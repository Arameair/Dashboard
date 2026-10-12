# Dashboard

One local page that controls several real browsers. Each project is its own
Chrome or Edge window with its own profile, so each stays signed in separately.
From the page you start and stop them, see what each is showing, navigate, send
a message, read the latest reply, auto-nudge on a timer and read each log.

Everything stays on this computer: the page is served on 127.0.0.1 only.

## Run

**Branch integration:** the Dashboard application lives on `Dashboard/main`, while the four browser definitions live in `Teams/main/dashboard/teams.json`. The Teams repository currently defaults to `stable`, which does **not** include that file. A default `git clone` of Teams will therefore start the Dashboard with no imported Team 01–04 cards. Do not overwrite or switch an active Teams worktree to fix this.

For a clean, independent Windows checkout, use a new parent directory:

```powershell
mkdir C:\Dev\TeamsDashboard-Split
cd C:\Dev\TeamsDashboard-Split
git clone --branch main https://github.com/Arameair/Dashboard.git Dashboard
git clone --branch main https://github.com/Arameair/Teams.git Teams
cd Dashboard
.\Start-Dashboard.cmd
```

Keep both clones side by side. Dashboard runs without Teams; Team definitions are imported only if the adjacent Teams checkout provides them. An existing `%LOCALAPPDATA%\TeamsDashboard\v5\config.json` takes precedence for IDs already in it, preserving browser profiles and local edits. In particular, importing `teams.json` does not rename existing `Seat 01` cards to `Team 01` automatically.

Double-click `Start-Dashboard.cmd` (or run `./Start-Dashboard.ps1`). It builds
the dashboard into `runtime/bin/`, opens the page at http://127.0.0.1:8777/ and
runs until the window is closed. Go 1.24 or newer is required; every dependency
is in `dashboard/vendor`, so nothing is downloaded.

Closing the dashboard leaves the browsers open. The next start reconnects.

## Projects

The dashboard itself starts empty. Projects come from:

- **+ Project** on the page, and
- definition files: `projects/*.json` in this repo and, when the
  [Teams](https://github.com/Arameair/Teams) repo is cloned next to this one,
  `../Teams/dashboard/*.json`.

So the usual layout is:

```text
<folder>/Dashboard    this repo: the app
<folder>/Teams        the Teams project: its browsers are defined in dashboard/teams.json
```

See [projects/README.md](projects/README.md) for the file format. Future
projects work the same way: add a definition file, or pass `-projects`.

## Layout

```text
dashboard/             Go source and the page (index.html); vendor/ holds dependencies
projects/              project definition files
runtime/               builds and Go's cache (ignored by Git)
Start-Dashboard.cmd    build and run
Start-Dashboard.ps1    same, with -Projects and -NoOpen
START-HERE.txt         how to use the page
```

Settings, browser profiles, logs and saved replies are kept outside the repo in
`%LOCALAPPDATA%\TeamsDashboard\v5\` (Settings > Open data folder).

## Branches

- `main`: this version.
- `stable`: the earlier project launcher that `main` replaced.
