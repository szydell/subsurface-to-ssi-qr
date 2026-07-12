package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

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

	// Raw sort keys, kept alongside the formatted *Text fields above so
	// column sorting can compare actual values instead of locale-formatted
	// strings.
	WhenTime    time.Time
	DurationMin float64
	DepthM      float64
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
			WhenTime:     d.StartTime,
			DurationMin:  d.DurationMin,
			DepthM:       d.MaxDepthM,
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

// diveTableColumnCount is the number of columns rendered in the dive table:
// index, date, duration, depth, water body, site.
const diveTableColumnCount = 6

// diveCellText returns the display text for a single dive table cell.
// waterBodyLabel is passed in separately since it is localized and
// diveListItem only stores the numeric water-body ID.
func diveCellText(item diveListItem, col int, waterBodyLabel string) string {
	switch col {
	case 0:
		return fmt.Sprintf("%d", item.Index)
	case 1:
		return item.WhenText
	case 2:
		return item.DurationText
	case 3:
		return item.DepthText
	case 4:
		return waterBodyLabel
	case 5:
		return item.SiteText
	default:
		return ""
	}
}

// diveCell is a single dive table cell widget. Fyne's hit-testing stops at
// the deepest object implementing any of Tappable/SecondaryTappable/etc., so
// once this widget catches right-clicks (TappedSecondary) it must also
// handle left-clicks (Tapped) itself; otherwise the Table's own cell
// selection would never receive them. Tapped forwards to row selection.
type diveCell struct {
	widget.BaseWidget
	bg             *canvas.Rectangle
	line           *widget.Label
	onTap          func()
	onSecondaryTap func(pos fyne.Position)
}

func newDiveCell() *diveCell {
	cell := &diveCell{
		bg:   canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		line: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true}),
	}
	cell.line.Wrapping = fyne.TextWrapOff
	cell.ExtendBaseWidget(cell)
	return cell
}

func (c *diveCell) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewMax(c.bg, container.NewPadded(c.line)))
}

// Tapped selects the row, taking over the Table's own left-click handling
// (see the type doc comment for why this is necessary).
func (c *diveCell) Tapped(*fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

// TappedSecondary opens the water-body picker at the right-click position.
func (c *diveCell) TappedSecondary(ev *fyne.PointEvent) {
	if c.onSecondaryTap != nil {
		c.onSecondaryTap(ev.AbsolutePosition)
	}
}

// sortIndicator returns the arrow suffix appended to a sortable column's
// header title, or "" when that column isn't the current sort column.
func sortIndicator(col, sortColumn int, ascending bool) string {
	if col != sortColumn {
		return ""
	}
	if ascending {
		return " \u25b2"
	}
	return " \u25bc"
}

// diveHeaderCell is a clickable dive table header cell. Clicking it triggers
// column sorting; unlike diveCell it only needs a left-click handler.
type diveHeaderCell struct {
	widget.BaseWidget
	line  *widget.Label
	onTap func()
}

func newDiveHeaderCell() *diveHeaderCell {
	cell := &diveHeaderCell{
		line: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	cell.line.Wrapping = fyne.TextWrapOff
	cell.ExtendBaseWidget(cell)
	return cell
}

func (c *diveHeaderCell) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewPadded(c.line))
}

// Tapped triggers this column's sort toggle.
func (c *diveHeaderCell) Tapped(*fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

// diveRowColor returns the background color for a dive table row, based on
// its selection state and alternating zebra striping.
func diveRowColor(id int, selected bool) color.NRGBA {
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
