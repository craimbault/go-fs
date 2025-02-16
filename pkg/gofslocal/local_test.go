package gofslocal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	gofs "github.com/craimbault/go-fs/pkg/gofs"
	"github.com/stretchr/testify/assert"
)

var (
	// Check if the LocalBackend is compatible with the generic interface
	_ gofs.Backend = &LocalBackend{}
)

func TestNewLocalBackend(t *testing.T) {
	t.Run("BasePathExists", func(t *testing.T) {
		expectedBasePath := "demo"
		_localStat = func(basePath string) (os.FileInfo, error) {
			assert.Equal(t, expectedBasePath, basePath, "BasePath not valid")
			return nil, nil
		}

		c, err := NewLocalBackend(expectedBasePath)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, expectedBasePath, c.basePath, "The basePath is not the expected one")
	})

	t.Run("BasePathNotExists_MkdirFailed", func(t *testing.T) {
		expectedBasePath := "demo"
		expectedErr := errors.New("mkdir failed")
		_localStat = func(basePath string) (os.FileInfo, error) {
			assert.Equal(t, expectedBasePath, basePath, "BasePath not valid")
			return nil, errors.New("not exists")
		}
		_localMkdirAll = func(basePath string, perms os.FileMode) error {
			assert.Equal(t, expectedBasePath, basePath, "BasePath not valid")
			assert.Equal(t, CREATE_DIR_PERMISSIONS, perms, "Perms not valid")
			return expectedErr
		}

		_, err := NewLocalBackend(expectedBasePath)

		assert.ErrorAs(t, err, &gofs.BackendInitError{})
		assert.ErrorAs(t, err, &CreateBaseDirectoryError{})
	})

	t.Run("BasePathNotExists_MkdirSuccess", func(t *testing.T) {
		expectedBasePath := "demo"
		_localStat = func(basePath string) (os.FileInfo, error) {
			assert.Equal(t, expectedBasePath, basePath, "BasePath not valid")
			return nil, errors.New("not exists")
		}
		_localMkdirAll = func(basePath string, perms os.FileMode) error {
			assert.Equal(t, expectedBasePath, basePath, "BasePath not valid")
			assert.Equal(t, CREATE_DIR_PERMISSIONS, perms, "Perms not valid")
			return nil
		}

		c, err := NewLocalBackend(expectedBasePath)

		assert.NoError(t, err, "Expect no error")
		assert.Equal(t, expectedBasePath, c.basePath, "The basePath is not the expected one")
	})
}

func TestGetFullFilePath(t *testing.T) {
	file := gofs.File{
		Path: "path",
		Name: "test.test",
	}
	base := "test"
	c := LocalBackend{basePath: base}

	expectedFfp := path.Join(base, file.Path, file.Name)
	ffp := c.getFullFilePathFromFile(file)

	assert.Equal(t, expectedFfp, ffp, "Full FilePath not valid")
}

func TestStat(t *testing.T) {
	file := gofs.File{
		Path: "path",
		Name: "test.test",
	}
	c := LocalBackend{basePath: "test"}
	ctx := context.Background()

	t.Run("StatErr", func(t *testing.T) {
		_localStat = func(fullFilePath string) (os.FileInfo, error) {
			assert.Equal(t, path.Join(c.basePath, file.Path, file.Name), fullFilePath, "Full FilePath not valid")
			return nil, errors.New("error")
		}
		defer func() { _localStat = os.Stat }()

		err := c.Stat(ctx, &file)

		assert.ErrorAs(t, err, &gofs.StatError{}, "Expected error does not implement StatError")
	})

	t.Run("Success", func(t *testing.T) {
		contentType := "test/my-content-type"
		fileInfo := new(mockFileInfo)
		_localStat = func(fullFilePath string) (os.FileInfo, error) {
			assert.Equal(t, path.Join(c.basePath, file.Path, file.Name), fullFilePath, "Full FilePath not valid")
			return fileInfo, nil
		}
		_localMimeTypeByExtension = func(ext string) string {
			assert.Equal(t, ".test", ext, "Extention not valid")
			return contentType
		}
		defer func() {
			_localStat = os.Stat
			_localMimeTypeByExtension = mime.TypeByExtension
		}()

		modTime := time.Now()
		fileInfo.On("ModTime").Return(modTime).Once()
		size := int64(1234567890)
		fileInfo.On("Size").Return(size).Once()

		err := c.Stat(ctx, &file)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, modTime, *file.LastModified, "Expected LastModified not valid")
		assert.Equal(t, size, file.Size, "Expected Size not valid")
		assert.Equal(t, contentType, *file.ContentType, "Expected ContentType not valid")
	})
}

