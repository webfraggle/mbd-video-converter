package ui

import (
	"os/exec"
	"runtime"
)

// openInFileManager asks the OS to reveal the given directory in its native
// file manager (Finder on macOS, Explorer on Windows, xdg-open on Linux).
// Returns an error if the launcher process couldn't be spawned; success of
// the actual reveal cannot be reliably observed here so callers should fall
// back gracefully (e.g. copy the path to the clipboard).
func openInFileManager(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
