//go:build !windows

package utils

import "os/exec"

// HideWindow no hace nada en sistemas Unix porque no tienen el problema de la ventana emergente.
func HideWindow(cmd *exec.Cmd) {
	// No op
}
