package main

import (
	"net/http"
	"path/filepath"
	"testing"
)

func TestResolveMihomoSettingsDefaultsToClashVergePipe(t *testing.T) {
	db, err := openDatabase(filepath.Join(t.TempDir(), "traffic.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	settings, err := resolveMihomoSettings(db, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if settings.URL != defaultMihomoEndpoint() {
		t.Fatalf("URL = %q, want %q", settings.URL, defaultMihomoEndpoint())
	}
}

func TestResolveMihomoTransportNamedPipe(t *testing.T) {
	base, client, err := resolveMihomoTransport(`npipe://./pipe/verge-mihomo`, nil)
	if err != nil {
		t.Fatalf("resolve named pipe transport: %v", err)
	}
	if base != "http://mihomo" {
		t.Fatalf("base = %q, want http://mihomo", base)
	}
	if client == nil || client.Transport == nil {
		t.Fatal("expected a named-pipe HTTP transport")
	}
}

func TestResolveMihomoTransportKeepsHTTPClient(t *testing.T) {
	fallback := &http.Client{}
	base, client, err := resolveMihomoTransport("http://127.0.0.1:9090/", fallback)
	if err != nil {
		t.Fatalf("resolve HTTP transport: %v", err)
	}
	if base != "http://127.0.0.1:9090" {
		t.Fatalf("base = %q", base)
	}
	if client != fallback {
		t.Fatal("HTTP endpoint should retain the configured client")
	}
}

func TestResolveMihomoTransportRejectsUnsupportedScheme(t *testing.T) {
	if _, _, err := resolveMihomoTransport("ftp://127.0.0.1", nil); err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}
