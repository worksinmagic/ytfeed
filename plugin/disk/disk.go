package disk

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type Disk struct {
	dirPath string
}

func (d *Disk) Exists(ctx context.Context, name string) (exists bool, err error) {
	name = filepath.Join(d.dirPath, name)

	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		err = nil
		return
	}
	if err != nil {
		return
	}

	exists = true

	return
}

func (d *Disk) Delete(ctx context.Context, name string) (err error) {
	name = filepath.Join(d.dirPath, name)
	err = os.Remove(name)

	return
}

func (d *Disk) SaveAs(ctx context.Context, name string, r io.Reader) (written int64, err error) {
	name = filepath.Join(d.dirPath, name)

	var f io.WriteCloser
	f, err = os.Create(name)
	if err != nil {
		return
	}

	written, err = io.Copy(f, r)

	return
}

func New(dirPath string) (d *Disk, err error) {
	d = &Disk{}
	d.dirPath = dirPath

	return
}
