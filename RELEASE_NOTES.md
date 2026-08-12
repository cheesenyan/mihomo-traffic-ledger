# Mihomo Traffic Ledger v1.0.0

The first independent Windows release of Mihomo Traffic Ledger.

## Highlights

- Tracks upload and download by application, executable path, destination, rule, route, proxy chain, and final node.
- Connects directly to Clash Verge Rev through `npipe://./pipe/verge-mihomo`; no external Controller port is required.
- Preserves connection sessions and minute facts, with hour/day/week/month SQLite rollups.
- Runs as a single-instance Windows tray application with an embedded local dashboard.
- Includes a per-user installer, desktop and Start menu shortcuts, optional sign-in startup, and safe in-place upgrades.
- Keeps the historical database when the application is uninstalled.

## Install

1. Download `Mihomo-Traffic-Ledger-Setup-v1.0.0.exe` and its `.sha256` file.
2. Verify the checksum.
3. Run the installer. Administrator privileges are not required.

The binary is currently unsigned, so Windows SmartScreen may display an “Unknown publisher” warning.

## Documentation

- [English README](https://github.com/severin-ye/mihomo-traffic-ledger#readme)
- [简体中文 README](https://github.com/severin-ye/mihomo-traffic-ledger/blob/main/README.zh-CN.md)

## Provenance

This release is derived from [zhf883680/clash-traffic-monitor](https://github.com/zhf883680/clash-traffic-monitor) under the MIT License. See [THIRD_PARTY_NOTICES.md](https://github.com/severin-ye/mihomo-traffic-ledger/blob/main/THIRD_PARTY_NOTICES.md) for details.
