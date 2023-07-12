package archive

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"

	"github.com/mholt/archiver/v4"
)

func (t Type) extractor() archiver.Extractor {
	return [...]archiver.Extractor{archiver.Zip{}, archiver.Rar{}}[t]
}

func (t Type) Extract(ctx context.Context, archive, pathInArchive string) (io.Reader, int64, error) {
	// Open archive on filesystem
	f, err := os.Open(archive)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	// Create empty buffer
	var r = bytes.NewBuffer(nil)
	var size int64

	// Define the handler for file extraction
	fn := func(ctx context.Context, f archiver.File) error {
		// Open file within archive
		if f.Open == nil {
			return errors.New("file in archive cannot be opened")
		}
		src, err := f.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Copy contents and size to buffer
		_, err = io.Copy(r, src)
		if err != nil {
			return err
		}
		size = f.Size()

		return nil
	}

	// Attempt file extraction
	err = t.extractor().Extract(ctx, f, []string{pathInArchive}, fn)
	if err != nil {
		return nil, 0, err
	}

	return r, size, nil
}

func (t Type) Walk(ctx context.Context, archive string, handler archiver.FileHandler) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.extractor().Extract(ctx, f, nil, handler)
}
