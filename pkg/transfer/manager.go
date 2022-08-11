package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/mangadex"
	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/internal/platform/fse"
	"github.com/fiwippi/tanuki/pkg/manga"
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

func NewManager(libraryPath string, workers int, store *storage.Store, done func() error, sbInterval time.Duration) *Manager {
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
	if sbInterval > 0 {
		go m.checkSubscriptionsOnInterval(sbInterval)

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

			// If the download was successful and user wants to subscribe
			// to the series then add the series' uuid to the database
			if d.Subscribe && err == nil {
				// Get the SID of the download
				folderPath := fmt.Sprintf("%s/%s", m.libraryPath, fse.Sanitise(d.MangaTitle))
				sid, err := manga.FolderID(folderPath)
				if err != nil {
					log.Error().Err(err).Str("dl", d.String()).Int("wid", id).Msg("failed to get sid for finished dl")
				}

				// If successful then update the last published in the subscriptions db
				err = m.store.SetSubscriptionWithTime(sid, d.MangaTitle, dbutil.NullString(d.Chapter.SeriesID), d.Chapter.PublishedAt, true)
				if err != nil {
					log.Error().Err(err).Str("dl", d.String()).Int("wid", id).Msg("failed to set dl subscription in db")
				}
			}
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

// Subscriptions

func (m *Manager) CheckSubscriptions() error {
	// Get all subscriptions
	sbs, err := m.store.GetAllSubscriptions()
	if err != nil {
		return err
	}

	// Download each subscription
	for _, sb := range sbs {
		title := sb.Title
		series, err := m.store.GetSeries(sb.SID)
		if err == nil {
			title = series.FolderTitle
		}

		// Get all new chapters for the subscription
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		chs, err := mangadex.NewChapters(ctx, string(sb.MdexUUID), sb.MdexLastPublishedAt.Time())
		cancel()
		if err != nil {
			return err
		}

		// Queue the new chapter for downloading
		for _, ch := range chs {
			m.Queue(title, ch, true)
		}
	}

	return nil
}

func (m *Manager) checkSubscriptionsOnInterval(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info().Msg("checking subscriptions for new chapters")

			err := m.CheckSubscriptions()
			if err != nil {
				log.Error().Err(err).Msg("error when checking subscription")
			}
		}
	}
}
