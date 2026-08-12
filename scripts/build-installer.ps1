param(
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$goCommand = Get-Command go.exe -ErrorAction SilentlyContinue
$goExe = if ($goCommand) { $goCommand.Source } else { $null }
if (-not $goExe -and $env:GOROOT) {
    $goExe = Join-Path $env:GOROOT "bin\go.exe"
}
if (-not (Test-Path -LiteralPath $goExe)) {
    throw "找不到 Go 编译器。"
}

$isccCandidates = @(
    (Join-Path $env:LOCALAPPDATA "Programs\Inno Setup 6\ISCC.exe"),
    "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
    "C:\Program Files\Inno Setup 6\ISCC.exe"
)
$iscc = $isccCandidates | Where-Object { Test-Path -LiteralPath $_ } | Select-Object -First 1
if (-not $iscc) {
    throw "找不到 Inno Setup 6。请先安装 JRSoftware.InnoSetup。"
}

Push-Location $repoRoot
try {
    New-Item -ItemType Directory -Force -Path "dist" | Out-Null
    $env:GOPROXY = "https://goproxy.cn,direct"
    & $goExe test ./... -count=1
    if ($LASTEXITCODE -ne 0) { throw "Go 测试失败。" }

    & $goExe build -trimpath -ldflags "-s -w -H=windowsgui" -o "dist\ClashTrafficMonitor.exe" .
    if ($LASTEXITCODE -ne 0) { throw "Windows 应用构建失败。" }

    & $iscc "/DMyAppVersion=$Version" "installer\ClashTrafficMonitor.iss"
    if ($LASTEXITCODE -ne 0) { throw "Inno Setup 编译失败。" }

    Get-Item "dist\Mihomo-Traffic-Ledger-Setup-v$Version.exe"
} finally {
    Pop-Location
}
