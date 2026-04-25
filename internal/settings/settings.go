package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Settings struct {
	Language        string `json:"language"`
	FFmpegPath      string `json:"ffmpeg_path"`
	DefaultOutDir   string `json:"default_out_dir"`
	FilenamePattern string `json:"filename_pattern"`
	OnExist         string `json:"on_exist"` // "overwrite" | "suffix" | "fail"
	LastProfileID   string `json:"last_profile_id"`
}

func defaults() Settings {
	return Settings{
		Language:        "",
		FilenamePattern: "{name}_{profile}_{fps}fps",
		OnExist:         "overwrite",
	}
}

type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() (Settings, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults(), nil
	}
	if err != nil {
		return Settings{}, err
	}
	out := defaults()
	if err := json.Unmarshal(data, &out); err != nil {
		return Settings{}, err
	}
	return out, nil
}

func (s *Store) Save(v Settings) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
