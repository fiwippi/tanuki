package tanuki

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParsing_ParseEntry(t *testing.T) {
	t.Run("20th Century Boys", func(t *testing.T) {
		path := "tests/lib/20th Century Boys/v1.zip"
		e, err := ParseEntry(path)
		require.NoError(t, err)

		// We have to manually add the SID since ParseEntry
		// doesn't add it, and the SID in present in our
		// parsed example
		e.SID = "PvHfuhL24GD6jo-PKLbPj_KvRikLn2WjCw_gOaXKRyI"
		require.Equal(t, centuryEntries[0], e)
	})

	t.Run("Akira", func(t *testing.T) {
		path := "tests/lib/Akira/Volume 01.zip"
		e, err := ParseEntry(path)
		require.NoError(t, err)

		// We have to manually add the SID since ParseEntry
		// doesn't add it, and the SID in present in our
		// parsed example
		e.SID = "rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c"
		require.Equal(t, akiraEntries[0], e)
	})

	t.Run("Amano", func(t *testing.T) {
		path := "tests/lib/Amano/Amano Megumi wa Suki Darake! v01.zip"
		e, err := ParseEntry(path)
		require.NoError(t, err)

		// We have to manually add the SID since ParseEntry
		// doesn't add it, and the SID in present in our
		// parsed example
		e.SID = "wNgocaIzfIjmFcxC-5I3S5pEpjRKjDY4nRxg9Ko-z7k"
		require.Equal(t, amanoEntries[0], e)
	})
}

func TestParsing_ParseSeries(t *testing.T) {
	t.Run("20th Century Boys", func(t *testing.T) {
		s, e, err := ParseSeries("tests/lib/20th Century Boys")
		require.NoError(t, err)
		require.Len(t, e, 2)
		require.Equal(t, centurySeries, s)
	})

	t.Run("Akira", func(t *testing.T) {
		s, e, err := ParseSeries("tests/lib/Akira")
		require.NoError(t, err)
		require.Len(t, e, 2)
		require.Equal(t, akiraSeries, s)
	})

	t.Run("Amano", func(t *testing.T) {
		s, e, err := ParseSeries("tests/lib/Amano")
		require.NoError(t, err)
		require.Len(t, e, 1)
		require.Equal(t, amanoSeries, s)
	})

	t.Run("Amano (.cbz)", func(t *testing.T) {
		s, e, err := ParseSeries("tests/lib-cbz/Amano")
		require.NoError(t, err)
		require.Len(t, e, 1)
		require.Equal(t, amanoSeries, s)
	})
}

func TestParsing_ParseLibrary(t *testing.T) {
	lib, err := ParseLibrary("tests/lib")
	require.NoError(t, err)
	require.Equal(t, parsedLib, lib)
}

// Parsed data

var folderPath = func() string {
	p, err := filepath.Abs("tests/lib")
	if err != nil {
		panic(err)
	}
	return p
}()

var parsedLib = map[Series][]Entry{
	centurySeries: centuryEntries,
	akiraSeries:   akiraEntries,
	amanoSeries:   amanoEntries,
}

var centurySeries = Series{
	SID:     "PvHfuhL24GD6jo-PKLbPj_KvRikLn2WjCw_gOaXKRyI",
	Title:   "20th Century Boys",
	Author:  "Naoki Urusawa",
	ModTime: parseTime("2022-08-11T16:53:23.8437325+01:00"),
}

var centuryEntries = []Entry{
	{
		EID:      "O_wmlZTvZJIo6adLqwDwQu_JHVrMb77jGjgugNQjiP4",
		SID:      "PvHfuhL24GD6jo-PKLbPj_KvRikLn2WjCw_gOaXKRyI",
		Title:    "v1",
		ModTime:  parseTime("2022-08-11T16:53:23.8317325+01:00"),
		Archive:  folderPath + "/20th Century Boys/v1.zip",
		Filesize: 27143,
		Pages: Pages{
			{Path: "0000.jpg", Mime: "image/jpeg"},
			{Path: "20th Century Boys v01 (001).png", Mime: "image/png"},
			{Path: "20th Century Boys v01 (002).png", Mime: "image/png"},
			{Path: "20th Century Boys v01 (003).png", Mime: "image/png"},
			{Path: "20th Century Boys v01 (004).png", Mime: "image/png"},
			{Path: "20th Century Boys v01 (005).png", Mime: "image/png"},
			{Path: "20th Century Boys v01 (006).png", Mime: "image/png"},
		},
	},
	{
		EID:      "-wTctpcOTD0Yc95R_VpQ17tGszgxE2AmZcNQ7EC1-ZA",
		SID:      "PvHfuhL24GD6jo-PKLbPj_KvRikLn2WjCw_gOaXKRyI",
		Title:    "v2",
		ModTime:  parseTime("2022-08-11T16:53:23.8437325+01:00"),
		Archive:  folderPath + "/20th Century Boys/v2.zip",
		Filesize: 27704,
		Pages: Pages{
			{Path: "0000.jpg", Mime: "image/jpeg"},
			{Path: "20th Century Boys v02 (001).png", Mime: "image/png"},
			{Path: "20th Century Boys v02 (002).png", Mime: "image/png"},
			{Path: "20th Century Boys v02 (003).png", Mime: "image/png"},
			{Path: "20th Century Boys v02 (004).png", Mime: "image/png"},
			{Path: "20th Century Boys v02 (005).png", Mime: "image/png"},
			{Path: "20th Century Boys v02 (006).png", Mime: "image/png"},
		},
	},
}

