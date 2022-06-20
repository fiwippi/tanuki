package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/platform/encryption"
	"github.com/fiwippi/tanuki/pkg/human"
)

// TODO: remove the cgo dependency
// TODO is there a way to reduce similar code, e.g. code used to get covers or thumbnails
// TODO all functions which don't mutate a pointer (not just in storage, should pass by value)
// TODO: string representation of the DB
// TODO: test stuff like user's progress being deleted if the user is deleted
// TODO: SQL VACCUUM MODE

type Store struct {
	libraryPath string
	pool        *sqlx.DB
}

func NewStore(dbPath, libraryPath string, recreate bool) (*Store, error) {
	s := &Store{
		libraryPath: libraryPath,
		pool:        sqlx.MustConnect("sqlite", dbPath+"?_pragma=foreign_keys(on)"),
	}

	// Drop if recreating
	if recreate {
		// We have to delete the tables which depend on other tables first
		// and work our way back to tables which don't depend on anything
		// to satisfy the foreign keys constraint
		for _, t := range []string{"progress", "entries", "series", "users", "downloads"} {
			if _, err := s.pool.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, t)); err != nil {
				return nil, err
			}
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
		display_title     TEXT,
		custom_cover      BLOB,
		custom_cover_type INTEGER,
		
		-- Subscription data
		mangadex_uuid              TEXT,
		mangadex_last_published_at TEXT
	);`
	if _, err := s.pool.Exec(stmt); err != nil {
		return nil, err
	}

	// TODO: test cannot set invalid cover type for series or entries
	// TODO: tests to ensure the custom cover type is set correctly

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
		display_title     TEXT,
		custom_cover      BLOB,
		custom_cover_type INTEGER,

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
		sid     TEXT    NOT NULL,
		eid     TEXT    NOT NULL,
		uid     TEXT    NOT NULL,
		current INTEGER NOT NULL,
		total   INTEGER NOT NULL,
		PRIMARY KEY (sid, eid, uid),
		FOREIGN KEY (sid) 
		    REFERENCES series (sid)
                ON UPDATE CASCADE 
                ON DELETE CASCADE,
		FOREIGN KEY (sid, eid) 
		    REFERENCES entries (sid, eid)
                ON UPDATE CASCADE 
                ON DELETE CASCADE,
		FOREIGN KEY (uid) 
		    REFERENCES users (uid)
                ON UPDATE CASCADE 
                ON DELETE CASCADE 
	) WITHOUT ROWID;`
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
