package storage

import (
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/pkg/human"
)

// TODO test for stuff not existing and it should return proper not exists error instead of sql.row not exists
// TODO if an entry or series was deleted its accompanying progress should be deleted as well
// TODO what error does get entry return if an entry does not exist
// TODO test that when creating series progress, all relevant progress entries are created in the table

var ErrUserNotExist = errors.New("user does not exist")
var ErrInvalidProgressAmount = errors.New("progress amount is invalid (below zero or above total number of pages)")

func (s *Store) setEntryProgress(tx *sqlx.Tx, sid, eid, uid string, num int, setUnread, setRead bool) error {
	// Validate user and entry exist
	if !s.hasUser(tx, uid) {
		return ErrUserNotExist
	}
	e, err := s.getEntry(tx, sid, eid)
	if err != nil {
		return err
	}

	// Validate amount is valid
	if num < 0 || num > e.Pages.Total() {
		return ErrInvalidProgressAmount
	}
	if setUnread {
		num = 0
	} else if setRead {
		num = e.Pages.Total()
	}

	// Update progress
	stmt := `
		REPLACE INTO progress
			(sid, eid, uid, current, total)
		Values
			(?, ?, ?, ?, ?)`
	_, err = tx.Exec(stmt, sid, eid, uid, num, e.Pages.Total())
	return err
}

func (s *Store) SetEntryProgressAmount(sid, eid, uid string, num int) error {
	fn := func(tx *sqlx.Tx) error {
		return s.setEntryProgress(tx, sid, eid, uid, num, false, false)
	}

	return s.tx(fn)
}

func (s *Store) SetEntryProgressUnread(sid, eid, uid string) error {
	fn := func(tx *sqlx.Tx) error {
		return s.setEntryProgress(tx, sid, eid, uid, 0, true, false)
	}

	return s.tx(fn)
}

func (s *Store) SetEntryProgressRead(sid, eid, uid string) error {
	fn := func(tx *sqlx.Tx) error {
		return s.setEntryProgress(tx, sid, eid, uid, 0, false, true)
	}

	return s.tx(fn)
}

func (s *Store) setSeriesProgress(tx *sqlx.Tx, sid, uid string, setUnread, setRead bool) error {
	// Validate user and entry exist
	if !s.hasUser(tx, uid) {
		return ErrUserNotExist
	}
	en, err := s.getEntries(tx, sid)
	if err != nil {
		return err
	}

	for _, e := range en {
		// Validate amount is valid
		var num int
		if setUnread {
			num = 0
		} else if setRead {
			num = e.Pages.Total()
		}

		// Update progress
		stmt := `
		REPLACE INTO progress
			(sid, eid, uid, current, total)
		Values
			(?, ?, ?, ?, ?)`
		_, err = tx.Exec(stmt, sid, e.EID, uid, num, e.Pages.Total())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) SetSeriesProgressUnread(sid, uid string) error {
	fn := func(tx *sqlx.Tx) error {
		return s.setSeriesProgress(tx, sid, uid, true, false)
	}

	return s.tx(fn)
}

func (s *Store) SetSeriesProgressRead(sid, uid string) error {
	fn := func(tx *sqlx.Tx) error {
		return s.setSeriesProgress(tx, sid, uid, false, true)
	}

	return s.tx(fn)
}

func (s *Store) GetEntryProgress(sid, eid, uid string) (human.EntryProgress, error) {
	var p human.EntryProgress
	fn := func(tx *sqlx.Tx) error {
		stmt := `SELECT eid, current, total FROM progress WHERE sid = ? AND eid = ? AND uid = ?`

		// Get the entry
		err := tx.Get(&p, stmt, sid, eid, uid)
		if err == nil {
			return nil
		}

		// The entry doesn't exist, so we have to create it as unread
		err = s.setEntryProgress(tx, sid, eid, uid, 0, true, false)
		if err != nil {
			return err
		}

		// Get the entry progress which should be now created
		return tx.Get(&p, stmt, sid, eid, uid)
	}

	if err := s.tx(fn); err != nil {
		return human.EntryProgress{}, err
	}
	return p, nil
}

func (s *Store) GetSeriesProgress(sid, uid string) (human.SeriesProgress, error) {
	var ep []human.EntryProgress
	stmt := `SELECT eid, current, total FROM progress WHERE sid = ? AND uid = ?`
	if err := s.pool.Select(&ep, stmt, sid, uid); err != nil {
		return human.SeriesProgress{}, err
	}

	sp := human.NewSeriesProgress()
	for _, e := range ep {
		sp.Add(e.EID, e)
	}
	return *sp, nil
}
