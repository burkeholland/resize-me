package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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

func TestNormalizeConfigHiddenPresetIDsAreNormalizedAndKeepAnActiveVisiblePreset(t *testing.T) {
	config := Config{
		Presets: []Preset{
			{ID: "a", Name: "A", Width: 800, Height: 600},
			{ID: "b", Name: "B", Width: 900, Height: 700},
		},
		ActivePresetID:    "a",
		FavoritePresetIDs: []string{"a"},
		HiddenPresetIDs:   []string{"a", "missing", "a"},
		CenterAfterResize: true,
		Hotkey:            defaultHotkey,
	}

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}

	if got, want := normalized.HiddenPresetIDs, []string{"a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("hiddenPresetIds = %#v, want %#v", got, want)
	}
	if normalized.ActivePresetID != "b" {
		t.Fatalf("activePresetId = %q, want visible preset %q", normalized.ActivePresetID, "b")
	}
	if !reflect.DeepEqual(normalized.FavoritePresetIDs, []string{"a"}) {
		t.Fatalf("favoritePresetIds = %#v, want hidden favorite to be preserved", normalized.FavoritePresetIDs)
	}
}

func TestNormalizeConfigDoesNotAllowAllPresetsToBeHidden(t *testing.T) {
	config := Config{
		Presets: []Preset{
			{ID: "a", Name: "A", Width: 800, Height: 600},
			{ID: "b", Name: "B", Width: 900, Height: 700},
		},
		ActivePresetID:    "b",
		HiddenPresetIDs:   []string{"a", "b"},
		CenterAfterResize: true,
		Hotkey:            defaultHotkey,
	}

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}

	if got, want := normalized.HiddenPresetIDs, []string{"a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("hiddenPresetIds = %#v, want %#v", got, want)
	}
	if normalized.ActivePresetID != "b" {
		t.Fatalf("activePresetId = %q, want %q", normalized.ActivePresetID, "b")
	}
}

