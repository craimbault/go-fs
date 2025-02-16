package gofslocal

import (
	"bytes"
	"errors"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReader struct {
	mock.Mock
}

func (m *mockReader) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func TestLocalWriteStream(t *testing.T) {
	contentStream := bytes.NewReader([]byte("content to write"))

	t.Run("OsCreateFailed", func(t *testing.T) {
		err := LocalWriteStream("/you-are-not-allowed-here.test", contentStream)

		assert.Error(t, err, "An error is expected")
	})

	t.Run("WriteFailed", func(t *testing.T) {
		expectedErr := errors.New("io copy failed")
		mr := new(mockReader)
		mr.On("Read", mock.Anything).Return(0, expectedErr).Once()
		err := LocalWriteStream("/tmp/test.test", mr)

		assert.Equal(t, expectedErr, err, "The expected error is not valid")
	})

	t.Run("WriteSuccess", func(t *testing.T) {
		filePath := "/tmp/test.test"
		os.Remove(filePath)

		err := LocalWriteStream(filePath, contentStream)

		assert.NoError(t, err, "No error expected")

		os.Remove(filePath)
	})
}

func TestGetFileFromPath(t *testing.T) {
	t.Run("WithSubFolder", func(t *testing.T) {
		filename := "myFile.txt"
		folder := "sub-folder"

		file := GetFileFromPath(
			path.Join(folder, filename),
		)

		assert.Equal(t, filename, file.Name, "File name is not valid")
		assert.Equal(t, folder, file.Path, "Path is not valid")
	})

	t.Run("WithoutSubFolder", func(t *testing.T) {
		filename := "myFile.txt"
		folder := ""

		file := GetFileFromPath(
			path.Join(folder, filename),
		)

		assert.Equal(t, filename, file.Name, "File name is not valid")
		assert.Equal(t, folder, file.Path, "Path is not valid")
	})
}
