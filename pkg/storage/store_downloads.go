package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/mangadex"
)

func (s *Store) AddDownloads(dls ...*mangadex.Download) error {
	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, d := range dls {
		stmt := `
		INSERT INTO downloads 
    		(manga_title, chapter, status, current_page, total_pages, time_taken) 
		Values 
			(:manga_title, :chapter, :status, :current_page, :total_pages, :time_taken)`
		_, err = tx.NamedExec(stmt, d)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Store) getDownloadsByStatus(sts ...mangadex.DownloadStatus) ([]*mangadex.Download, error) {
	tx, err := s.pool.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query, args, err := sqlx.In(`SELECT * FROM downloads WHERE status IN (?) ORDER BY ROWID DESC `, sts)
	if err != nil {
		return nil, err
	}

	var dls []*mangadex.Download
	err = tx.Select(&dls, s.pool.Rebind(query), args...)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
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
	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query, args, err := sqlx.In(`DELETE FROM downloads WHERE status IN (?)`, sts)
	if err != nil {
		return err
	}
	_, err = tx.Exec(s.pool.Rebind(query), args...)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteAllDownloads() error {
	return s.deleteDownloadsByStatus(mangadex.DownloadStatuses...)
}

func (s *Store) DeleteSuccessfulDownloads() error {
	return s.deleteDownloadsByStatus(mangadex.DownloadCancelled, mangadex.DownloadExists, mangadex.DownloadFinished)
}
