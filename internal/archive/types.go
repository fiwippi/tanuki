// Package archive implements functionality to walk through and create archives
package archive

import (
	"compress/flate"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
)

// Type defines what the archive format is
type Type int

const (
	Zip Type = iota
	Rar
)

// MimeType returns the archive's mimetype used for sending it
// over protocols like HTTP
func (at Type) MimeType() string {
	return [...]string{"application/zip", "application/x-rar"}[at]
}

func (at Type) String() string {
	return [...]string{"zip", "rar"}[at]
}

// Walker returns a walker which is used to iterate
// over each file within the archive
func (at Type) Walker() archiver.Walker {
	switch at {
	case Zip:
		a := archiver.NewZip()
		a.MkdirAll = true
		a.SelectiveCompression = true
		a.CompressionLevel = flate.BestSpeed
		return a
	case Rar:
		a := archiver.NewRar()
		a.MkdirAll = true
		return a
	}

	panic(fmt.Sprintf("invalid archive type: '%d'", at))
}

// InferType attempts to guess an archive's type from its
// filepath, if it cannot guess confidently then an error
// is returned
func InferType(fp string) (Type, error) {
	ext := filepath.Ext(fp)
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)

	switch ext {
	case "zip", "cbz":
		return Zip, nil
	case "rar", "cbr":
		return Rar, nil
	}

	return -1, fmt.Errorf("invalid archive type: '%s'", ext)
}
