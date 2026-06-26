package main

import "testing"

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