func TestList(t *testing.T) {
	basePath := "base"
	c := &LocalBackend{basePath: basePath}
	ctx := context.Background()
	dirPath := "demo"
	paths := map[string]*mockFileInfo{
		path.Join(basePath, dirPath, "file1.txt"):                   new(mockFileInfo),
		path.Join(basePath, dirPath, "subpath", "file-subpath.txt"): new(mockFileInfo),
	}
	for filepath, fileinfo := range paths {
		fileinfo.On("IsDir").Return(false)
		fileinfo.On("ModTime").Return(time.Now())
		fileinfo.On("Size").Return(int64(len(filepath)))
	}

	t.Run("LocalFilepathWalkFailed", func(t *testing.T) {
		expectedErr := errors.New("failed")
		flags := gofs.FLAG__RECURSIVE

		_localFilepathWalk = func(fullFilePath string, _ filepath.WalkFunc) error {
			assert.Equal(t, path.Join(c.basePath, dirPath), fullFilePath, "Full FilePath not valid")
			return expectedErr
		}
		defer func() { _localFilepathWalk = filepath.Walk }()

		files, err := c.List(ctx, dirPath, flags)

		assert.Zero(t, len(files), "Expected no files")
		assert.ErrorContains(t, err, expectedErr.Error(), "Expected error not valid")
		assert.ErrorAs(t, err, &gofs.ListError{}, "Expected error does not implement ListError")
	})

	t.Run("StatFailed", func(t *testing.T) {
		expectedErr := errors.New("error")
		flags := gofs.FLAG__RECURSIVE | gofs.FLAG__WITH_INFOS

		_localFilepathWalk = func(_ string, fn filepath.WalkFunc) (err error) {
			for filepath, fileinfo := range paths {
				err = fn(filepath, fileinfo, nil)
				if err != nil {
					break
				}
			}
			return err
		}

		_localStat = func(name string) (os.FileInfo, error) {
			return nil, expectedErr
		}

		defer func() {
			_localFilepathWalk = filepath.Walk
			_localStat = os.Stat
		}()

		_, err := c.List(ctx, dirPath, flags)

		assert.ErrorAs(t, err, &gofs.StatError{}, "Expected error does not implement StatError")
	})

	t.Run("SuccessWithRecusrsive", func(t *testing.T) {
		flags := gofs.FLAG__RECURSIVE | gofs.FLAG__WITH_INFOS

		_localFilepathWalk = func(_ string, fn filepath.WalkFunc) error {
			for filepath, fileinfo := range paths {
				fn(filepath, fileinfo, nil)
			}
			return nil
		}

		_localStat = func(name string) (os.FileInfo, error) {
			fileinfo, ok := paths[name]
			if !ok {
				return nil, errors.New("not found")
			}
			return fileinfo, nil
		}
		defer func() {
			_localFilepathWalk = filepath.Walk
			_localStat = os.Stat
		}()

		files, err := c.List(ctx, dirPath, flags)

		assert.Equal(t, 2, len(files), "Expected two files")
		assert.NoError(t, err, "Expected no error")
	})

	t.Run("SuccessWithoutRecusrsive", func(t *testing.T) {
		_localFilepathWalk = func(_ string, fn filepath.WalkFunc) error {
			for filepath, fileinfo := range paths {
				fn(filepath, fileinfo, nil)
			}
			return nil
		}
		_localStat = func(name string) (os.FileInfo, error) {
			fileinfo, ok := paths[name]
			if !ok {
				return nil, errors.New("not found")
			}
			return fileinfo, nil
		}
		defer func() {
			_localFilepathWalk = filepath.Walk
			_localStat = os.Stat
		}()

		files, err := c.List(ctx, dirPath, gofs.Flag(0))

		assert.Equal(t, 1, len(files), "Expected one files")
		assert.NoError(t, err, "Expected no error")
	})
}