func TestConfigStoreMigratesMissingHiddenPresetIDsAsVisible(t *testing.T) {
	path := filepath.Join(t.TempDir(), settingsFile)
	data := []byte(`{"presets":[{"id":"a","name":"A","width":800,"height":600}],"activePresetId":"a","favoritePresetIds":[],"centerAfterResize":true,"hotkey":"Ctrl+Alt+R","autoStart":false,"firstRun":false}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	loaded, err := (&ConfigStore{path: path}).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.HiddenPresetIDs == nil {
		t.Fatal("expected non-nil hiddenPresetIds slice")
	}
	if len(loaded.HiddenPresetIDs) != 0 || len(loaded.VisiblePresets()) != 1 {
		t.Fatalf("legacy settings should keep every preset visible, got hidden %#v", loaded.HiddenPresetIDs)
	}
}

func TestConfigStorePersistsHiddenPresetIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), settingsFile)
	config := Config{
		Presets: []Preset{
			{ID: "a", Name: "A", Width: 800, Height: 600},
			{ID: "b", Name: "B", Width: 900, Height: 700},
		},
		ActivePresetID:    "b",
		FavoritePresetIDs: []string{"a"},
		HiddenPresetIDs:   []string{"a"},
		CenterAfterResize: true,
		Hotkey:            defaultHotkey,
	}

	store := &ConfigStore{path: path}
	if err := store.Save(config); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !reflect.DeepEqual(loaded.HiddenPresetIDs, []string{"a"}) {
		t.Fatalf("hiddenPresetIds = %#v, want %#v", loaded.HiddenPresetIDs, []string{"a"})
	}
	if !reflect.DeepEqual(loaded.FavoritePresetIDs, []string{"a"}) {
		t.Fatalf("favoritePresetIds = %#v, want %#v", loaded.FavoritePresetIDs, []string{"a"})
	}
	if got, want := loaded.VisiblePresets(), []Preset{{ID: "b", Name: "B", Width: 900, Height: 700}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("visible presets = %#v, want %#v", got, want)
	}
}

func TestNormalizeConfigUsesWindowsFunctionKeyLimit(t *testing.T) {
	if maximumSupportedFunctionKey() != 24 {
		t.Skip("F21-F24 are only supported on Windows")
	}

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
	if normalized.Hotkey != "Ctrl+Alt+F21" {
		t.Fatalf("expected F21 hotkey to remain valid on Windows, got %q", normalized.Hotkey)
	}
	if got := hotkeyValidationMessage(config.Hotkey); got != "" {
		t.Fatalf("expected no Windows hotkey guidance, got %q", got)
	}
}

func TestNormalizeConfigDefaultsQuickPickHotkey(t *testing.T) {
	config := DefaultConfig()
	config.QuickPickHotkey = ""

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}
	if normalized.QuickPickHotkey != defaultQuickPickHotkey {
		t.Fatalf("quickPickHotkey = %q, want %q", normalized.QuickPickHotkey, defaultQuickPickHotkey)
	}
}

func TestNormalizeConfigKeepsLegacyResizeHotkeyWhenItUsesTheNewDefault(t *testing.T) {
	config := DefaultConfig()
	config.Hotkey = defaultQuickPickHotkey
	config.QuickPickHotkey = ""

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}
	if normalized.Hotkey != defaultQuickPickHotkey {
		t.Fatalf("hotkey = %q, want %q", normalized.Hotkey, defaultQuickPickHotkey)
	}
	if normalized.QuickPickHotkey != legacyQuickPickHotkeyFallback {
		t.Fatalf("quickPickHotkey = %q, want %q", normalized.QuickPickHotkey, legacyQuickPickHotkeyFallback)
	}
}

func TestNormalizeConfigNormalizesCustomQuickPickHotkey(t *testing.T) {
	config := DefaultConfig()
	config.QuickPickHotkey = "alt shift q"

	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig returned error: %v", err)
	}
	if normalized.QuickPickHotkey != "Alt+Shift+Q" {
		t.Fatalf("quickPickHotkey = %q, want %q", normalized.QuickPickHotkey, "Alt+Shift+Q")
	}
}

func TestNormalizeConfigRejectsDuplicateHotkeys(t *testing.T) {
	config := DefaultConfig()
	config.QuickPickHotkey = config.Hotkey

	_, err := NormalizeConfig(config, DefaultConfig())
	if err == nil || err.Error() != "quick-pick hotkey must differ from the resize hotkey" {
		t.Fatalf("expected duplicate hotkey error, got %v", err)
	}
}

func TestConfigStorePreservesWindowsFunctionKeysWithoutDiscardingSettings(t *testing.T) {
	if maximumSupportedFunctionKey() != 24 {
		t.Skip("F21-F24 are only supported on Windows")
	}

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
	if loaded.Hotkey != "Ctrl+Alt+F24" {
		t.Fatalf("expected persisted Windows hotkey to be preserved, got %q", loaded.Hotkey)
	}
	if loaded.LoadError != "" {
		t.Fatalf("expected no load error for a supported Windows hotkey, got %q", loaded.LoadError)
	}
	if len(loaded.Presets) != 1 || loaded.Presets[0].ID != "custom" {
		t.Fatalf("expected existing presets to be preserved, got %#v", loaded.Presets)
	}
}

func TestSaveSettingsAcceptsWindowsFunctionKeys(t *testing.T) {
	if maximumSupportedFunctionKey() != 24 {
		t.Skip("F21-F24 are only supported on Windows")
	}

	app := &App{
		config: DefaultConfig(),
		store:  &ConfigStore{path: filepath.Join(t.TempDir(), settingsFile)},
	}
	next := DefaultConfig()
	next.Hotkey = "Ctrl+Alt+F21"

	saved, err := app.SaveSettings(next)
	if err != nil {
		t.Fatalf("expected F21 to be accepted on Windows, got %v", err)
	}
	if saved.Hotkey != next.Hotkey {
		t.Fatalf("saved hotkey = %q, want %q", saved.Hotkey, next.Hotkey)
	}
}
