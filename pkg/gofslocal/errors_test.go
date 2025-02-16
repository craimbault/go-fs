package gofslocal

import (
	"errors"
	"testing"

	"github.com/craimbault/go-fs/pkg/gofs"
	"github.com/stretchr/testify/assert"
)

func TestNewCreateBaseDirectoryError(t *testing.T) {
	baseError := errors.New("base error")
	err := NewCreateBaseDirectoryError(baseError)

	assert.ErrorContains(t, err, baseError.Error(), "Error does not contains expected content")
	assert.ErrorContains(t, err, "unable to create base directory", "Error does not contains expected content")
	assert.ErrorAs(t, err, &gofs.GoFsError{})
	assert.ErrorAs(t, err, &CreateBaseDirectoryError{})
}
