//go:build windows

package main

import "testing"

func TestClassifyResizeOutcome(t *testing.T) {
	preset := Preset{Width: 1280, Height: 720}

	tests := []struct {
		name    string
		achieved rect
		exact   bool
	}{
		{
			name:     "exact dimensions",
			achieved: rect{Left: 100, Top: 50, Right: 1380, Bottom: 770},
			exact:    true,
		},
		{
			name:     "within tolerance",
			achieved: rect{Right: 1281, Bottom: 719},
			exact:    true,
		},
		{
			name:     "constrained width",
			achieved: rect{Right: 1278, Bottom: 720},
			exact:    false,
		},
		{
			name:     "constrained height",
			achieved: rect{Right: 1280, Bottom: 722},
			exact:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outcome := classifyResizeOutcome(preset, test.achieved)
			if outcome.isExact() != test.exact {
				t.Fatalf("isExact() = %t, want %t for achieved %dx%d", outcome.isExact(), test.exact, outcome.achievedWidth, outcome.achievedHeight)
			}
		})
	}
}

func TestWindowStateRequiresRestore(t *testing.T) {
	tests := []struct {
		name  string
		state windowState
		want  bool
	}{
		{name: "normal window", state: windowState{}, want: false},
		{name: "minimized window", state: windowState{minimized: true}, want: true},
		{name: "maximized window", state: windowState{maximized: true}, want: true},
		{name: "minimized and maximized window", state: windowState{minimized: true, maximized: true}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.state.requiresRestore(); got != test.want {
				t.Fatalf("requiresRestore() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestParseHotkeySupportsF24(t *testing.T) {
	mods, key, err := parseHotkey("Ctrl+Alt+F24")
	if err != nil {
		t.Fatalf("parseHotkey returned error: %v", err)
	}
	if mods != modControl|modAlt {
		t.Fatalf("mods = %#x, want %#x", mods, modControl|modAlt)
	}
	if key != 0x87 {
		t.Fatalf("key = %#x, want %#x", key, uint32(0x87))
	}
}

func TestResizeOutcomeConstrainedError(t *testing.T) {
	outcome := resizeOutcome{
		requestedWidth:  1280,
		requestedHeight: 720,
		achievedWidth:   1200,
		achievedHeight:  700,
	}

	const want = "Example constrained resize: requested 1280x720, achieved 1200x700"
	if got := outcome.constrainedError("Example").Error(); got != want {
		t.Fatalf("constrainedError() = %q, want %q", got, want)
	}
}

func TestResizePosition(t *testing.T) {
	tests := []struct {
		name            string
		current         rect
		workArea        rect
		requestedWidth  int32
		requestedHeight int32
		center          bool
		want            point
	}{
		{
			name:            "centers within work area",
			current:         rect{Left: 100, Top: 100, Right: 900, Bottom: 700},
			workArea:        rect{Left: 0, Top: 0, Right: 1920, Bottom: 1040},
			requestedWidth:  1280,
			requestedHeight: 720,
			center:          true,
			want:            point{X: 320, Y: 160},
		},
		{
			name:            "oversized centered resize keeps title bar reachable",
			current:         rect{Left: 100, Top: 100, Right: 900, Bottom: 700},
			workArea:        rect{Left: 0, Top: 0, Right: 1920, Bottom: 1040},
			requestedWidth:  3840,
			requestedHeight: 2160,
			center:          true,
			want:            point{X: 0, Y: 0},
		},
		{
			name:            "non-centered resize preserves reachable title bar",
			current:         rect{Left: 120, Top: 400, Right: 920, Bottom: 1000},
			workArea:        rect{Left: 0, Top: 0, Right: 1920, Bottom: 1040},
			requestedWidth:  1280,
			requestedHeight: 720,
			want:            point{X: 120, Y: 400},
		},
		{
			name:            "non-centered resize brings title bar back onto screen",
			current:         rect{Left: 120, Top: 1100, Right: 920, Bottom: 1700},
			workArea:        rect{Left: 0, Top: 0, Right: 1920, Bottom: 1040},
			requestedWidth:  1280,
			requestedHeight: 720,
			want:            point{X: 120, Y: 1000},
		},
		{
			name:            "left monitor with negative origin",
			current:         rect{Left: -1800, Top: -700, Right: -1000, Bottom: -100},
			workArea:        rect{Left: -1920, Top: -1080, Right: 0, Bottom: 0},
			requestedWidth:  1280,
			requestedHeight: 720,
			center:          true,
			want:            point{X: -1600, Y: -900},
		},
		{
			name:            "right monitor",
			current:         rect{Left: 2000, Top: 100, Right: 2800, Bottom: 700},
			workArea:        rect{Left: 1920, Top: 0, Right: 3840, Bottom: 1040},
			requestedWidth:  3840,
			requestedHeight: 2160,
			center:          true,
			want:            point{X: 1920, Y: 0},
		},
		{
			name:            "monitor above primary",
			current:         rect{Left: 100, Top: -900, Right: 900, Bottom: -300},
			workArea:        rect{Left: 0, Top: -1040, Right: 1920, Bottom: 0},
			requestedWidth:  3840,
			requestedHeight: 2160,
			center:          true,
			want:            point{X: 0, Y: -1040},
		},
		{
			name:            "monitor below primary",
			current:         rect{Left: 100, Top: 1100, Right: 900, Bottom: 1700},
			workArea:        rect{Left: 0, Top: 1040, Right: 1920, Bottom: 2080},
			requestedWidth:  3840,
			requestedHeight: 2160,
			center:          true,
			want:            point{X: 0, Y: 1040},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := resizePosition(test.current, test.requestedWidth, test.requestedHeight, test.workArea, test.center)
			if got != test.want {
				t.Fatalf("resizePosition() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestQuoteExecutablePath(t *testing.T) {
	const exePath = `C:\Program Files\ResizeMe\resize-me.exe`
	const want = `"C:\Program Files\ResizeMe\resize-me.exe"`

	if got := quoteExecutablePath(exePath); got != want {
		t.Fatalf("quoteExecutablePath(%q) = %q, want %q", exePath, got, want)
	}
}

func TestQuoteExecutablePathAvoidsDoubleQuotes(t *testing.T) {
	const exePath = `"C:\Program Files\ResizeMe\resize-me.exe"`
	const want = `"C:\Program Files\ResizeMe\resize-me.exe"`

	if got := quoteExecutablePath(exePath); got != want {
		t.Fatalf("quoteExecutablePath(%q) = %q, want %q", exePath, got, want)
	}
}

func TestTargetMenuLabel(t *testing.T) {
	tests := []struct {
		name        string
		className   string
		processName string
		isResizeMe  bool
		want        string
	}{
		{
			name:       "ResizeMe",
			isResizeMe: true,
			want:       "Target: ResizeMe (not resizable)",
		},
		{
			name:      "desktop",
			className: "Progman",
			want:      "Target: Windows desktop (not resizable)",
		},
		{
			name:      "desktop worker window",
			className: "WorkerW",
			want:      "Target: Windows desktop (not resizable)",
		},
		{
			name:      "taskbar",
			className: "Shell_TrayWnd",
			want:      "Target: Windows taskbar (not resizable)",
		},
		{
			name:      "secondary taskbar",
			className: "Shell_SecondaryTrayWnd",
			want:      "Target: Windows taskbar (not resizable)",
		},
		{
			name:        "application",
			processName: "Code",
			want:        "Target: Code",
		},
		{
			name: "unresolved application",
			want: "Target: Active window",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := targetMenuLabel(test.className, test.processName, test.isResizeMe)
			if got != test.want {
				t.Fatalf("targetMenuLabel(%q, %q, %t) = %q, want %q", test.className, test.processName, test.isResizeMe, got, test.want)
			}
		})
	}
}

func TestProcessNameFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: `C:\Program Files\Microsoft VS Code\Code.exe`, want: "Code"},
		{path: `C:\Tools\ResizeMe`, want: "ResizeMe"},
	}

	for _, test := range tests {
		if got := processNameFromPath(test.path); got != test.want {
			t.Fatalf("processNameFromPath(%q) = %q, want %q", test.path, got, test.want)
		}
	}
}
