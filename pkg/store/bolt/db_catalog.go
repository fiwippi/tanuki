package bolt

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/fvbommel/sortorder"
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/errors"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/internal/sets"
	"github.com/fiwippi/tanuki/pkg/store/bolt/buckets"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
)

var (
	ErrCoverEmpty     = errors.New("cover is empty")
	ErrThumbnailEmpty = errors.New("thumbnail is empty")
)

func (db *DB) catalogBucket(tx *bolt.Tx) *buckets.CatalogBucket {
	return &buckets.CatalogBucket{Bucket: tx.Bucket(keys.Catalog)}
}

// Catalog

func (db *DB) PopulateCatalog(series []*manga.ParsedSeries) error {
	db.cont.Pause()
	defer db.cont.Resume()

	var errs error
	var cModTime time.Time

	newCatalog := make(api.Catalog, 0)
	for _, s := range series {
		err := db.Update(func(tx *bolt.Tx) error {
			// Add the series
			root := db.catalogBucket(tx)
			err := root.AddSeries(s)
			if err != nil {
				return err
			}

			sid := hash.SHA1(s.Title)
			sb, err := root.Series(sid)
			if err != nil {
				return err
			}

			mt := sb.ModTime()
			if cModTime == (time.Time{}) {
				cModTime = mt
			} else if mt.Before(cModTime) {
				cModTime = mt
			}

			// How many pages the series has (used to calculate progress)
			pages := 0
			for _, e := range s.Entries {
				pages += len(e.Pages)
			}

			// Create tentative metadata
			d := &api.Series{
				Hash:         sid,
				Title:        sb.Title(),
				Entries:      len(sb.EntriesMetadata()),
				TotalPages:   pages,
				Tags:         sb.Tags().List(),
				Author:       manga.AuthorZeroValue,
				DateReleased: nil,
			}

			// We always choose to preserve old metadata if it's not zero value
			m := sb.Metadata()
			if m != nil {
				if m.Title != manga.TitleZeroValue {
					d.Title = m.Title
				}
				if m.Author != manga.AuthorZeroValue {
					d.Author = m.Author
				}
				if m.DateReleased != nil && m.DateReleased.Time != manga.TimeZeroValue {
					d.DateReleased = m.DateReleased
				}
			}

			err = sb.SetMetadata(&manga.SeriesMetadata{
				Title:        d.Title,
				Author:       d.Author,
				DateReleased: d.DateReleased,
			})
			if err != nil {
				return err
			}

			newCatalog = append(newCatalog, d)

			return nil
		})
		if err != nil {
			errs = errors.Wrap(errs, err)
		}
	}

	// Sort catalog in natural order
	sort.SliceStable(newCatalog, func(i, j int) bool {
		return sortorder.NaturalLess(newCatalog[i].Title, newCatalog[j].Title)
	})
	err := db.Update(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)

		// Set the catalog's mod time
		if err := root.SetModTime(cModTime); err != nil {
			return err
		}

		// Update the series bucket so they know which index to use
		// to access their catalog metadata
		for i := range newCatalog {
			order := i + 1
			newCatalog[i].Order = order
			sb, err := root.Series(newCatalog[i].Hash)
			if err != nil {
				return err
			}

			err = sb.SetOrder(order)
			if err != nil {
				return err
			}
		}

		return root.SetCatalog(newCatalog)
	})
	if err != nil {
		errs = errors.Wrap(errs, err)
	}

	return errs
}

func (db *DB) GetCatalog() api.Catalog {
	var c api.Catalog
	db.View(func(tx *bolt.Tx) error {
		c = db.catalogBucket(tx).Catalog()
		return nil
	})
	return c
}

// Series

func (db *DB) HasSeries(sid string) bool {
	err := db.View(func(tx *bolt.Tx) error {
		_, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}
		return nil
	})
	return err == nil
}

