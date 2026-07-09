package ssi

import (
	"github.com/szydell/subsurface-to-ssi-qr/internal/config"
	"github.com/szydell/subsurface-to-ssi-qr/internal/model"
	publicssi "github.com/szydell/subsurface-to-ssi-qr/pkg/ssi"
)

const (
	DiveTypeScuba                   = publicssi.DiveTypeScuba
	DiveTypeExtendedRange           = publicssi.DiveTypeExtendedRange
	DiveTypeRebreatherSelfContained = publicssi.DiveTypeRebreatherSelfContained
	DiveTypeFreediving              = publicssi.DiveTypeFreediving
	DiveTypeRebreatherClosedCircuit = publicssi.DiveTypeRebreatherClosedCircuit
)

// MapDive converts a normalized Subsurface dive into SSI payload values.
func MapDive(in model.DiveRecord, cfg config.MappingConfig) Payload {
	return publicssi.MapDive(publicssi.DiveInput{
		StartTime:    in.StartTime,
		DurationMin:  in.DurationMin,
		MaxDepthM:    in.MaxDepthM,
		DiveMode:     in.DiveMode,
		WaterTypeRaw: in.WaterTypeRaw,
		AirTempC:     in.AirTempC,
		WaterTempC:   in.WaterTempC,
		VisibilityM:  in.VisibilityM,
	}, publicssi.MappingConfig(cfg))
}
