package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/szydell/subsurface-to-ssi-qr/internal/config"
	"github.com/szydell/subsurface-to-ssi-qr/internal/model"
)

func TestLoadDivesFromFile(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "sample_subsurface.xml")
	cfg := config.DefaultMapping()

	items, err := loadDivesFromFile(path, cfg)
	if err != nil {
		t.Fatalf("loadDivesFromFile returned error: %v", err)
	}
	if got, want := len(items), 2; got != want {
		t.Fatalf("unexpected item count: got %d, want %d", got, want)
	}

	first := items[0]
	if first.Index != 1 {
		t.Errorf("unexpected index: got %d, want 1", first.Index)
	}
	if first.SiteText != "Blue Wall" {
		t.Errorf("unexpected site: got %q, want %q", first.SiteText, "Blue Wall")
	}
	if first.Payload == "" {
		t.Error("expected non-empty payload")
	}
}

func TestLoadDivesFromFile_MissingFile(t *testing.T) {
	cfg := config.DefaultMapping()

	if _, err := loadDivesFromFile("does-not-exist.xml", cfg); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestNormalizeSite(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "Blue Wall", want: "Blue Wall"},
		{in: "  ", want: "-"},
		{in: "", want: "-"},
		{in: "  Lagoon  ", want: "Lagoon"},
	}

	for _, tc := range tests {
		if got := normalizeSite(tc.in); got != tc.want {
			t.Errorf("normalizeSite(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMapDivesToItemsWithOverrides(t *testing.T) {
	dives := []model.DiveRecord{{
		StartTime:   time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC),
		DurationMin: 33,
		MaxDepthM:   18,
		Site:        "Red Sea",
	}}

	items := mapDivesToItemsWithOverrides(dives, config.DefaultMapping(), map[int]int{0: -1})
	if len(items) != 1 {
		t.Fatalf("expected one item, got %d", len(items))
	}
	if strings.Contains(items[0].Payload, "var_water_body_id:") {
		t.Fatalf("expected omitted water-body field, got: %s", items[0].Payload)
	}
}

func TestFormatDiveRow(t *testing.T) {
	item := diveListItem{
		Index:        1,
		WhenText:     "2025-09-20 16:23",
		DurationText: "48.5 min",
		DepthText:    "26.4 m",
		SiteText:     "Blue Wall",
	}

	got := formatDiveRow(item, "Ocean")
	want := "1   2025-09-20 16:23  48.5 min   26.4 m   Ocean      Blue Wall"
	if got != want {
		t.Errorf("formatDiveRow() = %q, want %q", got, want)
	}
}

func TestFormatDiveRow_BlankWaterBodyLabel(t *testing.T) {
	item := diveListItem{
		Index:        2,
		WhenText:     "2025-09-21 08:12",
		DurationText: "39.0 min",
		DepthText:    "18.2 m",
		SiteText:     "Lagoon",
	}

	got := formatDiveRow(item, "")
	want := "2   2025-09-21 08:12  39.0 min   18.2 m              Lagoon"
	if got != want {
		t.Errorf("formatDiveRow() = %q, want %q", got, want)
	}
}

func TestDiveRowColor(t *testing.T) {
	selected := diveRowColor(0, true)
	if selected.R != 212 || selected.G != 225 || selected.B != 245 {
		t.Errorf("selected row color = %+v, want highlight color", selected)
	}

	even := diveRowColor(2, false)
	if even.R != 255 || even.G != 255 || even.B != 255 {
		t.Errorf("even row color = %+v, want white", even)
	}

	odd := diveRowColor(1, false)
	if odd.R != 246 || odd.G != 248 || odd.B != 251 {
		t.Errorf("odd row color = %+v, want zebra stripe", odd)
	}
}

func TestEnsurePNGExtension(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "dive.png", want: "dive.png"},
		{in: "dive.PNG", want: "dive.PNG"},
		{in: "dive", want: "dive.png"},
		{in: "dive.jpg", want: "dive.jpg.png"},
	}

	for _, tc := range tests {
		if got := ensurePNGExtension(tc.in); got != tc.want {
			t.Errorf("ensurePNGExtension(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
