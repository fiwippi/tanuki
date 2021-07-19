package core

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mholt/archiver/v3"

	"github.com/fiwippi/tanuki/internal/fse"
)

// Validating and parsing Manga metadata (for formed archives)

func validateMetadata(metadata string) ([]string, error) {
	// Expects data in the form:
	// - "Chainsaw Man - c090 (NA) - p020 [web] [VIZ Media].jpg"
	// - "Chainsaw Man - c001 (v01) - p048-049 [dig] [VIZ Media].jpg"
	// - "Chainsaw Man - c001 (v01) - p000 [CoverImage] [dig] [VIZ Media].jpg"

	parts := strings.Split(metadata, " - ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("metadata '%s' is malformed, it does not have 3 parts separated by ' - '", metadata)
	}
	return parts, nil
}

func parseMetadataSeriesTitle(metadata string) (string, error) {
	parts, err := validateMetadata(metadata)
	if err != nil {
		return "", err
	}

	return parts[0], err
}

func parseMetadataChapterAndVolume(metadata string) (int, int, error) {
	parts, err := validateMetadata(metadata)
	if err != nil {
		return -1, -1, err
	}

	// Second part contains the chapter and volume
	order := strings.Split(parts[1], " ")
	if len(order) < 2 {
		return -1, -1, fmt.Errorf("chapter volume information '%s' could not be split into two parts", parts[1])
	}

	var ch, vol = -1, -1
	ch, err = strconv.Atoi(strings.TrimPrefix(order[0], "c"))
	if err != nil {
		return -1, -1, fmt.Errorf("could not convert chapter number '%s' to int", strings.TrimPrefix(order[0], "c"))
	}
	if strings.HasPrefix(order[1], "(v") {
		volumeStr := strings.TrimPrefix(order[1], "(v")
		volumeStr = strings.TrimSuffix(volumeStr, ")")
		v, err := strconv.Atoi(volumeStr)
		if err != nil {
			return -1, -1, fmt.Errorf("could not convert volume number '%s' to int", volumeStr)
		}
		vol = v
	}

	return ch, vol, nil
}

func parseMetadataPage(metadata string) (*Page, error) {
	parts, err := validateMetadata(metadata)
	if err != nil {
		return nil, err
	}

	p := &Page{IsCover: false}

	// Third part contains which pages the image file covers and whether it's a cover
	//pagesCombined := strings.Split(strings.TrimPrefix(parts[2], "p"), " ")[0]
	//for _, v := range strings.Split(pagesCombined, "-") {
	//	page, err := strconv.Atoi(v)
	//	if err != nil {
	//		return nil, errors.New("failed to parse page string as int")
	//	}
	//	p.Pages = append(p.Pages, page)
	//
	//}

	if strings.Contains(parts[2], "[CoverImage]") {
		p.IsCover = true
	}

	return p, nil
}

// Parsing Manga

func ParseMangaArchive(fp string) (*Manga, error) {
	// Ensure valid archive
	at, err := GetArchiveType(fp)
	if err != nil {
		return nil, err
	}
	absFp, err := filepath.Abs(fp)
	if err != nil {
		return nil, err
	}
	aStats, err := os.Stat(absFp)
	if err != nil {
		return nil, err
	}
	a := &Archive{
		Path:    absFp,
		Type:    at,
		Cover:   &Cover{},
		ModTime: aStats.ModTime(),
	}

	// Parse the archive into a Manga struct
	var m *Manga
	m, err = parseFormedArchive(a)
	if err != nil {
		m, err = parseMalformedArchive(a)
		if err != nil {
			return nil, err
		}
	}

	// We want to set general manga metadata after we parsed the
	// pages since if the manga is malformed the manga struct is
	// reset which clears our old metadata
	m.Title = fse.Filename(fp)
	m.Metadata.Title = m.Title

	// First item (which is a page) is guaranteed to be first in order
	// of pages since all cbz/cbr archive are ordered beforehand, so
	// we default to using the first page as cover if one could not be
	// set

	if m.Archive.Cover.Fp == "" {
		m.Archive.Cover.Fp = m.Pages[0].Path
		m.Archive.Cover.ImageType = m.Pages[0].ImageType
	}

	return m, nil
}

func parseFormedArchive(a *Archive) (*Manga, error) {
	m := newManga()
	m.Archive = a
	parsedOnce := false
	err := a.Walk(func(f archiver.File) error {
		// We only want to parse for metadata once but
		// it has to be successful so we don't use
		// sync.Once
		if !parsedOnce {
			//title, err1 := parseMetadataTitle(f.Name())
			ch, vol, err := parseMetadataChapterAndVolume(f.Name())

			if err == nil {
				m.Metadata.Chapter = ch
				m.Metadata.Volume = vol
				parsedOnce = true
			}
		}

		// Validate page has correct image type
		t, err := GetImageType(filepath.Ext(f.Name()))
		if err != nil {
			return err
		}

		// Parse page metadata
		page, err := parseMetadataPage(f.Name())
		if err != nil {
			return err
		}
		// Set extra metadata
		page.ImageType = t
		page.Path = f.Name()

		// Set the cover page
		if m.Archive.Cover.Fp == "" && page.IsCover {
			m.Archive.Cover.Fp = f.Name()
			m.Archive.Cover.ImageType = t
		}

		m.Pages = append(m.Pages, page)

		return nil

	})
	if err != nil {
		return nil, err
	}

	return m, err
}

func parseMalformedArchive(a *Archive) (*Manga, error) {
	m := newManga()
	m.Archive = a

	// Iterate through the files in the archive, try to parse metadata
	err := a.Walk(func(f archiver.File) error {
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			// Validate page has correct image type
			t, err := GetImageType(filepath.Ext(f.Name()))
			if err != nil {
				return err
			}

			// Allow one cover page
			if strings.HasPrefix(strings.ToLower(f.Name()), "cover") {
				m.Archive.Cover.Fp = f.Name()
				m.Archive.Cover.ImageType = t
			} else {
				// If image isn't a cover page then parse it as a proper page
				page := &Page{
					ImageType: t,
				}

				// If zip we need the file header to get the absolute filepath
				// but with rar calling f.Name() already gives it to us
				switch a.Type {
				case ZipArchive:
					page.Path = f.Header.(zip.FileHeader).Name
				case RarArchive:
					page.Path = f.Name()
				}

				//
				m.Pages = append(m.Pages, page)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, err
}

// Parsing Series

func ParseSeriesFolder(dir string) (*Series, []*Manga, error) {
	s := newSeries()
	e := make([]*Manga, 0)

	// Set the series title
	s.Title = fse.Filename(dir)

	// Set the entries
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// We want to avoid parsing non-archive files like cover images
			if _, err := GetArchiveType(path); err == nil {
				m, err := ParseMangaArchive(path)
				if err != nil {
					return err
				}
				e = append(e, m)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return s, e, err
}
