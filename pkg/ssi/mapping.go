package ssi

import "strings"

// MapDive converts a normalized dive into SSI payload values.
func MapDive(in DiveInput, cfg MappingConfig) Payload {
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

// BuildPayloadFromDive maps and serializes a dive in one call.
func BuildPayloadFromDive(in DiveInput, cfg MappingConfig, mode ValidationMode) (string, error) {
	mapped := MapDive(in, cfg)
	return BuildPayload(mapped, cfg.IncludeUserIDs, mode)
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
