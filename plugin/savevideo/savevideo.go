package savevideo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/worksinmagic/ytfeed"
	youtube "google.golang.org/api/youtube/v3"
)

const (
	LiveBroadcastContentLive      = "live"
	LiveBroadcastContentCompleted = "completed"
	LiveBroadcastContentNone      = "none"
	LiveBroadcastContentUpcoming  = "upcoming"

	DefaultTemplateName = "ytfeed-savestream-filename"

	DownloadingVideoStatus = true

	DefaultTemporaryDownloadDirectoryPermission = 0755

	ErrFileAlreadyExistsFormat = "file %s already exists"
)

var (
	defaultParts []string = []string{
		"snippet",
		"liveStreamingDetails",
	}
)

type Entry struct {
	Author                         string
	LinkURL                        string
	Title                          string
	VideoID                        string
	ChannelID                      string
	Published                      string
	PublishedYear                  int
	PublishedMonth                 string
	PublishedDay                   int
	PublishedHour                  int
	PublishedMinute                int
	PublishedSecond                int
	PublishedNanosecond            int
	PublishedTimeZone              string
	PublishedTimeZoneOffsetSeconds int
	VideoExtension                 string
	VideoQuality                   string
}

type YoutubeVideoLister interface {
	List(part []string) *youtube.VideosListCall
}

type DataSaver interface {
	SaveAs(ctx context.Context, name string, r io.Reader) (int64, error)
	Delete(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) (bool, error)
}

type FilenameTemplater interface {
	Execute(w io.Writer, data interface{}) error
}

type StreamScheduler interface {
	RegisterSchedule(runAt time.Time, data *ytfeed.Data) error
}

type SaveVideo struct {
	logger               ytfeed.Logger
	vs                   YoutubeVideoLister
	dataSaver            DataSaver
	filenameTemplate     FilenameTemplater
	videoFormatQuality   string
	videoFormatExtension string
	downloadingVideo     map[string]bool
	streamScheduler      StreamScheduler
	tmpDir               string
	maxRetries           int
	retryDelay           time.Duration
	downloadingVideoLock sync.Mutex
}

