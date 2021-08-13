// Package fse provides additional filesystem functions which extend
// the capability of stdlib
package fse

import (
	"os"
	"path/filepath"
)

// DeleteFileDirIfEmpty deletes the parent directory for a
// specified filepath if it's empty
func DeleteFileDirIfEmpty(fp string) error {
	return DeleteDirIfEmpty(filepath.Dir(fp))
}

// DeleteDirIfEmpty deletes a directory if it's empty
func DeleteDirIfEmpty(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return os.Remove(dir)
	}
	return nil
}
