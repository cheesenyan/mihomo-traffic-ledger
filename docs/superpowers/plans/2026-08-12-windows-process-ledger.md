# Windows Clash Process Ledger Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a single Windows tray EXE that permanently records every Mihomo connection observed after installation, attributes bytes to process and actual route, and exposes hour/day/week/month analysis.

**Architecture:** Extend the existing Go single-binary service. Keep the embedded static Web UI and SQLite store, add a transport abstraction for HTTP and Clash Verge's Windows named pipe, persist connection sessions plus permanent minute facts and rollups, then wrap the service in a Windows tray lifecycle.

**Tech Stack:** Go 1.25+, `database/sql` with pure-Go `modernc.org/sqlite`, `microsoft/go-winio`, `getlantern/systray`, `golang.org/x/sys/windows`, embedded HTML/CSS/JavaScript.

## Global Constraints

- Preserve the upstream single-binary architecture and embedded static frontend.
- Default Windows controller is `npipe://verge-mihomo`; HTTP Controller remains supported.
- Poll once per second and state that sub-second connections can be missed.
- Permanently retain observed connection sessions, minute facts, and hour/day/week/month rollups.
- Store proxy, DIRECT, and REJECT traffic separately; airport usage means proxy traffic only.
- Listen only on `127.0.0.1`; never expose the Mihomo secret to the frontend or logs.
- Store Windows data under `%LOCALAPPDATA%\ClashTrafficMonitor` and use per-user autostart without elevation.
- Do not change proxy rules, subscriptions, or selected nodes.
- Keep SQLite pure Go so local tests and Windows builds do not require GCC or MinGW.
- Every production behavior must follow a failing-test-first cycle.

---

### Task 1: Extended connection model and backward-compatible schema

**Files:**
- Modify: `main.go`
- Create: `ledger_schema_test.go`

**Interfaces:**
- Produces: expanded `connection.Metadata`, `routeType([]string) string`, and SQLite tables `connection_sessions` and `traffic_rollups`.
- Consumes: existing `openDatabase`, `connection`, and `traffic_aggregated` schema.

- [ ] **Step 1: Write failing model and migration tests**

Add tests which decode a representative Mihomo payload containing `processPath`, ports, network, type, rule and rule payload, then assert all fields are retained. Open an in-memory database through `openDatabase` and assert `connection_sessions`, `traffic_rollups`, and the new minute-fact columns exist.

```go
func TestConnectionMetadataDecodesLedgerFields(t *testing.T) {
	var payload connectionsResponse
	err := json.Unmarshal([]byte(`{"connections":[{"id":"c1","metadata":{"network":"tcp","type":"HTTPS","sourceIP":"127.0.0.1","sourcePort":"52000","destinationIP":"1.1.1.1","destinationPort":"443","host":"chatgpt.com","process":"codex.exe","processPath":"C:\\\\app\\\\codex.exe"},"chains":["KR-01","AI"],"rule":"DomainSuffix","rulePayload":"chatgpt.com","upload":11,"download":22}]}`), &payload)
	if err != nil { t.Fatal(err) }
	got := payload.Connections[0]
	if got.Metadata.ProcessPath == "" || got.Rule != "DomainSuffix" { t.Fatalf("missing ledger fields: %+v", got) }
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestConnectionMetadataDecodesLedgerFields|TestLedgerSchemaMigration'`

Expected: compilation failure because the new metadata fields and tables do not exist.

- [ ] **Step 3: Add metadata fields and idempotent schema**

Extend `connection` with `Start`, `Rule`, `RulePayload`, metadata network/type/ports/processPath. Add `connection_sessions` with a unique `(session_key)` and all fields from the design. Add `traffic_rollups` with unique key `(granularity,bucket_start,process,process_path,host,destination_ip,route_type,outbound,chains,rule,rule_payload)`. Add missing columns to `traffic_aggregated` using duplicate-column-tolerant migrations.

Implement:

```go
func routeType(chains []string) string {
	for _, chain := range chains {
		switch strings.ToUpper(strings.TrimSpace(chain)) {
		case "REJECT", "REJECT-DROP": return "REJECT"
		case "DIRECT": return "DIRECT"
		}
	}
	return "PROXY"
}
```

- [ ] **Step 4: Run focused and full tests**

Run: `go test ./... -run 'TestConnectionMetadataDecodesLedgerFields|TestLedgerSchemaMigration|TestRouteType'`

Expected: PASS.

Run: `go test ./...`

Expected: all upstream and new tests PASS.

- [ ] **Step 5: Commit**