func (s *SaveVideo) DataHandler(ctx context.Context, d *ytfeed.Data) {
	// if deletion entry, ignore
	if d.Feed.DeletedEntry.Link.Href != "" {
		s.logger.Warnf("Deletion entry for %s, ignored", d.Feed.DeletedEntry.Link.Href)
		return
	}

	entry := Entry{}
	entry.Author = d.Feed.Entry.Author.Name
	entry.ChannelID = d.Feed.Entry.ChannelID
	entry.VideoID = d.Feed.Entry.VideoID
	entry.LinkURL = d.Feed.Entry.Link.Href
	entry.VideoExtension = s.videoFormatExtension
	entry.VideoQuality = s.videoFormatQuality

	// check first if currently downloading the same video
	if _, ok := s.downloadingVideo[entry.LinkURL]; ok {
		s.logger.Warnf("Already downloading video %s", entry.LinkURL)
		return
	}

	// set the downloading status
	s.downloadingVideoLock.Lock()
	s.downloadingVideo[entry.LinkURL] = DownloadingVideoStatus
	s.downloadingVideoLock.Unlock()
	defer func() {
		s.downloadingVideoLock.Lock()
		delete(s.downloadingVideo, entry.LinkURL)
		s.downloadingVideoLock.Unlock()
	}()

	vlcall := s.vs.List(defaultParts)
	vlcall = vlcall.Id(entry.VideoID)
	vlresp, err := vlcall.Do()
	if err != nil {
		s.logger.Errorf("Failed to get video %s info: %v", entry.LinkURL, err)
		return
	}
	if len(vlresp.Items) < 1 {
		s.logger.Warnf("No item in video list response of video %s", entry.LinkURL)
		return
	}
	item := vlresp.Items[0]

	// use Youtube's published_at instead of feed's
	// if live broadcast, use scheduled start time instead
	// if there is none, use current time instead
	isLiveBroadcast := (item.LiveStreamingDetails != nil)
	if isLiveBroadcast {
		entry.Published = item.LiveStreamingDetails.ScheduledStartTime
	} else {
		entry.Published = item.Snippet.PublishedAt
	}
	publishedDate, err := time.Parse(time.RFC3339Nano, entry.Published)
	if err != nil {
		s.logger.Warnf("Invalid published date %s: %v, using current time instead", entry.Published, err)
		publishedDate = time.Now()
	}
	entry.PublishedYear = publishedDate.Year()
	entry.PublishedMonth = publishedDate.Month().String()
	entry.PublishedDay = publishedDate.Day()
	entry.PublishedHour = publishedDate.Hour()
	entry.PublishedMinute = publishedDate.Minute()
	entry.PublishedSecond = publishedDate.Second()
	entry.PublishedNanosecond = publishedDate.Nanosecond()
	entry.PublishedTimeZone, entry.PublishedTimeZoneOffsetSeconds = publishedDate.Zone()

	fileName := bytes.NewBuffer(nil)
	err = s.filenameTemplate.Execute(fileName, entry)
	if err != nil {
		s.logger.Errorf("Failed to render file template name: %v", err)
		return
	}
	switch item.Snippet.LiveBroadcastContent {
	case LiveBroadcastContentNone:
		fallthrough
	case LiveBroadcastContentLive:
		s.logger.Infof("Downloading video %s", entry.LinkURL)

		if s.retryDelay > 0 && s.maxRetries > 0 {
			err = s.DownloadVideoWithRetries(ctx, s.retryDelay, s.maxRetries, fileName.String(), entry.LinkURL, s.videoFormatQuality, s.videoFormatExtension, isLiveBroadcast, s.dataSaver)
		} else {
			err = s.DownloadVideo(ctx, fileName.String(), entry.LinkURL, s.videoFormatQuality, s.videoFormatExtension, isLiveBroadcast, s.dataSaver)
		}

		if err != nil {
			s.logger.Errorf("Failed to download video %s: %v. Original message was: `%s`", entry.LinkURL, err, d.OriginalXMLMessage)
			return
		}

		s.logger.Infof("Video %s downloaded", entry.LinkURL)
	case LiveBroadcastContentUpcoming:
		s.logger.Infof("Upcoming stream video %s at %s", entry.LinkURL, item.LiveStreamingDetails.ScheduledStartTime)
		if s.streamScheduler != nil {
			var runAt time.Time
			runAt, err = time.Parse(time.RFC3339, item.LiveStreamingDetails.ScheduledStartTime)
			if err != nil {
				s.logger.Errorf("Failed to register schedule for stream video %s: %v", entry.LinkURL, err)
				return
			}
			s.logger.Infof("Registering video %s to scheduler to be ran at %s", entry.LinkURL, runAt)
			err = s.streamScheduler.RegisterSchedule(runAt, d)
			if err != nil {
				s.logger.Errorf("Failed to register schedule for stream video %s: %v. Original message was: `%s`", entry.LinkURL, err, d.OriginalXMLMessage)
				return
			}
		}
	default:
		s.logger.Warnf("Unexpected broadcast content %s for url: %s", item.Snippet.LiveBroadcastContent, entry.LinkURL)
	}
}

func (s *SaveVideo) DownloadVideoWithRetries(ctx context.Context, retryDelay time.Duration, maxRetries int, videoName, url, quality, ext string, isLive bool, dataSaver DataSaver) (err error) {
	retries := 0
	for {
		err = s.DownloadVideo(ctx, videoName, url, quality, ext, isLive, dataSaver)
		if err == nil {
			return
		}
		if IsErrorAlreadyExists(err) {
			return
		}
		retries++
		if retries > maxRetries {
			return
		}

		s.logger.Warnf("Failed to download %s with error '%v', retrying %d/%d", url, err, retries, maxRetries)

		select {
		case <-time.After(retryDelay):
		case <-ctx.Done():
			return
		}
	}
}

