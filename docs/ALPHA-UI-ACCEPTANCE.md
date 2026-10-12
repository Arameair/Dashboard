# Alpha Dashboard — UI acceptance checklist

**Owner task:** #6 · **Specialist task:** #9 · **Seat:** DASH-04-SPEC-20261012  
**Status:** CHECKLIST PREPARED; ALPHA NOT ACCEPTED; ALL EXECUTABLE SCENARIOS NOT RUN  
**Authorized branch:** `prototype/dashboard-reliability` (existing branch only)  
**Stable baseline / owner's base (full SHA):** `a2d8c72482e3c9b5e7cb630dd8e96777437df074`  
**Prototype HEAD at checklist authoring (full SHA):** `8f129e76321bf7505491b7c9aa34683c4bff57cc`  
**Code changes by this task:** none. This document is the only allowed write path.

## Scope and non-negotiable safeguards

This is a **manual/offline, local synthetic-fixture** checklist. Do not exercise bundled real projects or `teams-nudger`; do not send ChatGPT or other external messages; do not authenticate, examine stored browser profiles, install a browser engine, or use paid services. The operator must use an isolated temporary project directory containing only known-safe local fixtures. Do not execute a project merely to see whether it is safe. All test results below are **NOT RUN**, not PASS. No approval to merge or promote follows from publishing this checklist.

Before any run, record machine OS/version, Go and PowerShell versions, clean branch and full `git rev-parse HEAD`, test directory, process ports, start/end timestamps, exact command and relevant stdout/stderr. A local browser visual observation must identify the browser/version and include screenshot or notes. Redact usernames, secrets and local personal paths in shared receipts. Re-check HEAD after changes and rerun impacted tests.

## Safe fixture and startup procedure (operator-run; not executed by Specialist 04)

1. Prepare a **separate isolated checkout** of existing `prototype/dashboard-reliability`. Never reuse the owner's installed Dashboard, active profiles, or live project directories; do not switch any protected lane. Ensure the checkout is clean. Before testing record `git status --short`, `git branch --show-current`, and `git rev-parse HEAD`.
2. Make a temporary fixture root containing at least two project directories, `fixture-a/` and `fixture-b/`, with self-contained `project.json` manifests conforming to the current `dashboard/main.go` project/action schema. Use only locally controlled programs that print deterministic lines and optionally wait for stdin; never point manifests to any existing project executable. A third fixture points to a **nonexistent** executable to test error display. Keep all fixture logs in a separate temporary folder. Do not commit fixture files.
3. From the isolated prototype checkout, run `go test ./dashboard` and `go vet ./dashboard`. On Windows with PowerShell available, run `./scripts/Check-Dashboard.ps1` (only if confirmed to use isolated synthetic fixtures). Record actual results, not assumptions.
4. Launch the dashboard with explicit safe fixture/log directories using `go run ./dashboard -projects "<isolated-fixture-root>" -logs "<isolated-log-root>" -no-browser` where supported by the checked-out `dashboard/main.go`. Open only its local HTTP page in a separate browser context with no personal profile. Capture the printed loopback URL and observe it locally. Stop the temporary server and verify no fixture process remains at teardown.
5. Windows launcher cold/warm checks are **separate**: follow issue #3 / Lead's safe-launch procedure for a separate clean checkout. Never point launcher/default paths to live projects. Do not call an unexecuted launcher check PASS.

## Acceptance scenarios

For each case, capture **PASS / FAIL / NOT RUN**, fixture used, reproduction steps, expected vs observed behavior, exact HEAD SHA, OS/browser version and evidence reference. PASS requires an executed observation; reading frontend/backend code alone is not a PASS.

