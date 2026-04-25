# MBD-Videoconverter — Design

**Status:** Approved 2026-04-25
**Owner:** Christoph (chkdesign@gmail.com)
**Repo:** https://github.com/webfraggle/mbd-video-converter

## 1. Ziel und Kontext

Modellbahn-Displays mit ESP32-Controller spielen Videos im MJPEG-Format ab, weil der Controller für andere Formate keine Performance hat. Heute müssen Kunden ihre `.mp4`-Videos per CLI-Script (`scripts/convert_*.sh`) durch ffmpeg konvertieren — fehleranfällig und für Nicht-Techniker eine Hürde.

Diese App ersetzt den CLI-Workflow durch eine **Desktop-GUI für Windows und macOS** mit beigelegtem ffmpeg, vorkonfigurierten Display-Profilen und Settings, die der User anpassen oder als eigene Profile speichern kann.

**Nicht-Ziele (explizit ausgeklammert):**
- Live-Vorschau mit Slider — Saturation/Gamma sind für die Displays getuned, am Monitor unrepräsentativ.
- CLI-Modus — Power-User nutzen direkt das beigelegte ffmpeg.
- Folder-Watch / Auto-Convert — User dropt Dateien explizit.
- Per-Datei-unterschiedliche Profile in einer Batch — ein globales Profil gilt für die gesamte Queue.

## 2. Stack und Distribution

| | |
|---|---|
| Sprache | Go |
| GUI | Fyne v2 |
| Cross-Compile | `fyne-cross` (mit `fyne.io/fyne/v2/cmd/fyne` als Voraussetzung) |
| Targets | macOS arm64, macOS x64, Windows amd64 |
| App-ID | `de.modellbahn-displays.mbd-videoconverter` |
| Distribution | ZIP, keine Installation. App-Binary + ffmpeg-Binary nebeneinander. |
| ffmpeg-Quelle | macOS: evermeet.cx (signed). Windows: BtbN/gyan.dev release-essentials (statisch gelinkt). |
| Persistenz | JSON-Dateien in `os.UserConfigDir()` |

## 3. Display-Profile (Factory)

