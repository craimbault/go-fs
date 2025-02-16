package gofslocal

import (
	"context"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/craimbault/go-fs/pkg/gofs"
)

const (
	CREATE_DIR_PERMISSIONS  os.FileMode = 0777
	CREATE_FILE_PERMISSIONS os.FileMode = 0644
)

var (
	_localMkdirAll            = os.MkdirAll
	_localReadFile            = os.ReadFile
	_localReadStream          = os.Open
	_localRemove              = os.Remove
	_localStat                = os.Stat
	_localWriteFile           = os.WriteFile
	_localWriteStream         = LocalWriteStream
	_localMoveFile            = os.Rename
	_localMimeTypeByExtension = mime.TypeByExtension
	_localFilepathWalk        = filepath.Walk
)

type LocalBackend struct {
	basePath string
}

func NewLocalBackend(basePath string) (LocalBackend, error) {
	b := LocalBackend{
		basePath: basePath,
	}

	// On controle que le basePath existe deja ou on le cree
	_, err := _localStat(basePath)
	if err != nil {
		err = _localMkdirAll(basePath, CREATE_DIR_PERMISSIONS)
		if err != nil {
			return b, gofs.NewBackendInitError(NewCreateBaseDirectoryError(nil))
		}
	}

	return b, nil
}

func (b *LocalBackend) getFullFilePathFromFile(file gofs.File) string {
	return path.Join(b.basePath, file.Path, file.Name)
}

func (b *LocalBackend) Stat(ctx context.Context, file *gofs.File) error {
	// On va chercher les infos
	localStat, err := _localStat(b.getFullFilePathFromFile(*file))
	if err != nil {
		return gofs.NewStatError(nil)
	}

	// On ajoute les infos
	lastModified := localStat.ModTime()
	file.LastModified = &lastModified
	size := localStat.Size()
	file.Size = size
	contentType := _localMimeTypeByExtension(path.Ext(file.Name))
	file.ContentType = &contentType

	return nil
}

func (b *LocalBackend) List(ctx context.Context, dirPath string, flags gofs.Flag) (files []gofs.File, err error) {
	// On initialise
	fullDirPath := path.Join(b.basePath, dirPath)
	fullDirPathLen := len(fullDirPath)
	basePathLen := len(b.basePath)

	// On regarde si on utilise la recursivite et le recuperation d'infos
	recursive := flags&gofs.FLAG__RECURSIVE != 0
	extendedFileInfos := flags&gofs.FLAG__WITH_INFOS != 0

	// On parcours tous les elements
	err = _localFilepathWalk(fullDirPath, func(currentPath string, info os.FileInfo, err error) error {
		// Si ce n'est pas le chemin en cours
		if fullDirPath != currentPath {
			// On verifie si l'on a un dossier
			isRecursiveFolder := len(strings.Split(currentPath[fullDirPathLen+1:], string(os.PathSeparator))) > 1

			// Si l'on est en recursive ou que l'on y est pas (ni global, ni local) mais et que l'on a pas un dossier
			if (recursive || (!recursive && !isRecursiveFolder)) && !info.IsDir() {
				// On genre le fichier
				file := GetFileFromPath(currentPath[basePathLen+1:])

				// Si on doit recuperer les infos du fichier
				if extendedFileInfos {
					// On va les chercher
					fileInfoErr := b.Stat(ctx, &file)
					if fileInfoErr != nil {
						return fileInfoErr
					}
				}

				// On ajoute le fichier dans la liste
				files = append(files, file)
			}
		}
		return nil
	})

	if err != nil {
		err = gofs.NewListError(err)
	}

	return
}

func (b *LocalBackend) Read(ctx context.Context, file gofs.File) (fb gofs.FileBytes, err error) {
	// On recupere le chemin complet
	fileFullPath := b.getFullFilePathFromFile(file)

	// On va chercher le fichier
	fb.Content, err = _localReadFile(fileFullPath)
	if err != nil {
		err = gofs.NewReadError(err)
		return
	}

	// On va chercher les infos
	b.Stat(ctx, &file)

	// On rajoute les infos
	fb.File = file
	return
}

func (b *LocalBackend) ReadStream(ctx context.Context, file gofs.File) (fs gofs.FileStream, err error) {
	// On recupere le chemin complet
	fileFullPath := b.getFullFilePathFromFile(file)

	// On va chercher le fichier
	fs.Content, err = _localReadStream(fileFullPath)
	if err != nil {
		err = gofs.NewReadError(err)
		return
	}

	// On va chercher les infos
	b.Stat(ctx, &file)

	// On rajoute les infos
	fs.File = file
	return
}

func (b *LocalBackend) prepareWrite(file gofs.File) (fileFullPath string, err error) {
	// On recupere les infos
	fileFullPath = b.getFullFilePathFromFile(file)
	fullDirPath := path.Dir(fileFullPath)

	// Si le chemin n'existe pas
	_, err = _localStat(fullDirPath)
	if err != nil {
		err = _localMkdirAll(fullDirPath, CREATE_DIR_PERMISSIONS)
	}

	return
}

func (b *LocalBackend) Write(ctx context.Context, file gofs.FileBytes) error {
	// On prepare l'ecriture
	fullFilePath, err := b.prepareWrite(file.File)
	if err != nil {
		return gofs.NewWriteError(err)
	}

	// On ecrit le fichier
	err = _localWriteFile(fullFilePath, file.Content, 0644)
	if err != nil {
		err = gofs.NewWriteError(err)
	}

	return err
}

func (b *LocalBackend) WriteStream(ctx context.Context, file gofs.FileStream) error {
	// On prepare l'ecriture
	fullFilePath, err := b.prepareWrite(file.File)
	if err != nil {
		return gofs.NewWriteError(err)
	}

	// On enregistre le contenu
	err = _localWriteStream(fullFilePath, file.Content)
	if err != nil {
		err = gofs.NewWriteError(err)
	}

	return err
}

func (b *LocalBackend) Move(ctx context.Context, srcFile, dstFile gofs.File) error {
	// On genere les chemins complets
	srcFullFilePath := b.getFullFilePathFromFile(srcFile)
	dstFullFilePath := b.getFullFilePathFromFile(dstFile)

	// On regarde si la source existe
	_, err := _localStat(srcFullFilePath)
	if err != nil {
		return gofs.NewMoveSrcDoesNotExistsError(nil)
	}

	// On regarde si la destination existe
	_, err = _localStat(dstFullFilePath)
	if err == nil {
		return gofs.NewMoveDstDoesNotExistsError(nil)
	}

	// On prepare l'ecriture
	_, err = b.prepareWrite(dstFile)
	if err != nil {
		return gofs.NewMoveError(err)
	}

	// On deplace le fichier
	err = _localMoveFile(srcFullFilePath, dstFullFilePath)
	if err != nil {
		err = gofs.NewMoveError(err)
	}

	return err
}

func (b *LocalBackend) Delete(ctx context.Context, file gofs.File) (err error) {
	fullFilePath := b.getFullFilePathFromFile(file)

	_, err = _localStat(fullFilePath)
	if err != nil {
		return gofs.NewFileDoesNotExistsError(nil)
	}

	err = _localRemove(fullFilePath)
	if err != nil {
		err = gofs.NewDeleteError(err)
	}

	return
}
