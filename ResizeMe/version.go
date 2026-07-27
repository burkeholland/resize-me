package main

import (
	"runtime/debug"
	"strings"
)

const developmentVersion = "Development"

func buildInfoVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return developmentVersion
	}

	version := strings.TrimPrefix(strings.TrimSpace(info.Main.Version), "v")
	if version == "" || version == "(devel)" {
		return developmentVersion
	}
	return version
}
