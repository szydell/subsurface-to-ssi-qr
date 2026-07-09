package main

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	sqdialog "github.com/sqweek/dialog"

	"github.com/szydell/subsurface-to-ssi-qr/internal/buildinfo"
	"github.com/szydell/subsurface-to-ssi-qr/internal/config"
	"github.com/szydell/subsurface-to-ssi-qr/internal/qr"
	"github.com/szydell/subsurface-to-ssi-qr/internal/ssi"
	"github.com/szydell/subsurface-to-ssi-qr/internal/subsurface"
)

type diveListItem struct {
	Index        int
	WhenText     string
	DurationText string
	DepthText    string
	SiteText     string
	Payload      string
}

const prefLastDir = "last_dialog_dir"
const prefLang = "ui_lang"

func main() {
	a := app.NewWithID("pl.szydell.subsurface-to-ssi-qr")
	uiLang := normalizeLang(a.Preferences().String(prefLang))
	tr, err := newTranslator(uiLang)
	if err != nil {
		panic(err)
	}

	w := a.NewWindow(tr.text("app_title"))
	w.Resize(fyne.NewSize(920, 680))

	cfg := config.DefaultMapping()
	listItems := make([]diveListItem, 0)
	selectedDiveID := -1

	startDir := ""
	if wd, wdErr := os.Getwd(); wdErr == nil {
		startDir = wd
	}
	if saved := strings.TrimSpace(a.Preferences().String(prefLastDir)); saved != "" {
		if stat, statErr := os.Stat(saved); statErr == nil && stat.IsDir() {
			startDir = saved
		}
	}
	if strings.TrimSpace(startDir) == "" {
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			startDir = home
		}
	}

	status := widget.NewLabel(tr.text("status_prompt_load"))
	payloadBox := widget.NewMultiLineEntry()
	payloadBox.SetPlaceHolder(tr.text("payload_placeholder"))
	payloadBox.Wrapping = fyne.TextWrapWord
	payloadBox.SetMinRowsVisible(3)
	payloadBox.Disable()

	img := canvas.NewImageFromImage(nil)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(360, 360))

	diveList := widget.NewList(
		func() int {
			return len(listItems)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			row := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
			row.Wrapping = fyne.TextWrapOff
			return container.NewMax(bg, container.NewPadded(row))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			item := listItems[id]
			row := obj.(*fyne.Container)
			bg := row.Objects[0].(*canvas.Rectangle)
			content := row.Objects[1].(*fyne.Container)
			line := content.Objects[0].(*widget.Label)

			if id == selectedDiveID {
				bg.FillColor = color.NRGBA{R: 212, G: 225, B: 245, A: 255}
			} else if id%2 == 0 {
				bg.FillColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			} else {
				bg.FillColor = color.NRGBA{R: 246, G: 248, B: 251, A: 255}
			}
			bg.Refresh()

			line.SetText(formatDiveRow(item))
		},
	)
	diveList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(listItems) {
			return
		}
		selectedDiveID = id
		diveList.Refresh()

		item := listItems[id]
		payload := item.Payload
		payloadBox.SetText(payload)

		png, err := qr.PNG(payload, 420)
		if err != nil {
			status.SetText(tr.textData("status_qr_gen_error", map[string]any{"Err": err.Error()}))
			return
		}
		img.Resource = fyne.NewStaticResource("dive.png", png)
		img.Refresh()
	}

	openBtn := widget.NewButton(tr.text("btn_open_ssrf"), func() {
		path, err := sqdialog.File().
			SetStartDir(startDir).
			Filter(tr.text("filter_subsurface"), "ssrf", "xml").
			Load()
		if err != nil {
			if errors.Is(err, sqdialog.Cancelled) {
				return
			}
			status.SetText(tr.textData("status_file_select_error", map[string]any{"Err": err.Error()}))
			return
		}

		startDir = filepath.Dir(path)
		a.Preferences().SetString(prefLastDir, startDir)

		parsed, err := subsurface.ParseFile(path)
		if err != nil {
			status.SetText(tr.textData("status_parse_error", map[string]any{"Err": err.Error()}))
			return
		}

		listItems = listItems[:0]
		for i, d := range parsed {
			mapped := ssi.MapDive(d, cfg)
			payload, err := ssi.BuildPayload(mapped, cfg.IncludeUserIDs, ssi.ValidationLenient)
			if err != nil {
				continue
			}
			idx := i + 1
			listItems = append(listItems, diveListItem{
				Index:        idx,
				WhenText:     d.StartTime.Format("2006-01-02 15:04"),
				DurationText: fmt.Sprintf("%.1f min", d.DurationMin),
				DepthText:    fmt.Sprintf("%.1f m", d.MaxDepthM),
				SiteText:     normalizeSite(d.Site),
				Payload:      payload,
			})
		}

		if len(listItems) == 0 {
			status.SetText(tr.text("status_no_valid_dives"))
			selectedDiveID = -1
			payloadBox.SetText("")
			img.Resource = nil
			img.Refresh()
			diveList.UnselectAll()
			diveList.Refresh()
			return
		}

		diveList.Refresh()
		diveList.Select(0)
		status.SetText(tr.textData("status_loaded_n_dives", map[string]any{
			"Count": len(listItems),
			"File":  filepath.Base(path),
		}))
	})

	saveBtn := widget.NewButton(tr.text("btn_save_png"), func() {
		if payloadBox.Text == "" {
			status.SetText(tr.text("status_payload_first"))
			return
		}
		path, err := sqdialog.File().
			SetStartDir(startDir).
			Filter(tr.text("filter_png"), "png").
			Save()
		if err != nil {
			if errors.Is(err, sqdialog.Cancelled) {
				return
			}
			status.SetText(tr.textData("status_file_select_error", map[string]any{"Err": err.Error()}))
			return
		}

		startDir = filepath.Dir(path)
		a.Preferences().SetString(prefLastDir, startDir)

		if !strings.HasSuffix(strings.ToLower(path), ".png") {
			path += ".png"
		}
		if err := qr.WritePNG(payloadBox.Text, 420, path); err != nil {
			status.SetText(tr.textData("status_png_save_error", map[string]any{"Err": err.Error()}))
			return
		}
		status.SetText(tr.text("status_png_saved"))
	})

	listHeader := widget.NewLabelWithStyle(
		tr.text("list_header"),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Monospace: true},
	)
	listHeader.Wrapping = fyne.TextWrapOff
	divesLabel := widget.NewLabel(tr.text("label_dives"))
	payloadLabel := widget.NewLabel(tr.text("label_payload"))
	previewLabel := widget.NewLabel(tr.text("label_preview_qr"))
	langLabel := widget.NewLabel(tr.text("label_language"))
	var toolbar *fyne.Container
	var content *fyne.Container
	var langSelect *widget.Select

	langSelect = widget.NewSelect([]string{"EN", "PL", "DE"}, func(choice string) {
		code := langCode(choice)
		tr.setLanguage(code)
		a.Preferences().SetString(prefLang, code)

		w.SetTitle(tr.text("app_title"))
		openBtn.SetText(tr.text("btn_open_ssrf"))
		saveBtn.SetText(tr.text("btn_save_png"))
		divesLabel.SetText(tr.text("label_dives"))
		payloadLabel.SetText(tr.text("label_payload"))
		previewLabel.SetText(tr.text("label_preview_qr"))
		langLabel.SetText(tr.text("label_language"))
		payloadBox.SetPlaceHolder(tr.text("payload_placeholder"))
		listHeader.SetText(tr.text("list_header"))
		if len(listItems) == 0 {
			status.SetText(tr.text("status_prompt_load"))
		}

		// Force relayout after text length changes across locales.
		openBtn.Refresh()
		saveBtn.Refresh()
		langSelect.Refresh()
		if toolbar != nil {
			toolbar.Refresh()
		}
		if content != nil {
			content.Refresh()
		}
	})
	langSelect.SetSelected(langOption(uiLang))

	toolbar = container.NewHBox(openBtn, saveBtn, layout.NewSpacer(), langLabel, langSelect)
	left := container.NewBorder(
		container.NewVBox(divesLabel, listHeader),
		container.NewVBox(payloadLabel, payloadBox),
		nil,
		nil,
		diveList,
	)
	right := container.NewVBox(previewLabel, img)

	mainSplit := container.NewHSplit(left, right)
	mainSplit.Offset = 0.55
	versionText := canvas.NewText(strings.TrimSpace(buildinfo.Version), color.NRGBA{R: 124, G: 132, B: 142, A: 255})
	versionText.TextSize = 11
	footer := container.NewHBox(layout.NewSpacer(), versionText)

	content = container.NewBorder(
		container.NewVBox(toolbar, status),
		footer,
		nil,
		nil,
		mainSplit,
	)

	w.SetContent(content)
	w.ShowAndRun()
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
