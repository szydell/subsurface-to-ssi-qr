package main

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var localeFS embed.FS

type translator struct {
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	lang      string
}

func newTranslator(lang string) (*translator, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	paths, err := fs.Glob(localeFS, "locales/*.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to list locale files: %w", err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no locale files found")
	}

	for _, p := range paths {
		b, readErr := fs.ReadFile(localeFS, p)
		if readErr != nil {
			return nil, fmt.Errorf("failed reading locale file %s: %w", p, readErr)
		}
		if _, parseErr := bundle.ParseMessageFileBytes(b, p); parseErr != nil {
			return nil, fmt.Errorf("failed parsing locale file %s: %w", p, parseErr)
		}
	}

	t := &translator{bundle: bundle}
	t.setLanguage(lang)
	return t, nil
}

func (t *translator) setLanguage(lang string) {
	t.lang = normalizeLang(lang)
	t.localizer = i18n.NewLocalizer(t.bundle, t.lang)
}

func (t *translator) text(id string) string {
	return t.textData(id, nil)
}

func (t *translator) textData(id string, data map[string]any) string {
	return t.localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
}

func (t *translator) textCount(id string, count int, data map[string]any) string {
	if data == nil {
		data = make(map[string]any, 1)
	}
	data["PluralCount"] = count

	return t.localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
		PluralCount:  count,
	})
}

func (t *translator) localize(cfg *i18n.LocalizeConfig) string {
	msg, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    cfg.MessageID,
		TemplateData: cfg.TemplateData,
		PluralCount:  cfg.PluralCount,
	})
	if err != nil {
		return cfg.MessageID
	}
	return msg
}

func normalizeLang(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pl":
		return "pl"
	case "de":
		return "de"
	default:
		return "en"
	}
}

func langOption(lang string) string {
	switch normalizeLang(lang) {
	case "pl":
		return "PL"
	case "de":
		return "DE"
	default:
		return "EN"
	}
}

func langCode(option string) string {
	switch strings.ToUpper(strings.TrimSpace(option)) {
	case "PL":
		return "pl"
	case "DE":
		return "de"
	default:
		return "en"
	}
}
