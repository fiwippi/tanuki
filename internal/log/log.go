package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func init() {
	// Create the default logger
	logger = log.Output(createCW()).Level(zerolog.InfoLevel)
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

func Panic() *zerolog.Event {
	return logger.Panic()
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

func Disable() {
	logger = logger.Level(zerolog.Disabled)
}

func Copy() zerolog.Logger {
	return logger.With().Logger()
}
