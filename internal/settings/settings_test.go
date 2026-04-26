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
