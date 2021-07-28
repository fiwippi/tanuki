package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logFileWriter zerolog.ConsoleWriter
var logConsoleWriter zerolog.ConsoleWriter

func init() {
	// Set Info as the default level and by default log to file and console
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create the basic console logger
	logConsoleWriter = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	log.Logger = log.Output(logConsoleWriter)
}

func SetupLogger(c *Config, fp string) {
	// Setup the log level
	zerolog.SetGlobalLevel(c.Level.Zerolog())

	// Decide which outputs the log will have
	if c.LogToFile {
		// Create the file writer
		f, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal().Err(err).Msg("log file could not be created")
		}
		logFileWriter = zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05", NoColor: true}

		if c.LogToConsole {
			w := zerolog.MultiLevelWriter(logConsoleWriter, logFileWriter)
			log.Logger = log.Output(w)
		} else {
			log.Logger = log.Output(logFileWriter)
		}
	} else if c.LogToConsole {
		log.Logger = log.Output(logConsoleWriter)
	}

	log.Info().Bool("console", c.LogToConsole).Bool("file", c.LogToFile).Str("level", c.Level.String()).Msg("setup logger")
}