func (db *DB) GetSeries(sid string) (*api.Series, error) {
	var e *api.Series
	err := db.View(func(tx *bolt.Tx) error {
		sm, err := db.catalogBucket(tx).SeriesMetadata(sid)
		if err != nil {
			return err
		}
		e = sm

		return nil
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

// Entries

func (db *DB) HasEntry(sid, eid string) bool {
	err := db.View(func(tx *bolt.Tx) error {
		_, err := db.catalogBucket(tx).Entry(sid, eid)
		if err == nil {
			return err
		}
		return nil
	})
	return err == nil
}

func (db *DB) GetEntry(sid, eid string) (*api.Entry, error) {
	var e *api.Entry
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}

		em, err := sb.EntryMetadata(eid)
		if err != nil {
			return err
		}
		e = em

		return nil
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (db *DB) GetEntries(sid string) (api.Entries, error) {
	var entries api.Entries
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}
		entries = sb.EntriesMetadata()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// Folder Title

func (db *DB) GetSeriesFolderTitle(sid string) (string, error) {
	var s string
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}
		s = sb.Title()
		return nil
	})
	if err != nil {
		return "", err
	}
	return s, nil
}

// Archive

func (db *DB) GetEntryArchive(sid, eid string) (*manga.Archive, error) {
	var a *manga.Archive
	err := db.View(func(tx *bolt.Tx) error {
		mb, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}
		a = mb.Archive()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// Page

func (db *DB) GetEntryPage(sid, eid string, num int) (*manga.Page, error) {
	var p *manga.Page
	err := db.View(func(tx *bolt.Tx) error {
		mb, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}

		pb := mb.PagesBucket()
		temp, err := pb.GetPage(num)
		if err != nil {
			return err
		}
		p = temp

		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ModTime

func (db *DB) GetEntryModTime(sid, eid string) (time.Time, error) {
	var t time.Time
	err := db.View(func(tx *bolt.Tx) error {
		b, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}
		t = b.Archive().ModTime
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (db *DB) GetSeriesModTime(sid string) (time.Time, error) {
	var t time.Time
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}
		t = sb.ModTime()

		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (db *DB) GetCatalogModTime() (time.Time, error) {
	var t time.Time
	err := db.View(func(tx *bolt.Tx) error {
		t = db.catalogBucket(tx).ModTime()
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// Cover

func (db *DB) GetSeriesCover(sid string) (*manga.Cover, error) {
	var s *manga.Cover
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}
		s = sb.Cover()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (db *DB) GetEntryCover(sid, eid string) (*manga.Cover, error) {
	var s *manga.Cover
	err := db.View(func(tx *bolt.Tx) error {
		eb, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}
		s = eb.Cover()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (db *DB) SetSeriesCover(sid string, cover *manga.Cover) error {
	err := db.Update(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}
		return sb.SetCover(cover)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetEntryCover(sid, eid string, cover *manga.Cover) error {
	err := db.Update(func(tx *bolt.Tx) error {
		eb, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}
		return eb.SetCover(cover)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetSeriesCoverFile(sid string) ([]byte, string, error) {
	return db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		root := db.catalogBucket(tx)

		// Get the custom series cover if exists
		// 1. ensure series exists
		// 2. get the series' has *core.Cover
		// 3. if a cover exists then ensure the file exists on the filesystem
		// 4. if a file exists then attempt to load it
		//
		// error only returned if the file exists and it could be loaded,
		// in cases where the cover exists but no file exists on the filesystem,
		// then this should be flagged as missing entry, on consecutive
		// delete-missing-items, the cover's fp should be deleted if it's file does
		// not exist

		// 1
		sb, err := root.Series(sid)
		if err != nil {
			return nil, "", err
		}
		// 2
		c := sb.Cover()
		// 3
		if c.Fp != "" && c.ExistsOnFS() {
			// 4
			data, err := c.ReadFile()
			if err != nil {
				return nil, "", err
			} else if len(data) == 0 {
				return nil, "", ErrCoverEmpty.Fmt(sid)
			}

			return data, c.ImageType.MimeType(), nil
		}

		// Otherwise get the cover from the first series entry
		firstEntry, err := root.FirstEntry(sid)
		if err != nil {
			return nil, "", err
		}
		c = firstEntry.Cover()
		tempData, err := firstEntry.Archive().CoverFile()
		if err != nil {
			return nil, "", err
		}
		return tempData, c.ImageType.MimeType(), nil
	})
}

func (db *DB) GetEntryCoverFile(sid, eid string) ([]byte, string, error) {
	return db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		root := db.catalogBucket(tx)

		// 1
		mb, err := root.Entry(sid, eid)
		if err != nil {
			return nil, "", err
		}
		// 2
		c := mb.Cover()
		// 3
		if c.Fp != "" && c.ExistsOnFS() {
			// 4
			data, err := c.ReadFile()
			if err != nil {
				return nil, "", err
			} else if len(data) == 0 {
				return nil, "", ErrCoverEmpty.Fmt(sid, eid)
			}

			return data, c.ImageType.MimeType(), nil
		}

		// Otherwise get the embedded cover
		c = mb.Archive().Cover
		tempData, err := mb.Archive().CoverFile()
		if err != nil {
			return nil, "", err
		}
		return tempData, c.ImageType.MimeType(), nil
	})
}

func (db *DB) DeleteSeriesCover(seriesHash string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(seriesHash)
		if err != nil {
			return err
		}

		// Get the cover
		c := sb.Cover()

		// DeleteEntry the cover image from the filesystem
		// and its directory if it's left empty
		os.Remove(c.Fp)
		fse.DeleteFileDirIfEmpty(c.Fp)

		// Clean references to the file
		c.Fp = ""
		err = sb.SetCover(c)
		if err != nil {
			return err
		}

		// DeleteEntry the thumbnail as well
		return sb.SetThumbnail(nil)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) DeleteEntryCover(sid, eid string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}

		// Get the cover
		c := b.Cover()

		// DeleteEntry the cover image from the filesystem
		// and its directory if it's left empty
		os.Remove(c.Fp)
		fse.DeleteFileDirIfEmpty(c.Fp)

		// Clean references to the file
		c.Fp = ""
		err = b.SetCover(c)
		if err != nil {
			return err
		}

		// DeleteEntry the thumbnail as well
		return b.SetThumbnail(nil)
	})
	if err != nil {
		return err
	}
	return nil
}

// Thumbnail

func (db *DB) GenerateThumbnails(forceNew bool) error {
	var errs error

	items := db.GetCatalog()
	for _, i := range items {
		db.cont.WaitIfPaused()
		db.Update(func(tx *bolt.Tx) error {
			root := db.catalogBucket(tx)

			sb, err := root.Series(i.Hash)
			if err != nil {
				return err
			}

			seriesThumbnailExists := sb.HasThumbnail()

			// Generate series thumbnail
			if !seriesThumbnailExists || (seriesThumbnailExists && forceNew) {
				c := sb.Cover()
				if c.Fp != "" && c.ExistsOnFS() {
					img, err := c.ThumbnailFile()
					if err != nil {
						errs = errors.Wrap(errs, err)
						return nil
					}

					err = sb.SetThumbnail(img)
					if err != nil {
						errs = errors.Wrap(errs, err)
						return nil
					}
				}
			}

			// Generate entries thumbnails
			return sb.ForEachEntry(func(_ string, eb *buckets.EntryBucket) error {
				thumbnailExists := eb.HasThumbnail()
				if !thumbnailExists || (thumbnailExists && forceNew) {
					// Create thumbnail of custom cover if it exists
					c := eb.Cover()
					if c.Fp != "" && c.ExistsOnFS() {
						img, err := c.ThumbnailFile()
						if err != nil {
							errs = errors.Wrap(errs, err)
							return nil
						}

						err = sb.SetThumbnail(img)
						if err != nil {
							errs = errors.Wrap(errs, err)
						}

						return nil
					}

					// Otherwise use thumbnail of default cover
					img, err := eb.Archive().ThumbnailFile()
					if err != nil {
						errs = errors.Wrap(errs, err)
						return nil
					}
					return eb.SetThumbnail(img)
				}
				return nil
			})
		})
	}

	return errs
}

func (db *DB) GenerateSeriesThumbnail(sid string, forceNew bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}

		seriesThumbnailExists := sb.HasThumbnail()
		if !seriesThumbnailExists || (seriesThumbnailExists && forceNew) {
			c := sb.Cover()
			if c.Fp != "" && c.ExistsOnFS() {
				img, err := c.ThumbnailFile()
				if err != nil {
					return err
				}

				err = sb.SetThumbnail(img)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (db *DB) GenerateEntryThumbnail(sid, eid string, forceNew bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}

		hasThumb := b.HasThumbnail()
		if !hasThumb || (hasThumb && forceNew) {
			// If custom cover exists try and create it
			c := b.Cover()
			if c.Fp != "" && c.ExistsOnFS() {
				data, err := c.ThumbnailFile()
				if err != nil {
					return err
				} else if len(data) == 0 {
					return ErrCoverEmpty.Fmt(sid, eid)
				}
				err = b.SetThumbnail(data)
				if err != nil {
					b.SetThumbnail(nil)
					return err
				}
				return nil
			}

			// If it doesn't exist then use the archive thumbnail
			data, err := b.Archive().ThumbnailFile()
			if err != nil {
				return err
			}
			return b.SetThumbnail(data)
		}
		return nil
	})
}

func (db *DB) GetSeriesThumbnail(sid string) ([]byte, string, error) {
	// Get thumbnail
	img, mimetype, err := db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		root := db.catalogBucket(tx)

		// Get the custom series cover if exists
		sb, err := root.Series(sid)
		if err != nil {
			return nil, "", err
		}
		if sb.HasThumbnail() {
			return sb.Thumbnail(), "image/jpeg", nil
		}
		return nil, "", ErrThumbnailEmpty.Fmt(sid)
	})
	if len(img) > 0 {
		return img, mimetype, err
	}

	// If can't get custom thumbnail then get thumbnail from the first entry
	var eid string
	err = db.View(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)
		eb, err := root.FirstEntry(sid)
		if err != nil {
			return err
		}
		eid = hash.SHA1(eb.Archive().Title)
		return nil
	})
	if err != nil {
		return nil, "", buckets.ErrEntryNotExist
	}

	return db.GetEntryThumbnail(sid, eid)
}

func (db *DB) GetEntryThumbnail(sid, eid string) ([]byte, string, error) {
	// Get the thumbnail
	img, mimetype, err := db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		b, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return nil, "", err
		}

		return b.Thumbnail(), "image/jpeg", nil
	})
	// Return it if it exists
	if len(img) > 0 {
		return img, mimetype, err
	}

	// If thumbnail doesn't exist try and recreate it
	err = db.GenerateEntryThumbnail(sid, eid, true)
	if err != nil {
		// We have no thumbnail and can't generate another one
		return nil, "", err
	}
	// Return the new thumbnail
	return db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		b, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return nil, "", err
		}

		return b.Thumbnail(), "image/jpeg", nil
	})
}

