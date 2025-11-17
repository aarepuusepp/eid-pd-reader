@echo off
REM Adjust the paths below
set HOSTNAME=ee.aarepuusepp.eid_helper
set JSON=%~dp0..\ee.aarepuusepp.eid_helper.json
reg add "HKCU\Software\Google\Chrome\NativeMessagingHosts\%HOSTNAME%" /ve /t REG_SZ /d "%JSON%" /f
reg add "HKCU\Software\Microsoft\Edge\NativeMessagingHosts\%HOSTNAME%" /ve /t REG_SZ /d "%JSON%" /f
echo Registered %HOSTNAME%

