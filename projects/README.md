# Projects

A project is one browser the dashboard runs: its own window, its own profile
(sign-in), a start page and a nudge message.

Projects can be added from the dashboard page (**+ Project**) or listed in a
definition file. The dashboard reads every `*.json` here, and the launcher also
reads `..\Teams\dashboard\` when the Teams repo is cloned next to this one.

```json
{
  "projects": [
    {
      "id": "example-01",
      "name": "Example 01",
      "url": "https://chatgpt.com/",
      "color": "#3b82f6",
      "nudge": "Text that Nudge now and auto-nudge send.",
      "nudge_minutes": 10,
      "auto_nudge": false,
      "save_replies": true,
      "console": false,
      "auto_start": false
    }
  ]
}
```

A definition only adds projects the dashboard doesn't have yet, matched by
`id`. Once a project exists, changes you make in the dashboard are kept and the
file is not applied again. The `id` also names the profile folder, so keep it
fixed. Files named `*.local.json` are ignored by Git.

To read definitions from somewhere else:

```powershell
./Start-Dashboard.ps1 -Projects 'C:\path\to\folder;C:\path\to\file.json'
```
