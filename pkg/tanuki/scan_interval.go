package tanuki

import (
	"time"

	"github.com/rs/zerolog/log"
)

type ScanInterval struct {
	time.Duration

	Ticker *time.Ticker
	Stop   chan struct{}
}

func parseMinutes(m int) time.Duration {
	return time.Minute * time.Duration(m)
}

func NewInterval(interval int) *ScanInterval {
	return &ScanInterval{Duration: parseMinutes(interval)}
}

func (si *ScanInterval) ChangeInterval(interval int) {
	si.Duration = parseMinutes(interval)

	if si.Ticker != nil {
		si.Ticker.Reset(si.Duration)
	}
}

func (si *ScanInterval) MarshalYAML() (interface{}, error) {
	return int(si.Minutes()), nil
}

func (si *ScanInterval) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var interval int
	if err := unmarshal(&interval); err != nil {
		return err
	}
	si.Duration = parseMinutes(interval)

	return nil
}

func (si *ScanInterval) RunTask(f func() error, taskName string)  {
	si.Ticker = time.NewTicker(si.Duration)
	si.Stop = make(chan struct{})

	go func() {
		for {
			select {
			case <- si.Ticker.C:
				err := f()
				if err != nil {
					log.Error().Err(err).Str("task name", taskName).Msg("error when running interval task")
				}
			case <- si.Stop:
				si.Ticker.Stop()
				return
			}
		}
	}()
}
