package gofss3

import (
	"github.com/craimbault/go-fs/pkg/gofs"
)

type NoCredentialsError struct {
	gofs.GofsSubError
}

func NewNoCredentialsError(previous error) gofs.GoFsError {
	return gofs.GoFsError{
		Err: NoCredentialsError{
			gofs.GofsSubError{
				Previous: previous,
				Message:  "no access-key/secret-key",
			},
		},
	}
}

type NoBucketError struct {
	gofs.GofsSubError
}

func NewNoBucketError(previous error) gofs.GoFsError {
	return gofs.GoFsError{
		Err: NoBucketError{
			gofs.GofsSubError{
				Previous: previous,
				Message:  "no bucket",
			},
		},
	}
}

type ClientCreateFailedError struct {
	gofs.GofsSubError
}

func NewClientCreateFailedError(previous error) gofs.GoFsError {
	return gofs.GoFsError{
		Err: ClientCreateFailedError{
			gofs.GofsSubError{
				Previous: previous,
				Message:  "s3 client creation error",
			},
		},
	}
}

type ZeroFileSizedError struct {
	gofs.GofsSubError
}

func NewZeroFileSizedError(previous error) gofs.GoFsError {
	return gofs.GoFsError{
		Err: ZeroFileSizedError{
			gofs.GofsSubError{
				Previous: previous,
				Message:  "zero sized file profived",
			},
		},
	}
}
