package gofsbcklocal

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/craimbault/go-fs/internal/backend"
	"github.com/rs/zerolog/log"
)

const BACKEND_NAME = "local"

type LocalConfig struct {
	BasePath string
	Debug    bool
}

type LocalBackend struct {
	Config LocalConfig
	Mu     sync.Mutex
}

func New(config LocalConfig) (*LocalBackend, error) {
	// On verifie que le BasePath existe
	basePathExists := true
	if !pathMustExists(config.BasePath) {
		// On indique que ca n'existe pas
		basePathExists = false

		// On le cree en reccursif
		err := createFolder(config.BasePath)
		if err != nil {
			return &LocalBackend{}, err
		}
	}

	// On informe
	log.Debug().
		Str("backend", "local").
		Str("basepath", config.BasePath).
		Bool("exists", basePathExists).
		Msg("Starting backend ...")

	// On initialise
	backend := LocalBackend{
		Config: config,
	}
	return &backend, nil
}

func (b *LocalBackend) List(path string, recursive bool) ([]string, error) {
	// On initialise
	prefixedPath := addPrefixedPath(b, path)
	files := make([]string, 0)
	prefixedPathLen := len(prefixedPath)

	log.Debug().
		Str("backend", "local").
		Str("action", "List").
		Str("path", prefixedPath).
		Send()

	// On parcours tous les elements
	filepath.Walk(prefixedPath, func(currentPath string, info os.FileInfo, err error) error {
		// Si ce n'est pas le chemin en cours
		if prefixedPath != currentPath {
			// On verifie si l'on a un dossier
			isRecursiveFolder := len(strings.Split(currentPath[prefixedPathLen+1:], string(os.PathSeparator))) > 1
			// Si l'on est en recursive ou que l'on y est pas (ni global, ni local) mais et que l'on a pas un dossier
			if (recursive || (!recursive && !isRecursiveFolder)) && !info.IsDir() {
				// On ajoute le nom du fichier dans la liste
				files = append(files, currentPath[prefixedPathLen:])
			}
		}
		return nil
	})

	return files, nil
}

func (b *LocalBackend) Stat(filePath string) (backend.FileInfo, error) {
	// On initialise
	prefixedFilePath := addPrefixedPath(b, filePath)
	fInfo := backend.FileInfo{}

	log.Debug().
		Str("backend", "local").
		Str("action", "Stat").
		Str("path", prefixedFilePath).
		Send()

	// On recupere les infos
	infos, err := os.Stat(prefixedFilePath)

	// Si le fichier n'existe pas
	if _, hasError := err.(*os.PathError); hasError {
		return fInfo, errors.New("filepath does not exists")
	} else if err != nil {
		return fInfo, errors.New("fs stat error : " + err.Error())
	}

	// On les ajoute au retour
	fInfo.ContentType = guessContentTypeFromFileExtention(infos.Name())
	fInfo.LastModified = infos.ModTime()
	fInfo.Size = infos.Size()

	// On renvoie les infos
	return fInfo, nil
}

func (b *LocalBackend) Read(filePath string) ([]byte, error) {
	// On initialise
	prefixedFilePath := addPrefixedPath(b, filePath)

	log.Debug().
		Str("backend", "local").
		Str("action", "Read").
		Str("path", prefixedFilePath).
		Send()

	// On verifie si le fichier existe
	exists := pathMustExists(prefixedFilePath)
	if !exists {
		return nil, errors.New("filepath does not exists")
	}

	// On lit le fichier
	return os.ReadFile(prefixedFilePath)
}

func (b *LocalBackend) ReadString(filePath string) (string, error) {
	// On utilise la methode existante
	data, err := b.Read(filePath)
	if err != nil {
		return "", err
	}

	// On converti en string
	return string(data), err
}

func (b *LocalBackend) ReadStream(filePath string) (backend.FileStream, error) {
	// On initialise
	prefixedFilePath := addPrefixedPath(b, filePath)
	fileStream := backend.FileStream{}

	log.Debug().
		Str("backend", "local").
		Str("action", "ReadStream").
		Str("path", prefixedFilePath).
		Send()

	// On verifie si le fichier existe
	exists := pathMustExists(prefixedFilePath)
	if !exists {
		return fileStream, errors.New("filepath does not exists")
	}

	// On recupere les infos du fichier
	objStat, err := b.Stat(filePath)
	if err != nil {
		return fileStream, errors.New("unable to get file info")
	}

	// On les ajoute au retour
	fileStream.ContentType = objStat.ContentType
	fileStream.Size = objStat.Size

	// On ouvre le stream
	fileStream.Content, err = os.OpenFile(prefixedFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return fileStream, errors.New("unable to read file")
	}

	// On revoi les infos
	return fileStream, nil
}

func (b *LocalBackend) Write(filePath string, data []byte) error {
	// On initialise
	prefixedFilePath := addPrefixedPath(b, filePath)

	log.Debug().
		Str("backend", "local").
		Str("action", "Write").
		Str("path", prefixedFilePath).
		Send()

	// On recupere le nom du dossier
	dirPath := filepath.Dir(prefixedFilePath)

	// Si le dossier existe pas
	if !pathMustExists(dirPath) {
		// On le cree
		createFolder(dirPath)
	}

	// On ecrit le fichier
	return os.WriteFile(prefixedFilePath, data, 0644)
}

func (b *LocalBackend) WriteString(filePath string, content string) error {
	// On utilise la methode existante
	return b.Write(filePath, []byte(content))
}

func (b *LocalBackend) WriteStream(filePath string, stream io.ReadCloser, length int64) error {
	// On initialise
	prefixedFilePath := addPrefixedPath(b, filePath)

	log.Debug().
		Str("backend", "local").
		Str("action", "Write").
		Str("path", prefixedFilePath).
		Send()

	// On recupere le nom du dossier
	dirPath := filepath.Dir(prefixedFilePath)

	// Si le dossier existe pas
	if !pathMustExists(dirPath) {
		// On le cree
		createFolder(dirPath)
	}

	// On ouvre le fichier en ecriture
	fd, err := os.OpenFile(prefixedFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	// On ecrit le fichier
	if _, err = io.Copy(fd, stream); err != nil {
		return err
	}
	defer stream.Close()

	// Tout est OK
	return nil
}

func (b *LocalBackend) Move(filePathSrc string, filePathDst string) error {
	// On initialise
	prefixedFilePathSrc := addPrefixedPath(b, filePathSrc)
	prefixedFilePathDst := addPrefixedPath(b, filePathDst)

	log.Debug().
		Str("backend", "local").
		Str("action", "Move").
		Str("src", prefixedFilePathSrc).
		Str("dst", prefixedFilePathDst).
		Send()

	// On deplace le fichier
	err := os.Rename(prefixedFilePathSrc, prefixedFilePathDst)
	if err != nil {
		return err
	}

	return nil
}

func (b *LocalBackend) Delete(filePath string) error {
	// On initialise
	prefixedFilePath := addPrefixedPath(b, filePath)

	log.Debug().
		Str("backend", "local").
		Str("action", "Delete").
		Str("path", prefixedFilePath).
		Send()

	return os.Remove(prefixedFilePath)
}
