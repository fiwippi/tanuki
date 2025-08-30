package tanuki

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
	"github.com/maruel/natural"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
	_ "modernc.org/sqlite"
)

// Store

const InMemory string = "file::memory:"

type Store struct {
	pool *sqlx.DB
}

func NewStore(path string) (*Store, error) {
	pool, err := sqlx.Connect("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", path, err)
	}
	pool.SetMaxOpenConns(1)
	pool.MapperFunc(func(s string) string {
		// For example, ModTime --> mod_time
		b := strings.Builder{}
		for i, r := range s {
			if unicode.IsUpper(r) && i > 0 {
				// Only split if the following letter is also not uppercase
				if i+1 < len(s) && !unicode.IsUpper(rune(s[i+1])) {
					b.WriteRune('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		}
		return b.String()
	})

	s := &Store{pool: pool}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			name TEXT PRIMARY KEY UNIQUE,
			pass TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS series (
			sid       TEXT     PRIMARY KEY UNIQUE,
			title     TEXT     NOT NULL    UNIQUE,
			author    TEXT,
			mod_time  DATETIME NOT NULL,
			position  INTEGER  NOT NULL,
		    missing   INTEGER  NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS entries (
			eid       TEXT     NOT NULL,
			sid       TEXT     NOT NULL,
			title     TEXT     NOT NULL,
			mod_time  DATETIME NOT NULL,
			archive   TEXT     NOT NULL,
			pages     TEXT     NOT NULL,
			filesize  INTEGER  NOT NULL,
			position  INTEGER  NOT NULL,
		    missing   INTEGER  NOT NULL,

			-- Relationships
			PRIMARY KEY (sid, eid),
			FOREIGN KEY (sid) 
		    	REFERENCES series (sid)
                	ON UPDATE CASCADE 
                	ON DELETE CASCADE 
			);`,
		`CREATE TABLE IF NOT EXISTS thumbnails (
			eid       TEXT     NOT NULL,
			sid       TEXT     NOT NULL,
			mod_time  DATETIME NOT NULL,
			data      BLOB     NOT NULL,

			-- Relationships
			PRIMARY KEY (sid, eid),
			FOREIGN KEY (sid, eid)
				REFERENCES entries (sid, eid)
                	ON UPDATE CASCADE 
                	ON DELETE CASCADE
			);`,
	}
	for _, stmt := range stmts {
		if _, err := s.pool.Exec(stmt); err != nil {
			return nil, err
		}
	}

	var exists bool
	if err := s.pool.Get(&exists, `SELECT COUNT(*) > 0 FROM users`); err != nil {
		return nil, err
	}
	if !exists {
		if err := s.AddUser(defaultUsername, defaultPassword); err != nil {
			return nil, err
		}
		slog.Info("Created default user",
			slog.String("username", defaultUsername), slog.String("password", defaultPassword))
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.pool.Close()
}

func (s *Store) Vacuum() error {
	_, err := s.pool.Exec("VACUUM")
	if err != nil {
		return err
	}
	_, err = s.pool.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	return err
}

func (s *Store) Dump() (string, error) {
	dump := ""

	return dump, s.tx(func(tx *sqlx.Tx) error {
		var names []string
		err := tx.Select(&names, `SELECT name FROM sqlite_master WHERE type = 'table'`)
		if err != nil {
			return err
		}

		for _, name := range names {
			// We don't want useless byte output
			if name == "thumbnails" {
				continue
			}

			dump += fmt.Sprintf("%s\n%s\n", name, strings.Repeat("-", utf8.RuneCountInString(name)))
			rows, err := tx.Queryx(fmt.Sprintf(`SELECT * FROM %s`, name))
			if err != nil {
				return err
			}

			i := 1
			for rows.Next() {
				results := make(map[string]any)
				err = rows.MapScan(results)
				if err != nil {
					return err
				}
				delete(results, "pages") // Too long
				dump += fmt.Sprintf("%d. %s\n", i, results)
				i++
			}
			dump += "\n\n"
		}

		return nil
	})
}

// Helpers

func (s *Store) tx(fn func(tx *sqlx.Tx) error) error {
	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Users

const (
	defaultUsername = "default"
	defaultPassword = "tanuki"
)

var (
	defaultPasswordHash = Sha256(defaultPassword)

	errEmptyUsername    = fmt.Errorf("username cannot be empty")
	errEmptyPassword    = fmt.Errorf("password cannot be empty")
	errNotEnoughUsers   = fmt.Errorf("not enough users left in store")
	errUserDoesNotExist = fmt.Errorf("user does not exist")
)

type User struct {
	Name string
	Pass string // Hashed
}

func (s *Store) AddUser(name, pass string) error {
	_, err := s.pool.Exec(`INSERT INTO users (name, pass) Values (?,?)`, name, Sha256(pass))
	return err
}

func (s *Store) DeleteUser(name string) error {
	return s.tx(func(tx *sqlx.Tx) error {
		var exists bool
		if err := tx.Get(&exists, `SELECT EXISTS(SELECT 1 FROM users WHERE name = ?)`, name); err != nil {
			return err
		}
		if !exists {
			return errUserDoesNotExist
		}

		var num int
		if err := tx.Get(&num, `SELECT COUNT(*) FROM users`); err != nil {
			return err
		}
		if num-1 == 0 {
			return errNotEnoughUsers
		}

		_, err := tx.Exec(`DELETE FROM users WHERE name = ?`, name)
		return err
	})
}

func (s *Store) ChangeUsername(oldName, newName string) error {
	if newName == "" {
		return errEmptyUsername
	}
	_, err := s.pool.Exec(`UPDATE users SET name = ? WHERE name = ?`, newName, oldName)
	return err
}

func (s *Store) ChangePassword(name, pass string) error {
	if pass == "" {
		return errEmptyPassword
	}
	_, err := s.pool.Exec(`UPDATE users SET pass = ? WHERE name = ?`, Sha256(pass), name)
	return err
}

func (s *Store) AuthLogin(name, pass string) bool {
	var valid bool
	err := s.pool.Get(&valid, `SELECT pass = ? FROM users WHERE name = ?`, Sha256(pass), name)
	if err != nil {
		return false
	}
	return valid
}

// Series

func (s *Store) addSeries(tx *sqlx.Tx, sr Series, position int) error {
	stmt := `INSERT INTO series (sid, title, author, mod_time, position, missing) 
			 Values (?, ?, ?, ?, ?, 0)
			 ON CONFLICT (sid)
			 DO UPDATE SET sid=excluded.sid, title=excluded.title, author=excluded.author,
						   mod_time=excluded.mod_time, position=excluded.position, 
                           missing=excluded.missing`
	_, err := tx.Exec(stmt, sr.SID, sr.Title, sr.Author, sr.ModTime, position)
	return err
}

func (s *Store) GetSeries(sid string) (Series, error) {
	var v Series
	return v, s.pool.Get(&v, `SELECT sid, title, author, mod_time FROM series 
		     				  WHERE sid = ?`, sid)
}

// Entries

func (s *Store) addEntry(tx *sqlx.Tx, e Entry, position int) error {
	stmt := `INSERT INTO entries (eid, sid, title, archive, pages, mod_time, filesize, position, missing) 
			 Values (?, ?, ?, ?, ?, ?, ?, ?, 0)
			 ON CONFLICT (eid, sid)
			 DO UPDATE SET eid=excluded.eid, sid=excluded.sid, title=excluded.title, archive=excluded.archive,
				           pages=excluded.pages, mod_time=excluded.mod_time, filesize=excluded.filesize,
						   position=excluded.position, missing=excluded.missing`
	_, err := tx.Exec(stmt, e.EID, e.SID, e.Title, e.Archive, e.Pages, e.ModTime, e.Filesize, position)
	return err
}

func (s *Store) getEntry(tx *sqlx.Tx, sid, eid string) (Entry, error) {
	var e Entry
	return e, tx.Get(&e, `SELECT eid, sid, title, mod_time, archive, filesize, pages
                          FROM entries WHERE sid = ? AND eid = ?`, sid, eid)
}

func (s *Store) GetEntry(sid, eid string) (Entry, error) {
	var e Entry
	return e, s.tx(func(tx *sqlx.Tx) error {
		var err error
		e, err = s.getEntry(tx, sid, eid)
		return err
	})
}

func (s *Store) GetEntries(sid string) ([]Entry, error) {
	stmt := `SELECT sid, eid, title, mod_time, archive, filesize, pages FROM entries
			 WHERE sid = ? ORDER BY position ASC, ROWID DESC `

	var es []Entry
	return es, s.pool.Select(&es, stmt, sid)
}

func (s *Store) getPage(tx *sqlx.Tx, sid, eid string, pageNum int) (*bytes.Buffer, string, error) {
	var archive string
	var ps Pages

	row := tx.QueryRow("SELECT archive, pages FROM entries WHERE sid = ? AND eid = ?", sid, eid)
	if err := row.Scan(&archive, &ps); err != nil {
		return nil, "", err
	}

	// Pages are zero-indexed
	if pageNum < 0 || pageNum >= len(ps) {
		return nil, "", fmt.Errorf("page num out of range")
	}
	p := ps[pageNum]

	r, err := zip.OpenReader(archive)
	if err != nil {
		return nil, "", err
	}
	defer r.Close()

	// If the path was originally non-UTF-8 encoded then we
	// can't directly Open the path, since it doesn't exist
	// under the UTF-8 name. Instead we need to do a page
	// by page comparison in the ZIP file... blegh!
	var f io.ReadCloser
	if !p.NonUtf8 {
		f, err = r.Open(p.Path)
	} else {
		for _, f2 := range r.File {
			if !f2.NonUTF8 {
				continue
			}
			f2Name, err := decodeCP437(f2.Name)
			if err != nil {
				return nil, "", err
			}
			if p.Path == f2Name {
				f, err = f2.Open()
				goto FoundFile
			}
		}
		return nil, "", fmt.Errorf("non-utf-8 page not found")
	}
FoundFile:
	if err != nil {
		return nil, "", err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, "", err
	}
	buf := bytes.NewBuffer(content)

	return buf, p.Mime, nil
}

func (s *Store) GetPage(sid, eid string, pageNum int) (*bytes.Buffer, string, error) {
	var mime string
	var buf *bytes.Buffer
	return buf, mime, s.tx(func(tx *sqlx.Tx) error {
		var err error
		buf, mime, err = s.getPage(tx, sid, eid, pageNum)
		return err
	})
}

func (s *Store) GetThumbnail(sid, eid string) (*bytes.Buffer, string, error) {
	buf := bytes.NewBuffer(nil)
	return buf, "image/jpeg", s.tx(func(tx *sqlx.Tx) error {
		e, err := s.getEntry(tx, sid, eid)
		if err != nil {
			return err
		}

		var thumbModTime time.Time
		err = tx.Get(&thumbModTime, `SELECT mod_time FROM thumbnails WHERE sid = ? AND eid = ?`, sid, eid)
		if err == nil && e.ModTime.Equal(thumbModTime) {
			var data []byte
			err = tx.Get(&data, `SELECT data FROM thumbnails WHERE sid = ? AND eid = ?`, sid, eid)
			buf.Write(data)
			return err
		}

		a := !thumbModTime.Equal(time.Time{}) && !thumbModTime.Equal(e.ModTime)
		b := err != nil && strings.Contains(err.Error(), "no rows in result set")
		if a || b {
			p, _, err := s.getPage(tx, sid, eid, 0)
			if err != nil {
				return err
			}
			img, _, err := image.Decode(p)
			if err != nil {
				return err
			}
			thumb := resize.Thumbnail(300, 300, img, resize.Bicubic)
			err = jpeg.Encode(buf, thumb, &jpeg.Options{Quality: 70})
			if err != nil {
				return err
			}

			stmt := `INSERT INTO thumbnails (sid, eid, mod_time, data) 
			 		 Values (?, ?, ?, ?)
			 	     ON CONFLICT (sid, eid)
					 DO UPDATE SET sid=excluded.sid, eid=excluded.eid, 
				     	           mod_time=excluded.mod_time, data=excluded.data`
			_, err = tx.Exec(stmt, e.SID, e.EID, e.ModTime, buf.Bytes())
			return err
		}

		return err
	})
}

// Catalog

func (s *Store) PopulateCatalog(input map[Series][]Entry) error {
	return s.tx(func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE series SET missing=1`)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE entries SET missing=1`)
		if err != nil {
			return err
		}

		// We need to sort our series input
		// since maps don't iterate in sorted
		// order
		ordered := make([]Series, 0)
		for series := range input {
			ordered = append(ordered, series)
		}
		sort.SliceStable(ordered, func(i, j int) bool {
			return natural.Less(ordered[i].Title, ordered[j].Title)
		})

		for i, series := range ordered {
			if err := s.addSeries(tx, series, i+1); err != nil {
				return err
			}
			for j, entry := range input[series] {
				if err := s.addEntry(tx, entry, j+1); err != nil {
					return err
				}
			}
		}

		_, err = tx.Exec(`DELETE FROM series WHERE missing=1`)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`DELETE FROM entries WHERE missing=1`)
		return err
	})
}

func (s *Store) GetCatalog() ([]Series, error) {
	stmt := `SELECT sid, title, author, mod_time FROM series 
		     WHERE missing=0 ORDER BY position ASC, ROWID DESC`

	var v []Series
	if err := s.pool.Select(&v, stmt); err != nil {
		return nil, err
	}

	return v, nil
}
