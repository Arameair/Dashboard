@echo off
setlocal
cd /d "%~dp0"
set "GOCACHE=%~dp0runtime\go-cache"
if not exist "runtime\bin" mkdir "runtime\bin"
go build -C dashboard -o "..\runtime\bin\Dashboard.exe" .
if errorlevel 1 (
    echo Dashboard build failed. Check that Go is installed and the dashboard is closed.
    pause
    exit /b 1
)
rem Projects come from this repo's projects\ folder and, if it is cloned next to
rem this one, from the Teams repo. Pass -projects yourself to use other places.
"runtime\bin\Dashboard.exe" -projects "%~dp0projects;%~dp0..\Teams\dashboard" %*
