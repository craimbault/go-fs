package gofss3

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/craimbault/go-fs/pkg/gofs"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestNewFromUri(t *testing.T) {
	var (
		bck gofs.ObjectBackend
		err error
	)

	uri := "s3://testusr:testpwd@127.0.0.1:9000/my-prefix?region=us-east-1&use_ssl=false&bucket=mybucket"

	t.Run("UriParseFailed", func(t *testing.T) {
		expectedErr := errors.New("error")
		_uriToOptions = func(uri string) (S3BackendOptions, error) {
			return S3BackendOptions{}, expectedErr
		}
		defer func() { _uriToOptions = UriToOptions }()

		bck, err = NewS3BackendFromUri(uri)

		assert.Nil(t, bck, "No backend expected")
		assert.Equal(t, expectedErr, err, "Expected error not valid")
	})

	t.Run("Success", func(t *testing.T) {
		expectedOpts := S3BackendOptions{Bucket: "my-bucket", Prefix: "my-prefix"}
		_uriToOptions = func(uri string) (S3BackendOptions, error) {
			return expectedOpts, nil
		}
		_newS3Backend = func(opts S3BackendOptions) (*S3Backend, error) {
			assert.Equal(t, expectedOpts, opts, "Expected options in NewS3Backend not valid")
			return &S3Backend{bucket: opts.Bucket, prefix: opts.Prefix}, nil
		}
		defer func() {
			_uriToOptions = UriToOptions
			_newS3Backend = NewS3Backend
		}()

		bck, err = NewS3BackendFromUri(uri)

		assert.NotNil(t, bck, "Backend expected and not provided")
		assert.NoError(t, err, "No error expected")
	})

}

func TestNew(t *testing.T) {
	s3Opts := S3BackendOptions{
		AccessKey: "test",
		SecretKey: "test",
		Bucket:    "test-bucket",
		Endpoint:  "test.test",
		Region:    "test",
		UseSsl:    true,
	}

	t.Run("NoBucket", func(t *testing.T) {
		_, err := NewS3Backend(S3BackendOptions{})

		assert.ErrorAs(t, err, &gofs.GoFsError{}, "Returned error does not implement expected parent GoFsError error")
		assert.ErrorAs(t, err, &gofs.BackendInitError{}, "Returned error does not implement expected parent BackendInitError error")
		assert.ErrorAs(t, err, &NoBucketError{}, "Returned error does not implement expected error")
	})

	t.Run("NoAccessKey", func(t *testing.T) {
		_, err := NewS3Backend(S3BackendOptions{
			Bucket: "test-bucket",
		})

		assert.ErrorAs(t, err, &gofs.GoFsError{}, "Returned error does not implement expected parent GoFsError error")
		assert.ErrorAs(t, err, &gofs.BackendInitError{}, "Returned error does not implement expected parent BackendInitError error")
		assert.ErrorAs(t, err, &NoCredentialsError{}, "Returned error does not implement expected error")
	})

	t.Run("NoSecretKey", func(t *testing.T) {
		_, err := NewS3Backend(S3BackendOptions{
			Bucket: "test-bucket",
		})

		assert.ErrorAs(t, err, &gofs.GoFsError{}, "Returned error does not implement expected parent GoFsError error")
		assert.ErrorAs(t, err, &gofs.BackendInitError{}, "Returned error does not implement expected parent BackendInitError error")
		assert.ErrorAs(t, err, &NoCredentialsError{}, "Returned error does not implement expected error")
	})

	t.Run("NewError", func(t *testing.T) {
		expectedError := errors.New("new error")

		_minioNew = func(endpoint string, opts *minio.Options) (*minio.Client, error) {
			assert.Equal(t, s3Opts.Endpoint, endpoint, "Endpoint not the expected one")
			assert.Equal(t, s3Opts.Region, opts.Region, "Region not the expected one")
			assert.Equal(t, s3Opts.UseSsl, opts.Secure, "UseSsl not the expected one")

			creds, err := opts.Creds.Get()
			assert.NoError(t, err, "Get Creds from minio options should not throw an error")

			assert.Equal(t, s3Opts.AccessKey, creds.AccessKeyID, "AccessKey not the expected one")
			assert.Equal(t, s3Opts.SecretKey, creds.SecretAccessKey, "SecretKey not the expected one")

			return nil, expectedError
		}
		defer func() { _minioNew = minio.New }()

		_, err := NewS3Backend(s3Opts)

		assert.ErrorAs(t, err, &gofs.GoFsError{}, "Returned error does not implement expected parent GoFsError error")
		assert.ErrorAs(t, err, &gofs.BackendInitError{}, "Returned error does not implement expected parent BackendInitError error")
		assert.ErrorAs(t, err, &ClientCreateFailedError{}, "Returned error does not implement expected error")
	})

	t.Run("New", func(t *testing.T) {
		s3c, err := minio.New(s3Opts.Endpoint, &minio.Options{})
		assert.NoError(t, err, "Creating a real Minio S3 client should not throw an error")

		_minioNew = func(endpoint string, opts *minio.Options) (*minio.Client, error) {
			assert.Equal(t, s3Opts.Endpoint, endpoint, "Endpoint not the expected one")
			assert.Equal(t, s3Opts.Region, opts.Region, "Region not the expected one")
			assert.Equal(t, s3Opts.UseSsl, opts.Secure, "UseSsl not the expected one")

			creds, err := opts.Creds.Get()
			assert.NoError(t, err, "Get Creds from minio options should not throw an error")

			assert.Equal(t, s3Opts.AccessKey, creds.AccessKeyID, "AccessKey not the expected one")
			assert.Equal(t, s3Opts.SecretKey, creds.SecretAccessKey, "SecretKey not the expected one")

			return s3c, nil
		}
		defer func() { _minioNew = minio.New }()

		b, err := NewS3Backend(s3Opts)

		assert.NoError(t, err, "Sould not throw an error")
		assert.Equal(t, s3c, b.s3c, "S3Client is not the expected one")
	})
}

