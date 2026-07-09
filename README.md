# subsurface-to-ssi-qr

Standalone desktop tool that converts Subsurface dive logs (`.xml` / `.ssrf`) to
SSI-compatible QR payloads and QR images.

## Status

Initial implementation (MVP) is available under:

- `subsurface-to-ssi-qr/`

Implemented in MVP:

- Subsurface XML parser
- SSI field mapping with configurable defaults
- SSI payload generator (`dive;noid;...`)
- QR PNG generation
- Minimal desktop GUI to load a file, choose dive, preview payload and QR
- Pure Go CLI mode (no GUI dependencies)
- Unit tests for parser and payload generation

## Quick Start

1. Enter project directory:

```bash
cd subsurface-to-ssi-qr
```

2. Run tests:

```bash
go test ./...
```

3. Start desktop app:

```bash
go run ./cmd/app
```

3a. Start pure Go CLI (no GUI, recommended on headless/minimal systems):

```bash
go run ./cmd/cli -input ./tests/testdata/sample_subsurface.xml -index 1 -out-png ./dive1.png
```

List dives first (recommended for real logs):

```bash
go run ./cmd/cli -input ../data/addr@email.com.ssrf -list
go run ./cmd/cli -input ../data/addr@email.com.ssrf -index 3 -out-png ./dive3.png
```

3b. Or use Makefile shortcuts from repository root:

```bash
make doctor
make test
make build-cli
make run-cli-sample
```

`make doctor` checks local environment: Go availability, `CGO_ENABLED`, and on
Linux also `pkg-config` + GUI compatibility libraries required by Fyne.

4. In the app:

- Open a Subsurface file (`.xml` / `.ssrf`)
- Select dive index
- Copy payload and/or save QR PNG

## Configuration

Default mapping profile is in:

- `subsurface-to-ssi-qr/internal/config/defaults.yaml`

This includes current reverse-engineered defaults for `var_*` fields.

## Documentation

- `INSTALLATION.md`
- `FORMAT.md`
- `API.md`
- `CONTRIBUTING.md`

## Localization

GUI translations use `go-i18n` with TOML message files (standard i18n approach):

- `subsurface-to-ssi-qr/cmd/app/locales/active.en.toml`
- `subsurface-to-ssi-qr/cmd/app/locales/active.pl.toml`
- `subsurface-to-ssi-qr/cmd/app/locales/active.de.toml`

Language is selectable in GUI (`EN` / `PL` / `DE`) and remembered across runs.
Default on first run is `EN`.

## Known Limitations

- SSI QR format is reverse-engineered from public sources.
- `var_water_body_id` dictionary is still incomplete in public data.
- Full in-app SSI import validation requires manual testing on iOS/Android.
- Desktop GUI uses Fyne, which links native GUI/OpenGL libraries via cgo.
- On Linux Wayland sessions this usually works through XWayland, so X11 compatibility/devel packages can still be required for build/runtime.
- GUI build and run are supported on Windows as well. The CLI mode is pure Go and easiest to build everywhere.
