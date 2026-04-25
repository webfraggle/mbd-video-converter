# MBD-Videoconverter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Fyne-based desktop application for Windows and macOS that converts videos to MJPEG for ESP32-driven model railway displays, replacing the existing CLI workflow.

**Architecture:** Single Go binary using Fyne v2 GUI; pure Go core packages (`profile`, `settings`, `i18n`, `version`) with no UI dependency; an `ffmpeg` package that drives the bundled ffmpeg binary as a subprocess; a `job` package that orchestrates a sequential queue with skip-on-error and cancellation; a `ui` package that is the only consumer of all others. ffmpeg ships as a separate binary in the ZIP next to the app and can be overridden via settings.

**Tech Stack:** Go 1.21+, Fyne v2 (`fyne.io/fyne/v2`), `fyne-cross` for cross-compilation, `github.com/google/shlex` for parsing Advanced ffmpeg args, `github.com/google/uuid` for user-profile IDs.

**Reference spec:** `docs/superpowers/specs/2026-04-25-mbd-videoconverter-design.md`. Read it before starting implementation.

---

## File Structure

```
main.go                              Entry: launches Fyne UI
go.mod / go.sum
build.sh                             Cross-compile + version bump + ZIP packaging
VERSION                              vX.Y.Z, patch auto-incremented by build.sh
Icon.png                             App icon (placeholder for now)
README.md                            User-facing docs

internal/
  version/version.go                 Single ldflags-injected variable
  i18n/i18n.go                       T(key) lookup, default locale detection
  i18n/de.go                         German strings
  i18n/en.go                         English strings
  profile/profile.go                 Profile struct, validation
  profile/factory.go                 4 hardcoded factory profiles
  profile/store.go                   JSON load/save of user profiles
  settings/settings.go               Settings struct, JSON load/save
  ffmpeg/locate.go                   Resolve ffmpeg binary path
  ffmpeg/command.go                  Build argument list from EncodingConfig
  ffmpeg/runner.go                   exec.CommandContext, progress parsing
  ffmpeg/runner_windows.go           HideWindow + CREATE_NO_WINDOW
  ffmpeg/runner_unix.go              No-op SysProcAttr stub
  job/job.go                         Job struct + JobStatus
  job/encoding_config.go             EncodingConfig + filename pattern resolver
  job/queue.go                       Sequential queue with cancel + progress channel
  ui/ui.go                           Run() — main window assembly
  ui/queue_view.go                   Queue table, file picker, drop overlay
  ui/profile_panel.go                Profile list + value editor (right column)
  ui/settings_dialog.go              Modal dialog
  ui/dnd.go                          Drag-over overlay helper

docs/
  superpowers/specs/2026-04-25-mbd-videoconverter-design.md
  superpowers/plans/2026-04-25-mbd-videoconverter-implementation.md
  manual-test-plan.md                Per-release manual QA checklist

scripts/                             Legacy reference scripts (already in repo)
```

Test files live next to the package they test (`profile_test.go` etc.). Integration tests for `ffmpeg/runner` use build tag `//go:build integration`.

---

## Phase 1: Bootstrap

### Task 1: Initialize Go module and minimal main.go

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `internal/version/version.go`

- [ ] **Step 1: Initialize Go module**

```bash
go mod init github.com/webfraggle/mbd-video-converter
```

- [ ] **Step 2: Add Fyne dependency**

```bash
go get fyne.io/fyne/v2@latest
go get fyne.io/fyne/v2/app
go get fyne.io/fyne/v2/widget
```

- [ ] **Step 3: Create `internal/version/version.go`**

```go
package version

// Version is set at build time via -ldflags "-X .../internal/version.Version=vX.Y.Z".
var Version = "dev"
```

- [ ] **Step 4: Create `main.go` that opens an empty window**

```go
package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/version"
)

func main() {
	a := app.NewWithID("de.modellbahn-displays.mbd-videoconverter")
	w := a.NewWindow("MBD-Videoconverter " + version.Version)
	w.SetContent(widget.NewLabel("Hello, world."))
	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()
}
```

(Add `"fyne.io/fyne/v2"` import for `fyne.NewSize`.)

- [ ] **Step 5: Verify it compiles and runs**

Run: `go run .`
Expected: A window titled "MBD-Videoconverter dev" with a "Hello, world." label.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum main.go internal/version/version.go
git commit -m "Bootstrap: minimal Fyne window + version package"
```

### Task 2: VERSION file and ldflags wiring (build script comes later)

**Files:**
- Create: `VERSION`
- Modify: `main.go` (no change actually — just verify the ldflags path works manually)

- [ ] **Step 1: Create `VERSION`**

Content (single line, no trailing newline added by editor):
```
v0.1.0
```

- [ ] **Step 2: Verify ldflags injection works**

Run:
```bash
go run -ldflags "-X github.com/webfraggle/mbd-video-converter/internal/version.Version=$(cat VERSION)" .
```
Expected: Window title contains "v0.1.0".

- [ ] **Step 3: Commit**

```bash
git add VERSION
git commit -m "Add VERSION file at v0.1.0"
```

---

## Phase 2: Core data types

### Task 3: `internal/profile/profile.go` — Profile struct and validation

**Files:**
- Create: `internal/profile/profile.go`
- Test: `internal/profile/profile_test.go`

- [ ] **Step 1: Write `profile_test.go` with validation tests**

```go
package profile

import "testing"

func TestProfileValidate(t *testing.T) {
	cases := []struct {
		name    string
		p       Profile
		wantErr bool
	}{
		{"valid", Profile{Name: "x", Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"}, false},
		{"empty name", Profile{Width: 120, Height: 240, FPS: 20, Quality: 9, Scaler: "lanczos"}, true},
		{"zero width", Profile{Name: "x", Width: 0, Height: 240, FPS: 20, Quality: 9, Scaler: "lanczos"}, true},
		{"negative height", Profile{Name: "x", Width: 120, Height: -1, FPS: 20, Quality: 9, Scaler: "lanczos"}, true},
		{"fps zero", Profile{Name: "x", Width: 120, Height: 240, FPS: 0, Quality: 9, Scaler: "lanczos"}, true},
		{"quality below range", Profile{Name: "x", Width: 120, Height: 240, FPS: 20, Quality: 0, Scaler: "lanczos"}, true},
		{"quality above range", Profile{Name: "x", Width: 120, Height: 240, FPS: 20, Quality: 32, Scaler: "lanczos"}, true},
		{"empty scaler", Profile{Name: "x", Width: 120, Height: 240, FPS: 20, Quality: 9, Scaler: ""}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests — should fail to compile**

```bash
go test ./internal/profile/...
```
Expected: FAIL — `Profile` undefined.

- [ ] **Step 3: Write `internal/profile/profile.go`**

```go
package profile

import "fmt"

type Profile struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Factory    bool    `json:"factory"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FPS        int     `json:"fps"`
	Quality    int     `json:"quality"`
	Saturation float64 `json:"saturation"`
	Gamma      float64 `json:"gamma"`
	Scaler     string  `json:"scaler"`
}

func (p Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if p.Width <= 0 {
		return fmt.Errorf("width must be > 0")
	}
	if p.Height <= 0 {
		return fmt.Errorf("height must be > 0")
	}
	if p.FPS <= 0 {
		return fmt.Errorf("fps must be > 0")
	}
	if p.Quality < 1 || p.Quality > 31 {
		return fmt.Errorf("quality must be in [1,31]")
	}
	if p.Scaler == "" {
		return fmt.Errorf("scaler must not be empty")
	}
	return nil
}
```

- [ ] **Step 4: Run tests — should pass**

```bash
go test ./internal/profile/...
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/profile/profile.go internal/profile/profile_test.go
git commit -m "profile: add Profile struct with validation"
```

### Task 4: `internal/profile/factory.go` — four factory profiles

**Files:**
- Create: `internal/profile/factory.go`
- Modify: `internal/profile/profile_test.go`

- [ ] **Step 1: Add a test that the four factory profiles match the design spec**

Append to `internal/profile/profile_test.go`:

```go
func TestFactoryProfiles(t *testing.T) {
	want := []Profile{
		{ID: "factory:0.96", Name: "0.96\" Display", Factory: true, Width: 80, Height: 160, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.05", Name: "1.05\" Display", Factory: true, Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.14", Name: "1.14\" Display", Factory: true, Width: 135, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.90", Name: "1.90\" Display", Factory: true, Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
	}
	got := Factory()
	if len(got) != len(want) {
		t.Fatalf("got %d profiles, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("profile %d:\n got=%+v\nwant=%+v", i, got[i], want[i])
		}
	}
	for _, p := range got {
		if err := p.Validate(); err != nil {
			t.Errorf("factory profile %s invalid: %v", p.ID, err)
		}
	}
}
```

- [ ] **Step 2: Run tests — should fail**

```bash
go test ./internal/profile/...
```
Expected: FAIL — `Factory` undefined.

- [ ] **Step 3: Implement `internal/profile/factory.go`**

```go
package profile

func Factory() []Profile {
	return []Profile{
		{ID: "factory:0.96", Name: "0.96\" Display", Factory: true, Width: 80, Height: 160, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.05", Name: "1.05\" Display", Factory: true, Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.14", Name: "1.14\" Display", Factory: true, Width: 135, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.90", Name: "1.90\" Display", Factory: true, Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
	}
}
```

- [ ] **Step 4: Run tests — pass**

```bash
go test ./internal/profile/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/profile/factory.go internal/profile/profile_test.go
git commit -m "profile: add four factory profiles"
```

### Task 5: `internal/profile/store.go` — load/save user profiles JSON

**Files:**
- Create: `internal/profile/store.go`
- Test: `internal/profile/store_test.go`

- [ ] **Step 1: Write `store_test.go`**

```go
package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.json")
	s := NewStore(path)

	users := []Profile{
		{ID: "user:1", Name: "Custom 1", Width: 100, Height: 200, FPS: 25, Quality: 5, Saturation: 1.0, Gamma: 1.0, Scaler: "bicubic"},
		{ID: "user:2", Name: "Custom 2", Width: 200, Height: 100, FPS: 30, Quality: 10, Saturation: 1.5, Gamma: 0.9, Scaler: "lanczos"},
	}
	if err := s.Save(users); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != 2 || got[0] != users[0] || got[1] != users[1] {
		t.Errorf("got=%+v want=%+v", got, users)
	}
}

func TestStoreLoadMissingReturnsEmpty(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "missing.json"))
	got, err := s.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d profiles, want 0", len(got))
	}
}

func TestStoreSaveRejectsFactory(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "p.json"))
	err := s.Save([]Profile{{ID: "factory:bad", Name: "x", Factory: true, Width: 1, Height: 1, FPS: 1, Quality: 1, Scaler: "x"}})
	if err == nil {
		t.Fatal("expected error when saving Factory=true profile")
	}
}

func TestAllMergesFactoryAndUser(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "p.json"))
	user := []Profile{{ID: "user:1", Name: "u", Width: 1, Height: 1, FPS: 1, Quality: 1, Scaler: "x"}}
	_ = s.Save(user)

	all, err := s.All()
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("expected 4 factory + 1 user = 5, got %d", len(all))
	}
	if !all[0].Factory {
		t.Errorf("expected first entry to be factory")
	}
	if all[len(all)-1].ID != "user:1" {
		t.Errorf("expected user profile last")
	}
}

func ensureDirExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("file not created: %s", path)
	}
}
```

- [ ] **Step 2: Run tests — fail**

```bash
go test ./internal/profile/...
```

- [ ] **Step 3: Implement `internal/profile/store.go`**

```go
package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() ([]Profile, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Profile
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	return out, nil
}

