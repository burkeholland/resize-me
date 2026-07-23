//go:build windows

package main

import "testing"

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
