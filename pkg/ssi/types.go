package ssi

import "time"

const (
	DiveTypeScuba                   = 0
	DiveTypeExtendedRange           = 2
	DiveTypeRebreatherSelfContained = 4
	DiveTypeFreediving              = 6
	DiveTypeRebreatherClosedCircuit = 8
)

// ValidationMode controls how strict required-field checks are.
type ValidationMode string

const (
	ValidationLenient ValidationMode = "lenient"
	ValidationStrict  ValidationMode = "strict"
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

// DiveInput is a normalized dive representation expected by SSI mapping.
type DiveInput struct {
	StartTime    time.Time
	DurationMin  float64
	MaxDepthM    float64
	DiveMode     string
	WaterTypeRaw string
	AirTempC     *float64
	WaterTempC   *float64
	VisibilityM  *float64
}

// MappingConfig stores default SSI var_* values and optional user metadata.
type MappingConfig struct {
	SiteID         int    `yaml:"site_id"`
	WeatherID      int    `yaml:"var_weather_id"`
	EntryID        int    `yaml:"var_entry_id"`
	WaterBodyID    int    `yaml:"var_water_body_id"`
	WaterTypeID    int    `yaml:"var_watertype_id"`
	CurrentID      int    `yaml:"var_current_id"`
	SurfaceID      int    `yaml:"var_surface_id"`
	DiveSubtypeID  int    `yaml:"var_divetype_id"`
	UserMasterID   int    `yaml:"user_master_id"`
	UserFirstName  string `yaml:"user_firstname"`
	UserLastName   string `yaml:"user_lastname"`
	UserLeaderID   int    `yaml:"user_leader_id"`
	IncludeUserIDs bool   `yaml:"include_user_ids"`
}

// DefaultMappingConfig is a conservative default profile based on public reverse-engineering.
func DefaultMappingConfig() MappingConfig {
	return MappingConfig{
		SiteID:         0,
		WeatherID:      2,
		EntryID:        21,
		WaterBodyID:    15,
		WaterTypeID:    5,
		CurrentID:      6,
		SurfaceID:      10,
		DiveSubtypeID:  24,
		UserMasterID:   0,
		UserFirstName:  "",
		UserLastName:   "",
		UserLeaderID:   0,
		IncludeUserIDs: false,
	}
}
