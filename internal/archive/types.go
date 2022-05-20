// Package archive implements functionality to walk through and create archives
package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v4"
)

// Type defines what the archive format is
type Type int

const (
	Zip Type = iota
	Rar

	Invalid = -1
)

// MimeType returns the archive's mimetype used for sending it
// over protocols like HTTP
func (at Type) MimeType() string {
	return [...]string{"application/zip", "application/x-rar"}[at]
}

func (at Type) String() string {
	return [...]string{"zip", "rar"}[at]
}

func (at Type) Walk(ctx context.Context, fp string, handleFile archiver.FileHandler) error {
	if at != Zip && at != Rar {
		return fmt.Errorf("invalid archive type")
	}

	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	var a archiver.Extractor
	if at == Zip {
		a = archiver.Zip{}
	} else {
		a = archiver.Rar{}
	}

	return a.Extract(ctx, f, nil, handleFile)
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

	return Invalid, fmt.Errorf("invalid archive type: '%s'", ext)
}
