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

func TestNormalizeConfigSupportsWindowsFunctionKeys(t *testing.T) {
	for _, hotkey := range []string{"Ctrl+Alt+F20", "Ctrl+Alt+F21", "Ctrl+Alt+F24"} {
		config := DefaultConfig()
		config.Hotkey = hotkey

		normalized, err := NormalizeConfig(config, DefaultConfig())
		if err != nil {
			t.Fatalf("NormalizeConfig(%q) returned error: %v", hotkey, err)
		}
		if normalized.Hotkey != hotkey {
			t.Fatalf("NormalizeConfig(%q) hotkey = %q", hotkey, normalized.Hotkey)
		}
	}

	config := DefaultConfig()
	config.Hotkey = "Ctrl+Alt+F25"
	normalized, err := NormalizeConfig(config, DefaultConfig())
	if err != nil {
		t.Fatalf("NormalizeConfig(F25) returned error: %v", err)
	}
	if normalized.Hotkey != defaultHotkey {
		t.Fatalf("F25 hotkey = %q, want %q", normalized.Hotkey, defaultHotkey)
	}
}

func TestConfigStorePreservesWindowsFunctionKeys(t *testing.T) {
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
	if loaded.Hotkey != config.Hotkey {
		t.Fatalf("loaded hotkey = %q, want %q", loaded.Hotkey, config.Hotkey)
	}
	if loaded.LoadError != "" {
		t.Fatalf("unexpected load error: %q", loaded.LoadError)
	}
	if len(loaded.Presets) != 1 || loaded.Presets[0].ID != "custom" {
		t.Fatalf("expected existing presets to be preserved, got %#v", loaded.Presets)
	}

	if err := (&ConfigStore{path: path}).Save(loaded); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	reloaded, err := (&ConfigStore{path: path}).Load()
	if err != nil {
		t.Fatalf("reload returned error: %v", err)
	}
	if reloaded.Hotkey != config.Hotkey {
		t.Fatalf("reloaded hotkey = %q, want %q", reloaded.Hotkey, config.Hotkey)
	}
}

func TestSaveSettingsAllowsWindowsFunctionKeys(t *testing.T) {
	app := &App{
		config: DefaultConfig(),
		store:  &ConfigStore{path: filepath.Join(t.TempDir(), settingsFile)},
	}
	next := DefaultConfig()
	next.Hotkey = "Ctrl+Alt+F21"

	saved, err := app.SaveSettings(next)
	if err != nil {
		t.Fatalf("SaveSettings returned error: %v", err)
	}
	if saved.Hotkey != next.Hotkey {
		t.Fatalf("saved hotkey = %q, want %q", saved.Hotkey, next.Hotkey)
	}
}
