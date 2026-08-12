//go:build windows

package main

import (
	"errors"

	"golang.org/x/sys/windows"
)

const shutdownEventName = `Local\ClashTrafficMonitor.Shutdown`

func requestExistingShutdown() error {
	name, err := windows.UTF16PtrFromString(shutdownEventName)
	if err != nil {
		return err
	}
	handle, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, name)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return nil
	}
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	return windows.SetEvent(handle)
}

func watchShutdownRequests() (<-chan struct{}, func(), error) {
	name, err := windows.UTF16PtrFromString(shutdownEventName)
	if err != nil {
		return nil, func() {}, err
	}
	handle, err := windows.CreateEvent(nil, 0, 0, name)
	if err != nil {
		return nil, func() {}, err
	}
	requests := make(chan struct{}, 1)
	go func() {
		if status, _ := windows.WaitForSingleObject(handle, windows.INFINITE); status == windows.WAIT_OBJECT_0 {
			requests <- struct{}{}
		}
	}()
	closeWatcher := func() {
		_ = windows.SetEvent(handle)
		_ = windows.CloseHandle(handle)
	}
	return requests, closeWatcher, nil
}