func TestGetFullFilePath(t *testing.T) {
	file := gofs.File{
		Path: "path",
		Name: "test.test",
	}
	myPrefix := "test"
	c := S3Backend{
		prefix: myPrefix,
	}

	expectedFfp := path.Join(myPrefix, file.Path, file.Name)
	ffp := c.getFullFilePathFromFile(file)

	assert.Equal(t, expectedFfp, ffp, "Full FilePath not valid")
}

func TestList(t *testing.T) {
	s3c := new(mockS3Client)
	ctx := context.Background()
	b := S3Backend{
		bucket: "test-bucket",
		s3c:    s3c,
		prefix: "test",
	}
	dirPath := "test-path"
	contentType := "content-type-test"
	lastModified := time.Now()
	size := int64(123456789)
	fileName := "file.test"

	t.Run("ListError", func(t *testing.T) {
		dirPath := "test-path"
		flag := gofs.Flag(gofs.FLAG__RECURSIVE)
		expectedErr := errors.New("list error")

		loo := minio.ListObjectsOptions{
			Prefix:    path.Join(b.prefix, dirPath),
			Recursive: true,
		}

		s3c.On("ListObjects", ctx, b.bucket, loo).
			Return(minio.ObjectInfo{Err: expectedErr}).Once()

		_, err := b.List(ctx, dirPath, flag)

		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the requied content")
		assert.ErrorAs(t, err, &gofs.ListError{}, "Returned error does not implement ListError")
	})

	t.Run("ListSuccessRecursiveWithPrefix", func(t *testing.T) {
		objKey := path.Join(b.prefix, dirPath, fileName)
		flag := gofs.Flag(gofs.FLAG__RECURSIVE)

		loo := minio.ListObjectsOptions{
			Prefix:    path.Join(b.prefix, dirPath),
			Recursive: true,
		}

		s3c.On("ListObjects", ctx, b.bucket, loo).
			Return(minio.ObjectInfo{
				ContentType:  contentType,
				LastModified: lastModified,
				Size:         size,
				Key:          objKey,
			}).Once()

		files, err := b.List(ctx, dirPath, flag)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, 1, len(files), "Single file expected")
		assert.Equal(t, size, files[0].Size, "Size not valid")
		assert.Equal(t, lastModified, *files[0].LastModified, "LastModified not valid")
		assert.Equal(t, contentType, *files[0].ContentType, "LastModified not valid")
		assert.Equal(t, fileName, files[0].Name, "Name not valid")
		assert.Equal(t, dirPath, files[0].Path, "Path not valid")
	})

	t.Run("ListSuccessNoRecursive", func(t *testing.T) {
		objKey := path.Join(b.prefix, dirPath, fileName)
		flag := gofs.Flag(0)

		loo := minio.ListObjectsOptions{
			Prefix:    path.Join(b.prefix, dirPath),
			Recursive: false,
		}

		s3c.On("ListObjects", ctx, b.bucket, loo).
			Return(minio.ObjectInfo{
				ContentType:  contentType,
				LastModified: lastModified,
				Size:         size,
				Key:          objKey,
			}).Once()

		_, _ = b.List(ctx, dirPath, flag)
	})

	t.Run("ListSuccessWithoutPrefix", func(t *testing.T) {
		previousPrefix := b.prefix
		b.prefix = ""
		defer func() { b.prefix = previousPrefix }()

		objKey := path.Join(b.prefix, dirPath, fileName)
		flag := gofs.Flag(0)

		loo := minio.ListObjectsOptions{
			Prefix:    path.Join(b.prefix, dirPath),
			Recursive: false,
		}

		s3c.On("ListObjects", ctx, b.bucket, loo).
			Return(minio.ObjectInfo{
				ContentType:  contentType,
				LastModified: lastModified,
				Size:         size,
				Key:          objKey,
			}).Once()

		files, err := b.List(ctx, dirPath, flag)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, fileName, files[0].Name, "Name not valid")
		assert.Equal(t, dirPath, files[0].Path, "Path not valid")
	})
}

