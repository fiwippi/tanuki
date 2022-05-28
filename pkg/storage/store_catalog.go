package storage

import (
	"github.com/fiwippi/tanuki/pkg/manga"
)

func (s *Store) GetCatalog() ([]*manga.Series, error) {
	var v []*manga.Series
	stmt := `
		SELECT 
			sid, folder_title, num_entries, num_pages, mod_time, 
			display_title, tags, mangadex_uuid, mangadex_last_published_at 
		FROM series ORDER BY ROWID ASC`
	err := s.pool.Select(&v, stmt)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// TODO if series a is deleted and then put back will it order correcty or put it at the end

// TODO generate thumbnails

// TODO get missing items

// TODO delete missing items

// Missing items related TODO: put this in store_catalog
