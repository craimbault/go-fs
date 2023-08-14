package gofsbcks3

func addPrefixedPath(b *S3Backend, path string) string {
	return b.Config.PathPrefix + path
}
