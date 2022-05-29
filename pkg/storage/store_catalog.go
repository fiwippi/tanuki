package storage

import (
	"context"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/platform/errors"
	"github.com/fiwippi/tanuki/internal/platform/fse"
	"github.com/fiwippi/tanuki/pkg/manga"
)

// TODO benchmark populate catalog and generate thumbnails

func (s *Store) PopulateCatalog() error {
	items, err := os.ReadDir(s.libraryPath)
	if err != nil {
		return err
	}

	var errs errors.Errors
	for _, item := range items {
		if !item.IsDir() {
			continue
		}

		fp := filepath.Join(s.libraryPath, item.Name())
		series, entries, err := manga.ParseSeries(context.Background(), fp)
		if err != nil {
			errs.Add(err)
			continue
		}

		err = s.AddSeries(series, entries)
		if err != nil {
			errs.Add(err)
		}
	}

	return errs.Ret()
}

func (s *Store) getCatalog(tx *sqlx.Tx) ([]*manga.Series, error) {
	var v []*manga.Series
	stmt := `
		SELECT 
			sid, folder_title, num_entries, num_pages, mod_time, 
			display_title, tags, mangadex_uuid, mangadex_last_published_at 
		FROM series ORDER BY ROWID ASC`
	err := tx.Select(&v, stmt)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *Store) GetCatalog() ([]*manga.Series, error) {
	var ctl []*manga.Series
	fn := func(tx *sqlx.Tx) error {
		var err error
		ctl, err = s.getCatalog(tx)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return ctl, nil
}

func (s *Store) GenerateThumbnails(tx *sqlx.Tx, overwrite bool) error {
	var errs errors.Errors
	var sids []string
	tx.Select(&sids, `SELECT sid FROM series`)

	for _, sid := range sids {
		_, err := s.generateSeriesThumbnail(tx, sid, overwrite)
		if err != nil {
			errs.Add(err)
			continue
		}

		var eids []string
		tx.Select(&eids, `SELECT eid FROM entries WHERE sid = ?`, sid)

		for _, eid := range eids {
			_, err := s.generateEntryThumbnail(tx, sid, eid, overwrite)
			if err != nil {
				errs.Add(err)
			}
		}
	}

	return errs.Ret()
}

func (s *Store) processMissingItems(tx *sqlx.Tx, del bool) ([]MissingItem, error) {
	var missing []MissingItem

	catalog, err := s.getCatalog(tx)
	if err != nil {
		return nil, err
	}

	for _, series := range catalog {
		fp := filepath.Join(s.libraryPath, series.FolderTitle)
		if !fse.Exists(fp) {
			missing = append(missing, MissingItem{
				Type:  "Series",
				Title: series.FolderTitle,
				Path:  fp,
			})

			if del {
				if err := s.deleteSeries(tx, series.SID); err != nil {
					return nil, err
				}
			}

			continue
		}

		entries, err := s.getEntries(tx, series.SID)
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			if !e.Archive.Exists() {
				missing = append(missing, MissingItem{
					Type:  "Entry",
					Title: e.Title,
					Path:  e.Archive.Path,
				})

				if del {
					if err := s.deleteEntry(tx, e.SID, e.EID); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return missing, nil
}

func (s *Store) GrocessMissingItems() ([]MissingItem, error) {
	var missing []MissingItem
	fn := func(tx *sqlx.Tx) error {
		var err error
		missing, err = s.processMissingItems(tx, false)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return missing, nil
}

func (s *Store) DeleteMissingItems() error {
	fn := func(tx *sqlx.Tx) error {
		_, err := s.processMissingItems(tx, true)
		return err
	}

	return s.tx(fn)
}
