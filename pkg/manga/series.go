package manga

import (
	"context"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/rs/xid"
	"golang.org/x/sync/errgroup"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/fse"
)

type Series struct {
	ID      xid.ID
	Title   string
	Entries []*Entry

	// Below are fields which aren't picked up by
	// the scan and shouldn't overwrite current
	// values that could exist
	// TODO: move these to within the DB and move all stuf in the manga pckag out?
	MangadexUUID         string    // UUID of a Mangadex entry
	MangadexLastAccessed time.Time // Last time the Mangadex entry was queried for new chapters
}

func folderID(dir string) (xid.ID, error) {
	f, err := os.Create(filepath.Join(dir, "/info.tanuki"))
	if err != nil {
		return xid.NilID(), err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return xid.NilID(), err
	}

	// If the file is empty then generate an ID
	if len(data) == 0 {
		id := xid.New()

		data, err := json.Marshal(id)
		if err != nil {
			return xid.NilID(), err
		}
		_, err = f.Write(data)
		if err != nil {
			return xid.NilID(), err
		}

		return id, nil
	}

	// Otherwise unmarshal it from the file
	var id xid.ID
	if err := json.Unmarshal(data, &id); err != nil {
		return xid.NilID(), err
	}
	return id, nil
}

func ParseSeries(ctx context.Context, dir string) (*Series, error) {
	id, err := folderID(dir)
	if err != nil {
		return nil, err
	}

	s := &Series{
		ID:      id,
		Title:   fse.Filename(dir),
		Entries: make([]*Entry, 0),
	}

	g, ctx := errgroup.WithContext(ctx)

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d != nil && !d.IsDir() {
			// We want to avoid parsing non-archive files like cover images
			_, err = archive.InferType(path)
			if err != nil {
				// We continue processing the folder if the file is not an archive
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
