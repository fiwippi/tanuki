package manga

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mholt/archiver/v4"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/internal/image"
)

// Entry represents an entry which you read, i.e. an archive file
type Entry struct {
	Hash    string   `json:"hash"`
	Title   string   `json:"title"`
	Archive *Archive `json:"archive"`
	Pages   []string `json:"pages"`
}

func ParseArchive(ctx context.Context, fp string) (*Entry, error) {
	// Ensure valid archive
	at, err := archive.InferType(fp)
	if err != nil {
		return nil, err
	}

	// Create the archive
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
		ModTime: aStats.ModTime(),
		Title:   fse.Filename(fp),
	}

	// Create the entry
	title := fse.Filename(a.Path)
	e := &Entry{
		Archive: a,
		Hash:    hash.SHA1(title),
		Title:   title,
		Pages:   make([]string, 0),
	}

	// Iterate through the files in the archive
	err = a.Walk(ctx, func(ctx context.Context, f archiver.File) error {
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			_, err := image.InferType(f.Name())
			if err != nil {
				return err
			}

			e.Pages = append(e.Pages, f.NameInArchive)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Archive should contain images
	if len(e.Pages) == 0 {
		return nil, fmt.Errorf("archive has no pages")
	}

	// Walker does not walk the archive in archived order so we need to sort the pages
	sort.SliceStable(e.Pages, func(i, j int) bool {
		a := strings.TrimSuffix(e.Pages[i], filepath.Ext(e.Pages[i]))
		b := strings.TrimSuffix(e.Pages[j], filepath.Ext(e.Pages[j]))
		return fse.SortNatural(a, b)
	})

	return e, nil
}
