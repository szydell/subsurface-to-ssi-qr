package config

import (
	"os"

	"github.com/szydell/subsurface-to-ssi-qr/pkg/ssi"

	"gopkg.in/yaml.v3"
)

// MappingConfig stores default SSI var_* values and user metadata.
type MappingConfig = ssi.MappingConfig

// DefaultMapping is a conservative default profile based on public reverse-engineering.
func DefaultMapping() MappingConfig {
	return ssi.DefaultMappingConfig()
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