func TestRead(t *testing.T) {
	c := &LocalBackend{basePath: "base"}
	ctx := context.Background()
	file := gofs.File{
		Path: "path",
		Name: "test.test",
	}

	t.Run("Failed", func(t *testing.T) {
		expectedErr := errors.New("read failed")
		_localReadFile = func(fileFullPath string) ([]byte, error) {
			assert.Equal(t, path.Join(c.basePath, file.Path, file.Name), fileFullPath, "FileFullPath not valid")
			return nil, expectedErr
		}

		_, err := c.Read(ctx, file)

		assert.ErrorContains(t, err, expectedErr.Error(), "Expected error not valid")
		assert.ErrorAs(t, err, &gofs.ReadError{}, "Expected error does not implement ReadError")
	})

	t.Run("Success", func(t *testing.T) {
		expected := []byte("data")
		_localReadFile = func(fileFullPath string) ([]byte, error) {
			assert.Equal(t, path.Join(c.basePath, file.Path, file.Name), fileFullPath, "FileFullPath not valid")
			return expected, nil
		}

		fb, err := c.Read(ctx, file)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, expected, fb.Content, "Expected content not valid")
	})
}

func TestReadStream(t *testing.T) {
	c := &LocalBackend{basePath: "base"}
	ctx := context.Background()
	file := gofs.File{
		Path: "path",
		Name: "test.test",
	}

	t.Run("Failed", func(t *testing.T) {
		expectedErr := errors.New("stream failed")
		_localReadStream = func(fileFullPath string) (*os.File, error) {
			assert.Equal(t, path.Join(c.basePath, file.Path, file.Name), fileFullPath, "FileFullPath not valid")
			return nil, expectedErr
		}

		_, err := c.ReadStream(ctx, file)

		assert.ErrorContains(t, err, expectedErr.Error(), "Expected error not valid")
		assert.ErrorAs(t, err, &gofs.ReadError{}, "Expected error does not implement ReadError")
	})

	t.Run("Success", func(t *testing.T) {
		expected := &os.File{}
		_localReadStream = func(fileFullPath string) (*os.File, error) {
			assert.Equal(t, path.Join(c.basePath, file.Path, file.Name), fileFullPath, "FileFullPath not valid")
			return expected, nil
		}

		fs, err := c.ReadStream(ctx, file)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, expected, fs.Content, "Expected content not valid")
	})
}

func TestPrepareWrite(t *testing.T) {
	basePath := "base"
	c := &LocalBackend{basePath: basePath}
	file := gofs.File{
		Path: "path",
		Name: "test.test",
	}

	t.Run("AlreadyExists", func(t *testing.T) {
		_localStat = func(dirPath string) (os.FileInfo, error) {
			assert.Equal(t, path.Join(c.basePath, file.Path), dirPath, "FileFullPath not valid")
			return nil, nil
		}

		fullFilePath, err := c.prepareWrite(file)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, path.Join(basePath, file.Path, file.Name), fullFilePath, "Full FilePath not valid")
	})

	t.Run("NotExistsMkdirSuccess", func(t *testing.T) {
		_localStat = func(dirPath string) (os.FileInfo, error) {
			assert.Equal(t, path.Join(c.basePath, file.Path), dirPath, "FileFullPath not valid")
			return nil, errors.New("not exists")
		}
		_localMkdirAll = func(dirPath string, perms os.FileMode) error {
			assert.Equal(t, path.Join(c.basePath, file.Path), dirPath, "FileFullPath not valid")
			assert.Equal(t, CREATE_DIR_PERMISSIONS, perms, "Perms not valid")
			return nil
		}

		fullFilePath, err := c.prepareWrite(file)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, path.Join(basePath, file.Path, file.Name), fullFilePath, "Full FilePath not valid")
	})
}

