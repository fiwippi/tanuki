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

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/internal/sortnat"
	"github.com/fiwippi/tanuki/internal/sqlutil"
)

// Entry represents an entry which you read, i.e. an archive file
type Entry struct {
	SID     string       `json:"sid" db:"sid"`
	EID     string       `json:"eid" db:"eid"`
	Title   string       `json:"title" db:"title"`
	Archive Archive      `json:"archive" db:"archive"`
	Pages   Pages        `json:"pages" db:"pages"`
	ModTime sqlutil.Time `json:"mod_time" db:"mod_time"`
}

func ParseEntry(ctx context.Context, fp string) (Entry, error) {
	// Ensure valid archive
	t, err := archive.InferType(fp)
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
	title := fse.Filename(absFp)
	arch := Archive{
		Path:  absFp,
		Type:  t,
		Title: title,
	}

	// Create the entry
	e := Entry{
		Archive: arch,
		EID:     hash.SHA1(title),
		Title:   title,
		Pages:   make(Pages, 0),
		ModTime: sqlutil.Time(aStats.ModTime().Round(time.Second)),
	}

	// Iterate through the files in the archive and add them
	err = arch.Walk(ctx, func(ctx context.Context, f archiver.File) error {
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

	// Walker does not walk the archive in order, so we need to sort the pages ourselves
	sort.SliceStable(e.Pages, func(i, j int) bool {
		a := strings.TrimSuffix(e.Pages[i].Path, filepath.Ext(e.Pages[i].Path))
		b := strings.TrimSuffix(e.Pages[j].Path, filepath.Ext(e.Pages[j].Path))
		return sortnat.Natural(a, b)
	})

	return e, nil
}
