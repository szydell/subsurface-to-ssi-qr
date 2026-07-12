package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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
	WaterBodyID  int
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
	return mapDivesToItemsWithOverrides(parsed, cfg, nil)
}

// mapDivesToItemsWithOverrides applies temporary per-import water-body choices.
// A negative override explicitly omits the SSI field; zero uses automatic mapping.
func mapDivesToItemsWithOverrides(parsed []model.DiveRecord, cfg config.MappingConfig, overrides map[int]int) []diveListItem {
	items := make([]diveListItem, 0, len(parsed))
	for i, d := range parsed {
		if override, ok := overrides[i]; ok {
			d.WaterBodyOverride = override
		}
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
			WaterBodyID:  mapped.VarWaterBody,
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

// formatDiveRow renders a dive's fixed-width columns: index, date, duration,
// depth, water body, then site. waterBodyLabel is passed in separately
// since it is localized and diveListItem only stores the numeric ID.
func formatDiveRow(item diveListItem, waterBodyLabel string) string {
	return fmt.Sprintf(
		"%-3d %-17s %-10s %-8s %-10s %s",
		item.Index,
		item.WhenText,
		item.DurationText,
		item.DepthText,
		waterBodyLabel,
		item.SiteText,
	)
}

// diveRow is a dive list row widget. Fyne's hit-testing stops at the
// deepest object implementing any of Tappable/SecondaryTappable/etc., so
// once this widget catches right-clicks (TappedSecondary) it must also
// handle left-clicks (Tapped) itself; otherwise the List's own row selection
// would never receive them. Tapped forwards to the List's own selection.
type diveRow struct {
	widget.BaseWidget
	bg             *canvas.Rectangle
	line           *widget.Label
	onTap          func()
	onSecondaryTap func(pos fyne.Position)
}

func newDiveRow() *diveRow {
	row := &diveRow{
		bg:   canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		line: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true}),
	}
	row.line.Wrapping = fyne.TextWrapOff
	row.ExtendBaseWidget(row)
	return row
}

func (r *diveRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewMax(r.bg, container.NewPadded(r.line)))
}

// Tapped selects the row, taking over the List's own left-click handling
// (see the type doc comment for why this is necessary).
func (r *diveRow) Tapped(*fyne.PointEvent) {
	if r.onTap != nil {
		r.onTap()
	}
}

// TappedSecondary opens the water-body picker at the right-click position.
func (r *diveRow) TappedSecondary(ev *fyne.PointEvent) {
	if r.onSecondaryTap != nil {
		r.onSecondaryTap(ev.AbsolutePosition)
	}
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
