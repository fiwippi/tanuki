package tanuki

import (
	"time"

	"github.com/rs/zerolog/log"
)

type Interval struct {
	time.Duration

	Ticker *time.Ticker
	Stop   chan struct{}
}

func parseMinutes(m int) time.Duration {
	return time.Minute * time.Duration(m)
}

func NewInterval(interval int) *Interval {
	return &Interval{Duration: parseMinutes(interval)}
}

func (i *Interval) ChangeInterval(interval int) {
	i.Duration = parseMinutes(interval)

	if i.Ticker != nil {
		i.Ticker.Reset(i.Duration)
	}
}

func (i *Interval) MarshalYAML() (interface{}, error) {
	return int(i.Minutes()), nil
}

func (i *Interval) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var interval int
	if err := unmarshal(&interval); err != nil {
		return err
	}
	i.Duration = parseMinutes(interval)

	return nil
}

func (i *Interval) RunTask(f func() error, taskName string) {
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
