# CLAUDE.md

**MBD-Videoconverter** — Desktop-Tool für Windows/macOS, das Videos in MJPEG für ESP32-Modellbahn-Displays konvertiert. Ersetzt einen alten CLI-ffmpeg-Workflow.

**Aktueller Stand (2026-04-26):** v0.1.0 fertig implementiert auf Branch `feat/v0.1.0-implementation`. PR gegen `main` und manueller Test stehen aus.

Repo: https://github.com/webfraggle/mbd-video-converter

## Stack

Go + Fyne v2, single binary, kein Installer. Cross-Compile via `fyne-cross` — braucht das **alte** `fyne.io/fyne/v2/cmd/fyne` CLI, **nicht** `fyne.io/tools/cmd/fyne`.

Targets: macOS arm64, macOS x64, Windows amd64. App-ID `de.modellbahn-displays.mbd-videoconverter`.

## Wichtige Gotchas

- **LSP-Rauschen:** gopls meldet in diesem Projekt häufig falsche `undefined symbol`-Fehler unter Windows-Cross-Analyse. Erst `go build ./... && go test ./...` prüfen, bevor reagiert wird.
- **Custom-Theme-Reihenfolge:** `app.Settings().SetTheme(...)` MUSS vor jeder Widget-Konstruktion stehen, sonst übernehmen die Widgets den OS-Default-Variant. `cHover`/`cPressed` müssen translucent sein (sonst überdecken sie HighImportance-Buttons).
- **Skript-Typo:** `scripts/convert_110_*.sh` ist eigentlich für **1.05"** (nicht 1.10"). Korrektes Display-Set: 0.96" / 1.05" / 1.14" / 1.90".
- **Windows-Subprozess:** ffmpeg muss mit `HideWindow=true`/`CREATE_NO_WINDOW` gestartet werden, sonst poppt für jeden Aufruf ein Konsolenfenster (`internal/ffmpeg/runner_windows.go`).
- **CI-Theme:** Weiß + Orange `#FD7014` (primary) + Teal `#037F8C` (secondary). Implementation in `internal/ui/theme.go`. Mehr Details in der Spec.

## Wo was steht

| | |
|---|---|
| Design-Spec | `docs/superpowers/specs/2026-04-25-mbd-videoconverter-design.md` |
| Implementation-Plan | `docs/superpowers/plans/2026-04-25-mbd-videoconverter-implementation.md` |
| Manueller QA-Plan | `docs/manual-test-plan.md` |
| Bedienung (User-facing) | `README.md` |
| Build-Voraussetzungen + ffmpeg-Quellen | Header von `build.sh` |

## Geschwister-Projekte zum Mustervergleich

- `~/Documents/_Projects/Zugzielanzeiger/TrainController-GIT/`
- `~/Documents/_Projects/Zugzielanzeiger/zza-generate-images/`

## Workflow

Brainstorming → Spec → Plan → Subagent-Driven-Development → Reviews → manueller Test.
