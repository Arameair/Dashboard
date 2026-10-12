# Synthetic, interactive workspace overview — issue #25

**SIMULATED WORKSPACES — NO LIVE PROJECT CONNECTION.** Two independently mounted sandboxed iframes are *instances* of one synthetic telemetry project, each with own editable input, action buttons, pause/disconnect, counter, bounded event log (120 entries). Open `index.html` from a disposable folder in a normal independent browser; if local file iframe policy blocks the demo, serve this directory only from a temporary loopback local static server, never user profiles or production sessions. No packages, network fetch, process execution or real project adapter.

Two mounted iframe nodes persist while the parent grid changes CSS to focus/back. Instance B remains mounted/logging when instance A is focused. Tabs are per instance, not project-wide; the Updates panel is disabled and does not claim update/rollback support.

## Provider contract for future integration (NOT IMPLEMENTED)
`WorkspaceProvider{projectId,instanceId,type,embeddedLocalUrl?,externalLaunchRef?,interactive,health,logCursor}`. **embedded-local:** allow only explicitly operator-approved local origins with safe framing and valid process ownership; **preview-only:** labeled noninteractive snapshots; **external-window:** NOT EMBEDDABLE / OPEN EXTERNALLY where iframe policies prevent safe embedding. Never infer control of arbitrary Chrome/Edge/ChatGPT windows or authenticated sessions. Logs are instance-scoped, bounded, and inserted by textContent; input remains inside sandbox frames.

## Acceptance ledger
- T01 two instance interfaces visible, editable and independently logged — NOT RUN (browser).
- T02 focus A/back preserves A input, counters and log; B continues logging — NOT RUN.
- T03 focus B/back repeatedly; keyboard and narrow screen — NOT RUN.
- T04 disconnect A does not disconnect B — NOT RUN.
- T05 external-window fallback truthful; Updates disabled — NOT RUN.
- T06 no source install/profile/real-runtime touch — design review only, NOT runtime verified.

Branch: `experiment/a4-version-manager-v5-20261012`. See #25 GitHub comment for exact final commit SHA. No Windows or browser automation tests have been executed.
