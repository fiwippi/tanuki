package storage

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrItemNotExist = errors.New("item does not exist")
var ErrInvalidCover = errors.New("cover is invalid")

type txFunc func(tx *sqlx.Tx) error

func (s *Store) tx(fn txFunc) error {
	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = fn(tx)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrItemNotExist
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
