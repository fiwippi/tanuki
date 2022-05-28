package mangadex

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	// SearchManga
	ls, err := SearchManga(context.TODO(), "hori", 2)
	require.Nil(t, err)
	require.NotZero(t, len(ls))

	// ListChapters
	chs, err := ls[0].ListChapters(context.Background())
	require.Nil(t, err)
	require.NotZero(t, len(chs))

	// NewChapters
	since, _ := time.Parse("2006-01-02", "2020-01-01")
	chsSince, err := ls[0].NewChapters(context.TODO(), since)
	require.Nil(t, err)
	require.NotZero(t, len(chsSince))
	require.NotEqual(t, chs, chsSince)

	// DownloadChapter - choose a chapter which has a few amount of pages
	smallCh := chs[0]
	if len(chs) > 1 {
		for _, ch := range chs[1:] {
			if ch.Pages < smallCh.Pages {
				smallCh = ch
			}
		}
	}
	require.NotZero(t, smallCh.Pages)

	progress := make(chan int)
	go func() {
		for p := range progress {
			t.Logf("DL Progress: %d/%d\n", p, smallCh.Pages)
		}
	}()
	zF, err := smallCh.downloadZip(context.TODO(), progress)
	require.Nil(t, err)
	require.NotZero(t, len(zF.Data()))
}
