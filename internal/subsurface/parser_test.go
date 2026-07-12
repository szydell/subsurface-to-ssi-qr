package subsurface

import (
	"os"
	"path/filepath"
	"strings"
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
	if first.SiteGeography == "" {
		t.Fatal("first parsed dive has empty site geography")
	}
}

func TestParse_SiteMetadataForWaterBodyMapping(t *testing.T) {
	input := `<divelog>
		<divesites>
			<site uuid="lake-1" name="Lake Alpha" description="Freshwater lake">
				<notes>Calm shore entry</notes>
				<geo cat="1" value="Poland"/>
			</site>
		</divesites>
		<dives>
			<dive date="2026-01-01" time="10:30" duration="30 min" maxdepth="18 m" divesiteid="lake-1" tags="training, lake">
				<notes>Night dive</notes>
			</dive>
		</dives>
	</divelog>`

	dives, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if got, want := len(dives), 1; got != want {
		t.Fatalf("unexpected dive count: got %d, want %d", got, want)
	}

	dive := dives[0]
	if dive.Site != "Lake Alpha" || dive.SiteDescription != "Freshwater lake" || dive.SiteNotes != "Calm shore entry" || dive.SiteGeography != "Poland" || dive.Tags != "training, lake" || dive.Notes != "Night dive" {
		t.Fatalf("unexpected parsed metadata: %+v", dive)
	}
}