```powershell
git add main.go ledger_schema_test.go
git commit -m "功能：扩展连接账本数据模型"
```

### Task 2: Permanent connection-session ledger and crash-safe deltas

**Files:**
- Create: `connection_ledger.go`
- Create: `connection_ledger_test.go`
- Modify: `main.go`

**Interfaces:**
- Consumes: `connection`, `connectionsResponse`, `routeType`, `service.db`, `service.lastConnections`.
- Produces: `(*service).persistConnectionSnapshot(now time.Time, payload *connectionsResponse) ([]trafficLog, error)` and `(*service).closeObservedSessions(now time.Time) error`.

- [ ] **Step 1: Write failing lifecycle tests**

Cover first observation, counter increase, disappearance, ID reuse with a different Mihomo start time, and counter reset. Assert one session row receives final cumulative bytes while returned `trafficLog` values contain only positive deltas.

```go
func TestPersistConnectionSnapshotTracksLifecycleWithoutDoubleCounting(t *testing.T) {
	svc := newLedgerTestService(t)
	first := responseWithConnection("c1", 100, 200)
	logs, err := svc.persistConnectionSnapshot(time.Unix(100, 0), first)
	if err != nil || sumLogs(logs) != 300 { t.Fatalf("first snapshot: logs=%+v err=%v", logs, err) }
	second := responseWithConnection("c1", 130, 260)
	logs, err = svc.persistConnectionSnapshot(time.Unix(101, 0), second)
	if err != nil || sumLogs(logs) != 90 { t.Fatalf("second snapshot: logs=%+v err=%v", logs, err) }
	_, err = svc.persistConnectionSnapshot(time.Unix(102, 0), &connectionsResponse{})
	if err != nil { t.Fatal(err) }
	assertSession(t, svc.db, "c1", 130, 260, true)
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestPersistConnectionSnapshot|TestConnectionIDReuse|TestCounterResetClosesSessions'`

Expected: compilation failure because ledger methods do not exist.

- [ ] **Step 3: Implement transactional session persistence**

Move connection delta calculation behind `persistConnectionSnapshot`. Use `connection.ID + "|" + connection.Start` as the stable session key, falling back to first-seen Unix milliseconds when `Start` is absent. In one transaction:

```sql
INSERT INTO connection_sessions (...) VALUES (...)
ON CONFLICT(session_key) DO UPDATE SET
  last_seen_at=excluded.last_seen_at,
  upload=excluded.upload,
  download=excluded.download,
  ended_at=NULL;
```

Mark sessions missing from the new active-ID set with `ended_at = now`. Update `processConnections` to call the new function, then feed returned deltas into the existing minute buffer. On shutdown call `closeObservedSessions` before closing SQLite.

- [ ] **Step 4: Verify ledger tests and regressions**

Run: `go test ./... -run 'TestPersistConnectionSnapshot|TestConnectionIDReuse|TestCounterResetClosesSessions|TestProcessConnections'`

Expected: PASS.

Run: `go test ./...`

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```powershell
git add main.go connection_ledger.go connection_ledger_test.go
git commit -m "功能：永久记录连接会话与流量增量"
```

### Task 3: Permanent minute facts and hour/day/week/month rollups

**Files:**
- Create: `rollups.go`
- Create: `rollups_test.go`
- Modify: `main.go`

**Interfaces:**
- Consumes: `traffic_aggregated`, `traffic_rollups`, `service.currentTime`.
- Produces: `bucketBounds(time.Time, granularity string) (start, end time.Time)`, `(*service).buildCompletedRollups(now time.Time) error`, and permanent cleanup behavior.

- [ ] **Step 1: Write failing boundary and idempotency tests**

Use `time.FixedZone("CST", 8*3600)` and test hour, local calendar day, Monday-start week, month, December-to-January, and leap-February boundaries. Insert minute facts, build rollups twice, and assert bytes are unchanged after the second run.

```go
func TestBucketBoundsUsesLocalCalendarWeekAndMonth(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 8, 12, 21, 30, 0, 0, loc)
	weekStart, weekEnd := bucketBounds(now, "week")
	if weekStart.Weekday() != time.Monday || weekEnd.Sub(weekStart) != 7*24*time.Hour {
		t.Fatalf("bad week: %v - %v", weekStart, weekEnd)
	}
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestBucketBounds|TestBuildCompletedRollups|TestPermanentRetention'`

Expected: compilation failure for missing rollup functions or assertion failure because cleanup deletes traffic.

- [ ] **Step 3: Implement deterministic rollups and remove traffic deletion**

