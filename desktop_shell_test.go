package main

import (
	"bytes"
	"image/png"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDesktopMenuLabels(t *testing.T) {
	want := []string{"打开流量看板", "打开数据目录", "退出"}
	if got := desktopMenuLabels(); !reflect.DeepEqual(got, want) {
		t.Fatalf("desktopMenuLabels = %v, want %v", got, want)
	}
}

func TestInstallerCreatesStartMenuAndDesktopShortcuts(t *testing.T) {
	data, err := os.ReadFile("scripts/install.ps1")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	for _, want := range []string{"WScript.Shell", "StartMenu", "Desktop", "Clash 软件流量账本.lnk"} {
		if !strings.Contains(script, want) {
			t.Fatalf("installer missing %q", want)
		}
	}
	if !strings.Contains(script, "$null = $process.WaitForExit(5000)") {
		t.Fatal("installer should suppress the WaitForExit return value")
	}
}

func TestTrayIconIsValid32PixelPNG(t *testing.T) {
	data := trayIconPNG()
	image, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode tray icon: %v", err)
	}
	if image.Bounds().Dx() != 32 || image.Bounds().Dy() != 32 {
		t.Fatalf("icon bounds = %v", image.Bounds())
	}
}

func TestAlreadyRunningOpensDashboard(t *testing.T) {
	called := ""
	original := launchURL
	launchURL = func(url string) error {
		called = url
		return nil
	}
	t.Cleanup(func() { launchURL = original })

	if err := handleAlreadyRunning(true); err != nil {
		t.Fatal(err)
	}
	if called != dashboardURL {
		t.Fatalf("opened %q, want %q", called, dashboardURL)
	}
}
