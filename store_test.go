package tanuki

import (
	"archive/zip"
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nfnt/resize"
	"github.com/stretchr/testify/require"
)

func TestNewStore(t *testing.T) {
	t.Run("default user created with default values", func(t *testing.T) {
		s, tf := mustOpenStoreFile(t, nil)
		defer tf.Close()
		defer mustCloseStore(t, s)

		u, err := s.GetUser(defaultUsername)
		require.NoError(t, err)
		require.Equal(t, User{Name: defaultUsername, Pass: defaultPasswordHash}, u)
	})

	t.Run("changed user data is preserved", func(t *testing.T) {
		// When the store opens and closes multiple times, the
		// changed password is not overwritten by the control
		// block which checks if the default user must be created
		s, tf := mustOpenStoreFile(t, nil)
		defer tf.Close()

		require.NoError(t, s.ChangeUsername(defaultUsername, "a"))
		require.NoError(t, s.ChangePassword("a", "b"))
		mustCloseStore(t, s)

		for range 25 {
			s, _ := mustOpenStoreFile(t, tf)
			u, err := s.GetUser("a")
			require.NoError(t, err)
			require.Equal(t, User{Name: "a", Pass: Sha256("b")}, u)
			mustCloseStore(t, s)
		}
	})
}

func TestStore_Vacuum(t *testing.T) {
	s, tf := mustOpenStoreFile(t, nil)
	defer mustCloseStore(t, s)

	// Add large amount of data
	for i := 1; i <= 200; i++ {
		if i == 1 || i%100 == 0 {
			t.Log("Adding:", i)
		}

		ss := centurySeries
		ss.SID += strconv.Itoa(i)
		ss.Title = ss.SID
		ee := centuryEntries[0]
		ee.SID = ss.SID
		ee.EID = strings.Repeat("-", i)
		ee.Title = ee.EID

		require.NoError(t, s.tx(func(tx *sqlx.Tx) error {
			if err := s.addSeries(tx, ss, i); err != nil {
				return err
			}
			return s.addEntry(tx, ee, i)
		}))
	}

	fi, err := tf.Stat()
	require.NoError(t, err)
	t.Log("Size (before deletion): ", fi.Size())

	// Delete library data
	require.NoError(t, s.PopulateCatalog(nil))

	fi, err = tf.Stat()
	require.NoError(t, err)
	sizeBef := fi.Size()
	t.Log("Size (after deletion): ", sizeBef)

	t.Log("Vacuuming...")
	require.NoError(t, s.Vacuum())

	fi, err = tf.Stat()
	require.NoError(t, err)
	sizeAft := fi.Size()
	t.Log("Size (vacuumed): ", sizeAft)

	require.Less(t, sizeAft, sizeBef)
	require.NoError(t, tf.Close())
}

// Users

func TestStore_AddUser(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	t.Run("can add new user", func(t *testing.T) {
		require.NoError(t, s.AddUser("a", "b"))
		u, err := s.GetUser("a")
		require.NoError(t, err)
		require.Equal(t, User{Name: "a", Pass: Sha256("b")}, u)
	})

	t.Run("cannot add existing user", func(t *testing.T) {
		require.ErrorContains(t, s.AddUser("a", "b"), "UNIQUE constraint failed")
	})
}

func TestStore_DeleteUser(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	// Test requires an initial store
	// with two users, (we already have
	// the default one on creation)
	require.NoError(t, s.AddUser("a", "b"))

	t.Run("can delete user", func(t *testing.T) {
		require.NoError(t, s.DeleteUser("a"))
		_, err := s.GetUser("a")
		require.ErrorContains(t, err, "no rows in result set")
	})

	t.Run("cannot delete user if last left", func(t *testing.T) {
		require.ErrorIs(t, s.DeleteUser(defaultUsername), errNotEnoughUsers)
	})
}

