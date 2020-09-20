package s3

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	UseSSL            = true
	UnknownObjectSize = -1

	DefaultToken            = ""
	DefaultCacheControl     = ""
	DefaultVideoContentType = "video/mp4"

	ErrorResponseNoSuchKey = "NoSuchKey"
)

type S3 struct {
	cli *minio.Client

	bucketName string
}

func (s *S3) Exists(ctx context.Context, name string) (exists bool, err error) {
	statOptions := minio.StatObjectOptions{}
	_, err = s.cli.StatObject(ctx, s.bucketName, name, statOptions)
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == ErrorResponseNoSuchKey {
			// file not exists
			err = nil
			return
		}

		return
	}

	exists = true

	return
}

func (s *S3) Delete(ctx context.Context, name string) (err error) {
	opts := minio.RemoveObjectOptions{}
	err = s.cli.RemoveObject(ctx, s.bucketName, name, opts)

	return
}

func (s *S3) SaveAs(ctx context.Context, name string, r io.Reader) (written int64, err error) {
	putOptions := minio.PutObjectOptions{}
	putOptions.CacheControl = DefaultCacheControl
	putOptions.ContentType = DefaultVideoContentType

	var info minio.UploadInfo
	info, err = s.cli.PutObject(ctx, s.bucketName, name, r, UnknownObjectSize, putOptions)
	if err != nil {
		return
	}
	written = info.Size

	return
}

func New(endpoint, accessKeyID, secretAccessKey, bucketName string, useSSL bool) (s *S3, err error) {
	s = &S3{}
	s.bucketName = bucketName
	s.cli, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, DefaultToken),
		Secure: useSSL,
	})

	return
}
