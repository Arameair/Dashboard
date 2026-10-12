# Project Identity Card — Four-Browser Dashboard Alpha

| Field | Value |
| --- | --- |
| Project ID | `DASH-A4-PRJ-20261012` |
| Company/team | Dashboard Alpha Observability Team |
| Repository | `Arameair/Dashboard` |
| Working baseline | Dashboard v5 `main` at `e4bfa2636acdbe1b1bbf7b2c1a988ad62ca8c59c` at enrollment; verify current SHA before working |
| Owner | Arameair |
| Company communication | Owner ↔ primary assistant ↔ 02 High (planning) / 01 Instant (urgent) |
| Initial state | **FORMING**; roles are defined, not autonomously activated |
| Parent request | [Four-browser Alpha reliability work order #11](https://github.com/Arameair/Dashboard/issues/11) |
| Release rule | Separate Alpha branch and draft PR; no changes to `main`, `stable`, or `alpha` without owner GO |

## Purpose

Get **four user-designated ChatGPT browser sessions** visible in the Dashboard and measure **which session is connecting/recovering/responding best**, using evidence-based, read-only health logs. Reuse existing **Dashboard-managed** Chrome/Edge browser sessions when their persistent profiles and control ports are reachable. Do **not** promise takeover of arbitrary already-open browsers, multiple independent browser engines, or autonomous AI work.

## Acceptance categories (future Alpha, NOT YET TESTED)

- 4 seats appear independently; model identities manually chosen in the respective ChatGPT chats.
- Dashboard restart reconnects to four already-running **Dashboard-managed** browser sessions without duplicating them or touching user sign-ins.
- Differentiate `RECONNECTED / NEW / NOT CONTROLLABLE`; stale port and failed connection show meaningful errors.
- Per-seat log: timestamps, poll success/failure, latency, last good response, input presence, page readiness, errors/recoveries; no secrets, tokens, conversation text, or full private URLs.
- Side-by-side report labels observation windows and PASS/FAIL/NOT RUN. Running process ≠ healthy ChatGPT ≠ useful project progress.
- Windows run instructions, specific build SHA, tests and operator-observed evidence recorded honestly.
- No messages, nudges, sign-ins, private-profile access, or production merges in this narrow Alpha experiment.

## Responsibility and escalation

**01 Instant — urgency:** pick the most time-critical actionable item, verify blockers, report result fast; route larger planning to 02.  
**02 High — full picture:** map all work, risks, dependencies, priorities, acceptance criteria, path exclusivity and owner questions; dispatch only after explicit scope authorization.  
**03 Medium — specialist A:** implement assigned browser connectivity/reconnect/diagnostics, with tests on own paths.  
**04 Medium — specialist B:** implement assigned UI/logging/summary presentation and independent tests on own paths.  
**Owner + primary assistant:** can work on tasks separately by agreement; 02 must record who owns each write path.

## Governance

All seat definitions and scope changes are tracked in this repository. Cards are GitHub-backed **identity/contracts only**; they cannot change a chat's underlying model, run unattended, or transmit messages across chats. Owner manually introduces all four chats. 02 must never assign code changes with overlapping write paths. Roles must ACK before receiving work, and no green traffic light without evidence.

## Operating signal

**01 Instant is the traffic-light interface; 02 High is the continuous coordinator/planner.** Each invocation of 01 returns 🟢 nudge now / 🔴 wait / ⚪ unknown for 02, 03 and 04, a recipient list, the standard [default nudge](DEFAULT-NUDGE.md), and one next action. Green is **whether owner should nudge**, not an assertion that the project works. 02 should actively create/coordinate/clear authorized work on each turn, but cannot operate unprompted in another ChatGPT conversation. 02 supplies 01 with verified GitHub evidence rather than guesses. Enrollment and write authorization remain separate gates.

## First action

**Only enrollment.** Read the four ID cards, introduce members, obtain ACKs. Then 01 identifies what needs action immediately, while 02 prepares the backlog from #11. Specialists WAIT until READY issues are dispatched.
