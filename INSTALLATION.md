# Installation

## Requirements

- Go 1.23+
- Linux, macOS, or Windows
- C toolchain required by GUI dependencies (platform dependent)

## Project Layout

Implementation lives in:

- repository root (`.`)

## Build And Run

```bash
cd subsurface-to-ssi-qr
go mod tidy
go test ./...
go run ./cmd/app
```

From repository root, you can also use Makefile targets:

```bash
make doctor
make test
make build-cli
make build-gui
make run-gui
```

### Build Version Injection

Both binaries embed a build version using Go `-ldflags`.

- Default value: derived from `git describe --tags --always --dirty` (fallback: `dev`)
- Manual override: pass `VERSION=...` to `make`

Examples:

```bash
make build-cli VERSION=v1.2.3
make build-gui VERSION=v1.2.3
```

Where version is shown:

- CLI: in `-h` / usage header
- GUI: discreet label in the bottom bar, right side

## GUI Language Selection

The desktop GUI supports three languages:

- `EN` (default on first run)
- `PL`
- `DE`

How it works:

- Use the language selector in the top toolbar (`EN` / `PL` / `DE`).
- The selected language is remembered across app restarts.
- Translation catalogs are loaded from standard `go-i18n` TOML files in:
	`cmd/app/locales/`

## Environment Check (`make doctor`)

Run from repository root:

```bash
make doctor
```

What it checks:

- `go` availability and version
- `GOOS`, `GOARCH`, `CGO_ENABLED`
- On Linux: presence of `pkg-config`
- On Linux: GUI link dependencies used by Fyne (`x11`, `xrandr`, `xi`, `xcursor`, `xinerama`, `xxf86vm`, `gl`)

If Linux GUI modules are missing, `make doctor` prints install hints for Fedora
and Ubuntu/Debian. Missing GUI modules do not block pure Go CLI usage.

## Pure Go Mode (No GUI)

If you want to avoid native GUI linkage on Linux, use the CLI target:

```bash
cd subsurface-to-ssi-qr
CGO_ENABLED=0 go build ./cmd/cli
./cli -input ./tests/testdata/sample_subsurface.xml -index 1 -out-png ./dive1.png
```

For real logs with many dives, list entries first and then select index:

```bash
./cli -input ../data/addr@email.com.ssrf -list
./cli -input ../data/addr@email.com.ssrf -index 3 -out-png ./dive3.png
```

## Notes For Linux

The GUI (`cmd/app`) uses Fyne and requires cgo plus native desktop/OpenGL libs.
Even if your desktop session is Wayland, many stacks still rely on XWayland for
compatibility, and the linker can still require X11 development libraries.

Fedora example for the missing `-lXxf86vm` linker error:

```bash
sudo dnf install libX11-devel libXrandr-devel libXi-devel libXcursor-devel libXinerama-devel libXxf86vm-devel mesa-libGL-devel
```

If you want to avoid these dependencies completely on Linux, use the pure Go CLI
target (`cmd/cli`) with `CGO_ENABLED=0`.

## Notes For Windows

The project works on Windows:

- CLI mode works as pure Go and is the simplest path.
- GUI mode is supported by Fyne on Windows, but still uses cgo (requires a C toolchain for builds from source).

Typical source build flow on Windows (PowerShell):

```powershell
cd .\subsurface-to-ssi-qr
go test ./...
go build -o .\bin\subsurface-ssi-cli.exe .\cmd\cli
go build -o .\bin\subsurface-ssi-gui.exe .\cmd\app
```

## Optional Config Customization

Edit:

- `internal/config/defaults.yaml`

Current app build uses built-in defaults and does not yet expose config file
selection in GUI. This is planned for next iteration.
