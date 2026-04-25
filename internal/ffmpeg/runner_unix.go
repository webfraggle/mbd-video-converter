//go:build !windows

package ffmpeg

import "os/exec"

func applyHideWindow(_ *exec.Cmd) {}
