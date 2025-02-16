package gofss3

import (
	"testing"

	"github.com/craimbault/go-fs/pkg/gofs"
	"github.com/stretchr/testify/assert"
)

func TestUriToOptions(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		uri := "s3://accesskey:secretkey@127.0.0.1:9000/my-prefix?region=us-east-1&use_ssl=true&bucket=mybucket"

		opts, err := UriToOptions(uri)

		assert.NoError(t, err)
		assert.Equal(t, "accesskey", opts.AccessKey)
		assert.Equal(t, "secretkey", opts.SecretKey)
		assert.Equal(t, "127.0.0.1:9000", opts.Endpoint)
		assert.Equal(t, "mybucket", opts.Bucket)
		assert.Equal(t, "/my-prefix", opts.Prefix)
		assert.Equal(t, true, opts.UseSsl)
	})

	t.Run("InvalidUri", func(t *testing.T) {
		uri := "://"

		_, err := UriToOptions(uri)

		assert.Error(t, err)
		assert.ErrorAs(t, err, &gofs.UriParsingError{})
		assert.ErrorAs(t, err, &gofs.GoFsError{})
	})

	t.Run("NoUseSsl", func(t *testing.T) {
		uri := "s3://accesskey:secretkey@127.0.0.1:9000/my-prefix?region=us-east-1&bucket=mybucket"

		opts, err := UriToOptions(uri)

		assert.NoError(t, err)
		assert.Equal(t, false, opts.UseSsl)
	})

	t.Run("UseSslTrue", func(t *testing.T) {
		uri := "s3://accesskey:secretkey@127.0.0.1:9000/my-prefix?region=us-east-1&use_ssl=true&bucket=mybucket"

		opts, err := UriToOptions(uri)

		assert.NoError(t, err)
		assert.True(t, opts.UseSsl)
	})

	t.Run("InvalidSsl", func(t *testing.T) {
		uri := "s3://accesskey:secretkey@127.0.0.1:9000/my-prefix?region=us-east-1&use_ssl=invalid&bucket=mybucket"

		_, err := UriToOptions(uri)

		assert.Error(t, err)
		assert.ErrorAs(t, err, &gofs.UriParsingError{})
		assert.ErrorAs(t, err, &gofs.GoFsError{})
	})
}
