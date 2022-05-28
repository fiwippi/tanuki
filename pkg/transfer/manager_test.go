package transfer

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/mangadex"
	"github.com/fiwippi/tanuki/pkg/storage"
)

func mustOpenStoreMem(t *testing.T) *storage.Store {
	s, err := storage.NewStore("file::memory:", false)
	require.Nil(t, err)
	return s
}

func TestMain(m *testing.M) {
	log.Disable()
	os.Exit(m.Run())
}

func TestManager(t *testing.T) {
	ls, err := mangadex.SearchManga(context.TODO(), "hori", 2)
	require.Nil(t, err)
	require.NotZero(t, len(ls))

	chs, err := ls[0].ListChapters(context.Background())
	require.Nil(t, err)
	require.NotZero(t, len(chs))

	m := NewManager(".", 1, mustOpenStoreMem(t), func() error { return nil })
	m.Queue(ls[0].Title, chs[1])

	for len(m.activeDownloads.l) != 0 {
		time.Sleep(1 * time.Second)
	}
	os.RemoveAll("./Horimiya")
}
