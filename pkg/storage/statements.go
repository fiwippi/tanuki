package storage

const (
	getSeriesStmt = `
		SELECT 
			sid, folder_title, num_entries, num_pages, mod_time, 
			display_title, tags, mangadex_uuid, mangadex_last_published_at 
		FROM series WHERE sid = ? ORDER BY ROWID DESC`
)