// Tags

func (db *DB) SetSeriesTags(sid string, tags []string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)
		sb, err := root.Series(sid)
		if err != nil {
			return err
		}

		// Set the new tags in the series bucket
		t := sb.Tags()
		t.Clear()
		t.Add(tags...)
		err = sb.SetTags(t)
		if err != nil {
			return err
		}

		// Set the new tags for the series entry
		sm, err := root.SeriesMetadata(sid)
		if err != nil {
			return err
		}
		sm.Tags = t.List()
		return root.SetSeriesMetadata(sid, sm)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetSeriesTags(sid string) (*sets.Set, error) {
	var t *sets.Set
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.catalogBucket(tx).Series(sid)
		if err != nil {
			return err
		}
		t = sb.Tags()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (db *DB) GetTags() *sets.Set {
	tags := sets.NewSet()
	db.View(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)
		return root.ForEachSeries(func(hash string, b *buckets.SeriesBucket) error {
			tags.Add(b.Tags().List()...)
			return nil
		})
	})
	return tags
}

func (db *DB) GetSeriesWithTag(tag string) api.Catalog {
	list := make(api.Catalog, 0)

	// We can ignore the error since we aren't returning
	// any errors in the ForEach traversal code
	db.View(func(tx *bolt.Tx) error {
		c := db.catalogBucket(tx).Catalog()

		for _, s := range c {
			for _, t := range s.Tags {
				if t == tag {
					list = append(list, s)
				}
			}
		}

		return nil
	})
	return list
}