func (s *Store) Save(users []Profile) error {
	for _, p := range users {
		if p.Factory {
			return fmt.Errorf("refusing to persist factory profile %q", p.ID)
		}
		if err := p.Validate(); err != nil {
			return fmt.Errorf("invalid profile %q: %w", p.ID, err)
		}
	}
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// All returns factory profiles followed by user profiles.
func (s *Store) All() ([]Profile, error) {
	users, err := s.Load()
	if err != nil {
		return nil, err
	}
	all := append([]Profile(nil), Factory()...)
	all = append(all, users...)
	return all, nil
}
```

- [ ] **Step 4: Run tests — pass**

```bash
go test ./internal/profile/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/profile/store.go internal/profile/store_test.go
git commit -m "profile: add Store for load/save user profiles + All() merge"
```

### Task 6: `internal/settings/settings.go` — app settings JSON persistence

**Files:**
- Create: `internal/settings/settings.go`
- Test: `internal/settings/settings_test.go`

- [ ] **Step 1: Write `settings_test.go`**

```go
package settings

import (
	"path/filepath"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s := New(path)

	in := Settings{
		Language:        "de",
		FFmpegPath:      "/usr/local/bin/ffmpeg",
		DefaultOutDir:   "/tmp",
		FilenamePattern: "{name}_{profile}_{fps}fps",
		OnExist:         "overwrite",
		LastProfileID:   "factory:1.05",
	}
	if err := s.Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != in {
		t.Errorf("got=%+v want=%+v", got, in)
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "missing.json"))
	got, err := s.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FilenamePattern == "" {
		t.Errorf("expected default FilenamePattern")
	}
	if got.OnExist != "overwrite" {
		t.Errorf("expected OnExist default 'overwrite', got %q", got.OnExist)
	}
}
```

- [ ] **Step 2: Run tests — fail**

- [ ] **Step 3: Implement `internal/settings/settings.go`**

```go
package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Settings struct {
	Language        string `json:"language"`
	FFmpegPath      string `json:"ffmpeg_path"`
	DefaultOutDir   string `json:"default_out_dir"`
	FilenamePattern string `json:"filename_pattern"`
	OnExist         string `json:"on_exist"` // "overwrite" | "suffix" | "fail"
	LastProfileID   string `json:"last_profile_id"`
}

func defaults() Settings {
	return Settings{
		Language:        "",
		FilenamePattern: "{name}_{profile}_{fps}fps",
		OnExist:         "overwrite",
	}
}

type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() (Settings, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults(), nil
	}
	if err != nil {
		return Settings{}, err
	}
	out := defaults()
	if err := json.Unmarshal(data, &out); err != nil {
		return Settings{}, err
	}
	return out, nil
}

func (s *Store) Save(v Settings) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
```

- [ ] **Step 4: Run tests — pass**

- [ ] **Step 5: Commit**

```bash
git add internal/settings/
git commit -m "settings: add Settings struct with JSON persistence and defaults"
```

### Task 7: `internal/i18n/` — translation map skeleton

**Files:**
- Create: `internal/i18n/i18n.go`
- Create: `internal/i18n/de.go`
- Create: `internal/i18n/en.go`
- Test: `internal/i18n/i18n_test.go`

- [ ] **Step 1: Write `i18n_test.go`**

```go
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
```

- [ ] **Step 2: Run tests — fail to compile**

- [ ] **Step 3: Implement files**

`internal/i18n/i18n.go`:
```go
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
```

`internal/i18n/de.go`:
```go
package i18n

var strings_de = map[string]string{
	"app.title":            "MBD-Videoconverter",
	"queue.header.file":    "Datei",
	"queue.header.status":  "Status",
	"queue.status.pending": "wartet",
	"queue.status.running": "läuft",
	"queue.status.done":    "fertig",
	"queue.status.failed":  "fehlgeschlagen",
	"queue.status.cancelled": "abgebrochen",
	"queue.btn.add":        "+ Datei…",
	"queue.btn.clear":      "Liste leeren",
	"queue.btn.cancel":     "Abbrechen",
	"queue.btn.convert":    "▶ Konvertieren",
	"profile.header":       "Aktives Profil",
	"profile.btn.new":      "+ Neu",
	"profile.btn.dup":      "Duplizieren",
	"profile.btn.del":      "Löschen",
	"profile.btn.save":     "Speichern",
	"profile.btn.saveAs":   "Als neues Profil…",
	"profile.field.width":  "Breite",
	"profile.field.height": "Höhe",
	"profile.field.fps":    "FPS",
	"profile.field.quality":    "Quality",
	"profile.field.saturation": "Saturation",
	"profile.field.gamma":  "Gamma",
	"profile.field.scaler": "Scaler",
	"profile.advanced":     "Advanced — rohe ffmpeg-Args",
	"settings.title":       "Einstellungen",
	"settings.btn.openLog": "Log-Ordner öffnen",
	"output.label.dir":     "Output-Ordner",
	"output.label.pattern": "Filename-Pattern",
	"dnd.overlay":          "Videos hier loslassen",
	"err.ffmpegNotFound":   "ffmpeg nicht gefunden. Pfad in Einstellungen prüfen.",
}
```

`internal/i18n/en.go`:
```go
package i18n

var strings_en = map[string]string{
	"app.title":            "MBD-Videoconverter",
	"queue.header.file":    "File",
	"queue.header.status":  "Status",
	"queue.status.pending": "waiting",
	"queue.status.running": "running",
	"queue.status.done":    "done",
	"queue.status.failed":  "failed",
	"queue.status.cancelled": "cancelled",
	"queue.btn.add":        "+ File…",
	"queue.btn.clear":      "Clear list",
	"queue.btn.cancel":     "Cancel",
	"queue.btn.convert":    "▶ Convert",
	"profile.header":       "Active profile",
	"profile.btn.new":      "+ New",
	"profile.btn.dup":      "Duplicate",
	"profile.btn.del":      "Delete",
	"profile.btn.save":     "Save",
	"profile.btn.saveAs":   "Save as new profile…",
	"profile.field.width":  "Width",
	"profile.field.height": "Height",
	"profile.field.fps":    "FPS",
	"profile.field.quality":    "Quality",
	"profile.field.saturation": "Saturation",
	"profile.field.gamma":  "Gamma",
	"profile.field.scaler": "Scaler",
	"profile.advanced":     "Advanced — raw ffmpeg args",
	"settings.title":       "Settings",
	"settings.btn.openLog": "Open log folder",
	"output.label.dir":     "Output folder",
	"output.label.pattern": "Filename pattern",
	"dnd.overlay":          "Drop videos here",
	"err.ffmpegNotFound":   "ffmpeg not found. Check the path in Settings.",
}
```

- [ ] **Step 4: Run tests — pass**

- [ ] **Step 5: Commit**

```bash
git add internal/i18n/
git commit -m "i18n: add T() lookup with de/en parity test"
```

---

## Phase 3: ffmpeg layer

### Task 8: `internal/ffmpeg/locate.go` — find ffmpeg binary

**Files:**
- Create: `internal/ffmpeg/locate.go`
- Test: `internal/ffmpeg/locate_test.go`

- [ ] **Step 1: Write `locate_test.go`**

```go
package ffmpeg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLocateUserPathPrefersUserSetting(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "myffmpeg")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Locate(bin, dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != bin {
		t.Errorf("got %q want %q", got, bin)
	}
}

func TestLocateFallsBackToBundled(t *testing.T) {
	dir := t.TempDir()
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}
	bin := filepath.Join(dir, name)
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Locate("", dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != bin {
		t.Errorf("got %q want %q", got, bin)
	}
}

func TestLocateMissingErrors(t *testing.T) {
	_, err := Locate("", t.TempDir())
	if err == nil {
		t.Fatal("expected error when ffmpeg missing")
	}
}

func TestLocateUserPathInvalidFallsBack(t *testing.T) {
	dir := t.TempDir()
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}
	bin := filepath.Join(dir, name)
	if err := os.WriteFile(bin, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Locate("/nonexistent/path", dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != bin {
		t.Errorf("got %q want %q (should fall back to bundled when user path invalid)", got, bin)
	}
}
```

- [ ] **Step 2: Run tests — fail**

- [ ] **Step 3: Implement `internal/ffmpeg/locate.go`**

```go
package ffmpeg

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// Locate returns the ffmpeg path to use.
// Order: userPath (if non-empty AND points to an executable file) → bundledDir/ffmpeg[.exe] → error.
func Locate(userPath, bundledDir string) (string, error) {
	if userPath != "" {
		if isExecutable(userPath) {
			return userPath, nil
		}
		// fall through to bundled
	}

	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}
	bundled := filepath.Join(bundledDir, name)
	if isExecutable(bundled) {
		return bundled, nil
	}
	return "", errors.New("ffmpeg binary not found (neither user setting nor bundled)")
}

func isExecutable(path string) bool {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true // Windows lacks +x bit; existence is enough.
	}
	return st.Mode().Perm()&0o111 != 0
}
```

- [ ] **Step 4: Run tests — pass**

- [ ] **Step 5: Commit**

```bash
git add internal/ffmpeg/locate.go internal/ffmpeg/locate_test.go
git commit -m "ffmpeg: add Locate() with user-path → bundled fallback"
```

### Task 9: `internal/ffmpeg/command.go` — build argument list

**Files:**
- Create: `internal/ffmpeg/command.go`
- Test: `internal/ffmpeg/command_test.go`

- [ ] **Step 1: Write `command_test.go` with snapshot-style tests against legacy scripts**

```go
package ffmpeg

import (
	"reflect"
	"testing"
)

