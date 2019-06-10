@echo off
set srv_name=wdebugger
set srv_ver=0.1.0
set OS=%1
del /s /a /q build\%srv_name%
mkdir build
mkdir build\%srv_name%
set GOOS=windows
set GOARCH=%1
go build -o build\%srv_name%\wdebugger.exe github.com/sutils/webdebugger/wdebugger
if NOT %ERRORLEVEL% EQU 0 goto :efail
xcopy win-%OS%\nssm.exe build\%srv_name%
xcopy cert.bat build\%srv_name%
xcopy wdebugger-conf.bat build\%srv_name%
xcopy wdebugger-install.bat build\%srv_name%
xcopy wdebugger-uninstall.bat build\%srv_name%
xcopy default-wdebugger.json /F build\%srv_name%
echo F | xcopy privoxy-Win.exe /F /Y build\%srv_name%\privoxy.exe


if NOT %ERRORLEVEL% EQU 0 goto :efail

cd build
del /s /a /q %srv_name%-%srv_ver%-Win-%OS%.zip
7z a -r %srv_name%-%srv_ver%-Win-%OS%.zip %srv_name%
if NOT %ERRORLEVEL% EQU 0 goto :efail
cd ..\
goto :esuccess

:efail
echo "Build fail"
pause
exit 1

:esuccess
echo "Build success"
pause