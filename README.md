# GOFS - Storage Provider
GOFS aims to provide the same API with multiple backend storages such as Local, S3, ...


Examples
--------

A basic example to list files with an S3 Backend:
```go
config := s3.S3Config{
    Endpoint:        "localhost:9000",
    Region:          "us-east-1",
    AccessKeyID:     "minioaccesskey",
    SecretAccessKey: "miniosecretkey",
    UseSSL:          false,
    BucketName:      "gofs-test",
    PathPrefix:      "test/",
    Debug:           true,
}

log.Println("Initializing backend ...")
goFS, err := gofs.New(gofs.BACKEND_TYPE_S3, config)
if err != nil {
    log.Fatal("GOSF Backend initialization error : " + err.Error())
}
log.Println("Backend initialized")

log.Println("Listing all files recursively ...")
currentFiles, err := goFS.List("", true)
if err != nil {
    log.Fatal("Unable to list files : " + err.Error())
}
for _, filePath := range currentFiles {
    log.Printf("\t %s\n", filePath)
}
log.Println("Files listing end")
```

Other full examples are available in cmd/gosflocal & cmd/gofss3 folders

---
## TODO
### Global
- Add an in-memory cache system for small files (local and/or shared)
- Improve logging messages and error handling
- Add the ability to retreive a file partially (by bytes range)
- Add unit tests
- Add other storage backends ? (Azure Blob, GCP Storage, Swift, ...)

### Local Backend
- Add Mutex to handle simultaneous file Read/Write conflicts

---

License
-------

MIT, see [LICENSE](LICENSE)