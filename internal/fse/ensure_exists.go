package fse

import (
	"os"
	"path/filepath"
)

// EnsureDir ensures a directory exists on the filesystem
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

// EnsureFileDir ensures the parent directory of the file
// exists on the filesystem
func EnsureFileDir(fp string) error {
	return EnsureDir(filepath.Dir(fp))
}
