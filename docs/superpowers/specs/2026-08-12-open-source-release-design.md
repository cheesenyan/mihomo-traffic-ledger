# Mihomo Traffic Ledger Open-Source Release Design

## Goal

Turn the Windows process traffic ledger into a clearly independent open-source project with an English-first landing page, a complete Chinese translation, explicit upstream attribution, and its own GitHub Releases page.

## Identity and positioning

- Public product name: **Mihomo Traffic Ledger**.
- Repository name: `mihomo-traffic-ledger` under the `severin-ye` GitHub account.
- Position it as a Windows tray application for understanding which processes use which Mihomo routes and how much they upload and download.
- Do not present the project as an official Mihomo, Clash, or Clash Verge Rev product.

## Documentation structure

- `README.md` is the canonical English landing page.
- `README.zh-CN.md` is a complete Chinese translation linked from the language switcher at the top of both files.
- Both pages lead with the Windows installer, then explain capabilities, screenshots, requirements, privacy, storage, source builds, limitations, and contribution paths.
- Technical behavior must match the current implementation: named-pipe connection to Clash Verge Rev, loopback-only dashboard, process and node accounting, hourly/daily/weekly/monthly rollups, tray operation, autostart, and retained SQLite history.

## Licensing and attribution

- Keep the project under the MIT License because the upstream project is MIT licensed.
- Preserve the original copyright notice for `zhf883680` and add a copyright notice for the independent Windows-derived work.
- Add `THIRD_PARTY_NOTICES.md` to identify the upstream repository, explain that this project is a modified derivative, link to upstream, and summarize direct Go dependencies and their license families.
- Clearly state that third-party names and trademarks belong to their owners and that the project is unofficial.

## Release delivery

- Replace inherited release automation with a Windows-focused GitHub Actions workflow.
- On a `v*` tag, run tests, build the GUI executable and Inno Setup installer, generate a SHA-256 checksum, and attach both to the GitHub Release.
- Create an independent public repository, preserve the original repository as the `upstream` remote, and use the new repository as `origin`.
- Publish the existing verified `v1.0.0` installer as the first release with concise English release notes and a link to Chinese documentation.

## Validation

- Verify all relative links and referenced files in both README files.
- Verify the workflow and installer script use the same artifact name and version format.
- Run `go test ./...`, rebuild the installer, and verify the setup executable and checksum.
- Verify the new repository is public, `origin` points to it, and the GitHub Release exposes the installer.

