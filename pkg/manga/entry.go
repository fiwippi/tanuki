package manga

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mholt/archiver/v4"

	"github.com/fiwippi/tanuki/internal/platform/archive"
	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/internal/platform/fse"
	"github.com/fiwippi/tanuki/internal/platform/hash"
	"github.com/fiwippi/tanuki/internal/platform/image"
)

// Entry represents an entry which you read, i.e. an archive file
type Entry struct {
	SID          string            `json:"sid" db:"sid"`
	EID          string            `json:"eid" db:"eid"`
	FileTitle    string            `json:"title" db:"title"`
	Archive      Archive           `json:"archive" db:"archive"`
	Pages        Pages             `json:"pages" db:"pages"`
	ModTime      dbutil.Time       `json:"mod_time" db:"mod_time"`
	DisplayTitle dbutil.NullString `json:"display_title" db:"display_title"`
}

func (e Entry) Title() string {
	if e.DisplayTitle != "" {
		return string(e.DisplayTitle)
	}
	return e.FileTitle
}

func ParseEntry(ctx context.Context, fp string) (Entry, error) {
	// Ensure valid archive
	at, err := archive.InferType(fp)
	if err != nil {
		return Entry{}, err
	}

	// Create the archive
	absFp, err := filepath.Abs(fp)
	if err != nil {
		return Entry{}, err
	}
	aStats, err := os.Stat(absFp)
	if err != nil {
		return Entry{}, err
	}
	a := Archive{
		Path:  absFp,
		Type:  at,
		Title: fse.Filename(fp),
	}

	// Create the entry
	title := fse.Filename(a.Path)
	e := Entry{
		Archive:   a,
		EID:       hash.SHA1(title),
		FileTitle: title,
		Pages:     make(Pages, 0),
		ModTime:   dbutil.Time(aStats.ModTime().Round(time.Second)),
	}

	// Iterate through the files in the archive
	err = a.Walk(ctx, func(ctx context.Context, f archiver.File) error {
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			it, err := image.InferType(f.Name())
			if err != nil {
				return err
			}

			e.Pages = append(e.Pages, Page{
				Path: f.NameInArchive,
				Type: it,
			})
		}
		return nil
	})
	if err != nil {
		return Entry{}, err
	}

	// Archive should contain images
	if len(e.Pages) == 0 {
		return Entry{}, fmt.Errorf("archive has no pages")
	}

	// Walker does not walk the archive in archived order so we need to sort the pages
	sort.SliceStable(e.Pages, func(i, j int) bool {
		a := strings.TrimSuffix(e.Pages[i].Path, filepath.Ext(e.Pages[i].Path))
		b := strings.TrimSuffix(e.Pages[j].Path, filepath.Ext(e.Pages[j].Path))
		return fse.SortNatural(a, b)
	})

	return e, nil
}
