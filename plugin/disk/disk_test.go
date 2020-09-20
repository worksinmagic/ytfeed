package disk

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisk(t *testing.T) {
	dirName, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	d, err := New(dirName)
	require.NoError(t, err)
	require.NotNil(t, d)

	t.Run("SaveAs success", func(t *testing.T) {
		data := bytes.NewBufferString("data")
		written, err := d.SaveAs(context.TODO(), "test.txt", data)
		require.NoError(t, err)
		require.Equal(t, int64(4), written)

		writtenData, err := ioutil.ReadFile(filepath.Join(dirName, "test.txt"))
		if err != nil {
			panic(err)
		}
		require.Equal(t, "data", string(writtenData))
	})

	var wg sync.WaitGroup

	wg.Add(1)
	t.Run("Exists success file exists", func(t *testing.T) {
		defer wg.Done()

		exists, err := d.Exists(context.TODO(), "test.txt")
		require.NoError(t, err)
		require.True(t, exists)
	})

	wg.Wait()

	t.Run("Delete success", func(t *testing.T) {
		err = d.Delete(context.TODO(), "test.txt")
		require.NoError(t, err)
	})

	t.Run("Exists failed file does not exists", func(t *testing.T) {
		exists, err := d.Exists(context.TODO(), "test.txt")
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("Delete failed", func(t *testing.T) {
		err = d.Delete(context.TODO(), "test.txt")
		require.Error(t, err)
	})
}
