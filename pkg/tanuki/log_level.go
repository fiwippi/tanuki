package tanuki

import (
	"errors"

	"github.com/rs/zerolog"
)

// LogLevel wraps zerolog.Level in order to
// marshal into the config.yml file as a string
type LogLevel zerolog.Level

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel

	TraceLevel LogLevel = -1
)

func (l LogLevel) String() string {
	return l.Zerolog().String()
}

func (l LogLevel) Zerolog() zerolog.Level {
	return zerolog.Level(l)
}

func (l LogLevel) MarshalYAML() (interface{}, error) {
	return l.String(), nil
}

func (l LogLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var level string
	if err := unmarshal(&level); err != nil {
		return err
	}

	switch level {
	case "panic":
		l = LogLevel(zerolog.PanicLevel)
		return nil
	case "fatal":
		l = LogLevel(zerolog.FatalLevel)
		return nil
	case "error":
		l = LogLevel(zerolog.ErrorLevel)
		return nil
	case "warn":
		l = LogLevel(zerolog.WarnLevel)
		return nil
	case "info":
		l = LogLevel(zerolog.InfoLevel)
		return nil
	case "debug":
		l = LogLevel(zerolog.DebugLevel)
		return nil
	case "trace":
		l = LogLevel(zerolog.TraceLevel)
		return nil
	}

	return errors.New("invalid log level name")
}
