package main

type PlatformAgent interface {
	Start() error
	Stop()
	ApplySettings(Config) error
	ResizeActiveWindow(Preset, bool) error
	Notify(title string, message string, warning bool)
}
