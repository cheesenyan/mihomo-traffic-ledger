# Mihomo Traffic Ledger v1.1.0

This release adds a persistent dashboard filter for direct traffic and an in-app Windows startup control.

## What's new

- Added the `隐藏直连` switch to the main toolbar.
- The last switch state is restored when the dashboard is opened again.
- When enabled, `DIRECT` traffic is excluded consistently from totals, upload/download cards, trend charts, rankings, secondary lists, and connection details.
- Detail and summary modes use the same filter semantics.
- Added a `开机自动启动` switch to Settings. It reads and updates the current user's real Windows startup entry, so startup can be enabled or disabled after installation.

## Data integrity

This is a display-only filter. Mihomo Traffic Ledger continues collecting and permanently storing all Mihomo-managed traffic, including `DIRECT` connections. Toggling the filter does not modify or delete the SQLite database.

## Install or upgrade

Download `Mihomo-Traffic-Ledger-Setup-v1.1.0.exe` and its `.sha256` file. The per-user installer safely replaces a running version and keeps the existing traffic database.

The binary remains unsigned, so Windows SmartScreen may display an “Unknown publisher” warning.

## Documentation

- [English README](https://github.com/severin-ye/mihomo-traffic-ledger#readme)
- [简体中文 README](https://github.com/severin-ye/mihomo-traffic-ledger/blob/main/README.zh-CN.md)
