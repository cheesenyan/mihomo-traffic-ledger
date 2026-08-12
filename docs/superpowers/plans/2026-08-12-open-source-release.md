# Mihomo Traffic Ledger Open-Source Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish Mihomo Traffic Ledger as an independent, English-first, correctly attributed open-source project with a reproducible Windows release.

**Architecture:** Keep the existing Go application unchanged. Improve the repository boundary around it: bilingual user documentation, explicit legal notices, a Windows-only tag release workflow, and a dedicated GitHub repository whose Releases page distributes the Inno Setup installer.

**Tech Stack:** Markdown, MIT License, Go 1.25, PowerShell, Inno Setup 6, GitHub Actions, GitHub Releases

## Global Constraints

- The canonical project name is `Mihomo Traffic Ledger`.
- `README.md` is English-first and links to a complete `README.zh-CN.md` translation.
- Preserve upstream copyright and explicitly identify the derivative relationship.
- Release artifacts target Windows x64 and retain existing application data on upgrade or uninstall.
- Do not claim official affiliation with Mihomo, Clash, or Clash Verge Rev.

---

### Task 1: Bilingual project landing pages

**Files:**
- Modify: `README.md`
- Create: `README.zh-CN.md`

**Interfaces:**
- Consumes: current runtime behavior and installer artifact naming.
- Produces: canonical English onboarding and equivalent Chinese onboarding.

- [ ] Rewrite `README.md` around the Windows application and the independent Releases URL.
- [ ] Create a section-for-section Chinese translation in `README.zh-CN.md`.
- [ ] Validate all local paths and GitHub links referenced by both files.
- [ ] Commit with a Chinese commit message.

### Task 2: License and provenance

**Files:**
- Modify: `LICENSE`
- Create: `THIRD_PARTY_NOTICES.md`

**Interfaces:**
- Consumes: upstream MIT license and `go.mod` direct dependencies.
- Produces: redistributable legal notice set for source and installer users.

- [ ] Preserve `Copyright (c) 2026 zhf883680` and add `Copyright (c) 2026 Severin Ye`.
- [ ] Document the upstream source URL, modification relationship, unofficial status, and direct dependency licenses.
- [ ] Confirm every direct dependency is represented or covered by module-distributed license files.
- [ ] Commit with a Chinese commit message.

### Task 3: Reproducible GitHub Release

**Files:**
- Modify: `.github/workflows/release.yml`
- Delete: `.github/workflows/build.yml`
- Modify: `scripts/build-installer.ps1`
- Modify: `installer/ClashTrafficMonitor.iss`
- Modify: `RELEASE_NOTES.md`

**Interfaces:**
- Consumes: a tag formatted as `vMAJOR.MINOR.PATCH`.
- Produces: `Mihomo-Traffic-Ledger-Setup-vMAJOR.MINOR.PATCH.exe` and matching `.sha256` file.

- [ ] Make the installer builder portable across developer and GitHub-hosted Windows machines.
- [ ] Update publisher and support URLs to `severin-ye/mihomo-traffic-ledger`.
- [ ] Build and release only the supported Windows installer on version tags.
- [ ] Replace inherited release notes with the independent v1.0.0 notes.
- [ ] Run PowerShell parser validation, `go test ./...`, and a local installer build.
- [ ] Commit with a Chinese commit message.

### Task 4: Independent repository and first release

**Files:**
- No source file changes.

**Interfaces:**
- Consumes: verified repository commit and installer artifacts.
- Produces: public GitHub repository and `v1.0.0` release page.

- [ ] Rename the current `origin` remote to `upstream`.
- [ ] Create `severin-ye/mihomo-traffic-ledger` as a public repository and add it as `origin`.
- [ ] Push the completed history to `main`.
- [ ] Create GitHub Release `v1.0.0` and upload installer plus SHA-256 file.
- [ ] Verify repository visibility, release URL, asset names, and download availability.
