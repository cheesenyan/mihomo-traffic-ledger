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
