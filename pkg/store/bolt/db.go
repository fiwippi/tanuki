package bolt

import (
	"fmt"
	"github.com/fiwippi/tanuki/internal/encryption"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/bolt/util"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
	"github.com/rs/zerolog/log"
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

func connect(path string) (*DB, error) {
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

		_, err = tx.CreateBucketIfNotExists(keys.Downloads)
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

func Startup(path string) (*DB, error) {
	db, err := connect(path)
	if err != nil {
		return nil, err
	}

	// If no users exist then create default user
	if !db.HasUsers() {
		pass := encryption.NewKey(32).Base64()
		err := db.CreateUser(users.NewUser("default", pass, users.Admin))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create default user")
		}
		log.Info().Str("username", "default").Str("pass", pass).Msg("created default user")
	}

	return db, err
}
