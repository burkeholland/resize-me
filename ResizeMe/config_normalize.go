package main

import (
	"fmt"
	"regexp"
	"strings"
)

var nonAlphanumericRe = regexp.MustCompile(`[^a-z0-9]+`)

// NormalizeConfig validates and normalizes a config, falling back to defaults
// from the provided fallback config where needed.
func NormalizeConfig(config Config, fallback Config) (Config, error) {
	next := config.Clone()
	next.LoadError = ""

	if strings.TrimSpace(next.Hotkey) == "" {
		next.Hotkey = defaultHotkey
	}
	next.Hotkey = normalizeHotkeyText(next.Hotkey)
	if !isValidHotkeyText(next.Hotkey) {
		next.Hotkey = defaultHotkey
	}

	if len(next.Presets) == 0 {
		next.Presets = append([]Preset(nil), fallback.Presets...)
	}

	seen := map[string]bool{}
	for i := range next.Presets {
		preset := &next.Presets[i]
		preset.Name = strings.TrimSpace(preset.Name)
		if preset.Name == "" {
			preset.Name = fmt.Sprintf("%dx%d", preset.Width, preset.Height)
		}
		if preset.Width < 100 || preset.Width > 10000 {
			return Config{}, fmt.Errorf("%s width must be between 100 and 10000", preset.Name)
		}
		if preset.Height < 100 || preset.Height > 10000 {
			return Config{}, fmt.Errorf("%s height must be between 100 and 10000", preset.Name)
		}
		if strings.TrimSpace(preset.ID) == "" {
			preset.ID = presetID(*preset, i)
		}
		baseID := preset.ID
		for suffix := 2; seen[preset.ID]; suffix++ {
			preset.ID = fmt.Sprintf("%s-%d", baseID, suffix)
		}
		seen[preset.ID] = true
	}

	if !next.HasPreset(next.ActivePresetID) {
		if fallback.ActivePresetID != "" && next.HasPreset(fallback.ActivePresetID) {
			next.ActivePresetID = fallback.ActivePresetID
		} else {
			next.ActivePresetID = next.Presets[0].ID
		}
	}

	seenFavoriteIDs := map[string]bool{}
	normalizedFavorites := make([]string, 0, len(next.FavoritePresetIDs))
	for _, id := range next.FavoritePresetIDs {
		if !next.HasPreset(id) || seenFavoriteIDs[id] {
			continue
		}
		seenFavoriteIDs[id] = true
		normalizedFavorites = append(normalizedFavorites, id)
	}
	next.FavoritePresetIDs = normalizedFavorites

	seenHiddenIDs := map[string]bool{}
	normalizedHidden := make([]string, 0, len(next.HiddenPresetIDs))
	for _, id := range next.HiddenPresetIDs {
		if !next.HasPreset(id) || seenHiddenIDs[id] {
			continue
		}
		seenHiddenIDs[id] = true
		normalizedHidden = append(normalizedHidden, id)
	}
	next.HiddenPresetIDs = normalizedHidden

	if len(next.VisiblePresets()) == 0 {
		for i, id := range next.HiddenPresetIDs {
			if id == next.ActivePresetID {
				next.HiddenPresetIDs = append(next.HiddenPresetIDs[:i], next.HiddenPresetIDs[i+1:]...)
				break
			}
		}
	}

	visiblePresets := next.VisiblePresets()
	if next.IsPresetHidden(next.ActivePresetID) {
		next.ActivePresetID = visiblePresets[0].ID
	}

	return next, nil
}

func presetID(preset Preset, index int) string {
	base := strings.ToLower(strings.TrimSpace(preset.Name))
	base = nonAlphanumericRe.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = fmt.Sprintf("preset-%d", index+1)
	}
	return base
}

func normalizeHotkeyText(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '+' || r == ' ' || r == '-'
	})
	modifiers := map[string]bool{}
	key := ""
	for _, part := range parts {
		normalized := strings.ToLower(strings.TrimSpace(part))
		switch normalized {
		case "", "plus":
			continue
		case "ctrl", "control":
			modifiers["Ctrl"] = true
		case "alt", "option":
			modifiers["Alt"] = true
		case "shift":
			modifiers["Shift"] = true
		case "win", "windows", "cmd", "meta":
			modifiers["Win"] = true
		default:
			key = strings.ToUpper(normalized)
		}
	}

	ordered := make([]string, 0, 5)
	for _, modifier := range []string{"Ctrl", "Alt", "Shift", "Win"} {
		if modifiers[modifier] {
			ordered = append(ordered, modifier)
		}
	}
	if key != "" {
		ordered = append(ordered, key)
	}
	return strings.Join(ordered, "+")
}

// isValidHotkeyText checks that a normalized hotkey string has at least one
// modifier and one Windows-supported key (A-Z, 0-9, or F1-F24).
func isValidHotkeyText(value string) bool {
	parts := strings.Split(value, "+")
	hasModifier := false
	hasKey := false
	for _, part := range parts {
		switch part {
		case "Ctrl", "Alt", "Shift", "Win":
			hasModifier = true
		default:
			if isValidHotkeyKey(part) {
				hasKey = true
			}
		}
	}
	return hasModifier && hasKey
}

func isValidHotkeyKey(value string) bool {
	if len(value) == 1 {
		ch := value[0]
		return (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
	}

	if !strings.HasPrefix(value, "F") {
		return false
	}
	n := 0
	for _, c := range strings.TrimPrefix(value, "F") {
		if c < '0' || c > '9' {
			return false
		}
		n = n*10 + int(c-'0')
	}
	return n >= 1 && n <= 24
}
