//go:build !windows

package main

func requestExistingShutdown() error { return nil }

func watchShutdownRequests() (<-chan struct{}, func(), error) {
	return make(chan struct{}), func() {}, nil
}
