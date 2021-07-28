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

func (i *Job) Run(f func() error, taskName string) {
	i.Ticker = time.NewTicker(i.Duration)
	i.Stop = make(chan struct{})

	go func() {
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
