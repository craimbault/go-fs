package gofss3

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/craimbault/go-fs/pkg/gofs"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Backend struct {
	s3c    s3client
	prefix string
	bucket string
}

type S3BackendOptions struct {
	Endpoint  string
	Prefix    string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	UseSsl    bool
}

type s3client interface {
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	StatObject(ctx context.Context, bucketName string, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	GetObject(ctx context.Context, bucketName string, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (info minio.UploadInfo, err error)
	CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error)
	RemoveObject(ctx context.Context, bucketName string, objectName string, opts minio.RemoveObjectOptions) error
	PresignedGetObject(ctx context.Context, bucketName string, objectName string, expires time.Duration, reqParams url.Values) (u *url.URL, err error)
	PresignedPutObject(ctx context.Context, bucketName string, objectName string, expires time.Duration) (u *url.URL, err error)
}

var (
	_uriToOptions    = UriToOptions
	_newS3Backend    = NewS3Backend
	_minioNew        = minio.New
	_customGetObject = minioGetObbject
)

func NewS3BackendFromUri(uri string) (*S3Backend, error) {
	// On parse l'uri
	opts, err := _uriToOptions(uri)
	if err != nil {
		return nil, err
	}

	return _newS3Backend(opts)
}

func NewS3Backend(opts S3BackendOptions) (*S3Backend, error) {
	// On initialise le backend
	b := S3Backend{
		bucket: opts.Bucket,
		prefix: opts.Prefix,
	}

	// Si on a pas de bucket
	if len(opts.Bucket) == 0 {
		return nil, gofs.NewBackendInitError(NewNoBucketError(nil))
	}

	// Si on a pas les infos d'authent
	if len(opts.AccessKey) == 0 || len(opts.SecretKey) == 0 {
		return nil, gofs.NewBackendInitError(NewNoCredentialsError(nil))
	}

	// On initialise Minio
	var err error
	b.s3c, err = _minioNew(opts.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""),
		Region: opts.Region,
		Secure: opts.UseSsl,
	})
	if err != nil {
		return nil, gofs.NewBackendInitError(NewClientCreateFailedError(err))
	}

	return &b, nil
}

func (b *S3Backend) getFullFilePathFromFile(file gofs.File) string {
	return path.Join(b.prefix, file.Path, file.Name)
}

func (b *S3Backend) List(ctx context.Context, dirPath string, flags gofs.Flag) (files []gofs.File, err error) {
	// On genere les elements
	prefixLen := len(b.prefix)
	if prefixLen > 0 {
		// On ajoute 1 car on aura un / avant le dirPath
		prefixLen += 1
	}

	// On va chercher les elements
	objects := b.s3c.ListObjects(
		ctx,
		b.bucket,
		minio.ListObjectsOptions{
			Prefix:    path.Join(b.prefix, dirPath),
			Recursive: flags&gofs.FLAG__RECURSIVE != 0,
		},
	)

	// On les parcours
	for object := range objects {
		// Si on a eu une erreur
		if object.Err != nil {
			err = gofs.NewListError(object.Err)
			break
		}

		// Si l'on a pas un dossier
		if object.Key[len(object.Key)-1:] != "/" {
			// On ajoute a la liste
			files = append(files, gofs.File{
				Path:         path.Dir(object.Key[prefixLen:]),
				Name:         path.Base(object.Key),
				ContentType:  &object.ContentType,
				LastModified: &object.LastModified,
				Size:         object.Size,
			})
		}
	}

	return
}

func (b *S3Backend) Stat(ctx context.Context, file *gofs.File) error {
	// On initialise
	fileFullPath := b.getFullFilePathFromFile(*file)

	// On va chercher les infos
	objStat, err := b.s3c.StatObject(
		ctx,
		b.bucket,
		fileFullPath,
		minio.StatObjectOptions{},
	)

	// Si on a un probleme
	if err != nil {
		err = gofs.NewStatError(err)
	}

	// On hydrate
	file.ContentType = &objStat.ContentType
	file.LastModified = &objStat.LastModified
	file.Size = objStat.Size

	return err
}

func (b *S3Backend) Read(ctx context.Context, file gofs.File) (gofs.FileBytes, error) {
	// On utilise le stream
	fs, err := b.ReadStream(ctx, file)
	if err != nil {
		return gofs.FileBytes{}, err
	}

	return fs.ToFileBytes()
}

