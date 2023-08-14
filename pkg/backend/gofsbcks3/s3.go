package gofsbcks3

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/craimbault/go-fs/internal/backend"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

const BACKEND_NAME = "s3"

type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	PathPrefix      string
	Debug           bool
}

type S3Backend struct {
	client *minio.Client
	Config S3Config
}

func New(config S3Config) (*S3Backend, error) {
	// On initialise le backend
	backend := S3Backend{
		Config: config,
	}

	// On informe
	log.Debug().
		Str("backend", "local").
		Str("endpoint", config.Endpoint).
		Str("region", config.Region).
		Str("bucket_name", config.BucketName).
		Str("access_key", config.AccessKeyID).
		Str("prefix", config.PathPrefix).
		Msg("Starting backend ...")

	// On initialise le client
	s3Client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return &backend, errors.New("unable to initalize s3 client : " + err.Error())
	}

	// Si le bucket n'existe pas
	exists, err := s3Client.BucketExists(context.Background(), config.BucketName)
	if err != nil {
		return &backend, errors.New("unable to check bucket exists : " + err.Error())
	} else if !exists {
		return &backend, errors.New("bucket[" + config.BucketName + "] does not exists")
	}

	// On ajoute au backend
	backend.client = s3Client

	// On envoie le backend
	return &backend, nil
}

func (b *S3Backend) List(path string, recursive bool) ([]string, error) {
	// On initialise
	pathWithPrefix := addPrefixedPath(b, path)
	files := make([]string, 0)
	pathLen := len(pathWithPrefix)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// On recupere la liste
	objects := b.client.ListObjects(ctx, b.Config.BucketName, minio.ListObjectsOptions{
		Prefix:    pathWithPrefix,
		Recursive: recursive,
	})

	// On passse tous les elements
	for object := range objects {
		if object.Err != nil {
			return nil, errors.New("S3 Object error : " + object.Err.Error())
		}
		// Si l'on a pas un dossier
		if object.Key[len(object.Key)-1:] != "/" {
			// On le garde
			files = append(files, object.Key[pathLen:])
		}
	}

	// On revoi les fichiers
	return files, nil
}

func (b *S3Backend) Stat(filePath string) (backend.FileInfo, error) {
	// On initialise
	var fileInfo = backend.FileInfo{}
	filePathWithPrefix := addPrefixedPath(b, filePath)

	// On recupere les infos
	stat, err := b.client.StatObject(
		context.Background(),
		b.Config.BucketName,
		filePathWithPrefix,
		minio.GetObjectOptions{},
	)

	// Si l'on a une erreur
	if err != nil {
		log.Debug().Str("filepath", filePathWithPrefix).Msg("Unable to get file stats")
		return fileInfo, err
	}

	// On renvoi les infos
	return backend.FileInfo{
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		ETag:         stat.ETag,
		LastModified: stat.LastModified,
	}, nil
}

func (b *S3Backend) Read(filePath string) ([]byte, error) {
	// On initialise
	filePathWithPrefix := addPrefixedPath(b, filePath)

	// On va chercher le fichier
	object, err := b.client.GetObject(
		context.Background(),
		b.Config.BucketName,
		filePathWithPrefix,
		minio.GetObjectOptions{},
	)

	if err != nil {
		return nil, err
	}

	defer object.Close()

	// On lit tout le contenu
	return io.ReadAll(object)
}

func (b *S3Backend) ReadString(filePath string) (string, error) {
	data, err := b.Read(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (b *S3Backend) ReadStream(filePath string) (backend.FileStream, error) {
	// On initialise
	fileStream := backend.FileStream{}
	filePathWithPrefix := addPrefixedPath(b, filePath)

	// On va chercher le fichier
	object, err := b.client.GetObject(
		context.Background(),
		b.Config.BucketName,
		filePathWithPrefix,
		minio.GetObjectOptions{},
	)
	if err != nil {
		log.Debug().Msg("1")
		return fileStream, errors.New("Unable to get the requested file : " + err.Error())
	}

	// On recupere les infos du fichier
	fileInfo, err := b.Stat(filePath)
	if err != nil {
		return fileStream, errors.New("Unable to get the informations about the requested file : " + err.Error())
	}

	// On hydrate notre retour
	fileStream.Content = object
	fileStream.ContentType = fileInfo.ContentType
	fileStream.Size = fileInfo.Size

	// On renvoi tout
	return fileStream, nil
}

func (b *S3Backend) Write(filePath string, data []byte) error {
	// On initialise
	filePathWithPrefix := addPrefixedPath(b, filePath)

	log.Debug().Msg("FILE PATH : " + filePathWithPrefix)

	// On ecrit le fichier
	_, err := b.client.PutObject(
		context.Background(),
		b.Config.BucketName,
		filePathWithPrefix,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{},
	)
	return err
}

func (b *S3Backend) WriteString(filePath string, content string) error {
	return b.Write(filePath, []byte(content))
}

func (b *S3Backend) WriteStream(filePath string, stream io.ReadCloser, length int64) error {
	// On initialise
	filePathWithPrefix := addPrefixedPath(b, filePath)

	// On ecrit le fichier
	_, err := b.client.PutObject(
		context.Background(),
		b.Config.BucketName,
		filePathWithPrefix,
		stream,
		int64(length),
		minio.PutObjectOptions{},
	)
	return err
}

func (b *S3Backend) Move(filePathSrc string, filePathDst string) error {
	// On recupere le contenu du fichier source
	srcFile, err := b.ReadStream(filePathSrc)
	if err != nil {
		return errors.New("Unable to read src file : " + err.Error())
	}
	defer srcFile.Content.Close()

	// On reecrit le fichier de destination
	err = b.WriteStream(
		filePathDst,
		srcFile.Content,
		srcFile.Size,
	)
	if err != nil {
		return err
	}

	// On supprime le fichier source
	return b.Delete(filePathSrc)
}

func (b *S3Backend) Delete(filePath string) error {
	// On initialise
	filePathWithPrefix := addPrefixedPath(b, filePath)

	// On supprime
	return b.client.RemoveObject(
		context.Background(),
		b.Config.BucketName,
		filePathWithPrefix,
		minio.RemoveObjectOptions{},
	)
}
