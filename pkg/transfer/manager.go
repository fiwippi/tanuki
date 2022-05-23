package transfer

import (
	"context"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/mangadex"
)

var downloadsPool = NewPool()

type Manager struct {
	libraryPath     string
	activeDownloads *DownloadList
	queue           chan *mangadex.Download
}

func NewManager(libraryPath string, workers int) *Manager {
	m := &Manager{
		libraryPath:     libraryPath,
		activeDownloads: NewDownloadList(),
		queue:           make(chan *mangadex.Download, 10),
	}

	for id := 0; id < workers; id++ {
		go m.worker(id)
	}

	return m
}

func (m *Manager) worker(id int) {
	log.Debug().Int("wid", id).Msg("starting worker")

	for d := range m.queue {
		log.Info().Str("dl", d.String()).Int("wid", id).Msg("download started")

		// Process the download
		if m.activeDownloads.Has(d) {
			err := d.Run(context.Background(), m.libraryPath)

			l := log.Info()
			if err != nil {
				l = log.Error()
			}
			l.Str("dl", d.String()).Err(err).Str("status", string(d.Status)).Int("wid", id).Msg("download finished")
		}

		// Remove it from the active downloads
		m.activeDownloads.Remove(d)

		// TODO: add the download to the store

		// TODO: if no other downloads are running then call the done func (scan the library)
	}
}

// State of downloads

func (m *Manager) Cancel() {
	m.activeDownloads.Cancel()
}

func (m *Manager) List() ([]*mangadex.Download, func()) {
	p := m.activeDownloads.List()
	//p = append(p, m.store.GetDownloads()...) TODO: appedn this

	doneFunc := func() {
		downloadsPool.Put(p)
	}

	return p, doneFunc
}

// Downloading

func (m *Manager) Queue(mangaTitle string, ch mangadex.Chapter) {
	d := ch.CreateDownload(mangaTitle)
	m.activeDownloads.Add(d)
	m.queue <- d
}