func TestWrite(t *testing.T) {
	var (
		mkdirAllErr  error
		writeFileErr error
	)

	basePath := "base"
	c := &LocalBackend{basePath: basePath}
	ctx := context.Background()
	fileContent := []byte("tes bytes content")
	file := gofs.FileBytes{
		File: gofs.File{
			Path: "path",
			Name: "test.test",
			Size: int64(len(fileContent)),
		},
		Content: fileContent,
	}

	_localStat = func(dirPath string) (os.FileInfo, error) {
		assert.Equal(t, path.Join(c.basePath, file.Path), dirPath, "FileFullPath not valid")
		return nil, errors.New("force mkdir all")
	}
	_localMkdirAll = func(dirPath string, perms os.FileMode) error {
		assert.Equal(t, path.Join(c.basePath, file.Path), dirPath, "FileFullPath not valid")
		assert.Equal(t, CREATE_DIR_PERMISSIONS, perms, "Perms not valid")
		return mkdirAllErr
	}
	_localWriteFile = func(fileFullPath string, content []byte, perms os.FileMode) error {
		assert.Equal(t, path.Join(c.basePath, file.Path, file.Name), fileFullPath, "FileFullPath not valid")
		assert.Equal(t, fileContent, content, "Content not valid")
		assert.Equal(t, CREATE_FILE_PERMISSIONS, perms, "Perms not valid")
		return writeFileErr
	}

	t.Run("PrepareWriteFailed", func(t *testing.T) {
		// On force une erreur dans le prepare
		mkdirAllErr = errors.New("prepare error")

		err := c.Write(ctx, file)

		assert.ErrorContains(t, err, mkdirAllErr.Error(), "Expected error not valid")
		assert.ErrorAs(t, err, &gofs.WriteError{}, "Expected error does not implement WriteError")
	})

	t.Run("WriteSuccess", func(t *testing.T) {
		// On force pour ne pas avoir d'erreur
		mkdirAllErr = nil
		writeFileErr = nil

		err := c.Write(ctx, file)

		assert.NoError(t, err, "No error expected")
	})

	t.Run("WriteError", func(t *testing.T) {
		// On force pour avoir le resultat attendu
		mkdirAllErr = nil
		writeFileErr = errors.New("error")

		err := c.Write(ctx, file)

		assert.ErrorContains(t, err, writeFileErr.Error(), "Error does not contains the expected content")
		assert.ErrorAs(t, err, &gofs.WriteError{}, "Expected error does not implement WriteError")
	})
}

func TestWriteStream(t *testing.T) {
	var (
		mkdirAllErr    error
		writeStreamErr error
	)

	basePath := "base"
	c := &LocalBackend{basePath: basePath}
	ctx := context.Background()
	fileContent := []byte("tes bytes content")
	fileStream := bytes.NewReader(fileContent)
	file := gofs.FileStream{
		File: gofs.File{
			Path: "path",
			Name: "test.test",
			Size: int64(len(fileContent)),
		},
		Content: fileStream,
	}

	_localStat = func(dirPath string) (os.FileInfo, error) {
		assert.Equal(t, path.Join(c.basePath, file.Path), dirPath, "FileFullPath not valid")
		return nil, errors.New("force mkdir all")
	}
	_localMkdirAll = func(dirPath string, perms os.FileMode) error {
		assert.Equal(t, path.Join(c.basePath, file.Path), dirPath, "FileFullPath not valid")
		assert.Equal(t, CREATE_DIR_PERMISSIONS, perms, "Perms not valid")
		return mkdirAllErr
	}
	_localWriteStream = func(fileFullPath string, content io.Reader) error {
		assert.Equal(t, path.Join(c.basePath, file.Path, file.Name), fileFullPath, "FileFullPath not valid")
		assert.Equal(t, fileStream, content, "Content not valid")
		return writeStreamErr
	}

	t.Run("PrepareWriteFailed", func(t *testing.T) {
		// On force une erreur dans le prepare
		mkdirAllErr = errors.New("prepare error")

		err := c.WriteStream(ctx, file)

		assert.ErrorContains(t, err, mkdirAllErr.Error(), "Expected error not valid")
		assert.ErrorAs(t, err, &gofs.WriteError{}, "Expected error does not implement WriteError")
	})

	t.Run("WriteSuccess", func(t *testing.T) {
		mkdirAllErr = nil
		writeStreamErr = nil

		err := c.WriteStream(ctx, file)

		assert.NoError(t, err, "No error expected")
	})

	t.Run("WriteError", func(t *testing.T) {
		mkdirAllErr = nil
		writeStreamErr = errors.New("write stream error")

		err := c.WriteStream(ctx, file)

		assert.ErrorContains(t, err, writeStreamErr.Error(), "Error does not contains the expected value")
		assert.ErrorAs(t, err, &gofs.WriteError{}, "Expected error does not implement WriteError")
	})
}

