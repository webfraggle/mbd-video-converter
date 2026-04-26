# CLAUDE.md

**MBD-Videoconverter** — Desktop-Tool für Windows/macOS, das Videos in MJPEG für ESP32-Modellbahn-Displays konvertiert. Ersetzt einen alten CLI-ffmpeg-Workflow.

**Stand (2026-04-26):** v0.1.0 auf `main`. Repo: https://github.com/webfraggle/mbd-video-converter

## Stack

Go + Fyne v2, single binary. Cross-Compile via `fyne-cross` (braucht das **alte** `fyne.io/fyne/v2/cmd/fyne` CLI, nicht `fyne.io/tools/cmd/fyne`). Targets: macOS arm64, macOS x64, Windows amd64. App-ID `de.modellbahn-displays.mbd-videoconverter`.

## Gotchas

- gopls meldet hier oft Phantom-Fehler (`undefined symbol`, GL build-constraints unter windows/amd64) — vor dem Reagieren mit `go build ./... && go test ./...` verifizieren.
- `app.Settings().SetTheme(...)` MUSS vor jeder Widget-Konstruktion stehen, sonst übernehmen die Widgets den OS-Default-Variant. `cHover`/`cPressed` müssen translucente Overlays sein (sonst überdecken sie HighImportance-Buttons).
- `scripts/convert_110_*.sh` ist Tippfehler — gemeint ist **1.05"**. Display-Set: 0.96 / 1.05 / 1.14 / 1.90.
- ffmpeg-Subprozess unter Windows braucht `HideWindow=true`/`CREATE_NO_WINDOW` (`internal/ffmpeg/runner_windows.go`).
- CI-Theme: weiß + Orange `#FD7014` + Teal `#037F8C`. Implementation in `internal/ui/theme.go`.

## Doku

`docs/superpowers/specs/2026-04-25-mbd-videoconverter-design.md` (Spec), `docs/superpowers/plans/2026-04-25-mbd-videoconverter-implementation.md` (Plan), `docs/manual-test-plan.md` (QA-Checkliste), `README.md` (Bedienung), `build.sh`-Header (Build-Voraussetzungen + ffmpeg-Quellen).

## Geschwister-Projekte

`~/Documents/_Projects/Zugzielanzeiger/TrainController-GIT/` und `.../zza-generate-images/` — gleicher Fyne+fyne-cross-Pattern.

## Build / Release

`./build.sh` baut alle drei Targets. `./build.sh --release` zusätzlich Tag + Push + GitHub-Release als Latest. `--prerelease` markiert als Pre-release.
