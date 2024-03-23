package main

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include <string.h>
// #include "ret.h"
import "C"
import (
	"log"
	"strconv"
	"unsafe"

	gofs "github.com/craimbault/go-fs"
	"github.com/craimbault/go-fs/pkg/backend/gofsbcklocal"
	"github.com/rs/zerolog"
)

func main() {}

func gofs_init() (gofs.GoFS, error) {
	backendType := gofs.BACKEND_TYPE_LOCAL
	// On initialise la config
	config := gofsbcklocal.LocalConfig{
		BasePath: "/tmp/gosf-data",
		Debug:    true,
	}

	// On indique le niveau de log
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if config.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Println("Initializing backend ...")

	// On initalise le backend
	goFS, err := gofs.New(backendType, config)
	if err != nil {
		log.Fatal("GOSF Backend initialization error : " + err.Error())
	}
	log.Println("Backend initialized")

	// On renvoi les infos
	return goFS, err
}

//export gofs_list
func gofs_list(prefix *C.char) *C.struct__RET {
	// Init Backend
	goFS, _ := gofs_init()

	log.Println("Listing all files recursively ...")
	currentFiles, err := goFS.List(C.GoString(prefix), true)
	if err != nil {
		log.Println("Unable to list files : " + err.Error())
		return nil
	}

	ret := C.RET_new(C.uint(0), C.int(0))

	for _, filePath := range currentFiles {
		log.Printf("\t %s\n", filePath)
		ptr := C.CString(filePath)
		C.RET_add(ret, ptr)
		C.free(unsafe.Pointer(ptr))
	}
	//for i := 0; i < int(C.RET_count(ret)); i++ {
	//  log.Printf(">%d:%s\n", i+1, C.GoString(C.RET_get(ret, C.int(i))))
	//}
	log.Println("Files listing end")

	return ret
}

//export gofs_read
func gofs_read(file *C.char) *C.struct__RET {
	// Init Backend
	goFS, _ := gofs_init()

	content, err := goFS.ReadString(C.GoString(file))
	if err != nil {
		log.Println("Unable to read file : " + err.Error())
		ret := C.RET_new(C.uint(1), C.int(1))
		ptr := C.CString(err.Error())
		C.RET_add(ret, ptr)
		C.free(unsafe.Pointer(ptr))
		return ret
	}
	log.Printf("\t File %s read with success\n", C.GoString(file))

	ret := C.RET_new(C.uint(1), C.int(0))
	ptr := C.CString(content)
	C.RET_add2(ret, ptr, C.uint(len(content)))
	C.free(unsafe.Pointer(ptr))

	return ret
}

//export gofs_write
func gofs_write(file *C.char, content *C.char) C.int {
	// Init Backend
	goFS, _ := gofs_init()

	err := goFS.WriteString(C.GoString(file), C.GoString(content))
	if err != nil {
		log.Println("Unable to write file : " + err.Error())
		return 1
	}
	log.Printf("\t File %s written with success\n", C.GoString(file))

	return 0
}

//export gofs_delete
func gofs_delete(file *C.char) C.int {
	// Init Backend
	goFS, _ := gofs_init()

	err := goFS.Delete(C.GoString(file))
	if err != nil {
		log.Println("Unable to delete file : " + err.Error())
		return 1
	}
	log.Printf("\t File %s deleted with success\n", C.GoString(file))

	return 0
}

//export gofs_move
func gofs_move(oldfile *C.char, newfile *C.char) C.int {
	// Init Backend
	goFS, _ := gofs_init()

	err := goFS.Move(C.GoString(oldfile), C.GoString(newfile))
	if err != nil {
		log.Println("Unable to move file : " + err.Error())
		return 1
	}
	log.Printf("\t File %s moved into %s with success\n", C.GoString(oldfile), C.GoString(newfile))

	return 0
}

//export gofs_stat
func gofs_stat(file *C.char) *C.struct__RET {
	// Init Backend
	goFS, _ := gofs_init()

	log.Println("Getting stats of the file ...")
	fileStat, err := goFS.Stat(C.GoString(file))
	if err != nil {
		log.Println("Unable to stat file : " + err.Error())
		ret := C.RET_new(C.uint(1), C.int(1))
		C.RET_add(ret, file)
		ptr := C.CString(err.Error())
		C.RET_add(ret, ptr)
		C.free(unsafe.Pointer(ptr))
		return ret
	}

	ret := C.RET_new(C.uint(5), C.int(0))

	C.RET_add(ret, file)

	var ptr *C.char

	ptr = C.CString(string(fileStat.LastModified.Format("02/01/2006 15:04:05")))
	C.RET_add(ret, ptr)
	C.free(unsafe.Pointer(ptr))

	ptr = C.CString(fileStat.ETag)
	C.RET_add(ret, ptr)
	C.free(unsafe.Pointer(ptr))

	ptr = C.CString(fileStat.ContentType)
	C.RET_add(ret, ptr)
	C.free(unsafe.Pointer(ptr))

	ptr = C.CString(strconv.FormatInt(fileStat.Size, 10))
	C.RET_add(ret, ptr)
	C.free(unsafe.Pointer(ptr))

	log.Println("File stat end")

	return ret
}
