package storage

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/internal/sqlutil"
	"github.com/fiwippi/tanuki/pkg/manga"
	"github.com/fiwippi/tanuki/pkg/user"
)

const (
	dbPath  = "../../tests/data/tanuki.db"
	libPath = "../../tests/lib"
)

var tempFiles = make([]string, 0)
var defaultUID = hash.SHA1("default")
var parsedData []struct {
	s manga.Series
	e []manga.Entry
}
var customCover []byte

func mustOpenStoreFile(t require.TestingT, f *os.File, recreate bool) (*Store, *os.File) {
	var err error
	if f == nil {
		f, err = os.CreateTemp("", "tanuki-store-test")
		require.Nil(t, err)
		tempFiles = append(tempFiles, f.Name())
	}

	s, err := NewStore(f.Name(), libPath, recreate)
	require.Nil(t, err)
	return s, f
}

func mustOpenStoreMem(t require.TestingT) *Store {
	s, err := NewStore("file::memory:", libPath, false)
	require.Nil(t, err)
	return s
}

func mustCloseStore(t require.TestingT, s *Store) {
	require.Nil(t, s.Close())
}

func TestMain(m *testing.M) {
	// Make an example db in tests/data/tanuki.db which
	// can be read by the IDE for linting etc.
	os.Remove(dbPath)
	_, err := NewStore(dbPath, libPath, true)
	if err != nil {
		panic(err)
	}

	// Parse example data which can be used in testing
	paths := []string{"../../tests/lib/20th Century Boys", "../../tests/lib/Akira", "../../tests/lib/Amano"}
	for _, p := range paths {
		series, entries, err := manga.ParseSeries(context.Background(), p)
		if err != nil {
			panic(err)
		}
		parsedData = append(parsedData, struct {
			s manga.Series
			e []manga.Entry
		}{s: series, e: entries})
	}

	// Read in the custom cover
	customCover, err = os.ReadFile("../../tests/data/customcover.png")
	if err != nil {
		panic(err)
	}
	if len(customCover) == 0 {
		panic("nil custom cover read in for example test")
	}

	// Run the tests
	code := m.Run()
	for _, f := range tempFiles {
		os.Remove(f)
	}
	os.Exit(code)
}

func TestNewStore(t *testing.T) {
	// Ensure no error on sartup
	s, tf := mustOpenStoreFile(t, nil, false)
	defer tf.Close()

	// Default user must exist in the DB
	has, err := s.HasUsers()
	require.Nil(t, err)
	require.True(t, has)
	has, err = s.HasUser(defaultUID)
	require.Nil(t, err)
	require.True(t, has)

	// Default user should have the right values
	u, err := s.GetUser(defaultUID)
	require.Nil(t, err)
	require.Equal(t, defaultUID, u.UID)
	require.Equal(t, "default", u.Name)
	require.Equal(t, user.Admin, u.Type)
	mustCloseStore(t, s)
	oldPass := u.Pass

	// If the store is recreated the default user
	// should have a different password
	s, _ = mustOpenStoreFile(t, tf, true)
	u, err = s.GetUser(defaultUID)
	require.Nil(t, err)
	require.NotEqual(t, oldPass, u.Pass)
	mustCloseStore(t, s)
	oldPass = u.Pass

	// If opening again without recreation the user should stay the same
	s, _ = mustOpenStoreFile(t, tf, false)
	u, err = s.GetUser(defaultUID)
	require.Nil(t, err)
	require.Equal(t, oldPass, u.Pass)
	mustCloseStore(t, s)
}

func TestVacuum(t *testing.T) {
	s, tf := mustOpenStoreFile(t, nil, false)
	defer mustCloseStore(t, s)

	// Add large amount of data
	require.Nil(t, s.AddSeries(parsedData[0].s, nil))
	for i := 1; i <= 200; i++ {
		if i == 1 || i%100 == 0 {
			t.Log("Adding:", i)
		}

		sid := parsedData[0].s.SID
		eid := strings.Repeat("-", i)

		fn := func(tx *sqlx.Tx) error {
			err := s.addEntry(tx, manga.Entry{
				SID:     sid,
				EID:     eid,
				Title:   eid,
				Archive: manga.Archive{},
				Pages:   nil,
				ModTime: sqlutil.Time{},
			}, i)
			return err
		}
		require.Nil(t, s.tx(fn))
		require.Nil(t, s.SetEntryCover(sid, eid, "cover.jpg", customCover))
	}

	// Check size of file
	fi, err := tf.Stat()
	require.Nil(t, err)
	t.Log("Size: ", fi.Size())

	// Delete the entries
	for i := 1; i <= 200; i++ {
		if i == 1 || i%100 == 0 {
			t.Log("Deleting:", i)
		}

		sid := parsedData[0].s.SID
		eid := strings.Repeat("-", i)

		fn := func(tx *sqlx.Tx) error {
			err := s.deleteEntry(tx, sid, eid)
			return err
		}
		require.Nil(t, s.tx(fn))
	}

	// Check size of DB
	fi, err = tf.Stat()
	require.Nil(t, err)
	sizeBef := fi.Size()
	t.Log("Size: ", sizeBef)

	// Run Vacuum
	t.Log("Vacuuming...")
	require.Nil(t, s.Vacuum())

	// Check size of DB
	fi, err = tf.Stat()
	require.Nil(t, err)
	sizeAft := fi.Size()
	t.Log("Size: ", sizeAft)

	// Size after should be less than size before
	require.Less(t, sizeAft, sizeBef)

	require.Nil(t, tf.Close())
}

func TestStore_CanReadConcurrently(t *testing.T) {
	// Files are used when testing concurrency
	s, tf := mustOpenStoreFile(t, nil, false)
	defer tf.Close()
	defer mustCloseStore(t, s)

	require.Nil(t, s.PopulateCatalog())

	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100000; i++ {
				_, err := s.GetCatalog()
				require.Nil(t, err)
			}
		}()
	}
	wg.Wait()
}

func TestStore_CanReadAndWriteConcurrently(t *testing.T) {
	// Files are used when testing concurrency
	s, tf := mustOpenStoreFile(t, nil, false)
	defer tf.Close()
	defer mustCloseStore(t, s)

	require.Nil(t, s.PopulateCatalog())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)

		t.Log("Getting catalog...")
		for i := 0; i < 8000; i++ {
			_, err := s.GetCatalog()
			require.Nil(t, err)
		}
		t.Log("Got catalog")
	}()

	t.Log("Writing...")
	start := time.Now()
	fn := func(tx *sqlx.Tx) error {
		var sids []string
		tx.Select(&sids, `SELECT sid FROM series`)
		require.NotZero(t, len(sids))

		for i := 0; i < 1250; i++ {
			for _, sid := range sids {
				_, err := s.generateSeriesThumbnail(tx, sid, true)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
	require.Nil(t, s.tx(fn))
	t.Log("Done writing...", time.Since(start).String())

	wg.Wait()
}
