package fse

import (
	"path/filepath"
	"strings"
)

// Filename returns a file's name without extensions given its filepath
func Filename(fp string) string {
	title := strings.TrimPrefix(fp, ".")
	title = strings.TrimPrefix(title, filepath.Dir(title))
	title = strings.TrimPrefix(title, "/")
	title = strings.TrimPrefix(title, "\\")
	title = strings.TrimSuffix(title, filepath.Ext(fp))

	return title
}
