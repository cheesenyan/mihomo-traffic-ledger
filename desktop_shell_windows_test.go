//go:build windows

package main

import (
	"context"
	"testing"
	"time"

	"github.com/gogpu/systray"
)

func TestWaitForTrayVisibilityRetriesUntilExplorerReportsBounds(t *testing.T) {
	shows := 0
	hides := 0
	show := func() *systray.SystemTray {
		shows++
		return nil
	}
	hide := func() *systray.SystemTray {
		hides++
		return nil
	}
	bounds := func() (int, int, int, int) {
		if shows < 3 {
			return 0, 0, 0, 0
		}
		return 10, 20, 16, 16
	}

	if err := waitForTrayVisibility(context.Background(), show, hide, bounds, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if shows != 3 || hides != 2 {
		t.Fatalf("shows=%d hides=%d want 3/2", shows, hides)
	}
}

func TestWaitForTrayVisibilityStopsOnShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	show := func() *systray.SystemTray { return nil }
	hide := func() *systray.SystemTray { return nil }
	bounds := func() (int, int, int, int) { return 0, 0, 0, 0 }

	if err := waitForTrayVisibility(ctx, show, hide, bounds, time.Hour); err == nil {
		t.Fatal("expected canceled context error")
	}
}
