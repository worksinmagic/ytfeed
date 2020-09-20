package savevideo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetYoutubeDLCommand(t *testing.T) {
	t.Run("success mp4", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		tmpFilePath := "./"
		url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
		q := "1080"
		m := "mp4"
		l := false
		x, err := getYoutubeDLCommand(ctx, tmpFilePath, url, q, m, l)
		require.NoError(t, err)
		require.NotNil(t, x)

		expectedArgs := []string{"youtube-dl", "-f", "bestvideo[ext=mp4][height=1080]+bestaudio[ext=m4a]", "--merge-output-format", "mp4", "-o", "./", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}
		require.Equal(t, expectedArgs, x.Args)
	})

	t.Run("success webm", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		tmpFilePath := "./"
		url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
		q := "1080"
		m := "webm"
		l := false
		x, err := getYoutubeDLCommand(ctx, tmpFilePath, url, q, m, l)
		require.NoError(t, err)
		require.NotNil(t, x)

		expectedArgs := []string{"youtube-dl", "-f", "bestvideo[ext=webm][height=1080]+bestaudio[ext=webm]", "--merge-output-format", "webm", "-o", "./", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}
		require.Equal(t, expectedArgs, x.Args)
	})

	t.Run("failed quality check", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		tmpFilePath := "./"
		url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
		q := "invalid"
		m := "webm"
		l := false
		x, err := getYoutubeDLCommand(ctx, tmpFilePath, url, q, m, l)
		require.Error(t, err)
		require.Nil(t, x)
	})

	t.Run("success live, and ignore extension setting", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		tmpFilePath := "./"
		url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
		q := "1080"
		m := "webm"
		l := true
		x, err := getYoutubeDLCommand(ctx, tmpFilePath, url, q, m, l)
		require.NoError(t, err)
		require.NotNil(t, x)

		expectedArgs := []string{"youtube-dl", "-f", "[height=1080]", "-o", "./", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}
		require.Equal(t, expectedArgs, x.Args)
	})
}
