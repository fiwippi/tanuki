package manga

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/platform/image"
)

func testArchive(t *testing.T, path, name string, pageCount int, pages Pages) {
	a, err := ParseEntry(context.TODO(), path)
	require.Nil(t, err)
	require.Equal(t, name, a.FileTitle)
	require.Equal(t, pageCount, len(a.Pages))
	require.Equal(t, pages, a.Pages)
	require.Equal(t, a.Archive.Exists(), true)
}

func TestParseArchive(t *testing.T) {
	pages := Pages{
		{Path: "akira_1_c001.jpg", Type: image.JPEG},
		{Path: "akira_1_ic01.jpg", Type: image.JPEG},
		{Path: "akira_1_ic02-ic03.jpg", Type: image.JPEG},
		{Path: "akira_1_ic04.jpg", Type: image.JPEG},
		{Path: "akira_1_ic05.jpg", Type: image.JPEG},
		{Path: "akira_1_p001.jpg", Type: image.JPEG},
		{Path: "akira_1_p002-p003.jpg", Type: image.JPEG},
		{Path: "Akira_1_p004-p005.jpg", Type: image.JPEG},
		{Path: "Akira_1_p006-p007.jpg", Type: image.JPEG},
		{Path: "Akira_1_p008.jpg", Type: image.JPEG},
		{Path: "Akira_1_p009.jpg", Type: image.JPEG},
		{Path: "Akira_1_p010.jpg", Type: image.JPEG},
		{Path: "Akira_1_p011.jpg", Type: image.JPEG},
		{Path: "Akira_1_p356-p357.jpg", Type: image.JPEG},
		{Path: "Akira_1_rc01.jpg", Type: image.JPEG},
	}
	testArchive(t, "../../tests/lib/Akira/Volume 01.zip", "Volume 01", 15, pages)

	pages = Pages{
		{Path: "0000.jpg", Type: image.JPEG},
		{Path: "20th Century Boys v01 (001).png", Type: image.PNG},
		{Path: "20th Century Boys v01 (002).png", Type: image.PNG},
		{Path: "20th Century Boys v01 (003).png", Type: image.PNG},
		{Path: "20th Century Boys v01 (004).png", Type: image.PNG},
		{Path: "20th Century Boys v01 (005).png", Type: image.PNG},
		{Path: "20th Century Boys v01 (006).png", Type: image.PNG},
	}
	testArchive(t, "../../tests/lib/20th Century Boys/v1.zip", "v1", 7, pages)

	pages = Pages{
		{Path: "Vol.01 Ch.0001 - A/001.jpg", Type: image.JPEG},
		{Path: "Vol.01 Ch.0001 - A/002.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0001 - A/003.jpg", Type: image.JPEG},
		{Path: "Vol.01 Ch.0001 - A/004.jpg", Type: image.JPEG},
		{Path: "Vol.01 Ch.0001 - A/005.jpg", Type: image.JPEG},
		{Path: "Vol.01 Ch.0001 - A/006.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0001 - A/007.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0001 - A/008.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0001 - A/009.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0001 - A/010.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0001 - A/011.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/001.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/002.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/003.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/004.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/005.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/006.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/007.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/008.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/009.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/010.png", Type: image.PNG},
		{Path: "Vol.01 Ch.0002 - B/011.png", Type: image.PNG},
	}
	testArchive(t, "../../tests/lib/Amano/Amano Megumi wa Suki Darake! v01.zip", "Amano Megumi wa Suki Darake! v01", 22, pages)
}
