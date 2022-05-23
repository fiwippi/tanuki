package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/platform/hash"
	"github.com/fiwippi/tanuki/pkg/human"
)

var tempFiles = make([]string, 0)
var defaultUID = hash.SHA1("default")

func mustOpenStoreFile(t *testing.T, f *os.File, recreate bool) (*Store, *os.File) {
	var err error
	if f == nil {
		f, err = os.CreateTemp("", "tanuki-store-test")
		require.Nil(t, err)
	}

	s, err := NewStore(f.Name(), recreate)
	require.Nil(t, err)
	return s, f
}

func mustOpenStoreMem(t *testing.T) *Store {
	s, err := NewStore("file::memory:", false)
	require.Nil(t, err)
	return s
}

func mustCloseStore(t *testing.T, s *Store) {
	require.Nil(t, s.Close())
}

func TestMain(m *testing.M) {
	log.Disable()
	code := m.Run()
	for _, f := range tempFiles {
		os.Remove(f)
	}
	os.Exit(code)
}

func TestNewStore(t *testing.T) {
	// Ensure no error on sartup
	s, tf := mustOpenStoreFile(t, nil, false)

	// Default user must exist in the DB
	has, err := s.HasUsers()
	assert.Nil(t, err)
	assert.True(t, has)
	has, err = s.HasUser(defaultUID)
	assert.Nil(t, err)
	assert.True(t, has)

	// Default user should have the right values
	u, err := s.GetUser(defaultUID)
	assert.Nil(t, err)
	assert.Equal(t, defaultUID, u.UID)
	assert.Equal(t, "default", u.Name)
	assert.Equal(t, human.Admin, u.Type)
	mustCloseStore(t, s)
	oldPass := u.Pass

	// If the store is recreated the default user
	// should have a different password
	s, _ = mustOpenStoreFile(t, tf, true)
	u, err = s.GetUser(defaultUID)
	assert.Nil(t, err)
	assert.NotEqual(t, oldPass, u.Pass)
	mustCloseStore(t, s)
	oldPass = u.Pass

	// If opening again without recreation the user should stay the same
	s, _ = mustOpenStoreFile(t, tf, false)
	u, err = s.GetUser(defaultUID)
	assert.Nil(t, err)
	assert.Equal(t, oldPass, u.Pass)
	mustCloseStore(t, s)
}
