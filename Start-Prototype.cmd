@echo off
setlocal
cd /d "%~dp0"
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0Start-Prototype.ps1"
if errorlevel 1 (
    echo.
    echo Prototype did not start. Check the error above.
    pause
)
endlocal