Aus den bestehenden Scripts abgeleitet, mit korrigierter Bezeichnung (`scripts/convert_110_*.sh` ist tatsächlich für 1.05", die Skript-Benennung war falsch):

| Profil | Width × Height | FPS | Quality | Saturation | Gamma | Scaler |
|---|---|---|---|---|---|---|
| 0.96" Display | 80 × 160 | 20 | 9 | 2.5 | 0.8 | lanczos |
| 1.05" Display | 120 × 240 | 20 | 9 | 2.5 | 0.8 | lanczos |
| 1.14" Display | 135 × 240 | 20 | 9 | 2.5 | 0.8 | lanczos |
| 1.90" Display | 120 × 240 | 20 | 9 | 2.5 | 0.8 | lanczos |

Factory-Profile sind read-only und in `internal/profile/factory.go` hardcoded. User kann zusätzlich eigene Profile anlegen (per „+ Neu" oder „Duplizieren" eines Factory-Profils); diese leben in `<UserConfigDir>/MBD-Videoconverter/profiles.json`.

## 4. UI-Layout

Ein einziges Hauptfenster, zweispaltig:

**Linke Spalte (~62%):**
- Warteschlange als Tabelle mit Spalten: Index, Datei, Status, ✕-Button.
- Buttons: „+ Datei…", „Liste leeren", „Abbrechen", „▶ Konvertieren".
- Output-Ordner-Feld (default leer = neben Input) mit Datei-Picker.
- Filename-Pattern-Feld (default `{name}_{profile}_{fps}fps`).

**Rechte Spalte (~38%):**
- Profil-Liste oben (Factory mit Schloss-Icon, dann User-Profile).
- Werte-Editor unten: Breite, Höhe, FPS, Quality, Saturation, Gamma, Scaler.
- Collapsible „Advanced — rohe ffmpeg-Args" (Textarea).
- Buttons: „+ Neu", „Duplizieren", „Löschen", „Speichern" (bei Factory disabled), „Als neues Profil…".

**Drag&Drop-Verhalten:**
- Das gesamte Anwendungsfenster ist Drop-Target für Dateien.
- Bei `dragover` wird der Inhalt abgedimmt (Opacity ≈ 0.18) und ein blau-gestricheltes Overlay „⬇ Videos hier loslassen" eingeblendet.
- Akzeptierte Endungen: `.mp4`, `.mov`, `.avi`, `.mkv`, `.webm` (Filterung beim Drop, unbekannte Endungen werden ignoriert mit kurzem Toast).

**„Einstellungen…" (Top-Right-Button) öffnet einen Modal-Dialog:**
- Sprache (Radio: Deutsch / English).
- ffmpeg-Pfad (Datei-Picker, leer = beigelegtes ffmpeg).
- Default-Output-Ordner und Default-Filename-Pattern.
- Verhalten bei vorhandener Output-Datei (Radio: überschreiben / Suffix anhängen / Job-Fehler). Default: überschreiben.
- „Log-Ordner öffnen"-Button.

## 5. Datenmodell

```go
// internal/profile/profile.go
type Profile struct {
    ID         string  // "factory:1.05" oder "user:<uuid>"
    Name       string
    Factory    bool
    Width      int
    Height     int
    FPS        int
    Quality    int     // 1..31, ffmpeg -q:v
    Saturation float64
    Gamma      float64
    Scaler     string  // "lanczos" default
}

// internal/settings/settings.go
type Settings struct {
    Language        string  // "de" | "en"
    FFmpegPath      string  // "" = bundled
    DefaultOutDir   string  // "" = neben Input
    FilenamePattern string  // default: "{name}_{profile}_{fps}fps"
    OnExist         string  // "overwrite" | "suffix" | "fail"
    LastProfileID   string  // zuletzt aktives Profil, beim Start wieder selektieren
}

// internal/job/job.go
type Job struct {
    ID        string
    InputPath string
    Status    JobStatus  // pending | running | done | failed | cancelled
    Progress  float64    // 0..1
    OutputPath string    // wird beim Snapshot-Bauen gesetzt
    Error     string
    cancel    context.CancelFunc  // intern, nicht serialisiert
}

// internal/job/encoding_config.go
// Snapshot, der beim Klick auf "Konvertieren" aus aktivem Profil + Editor-Werten + Output-Feldern eingefroren wird.
type EncodingConfig struct {
    ProfileNameForPattern string  // z.B. "1.05" für {profile}-Platzhalter
    Width, Height         int
    FPS                   int
    Quality               int
    Saturation, Gamma     float64
    Scaler                string
    AdvancedArgs          string  // optional, rohe Args als ein String, POSIX-Shell-split-Quoting (z.B. via google/shlex)
    OutputDir             string  // resolvierter Pfad
    FilenamePattern       string
    OnExist               string
}
```

**Persistenz:**
- `<UserConfigDir>/MBD-Videoconverter/settings.json` — eine Settings-Struct.
- `<UserConfigDir>/MBD-Videoconverter/profiles.json` — Liste der User-Profile (Factory wird nicht persistiert).

Beim Start: Factory-Liste + JSON-User-Liste mergen → Anzeige. Beim Speichern eines User-Profils wird nur die User-Liste zurückgeschrieben.

## 6. Datenfluss bei „▶ Konvertieren"

1. **Snapshot bauen** (`ui` → `job.NewEncodingConfig`): aktives Profil + aktuelle Editor-Werte + Output-Felder werden in einen `EncodingConfig` eingefroren. Diese eine Struct gilt für die gesamte laufende Batch. Spätere UI-Edits wirken nicht mehr.
2. **Queue starten** (`job/queue.go`): durchläuft Jobs mit Status `pending` sequenziell. Pro Job ein `context.Context` mit Cancel-Func — bildet Basis für ✕-Button und Global-Cancel.
3. **ffmpeg-Pfad ermitteln** (`ffmpeg/locate.go`):
   - User-Setting (wenn nicht-leer und Datei existiert+ausführbar)
   - sonst: `<dir(executable)>/ffmpeg[.exe]` (auf macOS: gleicher Ordner wie das Binary innerhalb der `.app`)
   - sonst: Fehler mit klarer Anweisung „Pfad in Einstellungen prüfen". Konvertieren wird vor Start blockiert (siehe Konfig-Fehler-Klasse).
4. **Output-Pfad auflösen:** Pattern + Platzhalter (`{name}`, `{profile}`, `{fps}`, `{w}`, `{h}`) → finaler Pfad. Wenn Datei existiert: Verhalten gemäß `OnExist`.
5. **Argument-Liste bauen** (`ffmpeg/command.go`):
   ```
   ffmpeg -y -i <input>
          -vf "fps=<FPS>,scale=<W>:<H>:flags=<scaler>,eq=saturation=<sat>:gamma=<gamma>"
          -q:v <quality>
          -progress pipe:2 -nostats -loglevel error
          <output>
   ```
   Wenn `AdvancedArgs` gefüllt ist, ersetzt das nur den Filter/Quality-Block — `-i <input>`, `-progress`/`-nostats`/`-loglevel` und `<output>` werden weiterhin von der App gesetzt; alles dazwischen kommt aus der Override.
6. **Subprozess starten** (`ffmpeg/runner.go`): `exec.CommandContext(ctx, ffmpegPath, args...)`. Plattform-spezifisch:
   - **Windows:** `SysProcAttr{ HideWindow: true, CreationFlags: CREATE_NO_WINDOW }`.
   - **macOS/Linux:** keine Sonderbehandlung.
   stderr wird gestreamt, `-progress pipe:2` Output zeilenweise geparst (`out_time_ms=`, `frame=`, …) → Progress-Channel an UI.
7. **UI-Update:** Channel-Werte werden über `fyne.CurrentApp().Driver().DoFromGoroutine(...)` auf den Fyne-Main-Thread gepusht.
8. **Job-Ende:** Exit 0 → `done`. Cancel → `cancelled`, ggf. Teil-Output-Datei löschen. Sonst → `failed` mit den letzten paar stderr-Zeilen als Error. Queue läuft mit nächstem Job weiter (skip-on-error).

**Beispiel-Argumentliste** (1.05" Default):
```
ffmpeg -y -i video.mp4
       -vf "fps=20,scale=120:240:flags=lanczos,eq=saturation=2.5:gamma=0.8"
       -q:v 9
       -progress pipe:2 -nostats -loglevel error
       video_1.05_20fps.mjpeg
```

## 7. Komponenten-Aufteilung (Verzeichnis)

```
MBD-Videoconverter/
├── main.go                      Entry: startet Fyne-UI
├── go.mod / go.sum
├── build.sh                     Cross-Compile (fyne-cross)
├── VERSION                      vX.Y.Z, Patch-Auto-Increment beim Build
├── Icon.png
├── README.md
├── internal/
│   ├── profile/                 Factory + User Profiles, JSON-Persistenz, Validierung
│   │   ├── factory.go
│   │   ├── store.go
│   │   └── profile.go
│   ├── settings/                App-Settings JSON-Persistenz
│   │   └── settings.go
│   ├── ffmpeg/                  ffmpeg-Aufruf-Schicht (kennt kein Fyne)
│   │   ├── locate.go
│   │   ├── command.go
│   │   └── runner.go
│   ├── job/                     Queue-Modell, EncodingConfig
│   │   ├── job.go
│   │   ├── encoding_config.go
│   │   └── queue.go
│   ├── i18n/                    Sprach-Strings
│   │   ├── i18n.go
│   │   ├── de.go
│   │   └── en.go
│   ├── version/                 Build-Version per ldflags
│   │   └── version.go
│   └── ui/                      Fyne-UI (einziger Konsument der anderen Pakete)
│       ├── ui.go                  Hauptfenster
│       ├── queue_view.go          Drop-Zone, Tabelle, Buttons
│       ├── profile_panel.go       Liste + Werte-Editor (rechte Spalte)
│       └── settings_dialog.go
├── docs/
│   ├── superpowers/specs/         Design-Dokumente (dieses Doc)
│   └── manual-test-plan.md        Checkliste für manuelles QA pro Release
└── dist/                          Build-Output (nicht im Repo)
```

**Boundary-Regeln:**
- `profile`, `settings`, `i18n`, `version` kennen nichts außer Standard-Library.
- `ffmpeg/` kennt kein Fyne — testbar mit echtem ffmpeg-Binary.
- `job/` orchestriert `profile + encoding_config + ffmpeg`, kennt keine UI.
- `ui/` ist der einzige Konsument aller anderen Pakete.

## 8. Fehlerbehandlung

| Klasse | Beispiel | Reaktion |
|---|---|---|
| Konfig-Fehler | ffmpeg nicht gefunden, Output-Ordner nicht beschreibbar | Dialog/Banner vor Queue-Start, „Konvertieren" disabled |
| Pro-Job-Fehler | ffmpeg Exit ≠ 0, Disk voll, Input nicht lesbar | Status `failed`, Error-Text aus letzten stderr-Zeilen, Queue läuft weiter |
| User-Cancel | ✕-Button, „Abbrechen" | Status `cancelled`, ffmpeg-Prozess via Context-Cancel beendet, Teil-Output-Datei gelöscht |
| Datei existiert | Output-Datei vorhanden | Verhalten aus `Settings.OnExist` |
| Profil-Validierung | W/H/FPS/Quality leer/außerhalb Range, Saturation/Gamma kein Float | Save/Convert disabled, Feld rot, Inline-Hinweis |

**Logging:** App schreibt nach `<UserConfigDir>/MBD-Videoconverter/debug.log` (rotated bei ~5 MB). Inhalt: ffmpeg-Args, Exit-Codes, stderr-Tail bei Fehlern, App-Lifecycle. „Log-Ordner öffnen"-Button im Einstellungen-Dialog.

## 9. Internationalisierung

- `internal/i18n/` mit zwei Maps `de` und `en`, Schlüssel-basiert (`i18n.T("queue.status.failed")`).
- Sprach-Wechsel im Einstellungen-Dialog setzt `Settings.Language` und löst UI-Refresh aus (oder fordert Neustart, falls Fyne-Refresh zu fummelig).
- Default-Sprache: System-Locale (`de` wenn Locale mit `de` beginnt, sonst `en`); überschreibbar.

## 10. Tests

| Paket | Tests |
|---|---|
| `internal/profile` | Unit: JSON-Roundtrip, Factory-Read-only-Schutz, Validierung, Merge-Logik. **Pflicht.** |
| `internal/ffmpeg/command` | Unit/Snapshot: Argumentliste pro Factory-Profil — muss exakt das ergeben, was die alten Scripts produzieren. |
| `internal/ffmpeg/locate` | Unit mit temp-Dirs: Reihenfolge User-Setting → Bundled → Fehler. |
| `internal/job/queue` | Unit mit gemocktem Runner: sequenziell, skip-on-error, Cancel propagiert. |
| `internal/ffmpeg/runner` | Integration (build-tag `integration`): Mini-Video durchwandeln, Output-Größe + Magic-Bytes prüfen. |
| `internal/i18n` | Unit: alle Keys in beiden Sprachen vorhanden. |
| `internal/ui` | Keine — manueller Test. |

**Manueller Test-Plan** in `docs/manual-test-plan.md` — wird mit dem Implementation-Plan zusammen erstellt. Pflicht-Punkte: alle 4 Factory-Profile, Drag&Drop, Cancel, ffmpeg-Pfad-Override, custom Profil anlegen + speichern + löschen, Sprach-Wechsel, alle drei `OnExist`-Optionen.

## 11. Versionierung und Build

- `VERSION`-Datei im Root, Format `vX.Y.Z`. Major.Minor manuell, Patch wird in `build.sh` auto-inkrementiert.
- Version per `-ldflags "-X github.com/webfraggle/mbd-video-converter/internal/version.Version=$(cat VERSION)"` ins Binary kompiliert.
- App zeigt Version im Einstellungen-Dialog unten.
- `build.sh` baut alle drei Targets, sammelt App-Binary + passendes ffmpeg in `dist/<target>/`, erstellt ZIPs.

## 12. Out of Scope für die erste Version

- macOS-Notarisierung und Windows-Code-Signing (separate Aufgabe nach erster funktionierender Version).
- Updater / Versionscheck.
- Help-System / Tutorial im UI (README reicht zunächst).
- Drag&Drop von URL oder kompletten Ordnern.
- Live-Preview oder Test-Render-Button.
