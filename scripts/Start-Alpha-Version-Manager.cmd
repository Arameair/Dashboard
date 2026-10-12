@echo off
setlocal
powershell.exe -NoProfile -File "%~dp0Start-Alpha-Version-Manager.ps1" %*
exit /b %ERRORLEVEL%
