# Dashboard Alpha synthetic UI fixture — seat 04 / issue #24

**MOCK ONLY.** Never a real Dashboard launch, never valid operator approval, never a source of Windows acceptance. The fixture binds only 127.0.0.1 and serves only /, /index.html, /app.js, /styles.css from the existing #16 static directory. No GitHub fetch, package installation, browser automation, shell command execution, sign-in, user profile access, persistent storage, or real binary startup.

## Run in a disposable shell

```sh
node --version
node --test tools/version-manager/ui-fixture/fixture-tests.mjs
node tools/version-manager/ui-fixture/server.mjs
```

Open the printed `http://127.0.0.1:<ephemeral-port>/` URL in an **independent test browser**; do not use signed-in Dashboard sessions. Server process starts only when invoked. Stop with Ctrl+C. Optional `ALPHA_FIXTURE_PORT=18994` selects a disposable loopback port.

### Manual fault injection
```sh
curl -X POST http://127.0.0.1:18994/__fixture/fault -d '{"failLog":true}'
curl -X POST http://127.0.0.1:18994/__fixture/fault -d '{"failLog":false}'
curl -X POST http://127.0.0.1:18994/__fixture/reset
```
Also supports `failStatus`, `rejectAdd`, and `disconnected` booleans. Reset is ephemeral. All registered versions are synthetic. Launch/Stop/Rollback always return 501; approval-preview 422; approvals 403. Fixture logs are synthetic. Do not interpret selected/running labels as real PID ownership.

## Acceptance matrix (fill with executable evidence, not assumed PASS)
S01 search name/channel/branch/SHA: NOT RUN (browser). S02 selected != running: NOT RUN (browser). S03 Add valid: NOT RUN (browser). S04 reject invalid: NOT RUN (browser). S05 protected/active delete: NOT RUN (browser). S06 unregister/re-add: NOT RUN (browser). S07 Launch 501: NOT RUN (browser). S08 failure not success: NOT RUN (browser). S09 rollback 501: NOT RUN (browser). S10 >500 cursor rollover: NOT RUN (browser). S11 failLog/restore/reconnect controls: NOT RUN (browser). S12 inert untrusted HTML: NOT RUN (browser). S13 keyboard, responsive, empty-state: NOT RUN (browser). S14 real Windows owner launch/rollback: NOT RUN (blocked #15).

The Node unit suite can PASS independently; it is **not** an interactive browser test nor a Go/Windows verification. Always record OS, Node version, commit SHA, and commands executed alongside outcomes.
