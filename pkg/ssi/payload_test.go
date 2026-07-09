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
