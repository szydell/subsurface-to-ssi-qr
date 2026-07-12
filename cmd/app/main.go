package main

import (
	"bytes"
	"image"
	"image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"golang.org/x/image/draw"

	"github.com/szydell/subsurface-to-ssi-qr/internal/assets"
)

// maxAppIconSize caps the window icon's largest dimension. X11 sends the
// window icon as raw pixel data via the _NET_WM_ICON property in a single
// X_ChangeProperty request; very large icons (e.g. 2048x2048, as shipped for
// desktop icon themes) can exceed the X server's max request length and
// crash the app on startup with "X Error ... BadLength ... X_ChangeProperty".
const maxAppIconSize = 256

const prefLastDir = "last_dialog_dir"
const prefLang = "ui_lang"

func main() {
	a := app.NewWithID("pl.szydell.subsurface-to-ssi-qr")
	if appIcon := loadAppIcon(); appIcon != nil {
		a.SetIcon(appIcon)
	}

	uiLang := normalizeLang(a.Preferences().String(prefLang))
	tr, err := newTranslator(uiLang)
	if err != nil {
		panic(err)
	}

	s := newAppState(a, tr)
	s.buildUI()
	s.win.ShowAndRun()
}

// loadAppIcon returns the application's window/taskbar icon. It prefers the
// system-installed pixmap (used on Linux when packaged, so desktop icon
// theming/overrides are respected), but always falls back to the icon
// embedded in the binary at build time (internal/assets). This guarantees
// the icon is available on every platform regardless of the current working
// directory or install layout, which matters most on Windows where the GUI
// is often run as a standalone portable .exe with no "assets" directory
// alongside it.
func loadAppIcon() fyne.Resource {
	name := "subsurface-to-ssi-qr.png"
	pixmap, _ := os.ReadFile("/usr/share/pixmaps/subsurface-to-ssi-qr.png")
	data, err := preferredIconPNG(assets.IconPNG, pixmap)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(name, data)
}

// preferredIconPNG normalizes the embedded icon and, when valid, uses the
// system pixmap in preference to it.
func preferredIconPNG(embedded, pixmap []byte) ([]byte, error) {
	normalizedEmbedded, err := normalizedIconPNG(embedded)
	if err != nil {
		return nil, err
	}
	if normalizedPixmap, err := normalizedIconPNG(pixmap); err == nil {
		return normalizedPixmap, nil
	}
	return normalizedEmbedded, nil
}

// normalizedIconPNG validates a PNG image and caps its largest dimension.
func normalizedIconPNG(data []byte) ([]byte, error) {
	return resizeIconPNG(data, maxAppIconSize)
}

// resizeIconPNG decodes a PNG image and, if either dimension exceeds
// maxSize, scales it down to fit within maxSize x maxSize, returning the
// re-encoded PNG bytes.
func resizeIconPNG(data []byte, maxSize int) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= maxSize && h <= maxSize {
		return data, nil
	}

	scale := float64(maxSize) / float64(w)
	if hScale := float64(maxSize) / float64(h); hScale < scale {
		scale = hScale
	}
	newW := int(float64(w) * scale)
	newH := int(float64(h) * scale)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
