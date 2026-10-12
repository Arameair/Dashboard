# Dashboard

Dashboard is a small Go app that serves a local web UI for switching between projects, starting actions, watching logs, and stopping or finishing runs.

## Included in this repo

- `source/main.go` — the dashboard server and project runner
- `source/page.go` — the embedded HTML/JS UI
- `projects/*/project.json` — example project definitions and their actions
- `START-HERE.txt` — quick usage notes for the packaged desktop version

## Run locally

1. Open a terminal in this folder.
2. Run `go run ./source`.
3. The app opens the browser and uses the local `projects/` folder as its configuration directory.

## Notes

The repo intentionally ignores large runtime artifacts such as browser engines, browser profiles, and generated logs so the source stays lightweight and easy to push to GitHub.
