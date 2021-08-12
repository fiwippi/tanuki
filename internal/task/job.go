package task

import (
	"time"

	"github.com/rs/zerolog/log"
)

type Job struct {
	time.Duration

	Ticker *time.Ticker
	Stop   chan struct{}
}

func NewJob(m *Minutes) *Job {
	if m == nil {
		panic("no minutes specified for interval")
	}

	return &Job{Duration: m.Duration}
}

func (i *Job) Run(f func() error, taskName string, runOnStart bool) {
	if i.Duration <= 0 {
		return
	}

	i.Ticker = time.NewTicker(i.Duration)
	i.Stop = make(chan struct{})

	go func() {
		if runOnStart {
			start := time.Now()
			log.Info().Str("task_name", taskName).Msg("running interval task on startup")
			err := f()
			if err != nil {
				log.Error().Err(err).Str("task_name", taskName).Msg("error when running interval task")
			} else {
				log.Info().Str("task_name", taskName).Str("task_time", time.Now().Sub(start).String()).
					Msg("finished running interval task on startup")
			}
		}

		for {
			select {
			case <-i.Ticker.C:
				start := time.Now()
				err := f()
				if err != nil {
					log.Error().Err(err).Str("task_name", taskName).Msg("error when running interval task")
					continue
				}
				timeTaken := time.Now().Sub(start).String()
				log.Debug().Str("task_time", timeTaken).Str("task_name", taskName).Msg("ran interval task")
			case <-i.Stop:
				i.Ticker.Stop()
				return
			}
		}
	}()
}
