package manga

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/mholt/archiver/v3"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/errors"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/image"
)

func isDigit(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

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
			page := &Page{
				ImageType: t,
				Path:      a.GetPath(f),
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
		a := strings.TrimSuffix(e.Pages[i].Path, filepath.Ext(e.Pages[i].Path))
		b := strings.TrimSuffix(e.Pages[j].Path, filepath.Ext(e.Pages[j].Path))

		aFirst := rune(a[0])
		bFirst := rune(b[0])

		// First case if both aren't digits
		if !(isDigit(string(aFirst)) && isDigit(string(bFirst))) {
			// Only sorts them if one is lowercase and one is uppercase
			// and they're alphanumeric, i.e. not '_' or '.' etc.
			if unicode.IsLetter(aFirst) && unicode.IsLetter(bFirst) {
				aLowercase := string(aFirst) == strings.ToLower(string(aFirst))
				bUppercase := string(bFirst) == strings.ToUpper(string(bFirst))

				if aLowercase && bUppercase {
					return true
				} else if !aLowercase && !bUppercase {
					return false
				}
			}
		} else {
			// Second case if both are digits
			aBase := filepath.Base(a)
			bBase := filepath.Base(b)
			if isDigit(aBase) && isDigit(bBase) {
				aNum, err := strconv.Atoi(aBase)
				if err != nil {
					panic(err)
				}
				bNum, err := strconv.Atoi(bBase)
				if err != nil {
					panic(err)
				}
				return aNum < bNum
			}
		}

		// Third case is an underscore and a letter/digit
		if (aFirst == '_' || bFirst == '_') && (aFirst != bFirst) {
			return aFirst == '_' && bFirst != '_'
		}

		// Fourth case is if they're all the same up to the
		// base then we sort them based on the base value
		if filepath.Dir(a) == filepath.Dir(b) {
			aBase := filepath.Base(a)
			bBase := filepath.Base(b)
			return aBase < bBase
		}

		// Fifth case is general
		return a < b
	})

	// First item (which is a page) is guaranteed to be first in order
	// of pages since all cbz/cbr archive are ordered beforehand, so
	// we default to using the first page as cover if one could not be set
	e.Archive.Cover.Fp = e.Pages[0].Path
	e.Archive.Cover.ImageType = e.Pages[0].ImageType

	return e, nil
}
