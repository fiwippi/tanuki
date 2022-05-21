package transfer

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/fiwippi/tanuki/internal/mangadex"
)

func TestManager(t *testing.T) {
	ls, err := mangadex.SearchManga(context.TODO(), "hori", 2)
	assert.Nil(t, err)
	assert.NotZero(t, len(ls))

	chs, err := ls[0].ListChapters(context.Background())
	assert.Nil(t, err)
	assert.NotZero(t, len(chs))

	m := NewManager(".", 1)
	m.Queue(ls[0].Title, chs[1])

	for len(m.activeDownloads.l) != 0 {
		time.Sleep(1 * time.Second)
	}
	os.RemoveAll("./Horimiya")
}
