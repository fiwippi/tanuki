package util

import (
	"fmt"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// DumpCursor adds a string representation of the current bucket to
// the given input string, "s"
func DumpCursor(tx *bolt.Tx, c *bolt.Cursor, indent int, s *string) {
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			*s += fmt.Sprintf(strings.Repeat("\t", indent)+"[%s]\n", k)
			newBucket := c.Bucket().Bucket(k)
			if newBucket == nil {
				newBucket = tx.Bucket(k)
			}
			newCursor := newBucket.Cursor()
			DumpCursor(tx, newCursor, indent+1, s)
		} else if string(k) == string(keys.Thumbnail) {
			// Avoid printing out the byte data of the thumbnail
			*s += fmt.Sprintf(strings.Repeat("\t", indent)+"%s\n", k)
			*s += fmt.Sprintf(strings.Repeat("\t", indent+1)+"%s\n", "*EXISTS*")
		} else {
			*s += fmt.Sprintf(strings.Repeat("\t", indent)+"%s\n", k)
			*s += fmt.Sprintf(strings.Repeat("\t", indent+1)+"%s\n", v)
		}
	}
}

func MoveBucket(oldParent, newParent *bolt.Bucket, oldkey, newkey []byte, deleteOld bool) error {
	oldBucket := oldParent.Bucket(oldkey)
	newBucket, err := newParent.CreateBucket(newkey)
	if err != nil {
		return err
	}

	err = oldBucket.ForEach(func(k, v []byte) error {
		if v == nil {
			// Nested bucket
			return RenameBucket(oldBucket, newBucket, k, k)
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

func RenameBucket(oldParent, newParent *bolt.Bucket, oldkey, newkey []byte) error {
	return MoveBucket(oldParent, newParent, oldkey, newkey, true)
}
