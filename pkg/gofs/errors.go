package gofs

import (
	"fmt"
)

type GoFsError struct {
	Err error
}

func (e GoFsError) Error() string {
	return e.Err.Error()
}

func (e GoFsError) Unwrap() error {
	return e.Err
}

type GofsSubError struct {
	Previous error
	Message  string
}

func (e GofsSubError) Error() string {
	msg := e.Message
	if e.Previous != nil {
		msg = fmt.Sprintf("%s: %s", msg, e.Previous.Error())
	}
	return msg
}

func (e GofsSubError) Unwrap() error {
	return e.Previous
}

type UriParsingError struct {
	GofsSubError
}

func NewUriParsingError(previous error) GoFsError {
	return GoFsError{
		Err: UriParsingError{
			GofsSubError{
				Previous: previous,
				Message:  "unable to parse uri",
			},
		},
	}
}

type BackendInitError struct {
	GofsSubError
}

func NewBackendInitError(previous error) GoFsError {
	return GoFsError{
		Err: BackendInitError{
			GofsSubError{
				Previous: previous,
				Message:  "backend init error",
			},
		},
	}
}

type FileDoesNotExistsError struct {
	GofsSubError
}

func NewFileDoesNotExistsError(previous error) GoFsError {
	return GoFsError{
		Err: FileDoesNotExistsError{
			GofsSubError{
				Previous: previous,
				Message:  "file does not exists",
			},
		},
	}
}

type MoveError struct {
	GofsSubError
}

func NewMoveError(previous error) GoFsError {
	return GoFsError{
		Err: MoveError{
			GofsSubError{
				Previous: previous,
				Message:  "move file error",
			},
		},
	}
}

type MoveSrcDoesNotExistsError struct {
	GofsSubError
}

func NewMoveSrcDoesNotExistsError(previous error) GoFsError {
	return GoFsError{
		Err: MoveSrcDoesNotExistsError{
			GofsSubError{
				Previous: previous,
				Message:  "source file does not exists",
			},
		},
	}
}

type MoveDstDoesNotExistsError struct {
	GofsSubError
}

func NewMoveDstDoesNotExistsError(previous error) GoFsError {
	return GoFsError{
		Err: MoveDstDoesNotExistsError{
			GofsSubError{
				Previous: previous,
				Message:  "source file does not exists",
			},
		},
	}
}

type StatError struct {
	GofsSubError
}

func NewStatError(previous error) GoFsError {
	return GoFsError{
		Err: StatError{
			GofsSubError{
				Previous: previous,
				Message:  "unable to stat file infos",
			},
		},
	}
}

type ListError struct {
	GofsSubError
}

func NewListError(previous error) GoFsError {
	return GoFsError{
		Err: ListError{
			GofsSubError{
				Previous: previous,
				Message:  "unable to list files",
			},
		},
	}
}

type ReadError struct {
	GofsSubError
}

func NewReadError(previous error) GoFsError {
	return GoFsError{
		Err: ReadError{
			GofsSubError{
				Previous: previous,
				Message:  "unable to read file",
			},
		},
	}
}

type WriteError struct {
	GofsSubError
}

func NewWriteError(previous error) GoFsError {
	return GoFsError{
		Err: WriteError{
			GofsSubError{
				Previous: previous,
				Message:  "unable to write file",
			},
		},
	}
}

type DeleteError struct {
	GofsSubError
}

func NewDeleteError(previous error) GoFsError {
	return GoFsError{
		Err: DeleteError{
			GofsSubError{
				Previous: previous,
				Message:  "unable to delete file",
			},
		},
	}
}

type PreSignedError struct {
	GofsSubError
}

func NewPreSignedError(previous error) GoFsError {
	return GoFsError{
		Err: PreSignedError{
			GofsSubError{
				Previous: previous,
				Message:  "unable to generate presigned url",
			},
		},
	}
}
