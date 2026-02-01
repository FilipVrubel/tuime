package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.WorkDuration != 25 {
		t.Errorf("Expected WorkDuration 25, got %d", cfg.WorkDuration)
	}

	if cfg.ShortBreakDuration != 5 {
		t.Errorf("Expected ShortBreakDuration 5, got %d", cfg.ShortBreakDuration)
	}

	if cfg.LongBreakDuration != 15 {
		t.Errorf("Expected LongBreakDuration 15, got %d", cfg.LongBreakDuration)
	}

	if cfg.CyclesBeforeLongBreak != 4 {
		t.Errorf("Expected CyclesBeforeLongBreak 4, got %d", cfg.CyclesBeforeLongBreak)
	}
}

func TestConfigDurationConversion(t *testing.T) {
	cfg := &Config{
		WorkDuration:       30,
		ShortBreakDuration: 10,
		LongBreakDuration:  20,
	}

	tests := []struct {
		name     string
		got      time.Duration
		expected time.Duration
	}{
		{
			name:     "work duration",
			got:      cfg.WorkDurationTime(),
			expected: 30 * time.Minute,
		},
		{
			name:     "short break duration",
			got:      cfg.ShortBreakDurationTime(),
			expected: 10 * time.Minute,
		},
		{
			name:     "long break duration",
			got:      cfg.LongBreakDurationTime(),
			expected: 20 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.got)
			}
		})
	}
}

func TestLoadNonExistentConfig(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Errorf("Expected no error loading non-existent config, got %v", err)
	}

	if cfg.WorkDuration != DefaultConfig().WorkDuration {
		t.Error("Expected default config when file doesn't exist")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	customCfg := &Config{
		WorkDuration:          30,
		ShortBreakDuration:    10,
		LongBreakDuration:     20,
		CyclesBeforeLongBreak: 3,
	}

	err := customCfg.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.WorkDuration != customCfg.WorkDuration {
		t.Errorf("Expected WorkDuration %d, got %d", customCfg.WorkDuration, loadedCfg.WorkDuration)
	}

	if loadedCfg.ShortBreakDuration != customCfg.ShortBreakDuration {
		t.Errorf("Expected ShortBreakDuration %d, got %d", customCfg.ShortBreakDuration, loadedCfg.ShortBreakDuration)
	}

	if loadedCfg.LongBreakDuration != customCfg.LongBreakDuration {
		t.Errorf("Expected LongBreakDuration %d, got %d", customCfg.LongBreakDuration, loadedCfg.LongBreakDuration)
	}

	if loadedCfg.CyclesBeforeLongBreak != customCfg.CyclesBeforeLongBreak {
		t.Errorf("Expected CyclesBeforeLongBreak %d, got %d", customCfg.CyclesBeforeLongBreak, loadedCfg.CyclesBeforeLongBreak)
	}
}

func TestConfigPathCreation(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	configDir := filepath.Dir(path)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	expectedPath := filepath.Join(tempDir, ".config", "tuime", "config.json")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestLoadCorruptedConfig(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	corruptedData := []byte("{ this is not valid json }")
	err = os.WriteFile(path, corruptedData, 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted config: %v", err)
	}

	cfg, err := Load()
	if err == nil {
		t.Error("Expected error when loading corrupted config")
	}

	if cfg.WorkDuration != DefaultConfig().WorkDuration {
		t.Error("Expected default config when JSON is invalid")
	}
}

func TestMultipleSaves(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	cfg1 := &Config{
		WorkDuration:          25,
		ShortBreakDuration:    5,
		LongBreakDuration:     15,
		CyclesBeforeLongBreak: 4,
	}

	err := cfg1.Save()
	if err != nil {
		t.Fatalf("Failed to save first config: %v", err)
	}

	cfg2 := &Config{
		WorkDuration:          50,
		ShortBreakDuration:    10,
		LongBreakDuration:     30,
		CyclesBeforeLongBreak: 3,
	}

	err = cfg2.Save()
	if err != nil {
		t.Fatalf("Failed to save second config: %v", err)
	}

	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.WorkDuration != cfg2.WorkDuration {
		t.Errorf("Expected WorkDuration %d (from second save), got %d", cfg2.WorkDuration, loadedCfg.WorkDuration)
	}
}
