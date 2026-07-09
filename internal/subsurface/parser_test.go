package subsurface

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_Sample(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "sample_subsurface.xml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("sample file missing: %v", err)
	}

	dives, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if got, want := len(dives), 2; got != want {
		t.Fatalf("unexpected dive count: got %d, want %d", got, want)
	}

	if dives[0].DurationMin <= 48.0 || dives[0].DurationMin >= 49.0 {
		t.Fatalf("unexpected parsed duration: %v", dives[0].DurationMin)
	}
	if dives[0].MaxDepthM != 26.4 {
		t.Fatalf("unexpected depth: %v", dives[0].MaxDepthM)
	}
	if dives[0].Site != "Blue Wall" {
		t.Fatalf("unexpected site: %q", dives[0].Site)
	}
}

func TestParseFile_RealDataSSRF(t *testing.T) {
	path := filepath.Join("..", "..", "..", "data", "addr@email.com.ssrf")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("real data fixture missing: %v", err)
	}

	dives, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile returned error for real SSRF: %v", err)
	}
	if len(dives) == 0 {
		t.Fatal("expected at least one parsed dive from real SSRF")
	}

	first := dives[0]
	if first.StartTime.IsZero() {
		t.Fatal("first parsed dive has zero start time")
	}
	if first.DurationMin <= 0 {
		t.Fatalf("first parsed dive has non-positive duration: %v", first.DurationMin)
	}
	if first.MaxDepthM <= 0 {
		t.Fatalf("first parsed dive has non-positive max depth: %v", first.MaxDepthM)
	}
	if first.Site == "" {
		t.Fatal("first parsed dive has empty site; expected divesiteid resolution")
	}
}
