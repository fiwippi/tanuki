package manga

import (
	"archive/zip"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mholt/archiver/v3"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/errors"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/image"
)

var ErrNoPages = errors.New("archive contains no pages")

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
		Title:   fse.Filename(fp),
	}

	// Parse the archive into a EntryProgress struct
	e := newEntry()
	e.Archive = a

	// Iterate through the files in the archive, try to parse metadata
	err = a.Walk(func(f archiver.File) error {
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			// Only process the page if it's a valid image
			t, err := image.InferType(f.Name())
			if err != nil {
				return nil
			}

			// Parse the page
			page := &Page{ImageType: t}

			// If zip we need the file header to get the absolute filepath
			// but with rar calling f.Name() already gives it to us
			switch a.Type {
			case archive.Zip:
				page.Path = f.Header.(zip.FileHeader).Name
			case archive.Rar:
				page.Path = f.Name()
			}

			// Finally add the page
			e.Pages = append(e.Pages, page)
		}
		return nil
	})
	if err != nil {
		return nil, errors.New(err.Error()).Fmt(e.Archive.Title)
	}

	// Archive should contain images
	if len(e.Pages) == 0 {
		return nil, ErrNoPages.Fmt(e.Archive.Title)
	}

	// Add the rest of the metadata
	e.Metadata.Title = e.Archive.Title

	// Walker does not walk the archive in archived order so we need to sort the pages
	sort.SliceStable(e.Pages, func(i, j int) bool {
		// Lowercase should be uppercase
		a := fse.Filename(e.Pages[i].Path)
		b := fse.Filename(e.Pages[j].Path)

		if len(a) > 0 && len(b) > 0 {
			aFirst := string(a[0])
			bFirst := string(b[0])

			aLowercase := aFirst == strings.ToLower(aFirst)
			bUppercase := bFirst == strings.ToUpper(bFirst)

			if aLowercase && bUppercase {
				return true
			} else if !aLowercase && !bUppercase {
				return false
			}
		}
		return a < b
	})

	// First item (which is a page) is guaranteed to be first in order
	// of pages since all cbz/cbr archive are ordered beforehand, so
	// we default to using the first page as cover if one could not be set
	e.Archive.Cover.Fp = e.Pages[0].Path
	e.Archive.Cover.ImageType = e.Pages[0].ImageType

	return e, nil
}
