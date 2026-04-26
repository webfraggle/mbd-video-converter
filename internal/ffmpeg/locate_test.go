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
