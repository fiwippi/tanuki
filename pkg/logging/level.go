package logging

import (
	"errors"

	"github.com/rs/zerolog"
)

// Level wraps zerolog.Level in order to
// marshal into the config.yml file as a string
type Level zerolog.Level

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel

	TraceLevel Level = -1
)

func (l Level) String() string {
	return l.Zerolog().String()
}

func (l Level) Zerolog() zerolog.Level {
	return zerolog.Level(l)
}

func (l Level) MarshalYAML() (interface{}, error) {
	return l.String(), nil
}

func (l Level) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var level string
	if err := unmarshal(&level); err != nil {
		return err
	}

	switch level {
	case "panic":
		l = Level(zerolog.PanicLevel)
		return nil
	case "fatal":
		l = Level(zerolog.FatalLevel)
		return nil
	case "error":
		l = Level(zerolog.ErrorLevel)
		return nil
	case "warn":
		l = Level(zerolog.WarnLevel)
		return nil
	case "info":
		l = Level(zerolog.InfoLevel)
		return nil
	case "debug":
		l = Level(zerolog.DebugLevel)
		return nil
	case "trace":
		l = Level(zerolog.TraceLevel)
		return nil
	}

	return errors.New("invalid log level name")
}
