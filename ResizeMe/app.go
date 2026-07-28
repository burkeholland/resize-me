package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	store     *ConfigStore
	agent     PlatformAgent
	loadError string

	mu     sync.RWMutex
	saveMu sync.Mutex // serialises the full save transaction
	config Config
}

var errSettingsChanged = errors.New("Settings changed elsewhere. Review the latest values before saving.")

func NewApp() *App {
	store := NewConfigStore()
	config, err := store.Load()
	loadError := ""
	if err != nil {
		config = DefaultConfig()
		loadError = err.Error()
	} else if config.LoadError != "" {
		loadError = config.LoadError
	}

	return &App{
		store:     store,
		config:    config,
		loadError: loadError,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.agent = NewPlatformAgent(a)
	if err := a.agent.Start(); err != nil {
		log.Printf("[ResizeMe] agent start failed: %v", err)
		a.mu.Lock()
		a.loadError = err.Error()
		a.mu.Unlock()
	}

	a.mu.RLock()
	firstRun := a.config.FirstRun
	a.mu.RUnlock()
	if firstRun {
		wailsruntime.WindowShow(ctx)
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.agent != nil {
		a.agent.Stop()
	}
}

func (a *App) GetSettings() Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	config := a.config.Clone()
	config.LoadError = a.loadError
	return config
}

func (a *App) SaveSettings(next Config) (Config, error) {
	return a.saveSettings(next, nil)
}

// SaveSettingsIfUnchanged applies a settings draft only when it is based on
// the current persisted configuration.
func (a *App) SaveSettingsIfUnchanged(next Config, expected Config) (Config, error) {
	return a.saveSettings(next, &expected)
}

func (a *App) saveSettings(next Config, expected *Config) (Config, error) {
	a.saveMu.Lock()
	defer a.saveMu.Unlock()

	a.mu.RLock()
	current := a.config.Clone()
	a.mu.RUnlock()

	if expected != nil && !samePersistedConfig(current, *expected) {
		return current, errSettingsChanged
	}

	if message := hotkeyValidationMessage(next.Hotkey); message != "" {
		return current, fmt.Errorf("%s", message)
	}

	normalized, err := NormalizeConfig(next, current)
	if err != nil {
		return current, err
	}

	if a.agent != nil {
		if err := a.agent.ApplySettings(normalized); err != nil {
			return a.restoreRuntimeSettings(current, err)
		}
	}

	if err := a.store.Save(normalized); err != nil {
		return a.restoreRuntimeSettings(current, err)
	}

	a.mu.Lock()
	a.config = normalized.Clone()
	a.loadError = ""
	a.mu.Unlock()
	a.emitSettingsUpdated()
	return normalized, nil
}

func (a *App) restoreRuntimeSettings(current Config, cause error) (Config, error) {
	if a.agent == nil {
		return current, cause
	}
	if err := a.agent.ApplySettings(current); err != nil {
		return current, fmt.Errorf("%v; failed to restore previous runtime settings: %w", cause, err)
	}
	return current, cause
}

func samePersistedConfig(left Config, right Config) bool {
	left.LoadError = ""
	right.LoadError = ""
	return reflect.DeepEqual(left, right)
}

func (a *App) SetActivePreset(id string) (Config, error) {
	a.mu.RLock()
	next := a.config.Clone()
	a.mu.RUnlock()

	if !next.HasPreset(id) {
		return next, fmt.Errorf("unknown preset %q", id)
	}
	next.ActivePresetID = id
	return a.SaveSettings(next)
}

func (a *App) SetCenterAfterResize(enabled bool) (Config, error) {
	a.mu.RLock()
	next := a.config.Clone()
	a.mu.RUnlock()

	next.CenterAfterResize = enabled
	return a.SaveSettings(next)
}

func (a *App) SetAutoStart(enabled bool) (Config, error) {
	a.mu.RLock()
	next := a.config.Clone()
	a.mu.RUnlock()

	next.AutoStart = enabled
	next.FirstRun = false
	return a.SaveSettings(next)
}

func (a *App) CompleteFirstRun(enableAutoStart bool) (Config, error) {
	a.mu.RLock()
	next := a.config.Clone()
	a.mu.RUnlock()

	next.AutoStart = enableAutoStart
	next.FirstRun = false
	return a.SaveSettings(next)
}

func (a *App) ResizeNow() error {
	a.mu.RLock()
	config := a.config.Clone()
	a.mu.RUnlock()

	preset, ok := config.ActivePreset()
	if !ok {
		return fmt.Errorf("no active resize preset is configured")
	}
	if a.agent == nil {
		return fmt.Errorf("window agent is not running")
	}
	return a.agent.ResizeActiveWindow(preset, config.CenterAfterResize)
}

func (a *App) ShowSettings() {
	if a.ctx == nil {
		return
	}
	wailsruntime.WindowShow(a.ctx)
	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.WindowSetAlwaysOnTop(a.ctx, true)
	wailsruntime.WindowSetAlwaysOnTop(a.ctx, false)
}

func (a *App) ShowAbout() {
	a.ShowSettings()
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "show-about")
	}
}

func (a *App) GetVersion() string {
	return applicationVersion()
}

func (a *App) Quit() {
	if a.agent != nil {
		a.agent.Stop()
	}
	if a.ctx != nil {
		wailsruntime.Quit(a.ctx)
	}
}

func (a *App) emitSettingsUpdated() {
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "settings-updated", a.GetSettings())
	}
}
