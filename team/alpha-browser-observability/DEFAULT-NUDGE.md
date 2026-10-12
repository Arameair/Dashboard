# Default Nudge and Traffic Lights — Dashboard Alpha

**Project ID:** `DASH-A4-PRJ-20261012`  
**Owner-facing reporter:** `DASH-A4-01-INSTANT-20261012` (01 Instant).  
**Work planner/dispatcher:** `DASH-A4-02-HIGH-20261012` (02 High).  
**Recipients:** 02 High, 03 Medium and 04 Medium **only when 01 verifies that a nudge is useful**.

## One default message (copy and paste into 01 to obtain lights)

> Check your company ID card, project charter, current GitHub issues/PRs and available ACKs. Return **only the nudge traffic lights** for **02 High, 03 Medium, and 04 Medium**: 🟢 NUDGE NOW, 🔴 WAIT/NO NUDGE, or ⚪ UNKNOWN. Include who to nudge, the exact default copy-and-paste message for each green seat, one next action and any critical blocker. Never guess green. Do not change code in this status turn.

## 01 Instant MUST respond in this concise layout

```text
DASHBOARD ALPHA — NUDGE LIGHTS
02 HIGH     🟢 NUDGE — #<issue>: <specific actionable work>
03 MEDIUM   🔴 WAIT — <reason/dependency>
04 MEDIUM   ⚪ UNKNOWN — <what is unverified>

SEND TO: 02 HIGH
COPY/PASTE NUDGE:
Check your Dashboard Alpha ID card and current GitHub assignments.
Continue the highest-priority authorized READY action. Coordinate with
the team through GitHub; post actual progress, commits/tests, or a precise
blocker. Return the next action and any seat requiring another nudge.

NEXT: <single next thing that will unblock the project>
```

The statuses above are **illustrative only**, not live findings.

## Traffic lights mean SHOULD OWNER NUDGE NOW?

| Light | Interpretation | Required evidence |
| --- | --- | --- |
| 🟢 **NUDGE NOW** | An introduced, ACKed seat has an authorized READY action, an incomplete deliverable needs a turn, or a verified stalled item needs attention | Current issue/PR/ACK, specific next action and no unresolved stop condition |
| 🔴 **WAIT / NO NUDGE** | There is no assigned work, work is complete, a dependency blocks it, or the seat is paused | Concrete reason and who can clear the dependency |
| ⚪ **UNKNOWN** | Cannot verify role ACK, issue, latest work, or readiness | Missing evidence; 01 asks 02 to reconcile rather than guessing |

## Recipient message for green seats

Use the **same** standard nudge for 02, 03 and 04 (03/04 must stay within issue's exclusive file paths):

> Check your Dashboard Alpha ID card and current GitHub assignments. Continue the highest-priority **authorized READY action**. Coordinate with the team through GitHub; post actual progress, commits/tests, or a precise blocker. Return the next action and any seat requiring another nudge. Do not modify `main`, `stable`, or `alpha`, or exceed the issue's allowed paths.

The member-specific issue or next action can precede the standard message. Do not mark **message sent** unless the owner or a verified delivery mechanism actually sent it. If no seat is green, output `SEND TO: NONE`, `COPY/PASTE NUDGE: NOT NEEDED`.

## Continuous coordination rule for 02

On each owner-triggered turn, 02 High should identify and perform a permitted next *planning, documentation, issue-dispatch, integration-review or setup* action, if one exists. Re-check prior work, use GitHub receipts to unblock the next dependency and make clear who is READY. If everything is blocked, report **blocked** and the specific owner input required; do not manufacture tasks or pretend to work asynchronously between ChatGPT messages.

**No autonomous cross-chat messaging.** GitHub is the shared written record; the owner pastes nudges to the designated conversations. This file defines team protocol, not an enabled scheduler or browser action.
