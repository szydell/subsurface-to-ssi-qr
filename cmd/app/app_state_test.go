package main

import (
	"os"
	"testing"
	"time"
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

func newSortTestState(t *testing.T) *appState {
	t.Helper()
	tr, err := newTranslator("en")
	if err != nil {
		t.Fatalf("newTranslator: %v", err)
	}
	return &appState{
		tr:                tr,
		selectedDiveID:    -1,
		selectedDiveIndex: -1,
		sortColumn:        -1,
		listItems: []diveListItem{
			{Index: 1, SiteText: "Zulu", WhenTime: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC), DurationMin: 30, DepthM: 10},
			{Index: 2, SiteText: "Alpha", WhenTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), DurationMin: 50, DepthM: 30},
			{Index: 3, SiteText: "mango", WhenTime: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), DurationMin: 20, DepthM: 20},
		},
	}
}

func TestSortListItems_BySiteAscThenDesc(t *testing.T) {
	s := newSortTestState(t)
	s.sortColumn = 5
	s.sortAscending = true
	s.sortListItems()

	got := []string{s.listItems[0].SiteText, s.listItems[1].SiteText, s.listItems[2].SiteText}
	want := []string{"Alpha", "mango", "Zulu"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ascending site sort = %v, want %v", got, want)
		}
	}

	s.sortAscending = false
	s.sortListItems()
	got = []string{s.listItems[0].SiteText, s.listItems[1].SiteText, s.listItems[2].SiteText}
	want = []string{"Zulu", "mango", "Alpha"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("descending site sort = %v, want %v", got, want)
		}
	}
}

func TestSortListItems_ByDate(t *testing.T) {
	s := newSortTestState(t)
	s.sortColumn = 1
	s.sortAscending = true
	s.sortListItems()

	got := []int{s.listItems[0].Index, s.listItems[1].Index, s.listItems[2].Index}
	want := []int{2, 3, 1}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ascending date sort = %v, want %v", got, want)
		}
	}
}

func TestSortListItems_NoSortColumnIsNoOp(t *testing.T) {
	s := newSortTestState(t)
	before := append([]diveListItem(nil), s.listItems...)
	s.sortListItems()
	for i := range before {
		if s.listItems[i].Index != before[i].Index {
			t.Fatalf("sortListItems with sortColumn=-1 changed order: got %+v, want %+v", s.listItems, before)
		}
	}
}

func TestSyncSelectedRow_FollowsDiveAcrossResort(t *testing.T) {
	s := newSortTestState(t)
	// Select the dive with Index 2 ("Alpha"), initially at row 1.
	s.selectedDiveIndex = 1

	s.sortColumn = 5
	s.sortAscending = true
	s.sortListItems()
	s.syncSelectedRow()

	if got := s.listItems[s.selectedDiveID].Index; got != 2 {
		t.Errorf("selected row now points at Index %d, want 2 (Alpha)", got)
	}
}

func TestSyncSelectedRow_ClearsWhenDiveGone(t *testing.T) {
	s := newSortTestState(t)
	s.selectedDiveIndex = 99 // no dive with this original index
	s.syncSelectedRow()

	if s.selectedDiveID != -1 {
		t.Errorf("selectedDiveID = %d, want -1 when the selected dive is no longer present", s.selectedDiveID)
	}
}
