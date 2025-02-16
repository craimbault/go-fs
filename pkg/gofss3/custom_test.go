package gofss3

import (
	"context"
	"errors"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestMinioGetObject(t *testing.T) {
	ctx := context.Background()
	bucketName := "bucket-test"
	path := "demo"
	opts := minio.GetObjectOptions{}

	expectedMinioObj := minio.Object{}
	expectedErr := errors.New("custom error")

	s3c := new(mockS3Client)
	s3c.
		On("GetObject", ctx, bucketName, path, opts).
		Return(&expectedMinioObj, expectedErr).
		Once()

	obj, err := minioGetObbject(ctx, s3c, bucketName, path, opts)

	assert.Equal(t, expectedErr, err, "Expected error not valid")

	// On le re-converti en minio.Object pour pouvoir comparer
	minioObj, ok := obj.(*minio.Object)
	assert.True(t, ok, "Object should be a *minio.Object")
	assert.Equal(t, expectedMinioObj, *minioObj, "Expected object not valid")

}
