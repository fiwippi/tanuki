package db

import (
	"fmt"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (db *DB) String() string {
	var s string
	err := db.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		dumpCursor(tx, c, 0, &s)
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("error when string on db: %s", err))
	}
	return s
}

func dumpCursor(tx *bolt.Tx, c *bolt.Cursor, indent int, s *string) {
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			*s += fmt.Sprintf(strings.Repeat("\t", indent)+"[%s]\n", k)
			newBucket := c.Bucket().Bucket(k)
			if newBucket == nil {
				newBucket = tx.Bucket(k)
			}
			newCursor := newBucket.Cursor()
			dumpCursor(tx, newCursor, indent+1, s)
		} else if string(k) == string(keyThumbnail) {
			*s += fmt.Sprintf(strings.Repeat("\t", indent)+"%s\n", k)
			*s += fmt.Sprintf(strings.Repeat("\t", indent+1)+"%s\n", "*EXISTS*")
		} else {
			*s += fmt.Sprintf(strings.Repeat("\t", indent)+"%s\n", k)
			*s += fmt.Sprintf(strings.Repeat("\t", indent+1)+"%s\n", v)
		}
	}
}
