package savevideo

import (
	"context"
	"fmt"
	"os/exec"
)

const (
	ErrInvalidVideoQuality   = "invalid video quality: %s"
	ErrInvalidVideoExtension = "invalid video extension: %s"

	// for regular video
	// youtube-dl -f "bestvideo[ext=%s][height=%s]+bestaudio[ext=%s]" --merge-output-format ext -o temporaryfiledir/randomname url
	// for live download
	// youtube-dl -f "[height=%s]" -o temporaryfiledir/randomname https://www.youtube.com/watch?v=yF2va6QbnOs
	YoutubeDLCommand         = "youtube-dl"
	FormatArg                = "-f"
	FormatArgValueFormat     = "bestvideo[ext=%s][height=%s]+bestaudio[ext=%s]"
	FormatArgValueLiveFormat = "[height=%s]"
	MergeOutputFormatArg     = "--merge-output-format"
	OutputArg                = "-o"

	AudioM4A  = "m4a"
	AudioWebm = "webm"
	AudioMP4  = "mp4"

	ExtensionMP4  = "mp4"
	ExtensionWebm = "webm"
	ExtensionMKV  = "mkv"

	Quality1080 = "1080"
	Quality720  = "720"
	Quality640  = "640"
	Quality480  = "480"
	Quality360  = "360"
	Quality240  = "240"
	Quality144  = "144"
)

func getYoutubeDLCommand(ctx context.Context, tmpFilePath, url, quality, ext string, isLive bool) (cmd *exec.Cmd, err error) {
	var videoExt, audioExt string
	switch ext {
	case ExtensionMP4:
		videoExt = ExtensionMP4
		audioExt = AudioM4A
	case ExtensionWebm:
		fallthrough
	case ExtensionMKV:
		videoExt = ExtensionWebm
		audioExt = AudioWebm
	default:
		err = fmt.Errorf(ErrInvalidVideoExtension, ext)
		return
	}

	switch quality {
	case Quality1080:
	case Quality720:
	case Quality640:
	case Quality480:
	case Quality360:
	case Quality240:
	case Quality144:
	default:
		err = fmt.Errorf(ErrInvalidVideoQuality, quality)
		return
	}

	ytdlArgs := make([]string, 0, 7)
	ytdlArgs = append(ytdlArgs, FormatArg)
	if isLive {
		ytdlArgs = append(ytdlArgs, fmt.Sprintf(FormatArgValueLiveFormat, quality))
	} else {
		ytdlArgs = append(ytdlArgs, fmt.Sprintf(FormatArgValueFormat, videoExt, quality, audioExt))
		ytdlArgs = append(ytdlArgs, MergeOutputFormatArg)
		ytdlArgs = append(ytdlArgs, ext)
	}
	ytdlArgs = append(ytdlArgs, OutputArg)
	ytdlArgs = append(ytdlArgs, tmpFilePath)
	ytdlArgs = append(ytdlArgs, url)

	cmd = exec.CommandContext(ctx, YoutubeDLCommand, ytdlArgs...)

	return
}
