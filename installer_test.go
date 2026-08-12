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
		`PrivilegesRequired=lowest`,
		`DefaultDirName={localappdata}\ClashTrafficMonitor\app`,
		`Name: "desktopicon"`,
		`Name: "autostart"`,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		`ClashTrafficMonitor.exe`,
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

func TestInstallerBuildScriptUsesPinnedCompilerAndGUIBuild(t *testing.T) {
	data, err := os.ReadFile("scripts/build-installer.ps1")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	for _, want := range []string{"Inno Setup 6", "-H=windowsgui", "ISCC.exe", "ClashTrafficMonitor.iss"} {
		if !strings.Contains(script, want) {
			t.Fatalf("build-installer.ps1 missing %q", want)
		}
	}
}