var akiraSeries = Series{
	SID:     "rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c",
	Title:   "Akira",
	Author:  "Katsuhiro Otomo",
	ModTime: parseTime("2022-08-11T16:53:23.8677336+01:00"),
}

var akiraEntries = []Entry{
	{
		EID:      "1f2Xo_TQk-nS-9I9QsRm3zVNawdW6HlOUYJsV22wENk",
		SID:      "rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c",
		Title:    "Volume 01",
		ModTime:  parseTime("2022-08-11T16:53:23.856733+01:00"),
		Archive:  folderPath + "/Akira/Volume 01.zip",
		Filesize: 26881,
		Pages: Pages{
			{Path: "akira_1_c001.jpg", Mime: "image/jpeg"},
			{Path: "akira_1_ic01.jpg", Mime: "image/jpeg"},
			{Path: "akira_1_ic02-ic03.jpg", Mime: "image/jpeg"},
			{Path: "akira_1_ic04.jpg", Mime: "image/jpeg"},
			{Path: "akira_1_ic05.jpg", Mime: "image/jpeg"},
			{Path: "akira_1_p001.jpg", Mime: "image/jpeg"},
			{Path: "akira_1_p002-p003.jpg", Mime: "image/jpeg"},
			{Path: "Akira_1_p004-p005.jpg", Mime: "image/jpeg"},
			{Path: "Akira_1_p006-p007.jpg", Mime: "image/jpeg"},
			{Path: "Akira_1_p008.jpg", Mime: "image/jpeg"},
			{Path: "Akira_1_p009.jpg", Mime: "image/jpeg"},
			{Path: "Akira_1_p010.jpg", Mime: "image/jpeg"},
			{Path: "Akira_1_p011.jpg", Mime: "image/jpeg"},
			{Path: "Akira_1_p356-p357.jpg", Mime: "image/jpeg"},
			{Path: "Akira_1_rc01.jpg", Mime: "image/jpeg"},
		},
	},
	{
		EID:      "ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o",
		SID:      "rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c",
		Title:    "Volume 02",
		ModTime:  parseTime("2022-08-11T16:53:23.8677336+01:00"),
		Archive:  folderPath + "/Akira/Volume 02.zip",
		Filesize: 18690,
		Pages: Pages{
			{Path: "Akira_2_c001.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_ic01.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_ic02-ic03.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_ic04-ic05.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_p001-p002.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_p003.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_p004.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_p005.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_p006.jpg", Mime: "image/jpeg"},
			{Path: "Akira_2_rc01.jpg", Mime: "image/jpeg"},
		},
	},
}

var amanoSeries = Series{
	SID:     "wNgocaIzfIjmFcxC-5I3S5pEpjRKjDY4nRxg9Ko-z7k",
	Title:   "Amano",
	Author:  "Nekoguchi",
	ModTime: parseTime("2022-08-11T16:53:23.888737+01:00"),
}

var amanoEntries = []Entry{
	{
		EID:      "r60ZPxCs2SaWHRLpogVWVibDPnkquh8REuYXO4mTYTg",
		SID:      "wNgocaIzfIjmFcxC-5I3S5pEpjRKjDY4nRxg9Ko-z7k",
		Title:    "Amano Megumi wa Suki Darake! v01",
		ModTime:  parseTime("2022-08-11T16:53:23.888737+01:00"),
		Archive:  folderPath + "/Amano/Amano Megumi wa Suki Darake! v01.zip",
		Filesize: 118344,
		Pages: Pages{
			{Path: "Vol.01 Ch.0001 - A/001.jpg", Mime: "image/jpeg"},
			{Path: "Vol.01 Ch.0001 - A/002.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0001 - A/003.jpg", Mime: "image/jpeg"},
			{Path: "Vol.01 Ch.0001 - A/004.jpg", Mime: "image/jpeg"},
			{Path: "Vol.01 Ch.0001 - A/005.jpg", Mime: "image/jpeg"},
			{Path: "Vol.01 Ch.0001 - A/006.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0001 - A/007.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0001 - A/008.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0001 - A/009.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0001 - A/010.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0001 - A/011.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/001.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/002.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/003.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/004.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/005.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/006.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/007.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/008.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/009.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/010.png", Mime: "image/png"},
			{Path: "Vol.01 Ch.0002 - B/011.png", Mime: "image/png"},
		},
	},
}

// Utils

func parseTime(str string) time.Time {
	p, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic(err)
	}
	return p
}
