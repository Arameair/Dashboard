# Dashboard Alpha — Four-Seat Team Directory

**Company-style project ID:** `DASH-A4-PRJ-20261012`  
**Project:** [Arameair/Dashboard](https://github.com/Arameair/Dashboard)  
**Target:** Four live browser sessions, reliable reconnection, health diagnostics, side-by-side comparison. **Team onboarding comes first.**  
**Project charter:** [PROJECT.md](PROJECT.md).

## Team roster

| Seat | User-selected ChatGPT configuration | Role | ID card |
| --- | --- | --- | --- |
| 01 | Instant | Urgent Execution: what needs action **now** | [01-INSTANT.md](01-INSTANT.md) |
| 02 | High thinking | Project Planner: **everything** that needs doing, dependencies and delegation | [02-HIGH.md](02-HIGH.md) |
| 03 | Medium thinking | Specialist A: scoped engineering and tests | [03-MEDIUM.md](03-MEDIUM.md) |
| 04 | Medium thinking | Specialist B: scoped engineering and UI/log review | [04-MEDIUM.md](04-MEDIUM.md) |

**Owner and primary assistant** are outside these four roles: owner approves scope and promotions; primary assistant helps owner coordinate and can take explicitly requested tasks. Models are **manually selected in each ChatGPT chat**, not configured by GitHub or Dashboard.

## Who talks to whom?

Owner → primary assistant → **02 High** for project-wide intake/planning; **01 Instant** for urgent immediate actions. 02 breaks work into separately owned GitHub issues; 03/04 execute after owner-led chat handoff and role ACK. 01 may take a narrowly scoped urgent issue without waiting for broad planning, provided it does not collide with 03/04 files and never changes protected branches. 02 returns the consolidated status, 01 flags blockers that need owner attention.

## Enrollment order

1. Owner opens/designates **four separate ChatGPT conversations** and selects Instant / High / Medium / Medium. A GitHub file does not create a ChatGPT conversation.
2. Paste each card's **Introduction message** into its corresponding chat, preferably with the exact GitHub card link.
3. Each conversation reads its card, identifies its persistent ID, and ACKs. Acknowledge no unverified seat.
4. 02 confirms the roster and prepares a work breakdown. 01 identifies the immediate next action. 03/04 initially WAIT for an authorized READY issue. Owner authorizes actual development.
5. For actual work, use [Dashboard Alpha issue #11](https://github.com/Arameair/Dashboard/issues/11) as **the backlog input, not a pre-activated assignment**. Branches/files/acceptance tests must be settled before specialist writes.

## Existing work protection

This new **Alpha test team** is distinct from older Dashboard role cards/issue assignments (#3, #4, #6, #7, #8, #9). Never silently claim or supersede their work. The existing Dashboard v5 main remains stable while the new team is enrolled. The preexisting [PR #10](https://github.com/Arameair/Dashboard/pull/10) and `feature/embedded-view` remain separately owned and unmerged.
