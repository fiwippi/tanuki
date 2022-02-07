package downloading

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/sync"
	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/store/bolt"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

var downloadsPool = NewPool()

// Manager which synchronises downloads from Mangadex
type Manager struct {
	root      string             // Root path to download to
	downloads *DownloadList      // List of active downloads
	queue     chan *api.Download // Channel which feeds downloads to the workers
	mangadex  *mangadex.Client   // Mangadex client used to request chapters
	store     *bolt.DB           // Store for keeping track of past downloads
	cont      *sync.Controller   // Controller for workers so they can be stopped/paused/canceled
}

func NewManager(c *mangadex.Client, root string, store *bolt.DB, workers int) *Manager {
	m := &Manager{
		root:      root,
		queue:     make(chan *api.Download, 10),
		downloads: NewDownloadList(),
		mangadex:  c,
		store:     store,
		cont:      sync.NewController(),
	}

	for id := 0; id < workers; id++ {
		go m.worker(id)
	}

	return m
}

func (m *Manager) Paused() bool {
	return m.cont.Paused()
}

func (m *Manager) Pause() {
	m.cont.Pause()
}

func (m *Manager) Resume() {
	m.cont.Resume()
}

func (m *Manager) Cancel() {
	m.Resume() // Ensures paused downloads also get cancelled
	m.downloads.Cancel()
}

// Downloading

func (m *Manager) StartDownload(manga string, ch *mangadex.Chapter) {
	d := api.NewDownload(manga, ch)
	m.downloads.Add(d)
	m.queue <- d
}

func (m *Manager) worker(id int) {
	log.Debug().Int("wid", id).Msg("starting worker")
	for d := range m.queue {
		log.Debug().Int("wid", id).Str("key", d.Key()).Msg("worker received download")

		// We might not have the download because it's been cancelled
		if m.downloads.Has(d) {
			// Start the download
			d.Start()

			// Check if download exists
			fp := fmt.Sprintf("%s/%s", m.root, d.Filepath())
			if fse.Exists(fp) {
				log.Debug().Int("wid", id).Str("key", d.Key()).Str("state", "already exists").Msg("download finished")
				d.FinishExists()
			} else {
				// If it doesn't then execute it
				err := m.download(d, fp)
				if err != nil {
					cancelled := errors.Is(err, api.ErrDownloadCancelled)
					log.Debug().Err(err).Int("wid", id).Str("key", d.Key()).Bool("cancelled", cancelled).Msg("error while downloading")
					if cancelled {
						d.FinishCancelled()
					} else {
						d.FinishFailed()
					}
				} else {
					log.Debug().Int("wid", id).Str("key", d.Key()).Str("state", "finished").Msg("download finished")
					d.Finish()
				}
			}

			// Delete the download from the map
			m.downloads.Remove(d)
		}

		// Save the download to the db
		err := m.store.AddDownload(d)
		if err != nil {
			log.Debug().Err(err).Int("wid", id).Str("key", d.Key()).Msg("could not save dl to db")
		}
	}
}

func (m *Manager) download(d *api.Download, fp string) error {
	// Get the home url
	chData, err := m.mangadex.GetHomeUrl(d.Chapter.ID)
	if err != nil {
		return fmt.Errorf("could not get mangadex home url: %w", err)
	}

	// Ensure valid data
	if len(chData.Chapter.Data) == 0 {
		return fmt.Errorf("no pages exist for chapter: %s", d.Chapter.Attributes.Title)
	}
	if chData.Chapter.Hash == "" {
		return fmt.Errorf("no hash exists for chapter: %s", d.Chapter.Attributes.Title)
	}

	m.cont.WaitIfPaused()
	if d.Status == api.DownloadCancelled {
		return api.ErrDownloadCancelled
	}

	// Get the archive file
	forChapter := func(i int) error {
		d.CurrentPage = i
		if d.Status == api.DownloadCancelled {
			return api.ErrDownloadCancelled
		}
		return nil
	}
	archive, err := m.mangadex.CreateChapterArchive(d.Chapter, chData, forChapter, m.cont)
	if err != nil {
		return fmt.Errorf("could not get create the archive file: %w", err)
	}

	// Save the archive to disk
	err = fse.EnsureWriteFile(fp, archive.Data(), 0666)
	if err != nil {
		return fmt.Errorf("could not save archive file: %w", err)
	}

	return nil
}

func (m *Manager) Downloads() ([]*api.Download, func()) {
	p := m.downloads.List()
	p = append(p, m.store.GetDownloads()...)

	doneFunc := func() {
		downloadsPool.Put(p)
	}

	return p, doneFunc
}

// Editing the manager state

func (m *Manager) DeleteFinishedTasks() error {
	return m.store.ClearFinishedDownloads()
}

func (m *Manager) RetryFailedTasks() {
	tasks := m.store.GetFailedDownloads()
	for _, d := range tasks {
		m.downloads.Add(d)
		m.queue <- d
	}
}
