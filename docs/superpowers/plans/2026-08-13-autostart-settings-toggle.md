# Application Autostart Toggle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an application settings switch backed directly by the current-user Windows startup registry value.

**Architecture:** Expose a small GET/PUT settings endpoint whose backend is split into Windows registry and non-Windows implementations. The existing settings form loads the real state and saves it alongside existing settings, using the installer registry value as the single source of truth.

**Tech Stack:** Go 1.25+, `golang.org/x/sys/windows/registry`, embedded HTML/JavaScript, Go tests

## Global Constraints

- Registry key: `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`.
- Registry value: `ClashTrafficMonitor`.
- No duplicate SQLite or browser setting.
- Disable removes only the named value and never stops the running application.
- Tests must not modify the user's real registry.

---

### Task 1: Autostart backend and API

**Files:**
- Create: `autostart.go`
- Create: `autostart_windows.go`
- Create: `autostart_other.go`
- Create: `autostart_test.go`
- Modify: `main.go`

**Interfaces:**
- Produces: `readAutostartEnabled() (bool, error)`, `writeAutostartEnabled(bool) error`, and `/api/settings/autostart`.

- [ ] Write handler tests with substituted read/write functions for GET, PUT enable, PUT disable, invalid JSON, and backend errors.
- [ ] Run `go test ./... -run TestAutostart -count=1` and verify failure because the endpoint is absent.
- [ ] Implement platform backend and handler, then register the route.
- [ ] Run focused and full tests.
- [ ] Commit as `功能：支持应用内管理开机自启`.

### Task 2: Settings panel integration

**Files:**
- Modify: `main_test.go`
- Modify: `web/index.html`
- Modify: `web/app.js`

**Interfaces:**
- Consumes: `/api/settings/autostart` GET and PUT.
- Produces: `#autostartEnabled` reflecting the registry state.

- [ ] Write an embedded-asset test for the switch and load/save API calls.
- [ ] Run the test and verify it fails because the switch is absent.
- [ ] Add the settings control, state field, load synchronization, and save request.
- [ ] Run JavaScript syntax, focused, and full tests.
- [ ] Commit as `功能：在设置中增加开机自启开关`.

### Task 3: Combined v1.1.0 delivery

**Files:**
- Modify: current pending `README.md`, `README.zh-CN.md`, `RELEASE_NOTES.md`, version and release configuration files.

**Interfaces:**
- Produces: installed v1.1.0 application and public v1.1.0 release.

- [ ] Document both the dashboard-only direct filter and application autostart control.
- [ ] Run full tests and build the v1.1.0 installer.
- [ ] Install locally, test actual registry disable/enable, verify health and dashboard behavior.
- [ ] Commit release metadata, push HEAD to `origin/main`, create and push `v1.1.0`.
- [ ] Verify GitHub Actions, Release assets, downloaded SHA-256, remote HEAD, and clean worktree.
