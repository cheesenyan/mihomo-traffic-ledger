$ErrorActionPreference = "Stop"
$installRoot = Join-Path $env:LOCALAPPDATA "ClashTrafficMonitor"
$targetExe = Join-Path $installRoot "app\ClashTrafficMonitor.exe"
$runKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"

Get-Process -Name "ClashTrafficMonitor" -ErrorAction SilentlyContinue |
    Where-Object { $_.Path -eq $targetExe } |
    Stop-Process -Force
Remove-ItemProperty -Path $runKey -Name "ClashTrafficMonitor" -ErrorAction SilentlyContinue

Write-Host "已停止并取消开机自启。历史数据库仍保留在：$installRoot\data"
Write-Host "如确实不再需要历史数据，请手工删除整个目录：$installRoot"
