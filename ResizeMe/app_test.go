package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type recordingAgent struct {
	applied   []Config
	applyErrs []error
}

func (a *recordingAgent) Start() error {
	return nil
}

func (a *recordingAgent) Stop() {}

func (a *recordingAgent) ApplySettings(config Config) error {
	a.applied = append(a.applied, config.Clone())
	if len(a.applyErrs) == 0 {
		return nil
	}
	err := a.applyErrs[0]
	a.applyErrs = a.applyErrs[1:]
	return err
}

func (a *recordingAgent) ResizeActiveWindow(Preset, bool) error {
	return nil
}

func (a *recordingAgent) Notify(string, string, bool) {}

func TestSaveSettingsIfUnchangedRejectsStaleDraft(t *testing.T) {
	expected := DefaultConfig()
	current := expected.Clone()
	current.ActivePresetID = "720p-landscape"
	next := expected.Clone()
	next.CenterAfterResize = false
	agent := &recordingAgent{}
	app := &App{
		config: current,
		store:  &ConfigStore{path: filepath.Join(t.TempDir(), settingsFile)},
		agent:  agent,
	}

	saved, err := app.SaveSettingsIfUnchanged(next, expected)

	if !errors.Is(err, errSettingsChanged) {
		t.Fatalf("expected stale settings error, got %v", err)
	}
	if !samePersistedConfig(saved, current) {
		t.Fatalf("returned config = %#v, want %#v", saved, current)
	}
	if !samePersistedConfig(app.config, current) {
		t.Fatalf("app config = %#v, want %#v", app.config, current)
	}
	if len(agent.applied) != 0 {
		t.Fatalf("expected no side effects for a stale draft, got %#v", agent.applied)
	}
}

func TestSaveSettingsRestoresRuntimeAfterApplyFailure(t *testing.T) {
	current := DefaultConfig()
	next := current.Clone()
	next.Hotkey = "Ctrl+Shift+R"
	agent := &recordingAgent{applyErrs: []error{errors.New("register hotkey failed")}}
	app := &App{
		config: current,
		store:  &ConfigStore{path: filepath.Join(t.TempDir(), settingsFile)},
		agent:  agent,
	}

	_, err := app.SaveSettings(next)

	if err == nil || !strings.Contains(err.Error(), "register hotkey failed") {
		t.Fatalf("expected apply failure, got %v", err)
	}
	if len(agent.applied) != 2 {
		t.Fatalf("expected failed settings and rollback, got %#v", agent.applied)
	}
	if !samePersistedConfig(agent.applied[0], next) {
		t.Fatalf("first apply = %#v, want %#v", agent.applied[0], next)
	}
	if !samePersistedConfig(agent.applied[1], current) {
		t.Fatalf("rollback = %#v, want %#v", agent.applied[1], current)
	}
	if !samePersistedConfig(app.config, current) {
		t.Fatalf("app config = %#v, want %#v", app.config, current)
	}
}
func TestSaveSettingsRestoresRuntimeAfterPersistenceFailure(t *testing.T) {
	current := DefaultConfig()
	next := current.Clone()
	next.CenterAfterResize = false
	blockingPath := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blockingPath, []byte("file"), 0o600); err != nil {
		t.Fatalf("create blocking file: %v", err)
	}
	agent := &recordingAgent{}
	app := &App{
		config: current,
		store:  &ConfigStore{path: filepath.Join(blockingPath, settingsFile)},
		agent:  agent,
	}

	_, err := app.SaveSettings(next)

	if err == nil || !strings.Contains(err.Error(), "create settings directory") {
		t.Fatalf("expected persistence failure, got %v", err)
	}
	if len(agent.applied) != 2 {
		t.Fatalf("expected persisted-settings rollback, got %#v", agent.applied)
	}
	if !samePersistedConfig(agent.applied[0], next) {
		t.Fatalf("first apply = %#v, want %#v", agent.applied[0], next)
	}
	if !samePersistedConfig(agent.applied[1], current) {
		t.Fatalf("rollback = %#v, want %#v", agent.applied[1], current)
	}
	if !samePersistedConfig(app.config, current) {
		t.Fatalf("app config = %#v, want %#v", app.config, current)
	}
}
