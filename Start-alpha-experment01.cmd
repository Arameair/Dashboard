@echo off
setlocal
cd /d "%~dp0"
set "GOTOOLCHAIN=local"
set "GOPROXY=off"
where go >nul 2>&1 || (echo Go is missing from PATH. & pause & exit /b 1)
echo Open http://127.0.0.1:18999/
start "" "http://127.0.0.1:18999/"
go run .\experiments\alpha-experment01\main.go .\experiments\alpha-experment01\page.go
if errorlevel 1 (echo Prototype startup failed. & pause)
