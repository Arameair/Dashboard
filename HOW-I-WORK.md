# How I work

This is the controlled way to build, test, and automate prototypes with Dashboard.
Keep it simple. The goal is a steady flow of real results, not more machinery.

## The prototype loop

1. **One folder per prototype** under `projects/`, with a `project.json` that defines the buttons.
2. **Every prototype proves itself in Dashboard.** It counts as working only when you click a button here and the live log shows it. A team saying it passed does not count.
3. **Automation is just another button.** If something should repeat (for example every 10 minutes), that is an action the *program* handles. Dashboard starts it and shows the log; it does not schedule jobs itself (see Gaps below).
4. **Keep the original beside the new version** so you can always compare against something you know works (like Original and Prototype 1).
5. **Dashboard itself only changes when a real prototype needs something it cannot do yet.** One small change per branch, one pull request, only Dustin merges. Never push to `main`.

## Buttons that matter

For each prototype, aim for these actions in `project.json`:

- **Smoke test** — fast, offline or dry check. Closes by itself when it can.
- **Run** — the real demo or main path (mark `primary: true`).
- **Stop / Finish** — these are built into Dashboard, not `project.json`:
  - **Stop** force-closes the program (and its browser on Windows).
  - **Finish** presses Enter when the log shows the program is waiting ("Press Enter").

Optional: Install engine, Sign in, Test run (sends nothing), and other args your program understands.

## Rules that keep control

- Everything stays local on `127.0.0.1`. Dashboard never binds to `0.0.0.0`.
- Dashboard never commits, pushes, resets, or checks out git for you.
- One window: start, watch the log, test, stop or finish. You always know which version is which because each has its own folder and sidebar entry.
- Done looks like this: you can start, watch, test, and stop anything you are building from Dashboard, and you never wonder which version is the real one.

## How to add a new prototype

1. Copy `docs/project.template.json` into a new folder:
   `projects/<short-name>/project.json`
   Use a lowercase slug for the folder name (that becomes the project id).
2. Edit `name`, `group`, `about`, `exe`, `order`, and the `actions` list.
   Point `exe` at the program (relative to the project folder, or absolute).
   Use `dir` only if the working folder is not the project folder itself.
3. Put the program next to `project.json` (or where `exe` / `dir` say).
4. In Dashboard, click **Rescan** (or restart). The new project appears in the sidebar.
5. Click **Smoke test**. Then **Run**. Watch the log. Use **Finish** if it waits, **Stop** if you need to force close.
6. Keep an "Original" (or baseline) project in the same group until you are sure the new one is better.

Do **not** put a template under `projects/` with a real `project.json` — Rescan loads every subdirectory that has one, including names that start with `_`.

## Toward self-updating

Safe path, in order:

1. **Today:** prototypes write logs (and screenshots) that Dashboard already shows. That is the feedback loop.
2. **Next:** a button like "Check for updates" that a program can use to *download a reviewed release* into a staging folder and report what changed in the log.
3. **Never:** auto-merge pull requests, auto-push to `main`, or rewrite git history from Dashboard.

Self-updating means the *results* and *releases* you already approved show up here. You stay the one who merges.

## Gaps (what Dashboard does not do yet)

These are real limits in `source/main.go` today — do not invent fields for them:

- **No scheduling.** There is no cron, interval, or "every N minutes" field in `project.json`. "Run every 10 min" works only because the *program* loops; Dashboard just starts that action once.
- **No Stop/Finish actions in JSON.** Stop and Finish are UI controls on every running project.
- **No git actions.** By design.

When a prototype truly needs one of these from Dashboard itself, that is a small Dashboard change: one branch, one PR, Dustin merges.

## project.json fields (real schema)

| Field | Required | Meaning |
|-------|----------|--------|
| `name` | yes* | Display name (*folder name used if empty) |
| `group` | no | Sidebar group heading |
| `about` | no | Short description |
| `exe` | yes | Program path (relative to project folder or work dir, or absolute) |
| `dir` | no | Working folder (absolute, or relative to the project folder) |
| `screenshot` | no | Path relative to the working folder for the shot viewer |
| `order` | no | Sort order in the sidebar (lower first) |
| `actions` | yes | List of buttons |

Each action:

| Field | Required | Meaning |
|-------|----------|--------|
| `label` | yes | Button text |
| `args` | yes | String args passed to `exe` (use `[]` for none) |
| `primary` | no | Highlighted main button |
| `note` | no | Hint under the button |

Project id is always the **folder name**, not a field in the JSON.

See `docs/project.template.json` for a copyable starting point.
