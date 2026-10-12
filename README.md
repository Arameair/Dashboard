# Dashboard

A small Windows-friendly Go dashboard for switching between local projects, starting and stopping programs, and viewing live logs. Project definitions live in `projects/*/project.json`.

## Quick start (Windows / PowerShell)

Requires Go 1.22 or later and internet access for the first download of project dependencies.

```powershell
git clone https://github.com/Arameair/Dashboard.git
Set-Location Dashboard
pwsh -File ./build.ps1 -Run
```

The build script compiles `Dashboard.exe` and the three example project executables. It does **not** download the Playwright browser engine or start any project action. The `-Run` switch launches the dashboard after building; omit it if you only want to build.

Once the dashboard opens, click **Install browser engine** once before running a Playwright demo. The dashboard binds only to `127.0.0.1`. Closing its console stops the dashboard; use **Stop all** first to close started projects cleanly.

## Run the dashboard directly from Go

```powershell
go run ./source
```

The root `go.work` makes this command work from the repository directory. The dashboard resolves `projects/` from the executable directory first and falls back to the checkout/current working directory. **Example project executables still need to be built** by `build.ps1` before their buttons can launch them.

To use an external projects folder, set the `DASHBOARD_PROJECTS_DIR` environment variable to its full path before starting.

## Repository layout

- `source/`: dashboard web server and embedded HTML/JS interface
- `projects/*/project.json`: project definitions and action arguments
- `projects/*/source/`: example project Go modules
- `build.ps1`: compile dashboard and bundled projects on Windows
- `START-HERE.txt`: notes for the separately packaged desktop distribution

## Verify

```powershell
go test ./source
go vet ./source
```

Browser automation tests and sign-in require a locally installed Playwright engine and are **not** exercised by these commands. Do not infer that the Teams nudger or live ChatGPT interactions are validated from a successful Go unit test.

Runtime artifacts (`projects/managed`, `projects/runs`, `projects/artifacts`, project logs, and compiled binaries) are intentionally excluded from Git.
