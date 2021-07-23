package fse

import (
	"io/fs"
	"os"
)

// EnsureWriteFile performs os.WriteFile but ensures that
// all directories leading up to the file will be created
func EnsureWriteFile(filename string, data []byte, perm fs.FileMode) error {
	err := EnsureFileDir(filename)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, perm)
}
