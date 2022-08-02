package config

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/internal/platform/encryption"
	"github.com/fiwippi/tanuki/internal/platform/fse"
	"github.com/fiwippi/tanuki/internal/platform/task"
)

const (
	ScanInterval = 180 // Every 3 hours
)

type Config struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	Logging struct {
		Level        log.Level `yaml:"level"`
		LogToFile    bool      `yaml:"log_to_file"`
		LogToConsole bool      `yaml:"log_to_console"`
	} `yaml:"logging"`
	Paths                  Paths           `yaml:"paths"`
	SessionSecret          *encryption.Key `yaml:"session_secret"`
	ScanInterval           *task.Job       `yaml:"scan_interval_minutes"`
	MaxUploadedFileSizeMiB int             `yaml:"max_uploaded_file_size_mib"`
	DebugMode              bool            `yaml:"debug_mode"`
}

func DefaultConfig() *Config {
	defautLog := struct {
		Level        log.Level `yaml:"level"`
		LogToFile    bool      `yaml:"log_to_file"`
		LogToConsole bool      `yaml:"log_to_console"`
	}{Level: log.InfoLevel, LogToFile: true, LogToConsole: true}

	return &Config{
		Host:                   "0.0.0.0",
		Port:                   "8096",
		Logging:                defautLog,
		Paths:                  defaultPaths(),
		SessionSecret:          encryption.NewKey(32),
		ScanInterval:           task.NewJob(ScanInterval),
		MaxUploadedFileSizeMiB: 10,
		DebugMode:              false,
	}
}

// LoadConfig attempts to read Config from a filepath
// and returns the default Config if unsuccessful
func LoadConfig(fp string) *Config {
	c, err := readConfig(fp)
	if err != nil {
		log.Error().Err(err).Msg("failed to load config file, using defaults instead")
		c = DefaultConfig()
	}

	// If in debug mode then set the log level to at least debug
	if c.DebugMode && c.Logging.Level > log.DebugLevel {
		c.Logging.Level = log.DebugLevel
	}

	// Ensures the file/dir paths which tanuki uses exist
	err = c.Paths.EnsureExist(c.Logging.LogToFile)
	if err != nil {
		log.Panic().Err(err).Msg("paths can't be created")
	}

	// Ensure Job intervals can't be nil
	if c.ScanInterval == nil {
		c.ScanInterval = task.NewJob(ScanInterval)
	}

	return c
}

// Attempts to read Config and returns error if unsuccessful
func readConfig(fp string) (*Config, error) {
	buf, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Save(fp string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return fse.EnsureWriteFile(fp, data, 0666)
}
