package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/webfraggle/mbd-video-converter/internal/profile"
)

func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func stringDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// profileNameForPattern extracts the display-size prefix from factory profile names.
// E.g. `"0.96\" Display"` → `"0.96"`, `"Mein Bahnhof"` → `"Mein Bahnhof"`.
func profileNameForPattern(p profile.Profile) string {
	name := p.Name
	if i := strings.Index(name, "\""); i > 0 {
		return name[:i]
	}
	return name
}
