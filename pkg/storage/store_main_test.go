package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/platform/hash"
	"github.com/fiwippi/tanuki/pkg/human"
	"github.com/fiwippi/tanuki/pkg/manga"
)

const (
	dbPath  = "../../tests/data/tanuki.db"
	libPath = "../../tests/lib"
)

// TODO remove ROWID from some rows

var tempFiles = make([]string, 0)
var defaultUID = hash.SHA1("default")
var parsedData []struct {
	s *manga.Series
	e []*manga.Entry
}
var customCover []byte

func mustOpenStoreFile(t *testing.T, f *os.File, recreate bool) (*Store, *os.File) {
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

func mustOpenStoreMem(t *testing.T) *Store {
	s, err := NewStore("file::memory:", libPath, false)
	require.Nil(t, err)
	return s
}

func mustCloseStore(t *testing.T, s *Store) {
	require.Nil(t, s.Close())
}

func TestMain(m *testing.M) {
	log.Disable()

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
			s *manga.Series
			e []*manga.Entry
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
	require.Equal(t, human.Admin, u.Type)
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