func TestStore_ChangeUsername(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	t.Run("valid change", func(t *testing.T) {
		require.NoError(t, s.ChangeUsername(defaultUsername, "a"))
		u, err := s.GetUser("a")
		require.NoError(t, err)
		require.Equal(t, User{Name: "a", Pass: defaultPasswordHash}, u)
	})

	t.Run("cannot change to empty name", func(t *testing.T) {
		require.ErrorIs(t, s.ChangeUsername("a", ""), errEmptyUsername)
	})

	t.Run("cannot overwrite existing username", func(t *testing.T) {
		// We need to add a second user to test this
		require.NoError(t, s.AddUser("b", "c"))
		require.ErrorContains(t, s.ChangeUsername("a", "b"), "UNIQUE constraint failed")
	})
}

func TestStore_ChangePassword(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	t.Run("valid change", func(t *testing.T) {
		require.NoError(t, s.ChangePassword(defaultUsername, "b"))
		u, err := s.GetUser(defaultUsername)
		require.NoError(t, err)
		require.Equal(t, User{Name: defaultUsername, Pass: Sha256("b")}, u)
	})

	t.Run("cannot change to empty password", func(t *testing.T) {
		require.ErrorIs(t, s.ChangePassword(defaultUsername, ""), errEmptyPassword)
	})
}

func TestStore_AuthLogin(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	t.Run("correct credentials", func(t *testing.T) {
		require.True(t, s.AuthLogin("default", "tanuki"))
	})
	t.Run("wrong password", func(t *testing.T) {
		require.False(t, s.AuthLogin("default", ""))
	})
	t.Run("wrong username", func(t *testing.T) {
		require.False(t, s.AuthLogin("", "tanuki"))
	})
}

// Entries

func TestStore_AddEntry(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	t.Run("original entry", func(t *testing.T) {
		// Entries cannot exist without their
		// related series
		require.NoError(t, s.AddSeries(Series{SID: "b"}, 1))

		e := Entry{
			EID:      "a",
			SID:      "b",
			Title:    "c",
			Archive:  "d",
			ModTime:  time.Now().Round(0), // Strip the monotonic clock reading
			Pages:    Pages{{Path: "e", Mime: "f"}},
			Filesize: 1000,
		}
		require.NoError(t, s.AddEntry(e, 1))

		ee, err := s.GetEntry(e.SID, e.EID)
		require.NoError(t, err)
		require.Equal(t, e, ee)

		pos, missing := 0, false
		row := s.pool.QueryRow(`SELECT position, missing FROM entries WHERE sid = ? AND eid = ?`, e.SID, e.EID)
		require.NoError(t, row.Err())
		require.NoError(t, row.Scan(&pos, &missing))
		require.Equal(t, 1, pos)
		require.Equal(t, false, missing)
	})

	t.Run("modified entry", func(t *testing.T) {
		e := Entry{
			EID: "a",
			SID: "b",
			// Fields below this comment have been modified
			Title:    "w",
			Archive:  "x",
			ModTime:  time.Now().Round(0), // Strip the monotonic clock reading
			Pages:    Pages{{Path: "y", Mime: "z"}},
			Filesize: 2000,
		}
		require.NoError(t, s.AddEntry(e, 2))

		ee, err := s.GetEntry(e.SID, e.EID)
		require.NoError(t, err)
		require.Equal(t, e, ee)

		pos, missing := 0, false
		row := s.pool.QueryRow(`SELECT position, missing FROM entries WHERE sid = ? AND eid = ?`, e.SID, e.EID)
		require.NoError(t, row.Err())
		require.NoError(t, row.Scan(&pos, &missing))
		require.Equal(t, 2, pos)
		require.Equal(t, false, missing)
	})
}

func TestStore_GetEntry(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	// Entries cannot exist without their
	// related series
	require.NoError(t, s.AddSeries(Series{SID: "b"}, 1))

	e := Entry{
		EID:      "a",
		SID:      "b",
		Title:    "c",
		Archive:  "d",
		ModTime:  time.Now().Round(0), // Strip the monotonic clock reading
		Pages:    Pages{{Path: "e", Mime: "f"}},
		Filesize: 1000,
	}
	require.NoError(t, s.AddEntry(e, 1))

	ee, err := s.GetEntry(e.SID, e.EID)
	require.NoError(t, err)
	require.Equal(t, e, ee)
}

