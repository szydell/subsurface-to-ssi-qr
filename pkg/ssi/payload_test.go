package ssi

import (
	"strings"
	"testing"
	"time"
)

func TestMapDiveAndBuildPayload(t *testing.T) {
	rec := DiveInput{
		StartTime:    time.Date(2025, 9, 20, 16, 23, 0, 0, time.UTC),
		DurationMin:  48.5,
		MaxDepthM:    26.4,
		DiveMode:     "scuba",
		WaterTypeRaw: "salt",
	}
	cfg := DefaultMappingConfig()
	mapped := MapDive(rec, cfg)

	payload, err := BuildPayload(mapped, false, ValidationStrict)
	if err != nil {
		t.Fatalf("BuildPayload error: %v", err)
	}

	expectedPrefix := "dive;noid;dive_type:0;divetime:48.5;datetime:202509201623;depth_m:26.4"
	if !strings.HasPrefix(payload, expectedPrefix) {
		t.Fatalf("payload prefix mismatch\n got: %s\nwant: %s", payload, expectedPrefix)
	}
	if !strings.Contains(payload, "var_watertype_id:5") {
		t.Fatalf("missing mapped water type in payload: %s", payload)
	}
}

func TestResolveWaterBodyID(t *testing.T) {
	tests := []struct {
		name string
		in   DiveInput
		cfg  MappingConfig
		want int
	}{
		{
			name: "per-dive override has precedence",
			in:   DiveInput{Site: "Red Sea", WaterBodyOverride: WaterBodyLake},
			cfg:  MappingConfig{WaterBodyRules: map[string]int{"red sea": WaterBodyQuarry}, WaterBodyID: WaterBodyRiver},
			want: WaterBodyLake,
		},
		{
			name: "site rule has precedence over inference",
			in:   DiveInput{Site: "Red Sea"},
			cfg:  MappingConfig{WaterBodyRules: map[string]int{"  RED sea  ": WaterBodyQuarry}},
			want: WaterBodyQuarry,
		},
		{
			name: "recognizes Polish keyword",
			in:   DiveInput{SiteDescription: "Jezioro Nieslysz"},
			cfg:  MappingConfig{},
			want: WaterBodyLake,
		},
		{
			name: "recognizes German keyword",
			in:   DiveInput{Tags: "Steinbruch"},
			cfg:  MappingConfig{},
			want: WaterBodyQuarry,
		},
		{
			// Real-world confirmed example: an SSI-generated QR payload for
			// a dive at an indoor facility named "Centrum Indoor" included
			// var_water_body_id:17, matching WaterBodyIndoor.
			name: "recognizes indoor facility name",
			in:   DiveInput{Site: "Centrum Indoor"},
			cfg:  MappingConfig{},
			want: WaterBodyIndoor,
		},
		{
			// "Deepspot" is a unique brand name for a dedicated dive pool
			// in Poland, so it unambiguously identifies an indoor facility.
			name: "recognizes Deepspot dive pool",
			in:   DiveInput{Site: "Deepspot"},
			cfg:  MappingConfig{},
			want: WaterBodyIndoor,
		},
		{
			name: "uses unambiguous inference",
			in:   DiveInput{SiteGeography: "Red Sea Egypt"},
			cfg:  MappingConfig{},
			want: WaterBodyOcean,
		},
		{
			name: "unknown place omits field by default",
			in:   DiveInput{Site: "Blue Wall"},
			cfg:  MappingConfig{},
			want: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ResolveWaterBodyID(test.in, test.cfg); got != test.want {
				t.Errorf("ResolveWaterBodyID() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestBuildPayloadFromDive_OmitsUnknownWaterBody(t *testing.T) {
	in := DiveInput{
		StartTime:   time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC),
		DurationMin: 33,
		MaxDepthM:   18,
		Site:        "Blue Wall",
	}

	payload, err := BuildPayloadFromDive(in, DefaultMappingConfig(), ValidationStrict)
	if err != nil {
		t.Fatalf("BuildPayloadFromDive error: %v", err)
	}
	if strings.Contains(payload, "var_water_body_id:") {
		t.Fatalf("unexpected water-body field in payload: %s", payload)
	}
}

func TestBuildPayload_StrictValidation(t *testing.T) {
	_, err := BuildPayload(Payload{}, false, ValidationStrict)
	if err == nil {
		t.Fatal("expected strict validation error, got nil")
	}
}

func TestBuildPayloadFromDive_UsesConfigFlag(t *testing.T) {
	rec := DiveInput{
		StartTime:   time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC),
		DurationMin: 33,
		MaxDepthM:   18,
	}
	cfg := DefaultMappingConfig()
	cfg.IncludeUserIDs = true
	cfg.UserFirstName = "Ada"
	cfg.UserLastName = "Lovelace"

	payload, err := BuildPayloadFromDive(rec, cfg, ValidationStrict)
	if err != nil {
		t.Fatalf("BuildPayloadFromDive error: %v", err)
	}

	if !strings.Contains(payload, "user_firstname:Ada") {
		t.Fatalf("expected user_firstname in payload, got: %s", payload)
	}
	if !strings.Contains(payload, "user_lastname:Lovelace") {
		t.Fatalf("expected user_lastname in payload, got: %s", payload)
	}
}
