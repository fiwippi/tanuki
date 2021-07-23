package db

import (
	bolt "go.etcd.io/bbolt"
)

var (
	bucketUsers   = []byte("users")
	bucketCatalog = []byte("catalog")
)

func (db *DB) usersBucket(tx *bolt.Tx) *UsersBucket {
	return &UsersBucket{tx.Bucket(bucketUsers)}
}

func (db *DB) catalogBucket(tx *bolt.Tx) *CatalogBucket {
	return &CatalogBucket{tx.Bucket(bucketCatalog)}
}

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
