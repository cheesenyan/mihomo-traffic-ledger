# Third-Party Notices

Mihomo Traffic Ledger is an independent community project and is not affiliated with or endorsed by the projects listed below. All product names and trademarks belong to their respective owners.

## Upstream project

This repository is a modified derivative of:

- **clash-traffic-monitor** — <https://github.com/zhf883680/clash-traffic-monitor>
- Copyright (c) 2026 zhf883680
- License: MIT

The upstream project supplied the original Go service, Mihomo Controller integration, SQLite aggregation, embedded dashboard, and automatic proxy-switching functionality. Mihomo Traffic Ledger adds Windows process attribution, Clash Verge Rev named-pipe transport, permanent connection sessions and multi-level rollups, a tray shell, per-user installation, and related documentation and tests.

The original MIT notice is preserved in the repository's [LICENSE](LICENSE) file and is included with binary installations.

## Runtime dependencies

The Windows binary contains the following Go modules. Version numbers are pinned by `go.mod` and `go.sum`.

| Component | Copyright holder(s) | License |
| --- | --- | --- |
| [Microsoft/go-winio](https://github.com/microsoft/go-winio) | Microsoft | MIT |
| [dustin/go-humanize](https://github.com/dustin/go-humanize) | Dustin Sallings | MIT |
| [gogpu/systray](https://github.com/gogpu/systray) | Andrey Kolkov and GoGPU Contributors | MIT |
| [mattn/go-isatty](https://github.com/mattn/go-isatty) | Yasuhiro Matsumoto | MIT |
| [ncruces/go-strftime](https://github.com/ncruces/go-strftime) | Nuno Cruces | MIT |
| [remyoudompheng/bigfft](https://github.com/remyoudompheng/bigfft) | The Go Authors | BSD-3-Clause |
| [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys) | The Go Authors | BSD-3-Clause |
| [modernc.org/libc](https://pkg.go.dev/modernc.org/libc) | The Libc Authors and third-party contributors | BSD-3-Clause and bundled third-party terms |
| [modernc.org/mathutil](https://pkg.go.dev/modernc.org/mathutil) | The mathutil Authors | BSD-3-Clause |
| [modernc.org/memory](https://pkg.go.dev/modernc.org/memory) | The Memory Authors and third-party contributors | BSD-3-Clause and bundled third-party terms |
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | The Sqlite Authors | BSD-3-Clause |

The authoritative license texts shipped by each module are part of that module's source distribution. Source builds retrieve the exact versions and license files through the Go module system. Binary installations include this notice and the project MIT license; consult the linked module source for any component-specific bundled notices.

## Referenced ecosystem projects

The application interoperates with or acknowledges these independent projects but does not redistribute their source as part of this repository:

- [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo)
- [clash-verge-rev/clash-verge-rev](https://github.com/clash-verge-rev/clash-verge-rev)
- [MetaCubeX/metacubexd](https://github.com/MetaCubeX/metacubexd)
- [foru17/neko](https://github.com/foru17/neko)