func TestStat(t *testing.T) {
	s3c := new(mockS3Client)
	ctx := context.Background()
	b := S3Backend{
		bucket: "test-bucket",
		s3c:    s3c,
		prefix: "test",
	}
	dirPath := "test-path"
	objInfo := minio.ObjectInfo{
		ContentType:  "content-type-test",
		LastModified: time.Now(),
		Size:         123456789,
	}
	fileName := "file.test"

	t.Run("Success", func(t *testing.T) {
		s3c.On("StatObject",
			ctx, b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			minio.StatObjectOptions{},
		).Return(objInfo, nil).Once()

		file := gofs.File{
			Name: fileName,
			Path: dirPath,
		}

		err := b.Stat(ctx, &file)

		assert.NoError(t, err, "No error expected")
	})

	t.Run("Failed", func(t *testing.T) {
		expectedErr := errors.New("stat error details")

		s3c.On("StatObject",
			ctx, b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			minio.StatObjectOptions{},
		).Return(minio.ObjectInfo{}, expectedErr).Once()

		file := gofs.File{
			Name: fileName,
			Path: dirPath,
		}

		err := b.Stat(ctx, &file)

		assert.Error(t, err, "An error is expected")
		assert.ErrorAs(t, err, &gofs.GoFsError{}, "The error must have a GoFsError parent")
		assert.ErrorAs(t, err, &gofs.StatError{}, "The error must have a StatError parent")
	})
}

