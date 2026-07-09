# API

## Overview

The app is structured into independent modules:

- `internal/subsurface`: Parse Subsurface XML into normalized records
- `internal/model`: Shared domain model (`DiveRecord`)
- `internal/config`: Mapping defaults and config loading
- `pkg/ssi`: Public SSI payload model, mapping and serialization module
- `internal/ssi`: Backward-compatible adapter for app internals
- `internal/qr`: QR PNG generation
- `cmd/app`: Desktop GUI

## Data Flow

1. User selects Subsurface XML file.
2. Parser reads dives to `[]model.DiveRecord`.
3. Each dive is mapped to `ssi.Payload`.
4. Payload serializer creates SSI string (`dive;noid;...`).
5. QR module renders PNG bytes.
6. GUI shows text payload and QR preview.

## Public Module For External Integrators

Import path:

- `github.com/szydell/subsurface-to-ssi-qr/pkg/ssi`

Purpose:

- build SSI payload from your own parser/input model,
- use neutral `ssi.DiveInput` (no dependency on this repository internals),
- map to `ssi.Payload` and serialize to SSI QR string.

## Key Functions

### Subsurface Parser

- `ParseFile(path string) ([]model.DiveRecord, error)`
- `Parse(r io.Reader) ([]model.DiveRecord, error)`

### Mapping And Serialization

- `MapDive(in ssi.DiveInput, cfg ssi.MappingConfig) ssi.Payload`
- `BuildPayloadFromDive(in ssi.DiveInput, cfg ssi.MappingConfig, mode ssi.ValidationMode) (string, error)`
- `BuildPayload(p ssi.Payload, includeUser bool, mode ValidationMode) (string, error)`
- `ValidateRequired(p ssi.Payload) error`

### QR

- `PNG(payload string, size int) ([]byte, error)`
- `WritePNG(payload string, size int, path string) error`

## Validation Modes

- `lenient`: generate payload whenever possible
- `strict`: fail if required fields are missing

## Error Handling

- Parser returns error for malformed XML or missing valid dives.
- Strict payload validation returns descriptive field-specific errors.
- GUI reports errors in status bar and continues running.
