@echo off
setlocal
cd /d "%~dp0"
powershell.exe -NoProfile -File "%~dp0scripts\Start-Alpha-Owner-Smoke.ps1"
set "result=%ERRORLEVEL%"
if not "%result%"=="0" (
  echo.
  echo Alpha did not start. Existing Dashboard installation and profiles were not intentionally modified.
  echo Record this error and send it back for diagnosis.
  pause
)
exit /b %result%
