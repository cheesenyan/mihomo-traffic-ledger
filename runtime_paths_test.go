package main

import (
	"path/filepath"
	"testing"
)

func TestUserDataRootUsesLocalAppData(t *testing.T) {
	localData := t.TempDir()
	t.Setenv("LOCALAPPDATA", localData)
	if got, want := userDataRoot(), filepath.Join(localData, "ClashTrafficMonitor"); got != want {
		t.Fatalf("userDataRoot = %q, want %q", got, want)
	}
}
