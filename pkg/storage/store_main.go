package storage

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/platform/encryption"
	"github.com/fiwippi/tanuki/pkg/human"
)

// TODO is there a way to reduce similar code, e.g. code used to get covers or thumbnails
// TODO all functions which don't mutate a pointer (not just in storage, should pass by value)
// TODO: string representation of the DB
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
		DROP TABLE IF EXISTS users;
		DROP TABLE IF EXISTS series;
		DROP TABLE IF EXISTS entries;
		DROP TABLE IF EXISTS progress;`
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
		time_taken   TEXT    NOT NULL
    );`
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

	// Create the series table
	stmt = `CREATE TABLE IF NOT EXISTS series (
		sid          TEXT    PRIMARY KEY UNIQUE,
		folder_title TEXT    NOT NULL    UNIQUE,
		num_entries  INTEGER NOT NULL,
		num_pages    INTEGER NOT NULL,
		mod_time     TEXT    NOT NULL,
		thumbnail    BLOB,
		tags         TEXT,
		
		-- Custom metadata
		display_title TEXT,
		custom_cover  BLOB,
		
		-- Subscription data
		mangadex_uuid              TEXT,
		mangadex_last_published_at TEXT
	);`
	if _, err := s.pool.Exec(stmt); err != nil {
		return nil, err
	}

	// Create the entries table
	stmt = `CREATE TABLE IF NOT EXISTS entries (
		eid       TEXT NOT NULL,
		sid       TEXT NOT NULL,
		title     TEXT NOT NULL,
		archive   TEXT NOT NULL,
		pages     TEXT NOT NULL,
		mod_time  TEXT NOT NULL,
		position  INTEGER,
		thumbnail BLOB,
		
		-- Custom metadata
		display_title TEXT,
		custom_cover  BLOB,
		
		-- Relationships
		PRIMARY KEY (sid, eid),
		FOREIGN KEY (sid) 
		    REFERENCES series (sid)
                ON UPDATE CASCADE 
                ON DELETE CASCADE 
	);`
	if _, err := s.pool.Exec(stmt); err != nil {
		return nil, err
	}

	// Create the progress table
	stmt = `CREATE TABLE IF NOT EXISTS progress (
		sid     TEXT NOT NULL,
		uid     TEXT NOT NULL,
		data    TEXT NOT NULL,
		PRIMARY KEY (sid, uid),
		FOREIGN KEY (sid) 
		    REFERENCES series (sid)
                ON UPDATE CASCADE 
                ON DELETE CASCADE,
		FOREIGN KEY (uid) 
		    REFERENCES users (uid)
                ON UPDATE CASCADE 
                ON DELETE CASCADE 
	) WITHOUT ROWID ;`
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
		err = s.AddUser(human.NewUser("default", pass, human.Admin), true)
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
