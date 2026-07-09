package ssi

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidationMode controls how strict required-field checks are.
type ValidationMode string

const (
	ValidationLenient ValidationMode = "lenient"
	ValidationStrict  ValidationMode = "strict"
)

// BuildPayload converts payload struct into SSI QR text format.
func BuildPayload(p Payload, includeUser bool, mode ValidationMode) (string, error) {
	if mode == ValidationStrict {
		if err := ValidateRequired(p); err != nil {
			return "", err
		}
	}

	parts := []string{"dive", "noid"}

	parts = append(parts,
		kv("dive_type", strconv.Itoa(p.DiveType)),
		kv("divetime", formatFloat(p.DiveTimeMin)),
		kv("datetime", p.DateTime),
		kv("depth_m", formatFloat(p.DepthM)),
	)

	if p.SiteID > 0 {
		parts = append(parts, kv("site", strconv.Itoa(p.SiteID)))
	}
	if p.VarWeatherID > 0 {
		parts = append(parts, kv("var_weather_id", strconv.Itoa(p.VarWeatherID)))
	}
	if p.VarEntryID > 0 {
		parts = append(parts, kv("var_entry_id", strconv.Itoa(p.VarEntryID)))
	}
	if p.VarWaterBody > 0 {
		parts = append(parts, kv("var_water_body_id", strconv.Itoa(p.VarWaterBody)))
	}
	if p.VarWaterType > 0 {
		parts = append(parts, kv("var_watertype_id", strconv.Itoa(p.VarWaterType)))
	}
	if p.VarCurrentID > 0 {
		parts = append(parts, kv("var_current_id", strconv.Itoa(p.VarCurrentID)))
	}
	if p.VarSurfaceID > 0 {
		parts = append(parts, kv("var_surface_id", strconv.Itoa(p.VarSurfaceID)))
	}
	if p.VarDiveTypeID > 0 {
		parts = append(parts, kv("var_divetype_id", strconv.Itoa(p.VarDiveTypeID)))
	}

	if includeUser {
		if p.UserMasterID > 0 {
			parts = append(parts, kv("user_master_id", strconv.Itoa(p.UserMasterID)))
		}
		if p.UserFirstName != "" {
			parts = append(parts, kv("user_firstname", sanitize(p.UserFirstName)))
		}
		if p.UserLastName != "" {
			parts = append(parts, kv("user_lastname", sanitize(p.UserLastName)))
		}
		if p.UserLeaderID > 0 {
			parts = append(parts, kv("user_leader_id", strconv.Itoa(p.UserLeaderID)))
		}
	}

	if p.AirTempC != nil {
		parts = append(parts, kv("airtemp_c", formatFloat(*p.AirTempC)))
	}
	if p.WaterTempC != nil {
		parts = append(parts, kv("watertemp_c", formatFloat(*p.WaterTempC)))
	}
	if p.Visibility != nil {
		parts = append(parts, kv("vis_m", formatFloat(*p.Visibility)))
	}

	return strings.Join(parts, ";"), nil
}

// ValidateRequired verifies required fields for strict mode.
func ValidateRequired(p Payload) error {
	if p.DateTime == "" {
		return fmt.Errorf("missing required field: datetime")
	}
	if p.DiveTimeMin <= 0 {
		return fmt.Errorf("missing required field: divetime")
	}
	if p.DepthM <= 0 {
		return fmt.Errorf("missing required field: depth_m")
	}
	return nil
}

func kv(k, v string) string {
	return k + ":" + v
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func sanitize(v string) string {
	return strings.TrimSpace(strings.ReplaceAll(v, ";", " "))
}
