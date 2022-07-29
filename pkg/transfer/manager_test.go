package transfer

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/mangadex"
	"github.com/fiwippi/tanuki/internal/platform/fse"
	"github.com/fiwippi/tanuki/pkg/manga"
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
	log.Disable()
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
	m := NewManager(".", 1, s, func() error { return nil }, 20*time.Second)
	m.Queue(l.Title, ch, true)

	// Wait for download to finish
	for len(m.activeDownloads.l) != 0 {
		t.Log("Page:", m.activeDownloads.l[0].CurrentPage, "/", m.activeDownloads.l[0].TotalPages)
		time.Sleep(3 * time.Second)
	}

	// Check that the folder exists and that it has an SID since a subscription was created
	require.True(t, fse.Exists(fmt.Sprintf("./%s/info.tanuki", l.Title)))
	sid, err := manga.FolderID("./" + l.Title)
	require.Nil(t, err)

	// Check the subscription exists in the db
	sb, err := s.GetSubscription(sid)
	require.Nil(t, err)
	require.Equal(t, sid, sb.SID)
	require.Equal(t, l.ID, string(sb.MdexUUID))
	require.True(t, ch.PublishedAt.Equal(sb.MdexLastPublishedAt))

	// Wait for the subscription checker to start running
	// and then ensure the newly downloaded chapter is
	// published after the first chapter
	for len(m.activeDownloads.l) == 0 {
		t.Log("Waiting for subscription...")
		time.Sleep(3 * time.Second)
	}
	require.True(t, m.activeDownloads.l[0].Chapter.PublishedAt.After(ch.PublishedAt))

	// Wait for the new chapter to finish
	for len(m.activeDownloads.l) != 0 {
		t.Log("Page:", m.activeDownloads.l[0].CurrentPage, "/", m.activeDownloads.l[0].TotalPages)
		time.Sleep(3 * time.Second)
	}

	// Ensure the new chapter downloads in the same folder
	count := 0
	fis, err := ioutil.ReadDir("./" + l.Title)
	require.Nil(t, err)
	for _, fi := range fis {
		if !fi.IsDir() && !strings.Contains(fi.Name(), ".tanuki") {
			count += 1
		}
	}

	// We should have 2 files, the penultimate published chapter and the final one
	require.Equal(t, 2, count)

	// Delete all data
	os.RemoveAll("./" + l.Title)
}