func TestStore_GetEntries(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	// Entries cannot exist without their
	// related series
	require.NoError(t, s.AddSeries(Series{SID: "b"}, 1))

	es := []Entry{
		{
			EID:      "a",
			SID:      "b",
			Title:    "c",
			Archive:  "d",
			ModTime:  time.Now().Round(0), // Strip the monotonic clock reading
			Pages:    Pages{{Path: "e", Mime: "f"}},
			Filesize: 1000,
		},
		{
			EID:      "b",
			SID:      "b",
			Title:    "w",
			Archive:  "x",
			ModTime:  time.Now().Round(0), // Strip the monotonic clock reading
			Pages:    Pages{{Path: "y", Mime: "z"}},
			Filesize: 2000,
		},
		{
			EID:      "c",
			SID:      "b",
			Title:    "r",
			Archive:  "s",
			ModTime:  time.Now().Round(0), // Strip the monotonic clock reading
			Pages:    Pages{{Path: "t", Mime: "u"}},
			Filesize: 3000,
		},
	}

	// Add them in reverse order, but they should
	// still be returned in their sorted order
	require.NoError(t, s.AddEntry(es[2], 3))
	require.NoError(t, s.AddEntry(es[1], 2))
	require.NoError(t, s.AddEntry(es[0], 1))

	ees, err := s.GetEntries("b")
	require.NoError(t, err)
	require.Equal(t, es, ees)
}

func TestStore_GetPage(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	path := "tests/lib/Akira/Volume 01.zip"
	e, err := ParseEntry(path)
	require.NoError(t, err)

	t.Run("add the series and entry", func(t *testing.T) {
		require.NoError(t, s.AddSeries(Series{SID: e.SID}, 1))
		require.NoError(t, s.AddEntry(e, 1))
	})

	t.Run("page out of bounds", func(t *testing.T) {
		_, _, err := s.GetPage(e.SID, e.EID, -1)
		require.Error(t, err)
		_, _, err = s.GetPage(e.SID, e.EID, len(e.Pages))
		require.Error(t, err)
	})

	t.Run("page in bounds", func(t *testing.T) {
		r, err := zip.OpenReader(e.Archive)
		require.NoError(t, err)
		defer r.Close()

		pageData := make(map[string][]byte)
		for _, f := range r.File {
			fr, err := f.Open()
			require.NoError(t, err)
			data, err := io.ReadAll(fr)
			require.NoError(t, err)
			pageData[f.Name] = data
		}

		for i, p := range e.Pages {
			data, mime, err := s.GetPage(e.SID, e.EID, i)
			require.NoError(t, err)
			require.Equal(t, pageData[p.Path], data.Bytes())
			require.Equal(t, p.Mime, mime)
		}
	})
}

