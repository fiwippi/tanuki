package fse

import (
	"io/fs"
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

// EnsureWriteFile performs os.WriteFile but ensures that
// all directories leading up to the file will be created
func EnsureWriteFile(filename string, data []byte, perm fs.FileMode) error {
	err := EnsureFileDir(filename)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, perm)
}
