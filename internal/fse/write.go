package fse

import (
	"io/fs"
	"io/ioutil"
)

// EnsureWriteFile performs ioutil.WriteFile but ensures that
// all directories leading up to the file will be created
func EnsureWriteFile(filename string, data []byte, perm fs.FileMode) error {
	err := EnsureFileDir(filename)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, perm)
}
