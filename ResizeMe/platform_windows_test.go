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