func TestReadStreamAndRead(t *testing.T) {
	// Read appel ReadStream donc on test d'abord ReadStream puis Read
	s3c := new(mockS3Client)
	ctx := context.Background()
	b := S3Backend{
		bucket: "test-bucket",
		s3c:    s3c,
		prefix: "test",
	}
	dirPath := "test-path"
	contentBytes := []byte("test")
	objInfo := minio.ObjectInfo{
		ContentType:  "content-type-test",
		LastModified: time.Now(),
		Size:         int64(len(contentBytes)),
	}
	fileName := "file.test"

	t.Run("ReadStreamGetObjectFailed", func(t *testing.T) {
		expectedErr := errors.New("failed to get object")
		defer func() { _customGetObject = minioGetObbject }()
		_customGetObject = func(_ctx context.Context, _client s3client, _bucketName, _path string, _ minio.GetObjectOptions) (s3BackendObject, error) {
			assert.Equal(t, ctx, _ctx, "Context is not the expected one")
			assert.Equal(t, s3c, _client, "S3Client is not the expected one")
			assert.Equal(t, b.bucket, _bucketName, "BucketName is not the expected one")
			assert.Equal(t, path.Join(b.prefix, dirPath, fileName), _path, "Path is not the expected one")

			return nil, expectedErr
		}

		file := gofs.File{
			Name: fileName,
			Path: dirPath,
		}

		_, err := b.ReadStream(ctx, file)

		assert.Error(t, err, "An error is expected")
		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the throwed error")
		assert.ErrorAs(t, err, &gofs.ReadError{}, "Returned error should have ReadError as parent")
	})

	t.Run("ReadStreamStatObjectFailed", func(t *testing.T) {
		expectedErr := errors.New("failed to stat object")
		customObj := new(mockS3BackendObject)
		defer func() { _customGetObject = minioGetObbject }()
		_customGetObject = func(_ctx context.Context, _client s3client, _bucketName, _path string, _ minio.GetObjectOptions) (s3BackendObject, error) {
			assert.Equal(t, ctx, _ctx, "Context is not the expected one")
			assert.Equal(t, s3c, _client, "S3Client is not the expected one")
			assert.Equal(t, b.bucket, _bucketName, "BucketName is not the expected one")
			assert.Equal(t, path.Join(b.prefix, dirPath, fileName), _path, "Path is not the expected one")

			return customObj, nil
		}

		file := gofs.File{
			Name: fileName,
			Path: dirPath,
		}

		customObj.On("Stat").Return(minio.ObjectInfo{}, expectedErr).Once()

		_, err := b.ReadStream(ctx, file)

		assert.Error(t, err, "An error is expected")
		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the throwed error")
		assert.ErrorAs(t, err, &gofs.ReadError{}, "Returned error should have ReadError as parent")
		assert.ErrorAs(t, err, &gofs.StatError{}, "Returned error should have StatError as parent")
	})

	t.Run("ReadStreamSuccess", func(t *testing.T) {
		customObj := new(mockS3BackendObject)
		defer func() { _customGetObject = minioGetObbject }()
		_customGetObject = func(_ctx context.Context, _client s3client, _bucketName, _path string, _ minio.GetObjectOptions) (s3BackendObject, error) {
			assert.Equal(t, ctx, _ctx, "Context is not the expected one")
			assert.Equal(t, s3c, _client, "S3Client is not the expected one")
			assert.Equal(t, b.bucket, _bucketName, "BucketName is not the expected one")
			assert.Equal(t, path.Join(b.prefix, dirPath, fileName), _path, "Path is not the expected one")

			return customObj, nil
		}

		file := gofs.File{
			Name: fileName,
			Path: dirPath,
		}

		customObj.On("Stat").Return(objInfo, nil).Once()

		fs, err := b.ReadStream(ctx, file)

		assert.NoError(t, err, "Should not have any error")
		assert.Equal(t, customObj, fs.Content, "Content is not the expected one")
		assert.Equal(t, objInfo.ContentType, *fs.ContentType, "ContentType is not the expected one")
		assert.Equal(t, objInfo.LastModified, *fs.LastModified, "LastModified is not the expected one")
		assert.Equal(t, objInfo.Size, fs.Size, "Size is not the expected one")
		assert.Equal(t, file.Name, fs.Name, "Name is not the expected one")
		assert.Equal(t, file.Path, fs.Path, "Path is not the expected one")
	})

	t.Run("ReadFailed", func(t *testing.T) {
		expectedErr := errors.New("failed to get object")
		defer func() { _customGetObject = minioGetObbject }()
		_customGetObject = func(_ctx context.Context, _client s3client, _bucketName, _path string, _ minio.GetObjectOptions) (s3BackendObject, error) {
			return nil, expectedErr
		}

		file := gofs.File{
			Name: fileName,
			Path: dirPath,
		}

		_, err := b.Read(ctx, file)

		// On ne test que le fait d'avoir une erreur car
		//  c'est le ReadStream qui est appele dans le Read
		//  et qui est deja teste au dessus
		assert.Error(t, err, "An error is expected")
	})

	t.Run("ReadSuccess", func(t *testing.T) {
		customObj := customMockS3BackendObject{
			obj:     bytes.NewReader(contentBytes),
			objInfo: objInfo,
		}
		defer func() { _customGetObject = minioGetObbject }()
		_customGetObject = func(_ context.Context, _ s3client, _, _ string, _ minio.GetObjectOptions) (s3BackendObject, error) {
			return customObj, nil
		}

		file := gofs.File{
			Name: fileName,
			Path: dirPath,
		}

		fb, err := b.Read(ctx, file)

		assert.NoError(t, err, "Should not have any error")
		assert.Equal(t, contentBytes, fb.Content, "Content is not the expected one")
		assert.Equal(t, objInfo.ContentType, *fb.ContentType, "ContentType is not the expected one")
		assert.Equal(t, objInfo.LastModified, *fb.LastModified, "LastModified is not the expected one")
		assert.Equal(t, objInfo.Size, fb.Size, "Size is not the expected one")
		assert.Equal(t, file.Name, fb.Name, "Name is not the expected one")
		assert.Equal(t, file.Path, fb.Path, "Path is not the expected one")
	})
}

