package main

import (
	"testing"
	"time"

	"github.com/szydell/subsurface-to-ssi-qr/internal/model"
	"github.com/szydell/subsurface-to-ssi-qr/internal/ssi"
)

func TestSameSite(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{a: "Blue Wall", b: "Blue Wall", want: true},
		{a: "Blue Wall", b: "blue wall", want: true},
		{a: "  Blue Wall  ", b: "Blue Wall", want: true},
		{a: "Blue Wall", b: "Lagoon", want: false},
		{a: "", b: "", want: true},
	}
	for _, tc := range tests {
		if got := sameSite(tc.a, tc.b); got != tc.want {
			t.Errorf("sameSite(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestWaterBodyLabelRoundTrip(t *testing.T) {
	tr, err := newTranslator("pl")
	if err != nil {
		t.Fatalf("newTranslator: %v", err)
	}
	s := &appState{tr: tr}

	ids := []int{0, -1, ssi.WaterBodyOcean, ssi.WaterBodyRiver, ssi.WaterBodyQuarry, ssi.WaterBodyLake, ssi.WaterBodyIndoor, ssi.WaterBodyOpenWater}
	for _, id := range ids {
		label := s.waterBodyOptionLabel(id, true)
		if label == "" {
			t.Fatalf("empty label for id %d", id)
		}
		got, ok := s.waterBodyIDForLabel(label)
		if !ok || got != id {
			t.Fatalf("round trip failed for id %d: label=%q got=%d ok=%v", id, label, got, ok)
		}
	}
}

func TestWaterBodyColumnLabel(t *testing.T) {
	tr, err := newTranslator("en")
	if err != nil {
		t.Fatalf("newTranslator: %v", err)
	}
	s := &appState{tr: tr}

	if got := s.waterBodyColumnLabel(0); got != "" {
		t.Errorf("waterBodyColumnLabel(0) = %q, want empty", got)
	}
	if got := s.waterBodyColumnLabel(-1); got != "" {
		t.Errorf("waterBodyColumnLabel(-1) = %q, want empty", got)
	}
	if got := s.waterBodyColumnLabel(ssi.WaterBodyQuarry); got == "" {
		t.Error("waterBodyColumnLabel(quarry) is empty, want a label")
	}
}

func TestApplyWaterBodyChoice_SingleDiveOnly(t *testing.T) {
	tr, err := newTranslator("en")
	if err != nil {
		t.Fatalf("newTranslator: %v", err)
	}

	s := &appState{
		tr: tr,
		parsedDives: []model.DiveRecord{
			{Site: "Red Sea", StartTime: time.Now(), DurationMin: 10, MaxDepthM: 5},
			{Site: "red sea", StartTime: time.Now(), DurationMin: 10, MaxDepthM: 5},
		},
		waterBodyOverrides: make(map[int]int),
		selectedDiveID:     -1,
		selectedDiveIndex:  -1,
	}

	s.applyWaterBodyChoice(0, ssi.WaterBodyOcean, false)

	if got := s.waterBodyOverrides[0]; got != ssi.WaterBodyOcean {
		t.Errorf("waterBodyOverrides[0] = %d, want %d", got, ssi.WaterBodyOcean)
	}
	if _, ok := s.waterBodyOverrides[1]; ok {
		t.Error("expected the other dive to remain unaffected without apply-all")
	}
}

func TestApplyWaterBodyChoice_AppliesToMatchingSites(t *testing.T) {
	tr, err := newTranslator("en")
	if err != nil {
		t.Fatalf("newTranslator: %v", err)
	}

	s := &appState{
		tr: tr,
		parsedDives: []model.DiveRecord{
			{Site: "Red Sea", StartTime: time.Now(), DurationMin: 10, MaxDepthM: 5},
			{Site: "  red sea  ", StartTime: time.Now(), DurationMin: 10, MaxDepthM: 5},
			{Site: "Lagoon", StartTime: time.Now(), DurationMin: 10, MaxDepthM: 5},
		},
		waterBodyOverrides: make(map[int]int),
		selectedDiveID:     -1,
		selectedDiveIndex:  -1,
	}

	s.applyWaterBodyChoice(0, ssi.WaterBodyOcean, true)

	if got := s.waterBodyOverrides[0]; got != ssi.WaterBodyOcean {
		t.Errorf("waterBodyOverrides[0] = %d, want %d", got, ssi.WaterBodyOcean)
	}
	if got := s.waterBodyOverrides[1]; got != ssi.WaterBodyOcean {
		t.Errorf("waterBodyOverrides[1] = %d, want %d (matching site)", got, ssi.WaterBodyOcean)
	}
	if _, ok := s.waterBodyOverrides[2]; ok {
		t.Error("unrelated site must not be affected")
	}
}

func TestApplyWaterBodyChoice_AutomaticClearsOverride(t *testing.T) {
	tr, err := newTranslator("en")
	if err != nil {
		t.Fatalf("newTranslator: %v", err)
	}

	s := &appState{
		tr: tr,
		parsedDives: []model.DiveRecord{
			{Site: "Red Sea", StartTime: time.Now(), DurationMin: 10, MaxDepthM: 5},
		},
		waterBodyOverrides: map[int]int{0: ssi.WaterBodyQuarry},
		selectedDiveID:     -1,
		selectedDiveIndex:  -1,
	}

	s.applyWaterBodyChoice(0, 0, false)

	if _, ok := s.waterBodyOverrides[0]; ok {
		t.Error("expected automatic choice to clear the override")
	}
}