func TestBuildArgs_Factory105(t *testing.T) {
	in := EncodingInput{
		InputPath:  "video.mp4",
		OutputPath: "video_1.05_20fps.mjpeg",
		Width:      120, Height: 240, FPS: 20, Quality: 9,
		Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos",
	}
	want := []string{
		"-y",
		"-i", "video.mp4",
		"-vf", "fps=20,scale=120:240:flags=lanczos,eq=saturation=2.5:gamma=0.8",
		"-q:v", "9",
		"-progress", "pipe:2",
		"-nostats",
		"-loglevel", "error",
		"video_1.05_20fps.mjpeg",
	}
	got := BuildArgs(in)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestBuildArgs_AdvancedOverridesFilterBlock(t *testing.T) {
	in := EncodingInput{
		InputPath:    "in.mp4",
		OutputPath:   "out.mjpeg",
		AdvancedArgs: `-vf "fps=15,scale=80:160" -q:v 5`,
	}
	want := []string{
		"-y",
		"-i", "in.mp4",
		"-vf", "fps=15,scale=80:160",
		"-q:v", "5",
		"-progress", "pipe:2",
		"-nostats",
		"-loglevel", "error",
		"out.mjpeg",
	}
	got := BuildArgs(in)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestBuildArgs_FactoryAllFour(t *testing.T) {
	cases := []struct {
		name string
		in   EncodingInput
		vf   string
		q    string
	}{
		{"0.96", EncodingInput{Width: 80, Height: 160, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
			"fps=20,scale=80:160:flags=lanczos,eq=saturation=2.5:gamma=0.8", "9"},
		{"1.05", EncodingInput{Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
			"fps=20,scale=120:240:flags=lanczos,eq=saturation=2.5:gamma=0.8", "9"},
		{"1.14", EncodingInput{Width: 135, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
			"fps=20,scale=135:240:flags=lanczos,eq=saturation=2.5:gamma=0.8", "9"},
		{"1.90", EncodingInput{Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
			"fps=20,scale=120:240:flags=lanczos,eq=saturation=2.5:gamma=0.8", "9"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.in.InputPath = "in.mp4"
			tc.in.OutputPath = "out.mjpeg"
			args := BuildArgs(tc.in)
			vf := argAfter(args, "-vf")
			q := argAfter(args, "-q:v")
			if vf != tc.vf {
				t.Errorf("vf got %q want %q", vf, tc.vf)
			}
			if q != tc.q {
				t.Errorf("q got %q want %q", q, tc.q)
			}
		})
	}
}

func argAfter(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
```

- [ ] **Step 2: Run tests — fail**

- [ ] **Step 3: Implement `internal/ffmpeg/command.go`**

```go
package ffmpeg

import (
	"fmt"

	"github.com/google/shlex"
)

type EncodingInput struct {
	InputPath    string
	OutputPath   string
	Width        int
	Height       int
	FPS          int
	Quality      int
	Saturation   float64
	Gamma        float64
	Scaler       string
	AdvancedArgs string // optional, replaces the filter+quality block
}

// BuildArgs returns the ffmpeg argument list (excluding the binary name itself).
func BuildArgs(in EncodingInput) []string {
	args := []string{"-y", "-i", in.InputPath}

	if in.AdvancedArgs != "" {
		mid, err := shlex.Split(in.AdvancedArgs)
		if err != nil {
			// Unparseable input: fall back to generated filter block.
			args = append(args, generatedFilterBlock(in)...)
		} else {
			args = append(args, mid...)
		}
	} else {
		args = append(args, generatedFilterBlock(in)...)
	}

	args = append(args, "-progress", "pipe:2", "-nostats", "-loglevel", "error", in.OutputPath)
	return args
}

func generatedFilterBlock(in EncodingInput) []string {
	vf := fmt.Sprintf("fps=%d,scale=%d:%d:flags=%s,eq=saturation=%s:gamma=%s",
		in.FPS, in.Width, in.Height, in.Scaler,
		formatFloat(in.Saturation), formatFloat(in.Gamma))
	return []string{"-vf", vf, "-q:v", fmt.Sprintf("%d", in.Quality)}
}

func formatFloat(f float64) string {
	// Trim trailing zeros: 2.5 → "2.5", 2.0 → "2", 0.8 → "0.8".
	s := fmt.Sprintf("%g", f)
	return s
}
```

- [ ] **Step 4: Add shlex dependency**

```bash
go get github.com/google/shlex
```

- [ ] **Step 5: Run tests — pass**

- [ ] **Step 6: Commit**

```bash
git add internal/ffmpeg/command.go internal/ffmpeg/command_test.go go.mod go.sum
git commit -m "ffmpeg: add BuildArgs() with snapshot tests for all 4 factory profiles"
```

### Task 10: `internal/ffmpeg/runner.go` — subprocess + progress

**Files:**
- Create: `internal/ffmpeg/runner.go`
- Create: `internal/ffmpeg/runner_windows.go`
- Create: `internal/ffmpeg/runner_unix.go`
- Test: `internal/ffmpeg/runner_test.go` (unit, with mock)
- Test: `internal/ffmpeg/runner_integration_test.go` (with build tag)

- [ ] **Step 1: Write unit test for progress parser**

`internal/ffmpeg/runner_test.go`:
```go
package ffmpeg

import "testing"

func TestParseProgressLine(t *testing.T) {
	cases := []struct {
		in       string
		key, val string
		ok       bool
	}{
		{"out_time_ms=12345", "out_time_ms", "12345", true},
		{"frame=42", "frame", "42", true},
		{"progress=continue", "progress", "continue", true},
		{"weird line without equals", "", "", false},
		{"", "", "", false},
	}
	for _, tc := range cases {
		k, v, ok := parseProgressLine(tc.in)
		if k != tc.key || v != tc.val || ok != tc.ok {
			t.Errorf("parseProgressLine(%q) = (%q,%q,%v) want (%q,%q,%v)",
				tc.in, k, v, ok, tc.key, tc.val, tc.ok)
		}
	}
}

func TestProgressRatio(t *testing.T) {
	if r := progressRatio(5_000_000, 10_000_000); r < 0.49 || r > 0.51 {
		t.Errorf("ratio %v not ~0.5", r)
	}
	if r := progressRatio(0, 0); r != 0 {
		t.Errorf("ratio for zero duration = %v, want 0", r)
	}
	if r := progressRatio(20_000_000, 10_000_000); r != 1.0 {
		t.Errorf("ratio capped at 1, got %v", r)
	}
}
```

- [ ] **Step 2: Implement `internal/ffmpeg/runner.go`**

```go
package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// ProgressUpdate is delivered while ffmpeg is encoding.
type ProgressUpdate struct {
	Ratio float64 // 0..1; 0 if total duration unknown
	Frame int     // most recent frame count (-1 if unknown)
}

// RunResult is delivered exactly once after the process exits.
type RunResult struct {
	Err      error
	Stderr   string // tail of stderr lines (last ~20)
	ExitCode int
}

// Run starts ffmpeg with args, parses progress, reports updates on out.
// totalDurationMicros is the source video's duration in microseconds (use 0 if unknown).
// Returns when the process ends, ctx is cancelled, or stdin is closed.
func Run(ctx context.Context, ffmpegPath string, args []string, totalDurationMicros int64, updates chan<- ProgressUpdate) RunResult {
	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	applyHideWindow(cmd) // platform-specific (windows: hide console)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return RunResult{Err: err, ExitCode: -1}
	}

	if err := cmd.Start(); err != nil {
		return RunResult{Err: err, ExitCode: -1}
	}

	tail := newRingTail(20)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := bufio.NewScanner(stderr)
		s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		var frame int = -1
		for s.Scan() {
			line := s.Text()
			tail.push(line)
			k, v, ok := parseProgressLine(line)
			if !ok {
				continue
			}
			switch k {
			case "frame":
				if n, err := strconv.Atoi(v); err == nil {
					frame = n
				}
			case "out_time_ms":
				if cur, err := strconv.ParseInt(v, 10, 64); err == nil {
					select {
					case updates <- ProgressUpdate{Ratio: progressRatio(cur, totalDurationMicros), Frame: frame}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	_ = io.Discard // (placeholder to keep import if compiler complains)

	waitErr := cmd.Wait()
	wg.Wait()
	close(updates)

	exitCode := 0
	if waitErr != nil {
		exitCode = -1
		if ee, ok := waitErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}

	return RunResult{
		Err:      waitErr,
		Stderr:   strings.Join(tail.snapshot(), "\n"),
		ExitCode: exitCode,
	}
}

func parseProgressLine(line string) (string, string, bool) {
	if line == "" {
		return "", "", false
	}
	i := strings.IndexByte(line, '=')
	if i <= 0 {
		return "", "", false
	}
	return line[:i], line[i+1:], true
}

func progressRatio(currentMicros, totalMicros int64) float64 {
	if totalMicros <= 0 {
		return 0
	}
	r := float64(currentMicros) / float64(totalMicros)
	if r > 1 {
		return 1
	}
	if r < 0 {
		return 0
	}
	return r
}

type ringTail struct {
	buf  []string
	max  int
	next int
	full bool
}

func newRingTail(n int) *ringTail { return &ringTail{buf: make([]string, n), max: n} }

func (r *ringTail) push(s string) {
	r.buf[r.next] = s
	r.next = (r.next + 1) % r.max
	if r.next == 0 {
		r.full = true
	}
}

func (r *ringTail) snapshot() []string {
	if !r.full {
		return append([]string(nil), r.buf[:r.next]...)
	}
	out := make([]string, 0, r.max)
	out = append(out, r.buf[r.next:]...)
	out = append(out, r.buf[:r.next]...)
	return out
}

// errors-package convenience
var _ = fmt.Errorf
```

- [ ] **Step 3: Implement platform stubs**

`internal/ffmpeg/runner_windows.go`:
```go
//go:build windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

func applyHideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}
```

`internal/ffmpeg/runner_unix.go`:
```go
//go:build !windows

package ffmpeg

import "os/exec"

func applyHideWindow(_ *exec.Cmd) {}
```

- [ ] **Step 4: Run unit tests — pass**

```bash
go test ./internal/ffmpeg/...
```

- [ ] **Step 5: Add integration test (skipped if ffmpeg missing)**

`internal/ffmpeg/runner_integration_test.go`:
```go
//go:build integration

package ffmpeg

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestRunner_RealFFmpeg_Roundtrip(t *testing.T) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not in PATH")
	}
	dir := t.TempDir()
	in := filepath.Join(dir, "in.mp4")
	// Generate a 1-second test pattern.
	gen := exec.Command(bin, "-y", "-f", "lavfi", "-i", "testsrc=duration=1:size=120x240:rate=20", in)
	gen.Stderr = os.Stderr
	if err := gen.Run(); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out := filepath.Join(dir, "out.mjpeg")
	args := BuildArgs(EncodingInput{
		InputPath: in, OutputPath: out,
		Width: 120, Height: 240, FPS: 20, Quality: 9,
		Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos",
	})
	updates := make(chan ProgressUpdate, 32)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res := Run(ctx, bin, args, 1_000_000, updates)
	if res.Err != nil {
		t.Fatalf("run: %v\nstderr:\n%s", res.Err, res.Stderr)
	}
	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("output: %v", err)
	}
	if st.Size() < 1000 {
		t.Errorf("output suspiciously small: %d bytes", st.Size())
	}
}
```

Run with: `go test -tags integration ./internal/ffmpeg/...` (skipped in normal `go test`).

- [ ] **Step 6: Commit**

```bash
git add internal/ffmpeg/runner.go internal/ffmpeg/runner_windows.go internal/ffmpeg/runner_unix.go internal/ffmpeg/runner_test.go internal/ffmpeg/runner_integration_test.go
git commit -m "ffmpeg: add Run() with stderr progress parsing and platform hide-window"
```

### Task 11: Probe video duration

**Files:**
- Modify: `internal/ffmpeg/runner.go` (add Probe function)
- Modify: `internal/ffmpeg/runner_test.go` (add unit test for parser only)
- Modify: `internal/ffmpeg/runner_integration_test.go` (test Probe)

ffmpeg's `-progress` reports `out_time_ms` but not the total. We probe duration up front so progress can be reported as a 0..1 ratio.

- [ ] **Step 1: Add unit test for parsing the duration line**

Append to `internal/ffmpeg/runner_test.go`:

```go
func TestParseDurationLine(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"  Duration: 00:01:23.45, start: 0.000000, bitrate: ...", int64((1*60+23)*1_000_000 + 450_000), true},
		{"  Duration: 00:00:01.00, ...", 1_000_000, true},
		{"random line", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseDurationLine(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("parseDurationLine(%q) = (%d, %v); want (%d, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
```

- [ ] **Step 2: Implement `Probe(ctx, ffmpegPath, inputPath) (durationMicros int64, err error)` in `runner.go`**

Append to `internal/ffmpeg/runner.go`:

```go
import (
	// existing imports kept
	"regexp"
)

var durationRE = regexp.MustCompile(`Duration:\s+(\d+):(\d+):(\d+)\.(\d+)`)

// Probe runs `ffmpeg -i <input>` and parses the Duration line from stderr.
// Returns 0 if the duration cannot be determined.
func Probe(ctx context.Context, ffmpegPath, inputPath string) (int64, error) {
	cmd := exec.CommandContext(ctx, ffmpegPath, "-i", inputPath)
	applyHideWindow(cmd)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	var dur int64
	s := bufio.NewScanner(stderr)
	for s.Scan() {
		if d, ok := parseDurationLine(s.Text()); ok {
			dur = d
		}
	}
	// ffmpeg returns non-zero when no output is specified — that is expected.
	_ = cmd.Wait()
	return dur, nil
}

func parseDurationLine(line string) (int64, bool) {
	m := durationRE.FindStringSubmatch(line)
	if m == nil {
		return 0, false
	}
	h, _ := strconv.Atoi(m[1])
	mi, _ := strconv.Atoi(m[2])
	s, _ := strconv.Atoi(m[3])
	frac, _ := strconv.Atoi(m[4])
	// Normalize frac to microseconds: ffmpeg emits 2-digit hundredths typically; pad.
	micros := frac
	for i := len(m[4]); i < 6; i++ {
		micros *= 10
	}
	for i := len(m[4]); i > 6; i-- {
		micros /= 10
	}
	total := int64(h*3600+mi*60+s)*1_000_000 + int64(micros)
	return total, true
}
```

- [ ] **Step 3: Add a Probe test in the integration suite**

Append to `internal/ffmpeg/runner_integration_test.go`:

```go
func TestRunner_Probe(t *testing.T) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not in PATH")
	}
	dir := t.TempDir()
	in := filepath.Join(dir, "in.mp4")
	gen := exec.Command(bin, "-y", "-f", "lavfi", "-i", "testsrc=duration=2:size=120x240:rate=20", in)
	if err := gen.Run(); err != nil {
		t.Fatal(err)
	}
	d, err := Probe(context.Background(), bin, in)
	if err != nil {
		t.Fatal(err)
	}
	// Allow ±0.2s.
	if d < 1_800_000 || d > 2_200_000 {
		t.Errorf("duration micros = %d, expected ~2_000_000", d)
	}
}
```

- [ ] **Step 4: Run unit tests — pass**

```bash
go test ./internal/ffmpeg/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/ffmpeg/runner.go internal/ffmpeg/runner_test.go internal/ffmpeg/runner_integration_test.go
git commit -m "ffmpeg: add Probe() to extract source duration for progress ratio"
```

---

## Phase 4: Job and Queue

### Task 12: `internal/job/encoding_config.go` — config + filename pattern resolver

**Files:**
- Create: `internal/job/encoding_config.go`
- Test: `internal/job/encoding_config_test.go`

- [ ] **Step 1: Write `encoding_config_test.go`**

```go
package job

import "testing"

func TestResolveOutputPath(t *testing.T) {
	cfg := EncodingConfig{
		ProfileNameForPattern: "1.05",
		Width: 120, Height: 240, FPS: 20,
		FilenamePattern: "{name}_{profile}_{fps}fps",
		OutputDir:       "/out",
	}
	got := cfg.ResolveOutputPath("/in/My Video.mp4")
	want := "/out/My Video_1.05_20fps.mjpeg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestResolveOutputPath_DefaultDirIsInputDir(t *testing.T) {
	cfg := EncodingConfig{
		ProfileNameForPattern: "1.90",
		Width: 120, Height: 240, FPS: 20,
		FilenamePattern: "{name}_{profile}",
		OutputDir:       "",
	}
	got := cfg.ResolveOutputPath("/in/clip.mp4")
	want := "/in/clip_1.90.mjpeg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestResolveOutputPath_AllPlaceholders(t *testing.T) {
	cfg := EncodingConfig{
		ProfileNameForPattern: "X",
		Width: 80, Height: 160, FPS: 25,
		FilenamePattern: "{name}-{w}x{h}-{fps}-{profile}",
	}
	got := cfg.ResolveOutputPath("/foo/bar.mov")
	want := "/foo/bar-80x160-25-X.mjpeg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestApplyOnExist_Suffix(t *testing.T) {
	got := applySuffixForCollision("/x/foo.mjpeg", func(p string) bool {
		return p == "/x/foo.mjpeg" || p == "/x/foo (1).mjpeg"
	})
	if got != "/x/foo (2).mjpeg" {
		t.Errorf("got %q want /x/foo (2).mjpeg", got)
	}
}
```

- [ ] **Step 2: Run tests — fail**

- [ ] **Step 3: Implement `internal/job/encoding_config.go`**

```go
package job

import (
	"fmt"
	"path/filepath"
	"strings"
)

type EncodingConfig struct {
	ProfileNameForPattern string  // for {profile}
	Width, Height         int
	FPS                   int
	Quality               int
	Saturation, Gamma     float64
	Scaler                string
	AdvancedArgs          string
	OutputDir             string // "" → input dir
	FilenamePattern       string
	OnExist               string // "overwrite" | "suffix" | "fail"
}

// ResolveOutputPath resolves {name}/{profile}/{fps}/{w}/{h} placeholders, sets ".mjpeg" extension.
func (c EncodingConfig) ResolveOutputPath(inputPath string) string {
	dir := c.OutputDir
	if dir == "" {
		dir = filepath.Dir(inputPath)
	}
	base := filepath.Base(inputPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	r := strings.NewReplacer(
		"{name}", name,
		"{profile}", c.ProfileNameForPattern,
		"{fps}", fmt.Sprintf("%d", c.FPS),
		"{w}", fmt.Sprintf("%d", c.Width),
		"{h}", fmt.Sprintf("%d", c.Height),
	)
	resolved := r.Replace(c.FilenamePattern)
	return filepath.Join(dir, resolved+".mjpeg")
}

// applySuffixForCollision returns a path with " (N)" appended before the extension
// where N is the smallest positive integer that does not collide. exists tells us
// whether a candidate path already exists (injectable for tests).
func applySuffixForCollision(path string, exists func(string) bool) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, i, ext))
		if !exists(candidate) {
			return candidate
		}
	}
	return path
}
```

- [ ] **Step 4: Run tests — pass**

- [ ] **Step 5: Commit**

```bash
git add internal/job/encoding_config.go internal/job/encoding_config_test.go
git commit -m "job: add EncodingConfig with output-path resolver and collision suffix"
```

### Task 13: `internal/job/job.go` and `internal/job/queue.go`

**Files:**
- Create: `internal/job/job.go`
- Create: `internal/job/queue.go`
- Test: `internal/job/queue_test.go`

- [ ] **Step 1: Write `queue_test.go` with a mock runner**

```go
package job

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeRunner struct {
	mu       sync.Mutex
	calls    []string // input paths
	failOn   map[string]error
	delay    time.Duration
}

func (f *fakeRunner) Run(ctx context.Context, j *Job, cfg EncodingConfig, onProgress func(float64)) error {
	f.mu.Lock()
	f.calls = append(f.calls, j.InputPath)
	wantErr, hasErr := f.failOn[j.InputPath]
	d := f.delay
	f.mu.Unlock()

	if d > 0 {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if hasErr {
		return wantErr
	}
	onProgress(1.0)
	return nil
}

func TestQueueSequentialAndSkipOnError(t *testing.T) {
	r := &fakeRunner{failOn: map[string]error{"b.mp4": errors.New("boom")}}
	q := NewQueue(r)
	q.Add(&Job{InputPath: "a.mp4"})
	q.Add(&Job{InputPath: "b.mp4"})
	q.Add(&Job{InputPath: "c.mp4"})

	updates := make(chan QueueEvent, 32)
	q.SubscribeEvents(updates)

	done := make(chan struct{})
	go func() {
		q.Run(context.Background(), EncodingConfig{})
		close(done)
	}()
	<-done

	jobs := q.Snapshot()
	if jobs[0].Status != StatusDone {
		t.Errorf("a: status=%v err=%q", jobs[0].Status, jobs[0].Error)
	}
	if jobs[1].Status != StatusFailed || jobs[1].Error == "" {
		t.Errorf("b: status=%v err=%q", jobs[1].Status, jobs[1].Error)
	}
	if jobs[2].Status != StatusDone {
		t.Errorf("c: status=%v err=%q", jobs[2].Status, jobs[2].Error)
	}

	if len(r.calls) != 3 || r.calls[0] != "a.mp4" || r.calls[1] != "b.mp4" || r.calls[2] != "c.mp4" {
		t.Errorf("calls = %v want [a b c]", r.calls)
	}
}

func TestQueueCancelStopsAfterCurrent(t *testing.T) {
	r := &fakeRunner{delay: 200 * time.Millisecond}
	q := NewQueue(r)
	q.Add(&Job{InputPath: "a.mp4"})
	q.Add(&Job{InputPath: "b.mp4"})

	ctx, cancel := context.WithCancel(context.Background())
	go q.Run(ctx, EncodingConfig{})
	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(300 * time.Millisecond)

	jobs := q.Snapshot()
	if jobs[0].Status != StatusCancelled && jobs[0].Status != StatusDone {
		t.Errorf("first job: status %v", jobs[0].Status)
	}
	if jobs[1].Status != StatusPending && jobs[1].Status != StatusCancelled {
		t.Errorf("second job: status %v (should not have run)", jobs[1].Status)
	}
}

func TestQueueCancelOne(t *testing.T) {
	r := &fakeRunner{delay: 200 * time.Millisecond}
	q := NewQueue(r)
	q.Add(&Job{InputPath: "a.mp4"})
	q.Add(&Job{InputPath: "b.mp4"})

	go q.Run(context.Background(), EncodingConfig{})
	time.Sleep(50 * time.Millisecond)
	q.CancelJob(q.Snapshot()[0].ID)
	time.Sleep(400 * time.Millisecond)

	jobs := q.Snapshot()
	if jobs[0].Status != StatusCancelled {
		t.Errorf("first job: status %v want cancelled", jobs[0].Status)
	}
	if jobs[1].Status != StatusDone {
		t.Errorf("second job: status %v want done (queue continues)", jobs[1].Status)
	}
}
```

- [ ] **Step 2: Implement `internal/job/job.go`**

```go
package job

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type JobStatus int

const (
	StatusPending JobStatus = iota
	StatusRunning
	StatusDone
	StatusFailed
	StatusCancelled
)

type Job struct {
	ID         string
	InputPath  string
	OutputPath string // set after EncodingConfig snapshot
	Status     JobStatus
	Progress   float64
	Error      string

	cancelMu sync.Mutex
	cancel   context.CancelFunc
}

func NewJob(input string) *Job {
	return &Job{ID: uuid.NewString(), InputPath: input, Status: StatusPending}
}

func (j *Job) setCancel(c context.CancelFunc) {
	j.cancelMu.Lock()
	defer j.cancelMu.Unlock()
	j.cancel = c
}

func (j *Job) Cancel() {
	j.cancelMu.Lock()
	c := j.cancel
	j.cancelMu.Unlock()
	if c != nil {
		c()
	}
}
```

- [ ] **Step 3: Implement `internal/job/queue.go`**

```go
package job

import (
	"context"
	"sync"
)

// Runner abstracts ffmpeg invocation so the queue can be tested with a fake.
type Runner interface {
	Run(ctx context.Context, j *Job, cfg EncodingConfig, onProgress func(float64)) error
}

type QueueEvent struct {
	JobID    string
	Status   JobStatus
	Progress float64
	Error    string
}

type Queue struct {
	mu       sync.Mutex
	jobs     []*Job
	runner   Runner
	events   []chan QueueEvent
}

func NewQueue(r Runner) *Queue { return &Queue{runner: r} }

func (q *Queue) Add(j *Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, j)
}

func (q *Queue) Snapshot() []Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Job, len(q.jobs))
	for i, j := range q.jobs {
		out[i] = Job{ID: j.ID, InputPath: j.InputPath, OutputPath: j.OutputPath, Status: j.Status, Progress: j.Progress, Error: j.Error}
	}
	return out
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = nil
}

func (q *Queue) SubscribeEvents(ch chan QueueEvent) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.events = append(q.events, ch)
}

