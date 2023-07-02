package backend

import (
	"io"
	"time"
)

type Backend interface {
	List(path string, recursive bool) ([]string, error)
	Stat(filepath string) (FileInfo, error)
	Read(filepath string) ([]byte, error)
	ReadString(filepath string) (string, error)
	ReadStream(filepath string) (FileStream, error)
	Write(filepath string, data []byte) error
	WriteString(filepath string, content string) error
	WriteStream(filepath string, stream io.ReadCloser, length int64) error
	Move(filepathSrc string, filepathDst string) error
	Delete(filepath string) error
}

type FileInfo struct {
	LastModified time.Time
	ETag         string
	ContentType  string
	Size         int64
}

type FileStream struct {
	Size        int64
	ContentType string
	Content     io.ReadCloser
}
