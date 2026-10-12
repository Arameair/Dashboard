@echo off
setlocal
cd /d "%~dp0"
set "GOCACHE=%~dp0runtime\go-cache"
if not exist "runtime\bin" mkdir "runtime\bin"
go build -o "runtime\bin\Dashboard.exe" ./dashboard
if errorlevel 1 (
    echo Dashboard build failed. Check that Go is installed and the new dashboard is closed.
    pause
    exit /b 1
)
"runtime\bin\Dashboard.exe" %*