func (q *Queue) emit(j *Job) {
	ev := QueueEvent{JobID: j.ID, Status: j.Status, Progress: j.Progress, Error: j.Error}
	q.mu.Lock()
	subs := append([]chan QueueEvent(nil), q.events...)
	q.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
			// drop if full
		}
	}
}

func (q *Queue) CancelJob(id string) {
	q.mu.Lock()
	for _, j := range q.jobs {
		if j.ID == id {
			q.mu.Unlock()
			j.Cancel()
			return
		}
	}
	q.mu.Unlock()
}

// Run processes pending jobs sequentially.
// Skips on error. Honors ctx cancellation between (and during) jobs.
func (q *Queue) Run(ctx context.Context, cfg EncodingConfig) {
	for {
		q.mu.Lock()
		var next *Job
		for _, j := range q.jobs {
			if j.Status == StatusPending {
				next = j
				break
			}
		}
		q.mu.Unlock()

		if next == nil {
			return
		}
		if ctx.Err() != nil {
			next.Status = StatusCancelled
			q.emit(next)
			return
		}

		jobCtx, jobCancel := context.WithCancel(ctx)
		next.setCancel(jobCancel)
		next.Status = StatusRunning
		q.emit(next)

		err := q.runner.Run(jobCtx, next, cfg, func(p float64) {
			next.Progress = p
			q.emit(next)
		})
		jobCancel()

		switch {
		case err == nil:
			next.Status = StatusDone
			next.Progress = 1
		case ctx.Err() != nil || jobCtx.Err() == context.Canceled:
			next.Status = StatusCancelled
		default:
			next.Status = StatusFailed
			next.Error = err.Error()
		}
		q.emit(next)
	}
}
```

- [ ] **Step 4: Add uuid dependency**

```bash
go get github.com/google/uuid
```

- [ ] **Step 5: Run tests — pass**

```bash
go test ./internal/job/...
```

- [ ] **Step 6: Commit**

```bash
git add internal/job/ go.mod go.sum
git commit -m "job: add Job, Queue, sequential run with skip-on-error and per-job cancel"
```

### Task 14: ffmpeg-backed Runner adapter

**Files:**
- Create: `internal/job/ffmpeg_runner.go`

The Queue uses an interface; this is the production implementation that bridges into `internal/ffmpeg`.

- [ ] **Step 1: Implement adapter**

```go
package job

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/webfraggle/mbd-video-converter/internal/ffmpeg"
)

