param(
    [string]$BuildPath = (Join-Path $PSScriptRoot "..\dist\ClashTrafficMonitor.exe"),
    [switch]$NoOpen
)

$ErrorActionPreference = "Stop"
$installRoot = Join-Path $env:LOCALAPPDATA "ClashTrafficMonitor"
$appDir = Join-Path $installRoot "app"
$targetExe = Join-Path $appDir "ClashTrafficMonitor.exe"
$runKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
$dashboardUrl = "http://127.0.0.1:18080"

$resolvedBuild = (Resolve-Path -LiteralPath $BuildPath).Path
New-Item -ItemType Directory -Force -Path $appDir | Out-Null

$running = Get-Process -Name "ClashTrafficMonitor" -ErrorAction SilentlyContinue |
    Where-Object { $_.Path -eq $targetExe }
foreach ($process in $running) {
    $process | Stop-Process -Force
    $null = $process.WaitForExit(5000)
}

$copied = $false
for ($attempt = 1; $attempt -le 10 -and -not $copied; $attempt++) {
    try {
        Copy-Item -LiteralPath $resolvedBuild -Destination $targetExe -Force
        $copied = $true
    } catch {
        if ($attempt -eq 10) { throw }
        Start-Sleep -Milliseconds 250
    }
}
New-Item -Path $runKey -Force | Out-Null
New-ItemProperty -Path $runKey -Name "ClashTrafficMonitor" -Value ('"' + $targetExe + '"') -PropertyType String -Force | Out-Null

$shell = New-Object -ComObject WScript.Shell
$shortcutName = "Clash 软件流量账本.lnk"
$shortcutPaths = @(
    (Join-Path (Join-Path $shell.SpecialFolders.Item("StartMenu") "Programs") $shortcutName),
    (Join-Path $shell.SpecialFolders.Item("Desktop") $shortcutName)
)
foreach ($shortcutPath in $shortcutPaths) {
    $shortcut = $shell.CreateShortcut($shortcutPath)
    $shortcut.TargetPath = $targetExe
    $shortcut.WorkingDirectory = $appDir
    $shortcut.Description = "查看 Clash 中每个软件使用的节点和流量"
    $shortcut.IconLocation = "$targetExe,0"
    $shortcut.Save()
}

Start-Process -FilePath $targetExe -WorkingDirectory $appDir -WindowStyle Hidden
$deadline = (Get-Date).AddSeconds(15)
do {
    try {
        $health = Invoke-RestMethod -Uri "$dashboardUrl/health" -TimeoutSec 2
        break
    } catch {
        Start-Sleep -Milliseconds 500
    }
} while ((Get-Date) -lt $deadline)

if (-not $health) {
    throw "服务未能在 15 秒内启动。请检查 $installRoot\logs\monitor.log"
}

if (-not $NoOpen) {
    Start-Process $dashboardUrl
}

[pscustomobject]@{
    InstalledExe = $targetExe
    Database = (Join-Path $installRoot "data\traffic_monitor.db")
    Log = (Join-Path $installRoot "logs\monitor.log")
    Dashboard = $dashboardUrl
    AutoStart = $true
    StartMenuShortcut = $shortcutPaths[0]
    DesktopShortcut = $shortcutPaths[1]
}
