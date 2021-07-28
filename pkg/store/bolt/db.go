package bolt

import (
	"fmt"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/bolt/util"
	bolt "go.etcd.io/bbolt"
)

type DB struct {
	*bolt.DB
}

func (db *DB) String() string {
	var s string
	err := db.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		util.DumpCursor(tx, c, 0, &s)
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("error when viewing string of db: %s", err))
	}
	return s
}

func Create(path string) (*DB, error) {
	temp, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	db := &DB{DB: temp}

	// Guarantees the buckets exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(keys.Users)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(keys.Catalog)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
