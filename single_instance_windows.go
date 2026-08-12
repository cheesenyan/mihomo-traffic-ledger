//go:build windows

package main

import (
	"errors"

	"golang.org/x/sys/windows"
)

func acquireSingleInstance() (func(), bool, error) {
	name, err := windows.UTF16PtrFromString(`Local\ClashTrafficMonitor.SevenLedger`)
	if err != nil {
		return func() {}, false, err
	}
	handle, err := windows.CreateMutex(nil, false, name)
	if err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return func() {}, false, err
	}
	alreadyRunning := errors.Is(err, windows.ERROR_ALREADY_EXISTS)
	return func() { _ = windows.CloseHandle(handle) }, alreadyRunning, nil
}
