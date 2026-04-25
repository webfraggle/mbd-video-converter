package logx

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

const maxBytes = 5 * 1024 * 1024

// Setup opens (or rotates) the debug log next to settings.json and returns its path.
func Setup(appCfgDir string) (string, error) {
	if err := os.MkdirAll(appCfgDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(appCfgDir, "debug.log")
	if st, err := os.Stat(path); err == nil && st.Size() > maxBytes {
		_ = os.Rename(path, path+".1")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	log.SetOutput(io.MultiWriter(f, os.Stderr))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	return path, nil
}
