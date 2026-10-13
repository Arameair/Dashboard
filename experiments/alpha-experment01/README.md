# alpha-experment01

Standalone GitHub-project gallery based on the isolated Dashboard Alpha source branch. Runs independently on port 18999.

Windows: clone branch alpha-experment01 into a new folder (never overwrite existing Dashboard Alpha), then double-click Start-alpha-experment01.cmd. Requires Go and Git. No PowerShell policy changes.

Dashboard card: checks the existing Dashboard Alpha on port 18777, reads real sample preview and logs, and displays the local repository branch/commit. Teams card: independent GitHub repo card; runtime not connected.

No automatic launch, stop, nudge, repository modification, or browser-profile access. Existing Dashboard remains unchanged.

Commands in PowerShell:
    cd C:\Dev\Projects
    git clone --branch alpha-experment01 --single-branch https://github.com/Arameair/Dashboard.git Dashboard-alpha-experment01
    cd .\Dashboard-alpha-experment01
    .\Start-alpha-experment01.cmd
