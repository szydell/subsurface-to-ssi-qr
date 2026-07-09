package ssi

import (
	"testing"
	"time"

	"github.com/szydell/subsurface-to-ssi-qr/internal/config"
	"github.com/szydell/subsurface-to-ssi-qr/internal/model"
	publicssi "github.com/szydell/subsurface-to-ssi-qr/pkg/ssi"
)

func TestMapDive_UsesInternalTypes(t *testing.T) {
	rec := model.DiveRecord{
		StartTime:    time.Date(2025, 9, 20, 16, 23, 0, 0, time.UTC),
		DurationMin:  48.5,
		MaxDepthM:    26.4,
		DiveMode:     "scuba",
		WaterTypeRaw: "salt",
	}
	cfg := config.DefaultMapping()
	mapped := MapDive(rec, cfg)

	if mapped.VarWaterType != 5 {
		t.Fatalf("expected mapped water type 5 for salt water, got: %d", mapped.VarWaterType)
	}
	if mapped.DateTime != "202509201623" {
		t.Fatalf("unexpected mapped datetime: %s", mapped.DateTime)
	}
}

func TestBuildPayload_MatchesPublicPackage(t *testing.T) {
	p := Payload{
		DiveType:    DiveTypeScuba,
		DiveTimeMin: 31,
		DateTime:    "202601011030",
		DepthM:      18,
	}

	got, err := BuildPayload(p, false, ValidationStrict)
	if err != nil {
		t.Fatalf("internal BuildPayload error: %v", err)
	}

	want, err := publicssi.BuildPayload(publicssi.Payload(p), false, publicssi.ValidationStrict)
	if err != nil {
		t.Fatalf("public BuildPayload error: %v", err)
	}

	if got != want {
		t.Fatalf("adapter mismatch\n got: %s\nwant: %s", got, want)
	}
}
