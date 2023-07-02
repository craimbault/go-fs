package gofs

import (
	"errors"

	"github.com/craimbault/go-fs/internal/backend"
	"github.com/craimbault/go-fs/internal/backend/local"
	"github.com/craimbault/go-fs/internal/backend/s3"
)

type GoFSBackendType string

const (
	BACKEND_TYPE_LOCAL GoFSBackendType = "local"
	BACKEND_TYPE_S3    GoFSBackendType = "s3"
)

type GoFS struct {
	bType GoFSBackendType
	b     backend.Backend
}

func New(backendType GoFSBackendType, backendConfig interface{}) (GoFS, error) {
	// On initialise le retour
	gofs := GoFS{
		bType: backendType,
	}
	var err error

	// En fonction du type de backend
	switch backendType {
	case BACKEND_TYPE_LOCAL:
		config, ok := backendConfig.(local.LocalConfig)
		if !ok {
			return gofs, errors.New("s3 config is not valid")
		}
		gofs.b, err = local.New(config)
	case BACKEND_TYPE_S3:
		config, ok := backendConfig.(s3.S3Config)
		if !ok {
			return gofs, errors.New("s3 config is not valid")
		}
		gofs.b, err = s3.New(config)
	default:
		return gofs, errors.New("unknown backend type")
	}

	// On renvoi les infos
	return gofs, err
}

func (gfs *GoFS) List(path string, recursive bool) ([]string, error) {
	return gfs.b.List(path, recursive)
}
func (gfs *GoFS) Stat(filepath string) (backend.FileInfo, error) {
	return gfs.b.Stat(filepath)
}
func (gfs *GoFS) Read(filepath string) ([]byte, error) {
	return gfs.b.Read(filepath)
}
func (gfs *GoFS) ReadString(filepath string) (string, error) {
	return gfs.b.ReadString(filepath)
}
func (gfs *GoFS) Write(filepath string, data []byte) error {
	return gfs.b.Write(filepath, data)
}
func (gfs *GoFS) WriteString(filepath string, content string) error {
	return gfs.b.WriteString(filepath, content)
}
func (gfs *GoFS) Move(filepathSrc string, filepathDst string) error {
	return gfs.b.Move(filepathSrc, filepathDst)
}
func (gfs *GoFS) Delete(filepath string) error {
	return gfs.b.Delete(filepath)
}
