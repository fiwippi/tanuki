package config

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"github.com/fiwippi/tanuki/internal/encryption"
	"github.com/fiwippi/tanuki/internal/fse"
)

const ScanInterval = 3 * 60 // Every 3 hours

type Config struct {
	Host                   string          `yaml:"host"`
	Port                   string          `yaml:"port"`
	DBPath                 string          `yaml:"db_path"`
	LibraryPath            string          `yaml:"library_Path"`
	SessionSecret          *encryption.Key `yaml:"session_secret"`
	ScanInterval           int             `yaml:"scan_interval_minutes"`
	MaxUploadedFileSizeMiB int             `yaml:"max_uploaded_file_size_mib"`
	DebugMode              bool            `yaml:"debug_mode"`
}

func DefaultConfig() *Config {
	return &Config{
		Host:                   "0.0.0.0",
		Port:                   "8096",
		DBPath:                 "./data/tanuki.db",
		LibraryPath:            "./library",
		SessionSecret:          encryption.NewKey(32),
		ScanInterval:           ScanInterval,
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

	// Ensures the file/dir paths which tanuki uses exist
	if err := fse.CreateDirs(filepath.Dir(c.DBPath), c.LibraryPath); err != nil {
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

func (c *Config) Save(fp string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(fp)
	err = fse.CreateDirs(parentDir)
	if err != nil {
		return err
	}
	return os.WriteFile(fp, data, 0666)
}
