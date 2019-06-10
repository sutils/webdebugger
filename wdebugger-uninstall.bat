@echo off
cd /d %~dp0
nssm stop "wdebugger"
nssm remove "wdebugger" confirm
pause