package i18n

import (
	"os"
	"strings"
	"sync"
)

var (
	mu      sync.RWMutex
	current = "en"
)

// SetLanguage switches the active language ("de" or "en"). Unknown values fall back to "en".
func SetLanguage(lang string) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := tables[lang]; ok {
		current = lang
	} else {
		current = "en"
	}
}

// CurrentLanguage returns the active language code.
func CurrentLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// T looks up the translation for key in the active language. If missing, the key itself is returned.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if v, ok := tables[current][key]; ok {
		return v
	}
	return key
}

// DefaultLanguage returns "de" if the OS locale starts with "de", otherwise "en".
func DefaultLanguage() string {
	for _, env := range []string{"LANG", "LC_ALL", "LC_MESSAGES"} {
		if v := os.Getenv(env); strings.HasPrefix(strings.ToLower(v), "de") {
			return "de"
		}
	}
	return "en"
}

var tables = map[string]map[string]string{
	"de": strings_de,
	"en": strings_en,
}
