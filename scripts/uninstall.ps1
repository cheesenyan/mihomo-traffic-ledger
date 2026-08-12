$ErrorActionPreference = "Stop"
$installRoot = Join-Path $env:LOCALAPPDATA "ClashTrafficMonitor"
$targetExe = Join-Path $installRoot "app\ClashTrafficMonitor.exe"
$runKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"

Get-Process -Name "ClashTrafficMonitor" -ErrorAction SilentlyContinue |
    Where-Object { $_.Path -eq $targetExe } |
    Stop-Process -Force
Remove-ItemProperty -Path $runKey -Name "ClashTrafficMonitor" -ErrorAction SilentlyContinue
$shell = New-Object -ComObject WScript.Shell
$shortcutName = "Clash 软件流量账本.lnk"
@(
    (Join-Path (Join-Path $shell.SpecialFolders.Item("StartMenu") "Programs") $shortcutName),
    (Join-Path $shell.SpecialFolders.Item("Desktop") $shortcutName)
) | ForEach-Object { Remove-Item -LiteralPath $_ -Force -ErrorAction SilentlyContinue }

Write-Host "已停止并取消开机自启。历史数据库仍保留在：$installRoot\data"
Write-Host "如确实不再需要历史数据，请手工删除整个目录：$installRoot"
