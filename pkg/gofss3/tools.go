package gofss3

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/craimbault/go-fs/pkg/gofs"
)

// s3://cradevcradev:cradevcradev@127.0.0.1:9000/my-prefix?region=us-east-1&use_ssl=false&bucket=mybucket
// URL : accesskey:secretkey@endpoint/path/prefix?region=xxx&use_ssl=true&bucket=yyy
func UriToOptions(uri string) (opts S3BackendOptions, err error) {
	// On parse l'URI
	var uriParts *url.URL
	uriParts, err = url.Parse(uri)
	if err != nil {
		return opts, gofs.NewUriParsingError(err)
	}

	// On traite le SSL
	useSslString := uriParts.Query().Get("use_ssl")
	if len(useSslString) > 0 {
		// On converti en bool
		opts.UseSsl, err = strconv.ParseBool(useSslString)
		if err != nil {
			return opts, gofs.NewUriParsingError(fmt.Errorf("use_ssl is not valid"))
		}
	}

	// On recupere les infos
	opts.AccessKey = uriParts.User.Username()
	opts.SecretKey, _ = uriParts.User.Password()
	opts.Endpoint = uriParts.Host
	opts.Bucket = uriParts.Query().Get("bucket")
	opts.Prefix = uriParts.Path

	return
}
