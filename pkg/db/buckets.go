package db

import (
	bolt "go.etcd.io/bbolt"
)

var (
	bucketUsers = []byte("users")
	bucketSeries = []byte("series")
)

func (db *DB) usersBucket(tx *bolt.Tx) *UsersBucket {
	return &UsersBucket{tx.Bucket(bucketUsers)}
}

func (db *DB) seriesListBucket(tx *bolt.Tx) *SeriesListBucket {
	return &SeriesListBucket{tx.Bucket(bucketSeries)}
}

// moveBucket moves the inner bucket with key 'oldkey' to a new bucket with key 'newkey'
// must be used within an Update-transaction, it automatically create the new bucket
// https://github.com/boltdb/bolt/issues/396
func moveBucket(oldParent, newParent *bolt.Bucket, oldkey, newkey []byte, deleteOld bool) error {
	oldBucket := oldParent.Bucket(oldkey)
	newBucket, err := newParent.CreateBucket(newkey)
	if err != nil {
		return err
	}

	err = oldBucket.ForEach(func(k, v []byte) error {
		if v == nil {
			// Nested bucket
			return renameBucket(oldBucket, newBucket, k, k)
		} else {
			// Regular value
			return newBucket.Put(k, v)
		}
	})
	if err != nil {
		return err
	}

	// This deletes the oldkey
	if deleteOld {
		return oldParent.DeleteBucket(oldkey)
	}
	return nil
}

func renameBucket(oldParent, newParent *bolt.Bucket, oldkey, newkey []byte) error {
	return moveBucket(oldParent, newParent, oldkey, newkey, true)
}

func copyBucket(oldParent, newParent *bolt.Bucket, oldkey, newkey []byte) error {
	return moveBucket(oldParent, newParent, oldkey, newkey, false)
}

