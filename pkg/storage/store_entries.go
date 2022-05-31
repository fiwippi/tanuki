package storage

import (
	"io"
	"io/ioutil"

	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/fiwippi/tanuki/pkg/manga"
)

// TODO test modtime changing causes entry metadata to be deleted

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

func (s *Store) getEntry(tx *sqlx.Tx, sid, eid string) (*manga.Entry, error) {
	var e manga.Entry
	stmt := `
		SELECT 
			sid, eid, title, archive, pages, mod_time, display_title
		FROM entries WHERE sid = ? AND eid = ?`
	err := tx.Get(&e, stmt, sid, eid)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) GetEntry(sid, eid string) (*manga.Entry, error) {
	var e *manga.Entry
	var err error
	fn := func(tx *sqlx.Tx) error {
		e, err = s.getEntry(tx, sid, eid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) getFirstEntry(tx *sqlx.Tx, sid string) (*manga.Entry, error) {
	var e manga.Entry
	stmt := `
		SELECT 
			sid, eid, title, archive, pages, mod_time, display_title
		FROM entries WHERE sid = ? ORDER BY position ASC, ROWID DESC LIMIT 1`
	err := tx.Get(&e, stmt, sid)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) getEntries(tx *sqlx.Tx, sid string) ([]*manga.Entry, error) {
	var e []*manga.Entry
	stmt := `
	SELECT 
	    sid, eid, title, archive, pages, mod_time, display_title
	FROM entries WHERE sid = ? ORDER BY position ASC, ROWID DESC `
	err := tx.Select(&e, stmt, sid)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) GetEntries(sid string) ([]*manga.Entry, error) {
	var e []*manga.Entry
	fn := func(tx *sqlx.Tx) error {
		var err error
		e, err = s.getEntries(tx, sid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) addEntry(tx *sqlx.Tx, e *manga.Entry, position int) error {
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

	stmt := `
		REPLACE INTO entries 
			(sid, eid, title, archive, pages, mod_time) 
		Values 
			(:sid, :eid, :title, :archive, :pages, :mod_time)`
	_, err := tx.NamedExec(stmt, e)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE entries SET position = ? WHERE sid = ? AND  eid = ?", position, e.SID, e.EID)
	return err
}

func (s *Store) deleteEntry(tx *sqlx.Tx, sid, eid string) error {
	_, err := tx.Exec(`DELETE FROM entries WHERE sid = ? AND eid = ?`, sid, eid)
	return err
}

// Cover / Thumbnail / Page

func (s *Store) getEntryCover(tx *sqlx.Tx, sid, eid string) ([]byte, error) {
	e, err := s.getEntry(tx, sid, eid)
	if err != nil {
		return nil, err
	}

	// Check if the custom cover exists
	var data []byte
	tx.Get(&data, "SELECT custom_cover FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	if len(data) > 0 {
		return data, nil
	}

	// If it doesn't then get the first page from the archive
	r, _, err := s.getPage(tx, sid, e.EID, 1)
	if err != nil {
		return nil, err
	}
	data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Store) GetEntryCover(sid, eid string) ([]byte, error) {
	var data []byte
	var err error
	fn := func(tx *sqlx.Tx) error {
		data, err = s.getEntryCover(tx, sid, eid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Store) SetEntryCover(sid, eid string, data []byte) error {
	if len(data) == 0 {
		return ErrInvalidCover
	}

	fn := func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE entries SET custom_cover = ? WHERE sid = ? AND eid = ?`, data, sid, eid)
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

	cover, err := s.getEntryCover(tx, sid, eid)
	if err != nil {
		return nil, err
	}
	thumb, err := image.EncodeThumbnail(cover)
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

func (s *Store) GetEntryThumbnail(sid, eid string) ([]byte, error) {
	var data []byte
	fn := func(tx *sqlx.Tx) error {
		var err error
		data, err = s.generateEntryThumbnail(tx, sid, eid, false)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Store) getPage(tx *sqlx.Tx, sid, eid string, page int) (io.Reader, int64, error) {
	var a manga.Archive
	err := tx.Get(&a, "SELECT archive FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	if err != nil {
		return nil, 0, err
	}
	var p manga.Pages
	err = tx.Get(&p, "SELECT pages FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	if err != nil {
		return nil, 0, err
	}

	// Pages are 1-indexed
	return a.ReaderForFile(p[page-1])
}

func (s *Store) GetPage(sid, eid string, page int, zeroBased bool) (io.Reader, int64, error) {
	if zeroBased {
		page += 1
	}

	var r io.Reader
	var size int64
	var err error
	fn := func(tx *sqlx.Tx) error {
		r, size, err = s.getPage(tx, sid, eid, page)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, 0, err
	}
	return r, size, nil
}

// Metadata

func (s *Store) SetEntryDisplayTitle(sid, eid string, title string) error {
	_, err := s.pool.Exec("UPDATE entries SET display_title = ? WHERE sid = ? AND eid = ?", title, sid, eid)
	return err
}
