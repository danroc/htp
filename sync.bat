@echo off

if not "%1"=="am_admin" (
    powershell Start -Verb RunAs '%0' am_admin & exit /b
)

cd /d "%~dp0"
powershell -Command "Set-Date -Adjust $([TimeSpan]::FromSeconds($(.\htp.exe -v)))"
pause