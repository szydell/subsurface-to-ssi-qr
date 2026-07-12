package subsurface

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/szydell/subsurface-to-ssi-qr/internal/model"
)

// ParseFile reads a Subsurface XML file and returns normalized dives.
func ParseFile(path string) ([]model.DiveRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

// Parse reads Subsurface XML from an io.Reader.
func Parse(r io.Reader) ([]model.DiveRecord, error) {
	dec := xml.NewDecoder(r)

	dives := make([]model.DiveRecord, 0)
	siteByUUID := map[string]siteMetadata{}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		switch strings.ToLower(start.Name.Local) {
		case "site":
			site, err := parseSite(dec, start)
			if err != nil {
				return nil, err
			}
			if site.UUID != "" {
				siteByUUID[site.UUID] = site
			}
			continue
		case "dive":
			// parsed below
		default:
			continue
		}

		dive, err := parseDive(dec, start, siteByUUID)
		if err != nil {
			return nil, err
		}
		if dive.StartTime.IsZero() || dive.DurationMin <= 0 || dive.MaxDepthM <= 0 {
			continue
		}
		dives = append(dives, dive)
	}

	if len(dives) == 0 {
		return nil, errors.New("no valid dives found in input")
	}
	return dives, nil
}

type siteMetadata struct {
	UUID        string
	Name        string
	Description string
	Notes       string
	Geography   string
}

func parseSite(dec *xml.Decoder, root xml.StartElement) (siteMetadata, error) {
	site := siteMetadata{}
	for _, attr := range root.Attr {
		switch strings.ToLower(attr.Name.Local) {
		case "uuid":
			site.UUID = strings.TrimSpace(attr.Value)
		case "name":
			site.Name = strings.TrimSpace(attr.Value)
		case "description":
			site.Description = strings.TrimSpace(attr.Value)
		}
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			return site, err
		}
		switch element := tok.(type) {
		case xml.StartElement:
			switch strings.ToLower(element.Name.Local) {
			case "geo":
				for _, attr := range element.Attr {
					if strings.EqualFold(attr.Name.Local, "value") {
						site.Geography = appendText(site.Geography, attr.Value)
					}
				}
			}
		case xml.CharData:
			value := strings.TrimSpace(string(element))
			if value != "" {
				// `notes` is the only free text child emitted by common SSRF exports.
				site.Notes = appendText(site.Notes, value)
			}
		case xml.EndElement:
			if strings.EqualFold(element.Name.Local, "site") {
				return site, nil
			}
		}
	}
}

func parseDive(dec *xml.Decoder, root xml.StartElement, siteByUUID map[string]siteMetadata) (model.DiveRecord, error) {
	rec := model.DiveRecord{}

	var rawDate string
	var rawTime string
	for _, a := range root.Attr {
		switch strings.ToLower(a.Name.Local) {
		case "date":
			rawDate = strings.TrimSpace(a.Value)
		case "time":
			rawTime = strings.TrimSpace(a.Value)
		case "duration", "divetime":
			if v, ok := parseDurationMin(a.Value); ok {
				rec.DurationMin = v
			}
		case "maxdepth", "depth":
			if v, ok := parseDepthM(a.Value); ok {
				rec.MaxDepthM = v
			}
		case "dive_mode", "mode":
			rec.DiveMode = strings.TrimSpace(a.Value)
		case "watertype", "water_type":
			rec.WaterTypeRaw = strings.TrimSpace(a.Value)
		case "divesiteid":
			uuid := strings.TrimSpace(a.Value)
			if site, ok := siteByUUID[uuid]; ok {
				rec.Site = site.Name
				rec.SiteDescription = site.Description
				rec.SiteNotes = site.Notes
				rec.SiteGeography = site.Geography
			}
		case "tags":
			rec.Tags = strings.TrimSpace(a.Value)
		}
	}

	if dt, ok := parseDateTime(rawDate, rawTime); ok {
		rec.StartTime = dt
	}

	var elementPath []string
	for {
		tok, err := dec.Token()
		if err != nil {
			return rec, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			elementPath = append(elementPath, name)

			if name == "depth" {
				for _, a := range t.Attr {
					if strings.ToLower(a.Name.Local) == "max" {
						if v, ok := parseDepthM(a.Value); ok {
							rec.MaxDepthM = v
						}
					}
				}
			}

			// Some Subsurface variants put values in attributes.
			if name == "temperature" || name == "temps" {
				for _, a := range t.Attr {
					switch strings.ToLower(a.Name.Local) {
					case "air":
						if v, ok := parseTemperatureC(a.Value); ok {
							rec.AirTempC = ptrFloat(v)
						}
					case "water":
						if v, ok := parseTemperatureC(a.Value); ok {
							rec.WaterTempC = ptrFloat(v)
						}
					}
				}
			}
			if name == "visibility" {
				for _, a := range t.Attr {
					if v, ok := parseDistanceM(a.Value); ok {
						rec.VisibilityM = ptrFloat(v)
					}
				}
			}
		case xml.CharData:
			if len(elementPath) == 0 {
				continue
			}
			current := elementPath[len(elementPath)-1]
			value := strings.TrimSpace(string(t))
			if value == "" {
				continue
			}
			switch current {
			case "location", "dive_site", "site", "place":
				rec.Site = value
			case "notes":
				rec.Notes = appendText(rec.Notes, value)
			case "duration", "divetime":
				if v, ok := parseDurationMin(value); ok {
					rec.DurationMin = v
				}
			case "maxdepth", "depth":
				if v, ok := parseDepthM(value); ok {
					rec.MaxDepthM = v
				}
			case "dive_mode", "mode":
				rec.DiveMode = value
			case "watertype", "water_type":
				rec.WaterTypeRaw = value
			case "when", "datetime":
				if dt, ok := parseWhen(value); ok {
					rec.StartTime = dt
				}
			case "date":
				rawDate = value
				if dt, ok := parseDateTime(rawDate, rawTime); ok {
					rec.StartTime = dt
				}
			case "time":
				rawTime = value
				if dt, ok := parseDateTime(rawDate, rawTime); ok {
					rec.StartTime = dt
				}
			case "airtemp", "air_temperature":
				if v, ok := parseTemperatureC(value); ok {
					rec.AirTempC = ptrFloat(v)
				}
			case "watertemp", "water_temperature":
				if v, ok := parseTemperatureC(value); ok {
					rec.WaterTempC = ptrFloat(v)
				}
			case "visibility":
				if v, ok := parseDistanceM(value); ok {
					rec.VisibilityM = ptrFloat(v)
				}
			}
		case xml.EndElement:
			name := strings.ToLower(t.Name.Local)
			if name == "dive" {
				return rec, nil
			}
			if len(elementPath) > 0 {
				elementPath = elementPath[:len(elementPath)-1]
			}
		}
	}
}

