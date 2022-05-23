package storage

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/platform/encryption"
	"github.com/fiwippi/tanuki/pkg/human"
)

// TODO: test stuff like user's progress being deleted if the user is deleted
// TODO: SQL VACCUUM MODE

type Store struct {
	pool *sqlx.DB
}

func NewStore(path string, recreate bool) (*Store, error) {
	// Create the pool of connections to the DB
	pl := sqlx.MustConnect("sqlite", path)
	s := &Store{pool: pl}

	// Drop if recreating
	if recreate {
		stmt := `
		DROP TABLE IF EXISTS downloads;
		DROP TABLE IF EXISTS users;`
		if _, err := s.pool.Exec(stmt); err != nil {
			return nil, err
		}
	}

	// Create the downloads table
	stmt := `CREATE TABLE IF NOT EXISTS downloads (
    manga_title  TEXT    NOT NULL,
    chapter      BLOB    NOT NULL,
    status       TEXT    NOT NULL,
    current_page INTEGER NOT NULL,
    total_pages  INTEGER NOT NULL,
    time_taken   TEXT    NOT NULL);`
	if _, err := s.pool.Exec(stmt); err != nil {
		return nil, err
	}

	// Create the users table
	stmt = `CREATE TABLE IF NOT EXISTS users (
		uid  TEXT PRIMARY KEY UNIQUE,
		name TEXT NOT NULL    UNIQUE,
		pass TEXT NOT NULL,
		type TEXT NOT NULL
	);`
	if _, err := s.pool.Exec(stmt); err != nil {
		return nil, err
	}

	// Ensure a default user exists in the DB
	hasUsers, err := s.HasUsers()
	if err != nil {
		return nil, err
	}
	if !hasUsers {
		pass := encryption.NewKey(32).Base64()
		err = s.CreateUser(human.NewUser("default", pass, human.Admin))
		if err != nil {
			return nil, err
		}
		log.Info().Str("username", "default").Str("pass", pass).Msg("created default user")
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.pool.Close()
}
