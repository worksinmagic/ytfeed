package gcs

import (
	"context"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

const (
	DefaultCacheControl     = ""
	DefaultVideoContentType = "video/mp4"
)

type GCS struct {
	cli *storage.Client

	bucketName string
}

func (g *GCS) Exists(ctx context.Context, name string) (exists bool, err error) {
	_, err = g.cli.Bucket(g.bucketName).Object(name).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		// file not exists
		err = nil
		return
	}
	if err != nil {
		return
	}

	exists = true

	return
}

func (g *GCS) Delete(ctx context.Context, name string) (err error) {
	err = g.cli.Bucket(g.bucketName).Object(name).Delete(ctx)

	return
}

func (g *GCS) SaveAs(ctx context.Context, name string, r io.Reader) (written int64, err error) {
	w := g.cli.Bucket(g.bucketName).Object(name).NewWriter(ctx)
	defer w.Close()

	w.CacheControl = DefaultCacheControl
	w.ContentType = DefaultVideoContentType

	written, err = io.Copy(w, r)

	return
}

func New(bucketName, credentialJSONFilePath string, httpClient *http.Client) (g *GCS, err error) {
	g = &GCS{}
	g.bucketName = bucketName
	options := make([]option.ClientOption, 0, 1)
	options = append(options, option.WithTelemetryDisabled())

	if credentialJSONFilePath != "" {
		options = append(options, option.WithCredentialsFile(credentialJSONFilePath))
	}
	if httpClient != nil {
		options = append(options, option.WithHTTPClient(httpClient))
	}

	g.cli, err = storage.NewClient(context.TODO(), options...)

	return
}
