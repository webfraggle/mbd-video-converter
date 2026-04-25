# CLAUDE.md

Hinweise für Claude Code beim Arbeiten in diesem Repository.

## Projekt-Überblick

**MBD-Videoconverter** ist eine Desktop-Anwendung für Windows und macOS, die Videos in das MJPEG-Format konvertiert, das von ESP32-getriebenen Modellbahn-Displays (Zugzielanzeiger u.ä.) abgespielt wird. Sie ersetzt den bisherigen CLI-Workflow mit nackten ffmpeg-Aufrufen durch eine intuitive GUI.

**Status (2026-04-25):** v0.1.0 vollständig implementiert auf Branch `feat/v0.1.0-implementation` (32+ Commits). Alle 6 Phasen des Implementation-Plans abgeschlossen, Build-Artefakte für macOS arm64/x64 und Windows amd64 erzeugt. UI-Theme ist auf MBD-CI angepasst. Manueller Test (`docs/manual-test-plan.md`) und PR gegen `main` stehen noch aus.

## Stack

- **Sprache:** Go (Single-Binary)
- **GUI-Framework:** Fyne v2 (v2.7.x)
- **Cross-Compile:** `fyne-cross` (benötigt das alte `fyne.io/fyne/v2/cmd/fyne` CLI, nicht das neue `fyne.io/tools/cmd/fyne`)
- **Fonts:** Bigelow & Holmes Go-Schriftfamilie via `golang.org/x/image/font/gofont/*` — bewusst nicht der generische Fyne-Default Noto Sans
- **Distribution:** ZIP mit App-Binary + ffmpeg daneben (kein Installer). App-ID: `de.modellbahn-displays.mbd-videoconverter`.
- **Targets:** macOS arm64, macOS x64, Windows amd64.

## CI-Theme

Aus dem Designpass am 2026-04-25 (siehe `internal/ui/theme.go`):

| Rolle | Farbe | Verwendung |
|---|---|---|
| Surface | `#FFFFFF` (pures Weiß) | Hauptfläche, Input-Background |
| Body-Text | `#171717` (near-black) | Standard-Foreground |
| Primary-Accent | `#FD7014` Orange | Section-Headlines, aktive Selection, Focus-Ring, HighImportance-Buttons (Convert, Save), Hairline unter Sections |
| Secondary-Accent | `#037F8C` Teal | Subsection-Labels („ENCODING"), Hyperlinks, Success-States |

Der HoverColor und PressedColor müssen **halbtransparente schwarze Overlays** sein (`#000000` mit ~8 % bzw. ~16 % Alpha) — opake Werte überdecken das Orange auf HighImportance-Buttons komplett.

`SetTheme(...)` muss VOR der Erzeugung jeglicher Widgets passieren, sonst übernehmen die Widgets beim ersten Render den OS-Default-Variant (z. B. dunkles macOS).

## Geschwister-Projekte zum Mustervergleich

- `/Users/christoph/Documents/_Projects/Zugzielanzeiger/TrainController-GIT/` — Fyne-Pattern, `build.sh` mit fyne-cross, App-ID-Konvention.
- `/Users/christoph/Documents/_Projects/Zugzielanzeiger/zza-generate-images/` — Versionsmanagement via `VERSION`-Datei + `-ldflags -X`.

## Entscheidungen aus dem Brainstorming

Vollständig in `docs/superpowers/specs/2026-04-25-mbd-videoconverter-design.md`. Kurzfassung:

- **Workflow:** Multi-File-Batch-Queue, Drag&Drop oder File-Picker, keine Ordner.
- **Profile:** 4 Factory-Profile (read-only): 0.96" → 80×160, 1.05" → 120×240, 1.14" → 135×240, 1.90" → 120×240. User können eigene Profile zusätzlich anlegen, Factory nicht überschreiben.
- **Globales Profil:** ein aktives Profil gilt für die gesamte Queue (nicht pro Datei).
- **Settings exposed:** Width, Height, FPS, Quality, Saturation, Gamma, Scaler — plus collapsibler „Advanced"-Bereich für rohe ffmpeg-Args.
- **ffmpeg:** liegt als separate Binary neben der App im ZIP. Pfad in Settings überschreibbar.
- **Output:** User wählt Ordner (default: neben Input). Filename-Pattern konfigurierbar mit Platzhaltern `{name}` `{profile}` `{fps}` `{w}` `{h}`. Endung fest `.mjpeg`.
- **Queue:** sequenziell, skip-on-error, Cancel pro Job + global möglich.
- **Sprache:** Deutsch + Englisch mit Sprach-Umschalter (Neustart erforderlich).
- **Keine Vorschau:** Saturation/Gamma sind für Display-Ansicht optimiert, am Monitor unrepräsentativ.
- **App als Drop-Zone:** ganzes Fenster akzeptiert Drops.

## Wichtige Achtung-Punkte

- Die alten `scripts/convert_*.sh` haben einen Tippfehler: `convert_110_*.sh` ist eigentlich für 1.05", nicht 1.10". Das korrekte Display-Set ist 0.96", **1.05"**, 1.14", 1.90".
- ffmpeg-Subprozess unter Windows muss mit `CREATE_NO_WINDOW`/`HideWindow=true` gestartet werden, sonst poppt für jeden Aufruf ein Konsolenfenster (`internal/ffmpeg/runner_windows.go`).
- Fyne-UI-Updates aus Goroutines nur über `fyne.Do(...)` (verfügbar seit Fyne 2.6).
- **LSP/gopls-Diagnostics** in diesem Projekt sind häufig falsch — der Workspace wird unter `windows/amd64` analysiert, was zu „undefined symbol"- und „build constraints exclude all Go files in github.com/go-gl/gl"-Meldungen führt, obwohl `go build` und `go test` sauber durchlaufen. Vor dem Reagieren immer mit `go build ./... && go test ./...` verifizieren.
- **Theme-Hover muss translucent sein**, sonst wird das Orange auf HighImportance-Buttons beim Hover komplett überdeckt — siehe `internal/ui/theme.go`.

## Build und Run

Lokal entwickeln (Version-String ist „dev"):

```bash
go run .
```

Mit Versions-Injection lokal bauen:

```bash
go build -ldflags "-X github.com/webfraggle/mbd-video-converter/internal/version.Version=$(cat VERSION)" -o /tmp/mbdvc . && /tmp/mbdvc
```

Voll-Build aller Targets (bumpt automatisch Patch in `VERSION`):

```bash
./build.sh
```

Ergebnis liegt unter `dist/`. fyne-cross v1.6.1 hat einen Bug mit `-ldflags -X`; `build.sh` arbeitet mit einer temporär generierten `internal/version/version_gen.go` als Workaround.

## Workflow

Brainstorming → Design-Dokument → Implementation-Plan → Implementierung in Phasen mit Reviews. Standard-Superpowers-Pattern: writing-plans nach approved Spec, dann subagent-driven-development oder executing-plans.

## Repository

GitHub: https://github.com/webfraggle/mbd-video-converter
Aktueller Arbeits-Branch: `feat/v0.1.0-implementation`
