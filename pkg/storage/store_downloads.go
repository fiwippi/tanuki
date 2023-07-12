package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/pkg/mangadex"
)

func (s *Store) AddDownloads(dls ...*mangadex.Download) error {
	fn := func(tx *sqlx.Tx) error {
		for _, d := range dls {
			stmt := `
				INSERT INTO downloads 
    				(manga_title, chapter, status, current_page, total_pages, time_taken) 
				Values 
					(:manga_title, :chapter, :status, :current_page, :total_pages, :time_taken)`
			_, err := tx.NamedExec(stmt, d)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return s.tx(fn)
}

func (s *Store) getDownloadsByStatus(sts ...mangadex.DownloadStatus) ([]*mangadex.Download, error) {
	var dls []*mangadex.Download
	fn := func(tx *sqlx.Tx) error {
		query, args, err := sqlx.In(`SELECT * FROM downloads WHERE status IN (?) ORDER BY ROWID DESC `, sts)
		if err != nil {
			return err
		}

		return tx.Select(&dls, tx.Rebind(query), args...)
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return dls, nil
}

func (s *Store) GetAllDownloads() ([]*mangadex.Download, error) {
	return s.getDownloadsByStatus(mangadex.DownloadStatuses...)
}

func (s *Store) GetFailedDownloads() ([]*mangadex.Download, error) {
	return s.getDownloadsByStatus(mangadex.DownloadFailed)
}

func (s *Store) deleteDownloadsByStatus(sts ...mangadex.DownloadStatus) error {
	fn := func(tx *sqlx.Tx) error {
		query, args, err := sqlx.In(`DELETE FROM downloads WHERE status IN (?)`, sts)
		if err != nil {
			return err
		}

		_, err = tx.Exec(tx.Rebind(query), args...)
		return err
	}

	return s.tx(fn)
}

func (s *Store) DeleteAllDownloads() error {
	return s.deleteDownloadsByStatus(mangadex.DownloadStatuses...)
}

func (s *Store) DeleteSuccessfulDownloads() error {
	return s.deleteDownloadsByStatus(mangadex.DownloadCancelled, mangadex.DownloadExists, mangadex.DownloadFinished)
}
