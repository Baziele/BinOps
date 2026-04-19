//go:build !linux

package main

import "os"

func populateELFStats(_ *ELFStats, _ os.FileInfo) {}
