package ssi

import publicssi "github.com/szydell/subsurface-to-ssi-qr/pkg/ssi"

// Backward-compatible aliases for code still importing internal/ssi.
type ValidationMode = publicssi.ValidationMode

const (
	ValidationLenient = publicssi.ValidationLenient
	ValidationStrict  = publicssi.ValidationStrict
)

type Payload = publicssi.Payload

func BuildPayload(p Payload, includeUser bool, mode ValidationMode) (string, error) {
	return publicssi.BuildPayload(p, includeUser, mode)
}

func ValidateRequired(p Payload) error {
	return publicssi.ValidateRequired(p)
}
