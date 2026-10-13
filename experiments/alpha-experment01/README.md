# alpha-experment01 — runnable project hub

This experiment adds a **real, owner-triggered Alpha process runner** to the Netflix-style project cards. It remains on a dedicated experiment branch, without altering the normal Dashboard Alpha.

## Start on Windows

Prerequisites: Git and Go 1.24+ already installed. ProjectHub's Go stdlib-only server is local and does not install packages.

Use **separate local clones** (do not replace existing installs or switch an in-use checkout):

```powershell
Set-Location C:\Dev\Projects
# Only when these folders do not already exist:
git clone --branch experiment/alpha-v5-owner-smoke-20261013 --single-branch https://github.com/Arameair/Dashboard.git Dashboard-Alpha-Owner-Test
git clone --branch alpha --single-branch https://github.com/Arameair/Teams.git Teams-Alpha

# Clone this prototype once, or git pull --ff-only if already cloned:
git clone --branch alpha-experment01 --single-branch https://github.com/Arameair/Dashboard.git Dashboard-alpha-experment01
Set-Location .\Dashboard-alpha-experment01
.\Start-alpha-experment01.cmd
```

Open http://127.0.0.1:18999/. For an existing clean local prototype clone: `git pull --ff-only` before launching. **Do not run two ProjectHub processes on port 18999.**

## Actual managed Alpha instances

| Project | Required checkout | Branch | Isolated test URL |
| --- | --- | --- | --- |
| Dashboard | `C:\Dev\Projects\Dashboard-Alpha-Owner-Test` | `experiment/alpha-v5-owner-smoke-20261013` | http://127.0.0.1:18877/ |
| Teams | `C:\Dev\Projects\Teams-Alpha` | `alpha` | http://127.0.0.1:18998/ |

Click project → **Start Alpha**. The backend verifies the local Git branch and repository origin, builds the selected Go source into the ProjectHub experiment's own ignored `runtime/alpha-experment01/bin` folder, and launches a new process with isolated `runtime/alpha-experment01/data/<project>` state. No automatic Git checkout/pull or module download; `GOPROXY=off` and `GOTOOLCHAIN=local` are enforced for builds. On Teams, missing cached Go dependencies cause a visible build failure. The first full Teams browser-run may still require network and Playwright provisioning after you explicitly launch it.

When the owned server responds, its **actual web UI** is embedded in the expanded workspace. Framing might be blocked by browser headers; **Open full screen** opens the actual localhost app regardless. The preview is a live web application, not a screenshot or simulated project render. The card shows process state and PID.

**Stop Alpha** invokes Windows `taskkill /PID <owned child PID> /T /F` for only the process started by this ProjectHub. It never attaches to or kills a pre-existing instance merely because a port responds. Force stop can leave app-created browser data/temporary files; prefer graceful in-app Finish/Stop controls for any Teams browser run before stopping its Alpha server.

The existing owner Alpha on port 18777 is separate and remains untouched.

## Important limitations

- Status and version correspond to **local checked-out sources**, not the latest remote GitHub commit. Branch verification does not download code.
- There is no arbitrary-repo runner, version selector, rollback, or branch switching. Only the two explicitly configured repos/branches are supported.
- The hub does not automatically send ChatGPT messages or operate logged-in personal browsers.
- If the app is running outside ProjectHub, it will not be adopted, stopped, or represented as its managed process.
- ProjectHub closing does not intentionally stop child apps. Stop owned instances in the UI before closing it.
- Not yet verified on the owner's Windows machine; test Start, iframe/full-screen, and Stop one project at a time.
