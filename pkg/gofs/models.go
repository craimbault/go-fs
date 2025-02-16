package gofs

import (
	"context"
	"net/url"
	"time"
)

type Backend interface {
	List(ctx context.Context, path string, flags Flag) ([]File, error)
	Stat(ctx context.Context, file *File) error
	Read(ctx context.Context, file File) (FileBytes, error)
	ReadStream(ctx context.Context, file File) (FileStream, error)
	Write(ctx context.Context, file FileBytes) error
	WriteStream(ctx context.Context, file FileStream) error
	Move(ctx context.Context, srcFile, dstFile File) error
	Delete(ctx context.Context, file File) error
}

type ObjectBackend interface {
	Backend
	PresignedRead(ctx context.Context, file File, expiry time.Duration, reqParams url.Values) (*url.URL, error)
	PresignedWrite(ctx context.Context, file File, expiry time.Duration) (*url.URL, error)
}
