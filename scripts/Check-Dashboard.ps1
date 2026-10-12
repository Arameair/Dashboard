param([switch]$KeepRunning)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$checkRoot = Join-Path $repo ('runtime\checks\' + [guid]::NewGuid().ToString('N'))
$projectRoot = Join-Path $checkRoot 'projects'
$fixtureFolder = Join-Path $projectRoot 'fixture'
$logs = Join-Path $checkRoot 'logs'
$dashboardExe = Join-Path $checkRoot 'Dashboard.exe'
$fixtureExe = Join-Path $fixtureFolder 'Fixture.exe'
$stdout = Join-Path $checkRoot 'dashboard.stdout.log'
$stderr = Join-Path $checkRoot 'dashboard.stderr.log'
$server = $null
$success = $false
$previousCache = $env:GOCACHE
New-Item -ItemType Directory -Path $fixtureFolder -Force | Out-Null
Push-Location -LiteralPath $repo
try {
    $env:GOCACHE = Join-Path $repo 'runtime\go-cache'
    go build -o $dashboardExe ./dashboard
    if ($LASTEXITCODE -ne 0) { throw 'Dashboard build failed.' }
    $fixtureSource = Join-Path $checkRoot 'fixture.go'
    [IO.File]::WriteAllText($fixtureSource, "package main`nimport `"fmt`"`nfunc main() { fmt.Println(`"offline fixture output`") }`n")
    go build -o $fixtureExe $fixtureSource
    if ($LASTEXITCODE -ne 0) { throw 'Fixture build failed.' }
    $definition = @{
        name = 'Offline check'; exe = 'Fixture.exe'; group = 'Local checks'
        actions = @(@{label = 'Run fixture'; args = @(); primary = $true})
    } | ConvertTo-Json -Depth 5
    $projectJSON = Join-Path $fixtureFolder 'project.json'
    [IO.File]::WriteAllText($projectJSON, $definition)
    $server = Start-Process -FilePath $dashboardExe -WorkingDirectory $repo -WindowStyle Hidden -PassThru -ArgumentList @('-no-browser', '-projects', ('"' + $projectRoot + '"'), '-logs', ('"' + $logs + '"')) -RedirectStandardOutput $stdout -RedirectStandardError $stderr
    $url = $null
    for ($i = 0; $i -lt 50; $i++) {
        $server.Refresh()
        if ($server.HasExited) { throw "Dashboard exited: $(Get-Content -LiteralPath $stderr -Raw)" }
        $output = Get-Content -LiteralPath $stdout -Raw -ErrorAction SilentlyContinue
        if ($output -match 'http://127\.0\.0\.1:\d+/') { $url = $Matches[0]; break }
        Start-Sleep -Milliseconds 100
    }
    if (-not $url) { throw 'Dashboard startup timed out.' }
    $state = Invoke-RestMethod ($url + 'api/state')
    if ($state.folder -ne $projectRoot -or $state.logs -ne $logs -or $state.projects.Count -ne 1 -or -not $state.projects[0].ExeFound) {
        throw 'External project loading failed.'
    }
    $initialLog = Invoke-RestMethod ($url + 'api/log?p=fixture')
    if ($null -eq $initialLog.lines -or $initialLog.lines.Count -ne 0) { throw 'Idle log must be an empty array.' }
    $html = (Invoke-WebRequest $url -UseBasicParsing).Content
    foreach ($text in @('Dashboard settings', 'Selected project', 'Project actions', 'Run output', 'project-sidebar')) {
        if (-not $html.Contains($text)) { throw "Missing UI section: $text" }
    }
    Invoke-RestMethod ($url + 'api/run?p=fixture&a=0') -Method Post | Out-Null
    $logPath = Join-Path $logs 'fixture\last-run.log'
    for ($i = 0; $i -lt 50; $i++) {
        if (Test-Path -LiteralPath $logPath) { break }
        Start-Sleep -Milliseconds 100
    }
    if (-not (Test-Path -LiteralPath $logPath) -or -not (Get-Content -LiteralPath $logPath -Raw).Contains('offline fixture output')) {
        throw 'Dashboard log was not saved in the dashboard log folder.'
    }
    if (Test-Path -LiteralPath (Join-Path $fixtureFolder 'last-run.log')) { throw 'A dashboard log leaked into the project folder.' }
    $state = Invoke-RestMethod ($url + 'api/state')
    if ($state.projects[0].State -ne 'ok') { throw 'Fixture did not finish successfully.' }
    $hiddenJSON = Join-Path $checkRoot 'project.saved.json'
    Move-Item -LiteralPath $projectJSON -Destination $hiddenJSON
    Invoke-RestMethod ($url + 'api/rescan') -Method Post | Out-Null
    $empty = Invoke-RestMethod ($url + 'api/state')
    if ($null -eq $empty.projects -or $empty.projects.Count -ne 0) { throw 'Empty workspace must return an empty project array.' }
    Move-Item -LiteralPath $hiddenJSON -Destination $projectJSON
    Invoke-RestMethod ($url + 'api/rescan') -Method Post | Out-Null
    Write-Output 'PASS: project loading, isolated run, log placement, empty workspace, and UI sections over HTTP.'
    $success = $true
    if ($KeepRunning) { Write-Output "Preview: $url (owned test PID $($server.Id))" }
} finally {
    if ($server -and (-not $KeepRunning -or -not $success)) {
        $server.Refresh()
        if (-not $server.HasExited) {
            if ($server.Path -ne $dashboardExe) { throw 'Refusing to stop a process with an unexpected executable.' }
            Stop-Process -Id $server.Id
            $server.WaitForExit()
        }
        Write-Output "Stopped owned check server PID $($server.Id)."
    }
    $env:GOCACHE = $previousCache
    Pop-Location
}
