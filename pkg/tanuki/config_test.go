package tanuki

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Remove("config.yml")
	os.RemoveAll("./config")
	os.Remove("tanuki.log")
	os.Exit(code)
}

func TestReadConfig(t *testing.T) {
	// Error should be called if no config file is found
	_, err := readConfig()
	if err == nil {
		t.Error(err)
	}
}

func TestSaveConfig(t *testing.T) {
	if saveConfig(loadConfig()) != nil {
		t.Error("failed to save default config")
	}
}

