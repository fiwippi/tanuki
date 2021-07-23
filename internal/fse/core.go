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

// DeleteFileDirIfEmpty deletes the parent directory for a
// specified filepath if it's empty
func DeleteFileDirIfEmpty(fp string) error {
	return DeleteDirIfEmpty(filepath.Dir(fp))
}

// DeleteDirIfEmpty deletes a directory if it's empty
func DeleteDirIfEmpty(fp string) error {
	files, err := os.ReadDir(fp)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return os.Remove(fp)
	}
	return nil
}
