package main

import (
	"testing"
	"time"
)

func TestFlushAggregateEntriesBuildsPermanentCalendarRollups(t *testing.T) {
	svc := newTestService(t)
	when := time.Date(2026, 8, 12, 22, 47, 0, 0, time.Local)
	entry := aggregatedEntry{
		BucketStart: when.UnixMilli(), BucketEnd: when.Add(time.Minute).UnixMilli(),
		SourceIP: "127.0.0.1", Host: "api.openai.com", Process: "Codex.exe",
		ProcessPath: `C:\\Program Files\\Codex\\Codex.exe`, RouteType: "PROXY", Outbound: "韩国 AnyTLS",
		Chains: `["韩国 AnyTLS"]`, Rule: "DomainSuffix", RulePayload: "openai.com",
		Upload: 100, Download: 900, Count: 1,
	}
	svc.aggregateBuffer["test"] = &entry
	if err := svc.flushAggregateBuffer(); err != nil {
		t.Fatalf("flushAggregateBuffer: %v", err)
	}

	rows, err := svc.db.Query(`SELECT granularity, upload, download FROM traffic_rollups ORDER BY granularity`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[string][2]int64{}
	for rows.Next() {
		var grain string
		var upload, download int64
		if err := rows.Scan(&grain, &upload, &download); err != nil {
			t.Fatal(err)
		}
		got[grain] = [2]int64{upload, download}
	}
	for _, grain := range []string{"hour", "day", "week", "month"} {
		if got[grain] != [2]int64{100, 900} {
			t.Fatalf("%s rollup = %v", grain, got[grain])
		}
	}
}

func TestQueryTrafficRollupsGroupsByProcessAndNode(t *testing.T) {
	svc := newTestService(t)
	when := time.Date(2026, 8, 12, 22, 47, 0, 0, time.Local)
	entry := aggregatedEntry{BucketStart: when.UnixMilli(), BucketEnd: when.Add(time.Minute).UnixMilli(),
		Host: "api.openai.com", Process: "Codex.exe", RouteType: "PROXY", Outbound: "韩国 AnyTLS",
		Chains: `["韩国 AnyTLS"]`, Upload: 20, Download: 80, Count: 1}
	svc.aggregateBuffer["test"] = &entry
	if err := svc.flushAggregateBuffer(); err != nil {
		t.Fatal(err)
	}
	rows, err := svc.queryTrafficRollups("day", "process", when.Add(-time.Hour).UnixMilli(), when.Add(time.Hour).UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Label != "Codex.exe" || rows[0].Total != 100 {
		t.Fatalf("process rows = %+v", rows)
	}
	rows, err = svc.queryTrafficRollups("day", "outbound", when.Add(-time.Hour).UnixMilli(), when.Add(time.Hour).UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Label != "韩国 AnyTLS" {
		t.Fatalf("node rows = %+v", rows)
	}
}
