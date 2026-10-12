# Universal Nudge + Compact Handoff — Dashboard Alpha

**Project:** `DASH-A4-PRJ-20261012` · **Repo:** https://github.com/Arameair/Dashboard  
**01 Instant:** owner-facing traffic-light reporter · **02 High:** work queue, dispatch and peer coordination · **03/04 Medium:** scoped execution.

## One universal default nudge (SAME text for every team member)

```text
Check your ID card and latest assignments in https://github.com/Arameair/Dashboard. Continue your highest-priority authorized READY work; coordinate directly in the relevant GitHub issue or PR. Post what changed (commit/test/link), what is blocked or needs help, and your single next action. Respect assigned files and protected branches.
```

Use this exact message for 01, 02, 03, or 04. No name, issue number or branch is needed in the message: **each seat reads its own card and its own latest GitHub assignments**. The owner pastes it only into the chat(s) identified as needing a nudge. GitHub is the cross-seat communication record; the nudge does not automatically reach other chats.

## 01 Instant's mandatory small handoff (at every status request)

```text
TEAM: <one short sentence of actual progress / top blocker>
01 INSTANT  🔴 WAIT / 🟢 NUDGE / ⚪ UNKNOWN — <few words>
02 HIGH     🟢 NUDGE / 🔴 WAIT / ⚪ UNKNOWN — <few words>
03 MEDIUM   🟢 NUDGE / 🔴 WAIT / ⚪ UNKNOWN — <few words>
04 MEDIUM   🟢 NUDGE / 🔴 WAIT / ⚪ UNKNOWN — <few words>
HELP: <who needs help + from whom, or NONE>
SEND TO: <green seat numbers, or NONE>
NEXT: <one smallest next step, with issue link>
DEFAULT NUDGE: <paste the exact universal text above, once>
```

Aim for **8 short lines plus the unchanged nudge**. No detailed project history, duplicate per-seat nudges, raw logs, speculative schedules or essay. List **all four seats**, even 01, so the owner sees a complete team snapshot. If no seat needs attention, say `SEND TO: NONE` and still show the universal template for future use.

## Light definitions — should the owner nudge?

- 🟢 **NUDGE NOW:** verified introduced/ACKed member has specific authorized work ready to advance, or a stalled deliverable needs a prompt. Say which GitHub issue.
- 🔴 **WAIT / NO NUDGE:** no ready assignment, blocked pending another person/system, deliberate pause, or work completed. State whose action can unblock.
- ⚪ **UNKNOWN:** no verified current ACK / issue / progress. Never guess green.
- **HELP:** separately flag assistance or decision needed, even when the blocked seat is 🔴. E.g. `HELP: 03 needs 02 to approve API contract`. Escalate real owner input only when necessary.

## Team responsibilities and channel

- **02 High:** at each invocation, inspect current issues/PRs and advance one authorized coordination/build-setup step if possible; give 01 an accurate queue with READY/BLOCKED/REVIEW, ownership, and dependencies. 02 communicates with 03/04 directly in GitHub; 01 does **not** become an owner-relay service.
- **03 and 04:** publish short work evidence, blockers and dependency requests in their assigned issues; read 02 instructions there.
- **01 Instant:** read verified current GitHub evidence and return only the compact handoff; if useful, post it to shared issue #14 so 02 can use it on its next invoked turn.
- **Owner/primary assistant:** sees who to nudge and can copy the same message to any green chat. No manual rewriting.

**Boundaries:** No automatic cross-chat wake-up, no invented green, no merges or protected branch writes, no running untrusted binaries or accessing personal browser profiles. **Priority:** runnable Windows Version Manager Alpha #14/#15/#16 first; multi-project safe reuse #18 follows.
