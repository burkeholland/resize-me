//go:build windows

package main

import "testing"

func TestFormatVersionOmitsZeroRevision(t *testing.T) {
	if got := formatVersion(1<<16|2, 3<<16); got != "1.2.3" {
		t.Fatalf("formatVersion() = %q, want %q", got, "1.2.3")
	}
}

func TestFormatVersionIncludesNonzeroRevision(t *testing.T) {
	if got := formatVersion(1<<16|2, 3<<16|4); got != "1.2.3.4" {
		t.Fatalf("formatVersion() = %q, want %q", got, "1.2.3.4")
	}
}

func TestProductVersionUsesExecutableProductMetadata(t *testing.T) {
	info := &fixedFileInfo{
		FileVersionMS:    1<<16 | 2,
		FileVersionLS:    3 << 16,
		ProductVersionMS: 9<<16 | 9,
		ProductVersionLS: 9<<16 | 9,
	}
	if got := productVersion(info); got != "9.9.9.9" {
		t.Fatalf("productVersion() = %q, want %q", got, "9.9.9.9")
	}
}