Build completed buckets by deleting and reinserting only the target bucket inside one transaction. Group by all ledger dimensions and use `SUM(upload)`, `SUM(download)`, `SUM(count)`. Run the builder after each completed minute and at startup to fill missing completed ranges. Change `cleanupOldLogs` so it only checkpoints WAL and performs scheduled `PRAGMA optimize`; it must not delete `connection_sessions`, `traffic_aggregated`, or `traffic_rollups`.

- [ ] **Step 4: Run tests and verify**

Run: `go test ./... -run 'TestBucketBounds|TestBuildCompletedRollups|TestPermanentRetention'`

Expected: PASS.

Run: `go test ./...`

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```powershell
git add main.go rollups.go rollups_test.go
git commit -m "功能：永久保存分层流量汇总"
```

### Task 4: HTTP and Clash Verge named-pipe transports

**Files:**
- Create: `mihomo_transport.go`
- Create: `mihomo_transport_windows.go`
- Create: `mihomo_transport_other.go`
- Create: `mihomo_transport_test.go`
- Modify: `main.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Produces: `newMihomoHTTPClient(endpoint string, timeout time.Duration) (*http.Client, string, error)`.
- Consumes: `microsoft/go-winio.DialPipeContext` on Windows and standard `http.Transport` elsewhere.

- [ ] **Step 1: Write failing endpoint-normalization tests**

Test `http://127.0.0.1:9090`, `npipe://verge-mihomo`, blank Windows default, malformed endpoints, and authorization preservation. Use an injected dial function to test named-pipe HTTP without depending on the live Clash process.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestNewMihomoHTTPClient|TestNamedPipeTransport'`

Expected: compilation failure because the transport factory does not exist.

- [ ] **Step 3: Implement transport abstraction**

For `npipe://name`, return an `http.Client` whose transport uses:

```go
transport := &http.Transport{
	DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return winio.DialPipeContext(ctx, `\\.\pipe\`+pipeName)
	},
}
return &http.Client{Transport: transport, Timeout: timeout}, "http://localhost", nil
```

Keep ordinary HTTP URLs unchanged. Set the Windows default endpoint to `npipe://verge-mihomo`. Refactor `/connections`, `/proxies`, and proxy-switch requests to use the same transport factory.

- [ ] **Step 4: Add dependency and verify both target sets**

Run: `go get github.com/Microsoft/go-winio@latest`

Run: `go test ./...`

Expected: PASS on Windows.

Run: `$env:GOOS='linux'; go test ./...; Remove-Item Env:GOOS`

Expected: package compiles for Linux tests that do not require CGO cross-linking; if SQLite prevents execution, use `go test -run '^$'` with the repository's existing cross-compiler configuration.

- [ ] **Step 5: Commit**

```powershell
git add main.go mihomo_transport*.go go.mod go.sum
git commit -m "功能：支持 Clash Verge 命名管道"
```

### Task 5: Process, route, node, and time-granularity APIs

**Files:**
- Create: `ledger_api.go`
- Create: `ledger_api_test.go`
- Modify: `main.go`

**Interfaces:**
- Produces: `GET /api/ledger/summary`, `/api/ledger/ranking`, `/api/ledger/trend`, `/api/ledger/connections`, and `/api/ledger/status`.
- Query parameters: `start`, `end`, `granularity`, `routeType`, `dimension`, `label`, `limit`, `offset`.

- [ ] **Step 1: Write failing handler tests**

Seed proxy, DIRECT and REJECT rows for `codex.exe` and `OneDrive.exe`. Assert default summary returns proxy-only bytes, `routeType=all` returns all bytes, process ranking separates applications, node ranking reports actual outbound, and pagination is stable.

```go
func TestLedgerSummaryDefaultsToProxyTraffic(t *testing.T) {
	svc := newLedgerHTTPTestService(t)
	seedLedgerTraffic(t, svc.db)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/ledger/summary?start=0&end=9999999999999", nil)
	svc.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	assertJSONNumber(t, rec.Body.Bytes(), "total", 300)
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestLedgerSummary|TestLedgerRanking|TestLedgerConnections|TestLedgerStatus'`

Expected: 404 responses because routes do not exist.

- [ ] **Step 3: Implement validated SQL queries**

Allowlist dimensions (`process`, `host`, `destinationIP`, `outbound`, `rule`) and granularities (`hour`, `day`, `week`, `month`); never interpolate user-provided column names directly. Default `routeType=proxy`. Return separate `upload`, `download`, `total`, and `connections` fields. Connections endpoint reads permanent `connection_sessions`, not active-memory state.

