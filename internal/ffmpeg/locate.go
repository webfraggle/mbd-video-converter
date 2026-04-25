package ffmpeg

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// Locate returns the ffmpeg path to use.
// Order: userPath (if non-empty AND points to an executable file) → bundledDir/ffmpeg[.exe] → error.
func Locate(userPath, bundledDir string) (string, error) {
	if userPath != "" {
		if isExecutable(userPath) {
			return userPath, nil
		}
		// fall through to bundled
	}

	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}
	bundled := filepath.Join(bundledDir, name)
	if isExecutable(bundled) {
		return bundled, nil
	}
	return "", errors.New("ffmpeg binary not found (neither user setting nor bundled)")
}

func isExecutable(path string) bool {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true // Windows lacks +x bit; existence is enough.
	}
	return st.Mode().Perm()&0o111 != 0
}
