package main

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func userDataRoot() string {
	if localData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localData != "" {
		return filepath.Join(localData, "ClashTrafficMonitor")
	}
	return filepath.Join(".", "data")
}

func setupFileLogging() (func(), error) {
	logDir := filepath.Join(userDataRoot(), "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return func() {}, err
	}
	file, err := os.OpenFile(filepath.Join(logDir, "monitor.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return func() {}, err
	}
	output := io.MultiWriter(os.Stderr, file)
	log.SetOutput(output)
	// Third-party desktop components use slog for Explorer/taskbar recovery
	// diagnostics. Route those messages to the same persistent monitor log.
	slog.SetDefault(slog.New(slog.NewTextHandler(output, nil)))
	return func() { _ = file.Close() }, nil
}