// Missing items

func (db *DB) GetMissingItems() api.MissingItems {
	items := make(api.MissingItems, 0)

	// Checks for invalid archive and cover
	db.View(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)
		cat := root.Catalog()

		return root.ForEachSeries(func(sid string, sb *buckets.SeriesBucket) error {
			// Check if the series exists
			i := sb.Order() - 1
			if i < 0 || i >= len(cat) || cat[i].Hash != sid {
				e := &api.MissingItem{
					Type:  "Series",
					Title: sb.Title(),
					Path:  "",
				}
				items = append(items, e)
				return nil
			}

			// Check if series cover exists
			c := sb.Cover()
			if c.Fp != "" && !c.ExistsOnFS() {
				e := &api.MissingItem{
					Type:  "Cover",
					Title: fse.FilenameWithExt(c.Fp),
					Path:  c.Fp,
				}
				items = append(items, e)
			}

			em := sb.EntriesMetadata()

			return sb.ForEachEntry(func(eid string, eb *buckets.EntryBucket) error {
				// Check if the entry exists
				i := eb.Order() - 1
				if i < 0 || i >= len(em) || em[i].Hash != eid {
					a := eb.Archive()
					e := &api.MissingItem{
						Type:  "Entry",
						Title: a.Title,
						Path:  a.Path,
					}
					items = append(items, e)
					return nil
				}

				// Check if archive for the entry exists
				if !eb.Archive().Exists() {
					a := eb.Archive()
					e := &api.MissingItem{
						Type:  "Archive",
						Title: a.Title,
						Path:  a.Path,
					}
					items = append(items, e)
				}

				// Check if custom archive cover exists
				c := eb.Cover()
				if c.Fp != "" && !c.ExistsOnFS() {
					e := &api.MissingItem{
						Type:  "Cover",
						Title: fse.FilenameWithExt(c.Fp),
						Path:  c.Fp,
					}
					items = append(items, e)
				}

				return nil
			})
		})
	})

	// Checks for invalid progress
	db.View(func(tx *bolt.Tx) error {
		userRoot := db.usersBucket(tx)
		seriesRoot := db.catalogBucket(tx)

		return userRoot.ForEachUser(func(u *buckets.UserBucket) error {
			// Get catalog progress
			cp := u.Progress()
			if cp == nil {
				return nil
			}

			// For each series for each entry
			for sid, sp := range cp.Data {
				// Get entry metadata so we can get the index for the progress entry
				sb, err := seriesRoot.Series(sid)
				if err != nil {
					e := &api.MissingItem{
						Type:  "Progress",
						Title: u.Name(),
						Path:  fmt.Sprintf("Series: %s", sid),
					}
					items = append(items, e)
					continue
				}
				entries := sb.EntriesMetadata()

				for i := range sp.Entries {
					if i >= len(entries) {
						e := &api.MissingItem{
							Type:  "Progress",
							Title: u.Name(),
							Path:  fmt.Sprintf("Series: %s, Entry: %s", sid, "N/A"),
						}
						items = append(items, e)
						continue
					}

					eid := entries[i].Hash
					entry, err := seriesRoot.Entry(sid, eid)
					if err == nil && !entry.Archive().Exists() {
						e := &api.MissingItem{
							Type:  "Progress",
							Title: u.Name(),
							Path:  fmt.Sprintf("Series: %s, Entry: %s", sid, eid),
						}
						items = append(items, e)
					}
				}
			}

			return nil
		})
	})

	return items
}