func TestStore_GetThumbnail(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	path := "tests/lib/Akira/Volume 01.zip"
	e, err := ParseEntry(path)
	require.NoError(t, err)

	t.Run("add the series and entry", func(t *testing.T) {
		require.NoError(t, s.AddSeries(Series{SID: e.SID}, 1))
		require.NoError(t, s.AddEntry(e, 1))
	})

	origThumb := bytes.NewBuffer(nil)
	newThumb := bytes.NewBuffer(nil)
	t.Run("generate example thumbnail", func(t *testing.T) {
		r, err := zip.OpenReader(e.Archive)
		require.NoError(t, err)
		defer r.Close()

		p, _, err := s.GetPage(e.SID, e.EID, 0)
		require.NoError(t, err)
		img, _, err := image.Decode(p)
		require.NoError(t, err)
		thumb := resize.Thumbnail(300, 300, img, resize.Bicubic)
		require.NoError(t, jpeg.Encode(origThumb, thumb, &jpeg.Options{Quality: 70}))

		p, _, err = s.GetPage(e.SID, e.EID, 1)
		require.NoError(t, err)
		img, _, err = image.Decode(p)
		require.NoError(t, err)
		thumb = resize.Thumbnail(300, 300, img, resize.Bicubic)
		require.NoError(t, jpeg.Encode(newThumb, thumb, &jpeg.Options{Quality: 70}))
	})

	t.Run("thumbnail generated", func(t *testing.T) {
		buf, mime, err := s.GetThumbnail(e.SID, e.EID)
		require.NoError(t, err)
		require.Equal(t, "image/jpeg", mime)
		require.Equal(t, origThumb, buf)
	})

	t.Run("thumbnail retrieved", func(t *testing.T) {
		buf, mime, err := s.GetThumbnail(e.SID, e.EID)
		require.NoError(t, err)
		require.Equal(t, "image/jpeg", mime)
		require.Equal(t, origThumb, buf)
	})

	t.Run("new thumbnail on changed mod time", func(t *testing.T) {
		// We remove the original cover page, so now the second
		// page becomes the new cover page. Since the mod time
		// has changed for this entry we also expect the thumbnail
		// to be regenerated and equal the second page
		e.ModTime = time.Now().Round(0)
		e.Pages = e.Pages[1:]
		require.NoError(t, s.AddEntry(e, 1))

		buf, mime, err := s.GetThumbnail(e.SID, e.EID)
		require.NoError(t, err)
		require.Equal(t, "image/jpeg", mime)
		require.Equal(t, newThumb, buf)
	})

	t.Run("thumbnails removed on deletion", func(t *testing.T) {
		// A nil population deletes all current
		// series and entries
		require.NoError(t, s.PopulateCatalog(nil))

		var exists bool
		require.NoError(t, s.pool.Get(&exists, "SELECT COUNT(*) > 0 FROM thumbnails"))
		require.False(t, exists)
	})
}

// Series

func TestStore_AddSeries(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	t.Run("original series", func(t *testing.T) {
		sr := Series{
			SID:     "a",
			Title:   "b",
			ModTime: time.Now().Round(0), // Strip the monotonic clock reading
		}
		require.NoError(t, s.AddSeries(sr, 1))

		ssr, err := s.GetCatalog()
		require.NoError(t, err)
		require.Equal(t, []Series{sr}, ssr)

		pos, missing := 0, false
		row := s.pool.QueryRow(`SELECT position, missing FROM series WHERE sid = ?`, sr.SID)
		require.NoError(t, row.Err())
		require.NoError(t, row.Scan(&pos, &missing))
		require.Equal(t, 1, pos)
		require.Equal(t, false, missing)
	})

	t.Run("modified entry", func(t *testing.T) {
		sr := Series{
			SID: "a",
			// Fields below this comment have been modified
			Title:   "z",
			ModTime: time.Now().Round(0), // Strip the monotonic clock reading
		}
		require.NoError(t, s.AddSeries(sr, 2))

		ssr, err := s.GetCatalog()
		require.NoError(t, err)
		require.Equal(t, []Series{sr}, ssr)

		pos, missing := 0, false
		row := s.pool.QueryRow(`SELECT position, missing FROM series WHERE sid = ?`, sr.SID)
		require.NoError(t, row.Err())
		require.NoError(t, row.Scan(&pos, &missing))
		require.Equal(t, 2, pos)
		require.Equal(t, false, missing)
	})
}

func TestStore_GetSeries(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	srs := []Series{
		{
			SID:     "a",
			Title:   "b",
			ModTime: time.Now().Round(0), // Strip the monotonic clock reading
		},
		{
			SID:     "c",
			Title:   "d",
			ModTime: time.Now().Round(0), // Strip the monotonic clock reading
		},
		{
			SID:     "e",
			Title:   "f",
			ModTime: time.Now().Round(0), // Strip the monotonic clock reading
		},
	}

	require.NoError(t, s.AddSeries(srs[0], 1))
	require.NoError(t, s.AddSeries(srs[1], 2))
	require.NoError(t, s.AddSeries(srs[2], 3))

	for i := 0; i <= 2; i++ {
		srss, err := s.GetSeries(srs[i].SID)
		require.NoError(t, err)
		require.Equal(t, srs[i], srss)
	}
}

// Catalog

