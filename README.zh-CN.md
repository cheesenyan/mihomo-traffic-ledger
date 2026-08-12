# Mihomo Traffic Ledger

<p align="center">
  面向 Mihomo 与 Clash Verge Rev 的本地优先 Windows 软件流量账本。
</p>

<p align="center">
  <a href="README.md">English</a> · <a href="README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <a href="https://github.com/severin-ye/mihomo-traffic-ledger/releases/latest"><img alt="最新版本" src="https://img.shields.io/github/v/release/severin-ye/mihomo-traffic-ledger?style=flat-square"></a>
  <a href="https://github.com/severin-ye/mihomo-traffic-ledger/actions/workflows/release.yml"><img alt="发布构建" src="https://img.shields.io/github/actions/workflow/status/severin-ye/mihomo-traffic-ledger/release.yml?style=flat-square&label=release"></a>
  <a href="LICENSE"><img alt="MIT License" src="https://img.shields.io/badge/license-MIT-blue?style=flat-square"></a>
  <img alt="Windows x64" src="https://img.shields.io/badge/Windows-10%2F11%20x64-0078D4?style=flat-square&logo=windows">
</p>

Mihomo Traffic Ledger 回答普通流量图无法回答的问题：**哪个软件走了哪个路由或代理节点，上传和下载分别用了多少流量？**

它常驻 Windows 系统托盘，读取 Clash Verge Rev 的 Mihomo 内核已经接管的连接，并在本机保存永久 SQLite 账本。默认不需要开放外部 Controller 端口，不需要云账号，也不需要单独部署数据库。

> [!IMPORTANT]
> 这是独立的社区项目，与 Mihomo、Clash、Clash Verge Rev 官方没有隶属、授权或背书关系。

## 下载

从 [GitHub Releases](https://github.com/severin-ye/mihomo-traffic-ledger/releases/latest) 下载当前 Windows x64 安装程序：

```text
Mihomo-Traffic-Ledger-Setup-v1.0.0.exe
```

安装器按当前用户安装，无需管理员权限；可以创建桌面和开始菜单快捷方式、选择登录后自动启动，并能安全升级正在运行的版本。卸载应用时默认保留历史数据库。

> [!NOTE]
> 当前发布文件尚未进行代码签名，Windows SmartScreen 可能显示“未知发布者”。安装前请核对 Release 附带的 SHA-256 文件。

## 记录内容

- 软件/进程名称与可执行文件路径
- 目标域名和 IP 地址
- 规则、规则载荷与路由类型（`DIRECT` 或 `PROXY`）
- 完整代理链与最终出口节点
- 每条连接的上传量与下载量
- 永久分钟事实，以及小时、天、周、月四层汇总

看板默认按**软件**统计，也可以切换到**代理节点**、主机或路由等维度；页面打开时每五秒自动刷新。

## 界面预览

![Mihomo Traffic Ledger 流量看板](readmeImg/image.png)

当前应用界面为中文；源码、安装说明与项目文档已同时提供英文版本，英文界面后续再完善。

## 运行要求

- Windows 10 或 Windows 11，x64
- Clash Verge Rev 使用标准 Mihomo 命名管道：`npipe://./pipe/verge-mihomo`
- 相关流量能够出现在 Mihomo 的 `/connections` 数据中

使用 Clash Verge Rev 默认配置时，无需额外开启外部 TCP Controller。

## 工作原理

1. 采集器每秒通过本机命名管道读取一次 Mihomo 连接快照。
2. 根据相邻快照的计数差，把流量归到 Mihomo 报告的软件、目标、规则、路由与代理链上。
3. 将分钟事实和连接会话写入 SQLite，同时更新小时、天、周、月汇总。
4. 内置看板只监听 `http://127.0.0.1:18080`。

程序首次看到连接时只建立基线，不会把启动前已经累计的连接流量错误计入当前时段。

## 隐私与数据位置

Mihomo Traffic Ledger 不包含遥测，不会上传流量历史；看板只绑定本机回环地址。

```text
%LOCALAPPDATA%\ClashTrafficMonitor\
├── app\ClashTrafficMonitor.exe
├── data\traffic_monitor.db
└── logs\monitor.log
```

历史流量会一直保留，直到你主动删除数据库。永久保留也意味着数据库体积会随使用时间增长。

## 已知边界

- 只能统计 Mihomo 实际接管的连接；完全绕过 Mihomo 的流量不可见。
- 软件归属依赖 Mihomo 内核和操作系统提供的元数据，少数连接可能没有进程名称。
- 当前桌面集成只支持 Windows。仓库仍保留上游的服务端/Docker 代码，但本项目的正式发布物只覆盖 Windows。
- 本地看板因为只监听 `127.0.0.1` 而未设置登录验证；如果自行通过反向代理暴露，请先增加访问控制。

## 从源码构建

需要 Go 1.25 或更高版本；构建安装器还需要 Inno Setup 6。

```powershell
git clone https://github.com/severin-ye/mihomo-traffic-ledger.git
cd mihomo-traffic-ledger

go test ./...
powershell -ExecutionPolicy Bypass -File .\scripts\build-installer.ps1
```

安装器输出到 `dist\Mihomo-Traffic-Ledger-Setup-v1.0.0.exe`，也可以通过 `-Version` 指定其他版本。

只构建托盘应用：

```powershell
go build -trimpath -ldflags "-s -w -H=windowsgui" -o .\dist\ClashTrafficMonitor.exe .
```

## 参与贡献

欢迎提交问题和范围明确的 Pull Request。提交前请运行：

```powershell
go test ./...
```

采集或计费问题请附复现步骤；可见界面变更请附截图。不要把真实流量数据库上传到公开 Issue，因为其中可能包含进程路径、域名、IP 地址和路由详情。

## 上游来源、署名与许可证

本项目修改自 [zhf883680/clash-traffic-monitor](https://github.com/zhf883680/clash-traffic-monitor)。上游提供了最初的 Go 服务、SQLite 流量聚合、Web 看板和 Mihomo 集成；本项目在此基础上增加了 Windows 软件流量账本、命名管道连接、永久多层汇总、托盘应用、安装程序及相关使用体验。

项目采用 [MIT License](LICENSE)。原作者的版权声明被完整保留，独立修改部分另行增加版权声明。第三方组件及其许可证见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。

Mihomo、Clash、Clash Verge Rev、Windows 及其他名称可能是各自权利人的商标。

## 致谢

- [zhf883680/clash-traffic-monitor](https://github.com/zhf883680/clash-traffic-monitor) — 本项目所基于的上游项目
- [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) — 兼容的代理内核与 Controller API
- [clash-verge-rev/clash-verge-rev](https://github.com/clash-verge-rev/clash-verge-rev) — 默认命名管道集成所面向的 Windows 客户端
- [MetaCubeX/metacubexd](https://github.com/MetaCubeX/metacubexd) 与 [foru17/neko](https://github.com/foru17/neko) — 上游项目注明的界面与生态参考

