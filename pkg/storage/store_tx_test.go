package storage

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestStore_tx(t *testing.T) {
	s := mustOpenStoreMem(t)

	// An arbitrary transaction should succeed
	fn := func(tx *sqlx.Tx) error {
		_, err := s.pool.Exec(`CREATE TABLE IF NOT EXISTS temp (x TEXT);`)
		if err != nil {
			return err
		}
		_, err = s.pool.Exec(`DROP TABLE IF EXISTS temp`)
		if err != nil {
			return err
		}
		return nil
	}
	require.Nil(t, s.tx(fn))

	mustCloseStore(t, s)
}
