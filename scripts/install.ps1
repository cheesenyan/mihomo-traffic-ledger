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

Get-Process -Name "ClashTrafficMonitor" -ErrorAction SilentlyContinue |
    Where-Object { $_.Path -eq $targetExe } |
    Stop-Process -Force

Copy-Item -LiteralPath $resolvedBuild -Destination $targetExe -Force
New-Item -Path $runKey -Force | Out-Null
New-ItemProperty -Path $runKey -Name "ClashTrafficMonitor" -Value ('"' + $targetExe + '"') -PropertyType String -Force | Out-Null

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
}
