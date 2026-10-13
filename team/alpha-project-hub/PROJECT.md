# Dashboard Alpha Project Hub — experimental team charter

**Project:** DASH-HUB-20261013  
**Repository:** https://github.com/Arameair/Dashboard  
**Development branch:** `experiment/alpha-project-hub-team-20261013`  
**Baseline:** Dashboard `alpha` at `7ba3b485b73d646204643e26b1efd895a914a51c` (verified at branch creation).  
**Status:** NOT ACKNOWLEDGED until the owner creates/designates each separate chat and the seat acknowledges its exact card.

## Ground rules
- Work only on the experimental branch or explicitly authorized child branches; never change `alpha`, `stable`, or `main`, merge, or promote without owner approval.
- GitHub issues/PRs are the shared communication record. Different ChatGPT conversations do not spontaneously run or communicate; owner sends nudges as needed.
- Before writing code, verify the GitHub branch HEAD and a scoped READY issue from 02, with exclusive file ownership and acceptance criteria. For this project, the owner has authorized a prototype, **not** production promotion or automation of ChatGPT.
- Marquee ZIP is **visual inspiration only**; do not import its code wholesale or assume it is safe. Preserve and extend Dashboard Alpha's working application.
- All app launch/stop operations require explicit user action, isolated builds and data, exact process ownership, no dangerous git clean/reset, no arbitrary user-supplied shell execution, no production profile access.
- Always report PASS/FAIL/NOT RUN based on actual evidence. Neither CI nor source inspection equals owner-Windows acceptance.

## Owner goal
Keep Dashboard's working Alpha as the source. Build an experiment with a Netflix-style repository gallery for **Dashboard** and **Teams**, featuring **real running Alpha demos**, not fake cards or GitHub links. Owners click a card to expand a running workspace and logs; each project has explicit Start and Stop for its own isolated managed instance. Marquee is inspiration only; do not overwrite Alpha with Marquee.

## Scope and acceptance
1. Exact Alpha parent SHA and current source identity are recorded; warn if Alpha advances and rebase/merge only after review.
2. Dashboard and Teams exist as distinct project entries, with verified repo, branch/commit and local runtime info.
3. Start actual approved Alpha binaries in isolated environments; show real app live view or true live screenshot with honest offline/error indicators. Stop only owned processes. Disallow self-destructive stops, port hijacking, untrusted command execution, browser-profile reuse, arbitrary code installs, and automatic Git updates.
4. Project logs stay separated; switching cards must not change application lifecycle.
5. Working Windows runner instructions with an observed test matrix: build, start, live view, logs, stop, restart, profile isolation; tag unrun checks NOT RUN.
6. Owner reviews experiment and explicitly approves any Alpha promotion; no automatic branch promotion.

## Work sequence
01 maintains evidence-backed red/green/unknown nudge signal. 02 produces issue-scoped plan/API contract, assigns 03 safe runner and 04 frontend. 03 and 04 coordinate through GitHub after ACK. Owner designates the separate ChatGPT chats and supplies nudges. No claimed background worker.

## Default nudge (same for all four)
Check your ID card and latest assignments in https://github.com/Arameair/Dashboard on branch `experiment/alpha-project-hub-team-20261013`. Continue only your highest-priority authorized READY work. Coordinate through GitHub issues/PRs and report actual commit/test evidence, blockers, and one next action. Respect exclusive file ownership and never change protected branches.

## Current state
Branch and identity cards only. **No team seat is ACKed yet; no live-runner implementation or Windows acceptance is claimed for this branch.**
