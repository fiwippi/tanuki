package config

import (
	"github.com/fiwippi/tanuki/internal/encryption"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/pkg/logging"
	"github.com/fiwippi/tanuki/pkg/task"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"os"
)

//
type Config struct {
	Host                    string          `yaml:"host"`
	Port                    string          `yaml:"port"`
	Logging                 *logging.Config `yaml:"logging"`
	Paths                   *Paths          `yaml:"paths"`
	SessionSecret           *encryption.Key `yaml:"session_secret"`
	ScanInterval            *task.Minutes   `yaml:"scan_interval_minutes"`
	ThumbGenerationInterval *task.Minutes   `yaml:"thumbnail_generation_interval_minutes"`
	MaxUploadedFileSizeMiB  int             `yaml:"max_uploaded_file_size_mib"`
	DebugMode               bool            `yaml:"debug_mode"`
}

//
func DefaultConfig() *Config {
	return &Config{
		Host:                    "0.0.0.0",
		Port:                    "8096",
		Logging:                 logging.DefaultConfig(),
		Paths:                   defaultPaths(),
		SessionSecret:           encryption.NewKey(32),
		ScanInterval:            task.NewMinutes(5),
		ThumbGenerationInterval: task.NewMinutes(60),
		MaxUploadedFileSizeMiB:  10,
		DebugMode:               false,
	}
}

// Attempts to read Config but returns default Config if unsuccessful
func LoadConfig(fp string) *Config {
	c, err := readConfig(fp)
	if err != nil {
		log.Error().Err(err).Msg("failed to load config file, using defaults instead")
		c = DefaultConfig()
	}

	// If in debug mode then set the log level to at least debug
	if c.DebugMode && c.Logging.Level > logging.DebugLevel {
		c.Logging.Level = logging.DebugLevel
	}

	// Ensures the file/dir paths which tanuki uses exist
	err = c.Paths.EnsureExist()
	if err != nil {
		log.Panic().Err(err).Msg("paths can't be created")
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

//
func (c *Config) Save(fp string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return fse.EnsureWriteFile(fp, data, 0666)
}
