package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	WorkDuration          int `json:"work_duration"`
	ShortBreakDuration    int `json:"short_break_duration"`
	LongBreakDuration     int `json:"long_break_duration"`
	CyclesBeforeLongBreak int `json:"cycles_before_long_break"`
}

func DefaultConfig() *Config {
	return &Config{
		WorkDuration:          25,
		ShortBreakDuration:    5,
		LongBreakDuration:     15,
		CyclesBeforeLongBreak: 4,
	}
}

func (c *Config) WorkDurationTime() time.Duration {
	return time.Duration(c.WorkDuration) * time.Minute
}

func (c *Config) ShortBreakDurationTime() time.Duration {
	return time.Duration(c.ShortBreakDuration) * time.Minute
}

func (c *Config) LongBreakDurationTime() time.Duration {
	return time.Duration(c.LongBreakDuration) * time.Minute
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "tuime")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
