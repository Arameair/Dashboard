param([string]$Projects = '', [switch]$NoBrowser)

$ErrorActionPreference = 'Stop'
$previousCache = $env:GOCACHE
Push-Location -LiteralPath $PSScriptRoot
try {
    $env:GOCACHE = Join-Path $PSScriptRoot 'runtime\go-cache'
    New-Item -ItemType Directory -Path 'runtime\bin' -Force | Out-Null
    go build -o runtime\bin\Dashboard.exe ./dashboard
    if ($LASTEXITCODE -ne 0) { throw 'Dashboard build failed.' }
    $launchArgs = @()
    if ($Projects) { $launchArgs += @('-projects', $Projects) }
    if ($NoBrowser) { $launchArgs += '-no-browser' }
    & .\runtime\bin\Dashboard.exe @launchArgs
} finally {
    $env:GOCACHE = $previousCache
    Pop-Location
}
