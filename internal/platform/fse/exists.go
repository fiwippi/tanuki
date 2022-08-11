package fse

import "os"

// Exists returns whether a file exists on the filesystem
func Exists(fp string) bool {
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return false
	}
	return true
}
