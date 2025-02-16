package main

import (
	"context"
	"log"
	"path"

	"github.com/craimbault/go-fs/pkg/gofs"
	"github.com/craimbault/go-fs/pkg/gofslocal"
	"github.com/rs/zerolog"
)

func main() {
	// On indique le niveau de log
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Println("Initializing backend ...")

	// On initalise le backend
	ctx := context.Background()
	goFS, err := gofslocal.NewLocalBackend("/tmp/gosf-data")
	if err != nil {
		log.Fatal("GOSF Backend initialization error : " + err.Error())
	}
	log.Println("Backend initialized")

	// On va creer des fichiers avec contenu
	log.Println("Writing multiple files ...")
	filesToWrite := []gofs.FileBytes{
		gofs.NewFileBytesFromFullFilePath("file1.txt", []byte("Fichier 1 dans le dossier racine")),
		gofs.NewFileBytesFromFullFilePath("file2.txt", []byte("Fichier 2 dans le dossier racine")),
		gofs.NewFileBytesFromFullFilePath("subfolder1/sf1_file1.txt", []byte("Fichier 1 dans le sous dossier 1")),
		gofs.NewFileBytesFromFullFilePath("subfolder1/sf1_file2.txt", []byte("Fichier 2 dans le sous dossier 1")),
		gofs.NewFileBytesFromFullFilePath("subfolder1/ssf1.1/sf1_ssf1_file1.txt", []byte("Fichier 1 dans le sous sous dossier 1 du sous dossier 1")),
	}
	for _, fileBytes := range filesToWrite {
		log.Printf("\t Writing file : %s\n", fileBytes.GetFullPath())
		err := goFS.Write(ctx, fileBytes)
		if err != nil {
			log.Printf("\t\t File write error : %s\n", err.Error())
		} else {
			log.Printf("\t\t File writed with success\n")
		}
	}
	log.Println("Files writing end")

	// On va lister tous les fichiers en reccursif
	log.Println("Listing all files recursively ...")
	files, err := goFS.List(ctx, "", gofs.Flag(gofs.FLAG__RECURSIVE))
	if err != nil {
		log.Fatal("Unable to list files : " + err.Error())
	}
	for _, file := range files {
		log.Printf("\t %s\n", file.GetFullPath())
	}
	log.Println("Files listing end")

	if len(files) > 0 {
		// On va faire un stat sur le premier fichier trouve
		log.Println("Getting stats of the first file ...")
		err = goFS.Stat(ctx, &files[0])
		if err != nil {
			log.Fatal("Unable to stat file[" + files[0].GetFullPath() + "] : " + err.Error())
		}
		log.Printf(
			"\t Stats : LastModified[%s] ContentType[%s] Size[%d]\n",
			*files[0].LastModified,
			*files[0].ContentType,
			files[0].Size,
		)
		log.Println("File stat end")

		// On va deplacer le premier fichier
		newFilename := "new_file_moved.txt"
		dstFile := gofs.NewFileFromFullFilePath(path.Join(files[0].Path, newFilename), 0, nil)
		log.Println("Moving the first file to another location ...")
		err = goFS.Move(ctx, files[0], dstFile)
		if err != nil {
			log.Fatal("Unable to move file[" + files[0].GetFullPath() + "] : " + err.Error())
		}
		log.Println("File move end")

		// On va lire le fichier deplace
		log.Println("Reading the moved file ...")
		file, err := goFS.Read(ctx, dstFile)
		if err != nil {
			log.Fatal("Unable to read file[" + dstFile.GetFullPath() + "] : " + err.Error())
		}
		log.Printf("\tFile moved content :\n\n%s\n\n", file.Content)
		log.Println("File read end")

		// On va supprimer le fichier deplace
		log.Println("Removing the moved file ...")
		err = goFS.Delete(ctx, dstFile)
		if err != nil {
			log.Fatal("Unable to remove file[" + dstFile.GetFullPath() + "] : " + err.Error())
		}
		log.Println("File removed")
	}
}
