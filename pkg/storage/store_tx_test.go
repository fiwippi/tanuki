package storage

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestStore_tx(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	// An arbitrary transaction should succeed
	fn := func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS temp (x TEXT);`)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`DROP TABLE IF EXISTS temp`)
		if err != nil {
			return err
		}
		return nil
	}
	require.Nil(t, s.tx(fn))

	// Retrieving something which doesn't exist
	// should return the item not exit error
	fn = func(tx *sqlx.Tx) error {
		_, err := s.getSeries(tx, "")
		return err
	}
	require.ErrorIs(t, s.tx(fn), ErrItemNotExist)

}
