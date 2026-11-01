@echo off
rem Switch console code page to UTF-8.
chcp 65001 >nul 2>&1
rem ARTEX supervisor launch script (Windows)
rem
rem Usage:
rem   start.bat                  Run in foreground (press Ctrl-C to stop)
rem   start.bat -addr :9000      Pass extra arguments directly to artex
rem
rem Purpose: launches artex.exe and decides whether to restart based on exit code:
rem   0      Clean shutdown by user   -> Exit loop
rem   75     Restart requested        -> Restart immediately (after selfupdate or rollback)
rem   other  Crash or abnormal exit  -> Backoff and restart (1->2->4... up to 60s)
rem
rem Note: Download, SHA256 validation, and binary replacement are handled inside Go
rem (selfupdate package) on startup. See start.sh for details.

setlocal enabledelayedexpansion
cd /d "%~dp0"

set "BIN=artex.exe"
if not exist "%BIN%" (
	echo [artex] binary not found or not executable: %BIN% 1>&2
	exit /b 1
)

set "RESTART_CODE=75"
set "MAX_DELAY=60"
set /a delay=1

:loop
"%BIN%" %*
set "code=!ERRORLEVEL!"

if "!code!"=="0" (
	echo [artex] clean shutdown
	exit /b 0
)

if "!code!"=="%RESTART_CODE%" (
	rem Update or rollback staged: restarting artex applies the change on boot.
	echo [artex] restart requested (applying new version)...
	set /a delay=1
	goto loop
)

echo [artex] abnormal exit ^(code=!code!^), restarting in !delay!s 1>&2
rem timeout fails in redirected console, use ping fallback.
set /a pings=!delay!+1
ping -n !pings! 127.0.0.1 >nul 2>&1
set /a delay=!delay!*2
if !delay! gtr %MAX_DELAY% set /a delay=%MAX_DELAY%
goto loop
