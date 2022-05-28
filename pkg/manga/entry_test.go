package manga

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func testArchive(t *testing.T, path, name string, pageCount int, pages Pages) {
	a, err := ParseEntry(context.TODO(), path)
	require.Nil(t, err)
	require.Equal(t, name, a.Title)
	require.Equal(t, pageCount, len(a.Pages))
	require.Equal(t, pages, a.Pages)
	require.Equal(t, a.Archive.Exists(), true)
}

func TestParseArchive(t *testing.T) {
	pages := Pages{
		"akira_1_c001.jpg",
		"akira_1_ic01.jpg",
		"akira_1_ic02-ic03.jpg",
		"akira_1_ic04.jpg",
		"akira_1_ic05.jpg",
		"akira_1_p001.jpg",
		"akira_1_p002-p003.jpg",
		"Akira_1_p004-p005.jpg",
		"Akira_1_p006-p007.jpg",
		"Akira_1_p008.jpg",
		"Akira_1_p009.jpg",
		"Akira_1_p010.jpg",
		"Akira_1_p011.jpg",
		"Akira_1_p356-p357.jpg",
		"Akira_1_rc01.jpg",
	}
	testArchive(t, "../../tests/lib/Akira/Volume 01.zip", "Volume 01", 15, pages)

	pages = Pages{
		"0000.jpg",
		"20th Century Boys v01 (001).png",
		"20th Century Boys v01 (002).png",
		"20th Century Boys v01 (003).png",
		"20th Century Boys v01 (004).png",
		"20th Century Boys v01 (005).png",
		"20th Century Boys v01 (006).png",
	}
	testArchive(t, "../../tests/lib/20th Century Boys/v1.zip", "v1", 7, pages)

	pages = Pages{
		"Vol.01 Ch.0001 - A/001.jpg",
		"Vol.01 Ch.0001 - A/002.png",
		"Vol.01 Ch.0001 - A/003.jpg",
		"Vol.01 Ch.0001 - A/004.jpg",
		"Vol.01 Ch.0001 - A/005.jpg",
		"Vol.01 Ch.0001 - A/006.png",
		"Vol.01 Ch.0001 - A/007.png",
		"Vol.01 Ch.0001 - A/008.png",
		"Vol.01 Ch.0001 - A/009.png",
		"Vol.01 Ch.0001 - A/010.png",
		"Vol.01 Ch.0001 - A/011.png",
		"Vol.01 Ch.0002 - B/001.png",
		"Vol.01 Ch.0002 - B/002.png",
		"Vol.01 Ch.0002 - B/003.png",
		"Vol.01 Ch.0002 - B/004.png",
		"Vol.01 Ch.0002 - B/005.png",
		"Vol.01 Ch.0002 - B/006.png",
		"Vol.01 Ch.0002 - B/007.png",
		"Vol.01 Ch.0002 - B/008.png",
		"Vol.01 Ch.0002 - B/009.png",
		"Vol.01 Ch.0002 - B/010.png",
		"Vol.01 Ch.0002 - B/011.png",
	}
	testArchive(t, "../../tests/lib/Amano/Amano Megumi wa Suki Darake! v01.zip", "Amano Megumi wa Suki Darake! v01", 22, pages)
}