func TestWriteStreamAndWrite(t *testing.T) {
	// Write appel WriteStream donc on test d'abord WriteStream puis Write
	s3c := new(mockS3Client)
	ctx := context.Background()
	b := S3Backend{
		bucket: "test-bucket",
		s3c:    s3c,
		prefix: "test",
	}
	dirPath := "test-path"
	contentBytes := []byte("test")
	fileName := "file.json"

	t.Run("ZeroFileSized", func(t *testing.T) {
		file := gofs.NewFileStream(
			dirPath, fileName,
			0,
			bytes.NewReader(contentBytes),
		)

		err := b.WriteStream(ctx, file)

		assert.Error(t, err, "An error is expected")
		assert.ErrorAs(t, err, &gofs.WriteError{}, "Returned error should have WriteError as parent")
		assert.ErrorAs(t, err, &ZeroFileSizedError{}, "Returned error should have ZeroFileSized as parent")
	})

	t.Run("NoContentType", func(t *testing.T) {
		file := gofs.NewFileStream(
			dirPath, fileName,
			int64(len(contentBytes)),
			bytes.NewReader(contentBytes),
		)
		file.ContentType = nil

		s3c.On("PutObject",
			ctx, b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			file.Content, file.Size,
			minio.PutObjectOptions{
				ContentType: "application/octet-stream",
			},
		).Return(minio.UploadInfo{}, nil).Once()

		b.WriteStream(ctx, file)
	})

	t.Run("WithContentType", func(t *testing.T) {
		contentTypeForced := "my-content-type"
		file := gofs.NewFileStream(
			dirPath, fileName,
			0,
			bytes.NewReader(contentBytes),
		)
		file.ContentType = &contentTypeForced

		s3c.On("PutObject",
			ctx, b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			file.Content, file.Size,
			minio.PutObjectOptions{
				ContentType: contentTypeForced,
			},
		).Return(minio.UploadInfo{}, nil).Once()

		b.WriteStream(ctx, file)
	})

	t.Run("WriteStreamPutObjectFailed", func(t *testing.T) {
		expectedErr := errors.New("failed to write object")

		file := gofs.NewFileStream(
			dirPath, fileName,
			int64(len(contentBytes)),
			bytes.NewReader(contentBytes),
		)

		s3c.On("PutObject",
			ctx, b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			file.Content, file.Size,
			minio.PutObjectOptions{
				ContentType: *file.ContentType,
			},
		).Return(minio.UploadInfo{}, expectedErr).Once()

		err := b.WriteStream(ctx, file)

		assert.Error(t, err, "An error is expected")
		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the throwed error")
		assert.ErrorAs(t, err, &gofs.WriteError{}, "Returned error should have WriteError as parent")
	})

	t.Run("WriteStreamSuccess", func(t *testing.T) {
		file := gofs.NewFileStream(
			dirPath, fileName,
			int64(len(contentBytes)),
			bytes.NewReader(contentBytes),
		)

		s3c.On("PutObject",
			ctx, b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			file.Content, file.Size,
			minio.PutObjectOptions{
				ContentType: *file.ContentType,
			},
		).Return(minio.UploadInfo{}, nil).Once()

		err := b.WriteStream(ctx, file)

		assert.NoError(t, err, "Should not have any error")
	})

	t.Run("WriteFailed", func(t *testing.T) {
		expectedErr := errors.New("failed to get object")

		file := gofs.NewFileBytes(
			dirPath, fileName,
			contentBytes,
		)

		s3c.On("PutObject",
			ctx, b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			bytes.NewReader(contentBytes), file.Size,
			minio.PutObjectOptions{
				ContentType: *file.ContentType,
			},
		).Return(minio.UploadInfo{}, expectedErr).Once()

		err := b.Write(ctx, file)

		// On ne test que le fait d'avoir une erreur car
		//  c'est le WriteStream qui est appele dans le Write
		//  et qui est deja teste au dessus
		assert.Error(t, err, "An error is expected")
	})

	t.Run("WriteSuccess", func(t *testing.T) {
		file := gofs.NewFileBytes(
			dirPath, fileName,
			contentBytes,
		)

		s3c.On("PutObject",
			ctx, b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			bytes.NewReader(contentBytes), file.Size,
			minio.PutObjectOptions{
				ContentType: *file.ContentType,
			},
		).Return(minio.UploadInfo{}, nil).Once()

		err := b.Write(ctx, file)

		assert.NoError(t, err, "Should not have any error")
	})
}

