# CI/owner-independent Windows smoke check. Optional browser test uses only an offline blank page. Never interacts with ChatGPT.
param([switch]$BrowserSmoke)
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$testRoot = Join-Path ([System.IO.Path]::GetTempPath()) ('dashboard-alpha-ci-' + [guid]::NewGuid().ToString('N'))
$dataDir = Join-Path $testRoot 'data'
$projectsDir = Join-Path $testRoot 'projects'
$bin = Join-Path $testRoot 'Dashboard-Alpha-Test.exe'
$out = Join-Path $testRoot 'server.stdout.log'
$err = Join-Path $testRoot 'server.stderr.log'
$server = $null
$previousToolchain = $env:GOTOOLCHAIN
$previousProxy = $env:GOPROXY
$previousCache = $env:GOCACHE

New-Item -ItemType Directory -Force -Path $projectsDir, $dataDir | Out-Null
try {
    $fixture = @'
{"projects":[{"id":"alpha-test-ci","name":"Alpha CI offline test","url":"about:blank","nudge":"","nudge_minutes":10,"auto_nudge":false,"auto_start":false,"save_replies":false,"console":false}]}
'@
    [System.IO.File]::WriteAllText((Join-Path $projectsDir 'ci.json'), $fixture, [System.Text.UTF8Encoding]::new($false))
    $env:GOTOOLCHAIN = 'local'
    $env:GOPROXY = 'off'
    $env:GOCACHE = Join-Path $testRoot 'go-cache'
    Push-Location -LiteralPath (Join-Path $repo 'dashboard')
    try {
        & go test -mod=vendor ./...
        if ($LASTEXITCODE -ne 0) { throw 'Go tests failed' }
        & go vet -mod=vendor ./...
        if ($LASTEXITCODE -ne 0) { throw 'Go vet failed' }
        & go build -mod=vendor -o $bin .
        if ($LASTEXITCODE -ne 0) { throw 'Go build failed' }
    } finally {
        Pop-Location
    }

    # Ask Windows for a free loopback port; start only our own process with a dedicated data root.
    $probe = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
    $probe.Start()
    try { $port = $probe.LocalEndpoint.Port } finally { $probe.Stop() }
    $argsLine = '-no-open -port {0} -data "{1}" -projects "{2}"' -f $port, $dataDir, $projectsDir
    $server = Start-Process -FilePath $bin -ArgumentList $argsLine -PassThru -RedirectStandardOutput $out -RedirectStandardError $err
    $base = "http://127.0.0.1:$port/"
    $ready = $false
    for ($i = 0; $i -lt 80; $i++) {
        $server.Refresh()
        if ($server.HasExited) { throw ('Dashboard terminated: ' + (Get-Content $err -Raw -ErrorAction SilentlyContinue)) }
        try {
            $ping = Invoke-RestMethod -Uri ($base + 'api/ping') -TimeoutSec 2
            if ($ping -eq 'teams-dashboard v5') { $ready = $true; break }
        } catch { Start-Sleep -Milliseconds 250 }
    }
    if (-not $ready) { throw 'Dashboard did not become ready within 20 seconds' }
    $state = Invoke-RestMethod -Uri ($base + 'api/state') -TimeoutSec 4
    $html = (Invoke-WebRequest -Uri $base -UseBasicParsing -TimeoutSec 4).Content
    if ($state.version -ne 'v5') { throw 'Unexpected version' }
    if ($state.data -ne $dataDir) { throw 'State reports an unexpected data directory' }
    if (@($state.projects).Count -ne 1) { throw 'Expected exactly one isolated project' }
    $p = @($state.projects)[0]
    if ($p.id -ne 'alpha-test-ci' -or $p.state -ne 'stopped') { throw 'Fixture identity/state mismatch' }
    if ($p.auto_nudge -ne $false -or $p.auto_start -ne $false) { throw 'Unsafe auto-start/nudge state' }
    if ($html -notmatch 'Teams Dashboard') { throw 'UI HTML missing' }
    if (-not (Test-Path -LiteralPath (Join-Path $dataDir 'config.json'))) { throw 'Isolated config not created' }
    Write-Output "PASS: Go tests, vet, vendor build, HTTP ping, UI page, isolated sample, no auto-start/nudge."

    if ($BrowserSmoke) {
        if ([string]::IsNullOrWhiteSpace($state.settings.browser_found)) {
            throw ("Browser smoke requested, but Chrome/Edge was not found: " + $state.settings.browser_error)
        }
        $startUrl = $base + 'api/act?id=alpha-test-ci&a=start'
        $stopUrl = $base + 'api/act?id=alpha-test-ci&a=stop'
        $started = Invoke-RestMethod -Method Post -Uri $startUrl -ContentType 'application/json' -Body '{}' -TimeoutSec 10
        if ($started.ok -ne $true) { throw 'Browser Start was rejected' }
        $running = $false
        for ($i = 0; $i -lt 120; $i++) {
            Start-Sleep -Milliseconds 250
            $current = Invoke-RestMethod -Uri ($base + 'api/state') -TimeoutSec 3
            $p = @($current.projects)[0]
            if ($p.state -eq 'running') { $running = $true; break }
            if ($p.state -eq 'stopped' -and -not [string]::IsNullOrWhiteSpace($p.err)) {
                throw ('Browser launch failed: ' + $p.err)
            }
        }
        if (-not $running) { throw 'Browser never reported RUNNING within 30s' }
        # A running browser is insufficient for our dashboard: prove its live preview and log.
        $imageReady = $false
        for ($i = 0; $i -lt 50; $i++) {
            Start-Sleep -Milliseconds 400
            try {
                $image = Invoke-WebRequest -Uri ($base + 'api/shot?id=alpha-test-ci') -UseBasicParsing -TimeoutSec 4
                if ($image.StatusCode -eq 200 -and $image.Headers['Content-Type'] -match 'image/jpeg' -and $image.RawContentLength -gt 100) {
                    $imageReady = $true
                    break
                }
            } catch {}
        }
        if (-not $imageReady) { throw 'Browser launched but Dashboard never returned a JPEG preview' }
        $log = (Invoke-WebRequest -Uri ($base + 'api/log?id=alpha-test-ci') -UseBasicParsing -TimeoutSec 4).Content
        if ($log -notmatch 'browser ready') { throw 'Browser launch log unavailable' }
        Write-Output "PASS: live browser preview JPEG and project-specific launch log."
        $stop = Invoke-RestMethod -Method Post -Uri $stopUrl -ContentType 'application/json' -Body '{}' -TimeoutSec 10
        if ($stop.ok -ne $true) { throw 'Browser Stop was rejected' }
        $stopped = $false
        for ($i = 0; $i -lt 80; $i++) {
            Start-Sleep -Milliseconds 250
            $current = Invoke-RestMethod -Uri ($base + 'api/state') -TimeoutSec 3
            if (@($current.projects)[0].state -eq 'stopped') { $stopped = $true; break }
        }
        if (-not $stopped) { throw 'Browser did not stop within 20s' }
        Write-Output "PASS: one real Chrome/Edge isolated browser launched and stopped on Windows."
    }
} finally {
    if ($server) {
        if ($BrowserSmoke) {
            try { Invoke-RestMethod -Method Post -Uri ("http://127.0.0.1:$port/api/act?id=alpha-test-ci&a=stop") -ContentType 'application/json' -Body '{}' -TimeoutSec 4 | Out-Null } catch {}
        }
        $server.Refresh()
        if (-not $server.HasExited) {
            # Only terminate the exact PID returned by Start-Process for this disposable binary.
            Stop-Process -Id $server.Id -ErrorAction Stop
            $server.WaitForExit()
        }
    }
    $env:GOTOOLCHAIN = $previousToolchain
    $env:GOPROXY = $previousProxy
    $env:GOCACHE = $previousCache
    # Keep testRoot on failure for local inspection; CI ephemeral runner clears it automatically.
}
