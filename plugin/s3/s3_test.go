package s3

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/require"
)

func TestS3(t *testing.T) {
	accessKeyID := "AKIAIOSFODNN7EXAMPLE"
	secretAccessKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	bucketName := "test"
	fileName := "test.txt"
	endpoint := "127.0.0.1:9000"
	useSSL := false

	s, err := New(endpoint, accessKeyID, secretAccessKey, bucketName, useSSL)
	require.NoError(t, err)
	require.NotNil(t, s)

	_ = s.cli.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})

	t.Run("SaveAs success", func(t *testing.T) {
		data := bytes.NewBufferString("data")
		written, err := s.SaveAs(context.Background(), fileName, data)
		require.NoError(t, err)
		require.Equal(t, int64(4), written)
	})

	var wg sync.WaitGroup

	wg.Add(1)
	t.Run("Exists success file exists", func(t *testing.T) {
		defer wg.Done()

		exists, err := s.Exists(context.TODO(), "test.txt")
		require.NoError(t, err)
		require.True(t, exists)
	})

	wg.Wait()

	t.Run("Delete success", func(t *testing.T) {
		err = s.Delete(context.TODO(), "test.txt")
		require.NoError(t, err)
	})

	t.Run("Exists failed file does not exists", func(t *testing.T) {
		exists, err := s.Exists(context.TODO(), "test.txt")
		require.NoError(t, err)
		require.False(t, exists)
	})

	// S3 doesn't return deletion failed on non-exist file
	// t.Run("Delete failed", func(t *testing.T) {
	// 	err = s.Delete(context.TODO(), "test.txt")
	// 	require.Error(t, err)
	// })
}
