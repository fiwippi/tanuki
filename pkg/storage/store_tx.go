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

func processTx(pool *sqlx.DB, fn txFunc) error {
	// Begin transaction
	tx, err := pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Run transaction
	err = fn(tx)
	if err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

const retries = 10
const retryTimout = 50 * time.Millisecond

func (s *Store) tx(fn txFunc) error {
	// Begin loop
	for i := 1; i <= retries; i++ {
		err := processTx(s.pool, fn)

		if dbBusy(err) && i < retries {
			time.Sleep(retryTimout)
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
	return nil
}