func appendText(current, addition string) string {
	addition = strings.TrimSpace(addition)
	if addition == "" {
		return current
	}
	if strings.TrimSpace(current) == "" {
		return addition
	}
	return current + " " + addition
}

func parseWhen(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseDateTime(dateRaw, timeRaw string) (time.Time, bool) {
	if strings.TrimSpace(dateRaw) == "" {
		return time.Time{}, false
	}
	if strings.TrimSpace(timeRaw) == "" {
		timeRaw = "00:00"
	}
	joined := strings.TrimSpace(dateRaw) + " " + strings.TrimSpace(timeRaw)
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, joined); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseDurationMin(v string) (float64, bool) {
	raw := strings.TrimSpace(strings.ToLower(v))
	if raw == "" {
		return 0, false
	}

	raw = strings.ReplaceAll(raw, "min", "")
	raw = strings.TrimSpace(raw)

	if strings.HasSuffix(raw, "s") {
		n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(raw, "s")), 64)
		if err == nil {
			return n / 60.0, true
		}
	}

	parts := strings.Split(raw, ":")
	if len(parts) == 2 {
		m, errM := strconv.ParseFloat(parts[0], 64)
		s, errS := strconv.ParseFloat(parts[1], 64)
		if errM == nil && errS == nil {
			return m + s/60.0, true
		}
	}
	if len(parts) == 3 {
		h, errH := strconv.ParseFloat(parts[0], 64)
		m, errM := strconv.ParseFloat(parts[1], 64)
		s, errS := strconv.ParseFloat(parts[2], 64)
		if errH == nil && errM == nil && errS == nil {
			return h*60.0 + m + s/60.0, true
		}
	}

	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func parseDepthM(v string) (float64, bool) {
	raw := strings.TrimSpace(strings.ToLower(v))
	if raw == "" {
		return 0, false
	}

	if strings.Contains(raw, "ft") {
		n, ok := parseNumber(raw)
		if !ok {
			return 0, false
		}
		return round1(n * 0.3048), true
	}
	n, ok := parseNumber(raw)
	if !ok {
		return 0, false
	}
	return round1(n), true
}

func parseTemperatureC(v string) (float64, bool) {
	raw := strings.TrimSpace(strings.ToLower(v))
	if raw == "" {
		return 0, false
	}
	n, ok := parseNumber(raw)
	if !ok {
		return 0, false
	}
	if strings.Contains(raw, "f") {
		return round1((n - 32.0) * 5.0 / 9.0), true
	}
	return round1(n), true
}

func parseDistanceM(v string) (float64, bool) {
	raw := strings.TrimSpace(strings.ToLower(v))
	if raw == "" {
		return 0, false
	}
	n, ok := parseNumber(raw)
	if !ok {
		return 0, false
	}
	if strings.Contains(raw, "ft") {
		return round1(n * 0.3048), true
	}
	return round1(n), true
}

func parseNumber(v string) (float64, bool) {
	clean := strings.TrimSpace(v)
	clean = strings.ReplaceAll(clean, ",", ".")
	for _, token := range []string{"m", "ft", "c", "f", "bar", "psi"} {
		clean = strings.ReplaceAll(clean, token, "")
	}
	clean = strings.TrimSpace(clean)
	n, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func round1(v float64) float64 {
	return math.Round(v*10.0) / 10.0
}

func ptrFloat(v float64) *float64 {
	return &v
}

// MustParseDateTime is a helper intended for tests.
func MustParseDateTime(dateRaw, timeRaw string) time.Time {
	t, ok := parseDateTime(dateRaw, timeRaw)
	if !ok {
		panic(fmt.Sprintf("failed to parse datetime from %q %q", dateRaw, timeRaw))
	}
	return t
}