type FFmpegRunner struct {
	BinaryPath string
}

func (r *FFmpegRunner) Run(ctx context.Context, j *Job, cfg EncodingConfig, onProgress func(float64)) error {
	out := cfg.ResolveOutputPath(j.InputPath)

	switch cfg.OnExist {
	case "fail":
		if _, err := os.Stat(out); err == nil {
			return fmt.Errorf("output already exists: %s", out)
		}
	case "suffix":
		out = applySuffixForCollision(out, func(p string) bool {
			_, err := os.Stat(p)
			return err == nil
		})
	default: // "overwrite" or unset
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	j.OutputPath = out

	dur, _ := ffmpeg.Probe(ctx, r.BinaryPath, j.InputPath)

	args := ffmpeg.BuildArgs(ffmpeg.EncodingInput{
		InputPath: j.InputPath, OutputPath: out,
		Width: cfg.Width, Height: cfg.Height, FPS: cfg.FPS, Quality: cfg.Quality,
		Saturation: cfg.Saturation, Gamma: cfg.Gamma, Scaler: cfg.Scaler,
		AdvancedArgs: cfg.AdvancedArgs,
	})

	updates := make(chan ffmpeg.ProgressUpdate, 8)
	go func() {
		for u := range updates {
			onProgress(u.Ratio)
		}
	}()

	res := ffmpeg.Run(ctx, r.BinaryPath, args, dur, updates)
	if res.Err != nil {
		// Clean up partial output on failure / cancellation.
		_ = os.Remove(out)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg exit %d: %s", res.ExitCode, lastLine(res.Stderr))
	}
	return nil
}

func lastLine(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\n' {
			return s[i+1:]
		}
	}
	return s
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/job/ffmpeg_runner.go
git commit -m "job: add FFmpegRunner adapter wiring queue to ffmpeg.Run/Probe"
```

---

### Task 14b: Logging package

**Files:**
- Create: `internal/logx/logx.go`

The app writes to `<UserConfigDir>/MBD-Videoconverter/debug.log`, rotated when it exceeds ~5 MB. ffmpeg arguments, exit codes, and stderr tails go here so support requests are diagnosable.

- [ ] **Step 1: Implement minimal rotating logger**

```go
package logx

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

const maxBytes = 5 * 1024 * 1024

