# Dashboard prototype — separate preview

**Branch:** `prototype/dashboard-reliability`

This is a reviewable, isolated variant of the current `dashboard/` app. It does not change `main`, the installed dashboard, or other checked-out projects.

## Install once (PowerShell)

```powershell
git clone --branch prototype/dashboard-reliability --single-branch https://github.com/Arameair/Dashboard.git "$env:USERPROFILE\Dashboard-Prototype"
& "$env:USERPROFILE\Dashboard-Prototype\Start-Prototype.cmd"
```

GitHub may ask you to authenticate to clone the private repository.

## See the latest prototype (every subsequent time)

**Double-click `Start-Prototype.cmd`** inside `Dashboard-Prototype`.

The launcher checks for branch updates, only fast-forwards (never resets your files), builds the dashboard when its revision changes, builds bundled sample project executables when needed, and opens the browser dashboard. The UI title and header show **PROTOTYPE**. Closing the preview console exits the dashboard; use **Stop all projects** first to stop child programs.

To update or start it via PowerShell:

```powershell
& "$env:USERPROFILE\Dashboard-Prototype\Start-Prototype.ps1"
```

If GitHub is unavailable, it runs the already-downloaded revision. A dirty checkout or unexpected branch causes the updater to refuse changes.

The prototype build runs from `runtime/prototype/` and uses a per-session executable copy so it cannot overwrite a Windows executable that is currently open. All build output stays ignored by Git.

## Optional controls

```powershell
.\Start-Prototype.ps1 -NoUpdate           # local version only
.\Start-Prototype.ps1 -SkipExamples       # skip compiling the Playwright example programs
.\Start-Prototype.ps1 -NoBrowser          # dashboard server without opening the browser
go test ./dashboard                       # unit tests
.\scripts\Check-Dashboard.ps1             # isolated HTTP smoke test
```

The Playwright examples have separate dependencies and require installing their browser engine before browser actions work. If those example programs fail to build (e.g. when offline), the dashboard still starts and reports the missing executable.

## What changed

- Responsive dashboard stays in the newer `dashboard/` layout from `main`.
- A visible prototype label differentiates it from the installed version.
- Log polling uses absolute cursors and a run token so restarting or trimming a log does not freeze the output.
- Project input writes do not hold the global application lock.
- Run completion drains logs deterministically; process-start failures return errors and close pipes.
- Project discovery reports directory read errors.
- Navigation avoids rebuilding identical buttons every polling interval.
- Added regression tests for log rollover, restart, empty logs, and missing projects folder.

**Verification:** Unit and HTTP checks can be run locally; live ChatGPT/browser automation is not covered by these tests. Do not promote the prototype into `main` until local checks and owner review pass.
