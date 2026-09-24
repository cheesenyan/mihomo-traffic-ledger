//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/gogpu/systray"
)

func openURL(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func openDataDirectory() error {
	return exec.Command("explorer.exe", userDataRoot()).Start()
}

func runPlatformDesktopShell(ctx context.Context, requestQuit func()) error {
	tray := systray.New()
	labels := desktopMenuLabels()
	menu := systray.NewMenu()
	menu.Add(labels[0], func() { _ = launchURL(dashboardURL) })
	menu.Add(labels[1], func() { _ = openDataDirectory() })
	menu.AddSeparator()
	menu.Add(labels[2], requestQuit)
	tray.SetIcon(trayIconPNG()).SetTooltip("Clash 软件流量账本").SetMenu(menu)
	tray.OnClick(func() { _ = launchURL(dashboardURL) })
	tray.OnDoubleClick(func() { _ = launchURL(dashboardURL) })
	if err := waitForTrayVisibility(ctx, tray.Show, tray.Hide, tray.Bounds, 2*time.Second); err != nil {
		tray.Remove()
		return err
	}
	go func() {
		<-ctx.Done()
		tray.Remove()
	}()
	return tray.Run()
}

func waitForTrayVisibility(
	ctx context.Context,
	show func() *systray.SystemTray,
	hide func() *systray.SystemTray,
	bounds func() (int, int, int, int),
	retryDelay time.Duration,
) error {
	for attempt := 1; ; attempt++ {
		show()
		_, _, width, height := bounds()
		if width > 0 && height > 0 {
			if attempt > 1 {
				log.Printf("desktop tray registered after %d attempts", attempt)
			}
			return nil
		}

		// gogpu/systray's fluent Show method does not expose the underlying
		// Shell_NotifyIcon error. Bounds is an independent Windows check that
		// proves whether Explorer actually accepted the icon.
		log.Printf("desktop tray registration attempt %d failed: Explorer did not report icon bounds; retrying in %s", attempt, retryDelay)
		hide()

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return fmt.Errorf("wait for desktop tray registration: %w", ctx.Err())
		case <-timer.C:
		}
	}
}