func (db *DB) DeleteMissingItems() error {
	// Checks for invalid archive and cover
	err := db.Update(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)

		cat := root.Catalog()
		return root.ForEachSeries(func(sid string, sb *buckets.SeriesBucket) error {
			// Check if the series exists
			i := sb.Order() - 1
			if i < 0 || i >= len(cat) || cat[i].Hash != sid {
				return root.DeleteSeries(sid)
			}

			// Check if series cover exists
			c := sb.Cover()
			if c.Fp != "" && !c.ExistsOnFS() {
				// If it doesn't exist then reset the cover
				c.Fp = ""
				if err := sb.SetCover(c); err != nil {
					return err
				}
			}

			em := sb.EntriesMetadata()
			err := sb.ForEachEntry(func(eid string, eb *buckets.EntryBucket) error {
				// Check if the entry exists
				i := eb.Order() - 1
				if i < 0 || i >= len(em) || em[i].Hash != eid {
					return sb.DeleteEntry(eid)
				}

				// Check if archive for the entry exists
				if !eb.Archive().Exists() {
					if err := sb.DeleteEntry(eid); err != nil {
						return err
					}
					// If the entry is deleted we can't retrieve the cover
					return nil
				}

				// Check if custom cover exists
				c := eb.Cover()
				if c.Fp != "" && !c.ExistsOnFS() {
					c.Fp = ""
					if err := eb.SetCover(c); err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				return err
			}

			if len(sb.EntriesMetadata()) == 0 {
				err := root.DeleteSeries(sid)
				if err != nil {
					return err
				}
			}

			return root.RegenerateCatalog()
		})
	})
	if err != nil {
		return err
	}

	// Checks for invalid progress
	err = db.Update(func(tx *bolt.Tx) error {
		userRoot := db.usersBucket(tx)
		seriesRoot := db.catalogBucket(tx)

		return userRoot.ForEachUser(func(u *buckets.UserBucket) error {
			// Get catalog progress
			cp := u.Progress()
			if cp == nil {
				return nil
			}

			// For each series for each entry

			for sid, sp := range cp.Data {
				// Get entry metadata so we can get the index for the progress entry
				sb, err := seriesRoot.Series(sid)
				if err != nil {
					delete(cp.Data, sid)
					continue
				}
				entries := sb.EntriesMetadata()

				for i := range sp.Entries {
					if i >= len(entries) {
						sp.DeleteEntry(i)
						continue
					}

					eid := entries[i].Hash
					entry, err := seriesRoot.Entry(sid, eid)
					if err == nil && !entry.Archive().Exists() {
						sp.DeleteEntry(i)
					}
				}

				if len(sp.Entries) == 0 {
					cp.DeleteSeries(sid)
				}
			}

			return u.ChangeProgress(cp)
		})
	})
	if err != nil {
		return err
	}

	return nil
}

