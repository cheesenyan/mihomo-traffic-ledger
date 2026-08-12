package main

import (
	"path/filepath"
	"testing"
	"time"
)

func newLedgerTestService(t *testing.T) *service {
	t.Helper()
	db, err := openDatabase(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return &service{db: db, lastConnections: make(map[string]connection), activeSessionKeys: make(map[string]string)}
}

func responseWithConnection(id, start string, upload, download int64) *connectionsResponse {
	conn := connection{ID: id, Start: start, Upload: upload, Download: download, Chains: []string{"KR-01", "AI"}, Rule: "DomainSuffix", RulePayload: "chatgpt.com"}
	conn.Network = "tcp"
	conn.ConnType = "HTTPS"
	conn.SourcePort = "52000"
	conn.DestinationPort = "443"
	conn.ProcessPath = `C:\app\codex.exe`
	conn.Metadata.SourceIP = "127.0.0.1"
	conn.Metadata.Host = "chatgpt.com"
	conn.Metadata.DestinationIP = "1.1.1.1"
	conn.Metadata.Process = "codex.exe"
	return &connectionsResponse{Connections: []connection{conn}, UploadTotal: upload, DownloadTotal: download}
}

func logBytes(logs []trafficLog) int64 {
	var total int64
	for _, item := range logs {
		total += item.Upload + item.Download
	}
	return total
}

func TestPersistConnectionSnapshotTracksLifecycleWithoutDoubleCounting(t *testing.T) {
	svc := newLedgerTestService(t)
	start := "2026-08-12T21:13:57+08:00"
	logs, err := svc.persistConnectionSnapshot(time.Unix(100, 0), responseWithConnection("c1", start, 100, 200))
	if err != nil || logBytes(logs) != 300 {
		t.Fatalf("first snapshot: logs=%+v err=%v", logs, err)
	}
	logs, err = svc.persistConnectionSnapshot(time.Unix(101, 0), responseWithConnection("c1", start, 130, 260))
	if err != nil || logBytes(logs) != 90 {
		t.Fatalf("second snapshot: logs=%+v err=%v", logs, err)
	}
	logs, err = svc.persistConnectionSnapshot(time.Unix(102, 0), &connectionsResponse{UploadTotal: 130, DownloadTotal: 260})
	if err != nil || len(logs) != 0 {
		t.Fatalf("close snapshot: logs=%+v err=%v", logs, err)
	}

	var upload, download int64
	var endedAt *int64
	if err := svc.db.QueryRow(`SELECT upload, download, ended_at FROM connection_sessions WHERE connection_id='c1'`).Scan(&upload, &download, &endedAt); err != nil {
		t.Fatal(err)
	}
	if upload != 130 || download != 260 || endedAt == nil {
		t.Fatalf("session upload=%d download=%d ended=%v", upload, download, endedAt)
	}
}

func TestPersistConnectionSnapshotSeparatesReusedConnectionID(t *testing.T) {
	svc := newLedgerTestService(t)
	first := responseWithConnection("same", "2026-08-12T21:00:00+08:00", 10, 20)
	second := responseWithConnection("same", "2026-08-12T21:05:00+08:00", 7, 8)
	if _, err := svc.persistConnectionSnapshot(time.Unix(100, 0), first); err != nil {
		t.Fatal(err)
	}
	logs, err := svc.persistConnectionSnapshot(time.Unix(101, 0), second)
	if err != nil {
		t.Fatal(err)
	}
	if logBytes(logs) != 15 {
		t.Fatalf("reused ID delta=%d want 15", logBytes(logs))
	}
	var count int
	if err := svc.db.QueryRow(`SELECT COUNT(*) FROM connection_sessions WHERE connection_id='same'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("sessions=%d want 2", count)
	}
}

func TestPersistConnectionSnapshotCounterResetStartsFreshBaseline(t *testing.T) {
	svc := newLedgerTestService(t)
	start := "2026-08-12T21:00:00+08:00"
	if _, err := svc.persistConnectionSnapshot(time.Unix(100, 0), responseWithConnection("c1", start, 100, 200)); err != nil {
		t.Fatal(err)
	}
	reset := responseWithConnection("c2", "2026-08-12T21:10:00+08:00", 3, 4)
	reset.UploadTotal, reset.DownloadTotal = 3, 4
	logs, err := svc.persistConnectionSnapshot(time.Unix(101, 0), reset)
	if err != nil {
		t.Fatal(err)
	}
	if logBytes(logs) != 7 {
		t.Fatalf("reset delta=%d want 7", logBytes(logs))
	}
	var ended int
	if err := svc.db.QueryRow(`SELECT COUNT(*) FROM connection_sessions WHERE connection_id='c1' AND ended_at IS NOT NULL`).Scan(&ended); err != nil {
		t.Fatal(err)
	}
	if ended != 1 {
		t.Fatalf("ended reset sessions=%d want 1", ended)
	}
}

func TestPersistConnectionSnapshotDoesNotBackdatePreexistingConnection(t *testing.T) {
	svc := newLedgerTestService(t)
	monitorStart := time.Date(2026, 8, 12, 22, 0, 0, 0, time.Local)
	svc.monitorStartedAt = monitorStart.UnixMilli()
	preexisting := responseWithConnection("old", monitorStart.Add(-time.Hour).Format(time.RFC3339Nano), 1000, 2000)
	logs, err := svc.persistConnectionSnapshot(monitorStart, preexisting)
	if err != nil {
		t.Fatal(err)
	}
	if logBytes(logs) != 0 {
		t.Fatalf("preexisting connection delta=%d want 0", logBytes(logs))
	}
	preexisting.Connections[0].Upload += 10
	preexisting.Connections[0].Download += 20
	logs, err = svc.persistConnectionSnapshot(monitorStart.Add(time.Second), preexisting)
	if err != nil || logBytes(logs) != 30 {
		t.Fatalf("later growth logs=%+v err=%v", logs, err)
	}
}
