param([string]$Projects = '', [switch]$NoOpen)

$ErrorActionPreference = 'Stop'
$previousCache = $env:GOCACHE
Push-Location -LiteralPath $PSScriptRoot
try {
    $env:GOCACHE = Join-Path $PSScriptRoot 'runtime\go-cache'
    New-Item -ItemType Directory -Path 'runtime\bin' -Force | Out-Null
    go build -C dashboard -o ..\runtime\bin\Dashboard.exe .
    if ($LASTEXITCODE -ne 0) { throw 'Dashboard build failed.' }
    # Default: this repo's projects\ folder plus the Teams repo, if cloned next to this one.
    if (-not $Projects) { $Projects = "$PSScriptRoot\projects;$PSScriptRoot\..\Teams\dashboard" }
    $launchArgs = @('-projects', $Projects)
    if ($NoOpen) { $launchArgs += '-no-open' }
    & .\runtime\bin\Dashboard.exe @launchArgs
} finally {
    $env:GOCACHE = $previousCache
    Pop-Location
}
