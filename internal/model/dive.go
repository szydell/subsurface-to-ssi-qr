package model

import "time"

// DiveRecord is a normalized dive view extracted from Subsurface XML.
type DiveRecord struct {
	StartTime    time.Time
	DurationMin  float64
	MaxDepthM    float64
	DiveMode     string
	Site         string
	WaterTypeRaw string
	AirTempC     *float64
	WaterTempC   *float64
	VisibilityM  *float64
}
