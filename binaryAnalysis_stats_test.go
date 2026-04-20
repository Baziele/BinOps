//go:build linux

package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestParseElfStatsPopulatesUIDFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	st := ParseElfStats(nil, path)
	if !st.HasUID || !st.HasGID {
		t.Fatalf("expected uid/gid to be populated, got HasUID=%v HasGID=%v", st.HasUID, st.HasGID)
	}
	want := uint32(os.Getuid())
	if st.UID != want {
		t.Errorf("UID: got %d want %d", st.UID, want)
	}
	wantg := uint32(os.Getgid())
	if st.GID != wantg {
		t.Errorf("GID: got %d want %d", st.GID, wantg)
	}
	if u, err := user.LookupId(fmt.Sprintf("%d", want)); err == nil {
		if st.UIDName != u.Username {
			t.Errorf("UIDName: got %q want %q", st.UIDName, u.Username)
		}
	}
	if g, err := user.LookupGroupId(fmt.Sprintf("%d", wantg)); err == nil {
		if st.GIDName != g.Name {
			t.Errorf("GIDName: got %q want %q", st.GIDName, g.Name)
		}
	}
}
