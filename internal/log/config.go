package log

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Setup(l Level, logToConsole, logToFile bool, fp string) {
	logger = logger.Level(l.Zerolog())

	if logToFile && logToConsole {
		w := zerolog.MultiLevelWriter(createCW(), createFW(fp))
		logger = log.Output(w)
	} else if logToFile {
		logger = log.Output(createFW(fp))
	} else if logToConsole {
		logger = log.Output(createCW())
	} else if !logToFile && !logToConsole {
		Disable()
	}

	Info().Bool("console", logToConsole).Bool("file", logToFile).Str("level", l.String()).Msg("setup logger")
}

func createFW(fp string) zerolog.ConsoleWriter {
	// Create the file writer
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		Fatal().Err(err).Msg("log file could not be created")
	}
	return zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05", NoColor: true}
}

func createCW() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
}
