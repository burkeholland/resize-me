//go:build !windows

package main

func applicationVersion() string {
	return buildInfoVersion()
}
