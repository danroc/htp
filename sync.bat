CD /D "%~dp0"
PowerShell -Command "Set-Date -Adjust $([TimeSpan]::FromSeconds($(.\htp.exe -v -n 12)))"
PAUSE