func (s *SaveVideo) DownloadVideo(ctx context.Context, videoName, url, quality, ext string, isLive bool, dataSaver DataSaver) (err error) {
	var exists bool
	exists, err = dataSaver.Exists(ctx, videoName)
	if err != nil {
		err = errors.Wrapf(err, "failed to check if file %s already exists", videoName)
		return
	}
	if exists {
		err = fmt.Errorf(ErrFileAlreadyExistsFormat, videoName)
		return
	}

	tmpDownloadDirPath := filepath.Join(s.tmpDir, videoName)
	// create the temporary download dir
	err = os.MkdirAll(tmpDownloadDirPath, DefaultTemporaryDownloadDirectoryPermission)
	if err != nil {
		err = errors.Wrapf(err, "failed to create temporary download dir at %s", tmpDownloadDirPath)
		return
	}

	// then delete the temporary download dir afterward
	defer func(logger ytfeed.Logger, tmpDownloadDirPath string) {
		err := os.RemoveAll(tmpDownloadDirPath)
		if err != nil {
			logger.Errorf("Failed to remove temporary download dir %s: %v", tmpDownloadDirPath, err)
		}
	}(s.logger, tmpDownloadDirPath)

	// the temporary file that will be saved
	tmpFilePath := filepath.Join(tmpDownloadDirPath, strings.Replace(videoName, "/", "-", -1))

	stdErrCollector := &bytes.Buffer{}
	var ytdlCmd *exec.Cmd
	ytdlCmd, err = getYoutubeDLCommand(ctx, tmpFilePath, url, quality, ext, isLive)
	if err != nil {
		err = errors.Wrapf(err, "failed to create youtube-dl command from parameters: %s, %s, %s, %s", videoName, url, quality, ext)
		return
	}
	ytdlCmd.Stderr = stdErrCollector

	// save to temporary file
	err = ytdlCmd.Run()
	if err != nil {
		err = errors.Wrapf(err, "failed to run youtube-dl command from parameters: %s, %s, %s, %s and stderr: %s", videoName, url, quality, ext, stdErrCollector.String())
		return
	}

	var tmpFile io.ReadCloser
	tmpFile, err = os.Open(tmpFilePath)
	if err != nil {
		err = errors.Wrapf(err, "failed to open temporary file %s", tmpFilePath)
		return
	}
	defer tmpFile.Close()

	// pipe to data saver
	_, err = dataSaver.SaveAs(ctx, videoName, tmpFile)
	if err != nil {
		err = errors.Wrapf(err, "failed to save stream from youtube-dl command args %v", ytdlCmd.Args)
		return
	}

	return
}

// SetStreamScheduler because stream scheduler is optional, it doesn't have to be present at constructor function
func (s *SaveVideo) SetStreamScheduler(sc StreamScheduler) {
	s.streamScheduler = sc
}

// SetRetry because retries is optional, it doesn't have to be present at constructor function
func (s *SaveVideo) SetRetries(retryDelay time.Duration, maxRetries int) {
	s.retryDelay = retryDelay
	s.maxRetries = maxRetries
}

func New(
	logger ytfeed.Logger, vs YoutubeVideoLister, dataSaver DataSaver,
	tmpDir, filenameTemplate, quality, ext string,
) (s *SaveVideo, err error) {
	s = &SaveVideo{}
	s.logger = logger
	s.vs = vs
	s.dataSaver = dataSaver
	s.filenameTemplate, err = template.New(DefaultTemplateName).Parse(filenameTemplate)
	s.videoFormatQuality = quality
	s.videoFormatExtension = ext
	s.downloadingVideo = make(map[string]bool, 8)
	s.tmpDir = tmpDir

	return
}
