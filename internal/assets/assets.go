// Package assets embeds static application assets (such as the app icon)
// directly into the compiled binary. This makes them available at runtime
// regardless of the current working directory or install location, which
// matters especially on Windows where the GUI binary is often distributed
// as a single portable .exe with no accompanying "assets" directory.
package assets

import _ "embed"

// IconPNG is the raw PNG-encoded application icon, embedded at build time
// from icon.png (kept in sync with ../../assets/icon.png, the source of
// truth used for packaging, e.g. the RPM spec's pixmap install).
//
//go:embed icon.png
var IconPNG []byte
