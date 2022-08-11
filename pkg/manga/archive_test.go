package manga

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchive_Exists(t *testing.T) {
	require.False(t, new(Archive).Exists())
	a := Archive{Path: "../../tests/lib/Akira/Volume 01.zip"}
	require.True(t, a.Exists())
}

func TestArchive_ReaderForFile(t *testing.T) {
	validArchiveFile := func(r io.Reader, size int64, err error) {
		require.Nil(t, err)
		require.True(t, size > 0)
		data, err := ioutil.ReadAll(r)
		require.Nil(t, err)
		require.True(t, len(data) > 0)
	}

	// Get reader for archive with no folder in filepath
	a := Archive{Path: "../../tests/lib/Akira/Volume 01.zip"}
	validArchiveFile(a.ReaderForFile("Akira_1_rc01.jpg"))

	// Get reader for archive with folder in filepath
	a = Archive{Path: "../../tests/lib/Amano/Amano Megumi wa Suki Darake! v01.zip"}
	validArchiveFile(a.ReaderForFile("Vol.01 Ch.0001 - A/001.jpg"))
}
