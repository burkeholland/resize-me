package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeConfigFavoritePresetIDsAreNormalized(t *testing.T) {
	config := Config{
		Presets: []Preset{
			{ID: "a", Name: "A", Width: 800, Height: 600},
			{ID: "b", Name: "B", Width: 900, Height: 700},
		},
		ActivePresetID:    "a",
		FavoritePresetIDs: []string{"b", "missing", "b", "a"},
		CenterAfterResize: true,
		Hotkey:            defaultHotkey,
	}

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}

	if len(normalized.FavoritePresetIDs) != 2 {
		t.Fatalf("expected 2 favorites, got %d", len(normalized.FavoritePresetIDs))
	}
	if normalized.FavoritePresetIDs[0] != "b" || normalized.FavoritePresetIDs[1] != "a" {
		t.Fatalf("unexpected favorites order: %#v", normalized.FavoritePresetIDs)
	}
}

func TestNormalizeConfigNilFavoritesBecomeEmpty(t *testing.T) {
	config := Config{
		Presets:        []Preset{{ID: "a", Name: "A", Width: 800, Height: 600}},
		ActivePresetID: "a",
		Hotkey:         defaultHotkey,
	}

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}

	if normalized.FavoritePresetIDs == nil {
		t.Fatalf("expected non-nil favoritePresetIds slice")
	}
	if len(normalized.FavoritePresetIDs) != 0 {
		t.Fatalf("expected 0 favorites, got %d", len(normalized.FavoritePresetIDs))
	}
}

func TestNormalizeConfigUsesPortableFunctionKeyLimit(t *testing.T) {
	config := DefaultConfig()
	config.Hotkey = "Ctrl+Alt+F20"

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}
	if normalized.Hotkey != "Ctrl+Alt+F20" {
		t.Fatalf("expected F20 hotkey to remain valid, got %q", normalized.Hotkey)
	}

	config.Hotkey = "Ctrl+Alt+F21"
	normalized, err = NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}
	if normalized.Hotkey != defaultHotkey {
		t.Fatalf("expected F21 hotkey to fall back to %q, got %q", defaultHotkey, normalized.Hotkey)
	}
	if got := hotkeyValidationMessage(config.Hotkey); got != portableHotkeyHelp {
		t.Fatalf("expected portable hotkey guidance %q, got %q", portableHotkeyHelp, got)
	}
}

func TestConfigStoreMigratesLegacyFunctionKeysWithoutDiscardingSettings(t *testing.T) {
	config := DefaultConfig()
	config.Hotkey = "Ctrl+Alt+F24"
	config.Presets = []Preset{{ID: "custom", Name: "Custom", Width: 800, Height: 600}}
	config.ActivePresetID = "custom"

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	path := filepath.Join(t.TempDir(), settingsFile)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	loaded, err := (&ConfigStore{path: path}).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.Hotkey != defaultHotkey {
		t.Fatalf("expected legacy hotkey to fall back to %q, got %q", defaultHotkey, loaded.Hotkey)
	}
	if loaded.LoadError != portableHotkeyHelp {
		t.Fatalf("expected migration guidance %q, got %q", portableHotkeyHelp, loaded.LoadError)
	}
	if len(loaded.Presets) != 1 || loaded.Presets[0].ID != "custom" {
		t.Fatalf("expected existing presets to be preserved, got %#v", loaded.Presets)
	}
}

func TestSaveSettingsRejectsLegacyFunctionKeys(t *testing.T) {
	app := &App{config: DefaultConfig()}
	next := DefaultConfig()
	next.Hotkey = "Ctrl+Alt+F21"

	_, err := app.SaveSettings(next)
	if err == nil || err.Error() != portableHotkeyHelp {
		t.Fatalf("expected portable hotkey guidance error %q, got %v", portableHotkeyHelp, err)
	}
}
