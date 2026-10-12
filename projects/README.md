# Projects

Each project owns its `project.json`, `source/`, and local executable:

- `round2-original/`: the original Round 2 Playwright demo
- `round2-p1/`: the clean-shutdown prototype
- `teams-nudger/`: the Teams chat nudger

Projects and the dashboard share this Git repository for now. The dashboard
application source lives in `../dashboard/`; its generated run logs live in
`../runtime/logs/<project-id>/last-run.log`.

Build a project from its source folder, for example:

```powershell
cd round2-p1/source
go build -o ../round2-p1.exe .
```

`managed/`, `runs/`, and `artifacts/` contain the shared browser engine, saved
browser profiles, and screenshots. These folders, executables, and test logs are
ignored. Keep saved sign-in profiles out of Git.

The dashboard also accepts `-projects <folder>` for a future move to a separate
workspace; moving the projects out is deferred.
