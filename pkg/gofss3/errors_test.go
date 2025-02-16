package gofss3

import (
	"errors"
	"testing"

	"github.com/craimbault/go-fs/pkg/gofs"
	"github.com/stretchr/testify/assert"
)

func TestNewNoCredentialsError(t *testing.T) {
	baseError := errors.New("base error")
	err := NewNoCredentialsError(baseError)

	assert.ErrorContains(t, err, baseError.Error(), "Error does not contains expected content")
	assert.ErrorContains(t, err, "no access-key/secret-key", "Error does not contains expected error message")
	assert.ErrorAs(t, err, &gofs.GoFsError{})
	assert.ErrorAs(t, err, &NoCredentialsError{})
}

func TestNewNoBucketError(t *testing.T) {
	baseError := errors.New("base error")
	err := NewNoBucketError(baseError)

	assert.ErrorContains(t, err, baseError.Error(), "Error does not contains expected content")
	assert.ErrorContains(t, err, "no bucket", "Error does not contains expected error message")
	assert.ErrorAs(t, err, &gofs.GoFsError{})
	assert.ErrorAs(t, err, &NoBucketError{})
}

func TestNewClientCreateFailedError(t *testing.T) {
	baseError := errors.New("base error")
	err := NewClientCreateFailedError(baseError)

	assert.ErrorContains(t, err, baseError.Error(), "Error does not contains expected content")
	assert.ErrorContains(t, err, "s3 client creation error", "Error does not contains expected error message")
	assert.ErrorAs(t, err, &gofs.GoFsError{})
	assert.ErrorAs(t, err, &ClientCreateFailedError{})
}
