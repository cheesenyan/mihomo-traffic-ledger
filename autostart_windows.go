//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	autostartRegistryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	autostartValueName    = "ClashTrafficMonitor"
)

func platformReadAutostartEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, autostartRegistryPath, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open autostart registry key: %w", err)
	}
	defer key.Close()
	_, _, err = key.GetStringValue(autostartValueName)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read autostart registry value: %w", err)
	}
	return true, nil
}

func platformWriteAutostartEnabled(enabled bool) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, autostartRegistryPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open autostart registry key: %w", err)
	}
	defer key.Close()
	if !enabled {
		err := key.DeleteValue(autostartValueName)
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	command := `"` + strings.Trim(executable, `"`) + `"`
	if err := key.SetStringValue(autostartValueName, command); err != nil {
		return fmt.Errorf("write autostart registry value: %w", err)
	}
	return nil
}
