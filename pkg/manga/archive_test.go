package manga

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/fse"
)

func TestArchive_Exists(t *testing.T) {
	require.False(t, fse.Exists(new(Archive).Path))
	a := Archive{Path: "../../tests/lib/Akira/Volume 01.zip"}
	require.True(t, fse.Exists(a.Path))
}

func TestArchive_ReaderForFile(t *testing.T) {
	validArchiveFile := func(r io.Reader, size int64, err error) {
		require.Nil(t, err)
		require.True(t, size > 0)
		data, err := io.ReadAll(r)
		require.Nil(t, err)
		require.True(t, len(data) > 0)
	}

	// Get reader for archive with no folder in filepath
	a := Archive{Path: "../../tests/lib/Akira/Volume 01.zip"}
	validArchiveFile(a.Extract(context.Background(), "Akira_1_rc01.jpg"))

	// Get reader for archive with folder in filepath
	a = Archive{Path: "../../tests/lib/Amano/Amano Megumi wa Suki Darake! v01.zip"}
	validArchiveFile(a.Extract(context.Background(), "Vol.01 Ch.0001 - A/001.jpg"))
}
