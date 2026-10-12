# Pull the latest prototype, compile missing/revised programs, and launch safely.
# Runs exclusively inside a clone of prototype/dashboard-reliability.
param(
    [switch]$NoUpdate,
    [switch]$SkipExamples,
    [switch]$NoBrowser
)

$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot
$branch = 'prototype/dashboard-reliability'
$previousWork = $env:GOWORK

Push-Location -LiteralPath $root
try {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw 'Go is required. Install Go, then rerun Start-Prototype.cmd.'
    }
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
        throw 'Git is required to safely update this prototype.'
    }
    if (-not (Test-Path -LiteralPath (Join-Path $root '.git'))) {
        throw 'This must run from a Git clone of the prototype branch, not a downloaded ZIP.'
    }

    $currentBranch = (& git branch --show-current).Trim()
    if ($LASTEXITCODE -ne 0 -or $currentBranch -ne $branch) {
        throw "Expected branch '$branch'; found '$currentBranch'. Your other checkouts were not changed."
    }

    if (-not $NoUpdate) {
        $changes = @(& git status --porcelain --untracked-files=normal)
        if ($LASTEXITCODE -ne 0) { throw 'Could not check local changes.' }
        if ($changes.Count -gt 0) {
            throw 'Prototype checkout contains local changes. Update skipped to protect them. Use a clean prototype checkout.'
        }
        Write-Host 'Checking GitHub for prototype updates...'
        & git fetch --no-tags origin $branch
        if ($LASTEXITCODE -ne 0) {
            Write-Warning 'GitHub was not reachable. Launching your already installed prototype instead.'
        } else {
            & git merge --ff-only FETCH_HEAD
            if ($LASTEXITCODE -ne 0) {
                throw 'Cannot fast-forward prototype. No branch was reset or forced.'
            }
        }
    }

    $revision = (& git rev-parse --short=12 HEAD).Trim()
    if ($LASTEXITCODE -ne 0 -or -not $revision) { throw 'Could not read the prototype revision.' }
    Write-Host "Prototype revision: $revision"

    $prototypeRoot = Join-Path $root 'runtime\prototype'
    $releaseFolder = Join-Path $prototypeRoot $revision
    $sessionFolder = Join-Path $prototypeRoot 'sessions'
    New-Item -ItemType Directory -Path $releaseFolder, $sessionFolder -Force | Out-Null

    if (-not $SkipExamples) {
        $projects = @(
            @{ Dir = 'projects/round2-original/source'; Out = 'projects/round2-original/round2-original.exe' },
            @{ Dir = 'projects/round2-p1/source'; Out = 'projects/round2-p1/round2-p1.exe' },
            @{ Dir = 'projects/teams-nudger/source'; Out = 'projects/teams-nudger/teams-nudger.exe' }
        )
        $marker = Join-Path $releaseFolder 'examples-built.txt'
        $missing = @($projects | Where-Object { -not (Test-Path -LiteralPath (Join-Path $root $_.Out)) })
        if (-not (Test-Path -LiteralPath $marker) -or $missing.Count -gt 0) {
            $failures = @()
            $env:GOWORK = 'off'
            foreach ($project in $projects) {
                Write-Host "Building $($project.Out)..."
                Push-Location -LiteralPath (Join-Path $root $project.Dir)
                try {
                    & go build -o (Join-Path $root $project.Out) .
                    if ($LASTEXITCODE -ne 0) { $failures += $project.Out }
                } finally {
                    Pop-Location
                }
            }
            $env:GOWORK = $previousWork
            if ($failures.Count -eq 0) {
                Set-Content -LiteralPath $marker -Value $revision
            } else {
                Write-Warning ("Some example programs did not build: " + ($failures -join ', '))
                Write-Warning 'The dashboard will still open. Those project buttons may be unavailable until dependencies/builds are fixed.'
            }
        }
    }

    $compiled = Join-Path $releaseFolder 'Dashboard.exe'
    if (-not (Test-Path -LiteralPath $compiled)) {
        Write-Host 'Building prototype dashboard...'
        $env:GOWORK = $previousWork
        & go build -o $compiled ./dashboard
        if ($LASTEXITCODE -ne 0) { throw 'Prototype dashboard failed to compile.' }
    }

    # Each launched server uses a private copy, so Windows never needs to overwrite
    # an EXE currently running from a previous preview.
    $session = Join-Path $sessionFolder ('Dashboard-' + [guid]::NewGuid().ToString('N') + '.exe')
    Copy-Item -LiteralPath $compiled -Destination $session
    Write-Host 'Opening Dashboard PROTOTYPE. Close this console to end this preview.'
    try {
        $launchArgs = @()
        if ($NoBrowser) { $launchArgs += '-no-browser' }
        & $session @launchArgs
        if ($LASTEXITCODE -ne 0) { throw "Dashboard exited with code $LASTEXITCODE" }
    } finally {
        Remove-Item -LiteralPath $session -Force -ErrorAction SilentlyContinue
    }
} finally {
    if ($null -eq $previousWork) {
        Remove-Item Env:GOWORK -ErrorAction SilentlyContinue
    } else {
        $env:GOWORK = $previousWork
    }
    Pop-Location
}
