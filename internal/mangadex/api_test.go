package mangadex

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	// SearchManga
	ls, err := SearchManga(context.TODO(), "hori", 2)
	assert.Nil(t, err)
	assert.NotZero(t, len(ls))

	// ListChapters
	chs, err := ls[0].ListChapters(context.Background())
	assert.Nil(t, err)
	assert.NotZero(t, len(chs))

	// NewChapters
	since, _ := time.Parse("2006-01-02", "2020-01-01")
	chsSince, err := ls[0].NewChapters(context.TODO(), since)
	assert.Nil(t, err)
	assert.NotZero(t, len(chsSince))
	assert.NotEqual(t, chs, chsSince)

	// DownloadChapter
	progress := make(chan int)
	go func() {
		for p := range progress {
			t.Logf("DL Progress: %d/%d\n", p, chs[0].Pages)
		}
	}()
	defer close(progress)
	zF, err := chs[0].DownloadZip(context.TODO(), progress)
	assert.Nil(t, err)
	assert.NotZero(t, len(zF.Data()))
}
