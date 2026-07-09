package ssi

import (
	"strings"

	"subsurface-to-ssi-qr/internal/config"
	"subsurface-to-ssi-qr/internal/model"
)

const (
	DiveTypeScuba                   = 0
	DiveTypeExtendedRange           = 2
	DiveTypeRebreatherSelfContained = 4
	DiveTypeFreediving              = 6
	DiveTypeRebreatherClosedCircuit = 8
)

// Payload is a typed representation of SSI QR fields.
type Payload struct {
	DiveType      int
	DiveTimeMin   float64
	DateTime      string
	DepthM        float64
	SiteID        int
	VarWeatherID  int
	VarEntryID    int
	VarWaterBody  int
	VarWaterType  int
	VarCurrentID  int
	VarSurfaceID  int
	VarDiveTypeID int

	UserMasterID  int
	UserFirstName string
	UserLastName  string
	UserLeaderID  int

	AirTempC   *float64
	WaterTempC *float64
	Visibility *float64
}

// MapDive converts a normalized Subsurface dive into SSI payload values.
func MapDive(in model.DiveRecord, cfg config.MappingConfig) Payload {
	waterTypeID := cfg.WaterTypeID
	rawWater := strings.ToLower(strings.TrimSpace(in.WaterTypeRaw))
	if strings.Contains(rawWater, "fresh") {
		waterTypeID = 4
	} else if strings.Contains(rawWater, "salt") || strings.Contains(rawWater, "sea") {
		waterTypeID = 5
	}

	return Payload{
		DiveType:      mapDiveType(in.DiveMode),
		DiveTimeMin:   in.DurationMin,
		DateTime:      in.StartTime.Format("200601021504"),
		DepthM:        in.MaxDepthM,
		SiteID:        cfg.SiteID,
		VarWeatherID:  cfg.WeatherID,
		VarEntryID:    cfg.EntryID,
		VarWaterBody:  cfg.WaterBodyID,
		VarWaterType:  waterTypeID,
		VarCurrentID:  cfg.CurrentID,
		VarSurfaceID:  cfg.SurfaceID,
		VarDiveTypeID: cfg.DiveSubtypeID,
		UserMasterID:  cfg.UserMasterID,
		UserFirstName: cfg.UserFirstName,
		UserLastName:  cfg.UserLastName,
		UserLeaderID:  cfg.UserLeaderID,
		AirTempC:      in.AirTempC,
		WaterTempC:    in.WaterTempC,
		Visibility:    in.VisibilityM,
	}
}

func mapDiveType(mode string) int {
	raw := strings.ToLower(strings.TrimSpace(mode))
	switch raw {
	case "freedive", "freediving":
		return DiveTypeFreediving
	case "extended_range", "tec", "technical":
		return DiveTypeExtendedRange
	case "rebreather_scr", "pscr", "scr":
		return DiveTypeRebreatherSelfContained
	case "rebreather_ccr", "ccr":
		return DiveTypeRebreatherClosedCircuit
	default:
		return DiveTypeScuba
	}
}
