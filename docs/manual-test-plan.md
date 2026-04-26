# Manual Test Plan — MBD-Videoconverter

Run before each tagged release on macOS arm64 and Windows amd64. The integration tests in `internal/ffmpeg/runner_integration_test.go` cover the lower-level encoding path; this checklist covers UI behavior and end-to-end flows that aren't automatable.

## Smoke

- [ ] App launches without console window (Windows) and with correct title (macOS).
- [ ] Title bar shows `MBD-Videoconverter vX.Y.Z` matching the `VERSION` file used at build time.
- [ ] No crash, no error dialog on first launch with empty config.

## Profiles

- [ ] Four factory profiles visible with lock icon (🔒): `0.96"`, `1.05"`, `1.14"`, `1.90"`.
- [ ] Selecting each shows correct W/H/FPS/Quality/Saturation/Gamma/Scaler:
  - 0.96 → 80×160 / 20 / 9 / 2.5 / 0.8 / lanczos
  - 1.05 → 120×240 / 20 / 9 / 2.5 / 0.8 / lanczos
  - 1.14 → 135×240 / 20 / 9 / 2.5 / 0.8 / lanczos
  - 1.90 → 120×240 / 20 / 9 / 2.5 / 0.8 / lanczos
- [ ] `Speichern` button is disabled when a factory profile is selected.
- [ ] `Löschen` button is disabled when a factory profile is selected.
- [ ] `+ Neu` creates a user profile that survives app restart.
- [ ] `Duplizieren` from a factory profile creates an editable copy.
- [ ] `Speichern` on a user profile updates `~/Library/Application Support/MBD-Videoconverter/profiles.json` (macOS) / `%APPDATA%\MBD-Videoconverter\profiles.json` (Windows).
- [ ] `Löschen` removes the user profile from disk.
- [ ] `Als neues Profil…` while a factory profile is selected creates a user copy with the current editor values (including any in-place edits to W/H/etc).

## Conversion (single file)

- [ ] Drop a 1080p `.mp4` onto the window — appears in the queue.
- [ ] Convert with factory `1.05` profile selected → `.mjpeg` is produced next to input.
- [ ] Output filename matches `{name}_{profile}_{fps}fps.mjpeg` (the default pattern).
- [ ] Status moves through `wartet` → `läuft` (with %) → `fertig`.
- [ ] Output `.mjpeg` plays correctly on the actual ESP32 display when transferred.

## Conversion (batch)

- [ ] Drop 3 files plus add a 4th via `+ Datei…`.
- [ ] Click `▶ Konvertieren` → all four process sequentially in queue order.
- [ ] During run, click the ✕ on the currently-running row → that row becomes `abgebrochen`, queue continues with the next.
- [ ] During run, click the global `Abbrechen` button → current job stops, all subsequent jobs stay `wartet` and are not started.
- [ ] Make one file unconvertible (e.g. rename `notes.txt` to `notes.mp4`) → status `fehlgeschlagen`, queue continues with the next file.

## Output options

- [ ] Set output dir → all jobs land there (not next to input).
- [ ] Empty output dir → output lands next to input.
- [ ] Set pattern `{name}-{w}x{h}` → applied correctly (e.g. `clip-120x240.mjpeg`).
- [ ] Settings → `OnExist=fail` → second run on same input errors with clear message.
- [ ] Settings → `OnExist=suffix` → second run produces `name (1).mjpeg`, third produces `name (2).mjpeg`.
- [ ] Settings → `OnExist=overwrite` (default) → second run silently replaces.

## ffmpeg path override

- [ ] Settings → set ffmpeg path to a non-existent file → save closes; `▶ Konvertieren` shows error toast/dialog mentioning ffmpeg not found.
- [ ] Settings → clear the path → bundled ffmpeg is used; conversion works.
- [ ] Settings → set ffmpeg path to an alternate real ffmpeg → that one is used (verify in `debug.log` — it shows the path being invoked).

## Drag & drop

- [ ] Drag a `.mp4` over the app → file is accepted on drop.
- [ ] Drag a `.mov`, `.avi`, `.mkv`, `.webm`, `.m4v` → all accepted.
- [ ] Drag a `.txt` → silently ignored (not added to queue).
- [ ] Drag multiple files at once → all accepted videos appear in queue.

## Language

- [ ] Settings → switch to English, OK → restart-hint dialog appears.
- [ ] After app restart, all UI text is English (queue header, buttons, status labels, profile field labels).
- [ ] `de.json`/`en.json` keys all resolve — no raw key strings like `queue.btn.add` shown in UI.

## Logging

- [ ] After at least one conversion, `<UserConfigDir>/MBD-Videoconverter/debug.log` exists and contains:
  - App startup line with version
  - `ffmpeg: <path> [args]` line per job
  - `ffmpeg: exit=<code>` line on failure
- [ ] Settings → "Log-Ordner öffnen" copies the path to clipboard (paste into Finder/Explorer to confirm).

## Cancellation cleanup

- [ ] Cancel a job mid-flight → the partial `.mjpeg` output file is removed.
- [ ] Quit the app while queue is running → no zombie ffmpeg process remains (`ps aux | grep ffmpeg` on macOS, Task Manager on Windows).
