package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() ([]Profile, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Profile
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	return out, nil
}

func (s *Store) Save(users []Profile) error {
	for _, p := range users {
		if p.Factory {
			return fmt.Errorf("refusing to persist factory profile %q", p.ID)
		}
		if err := p.Validate(); err != nil {
			return fmt.Errorf("invalid profile %q: %w", p.ID, err)
		}
	}
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// All returns factory profiles followed by user profiles.
func (s *Store) All() ([]Profile, error) {
	users, err := s.Load()
	if err != nil {
		return nil, err
	}
	all := append([]Profile(nil), Factory()...)
	all = append(all, users...)
	return all, nil
}
