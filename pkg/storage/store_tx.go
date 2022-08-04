package storage

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"modernc.org/sqlite"
)

var ErrItemNotExist = errors.New("item does not exist")
var ErrInvalidCover = errors.New("cover is invalid")

type txFunc func(tx *sqlx.Tx) error

func dbBusy(err error) bool {
	sqlErr, ok := err.(*sqlite.Error)
	if ok {
		return sqlErr.Code() == 5 || sqlErr.Code() == 517
	}
	return false
}

func (s *Store) tx(fn txFunc) error {
	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i := 0; i < 5; i++ {
		err = fn(tx)
		if dbBusy(err) {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err != nil {
			if err == sql.ErrNoRows {
				return ErrItemNotExist
			}
			return err
		}
		break
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
