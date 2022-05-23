package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func init() {
	// Create the default logger
	logger = logger.Level(zerolog.InfoLevel)
	logger = log.Output(createCW()).
		Level(zerolog.InfoLevel)
}

func Error() *zerolog.Event {
	return logger.Error()
}

func Warn() *zerolog.Event {
	return logger.Warn()
}

func Fatal() *zerolog.Event {
	return logger.Fatal()
}

func Info() *zerolog.Event {
	return logger.Info()
}

func Debug() *zerolog.Event {
	return logger.Debug()
}

func Trace() *zerolog.Event {
	return logger.Trace()
}

func With() zerolog.Context {
	return logger.With()
}