// Setup opens (or rotates) the debug log next to settings.json and returns its path.
func Setup(appCfgDir string) (string, error) {
	if err := os.MkdirAll(appCfgDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(appCfgDir, "debug.log")
	if st, err := os.Stat(path); err == nil && st.Size() > maxBytes {
		_ = os.Rename(path, path+".1")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	log.SetOutput(io.MultiWriter(f, os.Stderr))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	return path, nil
}
```

- [ ] **Step 2: Hook into `ui.Run` early**

In `internal/ui/ui.go`, after `appCfgDir` is computed:

```go
import "github.com/webfraggle/mbd-video-converter/internal/logx"

// inside Run():
logPath, _ := logx.Setup(appCfgDir)
log.Printf("MBD-Videoconverter %s starting (log: %s)", version.Version, logPath)
```

- [ ] **Step 3: Log ffmpeg invocations and exit codes**

In `internal/job/ffmpeg_runner.go`, before `ffmpeg.Run`:

```go
log.Printf("ffmpeg: %s %v", r.BinaryPath, args)
```

After it returns, on error:

```go
log.Printf("ffmpeg: exit=%d err=%v\n%s", res.ExitCode, res.Err, res.Stderr)
```

(Add `"log"` import.)

- [ ] **Step 4: Commit**

```bash
git add internal/logx/ internal/ui/ui.go internal/job/ffmpeg_runner.go
git commit -m "logx: add rotating debug.log + log ffmpeg invocations and errors"
```

## Phase 5: UI — skeleton and panels

### Task 15: `internal/ui/ui.go` — main window with two-column layout

**Files:**
- Create: `internal/ui/ui.go`
- Modify: `main.go`

- [ ] **Step 1: Replace `main.go` to delegate to `ui.Run`**

```go
package main

import "github.com/webfraggle/mbd-video-converter/internal/ui"

func main() {
	ui.Run()
}
```

- [ ] **Step 2: Implement `internal/ui/ui.go` with two-column placeholders**

```go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/version"
)

func Run() {
	i18n.SetLanguage(i18n.DefaultLanguage())

	a := app.NewWithID("de.modellbahn-displays.mbd-videoconverter")
	w := a.NewWindow(i18n.T("app.title") + " " + version.Version)
	w.Resize(fyne.NewSize(1000, 640))

	// Placeholders; subsequent tasks fill these in.
	left := container.NewBorder(widget.NewLabel(i18n.T("queue.header.file")), nil, nil, nil, widget.NewLabel("(queue placeholder)"))
	right := container.NewBorder(widget.NewLabel(i18n.T("profile.header")), nil, nil, nil, widget.NewLabel("(profile panel placeholder)"))

	split := container.NewHSplit(left, right)
	split.SetOffset(0.62)

	settingsBtn := widget.NewButton(i18n.T("settings.title")+"…", func() {
		// task 22
	})
	header := container.NewBorder(nil, nil, nil, settingsBtn, widget.NewLabel(i18n.T("app.title")))

	w.SetContent(container.NewBorder(header, nil, nil, nil, split))
	w.ShowAndRun()
}
```

- [ ] **Step 3: Verify it runs**

```bash
go run .
```
Expected: A window with header (title + Settings button) and two-column split with placeholder labels.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/ui.go main.go
git commit -m "ui: add two-column main window skeleton with i18n title"
```

### Task 16: `internal/ui/profile_panel.go` — list + value editor (read-only display)

**Files:**
- Create: `internal/ui/profile_panel.go`
- Modify: `internal/ui/ui.go`

- [ ] **Step 1: Implement `profile_panel.go`**

```go
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/binding"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/profile"
)

type ProfilePanel struct {
	store    *profile.Store
	all      []profile.Profile
	selected int

	list   *widget.List
	widthE, heightE, fpsE, qualityE, satE, gammaE, scalerE *widget.Entry
	advE   *widget.Entry
	saveBtn, saveAsBtn, dupBtn, delBtn, newBtn *widget.Button
	root   *fyne.Container

	OnSelectionChanged func(p profile.Profile)
}

func NewProfilePanel(store *profile.Store) *ProfilePanel {
	pp := &ProfilePanel{store: store}
	all, _ := store.All()
	pp.all = all

	pp.list = widget.NewList(
		func() int { return len(pp.all) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			lbl := o.(*widget.Label)
			p := pp.all[i]
			prefix := ""
			if p.Factory {
				prefix = "🔒 "
			}
			lbl.SetText(prefix + p.Name)
		},
	)
	pp.list.OnSelected = func(i widget.ListItemID) {
		pp.selected = i
		pp.refreshFields()
		if pp.OnSelectionChanged != nil {
			pp.OnSelectionChanged(pp.all[i])
		}
	}

	pp.widthE = widget.NewEntry()
	pp.heightE = widget.NewEntry()
	pp.fpsE = widget.NewEntry()
	pp.qualityE = widget.NewEntry()
	pp.satE = widget.NewEntry()
	pp.gammaE = widget.NewEntry()
	pp.scalerE = widget.NewEntry()
	pp.advE = widget.NewMultiLineEntry()
	pp.advE.Wrapping = fyne.TextWrapWord

	form := widget.NewForm(
		widget.NewFormItem(i18n.T("profile.field.width"), pp.widthE),
		widget.NewFormItem(i18n.T("profile.field.height"), pp.heightE),
		widget.NewFormItem(i18n.T("profile.field.fps"), pp.fpsE),
		widget.NewFormItem(i18n.T("profile.field.quality"), pp.qualityE),
		widget.NewFormItem(i18n.T("profile.field.saturation"), pp.satE),
		widget.NewFormItem(i18n.T("profile.field.gamma"), pp.gammaE),
		widget.NewFormItem(i18n.T("profile.field.scaler"), pp.scalerE),
	)

	advAcc := widget.NewAccordion(
		widget.NewAccordionItem(i18n.T("profile.advanced"), pp.advE),
	)

	pp.newBtn = widget.NewButton(i18n.T("profile.btn.new"), pp.onNew)
	pp.dupBtn = widget.NewButton(i18n.T("profile.btn.dup"), pp.onDup)
	pp.delBtn = widget.NewButton(i18n.T("profile.btn.del"), pp.onDel)
	pp.saveBtn = widget.NewButton(i18n.T("profile.btn.save"), pp.onSave)
	pp.saveAsBtn = widget.NewButton(i18n.T("profile.btn.saveAs"), pp.onSaveAs)

	listButtons := container.NewGridWithColumns(3, pp.newBtn, pp.dupBtn, pp.delBtn)
	formButtons := container.NewGridWithColumns(2, pp.saveBtn, pp.saveAsBtn)

	pp.root = container.NewBorder(
		container.NewVBox(widget.NewLabel(i18n.T("profile.header")), pp.list, listButtons),
		formButtons,
		nil, nil,
		container.NewVBox(form, advAcc),
	)

	if len(pp.all) > 0 {
		pp.list.Select(0)
	}
	return pp
}

func (pp *ProfilePanel) Container() fyne.CanvasObject { return pp.root }

func (pp *ProfilePanel) refreshFields() {
	if pp.selected < 0 || pp.selected >= len(pp.all) {
		return
	}
	p := pp.all[pp.selected]
	pp.widthE.SetText(fmt.Sprintf("%d", p.Width))
	pp.heightE.SetText(fmt.Sprintf("%d", p.Height))
	pp.fpsE.SetText(fmt.Sprintf("%d", p.FPS))
	pp.qualityE.SetText(fmt.Sprintf("%d", p.Quality))
	pp.satE.SetText(fmt.Sprintf("%g", p.Saturation))
	pp.gammaE.SetText(fmt.Sprintf("%g", p.Gamma))
	pp.scalerE.SetText(p.Scaler)

	pp.saveBtn.Disable()
	if !p.Factory {
		pp.saveBtn.Enable()
	}
	pp.delBtn.Disable()
	if !p.Factory {
		pp.delBtn.Enable()
	}
}

// Selected returns the currently selected profile merged with editor edits as an EncodingConfig-friendly snapshot.
func (pp *ProfilePanel) Selected() profile.Profile {
	if pp.selected < 0 || pp.selected >= len(pp.all) {
		return profile.Profile{}
	}
	p := pp.all[pp.selected]
	if v, err := atoi(pp.widthE.Text); err == nil {
		p.Width = v
	}
	if v, err := atoi(pp.heightE.Text); err == nil {
		p.Height = v
	}
	if v, err := atoi(pp.fpsE.Text); err == nil {
		p.FPS = v
	}
	if v, err := atoi(pp.qualityE.Text); err == nil {
		p.Quality = v
	}
	if v, err := atof(pp.satE.Text); err == nil {
		p.Saturation = v
	}
	if v, err := atof(pp.gammaE.Text); err == nil {
		p.Gamma = v
	}
	if pp.scalerE.Text != "" {
		p.Scaler = pp.scalerE.Text
	}
	return p
}

func (pp *ProfilePanel) AdvancedArgs() string { return pp.advE.Text }

// Stub handlers — wired up in Task 17.
func (pp *ProfilePanel) onNew()    {}
func (pp *ProfilePanel) onDup()    {}
func (pp *ProfilePanel) onDel()    {}
func (pp *ProfilePanel) onSave()   {}
func (pp *ProfilePanel) onSaveAs() {}

// Compile guard for binding (kept around for future preview features).
var _ = binding.NewString
```

Add helper `atoi` / `atof` in a new small file `internal/ui/parse.go`:

```go
package ui

import "strconv"

func atoi(s string) (int, error) { return strconv.Atoi(s) }
func atof(s string) (float64, error) { return strconv.ParseFloat(s, 64) }
```

- [ ] **Step 2: Wire the panel into `ui.go` right column**

Modify `internal/ui/ui.go` to construct the panel:

```go
import (
	"path/filepath"
	"os"

	"github.com/webfraggle/mbd-video-converter/internal/profile"
)

// inside Run():
cfgDir, _ := os.UserConfigDir()
appCfgDir := filepath.Join(cfgDir, "MBD-Videoconverter")
profileStore := profile.NewStore(filepath.Join(appCfgDir, "profiles.json"))
profilePanel := NewProfilePanel(profileStore)
right := profilePanel.Container()
```

(Replace the placeholder `right` from Task 15.)

- [ ] **Step 3: Run app, verify profile list shows 4 factory entries with locks**

```bash
go run .
```

- [ ] **Step 4: Commit**

```bash
git add internal/ui/profile_panel.go internal/ui/parse.go internal/ui/ui.go
git commit -m "ui: add profile panel with list + value editor (read-only behavior)"
```

### Task 17: Profile-panel actions (New / Duplicate / Delete / Save / Save As)

**Files:**
- Modify: `internal/ui/profile_panel.go`

- [ ] **Step 1: Implement the five action handlers**

Replace the stubs at the bottom of `profile_panel.go`:

```go
import (
	// existing imports kept
	"fyne.io/fyne/v2/dialog"
	"github.com/google/uuid"
)

func (pp *ProfilePanel) reload(selectID string) {
	all, _ := pp.store.All()
	pp.all = all
	pp.list.Refresh()
	for i, p := range pp.all {
		if p.ID == selectID {
			pp.list.Select(i)
			return
		}
	}
	if len(pp.all) > 0 {
		pp.list.Select(0)
	}
}

func (pp *ProfilePanel) currentUserList() []profile.Profile {
	out := []profile.Profile{}
	for _, p := range pp.all {
		if !p.Factory {
			out = append(out, p)
		}
	}
	return out
}

func (pp *ProfilePanel) saveUserList(users []profile.Profile, selectID string) {
	if err := pp.store.Save(users); err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	pp.reload(selectID)
}

func (pp *ProfilePanel) onNew() {
	pp.promptName("", func(name string) {
		newP := profile.Profile{
			ID: "user:" + uuid.NewString(), Name: name,
			Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos",
		}
		users := append(pp.currentUserList(), newP)
		pp.saveUserList(users, newP.ID)
	})
}

func (pp *ProfilePanel) onDup() {
	if pp.selected < 0 {
		return
	}
	src := pp.all[pp.selected]
	pp.promptName(src.Name+" (Kopie)", func(name string) {
		dup := src
		dup.Factory = false
		dup.ID = "user:" + uuid.NewString()
		dup.Name = name
		users := append(pp.currentUserList(), dup)
		pp.saveUserList(users, dup.ID)
	})
}

func (pp *ProfilePanel) onDel() {
	if pp.selected < 0 {
		return
	}
	cur := pp.all[pp.selected]
	if cur.Factory {
		return
	}
	users := pp.currentUserList()
	out := users[:0]
	for _, p := range users {
		if p.ID != cur.ID {
			out = append(out, p)
		}
	}
	pp.saveUserList(out, "")
}

func (pp *ProfilePanel) onSave() {
	if pp.selected < 0 {
		return
	}
	cur := pp.Selected()
	if cur.Factory {
		return
	}
	users := pp.currentUserList()
	for i := range users {
		if users[i].ID == cur.ID {
			users[i] = cur
		}
	}
	pp.saveUserList(users, cur.ID)
}

func (pp *ProfilePanel) onSaveAs() {
	pp.promptName("Mein Profil", func(name string) {
		base := pp.Selected()
		base.Factory = false
		base.ID = "user:" + uuid.NewString()
		base.Name = name
		users := append(pp.currentUserList(), base)
		pp.saveUserList(users, base.ID)
	})
}

func (pp *ProfilePanel) promptName(initial string, ok func(string)) {
	entry := widget.NewEntry()
	entry.SetText(initial)
	dialog.ShowForm("Profilname", "OK", "Abbrechen",
		[]*widget.FormItem{widget.NewFormItem("Name", entry)},
		func(confirm bool) {
			if !confirm || entry.Text == "" {
				return
			}
			ok(entry.Text)
		},
		fyne.CurrentApp().Driver().AllWindows()[0],
	)
}
```

- [ ] **Step 2: Manual verification**

Run `go run .`, then:
1. Click `+ Neu`, give a name → new entry appears, can be selected.
2. Select a factory profile, click `Duplizieren` → user copy appears, fields editable.
3. Edit fields on the user profile, click `Speichern` → values persist after restart.
4. Click `Löschen` on the user profile → it disappears.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/profile_panel.go
git commit -m "ui: wire profile panel actions (new, duplicate, delete, save, save as)"
```

### Task 18: `internal/ui/queue_view.go` — table, file picker, output fields

**Files:**
- Create: `internal/ui/queue_view.go`
- Modify: `internal/ui/ui.go`

- [ ] **Step 1: Implement queue view (no drop yet, no convert wiring yet)**

```go
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/job"
)

type QueueView struct {
	jobs []*job.Job
	tbl  *widget.Table

	outDirEntry  *widget.Entry
	patternEntry *widget.Entry

	addBtn, clearBtn, cancelBtn, convertBtn *widget.Button
	root                                    *fyne.Container

	OnConvert    func()
	OnCancel     func()
	OnCancelJob  func(jobID string)
}

func NewQueueView() *QueueView {
	qv := &QueueView{}

	qv.tbl = widget.NewTable(
		func() (int, int) { return len(qv.jobs) + 1, 4 }, // +1 for header
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, c fyne.CanvasObject) {
			lbl := c.(*widget.Label)
			if id.Row == 0 {
				switch id.Col {
				case 0:
					lbl.SetText("#")
				case 1:
					lbl.SetText(i18n.T("queue.header.file"))
				case 2:
					lbl.SetText(i18n.T("queue.header.status"))
				case 3:
					lbl.SetText("")
				}
				return
			}
			j := qv.jobs[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(fmt.Sprintf("%d", id.Row))
			case 1:
				lbl.SetText(j.InputPath)
			case 2:
				lbl.SetText(statusLabel(j))
			case 3:
				lbl.SetText("✕")
			}
		},
	)
	qv.tbl.SetColumnWidth(0, 40)
	qv.tbl.SetColumnWidth(1, 360)
	qv.tbl.SetColumnWidth(2, 110)
	qv.tbl.SetColumnWidth(3, 30)
	qv.tbl.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && id.Col == 3 && qv.OnCancelJob != nil {
			qv.OnCancelJob(qv.jobs[id.Row-1].ID)
		}
		qv.tbl.UnselectAll()
	}

	qv.addBtn = widget.NewButton(i18n.T("queue.btn.add"), qv.onAdd)
	qv.clearBtn = widget.NewButton(i18n.T("queue.btn.clear"), func() {
		qv.jobs = nil
		qv.tbl.Refresh()
	})
	qv.cancelBtn = widget.NewButton(i18n.T("queue.btn.cancel"), func() {
		if qv.OnCancel != nil {
			qv.OnCancel()
		}
	})
	qv.convertBtn = widget.NewButton(i18n.T("queue.btn.convert"), func() {
		if qv.OnConvert != nil {
			qv.OnConvert()
		}
	})

	topButtons := container.NewHBox(qv.addBtn, qv.clearBtn)
	rightButtons := container.NewHBox(qv.cancelBtn, qv.convertBtn)
	bottomBar := container.NewBorder(nil, nil, topButtons, rightButtons)

	qv.outDirEntry = widget.NewEntry()
	qv.outDirEntry.SetPlaceHolder("(leer = neben Input)")
	qv.patternEntry = widget.NewEntry()
	qv.patternEntry.SetText("{name}_{profile}_{fps}fps")

	outDirRow := container.NewBorder(nil, nil, widget.NewLabel(i18n.T("output.label.dir")), widget.NewButton("…", qv.onPickOutDir), qv.outDirEntry)
	patternRow := container.NewBorder(nil, nil, widget.NewLabel(i18n.T("output.label.pattern")), nil, qv.patternEntry)

	qv.root = container.NewBorder(
		nil,
		container.NewVBox(bottomBar, outDirRow, patternRow),
		nil, nil,
		qv.tbl,
	)
	return qv
}

