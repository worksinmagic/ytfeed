package gcs

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/require"
)

func TestGCS(t *testing.T) {
	bucketName := "test"
	fileName := "test.txt"

	svr := fakestorage.NewServer(nil)
	defer svr.Stop()

	svr.CreateBucketWithOpts(fakestorage.CreateBucketOpts{
		Name: bucketName,
	})
	httpClient := svr.HTTPClient()

	gcsClient, err := New(bucketName, "", httpClient)
	require.NoError(t, err)
	require.NotNil(t, gcsClient)

	t.Run("SaveAs success", func(t *testing.T) {
		data := bytes.NewBufferString("data")
		written, err := gcsClient.SaveAs(context.Background(), fileName, data)
		require.NoError(t, err)
		require.Equal(t, int64(4), written)

		obj, err := svr.GetObject(bucketName, fileName)
		if err != nil {
			panic(err)
		}
		require.Equal(t, "data", string(obj.Content))
	})

	var wg sync.WaitGroup

	wg.Add(1)
	t.Run("Exists success file exists", func(t *testing.T) {
		defer wg.Done()

		exists, err := gcsClient.Exists(context.TODO(), fileName)
		require.NoError(t, err)
		require.True(t, exists)
	})

	wg.Wait()

	t.Run("Delete success", func(t *testing.T) {
		err = gcsClient.Delete(context.TODO(), fileName)
		require.NoError(t, err)
	})

	t.Run("Exists failed file does not exists", func(t *testing.T) {
		exists, err := gcsClient.Exists(context.TODO(), fileName)
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("Delete failed", func(t *testing.T) {
		err = gcsClient.Delete(context.TODO(), fileName)
		require.Error(t, err)
	})
}