func (b *S3Backend) ReadStream(ctx context.Context, file gofs.File) (fs gofs.FileStream, err error) {
	// On initialise
	fileFullPath := b.getFullFilePathFromFile(file)

	// On va chercher le fichier
	object, err := _customGetObject(
		ctx,
		b.s3c,
		b.bucket,
		fileFullPath,
		minio.GetObjectOptions{},
	)
	if err != nil {
		err = gofs.NewReadError(err)
		return
	}

	// On recupere les infos du fichier
	objStat, err := object.Stat()
	if err != nil {
		err = gofs.NewReadError(gofs.NewStatError(err))
		return
	}

	// On hydrate le FileStream
	fs.Content = object
	fs.ContentType = &objStat.ContentType
	fs.LastModified = &objStat.LastModified
	fs.Size = objStat.Size
	fs.Name = file.Name
	fs.Path = file.Path

	// On renvoi tout
	return fs, err
}

func (b *S3Backend) Write(ctx context.Context, file gofs.FileBytes) error {
	// On genere le stream
	fs := gofs.FileStream{
		File:    file.File,
		Content: bytes.NewReader(file.Content),
	}

	// On utilise le Stream
	return b.WriteStream(ctx, fs)
}

func (b *S3Backend) WriteStream(ctx context.Context, file gofs.FileStream) error {
	// On initialise
	fileFullPath := b.getFullFilePathFromFile(file.File)

	// Si on a une taille null
	if file.Size == 0 {
		return gofs.NewWriteError(NewZeroFileSizedError(nil))
	}

	// Si on a rien
	var contentType string
	if file.ContentType != nil && len(*file.ContentType) > 0 {
		contentType = *file.ContentType
	} else {
		contentType = "application/octet-stream"
	}

	// On ecrit le fichier
	_, err := b.s3c.PutObject(
		ctx,
		b.bucket,
		fileFullPath,
		file.Content,
		file.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)

	if err != nil {
		err = gofs.NewWriteError(err)
	}

	return err
}

func (b *S3Backend) Move(ctx context.Context, srcFile, dstFile gofs.File) error {
	// On genere les chemins complets
	srcFullFilePath := b.getFullFilePathFromFile(srcFile)
	dstFullFilePath := b.getFullFilePathFromFile(dstFile)

	// On copie le fichier
	_, err := b.s3c.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: b.bucket,
			Object: dstFullFilePath,
		},
		minio.CopySrcOptions{
			Bucket: b.bucket,
			Object: srcFullFilePath,
		},
	)

	// Si on a pas d'erreur
	if err == nil {
		// On supprime le fichier source
		err = b.Delete(ctx, srcFile)
	}

	// Si on a une erreur (soit dans le Copy, soit dans le Delete)
	if err != nil {
		// On ajoute l'envelope
		err = gofs.NewMoveError(err)
	}

	return err
}

func (b *S3Backend) Delete(ctx context.Context, file gofs.File) error {
	// On initialise
	fileFullPath := b.getFullFilePathFromFile(file)

	// On supprime
	err := b.s3c.RemoveObject(
		ctx,
		b.bucket,
		fileFullPath,
		minio.RemoveObjectOptions{},
	)

	// Si on a une erreur
	if err != nil {
		// On ajoute l'envelope
		err = gofs.NewDeleteError(err)
	}

	return err
}

func (b *S3Backend) PresignedRead(ctx context.Context, file gofs.File, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
	// On initialise
	fileFullPath := b.getFullFilePathFromFile(file)

	// On appel le client
	u, err := b.s3c.PresignedGetObject(
		ctx,
		b.bucket,
		fileFullPath,
		expiry,
		reqParams,
	)

	// Si on a une erreur
	if err != nil {
		// On ajoute l'envelope
		err = gofs.NewPreSignedError(err)
	}

	return u, err
}

func (b *S3Backend) PresignedWrite(ctx context.Context, file gofs.File, expiry time.Duration) (*url.URL, error) {
	// On initialise
	fileFullPath := b.getFullFilePathFromFile(file)

	// On appel le client
	u, err := b.s3c.PresignedPutObject(
		ctx,
		b.bucket,
		fileFullPath,
		expiry,
	)

	// Si on a une erreur
	if err != nil {
		// On ajoute l'envelope
		err = gofs.NewPreSignedError(err)
	}

	return u, err
}
