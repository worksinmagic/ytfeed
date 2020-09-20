package streamschedule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/worksinmagic/ytfeed"
	"go.etcd.io/bbolt"
)

const (
	DefaultContentType         = "application/rss+xml"
	DefaultFilePermission      = 0666
	DefaultDatabaseOpenTimeout = time.Second
	DefaultBucketName          = "ytfeed"

	ErrFailedToResendFormat     = "failed to resend xml data with returned status code %d"
	ErrFailedToDeleteKeysFormat = "failed to delete key(s): %s"

	WarnExceedingRetriesFormat = "key %s is exceeding retries of %d"

	InfiniteRetries = 0
)

type HTTPSender interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type Databaser interface {
	Close() error
	Update(func(tx *bbolt.Tx) error) error
}

type Schedule struct {
	RunAt    time.Time `json:"run_at"`
	XMLData  string    `json:"xml_data"`
	VideoURL string    `json:"video_url"`
}

type StreamSchedule struct {
	logger         ytfeed.Logger
	client         HTTPSender
	targetURL      string
	workerInterval time.Duration
	database       Databaser
	retryDelay     time.Duration
	maxRetries     int

	retriesKeysMap  map[string]int
	retriesKeysLock sync.Mutex
}

func (s *StreamSchedule) RegisterSchedule(runAt time.Time, xmlData, videoURL string) (err error) {
	sch := Schedule{}
	sch.RunAt = runAt
	sch.XMLData = xmlData
	sch.VideoURL = videoURL

	var rawData []byte
	rawData, err = json.Marshal(sch)
	if err != nil {
		err = errors.Wrapf(err, "failed to json marshal schedule of video %s", sch.VideoURL)
		return
	}

	err = s.database.Update(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket([]byte(DefaultBucketName))

		err = b.Put([]byte(sch.VideoURL), rawData)
		return
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to schedule video %s", sch.VideoURL)
	}

	return
}

func (s *StreamSchedule) RunWorker(ctx context.Context) (err error) {
	ticker := time.NewTicker(s.workerInterval)
	defer ticker.Stop()

	for {
		err = s.work()
		if err != nil {
			s.logger.Errorf("Failed to work: %v", err)
		}

		select {
		case <-ticker.C:
			// continue
		case <-ctx.Done():
			return
		}
	}
}

func (s *StreamSchedule) work() (err error) {
	err = s.database.Update(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket([]byte(DefaultBucketName))

		successKeys := make([]string, 0, 16)
		err = b.ForEach(func(k, v []byte) (err error) {
			sch := Schedule{}
			err = json.Unmarshal(v, &sch)
			if err != nil {
				err = errors.Wrapf(err, "failed to unmarshal schedule json with key %s", string(k))
				return
			}

			// time to resend messages
			// if failed, we don't delete the key so it could be retried in the next pass
			if time.Now().After(sch.RunAt) {
				var resp *http.Response
				resp, err = s.client.Post(s.targetURL, DefaultContentType, bytes.NewReader([]byte(sch.XMLData)))
				if err != nil {
					// if failed increment failure count for this key, if failed after n times, prepare for deletion
					retry := s.incrementRetriesKey(string(k))

					if !retry {
						err = errors.Wrapf(err, "failed to resend xml data to %s with key %s", s.targetURL, string(k))
					}

					<-time.After(s.retryDelay)
					return
				}
				defer resp.Body.Close()
				defer func() {
					_, _ = io.Copy(ioutil.Discard, resp.Body)
				}()

				if resp.StatusCode >= http.StatusBadRequest {
					// if failed increment failure count for this key, if failed after n times, prepare for deletion
					retry := s.incrementRetriesKey(string(k))

					if !retry {
						err = fmt.Errorf(ErrFailedToResendFormat, resp.StatusCode)
						err = errors.Wrapf(err, "failed to resend xml data to %s with key %s", s.targetURL, string(k))
					}

					<-time.After(s.retryDelay)
					return
				}

				// prepare the success key(s) to be deleted
				successKeys = append(successKeys, string(k))
			}

			return
		})

		// we collect the error(s) instead of straight jumping out at the first error
		failedKeysToDelete := make([]FailedOperation, 0, len(successKeys))

		// delete exceeding retries key(s)
		for _, k := range s.failedRetriesKeys() {
			// delete key from database
			s.logger.Warnf(WarnExceedingRetriesFormat, k, s.maxRetries)
			err = b.Delete([]byte(k))
			if err != nil {
				failedKeysToDelete = append(failedKeysToDelete, FailedOperation{
					Error: err,
					Key:   k,
				})
			}
		}

		// delete success keys
		for _, k := range successKeys {
			// delete key from database
			err = b.Delete([]byte(k))
			if err != nil {
				failedKeysToDelete = append(failedKeysToDelete, FailedOperation{
					Error: err,
					Key:   k,
				})
			}
		}

		// if there is failure in deleting success keys, collect the error from those keys here
		if len(failedKeysToDelete) > 0 {
			errMessage := ""
			for _, fk := range failedKeysToDelete {
				errMessage += fmt.Sprintf("key %s and error %v, ", fk.Key, fk.Error)
			}
			err = fmt.Errorf(ErrFailedToDeleteKeysFormat, errMessage)
			return
		}

		return
	})

	return
}

func (s *StreamSchedule) failedRetriesKeys() (keys []string) {
	if s.maxRetries > InfiniteRetries {
		keys = make([]string, 0, 16)

		s.retriesKeysLock.Lock()
		defer s.retriesKeysLock.Unlock()

		for k, v := range s.retriesKeysMap {
			if v > s.maxRetries {
				keys = append(keys, k)

				delete(s.retriesKeysMap, k)
			}
		}
	}

	return
}

func (s *StreamSchedule) incrementRetriesKey(key string) (retry bool) {
	// only do this if not set to infinite retries
	if s.maxRetries > InfiniteRetries {
		s.retriesKeysLock.Lock()
		defer s.retriesKeysLock.Unlock()

		// increment the key
		s.retriesKeysMap[key]++

		// if key not exceeding max retries, set to retry
		if s.retriesKeysMap[key] <= s.maxRetries {
			retry = true
			return
		}
	}

	return
}

type FailedOperation struct {
	Error error
	Key   string
}

func (s *StreamSchedule) CloseDatabase() (err error) {
	return s.database.Close()
}

func New(logger ytfeed.Logger, databasePath, targetURL string, retryDelay, workerInterval time.Duration, maxRetries int) (s *StreamSchedule, err error) {
	s = &StreamSchedule{}
	s.logger = logger
	s.targetURL = targetURL
	s.workerInterval = workerInterval
	s.retryDelay = retryDelay
	s.client = http.DefaultClient
	s.retriesKeysMap = map[string]int{}
	s.maxRetries = maxRetries
	s.database, err = bbolt.Open(databasePath, DefaultFilePermission, &bbolt.Options{Timeout: DefaultDatabaseOpenTimeout})
	if err != nil {
		return
	}

	err = s.database.Update(func(tx *bbolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte(DefaultBucketName))
		return
	})

	return
}
