# Dashboard Alpha — isolated owner smoke test (2026-10-13)

**Status: CANDIDATE FOR OWNER WINDOWS SMOKE TEST — NOT WINDOWS-VERIFIED.**

This branch is based on Dashboard v5 `main` commit `e4bfa2636acdbe1b1bbf7b2c1a988ad62ca8c59c`. It is **not** the unfinished A4 Version Manager, **not** a promoted `alpha` or `stable` build, and **not** the future four-prototype gallery. Its purpose is to get one existing real Dashboard browser/project workspace running in a clean isolated copy. No code on `main` or `stable` was changed.

## At home — 3 steps

1. In PowerShell, from a safe parent folder (for example `C:\Dev\Projects`), make a NEW clone (do not clone over your existing Dashboard):

   ```powershell
   git clone --branch experiment/alpha-v5-owner-smoke-20261013 --single-branch https://github.com/Arameair/Dashboard.git Dashboard-Alpha-Owner-Test
   ```

2. Open `Dashboard-Alpha-Owner-Test` and double-click **`Start-Alpha-Owner-Smoke.cmd`**. Requires **locally installed Go 1.24 or newer**, Git, and Chrome or Edge for the optional browser sample. The script uses vendored modules only and refuses to download an SDK or packages. If Windows blocks it or any check fails, stop and report the error; do not work around security warnings.

3. Dashboard v5 should open at **http://127.0.0.1:18777/**. It should show an **Alpha browser smoke test** project. Check that it is STOPPED initially. Click **Start** on that one sample, confirm a separate Chrome/Edge window displays `example.com`, inspect the live preview and that project's log, then Stop it. No ChatGPT login, automated send, production project, or Teams integration is needed.

**Stopping:** Press Ctrl+C in the test launcher console after stopping the test browser. Use your original shortcut to return to your previously working Dashboard.

## Safety contract / scope

- Test code and builds live under the **new test clone**.
- Test catalog/config, browser profiles, replies and logs live under `Dashboard-Alpha-Owner-Test\runtime\alpha-owner-smoke\data`, **not** `%LOCALAPPDATA%\TeamsDashboard\v5`.
- Uses a dedicated local port **18777**. If occupied, the script stops rather than knowingly connecting to the existing application. Do not run simultaneous copies of this test launcher.
- Test project definitions are under `runtime\alpha-owner-smoke\project-definitions`; the new project is inert until manual Start. `auto_nudge=false`, `auto_start=false`, no default nudge text.
- Does not load `../Teams/dashboard` or an existing production project definition.
- No `git pull`, auto-update, branch promotion, protected-branch merge, signup, ChatGPT message send, or arbitrary worker task.
- Dashboard v5 may leave browser windows open after the server closes; **Stop the sample browser manually** before closing if you want a clean stop. Only the isolated test profile is associated with it.

## Evidence to collect

| Check | Current status | Owner PASS criteria |
| --- | --- | --- |
| Source branch, entry scripts | PASS — GitHub readback only | Candidate scripts exist at recorded ref |
| Go 1.24 vendor test / vet / build on Windows | NOT RUN | Launcher reports no errors |
| Start, exact port 18777 and page | NOT RUN | Correct browser page opens at the dedicated URL |
| Isolated data and profiles | NOT RUN | New data appears only in test runtime folder |
| Project loaded, default stopped / no auto nudge | NOT RUN | Sample listed and initially stopped |
| Start one browser, real live preview, logs | NOT RUN | Window, preview and log correspond to the sample |
| Stop / restart only the sample | NOT RUN | Test browser closes/reopens as expected |
| Existing Dashboard and sign-ins unaffected | NOT RUN | Existing install remains usable, data/profile untouched |
| Prototype comparison/promotion / worker automation | NOT IMPLEMENTED | Out of scope of this smoke test |

Record Windows version, `go version`, branch commit SHA (`git rev-parse HEAD`), which checks passed, and any errors. **Do not call the Alpha fully tested or safe for production until this Windows smoke test passes.** Browser and website behavior may vary with local network/browser policies.

## Follow-up

- A4 Version Manager issues [#14](https://github.com/Arameair/Dashboard/issues/14), [#15](https://github.com/Arameair/Dashboard/issues/15), [#16](https://github.com/Arameair/Dashboard/issues/16) still own safe selection/launch/rollback of multiple builds.
- This branch is a parallel **owner test copy of existing v5**, not a replacement for [draft PR #17](https://github.com/Arameair/Dashboard/pull/17).
- If this small smoke test works, the next task is to wire one actual disposable project instance into Dashboard and separately prove one Teams worker receipt, then connect the first complete conversation → prototype loop.
