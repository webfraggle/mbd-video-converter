# CLAUDE.md

Hinweise für Claude Code beim Arbeiten in diesem Repository.

## Projekt-Überblick

**MBD-Videoconverter** ist eine Desktop-Anwendung für Windows und macOS, die Videos in das MJPEG-Format konvertiert, das von ESP32-getriebenen Modellbahn-Displays (Zugzielanzeiger u.ä.) abgespielt wird. Sie ersetzt den bisherigen CLI-Workflow mit nackten ffmpeg-Aufrufen durch eine intuitive GUI.

Status (2026-04-25): Brainstorming-Phase abgeschlossen, Design-Dokument liegt unter `docs/superpowers/specs/2026-04-25-mbd-videoconverter-design.md`. Implementierung steht aus.

## Stack

- **Sprache:** Go (Single-Binary, keine Installation nötig)
- **GUI-Framework:** Fyne v2
- **Cross-Compile:** `fyne-cross` (benötigt das alte `fyne.io/fyne/v2/cmd/fyne` CLI, nicht das neue `fyne.io/tools/cmd/fyne`)
- **Distribution:** ZIP mit App-Binary + ffmpeg daneben (kein Installer). App-ID: `de.modellbahn-displays.mbd-videoconverter`.
- **Targets:** macOS arm64, macOS x64, Windows amd64.

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
- **Sprache:** Deutsch + Englisch mit Sprach-Umschalter.
- **Keine Vorschau:** Saturation/Gamma sind für Display-Ansicht optimiert, am Monitor unrepräsentativ.
- **App als Drop-Zone:** ganzes Fenster akzeptiert Drops, dimmt Inhalt bei dragover.

## Wichtige Achtung-Punkte

- Die alten `scripts/convert_*.sh` haben einen Tippfehler: `convert_110_*.sh` ist eigentlich für 1.05", nicht 1.10". Das korrekte Display-Set ist 0.96", **1.05"**, 1.14", 1.90".
- ffmpeg-Subprozess unter Windows muss mit `CREATE_NO_WINDOW`/`HideWindow=true` gestartet werden, sonst poppt für jeden Aufruf ein Konsolenfenster.
- Fyne UI-Updates aus Goroutines nur über `fyne.CurrentApp().Driver().DoFromGoroutine(...)`.

## Workflow

Brainstorming → Design-Dokument → Implementation-Plan → Implementierung in Phasen mit Reviews. Standard-Superpowers-Pattern: writing-plans nach approved Spec, dann executing-plans.
