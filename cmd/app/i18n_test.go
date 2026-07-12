package main

import "testing"

func TestTranslatorLoadedDivesPluralization(t *testing.T) {
	t.Parallel()

	tr, err := newTranslator("pl")
	if err != nil {
		t.Fatalf("newTranslator(pl): %v", err)
	}

	tests := []struct {
		lang  string
		count int
		want  string
	}{
		{lang: "pl", count: 1, want: "Wczytano nurkowanie z sample.ssrf"},
		{lang: "pl", count: 2, want: "Wczytano 2 nurkowania z sample.ssrf"},
		{lang: "pl", count: 4, want: "Wczytano 4 nurkowania z sample.ssrf"},
		{lang: "pl", count: 5, want: "Wczytano 5 nurkowań z sample.ssrf"},
		{lang: "pl", count: 22, want: "Wczytano 22 nurkowania z sample.ssrf"},
		{lang: "en", count: 1, want: "Loaded 1 dive from sample.ssrf"},
		{lang: "en", count: 3, want: "Loaded 3 dives from sample.ssrf"},
		{lang: "de", count: 1, want: "1 Tauchgang aus sample.ssrf geladen"},
		{lang: "de", count: 3, want: "3 Tauchgaenge aus sample.ssrf geladen"},
	}

	for _, tc := range tests {
		tr.setLanguage(tc.lang)
		got := tr.textCount("status_loaded_n_dives", tc.count, map[string]any{"File": "sample.ssrf"})
		if got != tc.want {
			t.Fatalf("lang=%s count=%d: got %q, want %q", tc.lang, tc.count, got, tc.want)
		}
	}
}
