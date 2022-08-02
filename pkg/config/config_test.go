package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	data, err := yaml.Marshal(c)
	require.Nil(t, err)

	expected := fmt.Sprintf(`
host: 0.0.0.0
port: "8096"
logging:
    level: info
    log_to_file: true
    log_to_console: true
paths:
    db: ./data/tanuki.db
    log: ./data/tanuki.log
    library: ./library
session_secret: %s
scan_interval_minutes: 180
max_uploaded_file_size_mib: 10
debug_mode: false`, c.SessionSecret.Base64())
	require.Equal(t, strings.Trim(expected, "\n"), strings.Trim(string(data), "\n"))
}
