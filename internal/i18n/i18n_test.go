package i18n

import "testing"

func TestParity(t *testing.T) {
	for k := range strings_de {
		if _, ok := strings_en[k]; !ok {
			t.Errorf("key %q present in de but missing in en", k)
		}
	}
	for k := range strings_en {
		if _, ok := strings_de[k]; !ok {
			t.Errorf("key %q present in en but missing in de", k)
		}
	}
}

func TestT_FallsBackToKey(t *testing.T) {
	SetLanguage("de")
	got := T("nonexistent.key")
	if got != "nonexistent.key" {
		t.Errorf("expected fallback to key, got %q", got)
	}
}

func TestT_PicksLanguage(t *testing.T) {
	SetLanguage("de")
	if got := T("app.title"); got != strings_de["app.title"] {
		t.Errorf("got %q want %q", got, strings_de["app.title"])
	}
	SetLanguage("en")
	if got := T("app.title"); got != strings_en["app.title"] {
		t.Errorf("got %q want %q", got, strings_en["app.title"])
	}
}
