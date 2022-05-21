package manga

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func testSeries(t *testing.T, path, name string, entryCount int, entries []string) {
	fp, err := filepath.Abs(path)
	assert.Nil(t, err)
	s, err := ParseSeries(context.TODO(), fp)
	assert.Nil(t, err)
	assert.NotEqual(t, s.ID, xid.NilID())
	assert.Equal(t, name, s.Title)
	assert.Equal(t, entryCount, len(s.Entries))
	names := make([]string, 0)
	for _, e := range s.Entries {
		names = append(names, e.Title)
	}
	assert.Equal(t, entries, names)
}

func TestParseSeries(t *testing.T) {
	defer os.Remove("../../tests/lib/20th Century Boys/info.tanuki")
	entries := []string{
		"v1",
		"v2",
	}
	testSeries(t, "../../tests/lib/20th Century Boys", "20th Century Boys", 2, entries)

	defer os.Remove("../../tests/lib/Akira/info.tanuki")
	entries = []string{
		"Volume 01",
		"Volume 02",
	}
	testSeries(t, "../../tests/lib/Akira", "Akira", 2, entries)

	defer os.Remove("../../tests/lib/Amano/info.tanuki")
	entries = []string{
		"Amano Megumi wa Suki Darake! v01",
	}
	testSeries(t, "../../tests/lib/Amano", "Amano", 1, entries)
}
