package tanuki

import (
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/pkg/auth"
)

//
type config struct {
	Host                   string          `yaml:"host"`
	Port                   string          `yaml:"port"`
	Logging                logConfig       `yaml:"logging"`
	Paths                  paths           `yaml:"paths"`
	SessionSecret          *auth.SecureKey `yaml:"session_secret"`
	ScanIntervalMinutes    *ScanInterval   `yaml:"scan_interval_minutes"`
	MaxUploadedFileSizeMiB int             `yaml:"max_uploaded_file_size_mib"`
	DebugMode              bool            `yaml:"debug_mode"`
}

//
func defaultConfig() *config {
	return &config{
		Host:                   "0.0.0.0",
		Port:                   "8096",
		Logging:                defaultLogConfig(),
		Paths:                  defaultPaths(),
		SessionSecret:          auth.GenerateSecureKey(32),
		ScanIntervalMinutes:    NewInterval(5),
		MaxUploadedFileSizeMiB: 10,
		DebugMode: false,
	}
}

// Attempts to read config but returns default config if unsuccessful
func loadConfig() *config {
	c, err := readConfig()
	if err != nil {
		log.Error().Err(err).Msg("failed to load config file, using defaults instead")
		return defaultConfig()
	}
	return c
}

// Attempts to read config and returns error if unsuccessful
func readConfig() (*config, error) {
	buf, err := ioutil.ReadFile("./config/config.yml")
	if err != nil {
		return nil, err
	}

	c := &config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Saves config to file
func saveConfig(c *config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return fse.EnsureWriteFile("./config/config.yml", data, 0666)
}