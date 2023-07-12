package transfer

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/storage"
)

var downloadsPool = NewPool()

type Manager struct {
	*controller
	activeDownloads *DownloadList
	queue           chan *mangadex.Download
	store           *storage.Store
	doneFunc        func() error
	libraryPath     string
	waitingOnDone   bool
}

func NewManager(libraryPath string, workers int, store *storage.Store, done func() error) *Manager {
	m := &Manager{
		store:           store,
		controller:      newController(),
		activeDownloads: NewDownloadList(),
		libraryPath:     libraryPath,
		queue:           make(chan *mangadex.Download, 10),
		doneFunc:        done,
	}

	for id := 0; id < workers; id++ {
		go m.worker(id)
	}

	return m
}

// Internal

func (m *Manager) worker(id int) {
	log.Debug().Int("wid", id).Msg("starting worker")

	for d := range m.queue {
		m.WaitIfPaused()

		log.Info().Str("dl", d.String()).Int("wid", id).Msg("download started")

		// Process the download
		if m.activeDownloads.Has(d) {
			err := d.Run(context.Background(), m.libraryPath)

			// Log the download's success
			l := log.Info()
			if err != nil {
				l = log.Error()
			}
			l.Str("dl", d.String()).Err(err).Str("status", string(d.Status)).Int("wid", id).Msg("download finished")
		}

		// Remove it from the active downloads
		m.activeDownloads.Remove(d)

		// Add the finished download to the store
		err := m.store.AddDownloads(d)
		if err != nil {
			log.Debug().Err(err).Int("wid", id).Str("dl", d.String()).Msg("could not save dl to store")
		}

		// If it's the only download left call the doneFunc
		if len(m.queue) == 0 && len(m.activeDownloads.l) == 0 && !m.waitingOnDone {
			m.Pause()
			m.waitingOnDone = true
			log.Debug().Int("wid", id).Msg("running doneFunc since list is empty")
			err := m.doneFunc()
			if err != nil {
				log.Error().Err(err).Int("wid", id).Msg("error when running doneFunc")
			}
			log.Debug().Int("wid", id).Msg("finished running doneFunc")
			m.waitingOnDone = false
			m.Resume()
		}
	}
}

// State

func (m *Manager) Paused() bool {
	return m.controller.Paused()
}

func (m *Manager) Waiting() bool {
	return m.waitingOnDone
}

// Downloading

func (m *Manager) Queue(mangaTitle string, ch mangadex.Chapter, createSubscription bool) {
	d := ch.CreateDownload(mangaTitle, createSubscription)
	m.activeDownloads.Add(d)
	m.queue <- d
}

func (m *Manager) GetAllDownloads() ([]*mangadex.Download, func(), error) {
	p := m.activeDownloads.List()
	dls, err := m.store.GetAllDownloads()
	if err != nil {
		return nil, nil, err
	}
	p = append(p, dls...)

	doneFunc := func() {
		downloadsPool.Put(p)
	}

	return p, doneFunc, nil
}

func (m *Manager) CancelDownloads() {
	m.activeDownloads.Cancel()
}

func (m *Manager) DeleteSuccessfulDownloads() error {
	return m.store.DeleteSuccessfulDownloads()
}

func (m *Manager) DeleteAllDownloads() error {
	return m.store.DeleteAllDownloads()
}

func (m *Manager) RetryFailedDownloads() error {
	dls, err := m.store.GetFailedDownloads()
	if err != nil {
		return err
	}

	go func() {
		for _, d := range dls {
			m.activeDownloads.Add(d)
			m.queue <- d
		}
	}()

	return nil
}
