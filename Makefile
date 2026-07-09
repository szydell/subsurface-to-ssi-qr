GO ?= go
PROJECT_DIR ?= subsurface-to-ssi-qr
BIN_DIR ?= $(PROJECT_DIR)/bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS ?= -X subsurface-to-ssi-qr/internal/buildinfo.Version=$(VERSION)
JOBS ?= $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
GO_PARALLEL_FLAGS ?= -p $(JOBS)

HOST_GOOS ?= $(shell $(GO) env GOOS)
HOST_GOARCH ?= $(shell $(GO) env GOARCH)
EXE_EXT ?= $(if $(filter windows,$(HOST_GOOS)),.exe,)

CLI_BIN ?= $(BIN_DIR)/subsurface-ssi-cli$(EXE_EXT)
GUI_BIN ?= $(BIN_DIR)/subsurface-ssi-gui$(EXE_EXT)

DIST_DIR ?= dist
CLI_ARTIFACT ?= $(DIST_DIR)/subsurface-ssi-cli-$(HOST_GOOS)-$(HOST_GOARCH)$(EXE_EXT)
GUI_ARTIFACT ?= $(DIST_DIR)/subsurface-ssi-gui-$(HOST_GOOS)-$(HOST_GOARCH)$(EXE_EXT)

SAMPLE_FILE ?= $(PROJECT_DIR)/tests/testdata/sample_subsurface.xml
SAMPLE_INDEX ?= 1
SAMPLE_QR ?= $(BIN_DIR)/sample-dive.png

.PHONY: help doctor test build-cli build-gui build-release-artifacts run-cli-sample run-gui clean

help:
	@echo "Available targets:"
	@echo "  make doctor          - Check local build environment and GUI prerequisites"
	@echo "  make test            - Run all Go tests"
	@echo "  make build-cli       - Build pure-Go CLI binary (CGO_ENABLED=0, embeds VERSION)"
	@echo "  make build-gui       - Build desktop GUI binary (requires native GUI deps, embeds VERSION)"
	@echo "  make build-release-artifacts - Build CLI+GUI and copy release files to ./dist with OS/ARCH suffixes"
	@echo "  (parallelism)        - Override with JOBS=<num>, e.g. make build-gui JOBS=16"
	@echo "  make run-cli-sample  - Generate payload + QR from sample XML"
	@echo "  make run-gui         - Start desktop GUI"
	@echo "  make clean           - Remove built binaries"
	@echo ""
	@echo "Optional override: make build-cli VERSION=v1.2.3"

doctor:
	@echo "==> Environment doctor"
	@status=0; \
	if command -v $(GO) >/dev/null 2>&1; then \
		echo "[OK] go found: $$($(GO) version)"; \
	else \
		echo "[FAIL] go not found in PATH"; \
		status=1; \
	fi; \
	if [ $$status -ne 0 ]; then \
		echo "Fix Go installation first."; \
		exit $$status; \
	fi; \
	echo "[INFO] GOOS=$$($(GO) env GOOS) GOARCH=$$($(GO) env GOARCH) CGO_ENABLED=$$($(GO) env CGO_ENABLED)"; \
	os="$$($(GO) env GOOS)"; \
	if [ "$$os" = "linux" ]; then \
		echo "[INFO] Linux GUI checks (Fyne/OpenGL/X11 compatibility libs)"; \
		if command -v pkg-config >/dev/null 2>&1; then \
			echo "[OK] pkg-config found"; \
			missing=""; \
			for dep in x11 xrandr xi xcursor xinerama xxf86vm gl; do \
				if pkg-config --exists $$dep; then \
					echo "[OK] pkg-config module: $$dep"; \
				else \
					echo "[WARN] missing pkg-config module: $$dep"; \
					missing="$$missing $$dep"; \
				fi; \
			done; \
			if [ -n "$$missing" ]; then \
				echo "[WARN] GUI may fail to link without these modules:$$missing"; \
				echo "[HINT] Fedora: sudo dnf install libX11-devel libXrandr-devel libXi-devel libXcursor-devel libXinerama-devel libXxf86vm-devel mesa-libGL-devel"; \
				echo "[HINT] Ubuntu/Debian: sudo apt install libgtk-3-dev libx11-dev libxrandr-dev libxi-dev libxcursor-dev libxinerama-dev libxxf86vm-dev libgl1-mesa-dev"; \
			else \
				echo "[OK] Linux GUI dependencies look available"; \
			fi; \
		else \
			echo "[WARN] pkg-config not found; cannot verify Linux GUI dependencies"; \
			echo "[HINT] Install pkg-config, then rerun make doctor"; \
		fi; \
	else \
		echo "[INFO] Non-Linux OS detected ($$os). Skipping Linux-specific GUI dependency checks."; \
		echo "[INFO] CLI should build everywhere; GUI build on Windows/macOS requires cgo toolchain."; \
	fi; \
	echo "[INFO] Doctor completed"

test:
	cd $(PROJECT_DIR) && GOMAXPROCS=$(JOBS) $(GO) test $(GO_PARALLEL_FLAGS) ./...

build-cli:
	mkdir -p $(BIN_DIR)
	cd $(PROJECT_DIR) && GOMAXPROCS=$(JOBS) CGO_ENABLED=0 $(GO) build $(GO_PARALLEL_FLAGS) -ldflags "$(LDFLAGS)" -o ./bin/subsurface-ssi-cli$(EXE_EXT) ./cmd/cli

build-gui:
	mkdir -p $(BIN_DIR)
	cd $(PROJECT_DIR) && GOMAXPROCS=$(JOBS) $(GO) build $(GO_PARALLEL_FLAGS) -ldflags "$(LDFLAGS)" -o ./bin/subsurface-ssi-gui$(EXE_EXT) ./cmd/app

build-release-artifacts:
	mkdir -p $(DIST_DIR)
	$(MAKE) -j2 build-cli build-gui VERSION=$(VERSION) JOBS=$(JOBS)
	cp $(CLI_BIN) $(CLI_ARTIFACT)
	cp $(GUI_BIN) $(GUI_ARTIFACT)
	@echo "Prepared release artifacts:"
	@echo "  $(CLI_ARTIFACT)"
	@echo "  $(GUI_ARTIFACT)"

run-cli-sample: build-cli
	$(CLI_BIN) -input $(SAMPLE_FILE) -index $(SAMPLE_INDEX) -out-png $(SAMPLE_QR)
	@echo "Saved sample QR to $(SAMPLE_QR)"

run-gui:
	cd $(PROJECT_DIR) && $(GO) run ./cmd/app

clean:
	rm -f $(CLI_BIN) $(GUI_BIN)