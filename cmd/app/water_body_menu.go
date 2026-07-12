package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/szydell/subsurface-to-ssi-qr/internal/ssi"
)

func isKnownWaterBodyID(id int) bool {
	for _, option := range ssi.WaterBodyOptions {
		if option.ID == id {
			return true
		}
	}
	return false
}

// waterBodyOptionLabel returns the localized label for a water-body choice.
// id == 0 means "no explicit category" (shown as "Automatic" in menus that
// offer automatic resolution, otherwise as "No field"). id < 0 means the
// dive explicitly omits the SSI field.
func (s *appState) waterBodyOptionLabel(id int, automatic bool) string {
	if id == 0 {
		if automatic {
			return s.tr.text("water_body_automatic")
		}
		return s.tr.text("water_body_none")
	}
	if id < 0 {
		return s.tr.text("water_body_none")
	}
	for _, option := range ssi.WaterBodyOptions {
		if option.ID == id {
			return s.tr.text("water_body_" + option.Key)
		}
	}
	return s.tr.text("water_body_automatic")
}

// waterBodyColumnLabel is the compact text shown in the dive list's water
// body column; it is blank when no category is resolved for the dive.
func (s *appState) waterBodyColumnLabel(id int) string {
	if id <= 0 {
		return ""
	}
	return s.waterBodyOptionLabel(id, false)
}

// waterBodyChoiceLabels lists every selectable option in the right-click
// water-body picker: automatic resolution, explicitly no field, then every
// known SSI water-body category.
func (s *appState) waterBodyChoiceLabels() []string {
	labels := []string{s.waterBodyOptionLabel(0, true), s.waterBodyOptionLabel(-1, true)}
	for _, option := range ssi.WaterBodyOptions {
		labels = append(labels, s.waterBodyOptionLabel(option.ID, true))
	}
	return labels
}

func (s *appState) waterBodyIDForLabel(label string) (int, bool) {
	ids := []int{0, -1}
	for _, option := range ssi.WaterBodyOptions {
		ids = append(ids, option.ID)
	}
	for _, id := range ids {
		if label == s.waterBodyOptionLabel(id, true) {
			return id, true
		}
	}
	return 0, false
}

// sameSite reports whether two Subsurface site names refer to the same
// place, for the purpose of bulk-applying a water-body choice.
func sameSite(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// showWaterBodyMenu opens a small popup at pos letting the user assign a
// water-body category to the dive at rowID, optionally applying the same
// choice to every other dive in the current import that shares its site
// name.
func (s *appState) showWaterBodyMenu(rowID widget.ListItemID, pos fyne.Position) {
	if rowID < 0 || rowID >= len(s.parsedDives) {
		return
	}

	applyAll := widget.NewCheck(s.tr.text("water_body_apply_all"), nil)

	var popUp *widget.PopUp
	choices := widget.NewRadioGroup(s.waterBodyChoiceLabels(), func(label string) {
		if id, ok := s.waterBodyIDForLabel(label); ok {
			s.applyWaterBodyChoice(rowID, id, applyAll.Checked)
		}
		if popUp != nil {
			popUp.Hide()
		}
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle(s.tr.text("water_body_menu_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		applyAll,
		widget.NewSeparator(),
		choices,
	)

	popUp = widget.NewPopUp(content, s.win.Canvas())
	popUp.ShowAtPosition(pos)
}

// applyWaterBodyChoice records a water-body choice for the dive at rowID.
// When applyAll is set, every other dive in the current import that shares
// the same site name receives the same choice.
func (s *appState) applyWaterBodyChoice(rowID widget.ListItemID, waterBodyID int, applyAll bool) {
	setOverride := func(idx int) {
		if waterBodyID == 0 {
			delete(s.waterBodyOverrides, idx)
		} else {
			s.waterBodyOverrides[idx] = waterBodyID
		}
	}

	setOverride(rowID)
	if applyAll {
		site := s.parsedDives[rowID].Site
		for idx, dive := range s.parsedDives {
			if idx != rowID && sameSite(dive.Site, site) {
				setOverride(idx)
			}
		}
	}
	s.refreshDiveItems()
}
