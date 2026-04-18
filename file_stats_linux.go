//go:build linux

package main

import (
	"os"
	"syscall"
)

func populateELFStats(out *ELFStats, fileInfo os.FileInfo) {
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return
	}

	out.Blocks = stat.Blocks
	out.HasBlocks = true
	out.BlockSize = stat.Blksize
	out.HasBlockSize = true
	out.Links = uint64(stat.Nlink)
	out.HasLinks = true
	out.UID = stat.Uid
	out.HasUID = true
	out.GID = stat.Gid
	out.HasGID = true
	out.LastAccessTime = stat.Atim.Sec
	out.HasLastAccessTime = true
	out.LastModificationTime = stat.Mtim.Sec
	out.HasLastModTime = true
}
