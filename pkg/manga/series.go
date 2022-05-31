package manga

import (
	"context"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/rs/xid"
	"golang.org/x/sync/errgroup"

	"github.com/fiwippi/tanuki/internal/platform/archive"
	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/internal/platform/fse"
)

// TODO test returning series as struct and values and as poitners
// TODO function to look for missing progress/entries/series/whatever
// TODO if an entry gets deleted on the filesystem instead of through the db, does the progress for it get delted as well

type Series struct {
	SID         string      `json:"sid" db:"sid"`
	FolderTitle string      `json:"folder_title" db:"folder_title"`
	NumEntries  int         `json:"num_entries" db:"num_entries"`
	NumPages    int         `json:"num_pages" db:"num_pages"`
	ModTime     dbutil.Time `json:"mod_time" db:"mod_time"`

	// Below are fields which aren't picked up by
	// the scan and shouldn't overwrite current
	// values that could exist
	Tags                *Tags             `json:"tags" db:"tags"`
	DisplayTile         dbutil.NullString `json:"display_title" db:"display_title"`
	MdexUUID            dbutil.NullString `json:"mangadex_uuid" db:"mangadex_uuid"`
	MdexLastPublishedAt dbutil.Time       `json:"mangadex_last_published_at" db:"mangadex_last_published_at"`
	// TODO test the manager works with subscriptions
}

func folderID(dir string) (string, error) {
	f, err := os.OpenFile(filepath.Join(dir, "info.tanuki"), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	// If the file is empty then generate an ID
	if len(data) == 0 {
		id := xid.New().String()

		_, err := f.WriteString(id)
		if err != nil {
			return "", err
		}

		return id, nil
	}

	// Otherwise return it from the file
	return string(data), nil
}

func ParseSeries(ctx context.Context, dir string) (*Series, []*Entry, error) {
	id, err := folderID(dir)
	if err != nil {
		return nil, nil, err
	}

	s := &Series{
		SID:         id,
		FolderTitle: fse.Filename(dir),
		ModTime:     dbutil.Time{},
	}
	en := make([]*Entry, 0)

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
				e, err := ParseEntry(ctx, p)
				if err != nil {
					return err
				}

				e.SID = s.SID
				s.NumPages += len(e.Pages)
				if e.ModTime.After(s.ModTime) {
					s.ModTime = e.ModTime
				}

				en = append(en, e)
				return nil
			})
		}

		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	s.NumEntries = len(en)

	// Sort the entries list and add the correct order for each one
	sort.SliceStable(en, func(i, j int) bool {
		return fse.SortNatural(en[i].Archive.Title, en[j].Archive.Title)
	})

	return s, en, nil
}
