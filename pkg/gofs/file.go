package gofs

import (
	"io"
	"mime"
	"path"
	"time"
)

var (
	_readAll = io.ReadAll
)

type File struct {
	Path         string
	Name         string
	Size         int64
	ContentType  *string
	LastModified *time.Time
}

func (f *File) HasInfos() bool {
	return f.ContentType != nil && f.LastModified != nil
}

func (f *File) GetFullPath() string {
	return path.Join(f.Path, f.Name)
}

func NewFile(filePath, fileName string, contentLen int64, lastModified *time.Time) File {
	contentType := mime.TypeByExtension(path.Ext(fileName))
	file := File{
		Path:         filePath,
		Name:         fileName,
		Size:         contentLen,
		ContentType:  &contentType,
		LastModified: lastModified,
	}

	if file.Path == "." {
		file.Path = ""
	}

	return file
}

func NewFileFromFullFilePath(filePath string, contentLen int64, lastModified *time.Time) File {
	return NewFile(
		path.Dir(filePath),
		path.Base(filePath),
		contentLen,
		lastModified,
	)
}

type FileBytes struct {
	File
	Content []byte
}

func NewFileBytes(dirPath, fileName string, content []byte) FileBytes {
	return FileBytes{
		File:    NewFile(dirPath, fileName, int64(len(content)), nil),
		Content: content,
	}
}

func NewFileBytesFromFullFilePath(filePath string, content []byte) FileBytes {
	return NewFileBytes(
		path.Dir(filePath),
		path.Base(filePath),
		content,
	)
}

type FileStream struct {
	File
	Content io.Reader
}

func (f *FileStream) ToBytes() ([]byte, error) {
	return _readAll(f.Content)
}

func (f *FileStream) ToFileBytes() (fb FileBytes, err error) {
	fb.File = f.File
	fb.Content, err = f.ToBytes()
	return
}

func NewFileStream(dirPath, fileName string, contentLen int64, content io.Reader) FileStream {
	return FileStream{
		File:    NewFile(dirPath, fileName, contentLen, nil),
		Content: content,
	}
}

func NewFileStreamFromFullFilePath(filePath string, contentLen int64, content io.Reader) FileStream {
	return NewFileStream(
		path.Dir(filePath),
		path.Base(filePath),
		contentLen,
		content,
	)
}
