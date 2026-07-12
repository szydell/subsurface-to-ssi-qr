package main

import (
	"errors"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	sqdialog "github.com/sqweek/dialog"

	"github.com/szydell/subsurface-to-ssi-qr/internal/buildinfo"
	"github.com/szydell/subsurface-to-ssi-qr/internal/config"
	"github.com/szydell/subsurface-to-ssi-qr/internal/qr"
)

// appState holds all mutable UI state and widgets for the main window, and
// provides the event handlers that wire them together.
type appState struct {
	fyneApp fyne.App
	win     fyne.Window
	tr      *translator

	cfg            config.MappingConfig
	listItems      []diveListItem
	selectedDiveID int
	loadedFileName string
	startDir       string

	status       *widget.Label
	payloadBox   *widget.Entry
	img          *canvas.Image
	diveList     *widget.List
	openBtn      *widget.Button
	saveBtn      *widget.Button
	listHeader   *widget.Label
	divesLabel   *widget.Label
	payloadLabel *widget.Label
	qrLabel      *widget.Label
	langLabel    *widget.Label
	langSelect   *widget.Select
	toolbar      *fyne.Container
	content      *fyne.Container
}

// resolveStartDir picks the initial directory for file dialogs: the saved
// preference (if it still exists and is a directory), otherwise the current
// working directory, falling back to the user's home directory.
func resolveStartDir(savedDir string) string {
	startDir := ""
	if wd, err := os.Getwd(); err == nil {
		startDir = wd
	}
	if saved := strings.TrimSpace(savedDir); saved != "" {
		if stat, err := os.Stat(saved); err == nil && stat.IsDir() {
			startDir = saved
		}
	}
	if strings.TrimSpace(startDir) == "" {
		if home, err := os.UserHomeDir(); err == nil {
			startDir = home
		}
	}
	return startDir
}

// newAppState creates the main window and initial state, ready for buildUI.
func newAppState(a fyne.App, tr *translator) *appState {
	w := a.NewWindow(tr.text("app_title"))
	w.Resize(fyne.NewSize(920, 680))

	return &appState{
		fyneApp:        a,
		win:            w,
		tr:             tr,
		cfg:            config.DefaultMapping(),
		listItems:      make([]diveListItem, 0),
		selectedDiveID: -1,
		startDir:       resolveStartDir(a.Preferences().String(prefLastDir)),
	}
}

// buildUI creates all widgets, wires up event handlers and sets the window
// content.
func (s *appState) buildUI() {
	s.status = widget.NewLabel(s.tr.text("status_prompt_load"))

	s.payloadBox = widget.NewMultiLineEntry()
	s.payloadBox.SetPlaceHolder(s.tr.text("payload_placeholder"))
	s.payloadBox.Wrapping = fyne.TextWrapWord
	s.payloadBox.SetMinRowsVisible(3)
	s.payloadBox.Disable()

	s.img = canvas.NewImageFromImage(nil)
	s.img.FillMode = canvas.ImageFillContain
	s.img.SetMinSize(fyne.NewSize(360, 360))

	s.diveList = newDiveList(s)
	s.diveList.OnSelected = s.onDiveSelected

	s.openBtn = widget.NewButton(s.tr.text("btn_open_ssrf"), s.onOpenClicked)
	s.saveBtn = widget.NewButton(s.tr.text("btn_save_png"), s.onSaveClicked)

	s.listHeader = widget.NewLabelWithStyle(
		s.tr.text("list_header"),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Monospace: true},
	)
	s.listHeader.Wrapping = fyne.TextWrapOff
	s.divesLabel = widget.NewLabel(s.tr.text("label_dives"))
	s.payloadLabel = widget.NewLabel(s.tr.text("label_payload"))
	s.qrLabel = widget.NewLabel(s.tr.text("label_ssi_qr"))
	s.langLabel = widget.NewLabel(s.tr.text("label_language"))

	s.langSelect = widget.NewSelect([]string{"EN", "PL", "DE"}, s.onLangSelected)
	s.langSelect.SetSelected(langOption(s.tr.lang))

	s.buildLayout()
	s.win.SetContent(s.content)
}

// newDiveList builds the dive list widget backed by s.listItems.
func newDiveList(s *appState) *widget.List {
	return widget.NewList(
		func() int { return len(s.listItems) },
		newDiveListRowTemplate,
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			s.updateDiveListRow(id, obj)
		},
	)
}

func newDiveListRowTemplate() fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	row := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	row.Wrapping = fyne.TextWrapOff
	return container.NewMax(bg, container.NewPadded(row))
}

func (s *appState) updateDiveListRow(id widget.ListItemID, obj fyne.CanvasObject) {
	item := s.listItems[id]
	row := obj.(*fyne.Container)
	bg := row.Objects[0].(*canvas.Rectangle)
	content := row.Objects[1].(*fyne.Container)
	line := content.Objects[0].(*widget.Label)

	bg.FillColor = diveRowColor(id, id == s.selectedDiveID)
	bg.Refresh()
	line.SetText(formatDiveRow(item))
}

