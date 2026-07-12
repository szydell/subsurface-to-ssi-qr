package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestResizeIconPNG_SmallImageIsUnchanged(t *testing.T) {
	data := testPNG(t, 64, 32)

	got, err := resizeIconPNG(data, maxAppIconSize)
	if err != nil {
		t.Fatalf("resizeIconPNG: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Error("resizeIconPNG changed an image that already fits the size limit")
	}
}

func TestResizeIconPNG_CapsDimensionsAndPreservesAspectRatio(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		wantWidth  int
		wantHeight int
	}{
		{name: "landscape", width: 1024, height: 512, wantWidth: 256, wantHeight: 128},
		{name: "portrait", width: 512, height: 1024, wantWidth: 128, wantHeight: 256},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resizeIconPNG(testPNG(t, test.width, test.height), maxAppIconSize)
			if err != nil {
				t.Fatalf("resizeIconPNG: %v", err)
			}

			img, err := png.Decode(bytes.NewReader(got))
			if err != nil {
				t.Fatalf("decode resized PNG: %v", err)
			}
			if width, height := img.Bounds().Dx(), img.Bounds().Dy(); width != test.wantWidth || height != test.wantHeight {
				t.Errorf("resized dimensions = %dx%d, want %dx%d", width, height, test.wantWidth, test.wantHeight)
			}
		})
	}
}

func TestResizeIconPNG_InvalidDataReturnsError(t *testing.T) {
	if _, err := resizeIconPNG([]byte("not a PNG"), maxAppIconSize); err == nil {
		t.Fatal("resizeIconPNG accepted invalid image data")
	}
}

func TestPreferredIconPNG_InvalidSystemPixmapUsesEmbeddedIcon(t *testing.T) {
	embedded := testPNG(t, 64, 64)

	got, err := preferredIconPNG(embedded, []byte("not a PNG"))
	if err != nil {
		t.Fatalf("preferredIconPNG: %v", err)
	}
	if !bytes.Equal(got, embedded) {
		t.Error("preferredIconPNG did not fall back to the embedded icon")
	}
}

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return buf.Bytes()
}
