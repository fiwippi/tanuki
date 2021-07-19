package db

import (
	bolt "go.etcd.io/bbolt"
)


type DB struct {
	*bolt.DB
}

func CreateDB(path string) (*DB, error) {
	_db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	db := &DB{DB: _db}

	// Guarantees the buckets exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(bucketUsers)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(bucketSeries)
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

func (db *DB) returnBytes(f func(tx *bolt.Tx) ([]byte, string, error)) ([]byte, string, error) {
	var data []byte
	var mimetype string

	err := db.View(func(tx *bolt.Tx) error {
		d, m, err := f(tx)
		if err != nil {
			return err
		}
		data = make([]byte, len(d))
		copy(data, d)
		mimetype = m

		return nil
	})
	if err != nil {
		return nil, "", err
	}
	return data, mimetype, nil
}



