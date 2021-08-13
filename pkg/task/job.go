// Package task implements jobs which can run at given intervals
package task

import (
	"time"

	"github.com/rs/zerolog/log"
)

// Job runs a given function using its duration as the interval
type Job struct {
	time.Duration
}

// NewJob creates a new job using the given amount of Minutes as
// the interval.
func NewJob(m int) *Job {
	return &Job{Duration: minutesDuration(m)}
}

func minutesDuration(m int) time.Duration {
	return time.Minute * time.Duration(m)
}

// Run executes a f at intervals specified by Job's duration. runOnStart can
// be specified to have f run when Run is called instead of only when the first
// tick from the ticker.
func (j *Job) Run(f func() error, taskName string, runOnStart bool) {
	if j.Duration <= 0 {
		return
	}

	ticker := time.NewTicker(j.Duration)

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
			case <-ticker.C:
				start := time.Now()
				err := f()
				if err != nil {
					log.Error().Err(err).Str("task_name", taskName).Msg("error when running interval task")
					continue
				}
				timeTaken := time.Now().Sub(start).String()
				log.Debug().Str("task_time", timeTaken).Str("task_name", taskName).Msg("ran interval task")
			}
		}
	}()
}
