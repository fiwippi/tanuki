package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/sortnat"
	"github.com/fiwippi/tanuki/pkg/manga"
)

type MissingStatus int

const (
	NotMissing MissingStatus = 0
	IsMissing  MissingStatus = 1
)

type MissingItem struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

func (s *Store) PopulateCatalog() error {
	fn := func(tx *sqlx.Tx) error {
		return s.populateCatalog(tx)
	}
	return s.tx(fn)
}

func (s *Store) populateCatalog(tx *sqlx.Tx) error {
	// Make all series and entries missing so that newly
	// added ones are classed as not missing
	_, err := tx.Exec(`UPDATE series SET missing=1`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE entries SET missing=1`)
	if err != nil {
		return err
	}

	// Scan series
	items, err := os.ReadDir(s.libraryPath)
	if err != nil {
		return err
	}

	// Add series
	var errs error
	for _, item := range items {
		// Only accept folders, (i.e. no standalone files)
		if !item.IsDir() {
			continue
		}

		fp := filepath.Join(s.libraryPath, item.Name())
		series, entries, err := manga.ParseSeries(context.Background(), fp)
		if err != nil {
			errs = errors.Join(err)
			continue
		}

		err = s.addSeries(tx, series, entries)
		if err != nil {
			errs = errors.Join(err)
		}
	}
	return errs
}

func (s *Store) getCatalog(tx *sqlx.Tx, mstatus MissingStatus) ([]manga.Series, error) {
	var v []manga.Series

	stmt, err := tx.Preparex(
		`SELECT sid, folder_title, num_entries, num_pages, mod_time, tags
		 FROM series 
		 WHERE missing=?
		 ORDER BY ROWID DESC`)
	if err != nil {
		return nil, err
	}
	err = stmt.Select(&v, int(mstatus))
	if err != nil {
		return nil, err
	}

	sort.Slice(v, func(i, j int) bool {
		a := v[i].Title
		b := v[j].Title
		return sortnat.Natural(a, b)
	})

	return v, nil
}

func (s *Store) GetCatalog() ([]manga.Series, error) {
	var ctl []manga.Series
	fn := func(tx *sqlx.Tx) error {
		var err error
		ctl, err = s.getCatalog(tx, NotMissing)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return ctl, nil
}

func (s *Store) GenerateThumbnails(overwrite bool) error {
	var errs error

	// Get all sids
	var sids []string
	s.pool.Select(&sids, `SELECT sid FROM series`)

	// Generate thumbnails for each series
	for _, sid := range sids {
		time.Sleep(1 * time.Second)

		fn := func(tx *sqlx.Tx) error {
			_, err := s.generateSeriesThumbnail(tx, sid, overwrite)
			return err
		}
		if err := s.tx(fn); err != nil {
			errs = errors.Join(err)
			continue
		}

		// Get entry thumbnails
		var eids []string
		s.pool.Select(&eids, `SELECT eid FROM entries WHERE sid = ?`, sid)

		// Generate thumbnails for each series
		for _, eid := range eids {
			time.Sleep(1 * time.Second)

			fn := func(tx *sqlx.Tx) error {
				_, err := s.generateEntryThumbnail(tx, sid, eid, overwrite)
				return err
			}
			if err := s.tx(fn); err != nil {
				errs = errors.Join(err)
			}
		}
	}

	return errs
}

func (s *Store) GetMissingItems() ([]MissingItem, error) {
	var missing []MissingItem
	fn := func(tx *sqlx.Tx) error {
		// Track all missing items
		err := s.populateCatalog(tx)
		if err != nil {
			return err
		}

		// Loop over each table and extract missing items
		series, err := s.getCatalog(tx, IsMissing)
		if err != nil {
			return err
		}
		// Series
		for _, ser := range series {
			fp := filepath.Join(s.libraryPath, ser.Title)
			missing = append(missing, MissingItem{
				Type:  "Series",
				Title: ser.Title,
				Path:  fp,
			})
		}

		// Entries - must check for both series which are missing and not missing
		var sids []string
		err = tx.Select(&sids, `SELECT sid FROM series`)
		if err != nil {
			return err
		}

		for _, sid := range sids {
			entries, err := s.getEntries(tx, sid, IsMissing)
			if err != nil {
				return err
			}
			for _, e := range entries {
				missing = append(missing, MissingItem{
					Type:  "Entry",
					Title: e.Title,
					Path:  e.Archive.Path,
				})
			}
		}

		return nil
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return missing, nil
}

func (s *Store) DeleteMissingItems() error {
	fn := func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM series WHERE missing = ?`, IsMissing)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`DELETE FROM entries WHERE missing = ?`, IsMissing)
		if err != nil {
			return err
		}

		return nil
	}

	return s.tx(fn)
}
