package streamschedule

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/worksinmagic/ytfeed"
	"github.com/worksinmagic/ytfeed/mock"
)

const (
	videoURL  = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	videoURL2 = "https://www.youtube.com/watch?v=989-7xsRLR4"
	xmlData   = "xml data"
)

func TestStreamSchedule(t *testing.T) {
	// handler must be run exactly twice for different video URL
	// and sent videoURL must be different than the previous one
	var mockDataHandlerWG sync.WaitGroup
	var lock sync.Mutex
	runTimes := 0
	runVideoURL := ""
	mockDataHandler := func(ctx context.Context, d *ytfeed.Data) {
		defer mockDataHandlerWG.Done()

		lock.Lock()
		defer lock.Unlock()

		require.NotNil(t, d)

		if runVideoURL != "" {
			require.NotEqual(t, runVideoURL, d.Feed.Entry.Link.Href)
		}

		runVideoURL = d.Feed.Entry.Link.Href
		runTimes++
	}

	// handler must be run exactly twice for same video URL
	// and sent videoURL must be the same as the previous one
	var mockDataHandlerWG2 sync.WaitGroup
	var lock2 sync.Mutex
	runTimes2 := 0
	runVideoURL2 := ""
	mockDataHandler2 := func(ctx context.Context, d *ytfeed.Data) {
		defer mockDataHandlerWG2.Done()

		lock2.Lock()
		defer lock2.Unlock()

		require.NotNil(t, d)

		if runVideoURL2 != "" {
			require.Equal(t, runVideoURL2, d.Feed.Entry.Link.Href)
		}

		runVideoURL2 = d.Feed.Entry.Link.Href
		runTimes2++
	}

	var wg sync.WaitGroup

	wg.Add(1)
	t.Run("RunWorker success with different video URL", func(t *testing.T) {
		defer wg.Done()

		databasePath := filepath.Join(os.TempDir(), "database.db")
		defer os.Remove(databasePath)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		logger := mock.NewMockLogger(ctrl)

		workerInterval := 10 * time.Millisecond

		s, err := New(logger, databasePath, workerInterval)
		require.NoError(t, err)
		require.NotNil(t, s)

		s.RegisterDataHandler(mockDataHandler)

		mockDataHandlerWG.Add(1)
		go func() {
			data := &ytfeed.Data{}
			data.OriginalXMLMessage = xmlData
			data.Feed.Entry.Link.Href = videoURL

			err := s.RegisterSchedule(time.Now().Add(100*time.Millisecond), data)
			require.NoError(t, err)
		}()

		mockDataHandlerWG.Add(1)
		go func() {
			data2 := &ytfeed.Data{}
			data2.OriginalXMLMessage = xmlData
			data2.Feed.Entry.Link.Href = videoURL2

			err := s.RegisterSchedule(time.Now().Add(100*time.Millisecond), data2)
			require.NoError(t, err)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err = s.RunWorker(ctx)
		require.NoError(t, err)

		// wait until handler finished all req
		mockDataHandlerWG.Wait()
		require.Equal(t, 2, runTimes)
	})

	wg.Wait()

	wg.Add(1)
	t.Run("RunWorker success with same video URL", func(t *testing.T) {
		defer wg.Done()

		databasePath := filepath.Join(os.TempDir(), "database.db")
		defer os.Remove(databasePath)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		logger := mock.NewMockLogger(ctrl)

		workerInterval := 10 * time.Millisecond

		s, err := New(logger, databasePath, workerInterval)
		require.NoError(t, err)
		require.NotNil(t, s)

		s.RegisterDataHandler(mockDataHandler2)

		mockDataHandlerWG2.Add(1)
		go func() {
			data := &ytfeed.Data{}
			data.OriginalXMLMessage = xmlData
			data.Feed.Entry.Link.Href = videoURL

			err := s.RegisterSchedule(time.Now().Add(100*time.Millisecond), data)
			require.NoError(t, err)
		}()

		mockDataHandlerWG.Add(1)
		go func() {
			data := &ytfeed.Data{}
			data.OriginalXMLMessage = xmlData
			data.Feed.Entry.Link.Href = videoURL

			err := s.RegisterSchedule(time.Now().Add(100*time.Millisecond), data)
			require.NoError(t, err)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err = s.RunWorker(ctx)
		require.NoError(t, err)

		// wait until handler finished all req
		mockDataHandlerWG2.Wait()
		require.Equal(t, 1, runTimes2)
	})

	wg.Wait()

	t.Run("CloseDatabase", func(t *testing.T) {
		databasePath := filepath.Join(os.TempDir(), "database.db")
		defer os.Remove(databasePath)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		logger := mock.NewMockLogger(ctrl)

		workerInterval := 50 * time.Millisecond

		s, err := New(logger, databasePath, workerInterval)
		require.NoError(t, err)
		require.NotNil(t, s)

		s.RegisterDataHandler(mockDataHandler)

		err = s.CloseDatabase()
		require.NoError(t, err)
	})
}
