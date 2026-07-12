package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2/widget"

	"github.com/szydell/subsurface-to-ssi-qr/internal/config"
	"github.com/szydell/subsurface-to-ssi-qr/internal/model"
	"github.com/szydell/subsurface-to-ssi-qr/internal/ssi"
	"github.com/szydell/subsurface-to-ssi-qr/internal/subsurface"
)

// diveListItem is a display-ready row derived from a parsed dive record.
type diveListItem struct {
	Index        int
	WhenText     string
	DurationText string
	DepthText    string
	SiteText     string
	Payload      string
}

// loadDivesFromFile parses a Subsurface file and maps its dives into
// display-ready list items using the given mapping config.
func loadDivesFromFile(path string, cfg config.MappingConfig) ([]diveListItem, error) {
	parsed, err := subsurface.ParseFile(path)
	if err != nil {
		return nil, err
	}
	return mapDivesToItems(parsed, cfg), nil
}

// mapDivesToItems converts parsed dive records into diveListItems, building
// the SSI QR payload for each dive. Dives whose payload fails to build are
// skipped.
func mapDivesToItems(parsed []model.DiveRecord, cfg config.MappingConfig) []diveListItem {
	items := make([]diveListItem, 0, len(parsed))
	for i, d := range parsed {
		mapped := ssi.MapDive(d, cfg)
		payload, err := ssi.BuildPayload(mapped, cfg.IncludeUserIDs, ssi.ValidationLenient)
		if err != nil {
			continue
		}
		items = append(items, diveListItem{
			Index:        i + 1,
			WhenText:     d.StartTime.Format("2006-01-02 15:04"),
			DurationText: fmt.Sprintf("%.1f min", d.DurationMin),
			DepthText:    fmt.Sprintf("%.1f m", d.MaxDepthM),
			SiteText:     normalizeSite(d.Site),
			Payload:      payload,
		})
	}
	return items
}

func normalizeSite(site string) string {
	site = strings.TrimSpace(site)
	if site == "" {
		site = "-"
	}
	return site
}

func formatDiveRow(item diveListItem) string {
	return fmt.Sprintf(
		"%-3d %-17s %-10s %-8s %s",
		item.Index,
		item.WhenText,
		item.DurationText,
		item.DepthText,
		item.SiteText,
	)
}

// diveRowColor returns the background color for a dive list row, based on
// its selection state and alternating zebra striping.
func diveRowColor(id widget.ListItemID, selected bool) color.NRGBA {
	switch {
	case selected:
		return color.NRGBA{R: 212, G: 225, B: 245, A: 255}
	case id%2 == 0:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	default:
		return color.NRGBA{R: 246, G: 248, B: 251, A: 255}
	}
}

// ensurePNGExtension appends a ".png" suffix if path doesn't already have
// one (case-insensitive).
func ensurePNGExtension(path string) string {
	if strings.HasSuffix(strings.ToLower(path), ".png") {
		return path
	}
	return path + ".png"
}