- [ ] **Step 4: Verify handler and full tests**

Run: `go test ./... -run 'TestLedgerSummary|TestLedgerRanking|TestLedgerConnections|TestLedgerStatus'`

Expected: PASS.

Run: `go test ./...`

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```powershell
git add main.go ledger_api.go ledger_api_test.go
git commit -m "功能：提供进程与节点流量查询接口"
```

### Task 6: Process-ledger Web UI

**Files:**
- Modify: `web/index.html`
- Modify: `web/app.js`
- Modify: `web/styles.css`
- Modify: `main_test.go`

**Interfaces:**
- Consumes: Task 5 `/api/ledger/*` JSON.
- Produces: operational Chinese dashboard with route and granularity filters.

- [ ] **Step 1: Add failing embedded-asset assertions**

Assert the HTML contains route options `代理流量`, `DIRECT`, `REJECT`, `全部`; dimensions include `软件`, `节点`, `域名`, `目标 IP`; granularity includes `小时`, `天`, `周`, `月`; JavaScript requests `/api/ledger/summary`, `/api/ledger/ranking`, `/api/ledger/trend`, and `/api/ledger/connections`.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestLedgerUI'`

Expected: assertion failure because the new controls and API calls are absent.

- [ ] **Step 3: Implement the UI**

Replace the primary dashboard data path with the ledger endpoints while preserving the settings and auto-switch panels. Default to proxy traffic and process ranking. Clicking a process opens its node/domain breakdown and permanent connection table. Format bytes with IEC units and display process path in a tooltip. Show collector state and database size without exposing secrets.

- [ ] **Step 4: Verify assets and backend**

Run: `go test ./... -run 'TestLedgerUI|TestEmbedded'`

Expected: PASS.

Run: `go test ./...`

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```powershell
git add web/index.html web/app.js web/styles.css main_test.go
git commit -m "界面：增加软件与节点流量账本"
```

### Task 7: Windows tray, single instance, and per-user autostart

**Files:**
- Create: `desktop.go`
- Create: `desktop_windows.go`
- Create: `desktop_other.go`
- Create: `desktop_windows_test.go`
- Create: `web/tray-icon.png`
- Modify: `main.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Produces: `runDesktop(ctx context.Context, app *application) error`, `setAutoStart(enabled bool) error`, `isAutoStartEnabled() bool`, `acquireSingleInstance() (release func(), alreadyRunning bool, err error)`.
- Consumes: application shutdown callback, local dashboard URL, embedded tray PNG.

- [ ] **Step 1: Write failing tests around platform-neutral seams**

Inject registry and process-launch functions. Assert autostart writes a quoted executable path with `--background`, disabling removes only the `ClashTrafficMonitor` value, and an already-held mutex causes the second invocation to open the dashboard and exit.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestAutoStart|TestSingleInstance|TestDesktopMenuActions'`

Expected: compilation failure because desktop functions do not exist.

- [ ] **Step 3: Implement Windows lifecycle**

Use `github.com/getlantern/systray` for the notification icon, `golang.org/x/sys/windows/registry` for `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, and a named Windows mutex for single-instance enforcement. Menu actions are exactly: open dashboard, pause/resume collection, back up database, enable/disable autostart, open data folder, exit. `--background` suppresses automatic browser opening; interactive launch opens it once.

Non-Windows `runDesktop` starts the HTTP service normally without tray behavior. Refactor `main` into a testable `application` lifecycle so Exit flushes aggregates and sessions before process termination.

- [ ] **Step 4: Verify tests and Windows build**

Run: `go get github.com/getlantern/systray@latest golang.org/x/sys@latest`

Run: `go test ./...`

Expected: all tests PASS.

Run: `go build -trimpath -ldflags "-H=windowsgui -s -w" -o dist/ClashTrafficMonitor.exe .`

Expected: exit 0 and `dist/ClashTrafficMonitor.exe` exists.

- [ ] **Step 5: Commit**

```powershell
git add main.go desktop*.go web/tray-icon.png go.mod go.sum
git commit -m "功能：增加 Windows 托盘与登录自启"
```

### Task 8: Windows installer, durable paths, backup and documentation

**Files:**
- Create: `install-windows.ps1`
- Create: `uninstall-windows.ps1`
- Create: `backup.go`
- Create: `backup_test.go`
- Modify: `README.md`
- Modify: `.github/workflows/release.yml`
- Modify: `web/index.html`

**Interfaces:**
- Produces: `defaultDataRoot() string`, `backupDatabase(destination string) (string, error)`, install/uninstall scripts, Windows release artifact.

- [ ] **Step 1: Write failing path and online-backup tests**

Assert Windows defaults resolve beneath `%LOCALAPPDATA%\ClashTrafficMonitor`, backup uses SQLite online backup or `VACUUM INTO`, and backup contains all three traffic tables while collection remains open.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestDefaultDataRoot|TestBackupDatabase'`