func TestMove(t *testing.T) {
	basePath := "base"
	c := &LocalBackend{basePath: basePath}
	ctx := context.Background()
	srcFile := gofs.File{
		Path: "path",
		Name: "src.test",
	}
	dstFile := gofs.File{
		Path: "path",
		Name: "dst.test",
	}

	var (
		mkdirAllErr error
		statErrors  []error
		moveFilErr  error
		i           int
	)
	_localStat = func(filepath string) (os.FileInfo, error) {
		err := statErrors[i]
		i += 1
		return nil, err
	}

	_localMkdirAll = func(_ string, _ os.FileMode) error {
		return mkdirAllErr
	}

	_localMoveFile = func(oldpath, newpath string) error {
		return moveFilErr
	}

	defer func() {
		_localMoveFile = os.Rename
		_localStat = os.Stat
		_localMkdirAll = os.MkdirAll
	}()

	t.Run("SrcStatFailed", func(t *testing.T) {
		expectedErr := errors.New("src stat err")
		i = 0
		statErrors = []error{
			expectedErr,
		}

		err := c.Move(ctx, srcFile, dstFile)

		assert.ErrorAs(t, err, &gofs.MoveSrcDoesNotExistsError{}, "Expected error does not implement MoveSrcDoesNotExistsError")
	})

	t.Run("DstAlreadyExists", func(t *testing.T) {
		i = 0
		statErrors = []error{
			nil,
			nil,
		}

		err := c.Move(ctx, srcFile, dstFile)

		assert.ErrorAs(t, err, &gofs.MoveDstDoesNotExistsError{}, "Expected error does not implement MoveDstDoesNotExistsError")
	})

	t.Run("PrepareWriteError", func(t *testing.T) {
		i = 0
		mkdirAllErr = errors.New("mkdir error")
		statErrors = []error{
			nil,
			errors.New("not exists"),
			errors.New("not exists"),
		}

		err := c.Move(ctx, srcFile, dstFile)

		assert.ErrorContains(t, err, mkdirAllErr.Error(), "Expected error not valid")
		assert.ErrorAs(t, err, &gofs.MoveError{}, "Expected error does not implement MoveError")
	})

	t.Run("MoveSuccess", func(t *testing.T) {
		i = 0
		mkdirAllErr = nil
		statErrors = []error{
			nil,
			errors.New("file exists"),
			nil,
		}
		moveFilErr = nil

		err := c.Move(ctx, srcFile, dstFile)

		assert.NoError(t, err, "Expected no error")
	})

	t.Run("MoveError", func(t *testing.T) {
		i = 0
		mkdirAllErr = nil
		statErrors = []error{
			nil,
			errors.New("file exists"),
			nil,
		}
		moveFilErr = errors.New("move file error")

		err := c.Move(ctx, srcFile, dstFile)

		assert.ErrorContains(t, err, moveFilErr.Error(), "Error does not contains the expected content")
		assert.ErrorAs(t, err, &gofs.MoveError{}, "Expected error does not implement MoveError")
	})
}

func TestDelete(t *testing.T) {
	var (
		statErr   error
		removeErr error
	)

	_localStat = func(_ string) (os.FileInfo, error) {
		return nil, statErr
	}
	_localRemove = func(_ string) error {
		return removeErr
	}

	basePath := "base"
	c := &LocalBackend{basePath: basePath}
	ctx := context.Background()
	file := gofs.File{
		Path: "path",
		Name: "test.test",
	}
	t.Run("NotExists", func(t *testing.T) {
		// On force une erreur dans le stat pour dire qu'il n'existe pas
		statErr = errors.New("does not exists")

		err := c.Delete(ctx, file)

		assert.ErrorContains(t, err, statErr.Error(), "Expected error not valid")
		assert.ErrorAs(t, err, &gofs.FileDoesNotExistsError{}, "Expected error does not implement FileDoesNotExistsError")
	})

	t.Run("DeleteSuccess", func(t *testing.T) {
		// On force pour ne pas avoir d'erreur
		statErr = nil
		removeErr = nil

		err := c.Delete(ctx, file)
		assert.NoError(t, err, "No error expected")
	})

	t.Run("DeleteError", func(t *testing.T) {
		// On force pour avoir ce qu'il faut
		statErr = nil
		removeErr = errors.New("delete error")

		err := c.Delete(ctx, file)
		assert.ErrorContains(t, err, removeErr.Error(), "Error does not contains the expected value")
		assert.ErrorAs(t, err, &gofs.DeleteError{}, "Expected error does not implement DeleteError")
	})
}
