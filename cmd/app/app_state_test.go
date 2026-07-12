package main

import (
	"os"
	"testing"
)

func TestResolveStartDir_SavedDirExists(t *testing.T) {
	tmp := t.TempDir()

	got := resolveStartDir(tmp)
	if got != tmp {
		t.Errorf("resolveStartDir(%q) = %q, want %q", tmp, got, tmp)
	}
}

func TestResolveStartDir_SavedDirMissingFallsBackToCwd(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	got := resolveStartDir("/path/does/not/exist-really-not")
	if got != wd {
		t.Errorf("resolveStartDir(missing) = %q, want cwd %q", got, wd)
	}
}

func TestResolveStartDir_SavedDirIsFileFallsBackToCwd(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	file := t.TempDir() + "/not-a-dir.txt"
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	got := resolveStartDir(file)
	if got != wd {
		t.Errorf("resolveStartDir(file) = %q, want cwd %q", got, wd)
	}
}

func TestResolveStartDir_EmptySavedDirFallsBackToCwd(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	got := resolveStartDir("   ")
	if got != wd {
		t.Errorf("resolveStartDir(blank) = %q, want cwd %q", got, wd)
	}
}
