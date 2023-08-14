package gofsbcklocal

import (
	"mime"
	"os"
	"path/filepath"
)

const DEFAULT_MIME_TYPE = "application/octet-stream"

func createFolder(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func guessContentTypeFromFileExtention(filename string) string {
	// On recupere le Mime depuis l'extention du fichier
	mime := mime.TypeByExtension(filepath.Ext(filename))

	// Si on a rien, on utilise celui par defaut
	if len(mime) == 0 {
		mime = DEFAULT_MIME_TYPE
	}

	return mime

}

func pathExists(path string) (bool, error) {
	// On recupere les infos sur le chemin
	_, err := os.Stat(path)

	// Si l'on a pas d'erreur
	if err == nil {
		return true, nil
	} else if err == os.ErrNotExist { // Si ça n'existe pas
		return false, nil
	} else { // Sinon si l'on a une erreur
		return false, err
	}

}

func pathMustExists(path string) bool {
	exists, _ := pathExists(path)
	return exists
}

func addPrefixedPath(b *LocalBackend, path string) string {
	return b.Config.BasePath + string(os.PathSeparator) + path
}
