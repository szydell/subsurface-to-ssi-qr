# Contributing

## Development Setup

```bash
cd subsurface-to-ssi-qr
go mod tidy
go test ./...
```

## Scope Priorities

1. Compatibility with real Subsurface exports
2. Stable SSI payload serialization
3. Clear handling for undocumented `var_*` fields
4. Manual validation in SSI mobile app

## Code Guidelines

- Keep modules small and focused.
- Prefer deterministic behavior (especially payload order).
- Avoid silently swallowing parse/validation errors.
- Add tests for every parser or mapping behavior change.

## Testing

- Unit tests: parser and payload modules
- Golden payload checks when adding new mappings
- Manual scan tests in SSI app before release tags

## Localization

- GUI uses `go-i18n` with TOML files in `cmd/app/locales/`.
- To add a language, copy `active.en.toml` to `active.<lang>.toml` and translate message values.
- Keep message IDs unchanged; only translate `other = "..."` values.
- Keep all locale files in sync (`active.en.toml`, `active.pl.toml`, `active.de.toml`).
- New message IDs must be added to every supported language file in the same PR.
- Runtime behavior requirement: first run defaults to `EN`; selected language must remain persisted.
- Verify GUI language switch and run `go test ./...`.

## Pull Request Checklist

- [ ] Tests added/updated
- [ ] `go test ./...` passes
- [ ] Docs updated (`FORMAT.md`, `API.md`, README when needed)
- [ ] Behavior verified with at least one real Subsurface export
- [ ] If `var_*` changed: mark confidence level and source

## Known Risk Areas

- Non-standard Subsurface XML variants
- Locale-dependent number/date formats
- SSI undocumented field expectations
