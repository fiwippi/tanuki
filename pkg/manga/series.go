package manga

import (
	"context"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/fse"
)

type Series struct {
	Title   string
	Entries []*Entry
}

func ParseSeries(ctx context.Context, dir string) (*Series, error) {
	s := &Series{
		Title:   fse.Filename(dir),
		Entries: make([]*Entry, 0),
	}

	g, ctx := errgroup.WithContext(ctx)

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d != nil && !d.IsDir() {
			// We want to avoid parsing non-archive files like cover images
			_, err = archive.InferType(path)
			if err != nil {
				// We log it but we continue processing the folder
				log.Trace().Err(err).Str("fp", path).Msg("file is not archive")
				return nil
			}

			// Parse the archive
			p := path
			g.Go(func() error {
				e, err := ParseArchive(ctx, p)
				if err != nil {
					return err
				}
				s.Entries = append(s.Entries, e)
				return nil
			})
		}

		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Sort the entries list and add the correct order for each one
	sort.SliceStable(s.Entries, func(i, j int) bool {
		return fse.SortNatural(s.Entries[i].Archive.Title, s.Entries[j].Archive.Title)
	})

	return s, nil
}
