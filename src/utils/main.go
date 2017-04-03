package utils

import (
    "strings"
    "io/ioutil"
    "log"
    "fmt"
    "time"
    "os"
    "path"
)

// Request.RemoteAddress contains port, which we want to remove i.e.:
// "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
    idx := strings.LastIndex(s, ":")
    if idx == -1 {
	return s
    }
    return s[:idx]
}

func CleanStorage(basePath string) {
    files, err := ioutil.ReadDir(basePath)
    if err != nil {
	log.Fatal(err)
    }

    for _, file := range files {
	fmt.Printf("Name: %s, IsDir: %t, ModTime: %s\n", file.Name(), file.IsDir(), file.ModTime())
	duration := time.Since(file.ModTime())

	if duration > time.Hour * 24 {
	    fullPath := path.Join(basePath, file.Name())
	    err := os.RemoveAll(fullPath)
	    if err != nil {
		log.Printf("Could not delete %s, %s\n", fullPath, err)
	    } else {
		log.Printf("Deleted: %s\n", fullPath)
	    }
	}
    }
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
	if os.IsNotExist(err) {
	    return false
	}
    }
    return true
}