// Metadata

func (db *DB) SetSeriesMetadata(sid string, m *manga.SeriesMetadata) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)

		// Set the metadata in the catalog entry
		sm, err := root.SeriesMetadata(sid)
		if err != nil {
			return err
		}
		sm.Title = m.Title
		sm.Author = m.Author
		sm.DateReleased = m.DateReleased
		err = root.SetSeriesMetadata(sid, sm)
		if err != nil {
			return err
		}

		// Set the metadata in the series bucket
		sb, err := root.Series(sid)
		if err != nil {
			return err
		}
		return sb.SetMetadata(m)
	})
}

func (db *DB) SetEntryMetadata(sid, eid string, m *manga.EntryMetadata) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)

		sb, err := root.Series(sid)
		if err != nil {
			return err
		}

		oldM, err := sb.EntryMetadata(eid)
		if err != nil {
			return err
		}

		// Set the metadata in the series entries-metadata key
		oldM.Title = m.Title
		oldM.Author = m.Author
		oldM.DateReleased = m.DateReleased
		oldM.Chapter = m.Chapter
		oldM.Volume = m.Volume
		err = sb.SetEntryMetadata(eid, oldM)
		if err != nil {
			return err
		}

		// Set the metadata in the actual entries bucket
		eb, err := root.Entry(sid, eid)
		if err != nil {
			return err
		}
		return eb.SetMetadata(m)
	})
}

// Order

func (db *DB) GetEntryOrder(sid, eid string) (int, error) {
	var o int
	err := db.View(func(tx *bolt.Tx) error {
		eb, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}
		o = eb.Order()
		return nil
	})
	if err != nil {
		return -1, err
	}
	return o, nil
}

// Helper functions

func (db *DB) returnBytes(f func(tx *bolt.Tx) ([]byte, string, error)) ([]byte, string, error) {
	var data []byte
	var mimetype string

	err := db.View(func(tx *bolt.Tx) error {
		d, m, err := f(tx)
		if err != nil {
			return err
		}
		data = make([]byte, len(d))
		copy(data, d)
		mimetype = m

		return nil
	})
	if err != nil {
		return nil, "", err
	}
	return data, mimetype, nil
}
