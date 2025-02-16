package gofslocal

import (
	"io"
	"os"
	"path"

	"github.com/craimbault/go-fs/pkg/gofs"
)

func LocalWriteStream(filePath string, content io.Reader) error {
	// On ouvre le fichier
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// On copie le contenu
	_, err = io.Copy(file, content)
	return err
}

func GetFileFromPath(filepath string) gofs.File {
	file := gofs.File{
		Name: path.Base(filepath),
		Path: path.Dir(filepath),
	}

	if file.Path == "." {
		file.Path = ""
	}

	return file
}
