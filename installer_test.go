package main

import (
	"os"
	"strings"
	"testing"
)

func TestInnoInstallerContract(t *testing.T) {
	data, err := os.ReadFile("installer/ClashTrafficMonitor.iss")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	for _, want := range []string{
		`AppName=Clash 软件流量账本`,
		`#define MyAppVersion "1.1.0"`,
		`AppPublisher=Severin Ye`,
		`AppPublisherURL=https://github.com/severin-ye/mihomo-traffic-ledger`,
		`AppSupportURL=https://github.com/severin-ye/mihomo-traffic-ledger/issues`,
		`PrivilegesRequired=lowest`,
		`DefaultDirName={localappdata}\ClashTrafficMonitor\app`,
		`Name: "desktopicon"`,
		`Name: "autostart"`,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		`ClashTrafficMonitor.exe`,
		`Source: "..\LICENSE"`,
		`Source: "..\THIRD_PARTY_NOTICES.md"`,
		`Flags: postinstall nowait skipifsilent`,
		`UninstallDisplayName=Clash 软件流量账本`,
		`--shutdown`,
		`CheckForMutexes`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("installer contract missing %q", want)
		}
	}
}

func TestInstallerBuildScriptIsPortableAndUsesGUIBuild(t *testing.T) {
	data, err := os.ReadFile("scripts/build-installer.ps1")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	for _, want := range []string{
		`[string]$Version = "1.1.0"`,
		`Get-Command go.exe`,
		`$env:GOROOT`,
		`Inno Setup 6`,
		`-H=windowsgui`,
		`ISCC.exe`,
		`ClashTrafficMonitor.iss`,
		`Mihomo-Traffic-Ledger-Setup-v$Version.exe`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("build-installer.ps1 missing %q", want)
		}
	}
	if strings.Contains(script, `C:\Users\`) {
		t.Fatal("build-installer.ps1 must not contain a user-specific compiler path")
	}
}

func TestReleaseWorkflowUsesUTF8SafePowerShell(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)
	if !strings.Contains(workflow, `default: "1.1.0"`) {
		t.Fatal("release workflow manual default must match v1.1.0")
	}
	if !strings.Contains(workflow, "shell: pwsh") {
		t.Fatal("release workflow must use PowerShell Core so UTF-8 scripts parse correctly")
	}
	if strings.Contains(workflow, "shell: powershell") {
		t.Fatal("release workflow must not use Windows PowerShell 5.1 for UTF-8 scripts")
	}
}
