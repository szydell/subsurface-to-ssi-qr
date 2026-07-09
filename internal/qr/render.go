package qr

import (
	"os"

	qrcode "github.com/skip2/go-qrcode"
)

// PNG returns QR image bytes for the provided payload.
func PNG(payload string, size int) ([]byte, error) {
	if size <= 0 {
		size = 300
	}
	return qrcode.Encode(payload, qrcode.Medium, size)
}

// WritePNG writes a QR PNG file to disk.
func WritePNG(payload string, size int, path string) error {
	bytes, err := PNG(payload, size)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0o644)
}
