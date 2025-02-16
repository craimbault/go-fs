package gofslocal

import gofs "github.com/craimbault/go-fs/pkg/gofs"

type CreateBaseDirectoryError struct {
	gofs.GofsSubError
}

func NewCreateBaseDirectoryError(previous error) gofs.GoFsError {
	return gofs.GoFsError{
		Err: CreateBaseDirectoryError{
			gofs.GofsSubError{
				Previous: previous,
				Message:  "unable to create base directory",
			},
		},
	}
}
