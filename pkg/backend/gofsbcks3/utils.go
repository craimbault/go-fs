package gofsbcks3

import "gopkg.in/ini.v1"

func addPrefixedPath(b *S3Backend, path string) string {
	return b.Config.PathPrefix + path
}

func NewConfigFromIniSection(section *ini.Section) S3Config {
	return S3Config{
		Endpoint:        section.Key("endpoint").MustString("localhost:9000"),
		Region:          section.Key("region").MustString("us-east-1"),
		AccessKeyID:     section.Key("access_key").MustString("minioaccesskey"),
		SecretAccessKey: section.Key("secret_key").MustString("miniosecretkey"),
		UseSSL:          section.Key("ssl").MustBool(true),
		BucketName:      section.Key("bucket_name").MustString("gofs"),
		PathPrefix:      section.Key("path_prefix").MustString(""),
		Debug:           section.Key("debug").MustBool(false),
	}
}
