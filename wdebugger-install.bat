@echo off
cd /d %~dp0
mkdir logs
nssm install "wdebugger" %CD%\wdebugger.exe -c -f %CD%\wdebugger.json
nssm set "wdebugger" AppStdout %CD%\logs\out.log
nssm set "wdebugger" AppStderr %CD%\logs\err.log
nssm start "wdebugger"
pause