package profile

import (
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
