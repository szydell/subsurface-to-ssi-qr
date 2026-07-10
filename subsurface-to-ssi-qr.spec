# SPDX-License-Identifier: Apache-2.0
#
# Fedora/COPR RPM spec for subsurface-to-ssi-qr.
# This package builds:
# - CLI tool (pure Go)
# - GUI app (Fyne + cgo) as a subpackage

%global forgeurl https://github.com/szydell/subsurface-to-ssi-qr
# Release version is kept in sync by the release workflow.
%global base_version 1.0.9
# Upstream release tags use format: vX.Y.Z.
%global tag v%{base_version}
# Keep the source archive basename in Name-Version form (no 'v' prefix),
# matching the directory name inside the GitHub tag tarball.
%global archivename %{name}-%{base_version}
# %%forgemeta must see %%{version} already matching %%{tag} (minus the 'v'
# prefix) to recognize this as a proper tagged release. Since it runs before
# the Version: tag below is parsed, %%{version} would otherwise still be
# empty, forgemeta would treat %%{tag} as an unmatched snapshot ref, and it
# would stuff a redundant ".<date>git<tag>" into %%dist (e.g. producing
# 1.1.2-1.20260710gitv1.1.2.fc44 instead of the expected 1.1.2-1.fc44).
%global version %{base_version}
# %%forgemeta (from go-rpm-macros/rpmautospec forge macros) computes Source0
# (%%{forgesource}) using RPM's built-in Lua only - no external shell tools
# involved - so it stays valid no matter which COPR build stage/chroot
# re-parses this spec. COPR's own backend fetches forge-style sources before
# invoking the sandboxed SRPM build, so no direct network access is required
# inside that sandbox.
%undefine _disable_source_fetch
%forgemeta

Name:           subsurface-to-ssi-qr
Version:        %{base_version}
Release:        %autorelease
Summary:        Convert Subsurface dive logs to SSI-compatible QR payloads and QR images

License:        Apache-2.0
URL:            %{forgeurl}
Source0:        %{forgesource}

BuildRequires:  go-rpm-macros
BuildRequires:  golang >= 1.26
BuildRequires:  make
BuildRequires:  gcc
BuildRequires:  desktop-file-utils
BuildRequires:  pkgconfig

# GUI/cgo build dependencies used by Fyne on Linux/X11.
BuildRequires:  libxkbcommon-devel
BuildRequires:  libXcursor-devel
BuildRequires:  mesa-libGL-devel
BuildRequires:  gtk3-devel
BuildRequires:  libX11-devel
BuildRequires:  libXrandr-devel
BuildRequires:  libXi-devel
BuildRequires:  libXinerama-devel
BuildRequires:  libXxf86vm-devel

%description
Standalone tool that converts Subsurface dive logs (.ssrf) to
SSI-compatible QR payloads and QR images.

The base package ships the pure Go CLI utility. The GUI desktop application is
provided by the %{name}-gui subpackage.

%package gui
Summary:        Desktop GUI for %{name} (Fyne)
Requires:       %{name}%{?_isa} = %{version}-%{release}

# Explicit runtime requirements requested for GUI deployment.
Requires:       libxkbcommon%{?_isa}
Requires:       libXcursor%{?_isa}
Requires:       mesa-libGL%{?_isa}

%description gui
Desktop GUI application for converting Subsurface dive logs to SSI-compatible
QR payloads and QR images.

This subpackage contains the Fyne-based graphical frontend.

%prep
%forgeautosetup

%build
# Ensure version string embedded in binaries matches upstream git tag format.
export VERSION=v%{version}
# GUI requires cgo; CLI target in Makefile still forces CGO_ENABLED=0 internally.
export CGO_ENABLED=1
# Fedora's build environment exports LDFLAGS with C/C++ hardening flags
# (-Wl,... -specs=...). The Makefile's `LDFLAGS ?= -X ...` respects a
# pre-set LDFLAGS env var, so without unsetting it those C-linker flags
# would be passed straight to `go build -ldflags`, which Go's linker does
# not understand ("flag provided but not defined: -Wl,-z,relro").
unset LDFLAGS
%make_build build-cli build-gui VERSION=${VERSION}

%check
# Run upstream unit tests.
go test ./...

%install
install -Dpm0755 bin/subsurface-ssi-cli %{buildroot}%{_bindir}/subsurface-ssi-cli
install -Dpm0755 bin/subsurface-ssi-gui %{buildroot}%{_bindir}/subsurface-ssi-gui
install -Dpm0644 assets/icon.png %{buildroot}%{_datadir}/pixmaps/subsurface-to-ssi-qr.png

# Install desktop entry for the GUI app.
install -d %{buildroot}%{_datadir}/applications
cat > %{buildroot}%{_datadir}/applications/%{name}-gui.desktop << 'EOF'
[Desktop Entry]
Type=Application
Name=Subsurface to SSI QR
Comment=Convert Subsurface dive logs to SSI-compatible QR payloads
Exec=subsurface-ssi-gui
Icon=subsurface-to-ssi-qr
Terminal=false
Categories=Utility;Education;
Keywords=diving;subsurface;ssi;qr;
StartupWMClass=subsurface-ssi-gui
EOF

desktop-file-install \
  --dir=%{buildroot}%{_datadir}/applications \
  --set-key=StartupNotify --set-value=true \
  %{buildroot}%{_datadir}/applications/%{name}-gui.desktop

desktop-file-validate %{buildroot}%{_datadir}/applications/%{name}-gui.desktop

%files
%license LICENSE
%doc README.md INSTALLATION.md FORMAT.md API.md
%{_bindir}/subsurface-ssi-cli

%files gui
%{_bindir}/subsurface-ssi-gui
%{_datadir}/applications/%{name}-gui.desktop
%{_datadir}/pixmaps/subsurface-to-ssi-qr.png

%changelog
%autochangelog
