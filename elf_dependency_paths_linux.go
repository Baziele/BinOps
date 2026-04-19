//go:build linux

package main

import (
	"os"
	"path/filepath"
)

var standardPaths = []string{
	"/lib",
	"/lib64",
	"/usr/lib",
	"/usr/lib64",
	"/usr/local/lib",
	"/usr/lib/x86_64-linux-gnu", // Common on Ubuntu/Debian
}

func resolveSystemELFDependencyPath(lib string) string {
	for _, searchPath := range standardPaths {
		fullPath := filepath.Join(searchPath, lib)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}
	return ""
}
