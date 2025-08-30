package tanuki

import (
	"archive/zip"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/maruel/natural"
	"golang.org/x/text/encoding/charmap"
)

// Pages

type Page struct {
	Path string
	Mime string
	// Was the path originally encoded using a non-UTF-8
	// encoding? The only alternate encoding we support
	// is CP437
	NonUtf8 bool
}

type Pages []Page

func (p Pages) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Pages) Scan(src any) error {
	src, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("incompatible type")
	}
	return json.Unmarshal(src.([]byte), &p)
}

// Entry

type Entry struct {
	EID      string
	SID      string
	Title    string
	ModTime  time.Time
	Archive  string
	Filesize int64
	Pages    Pages
}

var validImageTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/tiff": {},
	"image/bmp":  {},
}

func ParseEntry(path string) (Entry, error) {
	slog.Debug("Parsing entry", slog.String("path", path))

	abs, err := filepath.Abs(path)
	if err != nil {
		return Entry{}, err
	}
	stat, err := os.Stat(path)
	if err != nil {
		return Entry{}, err
	}
	title := strings.TrimSuffix(stat.Name(), filepath.Ext(stat.Name()))

	e := Entry{
		EID:      Sha256(title),
		Title:    title,
		Archive:  abs,
		Filesize: stat.Size(),
		ModTime:  stat.ModTime().Round(0), // Strip the monotonic clock reading
		Pages:    make([]Page, 0),
	}

	r, err := zip.OpenReader(abs)
	if err != nil {
		return Entry{}, err
	}
	defer r.Close()

	for _, f := range r.File {
		fi := f.FileInfo()
		if !fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
			m := mime.TypeByExtension(filepath.Ext(fi.Name()))
			if _, found := validImageTypes[m]; !found {
				return Entry{}, fmt.Errorf("invalid image mime for page %s: %s", fi.Name(), m)
			}

			name := f.Name
			if f.NonUTF8 {
				name, err = decodeCP437(f.Name)
				if err != nil {
					return Entry{}, fmt.Errorf("invalid CP437 name for page %s: %w", fi.Name(), err)
				}
			}
			e.Pages = append(e.Pages, Page{
				Path:    name,
				Mime:    m,
				NonUtf8: f.NonUTF8,
			})
		}

	}
	if len(e.Pages) == 0 {
		return Entry{}, fmt.Errorf("archive contains no pages")
	}

	// Go reads the ZIP files in string-sorted order, which means
	// they're read as out-of-order in some cases because they're
	// "natural" sorted. Some archives also have problems with bad
	// casing, so we just lowercase everything to be safe
	sort.SliceStable(e.Pages, func(i, j int) bool {
		a := strings.TrimSuffix(e.Pages[i].Path, filepath.Ext(e.Pages[i].Path))
		b := strings.TrimSuffix(e.Pages[j].Path, filepath.Ext(e.Pages[j].Path))
		return natural.Less(strings.ToLower(a), strings.ToLower(b))
	})

	return e, nil
}

// Series

type Series struct {
	SID     string
	Title   string
	Author  string
	ModTime time.Time
}

var validArchiveExtensions = map[string]struct{}{
	".zip": {},
	".cbz": {},
}

func ParseSeries(path string) (Series, []Entry, error) {
	slog.Debug("Parsing series", slog.String("path", path))

	stat, err := os.Stat(path)
	if err != nil {
		return Series{}, nil, err
	}

	s := Series{
		SID:     Sha256(stat.Name()),
		Title:   stat.Name(),
		ModTime: time.Time{},
	}
	entries := make([]Entry, 0)

	// Authors do not necessarily have to exist
	authorFile, err := os.Open(path + "/author.txt")
	if err == nil {
		author, err := io.ReadAll(authorFile)
		if err != nil {
			return Series{}, nil, fmt.Errorf("read author.txt")
		}
		s.Author = strings.TrimRight(string(author), "\n")
	} else if !errors.Is(err, fs.ErrNotExist) {
		slog.Error("Could not open author file", slog.Any("err", err))
	}

	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}
		_, valid := validArchiveExtensions[filepath.Ext(p)]
		if !valid {
			return nil
		}

		e, err := ParseEntry(p)
		if err != nil {
			return fmt.Errorf("parse entry %s: %w", p, err)
		}
		e.SID = s.SID
		if e.ModTime.After(s.ModTime) {
			s.ModTime = e.ModTime
		}
		entries = append(entries, e)

		return nil
	})
	if err != nil {
		return Series{}, nil, err
	}

	return s, entries, nil
}

// Library

type ParseErrorItem struct {
	Name string
	Err  error
}

type ParseError struct {
	Items []ParseErrorItem
}

func (pe *ParseError) Error() string {
	msgs := make([]string, len(pe.Items))
	for i, item := range pe.Items {
		msgs[i] = fmt.Sprintf("(%s, %s)", item.Name, item.Err)
	}
	return "parse errors: " + strings.Join(msgs, "; ")
}

func ParseLibrary(path string) (map[Series][]Entry, error) {
	lib := make(map[Series][]Entry)

	items, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var pErr ParseError
	for _, item := range items {
		if !item.IsDir() {
			continue
		}

		series, entries, err := ParseSeries(filepath.Join(path, item.Name()))
		if err != nil {
			pErr.Items = append(pErr.Items, ParseErrorItem{item.Name(), err})
			continue
		}

		lib[series] = entries
	}
	if len(pErr.Items) > 0 {
		// We return what we've succesfully managed to parse
		// instead of only returning an empty library
		return lib, &pErr
	}

	return lib, nil
}

// Hashing / Encoding

func Sha256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	digest := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(digest)
}

func decodeCP437(s string) (string, error) {
	decBytes, err := charmap.CodePage437.NewDecoder().Bytes([]byte(s))
	if err != nil {
		return "", err
	}
	return string(decBytes), nil
}
