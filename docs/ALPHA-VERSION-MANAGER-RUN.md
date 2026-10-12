# Alpha Version Manager — isolated Windows operator guide

**Project:** DASH-A4-PRJ-20261012. **Development:** Issue #15. **Status:** backend lifecycle incomplete; **not** ready to launch real Dashboard builds.

## Preconditions
- Separate checkout of existing branch `experiment/a4-version-manager-v5-20261012`; keep the main/stable/alpha installations and working data untouched.
- Windows PowerShell, Go 1.24 or newer, and local authorization to run the project scripts.
- No user account/browser sessions needed. No GitHub fetch, sign-in, or automatic messaging.

## Build, inspect and stop
From a separate, clean checkout:
```powershell
git rev-parse HEAD
go version
cd tools/version-manager
go test ./...
go vet ./...
go build ./...
cd ../..
.\scripts\Start-Alpha-Version-Manager.ps1
```
The script builds the isolated executable under `runtime/version-manager-alpha/`, uses only that directory for the catalog, and serves the UI at `http://127.0.0.1:18776`. Stop it using Ctrl+C in its PowerShell console. It does not open a browser automatically. If the operating system blocks the executable, STOP and record the error; do not bypass protection.

Confirm `GET http://127.0.0.1:18776/api/versions` and `/api/status` from a fresh test session. `GET /` should show the static UI. Until safe child-process lifecycle is implemented and verified, Launch/Stop/Rollback return HTTP 501; do **not** use this Alpha to manage a live Dashboard installation.

## Data isolation, reset and rollback
All manager-generated data is intended to live under `runtime/version-manager-alpha/`; the existing Dashboard binary, `%LOCALAPPDATA%\TeamsDashboard\v5`, user browser profiles and data are outside its permitted write scope. To reset **only the manager-owned synthetic data** after stopping the process, back up or rename the catalog file in that isolated directory; do not recursively delete unknown folders. For a failed test, return to the preexisting Dashboard using its unchanged shortcut/checkout. This is an operator fallback, **not** an implemented in-app Rollback.

## Evidence matrix
| Verification | Result |
| --- | --- |
| Source readback / script presence | PASS — static only |
| `go test ./...` | NOT RUN |
| `go vet ./...` | NOT RUN |
| `go build ./...` | NOT RUN |
| Windows PowerShell start + local UI | NOT RUN |
| Confirmed selected child PID/port and readiness | NOT IMPLEMENTED |
| Stop/restart/collision/rollback | NOT IMPLEMENTED |

Record exact checkout SHA, Windows version, command output, response codes, and observed errors during operator testing. A catalog-only page is not the owner-accepted runnable Alpha.

## Security
Never register or execute unreviewed binaries, bypass Smart App Control, interact with browser profiles, run project actions, or mutate the working Dashboard to exercise this experiment. The next milestone is secure digest-bound executable validation and owned process/port/profile lifecycle in Issue #15.
