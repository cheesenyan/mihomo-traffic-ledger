//go:build !windows

package main

import (
	"context"
	"fmt"
)

func openURL(string) error { return fmt.Errorf("desktop URL launch is unavailable") }

func runPlatformDesktopShell(ctx context.Context, requestQuit func()) error {
	<-ctx.Done()
	return nil
}
