package transfer

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/storage"
)

func mustOpenStoreMem(t *testing.T) *storage.Store {
	s, err := storage.NewStore("file::memory:", ".", false)
	require.Nil(t, err)
	return s
}

func mustCloseStore(t *testing.T, s *storage.Store) {
	require.Nil(t, s.Close())
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestManager(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	// Get the manga
	ls, err := mangadex.SearchManga(context.TODO(), "horimi", 2)
	require.Nil(t, err)
	require.NotZero(t, len(ls))
	l := ls[0]

	// Ensure the title is removed before the download starts
	os.RemoveAll("./" + l.Title)

	// Get its chapters
	chs, err := l.ListChapters(context.Background())
	require.Nil(t, err)
	require.NotZero(t, len(chs))

	// Select the penultimate published chapter
	sort.Slice(chs, func(i, j int) bool {
		return chs[i].PublishedAt.After(chs[j].PublishedAt)
	})
	ch := chs[1]
	require.True(t, ch.Pages <= 10)

	// Queue the download
	m := NewManager(".", 1, s, func() error { return nil })
	m.Queue(l.Title, ch, true)

	// Wait for download to finish
	for len(m.activeDownloads.l) != 0 {
		t.Log("Page:", m.activeDownloads.l[0].CurrentPage, "/", m.activeDownloads.l[0].TotalPages)
		time.Sleep(3 * time.Second)
	}

	// Delete all data
	os.RemoveAll("./" + l.Title)
}
