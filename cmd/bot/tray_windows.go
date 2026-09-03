//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var trayIcon []byte

func runWithTray(startApp func(), stopApp func()) {
	onReady := func() {
		systray.SetIcon(trayIcon)
		systray.SetTitle("MumbleBeats")
		systray.SetTooltip("MumbleBeats is running")

		mDashboard := systray.AddMenuItem("Open Dashboard", "Open the web dashboard")
		mQuit := systray.AddMenuItem("Quit", "Stop the bot and exit")

		// Handle graceful shutdown from OS signals (Ctrl+C during dev)
		go func() {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
			<-sig
			fmt.Println("\nOS Signal received, quitting tray...")
			systray.Quit()
		}()

		// Handle menu clicks
		go func() {
			for {
				select {
				case <-mDashboard.ClickedCh:
					openBrowser("http://localhost:8080")
				case <-mQuit.ClickedCh:
					fmt.Println("Quit clicked in tray...")
					systray.Quit()
					return
				}
			}
		}()

		startApp()
	}

	onExit := func() {
		stopApp()
	}

	systray.Run(onReady, onExit)
}
