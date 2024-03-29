package mangadex

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	// SearchManga
	ls, err := SearchManga(context.TODO(), "Hori-san to Miyamura", 2)
	t.Logf("Search listing: %+v\n", ls)
	require.Nil(t, err)
	require.NotZero(t, len(ls))
	require.Contains(t, ls[0].Title, "Hori")

	// ListChapters
	chs, err := ls[0].ListChapters(context.Background())
	t.Logf("List chapters: %+v\n", chs)
	require.Nil(t, err)
	require.NotZero(t, len(chs))

	// NewChapters
	since, _ := time.Parse("2006-01-02", "2020-01-01")
	chsSince, err := ls[0].NewChapters(context.TODO(), since)
	t.Logf("New chapters: %+v\n", chsSince)
	require.Nil(t, err)
	require.NotZero(t, len(chsSince))
	require.NotEqual(t, chs, chsSince)

	// DownloadChapter - choose a chapter which has the fewest amount of pages
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
	require.NotZero(t, len(zF.Bytes()))
}