func (s *appState) onDiveSelected(id widget.ListItemID) {
	if id < 0 || id >= len(s.listItems) {
		return
	}
	s.selectedDiveID = id
	s.diveList.Refresh()

	item := s.listItems[id]
	s.payloadBox.SetText(item.Payload)

	png, err := qr.PNG(item.Payload, 420)
	if err != nil {
		s.status.SetText(s.tr.textData("status_qr_gen_error", map[string]any{"Err": err.Error()}))
		return
	}
	s.img.Resource = fyne.NewStaticResource("dive.png", png)
	s.img.Refresh()
}

func (s *appState) onOpenClicked() {
	path, err := sqdialog.File().
		SetStartDir(s.startDir).
		Filter(s.tr.text("filter_subsurface"), "ssrf", "xml").
		Load()
	if err != nil {
		if errors.Is(err, sqdialog.Cancelled) {
			return
		}
		s.status.SetText(s.tr.textData("status_file_select_error", map[string]any{"Err": err.Error()}))
		return
	}

	s.startDir = filepath.Dir(path)
	s.fyneApp.Preferences().SetString(prefLastDir, s.startDir)

	items, err := loadDivesFromFile(path, s.cfg)
	if err != nil {
		s.status.SetText(s.tr.textData("status_parse_error", map[string]any{"Err": err.Error()}))
		return
	}
	s.listItems = items

	if len(s.listItems) == 0 {
		s.loadedFileName = ""
		s.showNoValidDives()
		return
	}

	s.diveList.Refresh()
	s.diveList.Select(0)
	s.loadedFileName = filepath.Base(path)
	s.status.SetText(s.statusLoadedText())
}

func (s *appState) showNoValidDives() {
	s.status.SetText(s.tr.text("status_no_valid_dives"))
	s.selectedDiveID = -1
	s.payloadBox.SetText("")
	s.img.Resource = nil
	s.img.Refresh()
	s.diveList.UnselectAll()
	s.diveList.Refresh()
}

func (s *appState) onSaveClicked() {
	if s.payloadBox.Text == "" {
		s.status.SetText(s.tr.text("status_payload_first"))
		return
	}
	path, err := sqdialog.File().
		SetStartDir(s.startDir).
		Filter(s.tr.text("filter_png"), "png").
		Save()
	if err != nil {
		if errors.Is(err, sqdialog.Cancelled) {
			return
		}
		s.status.SetText(s.tr.textData("status_file_select_error", map[string]any{"Err": err.Error()}))
		return
	}

	s.startDir = filepath.Dir(path)
	s.fyneApp.Preferences().SetString(prefLastDir, s.startDir)

	path = ensurePNGExtension(path)
	if err := qr.WritePNG(s.payloadBox.Text, 420, path); err != nil {
		s.status.SetText(s.tr.textData("status_png_save_error", map[string]any{"Err": err.Error()}))
		return
	}
	s.status.SetText(s.tr.text("status_png_saved"))
}

func (s *appState) onLangSelected(choice string) {
	code := langCode(choice)
	s.tr.setLanguage(code)
	s.fyneApp.Preferences().SetString(prefLang, code)

	s.win.SetTitle(s.tr.text("app_title"))
	s.openBtn.SetText(s.tr.text("btn_open_ssrf"))
	s.saveBtn.SetText(s.tr.text("btn_save_png"))
	s.divesLabel.SetText(s.tr.text("label_dives"))
	s.payloadLabel.SetText(s.tr.text("label_payload"))
	s.qrLabel.SetText(s.tr.text("label_ssi_qr"))
	s.langLabel.SetText(s.tr.text("label_language"))
	s.payloadBox.SetPlaceHolder(s.tr.text("payload_placeholder"))
	s.listHeader.SetText(s.tr.text("list_header"))
	s.refreshStatusText()

	// Force relayout after text length changes across locales.
	s.openBtn.Refresh()
	s.saveBtn.Refresh()
	s.langSelect.Refresh()
	if s.toolbar != nil {
		s.toolbar.Refresh()
	}
	if s.content != nil {
		s.content.Refresh()
	}
}

func (s *appState) refreshStatusText() {
	if s.loadedFileName != "" {
		s.status.SetText(s.statusLoadedText())
	} else if len(s.listItems) == 0 {
		s.status.SetText(s.tr.text("status_prompt_load"))
	}
}

func (s *appState) statusLoadedText() string {
	return s.tr.textCount("status_loaded_n_dives", len(s.listItems), map[string]any{
		"File": s.loadedFileName,
	})
}

func (s *appState) buildLayout() {
	s.toolbar = container.NewHBox(s.openBtn, s.saveBtn, layout.NewSpacer(), s.langLabel, s.langSelect)
	left := container.NewBorder(
		container.NewVBox(s.divesLabel, s.listHeader),
		container.NewVBox(s.payloadLabel, s.payloadBox),
		nil,
		nil,
		s.diveList,
	)
	right := container.NewVBox(s.qrLabel, s.img)

	mainSplit := container.NewHSplit(left, right)
	mainSplit.Offset = 0.55
	versionText := canvas.NewText(strings.TrimSpace(buildinfo.Version), color.NRGBA{R: 124, G: 132, B: 142, A: 255})
	versionText.TextSize = 11
	footer := container.NewHBox(layout.NewSpacer(), versionText)

	s.content = container.NewBorder(
		container.NewVBox(s.toolbar, s.status),
		footer,
		nil,
		nil,
		mainSplit,
	)
}
