package gofss3

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type s3BackendObject interface {
	Read(p []byte) (n int, err error)
	Close() error
	Seek(offset int64, whence int) (int64, error)
	ReadAt(p []byte, off int64) (n int, err error)
	Stat() (minio.ObjectInfo, error)
}

// On force le retour avec un model interne car minio.Object n'est pas mockable
func minioGetObbject(ctx context.Context, client s3client, bucketName string, path string, opts minio.GetObjectOptions) (s3BackendObject, error) {
	return client.GetObject(ctx, bucketName, path, opts)
}
