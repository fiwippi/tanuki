package manga

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
)

func testSeries(t *testing.T, path, name string, pageCount int, entries []string) {
	fp, err := filepath.Abs(path)
	require.Nil(t, err)
	s, en, err := ParseSeries(context.TODO(), fp)
	require.Nil(t, err)
	require.NotEqual(t, s.SID, xid.NilID())
	require.Equal(t, name, s.FolderTitle)
	require.Equal(t, len(entries), len(en))
	require.Equal(t, len(entries), s.NumEntries)
	require.Equal(t, pageCount, s.NumPages)
	names := make([]string, 0)
	for _, e := range en {
		names = append(names, e.FileTitle)
	}
	require.Equal(t, entries, names)
}

func TestParseSeries(t *testing.T) {
	defer os.Remove("../../tests/lib/20th Century Boys/info.tanuki")
	entries := []string{
		"v1",
		"v2",
	}
	testSeries(t, "../../tests/lib/20th Century Boys", "20th Century Boys", 14, entries)

	defer os.Remove("../../tests/lib/Akira/info.tanuki")
	entries = []string{
		"Volume 01",
		"Volume 02",
	}
	testSeries(t, "../../tests/lib/Akira", "Akira", 25, entries)

	defer os.Remove("../../tests/lib/Amano/info.tanuki")
	entries = []string{
		"Amano Megumi wa Suki Darake! v01",
	}
	testSeries(t, "../../tests/lib/Amano", "Amano", 22, entries)
}