func TestDelete(t *testing.T) {
	s3c := new(mockS3Client)
	ctx := context.Background()
	b := S3Backend{
		bucket: "test-bucket",
		s3c:    s3c,
		prefix: "test",
	}
	dirPath := "test-path"
	fileName := "file.json"

	t.Run("Failed", func(t *testing.T) {
		expectedErr := errors.New("delete file failed")

		s3c.On("RemoveObject", ctx,
			b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			minio.RemoveObjectOptions{},
		).Return(expectedErr).Once()

		file := gofs.NewFile(dirPath, fileName, 0, nil)

		err := b.Delete(ctx, file)

		assert.Error(t, err, "An error is expected")
		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the throwed error")
		assert.ErrorAs(t, err, &gofs.DeleteError{}, "Returned error should have DeleteError as parent")
	})

	t.Run("Success", func(t *testing.T) {
		s3c.On("RemoveObject", ctx,
			b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			minio.RemoveObjectOptions{},
		).Return(nil).Once()

		file := gofs.NewFile(dirPath, fileName, 0, nil)

		err := b.Delete(ctx, file)

		assert.NoError(t, err, "No error expected")
	})
}

func TestMove(t *testing.T) {
	// Write appel WriteStream donc on test d'abord WriteStream puis Write
	s3c := new(mockS3Client)
	ctx := context.Background()
	b := S3Backend{
		bucket: "test-bucket",
		s3c:    s3c,
		prefix: "test",
	}
	dirPath := "test-path"
	srcFileName := "src.json"
	dstFileName := "dst.json"
	srcFile := gofs.NewFile(dirPath, srcFileName, 0, nil)
	dstFile := gofs.NewFile(dirPath, dstFileName, 0, nil)

	t.Run("MoveFailed", func(t *testing.T) {
		expectedErr := errors.New("move file failed")

		s3c.On("CopyObject", ctx,
			minio.CopyDestOptions{
				Bucket: b.bucket,
				Object: path.Join(b.prefix, dirPath, dstFileName),
			},
			minio.CopySrcOptions{
				Bucket: b.bucket,
				Object: path.Join(b.prefix, dirPath, srcFileName),
			},
		).Return(minio.UploadInfo{}, expectedErr).Once()

		err := b.Move(ctx, srcFile, dstFile)

		assert.Error(t, err, "An error is expected")
		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the throwed error")
		assert.ErrorAs(t, err, &gofs.MoveError{}, "Returned error should have WriteError as parent")
	})

	t.Run("MoveSuccessDeleteFailed", func(t *testing.T) {
		expectedErr := errors.New("delete file failed")

		s3c.On("CopyObject", ctx,
			minio.CopyDestOptions{
				Bucket: b.bucket,
				Object: path.Join(b.prefix, dirPath, dstFileName),
			},
			minio.CopySrcOptions{
				Bucket: b.bucket,
				Object: path.Join(b.prefix, dirPath, srcFileName),
			},
		).Return(minio.UploadInfo{}, nil).Once()

		s3c.On("RemoveObject", ctx,
			b.bucket,
			path.Join(b.prefix, dirPath, srcFileName),
			minio.RemoveObjectOptions{},
		).Return(expectedErr).Once()

		err := b.Move(ctx, srcFile, dstFile)

		assert.Error(t, err, "An error is expected")
		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the throwed error")
		assert.ErrorAs(t, err, &gofs.MoveError{}, "Returned error should have MoveError as parent")
		assert.ErrorAs(t, err, &gofs.DeleteError{}, "Returned error should have DeleteError as parent")
	})

	t.Run("Success", func(t *testing.T) {
		s3c.On("CopyObject", ctx,
			minio.CopyDestOptions{
				Bucket: b.bucket,
				Object: path.Join(b.prefix, dirPath, dstFileName),
			},
			minio.CopySrcOptions{
				Bucket: b.bucket,
				Object: path.Join(b.prefix, dirPath, srcFileName),
			},
		).Return(minio.UploadInfo{}, nil).Once()

		s3c.On("RemoveObject", ctx,
			b.bucket,
			path.Join(b.prefix, dirPath, srcFileName),
			minio.RemoveObjectOptions{},
		).Return(nil).Once()

		err := b.Move(ctx, srcFile, dstFile)

		assert.NoError(t, err, "No error expected")
	})
}