func (qv *QueueView) Container() fyne.CanvasObject { return qv.root }
func (qv *QueueView) AddJob(j *job.Job)             { qv.jobs = append(qv.jobs, j); qv.tbl.Refresh() }
func (qv *QueueView) Jobs() []*job.Job              { return qv.jobs }
func (qv *QueueView) OutputDir() string             { return qv.outDirEntry.Text }
func (qv *QueueView) FilenamePattern() string       { return qv.patternEntry.Text }

func statusLabel(j *job.Job) string {
	switch j.Status {
	case job.StatusPending:
		return i18n.T("queue.status.pending")
	case job.StatusRunning:
		if j.Progress > 0 {
			return fmt.Sprintf("%.0f%%", j.Progress*100)
		}
		return i18n.T("queue.status.running")
	case job.StatusDone:
		return i18n.T("queue.status.done")
	case job.StatusFailed:
		return i18n.T("queue.status.failed")
	case job.StatusCancelled:
		return i18n.T("queue.status.cancelled")
	}
	return ""
}

func (qv *QueueView) onPickOutDir() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		qv.outDirEntry.SetText(uri.Path())
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func (qv *QueueView) onAdd() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		qv.AddJob(job.NewJob(reader.URI().Path()))
	}, fyne.CurrentApp().Driver().AllWindows()[0])
	_ = storage.NewExtensionFileFilter
}
```

- [ ] **Step 2: Wire into `ui.go`**

Replace the placeholder `left` in `ui.go`:

```go
qv := NewQueueView()
left := qv.Container()
```

Hold references to `qv` and `profilePanel` in a struct for later wiring (Task 21).

- [ ] **Step 3: Manual verification**

`go run .`, click `+ Datei…`, pick a video → row appears in queue.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/queue_view.go internal/ui/ui.go
git commit -m "ui: add queue view with file picker, output dir + filename pattern"
```

### Task 19: Whole-window drag-and-drop with dim overlay

**Files:**
- Create: `internal/ui/dnd.go`
- Modify: `internal/ui/ui.go`

- [ ] **Step 1: Implement drop handler at the window level**

Fyne v2 exposes `fyne.Window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI))`. Dragover detection is platform-limited; we approximate by toggling the overlay around drop time.

`internal/ui/dnd.go`:
```go
package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/job"
)

var acceptedExt = map[string]struct{}{
	".mp4": {}, ".mov": {}, ".avi": {}, ".mkv": {}, ".webm": {}, ".m4v": {},
}

type DropOverlay struct {
	bg     *canvas.Rectangle
	label  *widget.Label
	root   *fyne.Container
}

func NewDropOverlay() *DropOverlay {
	bg := canvas.NewRectangle(color.NRGBA{R: 59, G: 110, B: 165, A: 80})
	lbl := widget.NewLabelWithStyle("⬇ "+i18n.T("dnd.overlay"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	d := &DropOverlay{
		bg:    bg,
		label: lbl,
		root:  container.NewStack(bg, container.NewCenter(lbl)),
	}
	d.root.Hide()
	return d
}

func (d *DropOverlay) Show() { d.root.Show() }
func (d *DropOverlay) Hide() { d.root.Hide() }
func (d *DropOverlay) CanvasObject() fyne.CanvasObject { return d.root }

// FilterAccepted picks only video URIs and returns the local paths.
func FilterAccepted(uris []fyne.URI) []string {
	out := make([]string, 0, len(uris))
	for _, u := range uris {
		if u == nil {
			continue
		}
		ext := strings.ToLower(u.Extension())
		if _, ok := acceptedExt[ext]; ok {
			out = append(out, u.Path())
		}
	}
	return out
}

// JobsFromPaths is a small helper used by both Add-File and Drop paths.
func JobsFromPaths(paths []string) []*job.Job {
	jobs := make([]*job.Job, 0, len(paths))
	for _, p := range paths {
		jobs = append(jobs, job.NewJob(p))
	}
	return jobs
}
```

- [ ] **Step 2: Wire the overlay and drop handler in `ui.go`**

```go
overlay := NewDropOverlay()

stack := container.NewStack(
	container.NewBorder(header, nil, nil, nil, split),
	overlay.CanvasObject(),
)
w.SetContent(stack)

w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
	defer overlay.Hide()
	for _, p := range FilterAccepted(uris) {
		qv.AddJob(job.NewJob(p))
	}
})
```

The overlay is hidden by default. Fyne's drop callback fires once per drop; we briefly show and immediately hide it. (Live dragover styling is handled platform-natively if Fyne adds an event for it; we can wire `SetOnDragEnter`/`SetOnDragLeave` here when available — for now, the overlay is a static visual cue user sees if they hover with a file in some platforms, or a momentary flash on drop.)

- [ ] **Step 3: Manual verification**

`go run .`, drag a `.mp4` onto the window → it appears in the queue.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/dnd.go internal/ui/ui.go
git commit -m "ui: window-wide drop handler with extension filter"
```

### Task 20: Wire Convert button → snapshot + queue Run

**Files:**
- Modify: `internal/ui/ui.go`

The settings store is also constructed here so its `LastProfileID` and ffmpeg path are available.

- [ ] **Step 1: Construct settings store + queue + runner in `ui.go`**

```go
import (
	"context"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2/dialog"

	"github.com/webfraggle/mbd-video-converter/internal/ffmpeg"
	"github.com/webfraggle/mbd-video-converter/internal/job"
	"github.com/webfraggle/mbd-video-converter/internal/settings"
)

// inside Run(), after profilePanel is built:
settingsStore := settings.New(filepath.Join(appCfgDir, "settings.json"))
appSettings, _ := settingsStore.Load()
i18n.SetLanguage(stringDefault(appSettings.Language, i18n.DefaultLanguage()))

bundleDir := executableDir() // helper added below
runQueueOnce := func() {
	bin, err := ffmpeg.Locate(appSettings.FFmpegPath, bundleDir)
	if err != nil {
		dialog.ShowError(fmt.Errorf("%s", i18n.T("err.ffmpegNotFound")), w)
		return
	}
	sel := profilePanel.Selected()
	cfg := job.EncodingConfig{
		ProfileNameForPattern: profileNameForPattern(sel),
		Width: sel.Width, Height: sel.Height, FPS: sel.FPS, Quality: sel.Quality,
		Saturation: sel.Saturation, Gamma: sel.Gamma, Scaler: sel.Scaler,
		AdvancedArgs:    profilePanel.AdvancedArgs(),
		OutputDir:       qv.OutputDir(),
		FilenamePattern: qv.FilenamePattern(),
		OnExist:         appSettings.OnExist,
	}
	q := job.NewQueue(&job.FFmpegRunner{BinaryPath: bin})
	for _, j := range qv.Jobs() {
		q.Add(j)
	}
	events := make(chan job.QueueEvent, 64)
	q.SubscribeEvents(events)

	ctx, cancel := context.WithCancel(context.Background())
	cancelHandle.Store(cancel)

	go q.Run(ctx, cfg)
	go func() {
		for range events {
			fyne.Do(func() { qv.tbl.Refresh() })
		}
	}()
}
qv.OnConvert = runQueueOnce
qv.OnCancel = func() {
	if c, ok := cancelHandle.Load().(context.CancelFunc); ok && c != nil {
		c()
	}
}
qv.OnCancelJob = func(id string) {
	for _, j := range qv.Jobs() {
		if j.ID == id {
			j.Cancel()
			return
		}
	}
}
```

Add helpers:

```go
import (
	"sync/atomic"
)

var cancelHandle atomic.Value

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

func profileNameForPattern(p profile.Profile) string {
	// "1.05\" Display" → "1.05"; "Custom Foo" → "Custom Foo".
	name := p.Name
	if i := strings.Index(name, "\""); i > 0 {
		return name[:i]
	}
	return name
}
```

For Fyne's main-thread dispatch, replace `fyne.Do(...)` with the version available in your Fyne version: `fyne.CurrentApp().Driver().DoFromGoroutine(...)` (Fyne v2.4+).

- [ ] **Step 2: Manual verification (with a real ffmpeg next to the binary)**

```bash
mkdir -p ./dist-test
cp $(which ffmpeg) ./dist-test/ffmpeg
go build -o ./dist-test/MBD-Videoconverter .
cd ./dist-test
./MBD-Videoconverter
```
Drop a `.mp4`, click Convert → output `.mjpeg` should appear next to it.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/ui.go
git commit -m "ui: wire Convert/Cancel — snapshot config and run queue with ffmpeg runner"
```

### Task 21: Settings dialog

**Files:**
- Create: `internal/ui/settings_dialog.go`
- Modify: `internal/ui/ui.go`

- [ ] **Step 1: Implement modal dialog**

```go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/settings"
	"github.com/webfraggle/mbd-video-converter/internal/version"
)

// logFolderPath returns the directory containing debug.log for display in the dialog.
// The actual path is set by Run() at startup; we read a package-local var here.
var currentLogFolder = ""

func SetLogFolder(p string) { currentLogFolder = p }
func logFolderPath() string { return currentLogFolder }

func ShowSettingsDialog(parent fyne.Window, store *settings.Store, current settings.Settings, onSaved func(settings.Settings)) {
	lang := widget.NewSelect([]string{"de", "en"}, nil)
	lang.SetSelected(stringDefault(current.Language, i18n.DefaultLanguage()))

	ffmpegPath := widget.NewEntry()
	ffmpegPath.SetText(current.FFmpegPath)
	ffmpegPath.SetPlaceHolder("(leer = beigelegtes ffmpeg)")

	defaultOutDir := widget.NewEntry()
	defaultOutDir.SetText(current.DefaultOutDir)

	pattern := widget.NewEntry()
	pattern.SetText(stringDefault(current.FilenamePattern, "{name}_{profile}_{fps}fps"))

	onExist := widget.NewSelect([]string{"overwrite", "suffix", "fail"}, nil)
	onExist.SetSelected(stringDefault(current.OnExist, "overwrite"))

	openLogBtn := widget.NewButton(i18n.T("settings.btn.openLog"), func() {
		// Best-effort: copy path to clipboard; user opens manually.
		fyne.CurrentApp().Clipboard().SetContent(logFolderPath())
	})
	versionLabel := widget.NewLabel("Version: " + version.Version)

	form := []*widget.FormItem{
		{Text: "Sprache", Widget: lang},
		{Text: "ffmpeg-Pfad", Widget: ffmpegPath},
		{Text: "Default Output", Widget: defaultOutDir},
		{Text: "Pattern", Widget: pattern},
		{Text: "Bei vorhandener Datei", Widget: onExist},
		{Text: "", Widget: openLogBtn},
		{Text: "", Widget: versionLabel},
	}

	dialog.ShowForm(i18n.T("settings.title"), "OK", "Abbrechen", form, func(ok bool) {
		if !ok {
			return
		}
		updated := settings.Settings{
			Language:        lang.Selected,
			FFmpegPath:      ffmpegPath.Text,
			DefaultOutDir:   defaultOutDir.Text,
			FilenamePattern: pattern.Text,
			OnExist:         onExist.Selected,
			LastProfileID:   current.LastProfileID,
		}
		if err := store.Save(updated); err != nil {
			dialog.ShowError(err, parent)
			return
		}
		i18n.SetLanguage(updated.Language)
		onSaved(updated)
	}, parent)
}
```

- [ ] **Step 2: Wire the Settings button in `ui.go`**

