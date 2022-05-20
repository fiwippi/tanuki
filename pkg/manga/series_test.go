package manga

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testSeries(t *testing.T, path, name string, entryCount int, entries []string) {
	fp, err := filepath.Abs(path)
	assert.Nil(t, err)
	s, err := ParseSeries(context.TODO(), fp)
	assert.Nil(t, err)
	assert.Equal(t, name, s.Title)
	assert.Equal(t, entryCount, len(s.Entries))
	names := make([]string, 0)
	for _, e := range s.Entries {
		names = append(names, e.Title)
	}
	assert.Equal(t, entries, names)
}

func TestParseSeries(t *testing.T) {
	entries := []string{
		"v1",
		"v2",
	}
	testSeries(t, "../../tests/lib/20th Century Boys", "20th Century Boys", 2, entries)

	entries = []string{
		"Volume 01",
		"Volume 02",
	}
	testSeries(t, "../../tests/lib/Akira", "Akira", 2, entries)

	entries = []string{
		"Amano Megumi wa Suki Darake! v01",
	}
	testSeries(t, "../../tests/lib/Amano", "Amano", 1, entries)
}
