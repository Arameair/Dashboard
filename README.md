# Dashboard

A local Go dashboard for choosing projects, running their actions, and watching
their output. Dashboard controls and project controls have separate places in
the UI.

## Repository layout

```text
dashboard/                 Dashboard server and web UI (Go module)
projects/                  Project definitions and each project's source
  round2-original/
  round2-p1/
  teams-nudger/
  managed/                 Shared browser engine (ignored)
  runs/                    Browser profiles / saved sign-in (ignored)
  artifacts/               Generated screenshots (ignored)
runtime/                   Dashboard builds, test output, and logs (ignored)
  logs/<project-id>/last-run.log
go.work                    Lets dashboard commands run from the repo root
```

Dashboard and projects remain in one Git repository for now. The dashboard can
use an external projects folder later without changing project definitions.

## Run

Double-click `Start-Dashboard.cmd` (or run `./Start-Dashboard.ps1`). It builds the current dashboard
into `runtime/bin/`, opens its browser page, and runs until Ctrl+C. Go is required.
The existing root `Dashboard.exe` is an older build; the launcher builds the new
version without overwriting or stopping it.

Alternatively:

```powershell
go run ./dashboard
```

For environments that block executables in Go's temporary build folder, use the
launcher or build directly:

```powershell
go build -o runtime/bin/Dashboard.exe ./dashboard
./runtime/bin/Dashboard.exe
```

On the screen, choose a project from the Projects sidebar. Its actions, Finish,
Stop, and run output appear in the project pane. Dashboard settings contains the
shared engine controls. Stop all projects applies to the whole workspace.

## Project folders and logs

By default the dashboard reads `projects/`. `+ Add project` registers a program
in another folder without copying its source. Rescan reloads project definitions.

For a separate workspace later:

```powershell
./Start-Dashboard.ps1 -Projects 'C:\path\to\projects'
```

`-projects` overrides the optional, ignored `dashboard.local.json` setting:

```json
{"projects": "projects"}
```

Relative project paths in that configuration are resolved from the dashboard
repository/package root. The selected external folder must already exist.
`-logs <folder>` overrides the dashboard's `runtime/logs/` location.
`-no-browser` suppresses automatic browser opening.

Logs, rotated logs (`*.log.*`, `*.log-*`), `logs/`, browser state, test reports,
coverage output, and builds are ignored. Ignore rules do not remove files already
tracked by Git; no tracked logs were found during this change.

## Checks

```powershell
go test ./dashboard
go vet ./dashboard
./scripts/Check-Dashboard.ps1
```

The HTTP check starts an isolated dashboard and a local fixture, verifies project
loading and log placement, then stops only its own server. It does not send chat
messages or run the bundled browser projects.
