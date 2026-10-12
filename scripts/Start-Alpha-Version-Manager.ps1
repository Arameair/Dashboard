param(
    [switch]$NoBrowser,
    [string]$GoExe = "go"
)
$ErrorActionPreference = "Stop"
$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$module = Join-Path $root "tools\version-manager"
$runtime = Join-Path $root "runtime\version-manager-alpha"
$binary = Join-Path $runtime "version-manager-alpha.exe"
if (-not (Test-Path -LiteralPath (Join-Path $module "go.mod"))) { throw "Wrong checkout: missing tools/version-manager/go.mod" }
if (-not (Test-Path -LiteralPath (Join-Path $module "static\index.html"))) { throw "UI assets missing; obtain latest shared Alpha branch" }
New-Item -ItemType Directory -Path $runtime -Force | Out-Null
$old = Get-Location
try {
    Set-Location $module
    & $GoExe version
    if ($LASTEXITCODE -ne 0) { throw "Go 1.24+ is required" }
    & $GoExe build -o $binary .
    if ($LASTEXITCODE -ne 0) { throw "Go build failed; Alpha not started" }
    $env:ALPHA_VM_DATA = $runtime
    $env:ALPHA_VM_STATIC = (Join-Path $module "static")
    Write-Host "Alpha Version Manager: http://127.0.0.1:18776"
    Write-Host "Separate test only. CTRL+C stops the manager; no existing Dashboard is modified."
    & $binary
    if ($LASTEXITCODE -ne 0) { throw "Version Manager exited with code $LASTEXITCODE" }
} finally {
    Set-Location $old
}
