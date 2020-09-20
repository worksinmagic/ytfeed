package streamschedule

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/worksinmagic/ytfeed/mock"
)

const (
	invalidXmlData        = "invalid"
	xmlData               = "valid"
	specialInvalidXMLData = "special invalid"
)

func TestStreamSchedule(t *testing.T) {
	invalidCount := 0
	maxInvalidCount := 1
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// make sure the data received is correct
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		if string(data) != xmlData {
			if string(data) == specialInvalidXMLData {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				if invalidCount < maxInvalidCount {
					w.WriteHeader(http.StatusBadRequest)
					invalidCount++
				}
			}
		}

		fmt.Fprintln(w, "RESPONSE")
	}))
	defer ts.Close()

	targetURL := ts.URL
	retryDelay := time.Millisecond
	workerInterval := 100 * time.Millisecond

	var wg sync.WaitGroup

	wg.Add(1)
	t.Run("RunWorker success", func(t *testing.T) {
		defer wg.Done()

		databasePath := filepath.Join(os.TempDir(), "database.db")
		defer os.Remove(databasePath)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		logger := mock.NewMockLogger(ctrl)

		s, err := New(logger, databasePath, targetURL, retryDelay, workerInterval, InfiniteRetries)
		require.NoError(t, err)
		require.NotNil(t, s)

		videoURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
		err = s.RegisterSchedule(time.Now(), xmlData, videoURL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = s.RunWorker(ctx)
		require.NoError(t, err)
	})

	wg.Wait()

	wg.Add(1)
	t.Run("RunWorker failed", func(t *testing.T) {
		defer wg.Done()

		databasePath := filepath.Join(os.TempDir(), "database.db")
		defer os.Remove(databasePath)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		logger := mock.NewMockLogger(ctrl)

		s, err := New(logger, databasePath, targetURL, retryDelay, workerInterval, InfiniteRetries)
		require.NoError(t, err)
		require.NotNil(t, s)

		videoURL := "https://www.youtube.com/watch?v=d1YBv2mWll0"
		err = s.RegisterSchedule(time.Now(), invalidXmlData, videoURL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("main error"),
			gomock.Any(),
		)

		err = s.RunWorker(ctx)
		require.NoError(t, err)
	})

	wg.Wait()

	wg.Add(1)
	t.Run("RunWorker exceeding retries", func(t *testing.T) {
		defer wg.Done()

		databasePath := filepath.Join(os.TempDir(), "database.db")
		defer os.Remove(databasePath)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		logger := mock.NewMockLogger(ctrl)
		retries := 1

		s, err := New(logger, databasePath, targetURL, retryDelay, workerInterval, retries)
		require.NoError(t, err)
		require.NotNil(t, s)

		videoURL := "https://www.youtube.com/watch?v=d1YBv2mWll0"
		err = s.RegisterSchedule(time.Now(), specialInvalidXMLData, videoURL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		logger.EXPECT().Warnf(
			gomock.AssignableToTypeOf("exceeding retries"),
			gomock.AssignableToTypeOf("key"),
			gomock.AssignableToTypeOf(retries),
		)

		err = s.RunWorker(ctx)
		require.NoError(t, err)
	})

	wg.Wait()

	t.Run("CloseDatabase", func(t *testing.T) {
		databasePath := filepath.Join(os.TempDir(), "database.db")
		defer os.Remove(databasePath)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		logger := mock.NewMockLogger(ctrl)

		s, err := New(logger, databasePath, targetURL, retryDelay, workerInterval, InfiniteRetries)
		require.NoError(t, err)
		require.NotNil(t, s)

		err = s.CloseDatabase()
		require.NoError(t, err)
	})
}
