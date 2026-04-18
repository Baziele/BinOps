//go:build !linux

package main

import "runtime"

func resolveSystemELFDependencyPath(_ string) string {
	return "system ELF library lookup unsupported on " + runtime.GOOS
}
