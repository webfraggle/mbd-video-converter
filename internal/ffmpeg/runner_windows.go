//go:build windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

func applyHideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}
