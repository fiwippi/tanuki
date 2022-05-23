package log

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

func Disable() {
	logger = logger.Level(zerolog.Disabled)
}

func Setup(c *Config, fp string) {
	logger = logger.Level(c.Level.Zerolog())

	if c.LogToFile && c.LogToConsole {
		w := zerolog.MultiLevelWriter(createCW(), createFW(fp))
		logger = log.Output(w)
	} else if c.LogToFile {
		logger = log.Output(createFW(fp))
	} else if c.LogToConsole {
		logger = log.Output(createCW())
	}

	Info().Bool("console", c.LogToConsole).Bool("file", c.LogToFile).Str("level", c.Level.String()).Msg("setup logger")
}

type Config struct {
	Level        Level `yaml:"level"`
	LogToFile    bool  `yaml:"log_to_file"`
	LogToConsole bool  `yaml:"log_to_console"`
}

func DefaultConfig() Config {
	return Config{
		Level:        InfoLevel,
		LogToFile:    true,
		LogToConsole: true,
	}
}