| ID | Operator steps using safe fixtures | Expected result | Result | Evidence |
|---|---|---|---|---|
| UI-01 | Start isolated Dashboard twice (fresh and warm), load local root URL; observe response and console | Server starts cleanly, page loads and identifies PROTOTYPE; no change to installed Dashboard | NOT RUN | No local execution receipt |
| UI-02 | Load fixture root with two valid projects and one missing-executable entry; use Rescan | All valid projects listed with correct labels; errors reflected without hiding others | NOT RUN | No local execution receipt |
| UI-03 | Switch A → B → A via sidebar while idle; resize narrow/wide | Selection, details, action labels and responsive navigation correspond to selected project | NOT RUN | No browser receipt |
| UI-04 | Select missing executable fixture and inspect screen without pressing Run | Missing-program notice shows correct path; no misleading successful-start state | NOT RUN | No browser receipt |
| UI-05 | Start only safe fixture A, observe buttons in idle/running/waiting/completed states | Run actions disabled while busy; Finish enabled **only** when waiting; Stop enabled only when busy; Stop All enabled iff any active run | NOT RUN | No fixture receipt |
| UI-06 | Run fixture that waits for Enter and press Finish; separately run fixture and press Stop; then run two fixtures and press Stop All | Correct targeted finish/stop, terminal state and no lingering child; Stop All stops active fixtures only | NOT RUN | No fixture receipt |
| UI-07 | Stream known sequential fixture lines; compare UI output with isolated last-run logs | Output stays ordered, correctly escaped and appropriately scrolls; logs are stored under isolated log directory | NOT RUN | No fixture receipt |
| UI-08 | Have safe fixture emit >4,000 numbered lines while polling; use recorded sentinels after buffer rollover | No stalled polling/duplicate tail; recent lines visible and absolute cursor advances; record before/after log cursor if observable | NOT RUN | No fixture receipt |
| UI-09 | Delay A's `/api/log` response with local devtools or a synthetic local HTTP fixture; select B before A completes | Late response from A never appears under B; B eventually polls correctly | NOT RUN | No browser/network receipt |
| UI-10 | Start A, stop/finish and start it again; compare distinct run tokens and output | New run resets old cursor/display and shows new lines, not stale previous-run data | NOT RUN | No fixture receipt |
| UI-11 | While fixture A is active, interrupt only the isolated HTTP server/network; inspect UI, then restore | Disconnected indicator appears; Run/Finish/Stop/Stop All become disabled while status is stale; controls recover based on fresh server state | NOT RUN | No browser receipt |
| UI-12 | Use fixture names/descriptions/log lines containing HTML-special text such as `<script>` (display-only string) | Strings display as text; no HTML or script execution, and no unsafe insertion via manifest/logs | NOT RUN | No browser receipt |
| UI-13 | Simulate a non-2xx `/api/state` response through a local HTTP fixture and restore a valid response | UI treats failed response as disconnected and recovers on a successful poll | NOT RUN | No browser/network receipt |
| UI-14 | Verify workspace/project actions with no projects configured | Empty-state text appears; no stale selection, log, action controls or Stop All state | NOT RUN | No browser receipt |

## Execution receipt template

Copy one line per execution, including failures. Do not convert absence of evidence into PASS.

| Field | Record |
|---|---|
| Scenario | UI-__ |
| Status | PASS / FAIL / NOT RUN |
| OS / browser / versions | NOT RUN — environment unavailable |
| Stable / HEAD SHA | Stable: `a2d8c72482e3c9b5e7cb630dd8e96777437df074`; prototype snapshot: `8f129e76321bf7505491b7c9aa34683c4bff57cc` |
| Command / action | NOT RUN |
| Expected vs observed | Expected as defined above; observed NOT RUN |
| Output excerpt / screenshot / log | None |
| Reviewer / date | Unassigned |

## Verification matrix — as authored

| Check | Status | OS / SHA / evidence |
|---|---|---|
| `go test ./dashboard` | NOT RUN | No isolated executable checkout receipt |
| `go vet ./dashboard` | NOT RUN | No isolated executable checkout receipt |
| `./scripts/Check-Dashboard.ps1` | NOT RUN | Windows/PowerShell execution evidence unavailable |
| Windows cold/warm launcher on separate isolated checkout | NOT RUN | Issue #3 blocker |
| Synthetic-fixture browser scenarios UI-01 through UI-14 | NOT RUN | Local browser/fixture execution evidence unavailable |
| GitHub source inspection / checklist composition | STATIC ONLY | Cannot establish runtime correctness |

## Dependencies and handoff

- **Issue #4**: the UI disconnected-state change on the prototype is committed, but executable UI regression checks (including log rollover, project switching and recovery) are NOT RUN. This document does not close #4.
- **Issue #3**: Windows/PowerShell cold/warm launcher and HTTP smoke evidence remains blocked. This checklist does not certify launcher safety.
- **Issue #7**: Specialist 03 owns the stop-during-start race investigation in disjoint backend paths. Static analysis is not deterministic reproduction.
- **Issue #6 / owner Alpha build**: checklist availability is **not** owner Alpha acceptance, deployment clearance, or approval to merge. Lead must collect receipts and request owner's decision separately.

**Stop:** Do not modify `main`, `stable`, `alpha`, issue #4 frontend files, issue #7 backend files or any other file for #9. Do not create branches, merge, force-push, launch real project actions or use external automation.
