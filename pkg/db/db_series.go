package db

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/sets"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/core"
)

func (db *DB) SaveSeries(s *core.Series, m []*core.Manga) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.seriesListBucket(tx)
		return root.AddSeries(s, m)
	})
}

func (db *DB) GenerateSeriesList() api.SeriesList {
	list := make(api.SeriesList, 0)
	db.View(func(tx *bolt.Tx) error {
		root := db.seriesListBucket(tx)
		return root.ForEachSeries(func(hash string, b *SeriesBucket) error {
			list = append(list, b.ApiSeries())
			return nil
		})
	})

	sort.SliceStable(list, func(i, j int) bool {
		return strings.ToLower(list[i].Title) < strings.ToLower(list[j].Title)
	})

	return list
}

func (db *DB) GenerateSeriesThumbnail(seriesHash string, forceNew bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
		if err != nil {
			return err
		}

		seriesThumbnailExists := sb.HasThumbnail()
		if !seriesThumbnailExists || (seriesThumbnailExists && forceNew) {
			c := sb.Cover()
			if c.Fp != "" && c.ExistsOnFS() {
				img, err := c.ThumbnailFromFS()
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

func (db *DB) GenerateThumbnails(forceNew bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.seriesListBucket(tx)

		return root.ForEachSeries(func(_ string, sb *SeriesBucket) error {
			seriesThumbnailExists := sb.HasThumbnail()

			// Generate series thumbnail
			if !seriesThumbnailExists || (seriesThumbnailExists && forceNew) {
				c := sb.Cover()
				if c.Fp != "" && c.ExistsOnFS() {
					img, err := c.ThumbnailFromFS()
					if err != nil {
						return err
					}

					err = sb.SetThumbnail(img)
					if err != nil {
						return err
					}
				}
			}

			// Generate entries thumbnails
			return sb.ForEachEntry(func(_ string, mb *MangaBucket) error {
				thumbnailExists := mb.HasThumbnail()
				if !thumbnailExists || (thumbnailExists && forceNew) {
					img, err := mb.Archive().Thumbnail()
					if err != nil {
						return err
					}
					return mb.SetThumbnail(img)
				}
				return nil
			})
		})
	})
}

func (db *DB) GenerateEntryThumbnail(sid, eid string, forceNew bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := db.seriesListBucket(tx).GetEntry(sid, eid)
		if err != nil {
			return err
		}

		hasThumb := b.HasThumbnail()
		if !hasThumb || (hasThumb && forceNew) {
			// If custom cover exists try and create it
			c := b.Cover()
			if c.Fp != "" && c.ExistsOnFS() {
				data, err := c.ThumbnailFromFS()
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

func (db *DB) GetSeriesCover(seriesHash string) (*core.Cover, error) {
	var s *core.Cover
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
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
		sb, err := db.seriesListBucket(tx).GetEntry(sid, eid)
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

func (db *DB) SetSeriesCover(seriesHash string, cover *core.Cover) error {
	err := db.Update(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
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
		sb, err := db.seriesListBucket(tx).GetEntry(sid, eid)
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

func (db *DB) GetSeriesCoverBytes(seriesHash string) ([]byte, string, error) {
	return db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		root := db.seriesListBucket(tx)

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
		sb, err := root.GetSeries(seriesHash)
		if err != nil {
			return nil, "", err
		}
		// 2
		c := sb.Cover()
		// 3
		if c.Fp != "" && c.ExistsOnFS() {
			// 4
			data, err := c.FromFS()
			if err != nil {
				return nil, "", err
			} else if len(data) == 0 {
				return nil, "", ErrCoverEmpty
			}

			return data, c.ImageType.MimeType(), nil
		}

		// Otherwise get the cover from the first series entry
		firstEntry, err := root.GetFirstEntry(seriesHash)
		if err != nil {
			return nil, "", err
		}
		c = firstEntry.Cover()
		tempData, err := firstEntry.ArchiveCoverBytes()
		if err != nil {
			return nil, "", err
		}
		return tempData, c.ImageType.MimeType(), nil
	})
}

func (db *DB) GetSeriesEntryCoverBytes(sid, eid string) ([]byte, string, error) {
	return db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		root := db.seriesListBucket(tx)

		// 1
		mb, err := root.GetEntry(sid, eid)
		if err != nil {
			return nil, "", err
		}
		// 2
		c := mb.Cover()
		// 3
		if c.Fp != "" && c.ExistsOnFS() {
			// 4
			data, err := c.FromFS()
			if err != nil {
				return nil, "", err
			} else if len(data) == 0 {
				return nil, "", ErrCoverEmpty
			}

			return data, c.ImageType.MimeType(), nil
		}

		// Otherwise get the embedded cover
		c = mb.ArchiveCover()
		tempData, err := mb.ArchiveCoverBytes()
		if err != nil {
			return nil, "", err
		}
		return tempData, c.ImageType.MimeType(), nil
	})
}

func (db *DB) DeleteEntryCover(sid, eid string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := db.seriesListBucket(tx).GetEntry(sid, eid)
		if err != nil {
			return err
		}

		// Get the cover
		c := b.Cover()

		// Delete the cover image from the filesystem
		// and its directory if it's left empty
		os.Remove(c.Fp)
		fse.DeleteFileDirIfEmpty(c.Fp)

		// Clean references to the file
		c.Fp = ""
		err = b.SetCover(c)
		if err != nil {
			return err
		}

		// Delete the thumbnail as well
		return b.SetThumbnail(nil)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) DeleteSeriesCover(seriesHash string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
		if err != nil {
			return err
		}

		// Get the cover
		c := sb.Cover()

		// Delete the cover image from the filesystem
		// and its directory if it's left empty
		os.Remove(c.Fp)
		fse.DeleteFileDirIfEmpty(c.Fp)

		// Clean references to the file
		c.Fp = ""
		err = sb.SetCover(c)
		if err != nil {
			return err
		}

		// Delete the thumbnail as well
		return sb.SetThumbnail(nil)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetSeriesThumbnail(seriesHash string) ([]byte, string, error) {
	return db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		root := db.seriesListBucket(tx)

		// Get the custom series cover if exists
		sb, err := root.GetSeries(seriesHash)
		if err != nil {
			return nil, "", err
		}
		if sb.HasThumbnail() {
			return sb.Thumbnail(), "image/jpeg", nil
		}

		// Otherwise get first entry cover
		firstEntry, err := root.GetFirstEntry(seriesHash)
		if err != nil {
			return nil, "", err
		}

		return firstEntry.Thumbnail(), "image/jpeg", nil
	})
}

func (db *DB) GetEntryThumbnail(sid, eid string) ([]byte, string, error) {
	return db.returnBytes(func(tx *bolt.Tx) ([]byte, string, error) {
		b, err := db.seriesListBucket(tx).GetEntry(sid, eid)
		if err != nil {
			return nil, "", err
		}

		//// If a thumbnail doesn't exist then try and use the default embedded one
		//if !b.HasThumbnail() {
		//	data, err := b.Archive().Thumbnail()
		//	if err != nil {
		//		return nil, "", err
		//	}
		//	err =  b.SetThumbnail(data)
		//	if err != nil {
		//		return nil, "", err
		//	}
		//}

		return b.Thumbnail(), "image/jpeg", nil
	})
}

func (db *DB) GetMissingEntries() api.MissingEntries {
	entries := make(api.MissingEntries, 0)

	// Checks for invalid archive and cover
	db.View(func(tx *bolt.Tx) error {
		root := db.seriesListBucket(tx)

		return root.ForEachSeries(func(_ string, sb *SeriesBucket) error {
			// Check if series cover exists
			c := sb.Cover()
			if c.Fp != "" && !c.ExistsOnFS() {
				e := &api.MissingEntry{
					Type:  "Cover",
					Title: fse.FilenameWithExt(c.Fp),
					Path:  c.Fp,
				}
				entries = append(entries, e)
			}

			return sb.ForEachEntry(func(_ string, mb *MangaBucket) error {
				// Check if archive for the entry exists
				if !mb.Archive().Exists() {
					e := &api.MissingEntry{
						Type:  "Archive",
						Title: mb.Title(),
						Path:  mb.Archive().Path,
					}
					entries = append(entries, e)
				}

				// Check if custom archive cover exists
				c := mb.Cover()
				if c.Fp != "" && !c.ExistsOnFS() {
					e := &api.MissingEntry{
						Type:  "Cover",
						Title: fse.FilenameWithExt(c.Fp),
						Path:  c.Fp,
					}
					fmt.Println(fse.FilenameWithExt(c.Fp))
					entries = append(entries, e)
				}

				return nil
			})
		})
	})

	// Checks for invalid progress
	db.View(func(tx *bolt.Tx) error {
		userRoot := db.usersBucket(tx)
		seriesRoot := db.seriesListBucket(tx)

		return userRoot.ForEachUser(func(u *UserBucket) error {
			tracker := u.ProgressTracker()
			if tracker != nil {
				for sid := range tracker.Tracker {
					for eid := range tracker.Tracker[sid] {
						entry, err := seriesRoot.GetEntry(sid, eid)
						if err == nil && !entry.Archive().Exists() {
							e := &api.MissingEntry{
								Type:  "Progress",
								Title: u.Name(),
								Path:  fmt.Sprintf("Series: %s, Entry: %s", sid, eid),
							}
							entries = append(entries, e)
						}

					}
				}
			}

			return nil
		})
	})

	return entries
}

func (db *DB) DeleteMissingEntries() error {
	// Checks for invalid archive and cover
	err := db.Update(func(tx *bolt.Tx) error {
		root := db.seriesListBucket(tx)

		return root.ForEachSeries(func(_ string, sb *SeriesBucket) error {
			// Check if series cover exists
			c := sb.Cover()
			if c.Fp != "" && !c.ExistsOnFS() {
				c.Fp = ""
				if err := sb.SetCover(c); err != nil {
					return err
				}
			}

			return sb.ForEachEntry(func(mbhash string, mb *MangaBucket) error {
				// Check if archive for the entry exists
				if !mb.Archive().Exists() {
					if err := sb.DeleteEntry([]byte(mbhash)); err != nil {
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
		})
	})
	if err != nil {
		return err
	}

	// Checks for invalid progress
	err = db.Update(func(tx *bolt.Tx) error {
		userRoot := db.usersBucket(tx)
		seriesRoot := db.seriesListBucket(tx)

		return userRoot.ForEachUser(func(u *UserBucket) error {
			tracker := u.ProgressTracker()
			if tracker != nil {
				for sid := range tracker.Tracker {
					for eid := range tracker.Tracker[sid] {
						entry, err := seriesRoot.GetEntry(sid, eid)
						if err == nil && !entry.Archive().Exists() {
							delete(tracker.Tracker[sid], eid)
							return u.ChangeProgressTracker(tracker)
						}
					}
				}
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) HasSeries(seriesHash string) bool {
	err := db.View(func(tx *bolt.Tx) error {
		_, err := db.seriesListBucket(tx).GetSeries(seriesHash)
		if err != nil {
			return err
		}
		return nil
	})
	return err == nil
}

func (db *DB) HasSeriesEntry(seriesHash, entryHash string) bool {
	err := db.View(func(tx *bolt.Tx) error {
		_, err := db.seriesListBucket(tx).GetEntry(seriesHash, entryHash)
		if err == nil {
			return err
		}
		return nil
	})
	return err == nil
}

func (db *DB) GetSeriesEntries(seriesHash string) (api.SeriesEntries, error) {
	var entries = make(api.SeriesEntries, 0)
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
		if err != nil {
			return err
		}

		return sb.ForEachEntry(func(hash string, mb *MangaBucket) error {
			entries = append(entries, mb.ApiSeriesEntry())
			return nil
		})

	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Title) < strings.ToLower(entries[j].Title)
	})

	return entries, nil
}

func (db *DB) GetSeries(seriesHash string) (*api.Series, error) {
	var e *api.Series
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
		if err != nil {
			return err
		}
		e = sb.ApiSeries()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (db *DB) GetSeriesFolderTitle(seriesHash string) (string, error) {
	var s string
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
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

func (db *DB) GetEntry(seriesHash, entryHash string) (*api.SeriesEntry, error) {
	var e *api.SeriesEntry
	err := db.View(func(tx *bolt.Tx) error {
		mb, err := db.seriesListBucket(tx).GetEntry(seriesHash, entryHash)
		if err != nil {
			return err
		}
		e = mb.ApiSeriesEntry()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (db *DB) GetSeriesEntryArchive(seriesHash, entryHash string) (*core.Archive, error) {
	var a *core.Archive
	err := db.View(func(tx *bolt.Tx) error {
		mb, err := db.seriesListBucket(tx).GetEntry(seriesHash, entryHash)
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

func (db *DB) GetSeriesEntryPage(seriesHash, entryHash string, num int) (*core.Page, error) {
	var p *core.Page
	err := db.View(func(tx *bolt.Tx) error {
		mb, err := db.seriesListBucket(tx).GetEntry(seriesHash, entryHash)
		if err != nil {
			return err
		}
		p = mb.Pages().GetPage(num)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (db *DB) SetSeriesTags(seriesHash string, tags []string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
		if err != nil {
			return err
		}
		t := sb.Tags()
		t.Clear()
		t.Add(tags...)
		return sb.SetTags(t)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetSeriesTags(seriesHash string) (*sets.Set, error) {
	var t *sets.Set
	err := db.View(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
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
		root := db.seriesListBucket(tx)
		return root.ForEachSeries(func(hash string, b *SeriesBucket) error {
			tags.Add(b.Tags().List()...)
			return nil
		})
	})
	return tags
}

func (db *DB) GetSeriesWithTag(t string) api.SeriesList {
	list := make(api.SeriesList, 0)

	// We can ignore the error since we aren't returning
	// any errors in the ForEach traversal code
	db.View(func(tx *bolt.Tx) error {
		root := db.seriesListBucket(tx)
		return root.ForEachSeries(func(hash string, b *SeriesBucket) error {
			if b.Tags().Has(t) {
				list = append(list, b.ApiSeries())
			}
			return nil
		})
	})
	return list
}

func (db *DB) SetSeriesMetadata(seriesHash string, metadata *core.SeriesMetadata) error {
	err := db.Update(func(tx *bolt.Tx) error {
		sb, err := db.seriesListBucket(tx).GetSeries(seriesHash)
		if err != nil {
			return err
		}
		return sb.SetMetadata(metadata)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetEntryMetadata(sid, eid string, metadata *core.EntryMetadata) error {
	err := db.Update(func(tx *bolt.Tx) error {
		mb, err := db.seriesListBucket(tx).GetEntry(sid, eid)
		if err != nil {
			return err
		}
		return mb.SetMetadata(metadata)
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetEntryModTime(sid, eid string) (time.Time, error) {
	var t time.Time
	err := db.View(func(tx *bolt.Tx) error {
		b, err := db.seriesListBucket(tx).GetEntry(sid, eid)
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
		sb, err := db.seriesListBucket(tx).GetSeries(sid)
		if err != nil {
			return err
		}

		return sb.ForEachEntry(func(_ string, mb *MangaBucket) error {
			entryTime := mb.Archive().ModTime
			if t == (time.Time{}) {
				t = entryTime
			} else if entryTime.Before(t) {
				t = entryTime
			}
			return nil
		})
	})
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (db *DB) GetCatalogModTime() (time.Time, error) {
	var t time.Time
	err := db.View(func(tx *bolt.Tx) error {
		root := db.seriesListBucket(tx)
		return root.ForEachSeries(func(sid string, _ *SeriesBucket) error {
			modTime, err := db.GetSeriesModTime(sid)
			if err != nil {
				return err
			}

			if t == (time.Time{}) {
				t = modTime
			} else if modTime.Before(t) {
				t = modTime
			}
			return nil
		})
	})
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
