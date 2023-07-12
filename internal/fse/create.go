package fse

import (
	"os"
)

func CreateDirs(dirs ...string) error {
	for _, dir := range dirs {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
