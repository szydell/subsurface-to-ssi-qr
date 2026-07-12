package ssi

import "strings"

// WaterBodyOptions contains the currently known SSI water-body dictionary.
// The SSI format is reverse-engineered; callers should not assume it is exhaustive.
var WaterBodyOptions = []WaterBodyOption{
	{ID: WaterBodyOcean, Key: "ocean"},
	{ID: WaterBodyRiver, Key: "river"},
	{ID: WaterBodyQuarry, Key: "quarry"},
	{ID: WaterBodyLake, Key: "lake"},
	{ID: WaterBodyIndoor, Key: "indoor"},
	{ID: WaterBodyOpenWater, Key: "open_water"},
}

// WaterBodyOption is a known SSI water-body dictionary value.
type WaterBodyOption struct {
	ID  int
	Key string
}

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
		VarWaterBody:  ResolveWaterBodyID(in, cfg),
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

// ResolveWaterBodyID returns the SSI water-body value for a dive. Explicit
// per-dive overrides take precedence, followed by exact site rules, then
// unambiguous local text matching, and finally the configured fallback.
func ResolveWaterBodyID(in DiveInput, cfg MappingConfig) int {
	if in.WaterBodyOverride < 0 {
		return 0
	}
	if isKnownWaterBody(in.WaterBodyOverride) {
		return in.WaterBodyOverride
	}

	normalizedSite := normalizeWaterBodyText(in.Site)
	if rule, ok := cfg.WaterBodyRules[normalizedSite]; ok && isKnownWaterBody(rule) {
		return rule
	}
	for site, rule := range cfg.WaterBodyRules {
		if normalizeWaterBodyText(site) == normalizedSite && isKnownWaterBody(rule) {
			return rule
		}
	}

	if inferred, ok := inferWaterBody(in); ok {
		return inferred
	}
	if isKnownWaterBody(cfg.WaterBodyID) {
		return cfg.WaterBodyID
	}
	return 0
}

func inferWaterBody(in DiveInput) (int, bool) {
	text := normalizeWaterBodyText(strings.Join([]string{
		in.Site,
		in.SiteDescription,
		in.SiteNotes,
		in.SiteGeography,
		in.Tags,
		in.Notes,
	}, " "))

	matches := map[int]bool{}
	for id, keywords := range map[int][]string{
		WaterBodyOcean:  {"ocean", "sea", "morze", "ocean", "meer"},
		WaterBodyRiver:  {"river", "rzeka", "fluss", "fluss"},
		WaterBodyQuarry: {"quarry", "kamieniolom", "steinbruch"},
		WaterBodyLake:   {"lake", "jezioro", "see"},
		WaterBodyIndoor: {"indoor", "pool", "basen", "hallenbad", "deepspot"},
	} {
		for _, keyword := range keywords {
			if containsWaterBodyKeyword(text, keyword) {
				matches[id] = true
				break
			}
		}
	}
	if len(matches) != 1 {
		return 0, false
	}
	for id := range matches {
		return id, true
	}
	return 0, false
}

func containsWaterBodyKeyword(text, keyword string) bool {
	for _, word := range strings.Fields(text) {
		if word == keyword {
			return true
		}
	}
	return false
}

func normalizeWaterBodyText(raw string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			continue
		}
		switch r {
		case 'ą':
			builder.WriteByte('a')
		case 'ć':
			builder.WriteByte('c')
		case 'ę':
			builder.WriteByte('e')
		case 'ł':
			builder.WriteByte('l')
		case 'ń':
			builder.WriteByte('n')
		case 'ó':
			builder.WriteByte('o')
		case 'ś':
			builder.WriteByte('s')
		case 'ź', 'ż':
			builder.WriteByte('z')
		case 'ü':
			builder.WriteByte('u')
		case 'ä':
			builder.WriteByte('a')
		case 'ö':
			builder.WriteByte('o')
		case 'ß':
			builder.WriteString("ss")
		default:
			builder.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func isKnownWaterBody(id int) bool {
	for _, option := range WaterBodyOptions {
		if option.ID == id {
			return true
		}
	}
	return false
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