Expected: compilation failure because durable-path and backup functions do not exist.

- [ ] **Step 3: Implement durable installation and safe uninstall**

`install-windows.ps1` copies the EXE into `%LOCALAPPDATA%\ClashTrafficMonitor\app`, launches `--enable-autostart --background`, and opens the dashboard. `uninstall-windows.ps1` stops the process and removes the app/autostart entry while preserving `%LOCALAPPDATA%\ClashTrafficMonitor\data` by default. A separate explicit `-DeleteData` switch removes data only after printing the exact target path and confirmation.

Update release workflow to build the whole package with `go build .`, not only `main.go`, and add `-H=windowsgui` only to the Windows target. Update README with installation, data location, backup, traffic definitions, sub-second limitation and recovery procedure.

- [ ] **Step 4: Verify tests, script syntax, and release build**

Run: `go test ./...`

Expected: all tests PASS.

Run: `[scriptblock]::Create((Get-Content -Raw install-windows.ps1)) | Out-Null; [scriptblock]::Create((Get-Content -Raw uninstall-windows.ps1)) | Out-Null`

Expected: no PowerShell parser errors.

Run: `go build -trimpath -ldflags "-H=windowsgui -s -w" -o dist/ClashTrafficMonitor.exe .`

Expected: exit 0.

- [ ] **Step 5: Commit**

```powershell
git add install-windows.ps1 uninstall-windows.ps1 backup.go backup_test.go README.md .github/workflows/release.yml web/index.html
git commit -m "发布：完善 Windows 安装备份与构建"
```

### Task 9: Live Windows installation and end-to-end verification

**Files:**
- Create during runtime only: `%LOCALAPPDATA%\ClashTrafficMonitor\app\ClashTrafficMonitor.exe`
- Create during runtime only: `%LOCALAPPDATA%\ClashTrafficMonitor\data\traffic_monitor.db`
- Create during runtime only: `%LOCALAPPDATA%\ClashTrafficMonitor\logs\traffic-monitor.log`
- Modify only if verification exposes a defect: corresponding production file plus a failing regression test.

**Interfaces:**
- Consumes: Tasks 1–8 complete release artifact.
- Produces: installed background application and verification evidence.

- [ ] **Step 1: Run full automated verification**

Run: `gofmt -w *.go`

Run: `go vet ./...`

Expected: no diagnostics.

Run: `go test ./... -count=1`

Expected: all tests PASS without cached results.

- [ ] **Step 2: Build and inspect artifact**

Run: `New-Item -ItemType Directory -Force dist | Out-Null; go build -trimpath -ldflags "-H=windowsgui -s -w" -o dist/ClashTrafficMonitor.exe .`

Expected: exit 0 and a non-empty PE executable.

- [ ] **Step 3: Install for the current user**

Run: `powershell -ExecutionPolicy Bypass -File .\install-windows.ps1 -SourceExe .\dist\ClashTrafficMonitor.exe`

Expected: installed path printed, process running once, autostart value present.

- [ ] **Step 4: Verify live Clash Verge collection without synthetic traffic**

Wait for existing normal traffic only. Query `/api/ledger/status`, `/api/ledger/ranking?dimension=process&routeType=all`, and SQLite counts. Confirm the status is connected through `npipe://verge-mihomo`, rows include real process names and proxy/DIRECT classifications, and no subscription or Clash configuration file changed.

- [ ] **Step 5: Verify restart continuity and tray actions**

Exit through the tray, restart with `--background`, and confirm database counts never decrease, completed sessions have `ended_at`, only one process instance runs, and opening from the tray reaches the local dashboard.

- [ ] **Step 6: Create and validate a backup**

Invoke the backup action, open the copied SQLite database read-only, and confirm session, minute and rollup row counts are readable.

- [ ] **Step 7: Final repository verification and commit fixes if needed**

Run: `git status --short; git diff --check; go vet ./...; go test ./... -count=1`

Expected: no uncommitted production changes, no whitespace errors, no vet errors, all tests PASS.

If a live defect was found, first add a failing regression test, implement the minimal fix, rerun the command above, then commit with a Chinese message describing the user-visible correction.
