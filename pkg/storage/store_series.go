package storage

import (
	"bytes"
	"errors"
	"io/ioutil"

	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/fiwippi/tanuki/pkg/manga"
)

// TODO methods for subscriptions
// TODO will long operations like populating the catalog and thumbnails block other operations

// Core

func (s *Store) hasSeries(tx *sqlx.Tx, sid string) bool {
	var exists bool
	tx.Get(&exists, "SELECT COUNT(sid) > 0 FROM series WHERE sid = ?", sid)
	return exists
}

func (s *Store) HasSeries(sid string) (bool, error) {
	var exists bool
	fn := func(tx *sqlx.Tx) error {
		exists = s.hasSeries(tx, sid)
		return nil
	}

	if err := s.tx(fn); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) getSeries(tx *sqlx.Tx, sid string) (*manga.Series, error) {
	var v manga.Series
	stmt := `
		SELECT 
			sid, folder_title, num_entries, num_pages, mod_time, 
			display_title, tags, mangadex_uuid, mangadex_last_published_at 
		FROM series WHERE sid = ?`
	err := tx.Get(&v, stmt, sid)
	if err != nil {
		return nil, err
	}
	return &v, nil

}

func (s *Store) GetSeries(sid string) (*manga.Series, error) {
	var v *manga.Series
	var err error
	fn := func(tx *sqlx.Tx) error {
		v, err = s.getSeries(tx, sid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Store) AddSeries(series *manga.Series, entries []*manga.Entry) error {
	fn := func(tx *sqlx.Tx) error {
		// Insert the series data
		stmt := `
		REPLACE INTO series
			(sid, folder_title, num_entries, num_pages, mod_time)
		Values
			(:sid, :folder_title, :num_entries, :num_pages, :mod_time)`
		_, err := tx.NamedExec(stmt, series)
		if err != nil {
			return err
		}

		//Insert each entry
		for i, e := range entries {
			if err := s.addEntry(tx, e, i+1); err != nil {
				return err
			}
		}

		return nil
	}

	return s.tx(fn)
}

func (s *Store) deleteSeries(tx *sqlx.Tx, sid string) error {
	_, err := tx.Exec(`DELETE FROM series WHERE sid = ?`, sid)
	return err
}

func (s *Store) DeleteSeries(sid string) error {
	fn := func(tx *sqlx.Tx) error {
		return s.deleteSeries(tx, sid)
	}

	return s.tx(fn)
}

// Cover / Thumbnail

func (s *Store) getSeriesCover(tx *sqlx.Tx, sid string) ([]byte, image.Type, error) {
	e, err := s.getFirstEntry(tx, sid)
	if err != nil {
		return nil, image.Invalid, err
	}

	// Check if the custom cover exists
	var data []byte
	tx.Get(&data, "SELECT custom_cover FROM series WHERE sid = ?", sid)
	if len(data) > 0 {
		var it = image.Invalid
		tx.Get(&it, "SELECT custom_cover_type FROM series WHERE sid = ?", sid)
		if it == image.Invalid {
			return nil, it, errors.New("thumbnail is an invalid image")
		}
		return data, it, nil
	}

	// If it doesn't then get the first page from the archive
	r, _, it, err := s.getPage(tx, sid, e.EID, 1)
	if err != nil {
		return nil, image.Invalid, err
	}
	data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, image.Invalid, err
	}
	return data, it, nil
}

func (s *Store) GetSeriesCover(sid string) ([]byte, image.Type, error) {
	var data []byte
	var it image.Type
	fn := func(tx *sqlx.Tx) error {
		var err error
		data, it, err = s.getSeriesCover(tx, sid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, image.Invalid, err
	}
	return data, it, nil
}

// TODO: test what happens if invalid images are given, also for setentrycover

func (s *Store) SetSeriesCover(sid, name string, data []byte) error {
	if len(data) == 0 {
		return ErrInvalidCover
	}

	it, err := image.InferType(name)
	if err != nil {
		return err
	}

	fn := func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE series SET custom_cover = ?, custom_cover_type = ? WHERE sid = ?`, data, it, sid)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE series SET thumbnail = NULL WHERE sid = ?;`, sid)
		return err
	}

	return s.tx(fn)
}

func (s *Store) DeleteSeriesCustomCover(sid string) error {
	stmt := `
		UPDATE series SET custom_cover = NULL WHERE sid = ?;
		UPDATE series SET custom_cover_type = NULL WHERE sid = ?;
		UPDATE series SET thumbnail = NULL WHERE sid = ?;
	`
	_, err := s.pool.Exec(stmt, sid, sid)
	return err
}

func (s *Store) generateSeriesThumbnail(tx *sqlx.Tx, sid string, overwrite bool) ([]byte, error) {
	var data []byte
	tx.Get(&data, "SELECT thumbnail FROM series WHERE sid = ?", sid)
	if len(data) > 0 && !overwrite {
		return data, nil
	}

	cover, it, err := s.getSeriesCover(tx, sid)
	if err != nil {
		return nil, err
	}
	thumb, err := it.EncodeThumbnail(bytes.NewReader(cover))
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("UPDATE series SET thumbnail = ? WHERE sid = ?", thumb, sid)
	if err != nil {
		return nil, err
	}
	return thumb, nil
}

func (s *Store) GetSeriesThumbnail(sid string) ([]byte, image.Type, error) {
	var data []byte
	fn := func(tx *sqlx.Tx) error {
		var err error
		data, err = s.generateSeriesThumbnail(tx, sid, false)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, image.Invalid, err
	}
	return data, image.JPEG, nil
}

// Tags / Metadata

// TODO: test that modtime change on the archive deletes the custom metadata for entries
// TODO: can we make tags only values and not pointers

func (s *Store) SetSeriesTags(sid string, tags *manga.Tags) error {
	_, err := s.pool.Exec("UPDATE series SET tags = ? WHERE sid = ?", tags, sid)
	return err
}

func (s *Store) GetSeriesWithTag(t string) ([]*manga.Series, error) {
	ctl, err := s.GetCatalog()
	if err != nil {
		return nil, err
	}

	filtered := make([]*manga.Series, 0)
	for _, series := range ctl {
		if series.Tags != nil && series.Tags.Has(t) {
			filtered = append(filtered, series)
		}
	}

	return filtered, nil
}

func (s *Store) GetAllTags() (*manga.Tags, error) {
	ctl, err := s.GetCatalog()
	if err != nil {
		return nil, err
	}

	all := manga.NewTags()
	for _, series := range ctl {
		if series.Tags != nil {
			all.Combine(series.Tags)
		}
	}

	return all, nil
}

func (s *Store) SetSeriesDisplayTitle(sid string, title string) error {
	_, err := s.pool.Exec("UPDATE series SET display_title = ? WHERE sid = ?", title, sid)
	return err
}
