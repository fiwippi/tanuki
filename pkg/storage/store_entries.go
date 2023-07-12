package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/pkg/manga"
)

func (s *Store) hasEntry(tx *sqlx.Tx, sid, eid string) bool {
	var exists bool
	tx.Get(&exists, "SELECT COUNT(sid) > 0 FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	return exists
}

func (s *Store) HasEntry(sid, eid string) (bool, error) {
	var exists bool
	fn := func(tx *sqlx.Tx) error {
		exists = s.hasEntry(tx, sid, eid)
		return nil
	}

	if err := s.tx(fn); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) getEntry(tx *sqlx.Tx, sid, eid string) (manga.Entry, error) {
	var e manga.Entry
	stmt := `
		SELECT 
			sid, eid, title, archive, pages, mod_time
		FROM entries WHERE sid = ? AND eid = ?`
	err := tx.Get(&e, stmt, sid, eid)
	if err != nil {
		return manga.Entry{}, err
	}
	return e, nil
}

func (s *Store) GetEntry(sid, eid string) (manga.Entry, error) {
	var e manga.Entry
	var err error
	fn := func(tx *sqlx.Tx) error {
		e, err = s.getEntry(tx, sid, eid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return manga.Entry{}, err
	}
	return e, nil
}

func (s *Store) getFirstEntry(tx *sqlx.Tx, sid string) (manga.Entry, error) {
	var e manga.Entry
	stmt := `
		SELECT 
			sid, eid, title, archive, pages, mod_time
		FROM entries WHERE sid = ? ORDER BY position ASC, ROWID ASC LIMIT 1`
	err := tx.Get(&e, stmt, sid)
	if err != nil {
		return manga.Entry{}, err
	}
	return e, nil
}

func (s *Store) getEntries(tx *sqlx.Tx, sid string, mstatus MissingStatus) ([]manga.Entry, error) {
	var e []manga.Entry

	stmt, err := tx.Preparex(
		`SELECT sid, eid, title, archive, pages, mod_time
	     FROM entries
	     WHERE sid = ? AND missing = ?
	     ORDER BY position ASC, ROWID DESC `)
	if err != nil {
		return nil, err
	}

	err = stmt.Select(&e, sid, int(mstatus))
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) GetEntries(sid string) ([]manga.Entry, error) {
	var e []manga.Entry
	fn := func(tx *sqlx.Tx) error {
		var err error
		e, err = s.getEntries(tx, sid, 0)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) addEntry(tx *sqlx.Tx, e manga.Entry, position int) error {
	// If the entries modtimes are different, then the file
	// itself has changed which means we want to delete it
	// and then recreate it so data associated with it gets
	// wiped, e.g. progress data
	if s.hasEntry(tx, e.SID, e.EID) {
		currentEntry, err := s.getEntry(tx, e.SID, e.EID)
		if err != nil {
			return err
		}
		if !currentEntry.ModTime.Equal(e.ModTime) {
			err := s.deleteEntry(tx, e.SID, e.EID)
			if err != nil {
				return err
			}
		}
	}

	stmt := `INSERT INTO entries (sid, eid, title, archive, pages, mod_time, missing) 
					Values (:sid, :eid, :title, :archive, :pages, :mod_time, 0)
					ON CONFLICT (sid, eid)
					DO UPDATE SET sid=:sid, eid=:eid, title=:title, archive=:archive, pages=:pages, mod_time=:mod_time,
					              missing=0`
	_, err := tx.NamedExec(stmt, e)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE entries SET position = ? WHERE sid = ? AND eid = ?", position, e.SID, e.EID)
	return err
}

func (s *Store) deleteEntry(tx *sqlx.Tx, sid, eid string) error {
	_, err := tx.Exec(`DELETE FROM entries WHERE sid = ? AND eid = ?`, sid, eid)
	return err
}

// Cover / Thumbnail / Page

func (s *Store) getEntryCover(tx *sqlx.Tx, sid, eid string) ([]byte, image.Type, error) {
	e, err := s.getEntry(tx, sid, eid)
	if err != nil {
		return nil, image.Invalid, err
	}

	// Check if the custom cover exists
	var data []byte
	tx.Get(&data, "SELECT custom_cover FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	if len(data) > 0 {
		var it = image.Invalid
		tx.Get(&it, "SELECT custom_cover_type FROM entries WHERE sid = ? AND eid = ?", sid, eid)
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

func (s *Store) GetEntryCover(sid, eid string) ([]byte, image.Type, error) {
	var data []byte
	var it image.Type
	fn := func(tx *sqlx.Tx) error {
		var err error
		data, it, err = s.getEntryCover(tx, sid, eid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, image.Invalid, err
	}
	return data, it, nil
}

func (s *Store) GetEntryCoverType(sid, eid string) (image.Type, error) {
	it := image.Invalid
	fn := func(tx *sqlx.Tx) error {
		tx.Get(&it, "SELECT custom_cover_type FROM entries WHERE sid = ? AND eid = ?", sid, eid)
		return nil
	}

	if err := s.tx(fn); err != nil {
		return image.Invalid, err
	}
	return it, nil
}

func (s *Store) SetEntryCover(sid, eid, name string, data []byte) error {
	if len(data) == 0 {
		return ErrInvalidCover
	}
	it, err := image.InferType(name)
	if err != nil {
		return ErrInvalidCover
	}

	fn := func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE entries SET custom_cover = ?, custom_cover_type = ? WHERE sid = ? AND eid = ?`, data, it, sid, eid)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE entries SET thumbnail = NULL WHERE sid = ? AND eid = ?;`, sid, eid)
		return err
	}

	return s.tx(fn)
}

func (s *Store) DeleteEntryCustomCover(sid, eid string) error {
	stmt := `
		UPDATE entries SET custom_cover = NULL WHERE sid = ? AND eid = ?;
		UPDATE entries SET custom_cover_type = NULL WHERE sid = ? AND eid = ?;
		UPDATE entries SET thumbnail = NULL WHERE sid = ? AND eid = ?;
	`
	_, err := s.pool.Exec(stmt, sid, eid, sid, eid)
	return err
}

func (s *Store) generateEntryThumbnail(tx *sqlx.Tx, sid, eid string, overwrite bool) ([]byte, error) {
	var data []byte
	tx.Get(&data, "SELECT thumbnail FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	if len(data) > 0 && !overwrite {
		return data, nil
	}

	cover, it, err := s.getEntryCover(tx, sid, eid)
	if err != nil {
		return nil, err
	}
	img, err := it.Decode(bytes.NewReader(cover))
	if err != nil {
		return nil, err
	}
	thumb, err := it.EncodeThumbnail(img, 300, 300)
	if err != nil {
		return nil, err
	}

	// Save the created thumbnail
	_, err = tx.Exec("UPDATE entries SET thumbnail = ? WHERE sid = ? AND eid = ?", thumb, sid, eid)
	if err != nil {
		return nil, err
	}
	return thumb, nil

}

func (s *Store) GetEntryThumbnail(sid, eid string) ([]byte, image.Type, error) {
	var data []byte
	fn := func(tx *sqlx.Tx) error {
		var err error
		data, err = s.generateEntryThumbnail(tx, sid, eid, false)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, image.Invalid, err
	}
	return data, image.JPEG, nil
}

func (s *Store) getPage(tx *sqlx.Tx, sid, eid string, pageNum int) (io.Reader, int64, image.Type, error) {
	var a manga.Archive
	err := tx.Get(&a, "SELECT archive FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	if err != nil {
		return nil, 0, image.Invalid, err
	}
	var p manga.Pages
	err = tx.Get(&p, "SELECT pages FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	if err != nil {
		return nil, 0, image.Invalid, err
	}

	// Pages are 1-indexed
	index := pageNum - 1
	if index < 0 || index >= len(p) {
		return nil, 0, image.Invalid, fmt.Errorf("page num index out of range")
	}
	page := p[index]
	r, size, err := a.Extract(context.Background(), page.Path)
	if err != nil {
		return nil, 0, image.Invalid, err
	}
	return r, size, page.Type, nil
}

func (s *Store) GetPage(sid, eid string, page int, zeroBased bool) (io.Reader, int64, image.Type, error) {
	if zeroBased {
		page += 1
	}

	var r io.Reader
	var size int64
	var it image.Type
	fn := func(tx *sqlx.Tx) error {
		var err error
		r, size, it, err = s.getPage(tx, sid, eid, page)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, 0, image.Invalid, err
	}
	return r, size, it, nil
}
