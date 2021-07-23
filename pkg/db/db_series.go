package db

import (
	"fmt"
	"github.com/fiwippi/tanuki/pkg/auth"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/sets"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/core"
)

// GetCatalog

// TODO after commit, implement this without goroutine and check speed difference
func (db *DB) PopulateCatalog(series []*core.ParsedSeries) core.ErrorSlice {
	wg := sync.WaitGroup{}
	errors := core.NewErrorSlice()
	mu := sync.Mutex{}
	var cModTime time.Time

	oldCatalog := db.GetCatalog()
	newCatalog := make(api.Catalog, 0, len(oldCatalog))
	queue := make(chan *api.Series, 1)
	errorQueue := make(chan error, 1)

	go func() {
		for e := range errorQueue {
			errors = append(errors, e)
		}
	}()

	go func() {
		for s := range queue {
			newCatalog = append(newCatalog, s)
			wg.Done()
		}
	}()

	for _, s := range series {
		wg.Add(1)
		go func(series *core.ParsedSeries, oldC api.Catalog) {
			err := db.Batch(func(tx *bolt.Tx) error {
				// Add the series
				root := db.catalogBucket(tx)
				err := root.AddSeries(series)
				if err != nil {
					return err
				}

				sid := auth.SHA1(series.Title)
				sb, err := root.Series(sid)
				if err != nil {
					return err
				}

				mu.Lock()
				mt := sb.ModTime()
				if cModTime == (time.Time{}) {
					cModTime = mt
				} else if mt.Before(cModTime) {
					cModTime = mt
				}
				mu.Unlock()

				// Create tentative metadata
				d := &api.Series{
					Hash:         sid,
					Title:        sb.Title(),
					Entries:      len(sb.EntriesMetadata()),
					Tags:         sb.Tags().List(),
					Author:       core.AuthorZeroValue,
					DateReleased: nil,
				}

				// We always choose to preserve old metadata if it's not zero value
				o := sb.Order()
				if len(oldCatalog) > 0 && sb.Order() != -1 {
					oldData := oldC[o-1]
					if oldData.Title != core.TitleZeroValue {
						d.Title = oldData.Title
					}
					if oldData.Author != core.AuthorZeroValue {
						d.Author = oldData.Author
					}
					if oldData.DateReleased != nil && oldData.DateReleased.Time != core.TimeZeroValue {
						d.DateReleased = oldData.DateReleased
					}
				}

				queue <- d
				return nil
			})
			if err != nil {
				errorQueue <- err
			}
		}(s, oldCatalog)
	}

	wg.Wait()
	close(queue)
	close(errorQueue)

	// Sort catalog in lowercase lexical order
	sort.SliceStable(newCatalog, func(i, j int) bool {
		return strings.ToLower(newCatalog[i].Title) < strings.ToLower(newCatalog[j].Title)
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
		errors = append(errors, err)
	}

	return errors
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

func (db *DB) GetEntryArchive(sid, eid string) (*core.Archive, error) {
	var a *core.Archive
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

func (db *DB) GetEntryPage(sid, eid string, num int) (*core.Page, error) {
	var p *core.Page
	err := db.View(func(tx *bolt.Tx) error {
		mb, err := db.catalogBucket(tx).Entry(sid, eid)
		if err != nil {
			return err
		}

		pb := mb.PagesBucket()
		if num < 1 || num > pb.Num() {
			return ErrPageNotExist
		}
		p = pb.GetPage(num)

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

func (db *DB) GetSeriesCover(sid string) (*core.Cover, error) {
	var s *core.Cover
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

func (db *DB) GetEntryCover(sid, eid string) (*core.Cover, error) {
	var s *core.Cover
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

func (db *DB) SetSeriesCover(sid string, cover *core.Cover) error {
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

func (db *DB) SetEntryCover(sid, eid string, cover *core.Cover) error {
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
		// delete-missing-entries, the cover's fp should be deleted if it's file does
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
				return nil, "", ErrCoverEmpty
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
				return nil, "", ErrCoverEmpty
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
	errors := core.NewErrorSlice()

	db.Update(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)

		return root.ForEachSeries(func(_ string, sb *SeriesBucket) error {
			seriesThumbnailExists := sb.HasThumbnail()

			// Generate series thumbnail
			if !seriesThumbnailExists || (seriesThumbnailExists && forceNew) {
				c := sb.Cover()
				if c.Fp != "" && c.ExistsOnFS() {
					img, err := c.Thumbnail()
					if err != nil {
						errors = append(errors, err)
						return nil
					}

					err = sb.SetThumbnail(img)
					if err != nil {
						errors = append(errors, err)
						return nil
					}
				}
			}

			// Generate entries thumbnails
			return sb.ForEachEntry(func(_ string, eb *EntryBucket) error {
				thumbnailExists := eb.HasThumbnail()
				if !thumbnailExists || (thumbnailExists && forceNew) {
					// Create thumbnail of custom cover if it exists
					c := eb.Cover()
					if c.Fp != "" && c.ExistsOnFS() {
						img, err := c.Thumbnail()
						if err != nil {
							errors = append(errors, err)
							return nil
						}

						err = sb.SetThumbnail(img)
						if err != nil {
							errors = append(errors, err)
						}

						return nil
					}

					// Otherwise use thumbnail of default cover
					img, err := eb.Archive().Thumbnail()
					if err != nil {
						errors = append(errors, err)
						return nil
					}
					return eb.SetThumbnail(img)
				}
				return nil
			})
		})
	})

	if errors.Empty() {
		return nil
	}
	return errors
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
				img, err := c.Thumbnail()
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
				data, err := c.Thumbnail()
				if err != nil {
					return err
				} else if len(data) == 0 {
					return ErrCoverEmpty
				}
				err = b.SetThumbnail(data)
				if err != nil {
					b.SetThumbnail(nil)
					return err
				}
				return nil
			}

			// If it doesn't exist then use the archive thumbnail
			data, err := b.Archive().Thumbnail()
			if err != nil {
				return err
			}
			return b.SetThumbnail(data)
		}
		return nil
	})
}

func (db *DB) GetSeriesThumbnail(sid string) ([]byte, string, error) {
	return db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		root := db.catalogBucket(tx)

		// Get the custom series cover if exists
		sb, err := root.Series(sid)
		if err != nil {
			return nil, "", err
		}
		if sb.HasThumbnail() {
			return sb.Thumbnail(), "image/jpeg", nil
		}

		// Otherwise get first entry cover
		firstEntry, err := root.FirstEntry(sid)
		if err != nil {
			return nil, "", err
		}

		return firstEntry.Thumbnail(), "image/jpeg", nil
	})
}

func (db *DB) GetEntryThumbnail(sid, eid string) ([]byte, string, error) {
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
		return root.ForEachSeries(func(hash string, b *SeriesBucket) error {
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

		return root.ForEachSeries(func(_ string, sb *SeriesBucket) error {
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

			return sb.ForEachEntry(func(_ string, mb *EntryBucket) error {
				// Check if archive for the entry exists
				if !mb.Archive().Exists() {
					a := mb.Archive()
					e := &api.MissingItem{
						Type:  "Archive",
						Title: a.Title,
						Path:  a.Path,
					}
					items = append(items, e)
				}

				// Check if custom archive cover exists
				c := mb.Cover()
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

		return userRoot.ForEachUser(func(u *UserBucket) error {
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
					eid := entries[i].Hash

					entry, err := seriesRoot.Entry(sid, eid)
					if err == nil && !entry.Archive().Exists() {
						e := &api.MissingItem{
							Type:  "Progress",
							Title: u.Name(),
							Path:  fmt.Sprintf("Series: %s, EntryProgress: %s", sid, eid),
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

		return root.ForEachSeries(func(sid string, sb *SeriesBucket) error {
			// Check if series cover exists
			c := sb.Cover()
			if c.Fp != "" && !c.ExistsOnFS() {
				// If it doesn't exist then reset the cover
				c.Fp = ""
				if err := sb.SetCover(c); err != nil {
					return err
				}
			}

			err := sb.ForEachEntry(func(eid string, mb *EntryBucket) error {
				// Check if archive for the entry exists
				if !mb.Archive().Exists() {
					if err := sb.DeleteEntry(eid); err != nil {
						return err
					}
					// If the entry is deleted we can't retrieve the cover
					return nil
				}

				// Check if custom cover exists
				c := mb.Cover()
				if c.Fp != "" && !c.ExistsOnFS() {
					c.Fp = ""
					if err := mb.SetCover(c); err != nil {
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

			return root.regenerateCatalog()
		})
	})
	if err != nil {
		return err
	}

	// Checks for invalid progress
	err = db.Update(func(tx *bolt.Tx) error {
		userRoot := db.usersBucket(tx)
		seriesRoot := db.catalogBucket(tx)

		return userRoot.ForEachUser(func(u *UserBucket) error {
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

func (db *DB) SetSeriesMetadata(sid string, m *api.EditableSeriesMetadata) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.catalogBucket(tx)

		sm, err := root.SeriesMetadata(sid)
		if err != nil {
			return err
		}

		sm.Title = m.Title
		sm.Author = m.Author
		sm.DateReleased = m.DateReleased

		return root.SetSeriesMetadata(sid, sm)
	})
}

func (db *DB) SetEntryMetadata(sid, eid string, m *api.EditableEntryMetadata) error {
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

		oldM.Title = m.Title
		oldM.Author = m.Author
		oldM.DateReleased = m.DateReleased
		oldM.Chapter = m.Chapter
		oldM.Volume = m.Volume

		return sb.SetEntryMetadata(eid, oldM)
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
