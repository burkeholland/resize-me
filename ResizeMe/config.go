package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultHotkey = "Ctrl+Alt+R"
	appDirName    = "ResizeMe"
	settingsFile  = "settings.json"
)

type Preset struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Config struct {
	Presets           []Preset `json:"presets"`
	ActivePresetID    string   `json:"activePresetId"`
	FavoritePresetIDs []string `json:"favoritePresetIds"`
	CenterAfterResize bool     `json:"centerAfterResize"`
	Hotkey            string   `json:"hotkey"`
	AutoStart         bool     `json:"autoStart"`
	FirstRun          bool     `json:"firstRun"`
	LoadError         string   `json:"loadError,omitempty"`
}

type ConfigStore struct {
	path string
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{path: defaultConfigPath()}
}

func (c Config) Clone() Config {
	clone := c
	clone.Presets = append([]Preset(nil), c.Presets...)
	clone.FavoritePresetIDs = append([]string(nil), c.FavoritePresetIDs...)
	return clone
}

func (c Config) HasPreset(id string) bool {
	_, ok := c.FindPreset(id)
	return ok
}

func (c Config) FindPreset(id string) (Preset, bool) {
	for _, preset := range c.Presets {
		if preset.ID == id {
			return preset, true
		}
	}
	return Preset{}, false
}

func (c Config) ActivePreset() (Preset, bool) {
	return c.FindPreset(c.ActivePresetID)
}

func (s *ConfigStore) Load() (Config, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, fmt.Errorf("read settings: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse settings: %w", err)
	}

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		return Config{}, fmt.Errorf("validate settings: %w", err)
	}
	return normalized, nil
}

func (s *ConfigStore) Save(config Config) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create settings directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace settings: %w", err)
	}
	return nil
}

func DefaultConfig() Config {
	presets := []Preset{
		{ID: "360p-landscape", Name: "360p Landscape", Width: 640, Height: 360},
		{ID: "480p-landscape", Name: "480p Landscape", Width: 854, Height: 480},
		{ID: "540p-landscape", Name: "540p Landscape", Width: 960, Height: 540},
		{ID: "720p-landscape", Name: "720p Landscape", Width: 1280, Height: 720},
		{ID: "900p-landscape", Name: "900p Landscape", Width: 1600, Height: 900},
		{ID: "1080p-landscape", Name: "1080p Landscape", Width: 1920, Height: 1080},
		{ID: "1440p-landscape", Name: "1440p Landscape", Width: 2560, Height: 1440},
		{ID: "1800p-landscape", Name: "1800p Landscape", Width: 3200, Height: 1800},
		{ID: "4k-landscape", Name: "4K Landscape", Width: 3840, Height: 2160},
		{ID: "360p-portrait", Name: "360p Portrait", Width: 360, Height: 640},
		{ID: "480p-portrait", Name: "480p Portrait", Width: 480, Height: 854},
		{ID: "540p-portrait", Name: "540p Portrait", Width: 540, Height: 960},
		{ID: "720p-portrait", Name: "720p Portrait", Width: 720, Height: 1280},
		{ID: "900p-portrait", Name: "900p Portrait", Width: 900, Height: 1600},
		{ID: "1080p-portrait", Name: "1080p Portrait", Width: 1080, Height: 1920},
		{ID: "1440p-portrait", Name: "1440p Portrait", Width: 1440, Height: 2560},
		{ID: "1800p-portrait", Name: "1800p Portrait", Width: 1800, Height: 3200},
		{ID: "4k-portrait", Name: "4K Portrait", Width: 2160, Height: 3840},
	}

	return Config{
		Presets:           presets,
		ActivePresetID:    "1080p-landscape",
		FavoritePresetIDs: []string{},
		CenterAfterResize: true,
		Hotkey:            defaultHotkey,
		AutoStart:         false,
		FirstRun:          true,
	}
}

func defaultConfigPath() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, appDirName, settingsFile)
	}
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		return filepath.Join(".", settingsFile)
	}
	return filepath.Join(configDir, appDirName, settingsFile)
}
