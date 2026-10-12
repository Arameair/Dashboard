@echo off
setlocal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0Start-Alpha-Version-Manager.ps1" %*
exit /b %ERRORLEVEL%
