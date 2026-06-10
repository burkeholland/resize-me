//go:build !windows

package main

import "fmt"

type stubAgent struct{}

func NewPlatformAgent(app *App) PlatformAgent {
	return &stubAgent{}
}

func (s *stubAgent) Start() error {
	return nil
}

func (s *stubAgent) Stop() {}

func (s *stubAgent) ApplySettings(Config) error {
	return nil
}

func (s *stubAgent) ResizeActiveWindow(Preset, bool) error {
	return fmt.Errorf("window resizing is only available on Windows")
}

func (s *stubAgent) Notify(string, string, bool) {}
