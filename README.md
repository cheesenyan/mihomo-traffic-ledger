# Mihomo Traffic Ledger

<p align="center">
  A local-first Windows traffic ledger for Mihomo and Clash Verge Rev.
</p>

<p align="center">
  <a href="README.md">English</a> · <a href="README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <a href="https://github.com/severin-ye/mihomo-traffic-ledger/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/v/release/severin-ye/mihomo-traffic-ledger?style=flat-square"></a>
  <a href="https://github.com/severin-ye/mihomo-traffic-ledger/actions/workflows/release.yml"><img alt="Release build" src="https://img.shields.io/github/actions/workflow/status/severin-ye/mihomo-traffic-ledger/release.yml?style=flat-square&label=release"></a>
  <a href="LICENSE"><img alt="MIT License" src="https://img.shields.io/badge/license-MIT-blue?style=flat-square"></a>
  <img alt="Windows x64" src="https://img.shields.io/badge/Windows-10%2F11%20x64-0078D4?style=flat-square&logo=windows">
</p>

Mihomo Traffic Ledger answers a question that ordinary traffic charts cannot: **which application used which route or proxy node, and how much did it upload and download?**

It runs quietly in the Windows system tray, reads the connections already handled by Clash Verge Rev's Mihomo core, and stores a permanent local SQLite ledger. No external controller port, cloud account, or separate database is required.

> [!IMPORTANT]
> This is an independent community project. It is not affiliated with or endorsed by Mihomo, Clash, or Clash Verge Rev.

## Download

Download the current Windows x64 installer from [GitHub Releases](https://github.com/severin-ye/mihomo-traffic-ledger/releases/latest):

```text
Mihomo-Traffic-Ledger-Setup-v1.1.0.exe
```

The installer runs per-user and does not require administrator privileges. It can create desktop and Start menu shortcuts, optionally start the ledger at sign-in, and safely upgrade a running copy. Uninstalling the application keeps the historical database by default.

> [!NOTE]
> Current release binaries are not code-signed. Windows SmartScreen may show an “Unknown publisher” warning. Verify the attached SHA-256 file before installation.

## What it records

- Application/process name and executable path
- Destination host and IP address
- Rule, rule payload, and route type (`DIRECT` or `PROXY`)
- Full proxy chain and final outbound node
- Per-connection upload and download counters
- Permanent minute facts plus hour, day, week, and month rollups

The dashboard defaults to the **application** dimension and can switch to **proxy node**, host, or route views. It refreshes every five seconds while open.

Use the persistent **Hide direct traffic** (`隐藏直连`) switch to remove `DIRECT` traffic from every visible card, trend, ranking, and drilldown. This is a dashboard filter only: collection and permanent history continue to include direct connections.

The **Start with Windows** (`开机自动启动`) switch in Settings reads and updates the current user's real Windows startup entry. You can change it at any time without reinstalling the application.

## Screenshot

![Mihomo Traffic Ledger dashboard](readmeImg/image.png)

The current interface is Chinese. An English interface is planned; the source, installer instructions, and project documentation are available in English now.

## Requirements

- Windows 10 or Windows 11, x64
- Clash Verge Rev using its standard Mihomo named pipe: `npipe://./pipe/verge-mihomo`
- Traffic must be visible in Mihomo's `/connections` data

No external TCP Controller needs to be enabled for the default Clash Verge Rev setup.

## How it works

1. The collector reads Mihomo connection snapshots once per second through the local named pipe.
2. Counter differences are attributed to the process, destination, rule, route, and proxy chain reported by Mihomo.
3. Minute facts and connection sessions are written to SQLite; hour/day/week/month rollups are updated alongside them.
4. The embedded dashboard serves only on `http://127.0.0.1:18080`.

The first snapshot establishes a baseline, so counters accumulated before the ledger starts are not incorrectly charged to the current period.

## Privacy and data location

Mihomo Traffic Ledger has no telemetry and does not upload your traffic history. The dashboard is bound to loopback only.

```text
%LOCALAPPDATA%\ClashTrafficMonitor\
├── app\ClashTrafficMonitor.exe
├── data\traffic_monitor.db
└── logs\monitor.log
```

Traffic history is retained until you delete the database yourself. Keep in mind that permanent retention means the database can grow over time.

## Limitations

- The ledger can only count connections handled by Mihomo. Traffic that bypasses Mihomo entirely is invisible.
- Process attribution depends on the metadata exposed by the Mihomo core and operating system; some connections may appear without a process name.
- The current desktop integration is Windows-specific. The inherited server/Docker code remains in the repository but is not part of this project's supported release artifacts.
- The local dashboard has no authentication because it listens only on `127.0.0.1`; do not expose it through a reverse proxy without adding access control.

## Build from source

Requirements: Go 1.25 or newer and, for the installer, Inno Setup 6.

```powershell
git clone https://github.com/severin-ye/mihomo-traffic-ledger.git
cd mihomo-traffic-ledger

go test ./...
powershell -ExecutionPolicy Bypass -File .\scripts\build-installer.ps1
```

The installer is written to `dist\Mihomo-Traffic-Ledger-Setup-v1.1.0.exe`. You can select another version with `-Version`.

To build only the tray executable:

```powershell
go build -trimpath -ldflags "-s -w -H=windowsgui" -o .\dist\ClashTrafficMonitor.exe .
```

## Contributing

Bug reports and focused pull requests are welcome. Before submitting a change:

```powershell
go test ./...
```

Please include reproduction steps for collector or accounting issues and screenshots for visible dashboard changes. Never attach a real traffic database to a public issue; it can contain process paths, domains, IP addresses, and routing details.

## Upstream, attribution, and license

This project is a modified derivative of [zhf883680/clash-traffic-monitor](https://github.com/zhf883680/clash-traffic-monitor), which provided the original Go service, SQLite traffic aggregation, dashboard, and Mihomo integration. This derivative adds the Windows process ledger, named-pipe transport, permanent multi-level rollups, tray application, installer, and related user experience.

The project is distributed under the [MIT License](LICENSE). The original copyright notice is preserved, and the independent modifications carry an additional copyright notice. Third-party components and their licenses are listed in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

Mihomo, Clash, Clash Verge Rev, Windows, and other names may be trademarks of their respective owners.

## Acknowledgements

- [zhf883680/clash-traffic-monitor](https://github.com/zhf883680/clash-traffic-monitor) — the upstream project this work is derived from
- [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) — the compatible proxy core and controller API
- [clash-verge-rev/clash-verge-rev](https://github.com/clash-verge-rev/clash-verge-rev) — the Windows client used by the default named-pipe integration
- [MetaCubeX/metacubexd](https://github.com/MetaCubeX/metacubexd) and [foru17/neko](https://github.com/foru17/neko) — interface and ecosystem references acknowledged by the upstream project
