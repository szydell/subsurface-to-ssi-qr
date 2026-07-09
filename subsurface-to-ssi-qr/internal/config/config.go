package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// MappingConfig stores default SSI var_* values and user metadata.
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

// DefaultMapping is a conservative default profile based on public reverse-engineering.
func DefaultMapping() MappingConfig {
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

// LoadMappingConfig loads YAML mapping config from disk.
func LoadMappingConfig(path string) (MappingConfig, error) {
	cfg := DefaultMapping()
	bytes, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
