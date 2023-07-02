package main

import (
	"log"

	gofs "github.com/craimbault/go-fs"
	"github.com/craimbault/go-fs/internal/backend/local"
	"github.com/rs/zerolog"
)

func main() {

	// On initialise la config
	config := local.LocalConfig{
		BasePath: "/tmp/gosf-data",
		Debug:    true,
	}

	// On indique le niveau de log
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if config.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Println("Initializing backend ...")

	// On initalise le backend
	goFS, err := gofs.New(gofs.BACKEND_TYPE_LOCAL, config)
	if err != nil {
		log.Fatal("GOSF Backend initialization error : " + err.Error())
	}
	log.Println("Backend initialized")

	// On va creer des fichiers avec contenu
	log.Println("Writing multiple files ...")
	filesToWrite := map[string]string{
		"file1.txt":                            "Fichier 1 dans le dossier racine",
		"file2.txt":                            "Fichier 2 dans le dossier racine",
		"subfolder1/sf1_file1.txt":             "Fichier 1 dans le sous dossier 1",
		"subfolder1/sf1_file2.txt":             "Fichier 2 dans le sous dossier 1",
		"subfolder1/ssf1.1/sf1_ssf1_file1.txt": "Fichier 1 dans le sous sous dossier 1 du sous dossier 1",
	}
	for filePath, fileContent := range filesToWrite {
		log.Printf("\t Writing file : %s\n", filePath)
		goFS.WriteString(filePath, fileContent)
		log.Printf("\t File writed with success\n")
	}
	log.Println("Files writing end")

	// On va lister tous les fichiers en reccursif
	log.Println("Listing all files recursively ...")
	currentFiles, err := goFS.List("", true)
	if err != nil {
		log.Fatal("Unable to list files : " + err.Error())
	}
	for _, filePath := range currentFiles {
		log.Printf("\t %s\n", filePath)
	}
	log.Println("Files listing end")

	if len(currentFiles) > 0 {
		// On va faire un stat sur le premier fichier trouve
		log.Println("Getting stats of the first file ...")
		fileStat, err := goFS.Stat(currentFiles[0])
		if err != nil {
			log.Fatal("Unable to stat file[" + currentFiles[0] + "] : " + err.Error())
		}
		log.Printf(
			"\t Stats : LastModified[%s] Etag[%s] ContentType[%s] Size[%d]\n",
			fileStat.LastModified,
			fileStat.ETag,
			fileStat.ContentType,
			fileStat.Size,
		)
		log.Println("File stat end")

		// On va deplacer le premier fichier
		newFilename := "new_file_moved.txt"
		log.Println("Moving the first file to another location ...")
		err = goFS.Move(currentFiles[0], newFilename)
		if err != nil {
			log.Fatal("Unable to move file[" + currentFiles[0] + "] : " + err.Error())
		}
		log.Println("File move end")

		// On va lire le fichier deplace
		log.Println("Reading the moved file ...")
		content, err := goFS.ReadString(newFilename)
		if err != nil {
			log.Fatal("Unable to remove file[" + newFilename + "] : " + err.Error())
		}
		log.Printf("\tFile moved content :\n\n%s\n\n", content)
		log.Println("File read end")

		// On va supprimer le fichier deplace
		log.Println("Removing the moved file ...")
		err = goFS.Delete(newFilename)
		if err != nil {
			log.Fatal("Unable to remove file[" + newFilename + "] : " + err.Error())
		}
		log.Println("File removed")
	}
}
