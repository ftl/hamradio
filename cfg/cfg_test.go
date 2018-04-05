package cfg

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirectory(t *testing.T) {
	customAbsolute, err := Directory("/var/custom/config")
	if err != nil {
		t.Errorf("failed to resolve custom absolute directory: %v", err)
	}
	if customAbsolute != "/var/custom/config" {
		t.Errorf("expected /var/custom/config, but got %q", customAbsolute)
	}
	customRelative, err := Directory("custom/config")
	if err != nil {
		t.Errorf("failed to resolve custom relative directory: %v", err)
	}
	if !strings.HasSuffix(customRelative, "/custom/config") {
		t.Errorf("expected .../custom/config, but got %q", customRelative)
	}
	defaultDir, err := Directory("")
	if err != nil {
		t.Errorf("failed to resolve default directory: %v", err)
	}
	if !(strings.HasPrefix(defaultDir, "/home/") && strings.HasSuffix(defaultDir, "/.config/hamradio")) {
		t.Errorf("expected /home/.../.config/hamradio, but got %q", defaultDir)
	}
}

func TestPrepareDirectory(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "cfg.TestPrepareDirectory")
	if err != nil {
		t.Errorf("failed to create temp dir: %v", err)
		t.FailNow()
	}
	defer os.Remove(tempDir)
	testDir := filepath.Join(tempDir, "custom/config/dir")

	preparedDir, err := PrepareDirectory(testDir)

	if preparedDir != testDir {
		t.Errorf("expected %q, but got %q", testDir, preparedDir)
	}
	if _, err := os.Stat(preparedDir); os.IsNotExist(err) {
		t.Errorf("failed to prepare %v", preparedDir)
	}
}

func TestRead(t *testing.T) {
	testInput := `{"test_key": "test_value"}`
	in := strings.NewReader(testInput)
	config, err := Read(in)
	if err != nil {
		t.Errorf("failed to read test input: %v", err)
	}
	if config["test_key"] != "test_value" {
		t.Errorf("cannot retrieve test_key from config, got %v", config)
	}
}

func TestLoad(t *testing.T) {
	config, err := Load("./testdata", "")
	if err != nil {
		t.Errorf("failed to load test config file: %v", err)
		t.FailNow()
	}
	if config["test_key"] != "test_value" {
		t.Errorf("cannto retrieve test_key from test config, got %v", config)
	}
}
