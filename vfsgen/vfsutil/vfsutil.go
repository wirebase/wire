// Package vfsutil implements some I/O utility functions for http.FileSystem.
package vfsutil

import (
	"net/http"
	"os"
)

// ReadDir reads the contents of the directory associated with file and
// returns a slice of FileInfo values in directory order.
func ReadDir(fs http.FileSystem, name string) ([]os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdir(0)
}

// Stat returns the FileInfo structure describing file.
func Stat(fs http.FileSystem, name string) (os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Stat()
}
