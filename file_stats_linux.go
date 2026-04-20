//go:build linux

package main

import (
	"fmt"
	"os"
	"os/user"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func populateELFStats(out *ELFStats, path string, fileInfo os.FileInfo) {
	// Prefer unix.Stat(path): it always fills uid/gid from the inode for this path.
	// Relying only on os.FileInfo.Sys() can fail the *syscall.Stat_t assertion on some setups
	// or return incomplete metadata for certain filesystems.
	var st unix.Stat_t
	if err := unix.Stat(path, &st); err == nil {
		out.Blocks = st.Blocks
		out.HasBlocks = true
		out.BlockSize = st.Blksize
		out.HasBlockSize = true
		out.Links = st.Nlink
		out.HasLinks = true
		out.UID = st.Uid
		out.HasUID = true
		out.GID = st.Gid
		out.HasGID = true
		out.LastAccessAt = time.Unix(st.Atim.Sec, st.Atim.Nsec)
		out.HasLastAccessTime = true
		out.LastModAt = time.Unix(st.Mtim.Sec, st.Mtim.Nsec)
		out.HasLastModTime = true
		resolveOwnerNames(out)
		return
	}

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
	out.LastAccessAt = time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
	out.HasLastAccessTime = true
	out.LastModAt = time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
	out.HasLastModTime = true
	resolveOwnerNames(out)
}

func resolveOwnerNames(out *ELFStats) {
	if out.HasUID {
		if u, err := user.LookupId(fmt.Sprintf("%d", out.UID)); err == nil {
			out.UIDName = u.Username
		}
	}
	if out.HasGID {
		if g, err := user.LookupGroupId(fmt.Sprintf("%d", out.GID)); err == nil {
			out.GIDName = g.Name
		}
	}
}
