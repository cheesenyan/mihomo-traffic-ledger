//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"sync/atomic"
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
	for generation := 1; ; generation++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		tray := newPlatformTray(requestQuit)
		if err := waitForTrayVisibility(ctx, tray.Show, tray.Hide, tray.Bounds, 2*time.Second); err != nil {
			tray.Remove()
			return err
		}
		if generation > 1 {
			log.Printf("desktop tray restored with generation %d", generation)
		}

		watchCtx, stopWatch := context.WithCancel(ctx)
		var restart atomic.Bool
		var removeOnce sync.Once
		remove := func() { removeOnce.Do(tray.Remove) }
		go func() {
			<-watchCtx.Done()
			remove()
		}()
		go monitorTrayVisibility(watchCtx, tray.Bounds, 5*time.Second, 3, func() {
			restart.Store(true)
			log.Printf("desktop tray icon disappeared; rebuilding tray after Explorer restart")
			remove()
		})

		err := tray.Run()
		stopWatch()
		remove()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !restart.Load() {
			return err
		}
		if err != nil {
			log.Printf("desktop tray message loop ended during recovery: %v", err)
		}
	}
}

func newPlatformTray(requestQuit func()) *systray.SystemTray {
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
	return tray
}

func monitorTrayVisibility(
	ctx context.Context,
	bounds func() (int, int, int, int),
	checkInterval time.Duration,
	missingThreshold int,
	onMissing func(),
) {
	if missingThreshold < 1 {
		missingThreshold = 1
	}
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	missingChecks := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _, width, height := bounds()
			if width > 0 && height > 0 {
				missingChecks = 0
				continue
			}
			missingChecks++
			if missingChecks >= missingThreshold {
				onMissing()
				return
			}
		}
	}
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