func TestPresignedRead(t *testing.T) {
	s3c := new(mockS3Client)
	ctx := context.Background()
	b := S3Backend{
		bucket: "test-bucket",
		s3c:    s3c,
		prefix: "test",
	}
	dirPath := "test-path"
	fileName := "file.json"
	expiry := time.Duration(time.Hour * 5)
	reqParams := url.Values{}
	reqParams.Add("test-key", "test-value")

	t.Run("Failed", func(t *testing.T) {
		returnedUrl := url.URL{}
		expectedErr := errors.New("presigned failed")

		s3c.On("PresignedGetObject", ctx,
			b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			expiry,
			reqParams,
		).Return(&returnedUrl, expectedErr).Once()

		file := gofs.NewFile(dirPath, fileName, 0, nil)

		_, err := b.PresignedRead(ctx, file, expiry, reqParams)

		assert.Error(t, err, "An error is expected")
		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the throwed error")
		assert.ErrorAs(t, err, &gofs.PreSignedError{}, "Returned error should have PresignedError as parent")
	})

	t.Run("Success", func(t *testing.T) {
		returnedUrl := url.URL{
			Scheme: "https",
			Host:   "demo.demo.demo",
			Path:   "/demo-test",
		}

		s3c.On("PresignedGetObject", ctx,
			b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			expiry,
			reqParams,
		).Return(&returnedUrl, nil).Once()

		file := gofs.NewFile(dirPath, fileName, 0, nil)

		u, err := b.PresignedRead(ctx, file, expiry, reqParams)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, returnedUrl, *u, "Expected Url is not valid")
	})
}

func TestPresignedWrite(t *testing.T) {
	s3c := new(mockS3Client)
	ctx := context.Background()
	b := S3Backend{
		bucket: "test-bucket",
		s3c:    s3c,
		prefix: "test",
	}
	dirPath := "test-path"
	fileName := "file.json"
	expiry := time.Duration(time.Hour * 5)
	reqParams := url.Values{}
	reqParams.Add("test-key", "test-value")

	t.Run("Failed", func(t *testing.T) {
		returnedUrl := url.URL{}
		expectedErr := errors.New("presigned failed")

		s3c.On("PresignedPutObject", ctx,
			b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			expiry,
		).Return(&returnedUrl, expectedErr).Once()

		file := gofs.NewFile(dirPath, fileName, 0, nil)

		_, err := b.PresignedWrite(ctx, file, expiry)

		assert.Error(t, err, "An error is expected")
		assert.ErrorContains(t, err, expectedErr.Error(), "Returned error does not contains the throwed error")
		assert.ErrorAs(t, err, &gofs.PreSignedError{}, "Returned error should have PresignedError as parent")
	})

	t.Run("Success", func(t *testing.T) {
		returnedUrl := url.URL{
			Scheme: "https",
			Host:   "demo.demo.demo",
			Path:   "/demo-test",
		}

		s3c.On("PresignedPutObject", ctx,
			b.bucket,
			path.Join(b.prefix, dirPath, fileName),
			expiry,
		).Return(&returnedUrl, nil).Once()

		file := gofs.NewFile(dirPath, fileName, 0, nil)

		u, err := b.PresignedWrite(ctx, file, expiry)

		assert.NoError(t, err, "No error expected")
		assert.Equal(t, returnedUrl, *u, "Expected Url is not valid")
	})
}
