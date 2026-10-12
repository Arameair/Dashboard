param([string]$Projects = '', [switch]$NoOpen)

$ErrorActionPreference = 'Stop'
$previousCache = $env:GOCACHE
Push-Location -LiteralPath $PSScriptRoot
try {
    $env:GOCACHE = Join-Path $PSScriptRoot 'runtime\go-cache'
    New-Item -ItemType Directory -Path 'runtime\bin' -Force | Out-Null
    go build -C dashboard -o ..\runtime\bin\Dashboard.exe .
    if ($LASTEXITCODE -ne 0) { throw 'Dashboard build failed.' }
    # Dashboard is standalone; attach Teams only through its configuration files.
    # The Teams default branch is stable, which lacks dashboard/teams.json.
    if (-not $Projects) {
        $teamsRoot = Join-Path $PSScriptRoot '..\Teams'
        $teamsDefs = Join-Path $teamsRoot 'dashboard'
        $Projects = "$PSScriptRoot\projects;$teamsDefs"
        if ((Test-Path -LiteralPath $teamsRoot -PathType Container) -and
            -not (Test-Path -LiteralPath (Join-Path $teamsDefs 'teams.json') -PathType Leaf)) {
            Write-Warning 'Teams checkout found, but dashboard/teams.json is missing. Team 01-04 definitions are on Teams/main, not Teams/stable. Dashboard runs standalone; no Team cards will be imported from this checkout. Use a separate Teams/main checkout; do not switch an active Teams/stable worktree.'
        }
    }
    $launchArgs = @('-projects', $Projects)
    if ($NoOpen) { $launchArgs += '-no-open' }
    & .\runtime\bin\Dashboard.exe @launchArgs
} finally {
    $env:GOCACHE = $previousCache
    Pop-Location
}