func TestStore_GetCatalog(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	srs := []Series{
		{
			SID:     "a",
			Title:   "b",
			ModTime: time.Now().Round(0), // Strip the monotonic clock reading
		},
		{
			SID:     "c",
			Title:   "d",
			ModTime: time.Now().Round(0), // Strip the monotonic clock reading
		},
		{
			SID:     "e",
			Title:   "f",
			ModTime: time.Now().Round(0), // Strip the monotonic clock reading
		},
	}

	// Add them in reverse order, but they should
	// still be returned in their sorted order
	require.NoError(t, s.AddSeries(srs[2], 3))
	require.NoError(t, s.AddSeries(srs[1], 2))
	require.NoError(t, s.AddSeries(srs[0], 1))

	srss, err := s.GetCatalog()
	require.NoError(t, err)
	require.Equal(t, srs, srss)
}

func TestStore_PopulateCatalog(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	lib, err := ParseLibrary("tests/lib")
	require.NoError(t, err)

	t.Run("data added", func(t *testing.T) {
		require.NoError(t, s.PopulateCatalog(lib))

		t.Run("series", func(t *testing.T) {
			ctl, err := s.GetCatalog()
			require.NoError(t, err)
			require.Equal(t, []Series{centurySeries, akiraSeries, amanoSeries}, ctl)
		})
		t.Run("entries", func(t *testing.T) {
			entries, err := s.GetEntries("PvHfuhL24GD6jo-PKLbPj_KvRikLn2WjCw_gOaXKRyI")
			require.NoError(t, err)
			require.Equal(t, centuryEntries, entries)
			entries, err = s.GetEntries("rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c")
			require.NoError(t, err)
			require.Equal(t, akiraEntries, entries)
			entries, err = s.GetEntries("wNgocaIzfIjmFcxC-5I3S5pEpjRKjDY4nRxg9Ko-z7k")
			require.NoError(t, err)
			require.Equal(t, amanoEntries, entries)
		})
	})

	t.Run("data removed", func(t *testing.T) {
		delete(lib, centurySeries)
		delete(lib, amanoSeries)
		lib[akiraSeries] = lib[akiraSeries][:1]
		require.NoError(t, s.PopulateCatalog(lib))

		t.Run("series", func(t *testing.T) {
			ctl, err := s.GetCatalog()
			require.NoError(t, err)
			require.Equal(t, []Series{akiraSeries}, ctl)
		})
		t.Run("entries", func(t *testing.T) {
			entries, err := s.GetEntries("rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c")
			require.NoError(t, err)
			require.Equal(t, []Entry{akiraEntries[0]}, entries)
		})
	})
}

// Helpers

func (s *Store) GetUser(name string) (User, error) {
	var u User
	return u, s.pool.Get(&u, `SELECT name, pass FROM users WHERE name = ?`, name)
}

func (s *Store) AddSeries(sr Series, position int) error {
	return s.tx(func(tx *sqlx.Tx) error {
		return s.addSeries(tx, sr, position)
	})
}

func (s *Store) AddEntry(e Entry, position int) error {
	return s.tx(func(tx *sqlx.Tx) error {
		return s.addEntry(tx, e, position)
	})
}

// Utils

func TestMain(m *testing.M) {
	// Run the tests
	code := m.Run()

	// Remove any generated stores
	for _, f := range tempFiles {
		if err := os.Remove(f); err != nil {
			slog.Error("Failed to delete temp store", slog.Any("err", err))
		}
	}
	os.Exit(code)
}

var tempFiles = make([]string, 0)

func mustOpenStoreFile(t *testing.T, f *os.File) (*Store, *os.File) {
	var err error
	if f == nil {
		f, err = os.CreateTemp("", "tanuki-store-test")
		require.Nil(t, err)
		tempFiles = append(tempFiles, f.Name())
	}

	s, err := NewStore(f.Name())
	require.Nil(t, err)
	return s, f
}

func mustOpenStoreMem(t *testing.T) *Store {
	s, err := NewStore(InMemory)
	require.Nil(t, err)
	return s
}

func mustCloseStore(t *testing.T, s *Store) {
	require.Nil(t, s.Close())
}
