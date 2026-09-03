//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func runWithTray(startApp func(), stopApp func()) {
	startApp()

	// Mantener vivo el programa hasta que se reciba una señal de salida (Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	fmt.Println("\nApagando MumbleBeats...")
	stopApp()
}
