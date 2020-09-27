package streamschedule

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/worksinmagic/ytfeed"
	"go.etcd.io/bbolt"
)

const (
	DefaultFilePermission      = 0666
	DefaultDatabaseOpenTimeout = time.Second
	DefaultBucketName          = "ytfeed"

	ErrFailedToDeleteKeysFormat = "failed to delete key(s): %s"
)

type Databaser interface {
	Close() error
	Update(func(tx *bbolt.Tx) error) error
}

type Schedule struct {
	RunAt time.Time    `json:"run_at"`
	Data  *ytfeed.Data `json:"data"`
}

type StreamSchedule struct {
	logger         ytfeed.Logger
	workerInterval time.Duration
	database       Databaser
	dataHandlers   []ytfeed.DataHandlerFunc
}

func (s *StreamSchedule) RegisterDataHandler(d ...ytfeed.DataHandlerFunc) {
	s.dataHandlers = d
}

func (s *StreamSchedule) RegisterSchedule(runAt time.Time, data *ytfeed.Data) (err error) {
	sch := Schedule{}
	sch.RunAt = runAt
	sch.Data = data

	// ignore deleted entry
	if data.Feed.DeletedEntry.Link.Href != "" {
		return
	}

	var rawData []byte
	rawData, err = json.Marshal(sch)
	if err != nil {
		err = errors.Wrapf(err, "failed to json marshal schedule of video %s", sch.Data.Feed.Entry.Link.Href)
		return
	}

	err = s.database.Update(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket([]byte(DefaultBucketName))

		err = b.Put([]byte(sch.Data.Feed.Entry.Link.Href), rawData)
		return
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to schedule video %s", sch.Data.Feed.Entry.Link.Href)
	}

	return
}

func (s *StreamSchedule) RunWorker(ctx context.Context) (err error) {
	ticker := time.NewTicker(s.workerInterval)
	defer ticker.Stop()

	for {
		err = s.work(ctx)
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

func (s *StreamSchedule) work(ctx context.Context) (err error) {
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
			if time.Now().After(sch.RunAt) {
				for _, d := range s.dataHandlers {
					go d(ctx, sch.Data)
				}

				// prepare the success key(s) to be deleted
				successKeys = append(successKeys, string(k))
			}

			return
		})

		// we collect the error(s) instead of straight jumping out at the first error
		failedKeysToDelete := make([]FailedOperation, 0, len(successKeys))

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

type FailedOperation struct {
	Error error
	Key   string
}

func (s *StreamSchedule) CloseDatabase() (err error) {
	return s.database.Close()
}

func New(logger ytfeed.Logger, databasePath string, workerInterval time.Duration) (s *StreamSchedule, err error) {
	s = &StreamSchedule{}
	s.logger = logger
	s.workerInterval = workerInterval
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
