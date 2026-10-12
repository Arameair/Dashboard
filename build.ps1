# Build the dashboard and its bundled example programs into their expected folders.
# This does not install Chromium or execute any project action.
param([switch]$Run)

$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot
$previousWorkspace = $env:GOWORK
try {
    # Each example project has its own go.mod, outside the dashboard's go.work.
    $env:GOWORK = 'off'
    $targets = @(
        @{ Source = 'source'; Output = 'Dashboard.exe' },
        @{ Source = 'projects/round2-original/source'; Output = 'projects/round2-original/round2-original.exe' },
        @{ Source = 'projects/round2-p1/source'; Output = 'projects/round2-p1/round2-p1.exe' },
        @{ Source = 'projects/teams-nudger/source'; Output = 'projects/teams-nudger/teams-nudger.exe' }
    )
    foreach ($target in $targets) {
        $src = Join-Path $root $target.Source
        $output = Join-Path $root $target.Output
        Write-Host "Building $($target.Output)..."
        Push-Location $src
        try {
            & go build -o $output .
            if ($LASTEXITCODE -ne 0) { throw "Build failed: $($target.Source) (exit $LASTEXITCODE)" }
        } finally {
            Pop-Location
        }
    }
} finally {
    if ($null -eq $previousWorkspace) {
        Remove-Item Env:GOWORK -ErrorAction SilentlyContinue
    } else {
        $env:GOWORK = $previousWorkspace
    }
}
Write-Host 'Build complete. Project binaries are ready.'
if ($Run) {
    & (Join-Path $root 'Dashboard.exe')
}
