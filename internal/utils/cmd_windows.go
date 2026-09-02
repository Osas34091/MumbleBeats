//go:build windows

package utils

import (
	"os/exec"
	"syscall"
)

// HideWindow evita que se abra la ventana de consola negra en Windows (para ffmpeg y yt-dlp)
func HideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}
