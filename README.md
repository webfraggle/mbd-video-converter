# MBD-Videoconverter

Konvertiert Videos in das MJPEG-Format für ESP32-getriebene Modellbahn-Displays.

## Installation

Keine Installation nötig — das ZIP entpacken und die App starten:

- **macOS:** `MBD-Videoconverter.app` öffnen.
- **Windows:** `MBD-Videoconverter.exe` doppelklicken.

ffmpeg liegt im ZIP direkt neben dem App-Binary. Es sind keine externen Bibliotheken oder zusätzliche Installationen erforderlich.

## Bedienung

1. **Profil wählen** rechts in der Liste — vier Werks-Profile für die Display-Größen 0,96″, 1,05″, 1,14″, 1,90″.
2. **Werte anpassen** (optional) im Editor unter der Liste — Saturation, Gamma, FPS, Quality lassen sich für besondere Videos verändern. „Als neues Profil…" speichert die aktuellen Werte als eigenes Profil.
3. **Videos hineinziehen** — die ganze App-Fläche ist Drop-Zone. Alternativ über den `+ Datei…`-Button auswählen.
4. **„▶ Konvertieren"** klicken. Output-Dateien landen entweder neben dem Input oder in dem Ordner, den du im Output-Feld angibst.

Ein laufender Job kann jederzeit über die ✕-Schaltfläche in der Zeile abgebrochen werden — der Rest der Warteschlange läuft weiter. Der globale „Abbrechen"-Button stoppt zusätzlich alle wartenden Jobs.

## Eigene ffmpeg-Version

In den Einstellungen lässt sich ein eigener Pfad zu einer ffmpeg-Binary setzen. Bleibt das Feld leer, wird das mitgelieferte ffmpeg verwendet.

## Output-Pattern

Der Dateiname wird über ein Pattern mit Platzhaltern gebildet. Default: `{name}_{profile}_{fps}fps`. Verfügbare Platzhalter:

| Platzhalter | Inhalt |
|---|---|
| `{name}` | Name der Input-Datei ohne Endung |
| `{profile}` | Profil-Name (z.B. `1.05`) |
| `{fps}` | Bildrate |
| `{w}` | Breite in Pixeln |
| `{h}` | Höhe in Pixeln |

Die Endung ist immer `.mjpeg`.

## Verhalten bei vorhandener Output-Datei

In den Einstellungen wählbar:

- **overwrite** (default) — bestehende Datei wird überschrieben
- **suffix** — neue Datei bekommt ein `(1)`/`(2)`/… angehängt
- **fail** — Job schlägt fehl, andere Jobs in der Queue laufen weiter

## Logging

Die App schreibt nach `<UserConfigDir>/MBD-Videoconverter/debug.log` (rotiert bei ~5 MB). Inhalt: ffmpeg-Argumente, Exit-Codes, Stderr-Tail bei Fehlern, App-Lifecycle. Hilft bei Support-Anfragen — der Pfad lässt sich über den „Log-Ordner öffnen"-Button in den Einstellungen kopieren.

## Konvertierungs-Profile

Werks-Profile (read-only):

| Display | Auflösung | FPS | Quality | Saturation | Gamma | Scaler |
|---|---|---|---|---|---|---|
| 0,96″ | 80 × 160 | 20 | 9 | 2.5 | 0.8 | lanczos |
| 1,05″ | 120 × 240 | 20 | 9 | 2.5 | 0.8 | lanczos |
| 1,14″ | 135 × 240 | 20 | 9 | 2.5 | 0.8 | lanczos |
| 1,90″ | 120 × 240 | 20 | 9 | 2.5 | 0.8 | lanczos |

Eigene Profile können zusätzlich angelegt werden (`+ Neu` oder `Duplizieren`) und werden in `<UserConfigDir>/MBD-Videoconverter/profiles.json` gespeichert.

## Building from source

Voraussetzungen:

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
go install github.com/fyne-io/fyne-cross@latest
```

Plus Docker für Cross-Compilation (macOS x64 + Windows). Dann:

```bash
./build.sh
```

Erzeugt ZIPs unter `dist/` für macOS arm64, macOS x64 und Windows amd64. Das Script bumpt automatisch den Patch-Teil in `VERSION` bei jedem Lauf.

## Lizenz

MBD-Videoconverter selbst steht unter der **MIT-Lizenz** (siehe [LICENSE](LICENSE)).

Das mitgelieferte ffmpeg-Binary wird unter LGPL ausgeliefert. Quellen:

- macOS: [evermeet.cx](https://evermeet.cx/ffmpeg/) (signed, statisch gelinkt)
- Windows: [BtbN/FFmpeg-Builds](https://github.com/BtbN/FFmpeg-Builds/releases) Release-Essentials, LGPL-Variante (statisch gelinkt)
