# Version Manager frontend acceptance — issue #16

Seat: DASH-A4-04-MEDIUM-20261012. Scope: standalone UI files only, on `experiment/a4-version-manager-v5-20261012`. Backend is owned by #15. This file records **test procedures**, not fictional passes.

## Safe fixture requirements
Create an isolated synthetic HTTP fixture on a private loopback port implementing exactly the API in issue #16. Do not point it at the owner's running Dashboard installation or personal browser profiles. Start with four synthetic catalog entries: Main protected/installed/runnable, Alpha installed/runnable, Experimental source-only/not runnable, and Experimental installed/not runnable. Include selectedId, runningId and status. Use only test fixture data, and never send messages or sign in. Record OS, browser, exact commit SHA, fixture command, HTTP requests/responses, screenshots and PASS/FAIL/NOT RUN before closing this checklist.

| Case | Procedure / expected behavior | Status |
|---|---|---|
| S01 | Search name, branch, channel and SHA; four fixtures filter correctly with case-insensitive matching | NOT RUN |
| S02 | Select entry; compare selected vs running labels; clicking does not launch | NOT RUN |
| S03 | Add existing build form; validation errors surfaced; registration does not execute any binary | NOT RUN |
| S04 | Attempt invalid executable; server rejection stays visible and entry not claimed runnable | NOT RUN |
| S05 | Remove protected or running entry; controls disabled; manual DELETE rejection surfaced | NOT RUN |
| S06 | Remove regular entry; confirmation explicitly says unregister only; fixture executable remains untouched, then re-add | NOT RUN |
| S07 | Launch selected runnable fixture; show exact selected/running transition and process/port returned by server | NOT RUN |
| S08 | Launch failure/HTTP 501 or port collision; display failure and do not fabricate running state | NOT RUN |
| S09 | Stop and rollback; confirmation and status follow server truth, errors surfaced | NOT RUN |
| S10 | Poll logs by cursor with timestamps; enforce bounded 150-event DOM; cursor reset handled | NOT RUN |
| S11 | Disconnect API or return malformed/5xx responses; buttons disable, error visible, recovery refreshes state | NOT RUN |
| S12 | Inject HTML-like fixture text in names, paths and log message; it displays as inert text | NOT RUN |
| S13 | Keyboard arrow navigation, tab focus, labels, narrow-screen layout, empty catalog | NOT RUN |
| S14 | Windows owner test with actual #15 backend, build/start/launch/relaunch/collision/rollback and no profile loss | NOT RUN |

## Current evidence / limits
All browser and Windows tests: **NOT RUN** until actually executed. GitHub source read-back alone is **STATIC ONLY**, not build/runtime validation. Backend issue #15 currently returns HTTP 501 for launch, stop and rollback; therefore a runnable Alpha is not established by this UI. Preserve existing Dashboard binaries, data, browser profiles and protected branches. Never merge or promote without owner approval.
