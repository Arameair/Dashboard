# Isolated owner test for Dashboard v5. Never starts the production checkout.
# This is a manual smoke-test candidate, not the unfinished Version Manager.
$ErrorActionPreference = "Stop"
$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$testRoot = Join-Path $root "runtime\alpha-owner-smoke"
$binDir = Join-Path $testRoot "bin"
$binary = Join-Path $binDir "Dashboard-Alpha-Test.exe"
$dataDir = Join-Path $testRoot "data"
$projectsDir = Join-Path $testRoot "project-definitions"
$port = 18777

# Never attach to an existing Dashboard. Fail if the dedicated port is occupied.
$probe = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Parse("127.0.0.1"), $port)
try {
    $probe.Start()
} catch {
    throw "TEST PORT $port IS IN USE. Close the earlier Alpha test or choose another time. No existing Dashboard was touched."
} finally {
    $probe.Stop()
}

foreach ($dir in @($binDir, $dataDir, $projectsDir, (Join-Path $testRoot "go-cache"))) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
}

# The smoke project is inert until the owner manually selects Start.
# Never replace an existing test definition or reuse signed-in production profiles.
$sample = Join-Path $projectsDir "alpha-smoke.json"
if (-not (Test-Path -LiteralPath $sample)) {
    $json = @'
{
  "projects": [{
    "id": "owner-alpha-smoke",
    "name": "Alpha browser smoke test",
    "url": "https://example.com/",
    "color": "#3b82f6",
    "nudge": "",
    "nudge_minutes": 10,
    "auto_nudge": false,
    "save_replies": false,
    "console": false,
    "auto_start": false
  }]
}
'@
    [System.IO.File]::WriteAllText($sample, $json, [System.Text.UTF8Encoding]::new($false))
}

$previousToolchain = $env:GOTOOLCHAIN
$previousProxy = $env:GOPROXY
$previousCache = $env:GOCACHE
Push-Location -LiteralPath (Join-Path $root "dashboard")
try {
    # No automatic SDK/toolchain or module download, and no write to production data.
    $env:GOTOOLCHAIN = "local"
    $env:GOPROXY = "off"
    $env:GOCACHE = Join-Path $testRoot "go-cache"
    $goVersion = (& go version | Out-String).Trim()
    if ($goVersion -notmatch 'go version go(\d+)\.(\d+)') {
        throw "Go 1.24 or newer must already be installed. Found: $goVersion"
    }
    $major = [int]$Matches[1]
    $minor = [int]$Matches[2]
    if ($major -lt 1 -or ($major -eq 1 -and $minor -lt 24)) {
        throw "Go 1.24 or newer required (no auto-download). Found: $goVersion"
    }
    Write-Host "Checking Dashboard v5 source in isolated test checkout..."
    & go test -mod=vendor ./...
    if ($LASTEXITCODE -ne 0) { throw "Go tests/compile failed. Nothing launched." }
    & go vet -mod=vendor ./...
    if ($LASTEXITCODE -ne 0) { throw "Go vet failed. Nothing launched." }
    & go build -mod=vendor -o $binary .
    if ($LASTEXITCODE -ne 0) { throw "Build failed. Nothing launched." }
} finally {
    Pop-Location
    $env:GOTOOLCHAIN = $previousToolchain
    $env:GOPROXY = $previousProxy
    $env:GOCACHE = $previousCache
}

Write-Host ""
Write-Host "ALPHA TEST ONLY — not the Version Manager or a protected Alpha promotion"
Write-Host "URL: http://127.0.0.1:$port/"
Write-Host "Data and browser profiles: $dataDir"
Write-Host "Projects: $projectsDir (one inert sample, auto-nudge OFF)"
Write-Host "Press Ctrl+C to stop this Dashboard server; do not use Stop All on other installs."
Write-Host ""
# The v5 executable opens the page in the default browser after it starts.
# Keep it in the foreground, so the owner can stop it without process discovery.
& $binary -port $port -data $dataDir -projects $projectsDir
if ($LASTEXITCODE -ne 0) { throw "Alpha test process exited with code $LASTEXITCODE." }
