package fse

import (
	"os"
	"path/filepath"
	"strings"
)

// Filename returns a file's name without extensions given its filepath
func Filename(fp string) string {
	title := FilenameWithExt(fp)
	title = strings.TrimSuffix(title, filepath.Ext(fp))

	return title
}

// FilenameWithExt returns a file's name with its extensions given its filepath
func FilenameWithExt(fp string) string {
	title := strings.TrimPrefix(fp, ".")
	title = strings.TrimPrefix(title, filepath.Dir(title))
	title = strings.TrimPrefix(title, "/")
	title = strings.TrimPrefix(title, "\\")

	return title
}

// Exists returns whether a file exists on the filesystem
func Exists(fp string) bool {
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return false
	}
	return true
}
