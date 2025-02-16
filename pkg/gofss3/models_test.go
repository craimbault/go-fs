package gofss3

import (
	"context"
	"errors"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/mock"
)

type mockS3Client struct {
	mock.Mock
}

func (m *mockS3Client) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	args := m.Called(ctx, bucketName, opts)
	ch := make(chan minio.ObjectInfo, 1)
	ch <- args.Get(0).(minio.ObjectInfo)
	defer close(ch)
	return ch
}

func (m *mockS3Client) StatObject(ctx context.Context, bucketName string, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Get(0).(minio.ObjectInfo), args.Error(1)
}

func (m *mockS3Client) GetObject(ctx context.Context, bucketName string, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Get(0).(*minio.Object), args.Error(1)
}

func (m *mockS3Client) PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	return args.Get(0).(minio.UploadInfo), args.Error(1)
}

func (m *mockS3Client) CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, dst, src)
	return args.Get(0).(minio.UploadInfo), args.Error(1)
}

func (m *mockS3Client) RemoveObject(ctx context.Context, bucketName string, objectName string, opts minio.RemoveObjectOptions) error {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Error(0)
}

func (m *mockS3Client) PresignedGetObject(ctx context.Context, bucketName string, objectName string, expires time.Duration, reqParams url.Values) (u *url.URL, err error) {
	args := m.Called(ctx, bucketName, objectName, expires, reqParams)
	return args.Get(0).(*url.URL), args.Error(1)
}

func (m *mockS3Client) PresignedPutObject(ctx context.Context, bucketName string, objectName string, expires time.Duration) (u *url.URL, err error) {
	args := m.Called(ctx, bucketName, objectName, expires)
	return args.Get(0).(*url.URL), args.Error(1)
}

type mockS3BackendObject struct {
	mock.Mock
}

func (m *mockS3BackendObject) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Get(0).(int), args.Error(1)
}

func (m *mockS3BackendObject) Close() error {
	return m.Called().Error(0)
}

func (m *mockS3BackendObject) Seek(offset int64, whence int) (int64, error) {
	args := m.Called(offset, whence)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockS3BackendObject) ReadAt(p []byte, off int64) (n int, err error) {
	args := m.Called(p, off)
	return args.Get(0).(int), args.Error(1)
}

func (m *mockS3BackendObject) Stat() (minio.ObjectInfo, error) {
	args := m.Called()
	return args.Get(0).(minio.ObjectInfo), args.Error(1)
}

type customMockS3BackendObject struct {
	obj     io.Reader
	objInfo minio.ObjectInfo
	statErr error
}

func (m customMockS3BackendObject) Read(p []byte) (n int, err error) {
	return m.obj.Read(p)
}

func (m customMockS3BackendObject) Close() error {
	return errors.New("not implemented in tests")
}

func (m customMockS3BackendObject) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("not implemented in tests")
}

func (m customMockS3BackendObject) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, errors.New("not implemented in tests")
}

func (m customMockS3BackendObject) Stat() (minio.ObjectInfo, error) {
	return m.objInfo, m.statErr
}
