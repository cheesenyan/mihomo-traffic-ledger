//go:build !windows

package main

import "errors"

func platformReadAutostartEnabled() (bool, error) {
	return false, nil
}

func platformWriteAutostartEnabled(enabled bool) error {
	if enabled {
		return errors.New("autostart is only supported on Windows")
	}
	return nil
}
