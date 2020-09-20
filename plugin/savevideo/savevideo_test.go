package savevideo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/worksinmagic/ytfeed"
	"github.com/worksinmagic/ytfeed/mock"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

func TestSaveVideo(t *testing.T) {
	// TODO: don't forget to remove the credential before you make a commit!
	// For your convenience if you just want to press "run test" on VSCode
	// os.Setenv("YTFEED_YOUTUBE_API_KEY", "")
	// os.Setenv("YTFEED_YOUTUBE_VIDEO_ID", "PCicKydX5GE")
	// os.Setenv("YTFEED_YOUTUBE_VIDEO_URL", "https://www.youtube.com/watch?v=PCicKydX5GE")

	youtubeAPIKey := os.Getenv("YTFEED_YOUTUBE_API_KEY")
	require.NotEmpty(t, youtubeAPIKey)

	videoID := os.Getenv("YTFEED_YOUTUBE_VIDEO_ID")
	require.NotEmpty(t, videoID)

	videoURL := os.Getenv("YTFEED_YOUTUBE_VIDEO_URL")
	require.NotEmpty(t, videoURL)

	filenameTemplate := "{{.VideoID}}.webm"
	quality := "144"
	ext := "webm"
	tmpDir := os.TempDir()

	tmpVideoFilePath := filepath.Join(tmpDir, videoID+".webm")

	yts, err := youtube.NewService(context.Background(), option.WithAPIKey(youtubeAPIKey))
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	t.Run("DataHandler success download regular video", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("LiveBroadcastContentNoneFallthrough"),
			gomock.AssignableToTypeOf(videoURL),
		)
		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("Video downloaded"),
			gomock.AssignableToTypeOf(videoURL),
		)

		fileExists := false
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, nil)
		dataSaver.EXPECT().SaveAs(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
			gomock.Any(),
		).Return(int64(0), nil)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler failed download regular video, failed to save data", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("LiveBroadcastContentNoneFallthrough"),
			gomock.AssignableToTypeOf(videoURL),
		)

		fileExists := false
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, nil)
		dataSaver.EXPECT().SaveAs(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
			gomock.Any(),
		).Return(int64(0), fmt.Errorf("error"))

		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("Failed to download video"),
			gomock.AssignableToTypeOf(videoURL),
			gomock.Any(),
			gomock.AssignableToTypeOf("original xml message"),
		)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler failed DeletedEntry", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.DeletedEntry = ytfeed.DeletedEntry{}
		mockData.Feed.DeletedEntry.Link = ytfeed.Link{}
		mockData.Feed.DeletedEntry.Link.Href = "deleted video"

		logger.EXPECT().Warnf(
			gomock.AssignableToTypeOf("Deleted entry"),
			gomock.AssignableToTypeOf("deleted video"),
		)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler failed already downloading video", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL
		sv.downloadingVideo[videoURL] = true

		logger.EXPECT().Warnf(
			gomock.AssignableToTypeOf("Already downloading video"),
			gomock.AssignableToTypeOf(videoURL),
		)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler failed file existence checker error", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("LiveBroadcastContentNoneFallthrough"),
			gomock.AssignableToTypeOf(videoURL),
		)

		fileExists := false
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, fmt.Errorf("error"))

		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("Failed to download video"),
			gomock.AssignableToTypeOf(videoURL),
			gomock.Any(),
			gomock.AssignableToTypeOf("original xml message"),
		)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler failed file already exists", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("LiveBroadcastContentNoneFallthrough"),
			gomock.AssignableToTypeOf(videoURL),
		)

		fileExists := true
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, nil)

		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("Failed to download video"),
			gomock.AssignableToTypeOf(videoURL),
			gomock.Any(),
			gomock.AssignableToTypeOf("original xml message"),
		)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler failed to create youtube-dl command", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, "invalid extension")
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("LiveBroadcastContentNoneFallthrough"),
			gomock.AssignableToTypeOf(videoURL),
		)

		fileExists := false
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, nil)

		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("Failed to download video"),
			gomock.AssignableToTypeOf(videoURL),
			gomock.Any(),
			gomock.AssignableToTypeOf("original xml message"),
		)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler retries failed exceeding retries", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		retryDelay := 10 * time.Millisecond
		maxRetries := 1

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)
		sv.SetRetries(retryDelay, maxRetries)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("LiveBroadcastContentNoneFallthrough"),
			gomock.AssignableToTypeOf(videoURL),
		)

		fileExists := false
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, nil)
		dataSaver.EXPECT().SaveAs(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
			gomock.Any(),
		).Return(int64(0), fmt.Errorf("error"))
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, nil)
		dataSaver.EXPECT().SaveAs(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
			gomock.Any(),
		).Return(int64(0), fmt.Errorf("error"))

		logger.EXPECT().Warnf(
			gomock.AssignableToTypeOf("Retrying download"),
			gomock.AssignableToTypeOf(videoURL),
			gomock.Any(),
			gomock.AssignableToTypeOf(0),
			gomock.AssignableToTypeOf(maxRetries),
		)
		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("Failed to download video"),
			gomock.AssignableToTypeOf(videoURL),
			gomock.Any(),
			gomock.AssignableToTypeOf("original xml message"),
		)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler retries failed file already exists", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		retryDelay := 10 * time.Millisecond
		maxRetries := 1

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)
		sv.SetRetries(retryDelay, maxRetries)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("LiveBroadcastContentNoneFallthrough"),
			gomock.AssignableToTypeOf(videoURL),
		)

		fileExists := true
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, nil)

		logger.EXPECT().Errorf(
			gomock.AssignableToTypeOf("Failed to download video"),
			gomock.AssignableToTypeOf(videoURL),
			gomock.Any(),
			gomock.AssignableToTypeOf("original xml message"),
		)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()

	wg.Add(1)
	t.Run("DataHandler retries success", func(t *testing.T) {
		defer wg.Done()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		defer os.Remove(tmpVideoFilePath)

		logger := mock.NewMockLogger(ctrl)
		dataSaver := mock.NewMockDataSaver(ctrl)
		streamScheduler := mock.NewMockStreamScheduler(ctrl)

		retryDelay := 10 * time.Millisecond
		maxRetries := 1

		sv, err := New(logger, yts.Videos, dataSaver, tmpDir, filenameTemplate, quality, ext)
		require.NoError(t, err)
		require.NotNil(t, sv)
		sv.SetStreamScheduler(streamScheduler)
		sv.SetRetries(retryDelay, maxRetries)

		mockData := &ytfeed.Data{}
		mockData.Feed = ytfeed.Feed{}
		mockData.Feed.Entry = ytfeed.Entry{}
		mockData.Feed.Entry.Published = time.Now().Format(time.RFC3339Nano)
		mockData.Feed.Entry.VideoID = videoID
		mockData.Feed.Entry.Link = ytfeed.Link{}
		mockData.Feed.Entry.Link.Href = videoURL

		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("LiveBroadcastContentNoneFallthrough"),
			gomock.AssignableToTypeOf(videoURL),
		)
		logger.EXPECT().Infof(
			gomock.AssignableToTypeOf("Video downloaded"),
			gomock.AssignableToTypeOf(videoURL),
		)

		fileExists := false
		dataSaver.EXPECT().Exists(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
		).Return(fileExists, nil)
		dataSaver.EXPECT().SaveAs(
			gomock.Any(),
			gomock.AssignableToTypeOf("video name"),
			gomock.Any(),
		).Return(int64(0), nil)

		sv.DataHandler(context.TODO(), mockData)
	})
	wg.Wait()
}