```go
SetLogFolder(filepath.Dir(logPath))
settingsBtn := widget.NewButton(i18n.T("settings.title")+"…", func() {
	ShowSettingsDialog(w, settingsStore, appSettings, func(updated settings.Settings) {
		appSettings = updated
		dialog.ShowInformation("Hinweis", "Sprachänderung wird beim Neustart wirksam.", w)
	})
})
```

- [ ] **Step 3: Manual verification**

Open dialog, change ffmpeg path / pattern, save. Verify `~/Library/Application Support/MBD-Videoconverter/settings.json` contains the new values.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/settings_dialog.go internal/ui/ui.go
git commit -m "ui: add settings dialog (language, ffmpeg path, defaults, on-exist)"
```

---

## Phase 6: Build, distribution, and manual test plan

### Task 22: `build.sh` — cross-compile + version bump + ZIP

**Files:**
- Create: `build.sh`

- [ ] **Step 1: Write build script**

```bash
#!/usr/bin/env bash
# Build MBD-Videoconverter for macOS arm64, macOS x64, Windows amd64.
#
# Requires:
#   macOS arm64:  Go 1.21+, Xcode CLT
#   macOS x64 + Windows: Docker + fyne-cross
#     go install github.com/fyne-io/fyne-cross@latest
#     go install fyne.io/fyne/v2/cmd/fyne@latest

set -euo pipefail
export PATH="$HOME/go/bin:$PATH"

# ── Version bump (auto-increment patch) ──────────────────────────────────────
VFILE=VERSION
CURRENT=$(cat "$VFILE")
MAJORMINOR=${CURRENT%.*}
PATCH=${CURRENT##*.}
NEXT="${MAJORMINOR}.$((PATCH + 1))"
echo "$NEXT" > "$VFILE"
VERSION=$NEXT
echo "Version: $VERSION"

LDFLAGS="-s -w -X github.com/webfraggle/mbd-video-converter/internal/version.Version=$VERSION"
APPID="de.modellbahn-displays.mbd-videoconverter"
OUTDIR="dist"
mkdir -p "$OUTDIR"

# ── Locate or download ffmpeg per platform ───────────────────────────────────
# These URLs need updating to current evermeet/BtbN releases. Replace placeholders.
FFMPEG_DARWIN_ARM64="${FFMPEG_DARWIN_ARM64:-/opt/homebrew/bin/ffmpeg}"
FFMPEG_DARWIN_X64="${FFMPEG_DARWIN_X64:-/usr/local/bin/ffmpeg}"
FFMPEG_WIN_AMD64="${FFMPEG_WIN_AMD64:-./build-deps/ffmpeg.exe}"

# ── macOS arm64 (native) ─────────────────────────────────────────────────────
echo "Building macOS arm64..."
ARM_BUNDLE="$OUTDIR/macos-arm64/MBD-Videoconverter.app"
rm -rf "$ARM_BUNDLE"
mkdir -p "$ARM_BUNDLE/Contents/MacOS" "$ARM_BUNDLE/Contents/Resources"
go build -ldflags "$LDFLAGS" -o "$ARM_BUNDLE/Contents/MacOS/MBD-Videoconverter" .
cp "$FFMPEG_DARWIN_ARM64" "$ARM_BUNDLE/Contents/MacOS/ffmpeg"
cat > "$ARM_BUNDLE/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>CFBundleExecutable</key><string>MBD-Videoconverter</string>
  <key>CFBundleIdentifier</key><string>$APPID</string>
  <key>CFBundleName</key><string>MBD-Videoconverter</string>
  <key>CFBundleShortVersionString</key><string>$VERSION</string>
  <key>CFBundlePackageType</key><string>APPL</string>
</dict></plist>
EOF
(cd "$OUTDIR/macos-arm64" && zip -ry "../MBD-Videoconverter-${VERSION}-macos-arm64.zip" "MBD-Videoconverter.app")

# ── macOS x64 + Windows via fyne-cross ───────────────────────────────────────
if command -v fyne-cross &>/dev/null && docker info &>/dev/null; then
  echo "Building macOS x64..."
  fyne-cross darwin -arch amd64 -app-id "$APPID" -ldflags "$LDFLAGS" -output MBD-Videoconverter
  X64_BUNDLE="$OUTDIR/macos-x64/MBD-Videoconverter.app"
  rm -rf "$X64_BUNDLE"
  mkdir -p "$OUTDIR/macos-x64"
  cp -R "fyne-cross/dist/darwin-amd64/MBD-Videoconverter.app" "$X64_BUNDLE"
  cp "$FFMPEG_DARWIN_X64" "$X64_BUNDLE/Contents/MacOS/ffmpeg"
  rm -rf fyne-cross/
  (cd "$OUTDIR/macos-x64" && zip -ry "../MBD-Videoconverter-${VERSION}-macos-x64.zip" "MBD-Videoconverter.app")

  echo "Building Windows amd64..."
  fyne-cross windows -arch amd64 -app-id "$APPID" -ldflags "$LDFLAGS" -output MBD-Videoconverter
  WIN_DIR="$OUTDIR/windows-amd64"
  rm -rf "$WIN_DIR" && mkdir -p "$WIN_DIR"
  cp fyne-cross/dist/windows-amd64/MBD-Videoconverter.exe "$WIN_DIR/"
  cp "$FFMPEG_WIN_AMD64" "$WIN_DIR/ffmpeg.exe"
  rm -rf fyne-cross/
  (cd "$OUTDIR" && zip -r "MBD-Videoconverter-${VERSION}-windows-amd64.zip" "windows-amd64")
else
  echo "fyne-cross or Docker missing — skipping macOS x64 and Windows builds."
fi

echo "Done. Artifacts in $OUTDIR/"
```

- [ ] **Step 2: chmod and dry-run on macOS arm64**

```bash
chmod +x build.sh
./build.sh
```
Expected: `dist/MBD-Videoconverter-v0.1.1-macos-arm64.zip` is produced.

- [ ] **Step 3: Smoke-test the produced ZIP**

```bash
unzip -o dist/MBD-Videoconverter-v0.1.1-macos-arm64.zip -d /tmp/mbdvc-test
open /tmp/mbdvc-test/MBD-Videoconverter.app
```
Verify the window opens and the version in the title is correct.

- [ ] **Step 4: Commit**

```bash
git add build.sh VERSION
git commit -m "Add build.sh: cross-compile + version-bump + ZIP for 3 targets"
```

### Task 23: Manual test plan document

**Files:**
- Create: `docs/manual-test-plan.md`

- [ ] **Step 1: Write checklist**

```markdown
# Manual Test Plan — MBD-Videoconverter

Run before each tagged release on macOS arm64 and Windows amd64.

## Smoke
- [ ] App launches without console window (Windows) and with correct title (macOS).
- [ ] Title bar shows `MBD-Videoconverter vX.Y.Z`.

## Profiles
- [ ] Four factory profiles visible with lock icon.
- [ ] Selecting each shows correct W/H/FPS/Quality/Saturation/Gamma/Scaler.
- [ ] Save button is disabled when a factory profile is selected.
- [ ] `+ Neu` creates a user profile that survives app restart.
- [ ] `Duplizieren` from a factory profile creates an editable copy.
- [ ] `Speichern` on a user profile updates `~/Library/Application Support/MBD-Videoconverter/profiles.json` (macOS) / `%APPDATA%\MBD-Videoconverter\profiles.json` (Windows).
- [ ] `Löschen` removes the user profile from disk.

## Conversion (one file)
- [ ] Drop a 1080p `.mp4` onto the window — appears in queue.
- [ ] Convert with factory `1.05` profile → output `.mjpeg` is produced next to input.
- [ ] Output filename matches `{name}_{profile}_{fps}fps.mjpeg` default.
- [ ] Status moves through `wartet` → `läuft` (with %) → `fertig`.

## Conversion (batch)
- [ ] Drop 3 files (use `+ Datei…` for one of them).
- [ ] Click `Konvertieren` → all three process sequentially.
- [ ] Cancel one mid-flight via the row's ✕ → that row becomes `abgebrochen`, queue continues with the next.
- [ ] Make one file unconvertible (e.g. a renamed text file) → status `fehlgeschlagen`, queue continues.

## Output options
- [ ] Set output dir → all jobs land there.
- [ ] Set pattern `{name}-{w}x{h}` → applied correctly.
- [ ] In Settings, set `OnExist=fail` → second run on same file errors.
- [ ] Set `OnExist=suffix` → new file gets ` (1)`.

## ffmpeg path
- [ ] Settings → set ffmpeg path to a non-existent file → conversion shows clear error, rest of queue still processable after fix.
- [ ] Settings → set ffmpeg path to a real alternative → that one is used (verify with `lsof`/Process Monitor).

## Language
- [ ] Settings → switch to English → after restart, all UI text is English.
- [ ] All keys present (no raw key strings like `queue.btn.add` shown).
```

- [ ] **Step 2: Commit**

```bash
git add docs/manual-test-plan.md
git commit -m "Add manual test plan checklist"
```

### Task 24: README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write user-facing README**

```markdown
# MBD-Videoconverter

Konvertiert Videos in das MJPEG-Format für ESP32-getriebene Modellbahn-Displays (Zugzielanzeiger, Bahnhofs-Displays …).

## Installation

Keine Installation nötig — das ZIP entpacken und die App starten:

- **macOS:** `MBD-Videoconverter.app` öffnen.
- **Windows:** `MBD-Videoconverter.exe` doppelklicken.

ffmpeg ist im ZIP enthalten und liegt direkt neben dem App-Binary. Es können keine externen Bibliotheken oder Installationen erforderlich sein.

## Bedienung

1. **Profil wählen** rechts in der Liste (vier Werks-Profile für die Display-Größen 0,96″, 1,05″, 1,14″, 1,90″).
2. **Werte anpassen** (optional) im Editor unter der Liste — Saturation/Gamma/FPS/Quality lassen sich für besondere Videos verändern. „Als neues Profil…" speichert die aktuellen Werte als eigenes Profil.
3. **Videos hineinziehen** — die ganze App ist Drop-Zone.
4. **„▶ Konvertieren"** klicken. Output-Dateien landen entweder neben dem Input oder in dem Ordner, den du im Output-Feld angibst.

Ein laufender Job kann jederzeit über die ✕-Schaltfläche in der Zeile abgebrochen werden, der Rest der Warteschlange läuft weiter.

## Eigene ffmpeg-Version

In den Einstellungen lässt sich ein eigener Pfad zu einer ffmpeg-Binary setzen. Bleibt das Feld leer, wird das mitgelieferte ffmpeg verwendet.

## Building from source

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
go install github.com/fyne-io/fyne-cross@latest
./build.sh
```
Erzeugt ZIPs unter `dist/`.

## Lizenz und ffmpeg

ffmpeg wird unter LGPL/GPL ausgeliefert; siehe `dist/<plattform>/LICENSE.ffmpeg`.
```

- [ ] **Step 2: Commit and push**

```bash
git add README.md
git commit -m "Add README with usage instructions"
git push origin main
```

---

## Self-Review Checklist (engineer runs at the end)

- [ ] All factory profiles match the spec table (use `go test ./internal/profile/...`).
- [ ] `BuildArgs` produces exactly the same `-vf` strings as the legacy scripts (compare against `scripts/convert_*.sh`).
- [ ] `i18n` parity test green.
- [ ] Drop, file picker, queue, cancel, convert all hooked up.
- [ ] Settings persists round-trip across app restart.
- [ ] `build.sh` produces a runnable macOS app and a runnable Windows folder.
- [ ] No console window flashes on Windows during conversion.
- [ ] `docs/manual-test-plan.md` walked through end to end.
