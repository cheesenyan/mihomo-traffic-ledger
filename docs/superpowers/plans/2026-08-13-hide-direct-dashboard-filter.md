# Hide Direct Dashboard Filter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a persistent dashboard-only `隐藏直连` switch that excludes `DIRECT` traffic consistently without changing collection or stored history.

**Architecture:** Parse `excludeDirect=1` at each traffic endpoint and carry a boolean into the shared SQLite and in-memory query paths, where it adds `route_type <> 'DIRECT'` before aggregation. The embedded frontend restores the switch from `localStorage`, sends the parameter with every aggregate/trend/drilldown request, and reloads the current view when toggled.

**Tech Stack:** Go 1.25+, SQLite, embedded HTML/CSS/JavaScript, Go tests, Inno Setup 6

## Global Constraints

- Filtering affects dashboard reads only; collection and persistence continue to record all route types.
- Only `DIRECT` is excluded; `PROXY` and `REJECT` remain visible.
- A missing or false query parameter preserves existing include-all behavior.
- The last switch state is restored from browser `localStorage`.
- Cards, trend, ranking, secondary rows, and detail cards must use the same filter state.

---

### Task 1: Backend query filter

**Files:**
- Modify: `main_test.go`
- Modify: `main.go`
- Modify: `rollup_test.go`
- Modify: `rollup.go`

**Interfaces:**
- Consumes: query parameter `excludeDirect=1`.
- Produces: `excludeDirect(r *http.Request) bool` and boolean-aware aggregate, trend, rollup, substat, and detail query functions.

- [ ] **Step 1: Write failing mixed-route API tests**

Insert one `DIRECT` and one `PROXY` aggregate for the same process, call `/api/traffic/aggregate`, `/api/traffic/trend`, and `/api/traffic/details` with `excludeDirect=1`, and assert only proxy bytes are returned. Add a rollup test proving summary mode applies the same condition.

- [ ] **Step 2: Verify tests fail because direct bytes remain present**

Run:

```powershell
go test ./... -run 'Test.*ExcludeDirect' -count=1
```

Expected: FAIL with totals containing both test rows.

- [ ] **Step 3: Implement the shared backend condition**

Use this semantic helper:

```go
func shouldExcludeDirect(r *http.Request) bool {
    return r.URL.Query().Get("excludeDirect") == "1"
}
```

Thread the boolean through persisted and buffered query paths. Add `route_type <> 'DIRECT'` to SQL and reject buffered entries whose normalized route type is `DIRECT` before grouping.

- [ ] **Step 4: Verify focused and full tests pass**

```powershell
go test ./... -run 'Test.*ExcludeDirect' -count=1
go test ./... -count=1
```

- [ ] **Step 5: Commit**

```powershell
git add main.go main_test.go rollup.go rollup_test.go
git commit -m "功能：支持查询排除直连流量"
```

### Task 2: Persistent toolbar switch

**Files:**
- Modify: `main_test.go`
- Modify: `web/index.html`
- Modify: `web/styles.css`
- Modify: `web/app.js`

**Interfaces:**
- Consumes: backend `excludeDirect=1` support.
- Produces: checkbox `#excludeDirect`, storage key `traffic-monitor.exclude-direct`, and one shared request parameter derived from the checkbox.

- [ ] **Step 1: Write failing embedded-frontend contract tests**

Assert that embedded assets contain `id="excludeDirect"`, label `隐藏直连`, the storage key, restore/persist functions, a change listener that invokes `loadData()`, and `excludeDirect` on every traffic request parameter object.

- [ ] **Step 2: Verify the frontend contract test fails**

```powershell
go test ./... -run TestEmbeddedIndexIncludesExcludeDirectFilter -count=1
```

Expected: FAIL because the switch does not exist.

- [ ] **Step 3: Implement the compact switch and persistence**

Add a toolbar checkbox styled consistently with existing controls. Restore `true` only when the stored value is `"1"`; persist `"1"` or `"0"`. Use:

```js
const excludeDirect = elements.excludeDirect.checked ? "1" : ""
```

and include it in aggregate, trend, secondary, and detail request parameters.

- [ ] **Step 4: Verify JavaScript syntax and tests**

```powershell
node --check web/app.js
go test ./... -run 'TestEmbedded.*ExcludeDirect' -count=1
go test ./... -count=1
```

- [ ] **Step 5: Commit**

```powershell
git add web/index.html web/styles.css web/app.js main_test.go
git commit -m "功能：增加持久化隐藏直连开关"
```

### Task 3: Local installation and v1.1.0 release

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `RELEASE_NOTES.md`
- Modify: `web/index.html`

**Interfaces:**
- Consumes: completed filter implementation.
- Produces: updated local tray application, `v1.1.0` installer, repository `main`, and GitHub Release assets.

- [ ] **Step 1: Document the dashboard-only behavior and bump visible version**

Describe that `隐藏直连` changes only the dashboard and never collection/history. Set the footer to `v1.1.0` and replace release notes with the v1.1.0 change summary.

- [ ] **Step 2: Build, install, and verify locally**

```powershell
go test ./... -count=1
powershell -ExecutionPolicy Bypass -File scripts/build-installer.ps1 -Version 1.1.0
powershell -ExecutionPolicy Bypass -File scripts/install.ps1 -BuildPath dist/ClashTrafficMonitor.exe -NoOpen
Invoke-RestMethod http://127.0.0.1:18080/health
```

Open the live dashboard, toggle `隐藏直连`, and verify cards, trend, rankings, and detail cards update while the database remains present.

- [ ] **Step 3: Commit documentation and release metadata**

```powershell
git add README.md README.zh-CN.md RELEASE_NOTES.md web/index.html
git commit -m "发布：准备隐藏直连筛选 1.1.0"
```

- [ ] **Step 4: Push and publish**

Push the completed HEAD to `origin/main`, create tag/release `v1.1.0`, and attach `Mihomo-Traffic-Ledger-Setup-v1.1.0.exe` plus its SHA-256 file.

- [ ] **Step 5: Verify remote delivery**

Confirm the GitHub Actions release run succeeds, download both release assets, verify the checksum, and confirm local HEAD equals remote `main` with a clean worktree.
