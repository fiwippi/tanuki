package manga

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/rs/xid"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/sqlutil"
)

type Series struct {
	SID        string       `json:"sid" db:"sid"`
	Title      string       `json:"folder_title" db:"folder_title"`
	NumEntries int          `json:"num_entries" db:"num_entries"`
	NumPages   int          `json:"num_pages" db:"num_pages"`
	ModTime    sqlutil.Time `json:"mod_time" db:"mod_time"`

	// Below are fields which aren't picked up by
	// the scan and shouldn't overwrite current
	// values that may exist
	Tags *Tags `json:"tags" db:"tags"`
}

func folderID(dir string) (string, error) {
	f, err := os.OpenFile(filepath.Join(dir, "info.tanuki"), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	// If the file is empty then generate an ID
	if len(data) == 0 {
		id := xid.New().String()

		// Write the new ID to the file
		_, err := f.WriteString(id)
		if err != nil {
			return "", err
		}

		return id, nil
	}

	// Otherwise return it from the file
	return string(data), nil
}

func ParseSeries(ctx context.Context, dir string) (Series, []Entry, error) {
	id, err := folderID(dir)
	if err != nil {
		return Series{}, nil, err
	}

	s := Series{
		SID:     id,
		Title:   fse.Filename(dir),
		ModTime: sqlutil.Time{},
	}
	en := make([]Entry, 0)

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d != nil && !d.IsDir() {
			// We want to avoid parsing non-archive files which may exist
			// in the folder
			_, err = archive.InferType(path)
			if err != nil {
				// We continue processing the folder if the file is not an archive
				return nil
			}

			// Parse the entry archive
			e, err := ParseEntry(ctx, path)
			if err != nil {
				return err
			}

			// Fill in remaining details
			e.SID = s.SID
			s.NumPages += len(e.Pages)
			if e.ModTime.After(s.ModTime) {
				s.ModTime = e.ModTime
			}

			// Add the entry to the list of entries
			en = append(en, e)
		}

		return nil
	})
	if err != nil {
		return Series{}, nil, err
	}

	s.NumEntries = len(en)

	sort.SliceStable(en, func(i, j int) bool {
		return sortorder.NaturalLess(en[i].Archive.Title, en[j].Archive.Title)
	})

	return s, en, nil
}
