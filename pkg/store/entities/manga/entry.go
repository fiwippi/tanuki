package manga

import (
	"archive/zip"
	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/image"
	"github.com/mholt/archiver/v3"
	"os"
	"path/filepath"
	"strings"
)

// ParsedEntry represents an entry which you read, i.e. an archive file
type ParsedEntry struct {
	Order    int            // 1-indexed order
	Archive  *Archive       // EntriesMetadata about the manga archive file
	Metadata *EntryMetadata // Metadata of the manga
	Pages    []*Page        // Pages of the manga
}

func newEntry() *ParsedEntry {
	return &ParsedEntry{
		Archive:  &Archive{Cover: &Cover{}},
		Metadata: NewEntryMetadata(),
		Pages:    make([]*Page, 0),
	}
}

func ParseArchive(fp string) (*ParsedEntry, error) {
	// Ensure valid archive
	at, err := archive.InferType(fp)
	if err != nil {
		return nil, err
	}
	absFp, err := filepath.Abs(fp)
	if err != nil {
		return nil, err
	}
	aStats, err := os.Stat(absFp)
	if err != nil {
		return nil, err
	}
	a := &Archive{
		Path:    absFp,
		Type:    at,
		Cover:   &Cover{},
		ModTime: aStats.ModTime(),
	}

	// Parse the archive into a EntryProgress struct
	e := newEntry()
	e.Archive = a

	// Iterate through the files in the archive, try to parse metadata
	err = a.Walk(func(f archiver.File) error {
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			// Validate page has correct image type
			t, err := image.InferType(filepath.Ext(f.Name()))
			if err != nil {
				return err
			}

			// Allow one cover page
			if strings.HasPrefix(strings.ToLower(f.Name()), "cover") {
				e.Archive.Cover.Fp = f.Name()
				e.Archive.Cover.ImageType = t
			} else {
				// If image isn't a cover page then parse it as a proper page
				page := &Page{
					ImageType: t,
				}

				// If zip we need the file header to get the absolute filepath
				// but with rar calling f.Name() already gives it to us
				switch a.Type {
				case archive.Zip:
					page.Path = f.Header.(zip.FileHeader).Name
				case archive.Rar:
					page.Path = f.Name()
				}

				//
				e.Pages = append(e.Pages, page)
			}
		}

		return nil
	})

	// Add the rest of the metadata
	e.Archive.Title = fse.Filename(fp)
	e.Metadata.Title = e.Archive.Title

	// First item (which is a page) is guaranteed to be first in order
	// of pages since all cbz/cbr archive are ordered beforehand, so
	// we default to using the first page as cover if one could not be set
	if e.Archive.Cover.Fp == "" {
		e.Archive.Cover.Fp = e.Pages[0].Path
		e.Archive.Cover.ImageType = e.Pages[0].ImageType
	}

	return e, nil
}
