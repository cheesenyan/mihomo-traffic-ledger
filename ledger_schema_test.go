package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestConnectionMetadataDecodesLedgerFields(t *testing.T) {
	var payload connectionsResponse
	err := json.Unmarshal([]byte(`{
		"connections": [{
			"id": "c1",
			"start": "2026-08-12T21:13:57.9317737+08:00",
			"metadata": {
				"network": "tcp",
				"type": "HTTPS",
				"sourceIP": "127.0.0.1",
				"sourcePort": "52000",
				"destinationIP": "1.1.1.1",
				"destinationPort": "443",
				"host": "chatgpt.com",
				"process": "codex.exe",
				"processPath": "C:\\app\\codex.exe"
			},
			"chains": ["KR-01", "AI"],
			"providerChains": ["provider-ai", ""],
			"rule": "DomainSuffix",
			"rulePayload": "chatgpt.com",
			"upload": 11,
			"download": 22
		}]
	}`), &payload)
	if err != nil {
		t.Fatal(err)
	}
	got := payload.Connections[0]
	if got.Start == "" || got.Network != "tcp" || got.ConnType != "HTTPS" {
		t.Fatalf("missing connection metadata: %+v", got)
	}
	if got.ProcessPath != `C:\app\codex.exe` || got.SourcePort != "52000" || got.DestinationPort != "443" {
		t.Fatalf("missing process or port metadata: %+v", got)
	}
	if got.Rule != "DomainSuffix" || got.RulePayload != "chatgpt.com" {
		t.Fatalf("missing rule metadata: %+v", got)
	}
	if len(got.ProviderChains) != 2 || got.ProviderChains[0] != "provider-ai" {
		t.Fatalf("missing provider chain metadata: %+v", got)
	}
}

func TestLedgerSchemaMigration(t *testing.T) {
	db, err := openDatabase(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, table := range []string{"connection_sessions", "traffic_rollups"} {
		var name string
		if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name); err != nil {
			t.Fatalf("missing table %s: %v", table, err)
		}
	}

	for _, table := range []string{"traffic_aggregated", "connection_sessions", "traffic_rollups"} {
		columns := map[string]bool{}
		rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var cid, notNull, pk int
			var name, columnType string
			var defaultValue any
			if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			columns[name] = true
		}
		rows.Close()
		for _, name := range []string{"policy_group", "provider_chains"} {
			if !columns[name] {
				t.Fatalf("missing %s column %q", table, name)
			}
		}
	}
}

func TestRouteType(t *testing.T) {
	tests := []struct {
		chains []string
		want   string
	}{
		{[]string{"KR-01", "AI"}, "PROXY"},
		{[]string{"DIRECT"}, "DIRECT"},
		{[]string{"REJECT-DROP"}, "REJECT"},
		{nil, "PROXY"},
	}
	for _, test := range tests {
		if got := routeType(test.chains); got != test.want {
			t.Fatalf("routeType(%v)=%q want %q", test.chains, got, test.want)
		}
	}
}
