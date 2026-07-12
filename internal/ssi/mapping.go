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

	WaterBodyOcean     = publicssi.WaterBodyOcean
	WaterBodyRiver     = publicssi.WaterBodyRiver
	WaterBodyQuarry    = publicssi.WaterBodyQuarry
	WaterBodyLake      = publicssi.WaterBodyLake
	WaterBodyIndoor    = publicssi.WaterBodyIndoor
	WaterBodyOpenWater = publicssi.WaterBodyOpenWater
)

// WaterBodyOptions lists the currently known SSI water-body dictionary
// entries. The SSI format is reverse-engineered; this list is not
// guaranteed to be exhaustive.
var WaterBodyOptions = publicssi.WaterBodyOptions

// MapDive converts a normalized Subsurface dive into SSI payload values.
func MapDive(in model.DiveRecord, cfg config.MappingConfig) Payload {
	return publicssi.MapDive(publicssi.DiveInput{
		StartTime:         in.StartTime,
		DurationMin:       in.DurationMin,
		MaxDepthM:         in.MaxDepthM,
		DiveMode:          in.DiveMode,
		Site:              in.Site,
		SiteDescription:   in.SiteDescription,
		SiteNotes:         in.SiteNotes,
		SiteGeography:     in.SiteGeography,
		Tags:              in.Tags,
		Notes:             in.Notes,
		WaterBodyOverride: in.WaterBodyOverride,
		WaterTypeRaw:      in.WaterTypeRaw,
		AirTempC:          in.AirTempC,
		WaterTempC:        in.WaterTempC,
		VisibilityM:       in.VisibilityM,
	}, publicssi.MappingConfig(cfg))
}
