//go:build windows

package main

import (
	"context"
	"os/exec"

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
	menu.Add(labels[2], func() {
		requestQuit()
		tray.Remove()
	})
	tray.SetIcon(trayIconPNG()).SetTooltip("Clash 软件流量账本").SetMenu(menu)
	tray.OnClick(func() { _ = launchURL(dashboardURL) })
	tray.OnDoubleClick(func() { _ = launchURL(dashboardURL) })
	tray.Show()
	go func() {
		<-ctx.Done()
		tray.Remove()
	}()
	return tray.Run()
}
