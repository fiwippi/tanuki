// Package archive implements functionality to walk through and create archives
package archive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

func (at Type) extractor(fp string) (archiver.Extractor, *os.File, error) {
	if at != Zip && at != Rar {
		return nil, nil, fmt.Errorf("invalid archive type")
	}

	f, err := os.Open(fp)
	if err != nil {
		return nil, nil, err
	}

	var e archiver.Extractor
	if at == Zip {
		e = archiver.Zip{}
	} else {
		e = archiver.Rar{}
	}

	return e, f, nil
}

func (at Type) Walk(ctx context.Context, fp string, handleFile archiver.FileHandler) error {
	e, f, err := at.extractor(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	return e.Extract(ctx, f, nil, handleFile)
}

func (at Type) Extract(ctx context.Context, archivePath, fp string) (io.Reader, int64, error) {
	e, f, err := at.extractor(archivePath)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	var r = bytes.NewBuffer(nil)
	var size int64
	fn := func(ctx context.Context, f archiver.File) error {
		if f.Open == nil {
			return errors.New("file in archive cannot be opened")
		}

		src, err := f.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(r, src)
		if err != nil {
			return err
		}

		size = f.Size()

		return nil
	}

	if err := e.Extract(ctx, f, []string{fp}, fn); err != nil {
		return nil, 0, err
	}
	if size == 0 {
		return nil, 0, errors.New("extracted file has no size")
	}
	return r, size, nil
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
