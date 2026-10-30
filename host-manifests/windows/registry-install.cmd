@echo off
REM Adjust the paths below
set HOSTNAME=ee.aarepuusepp.eid_helper
set JSON=%~dp0..\ee.aarepuusepp.eid_helper.json
reg add "HKLM\Software\Google\Chrome\NativeMessagingHosts\%HOSTNAME%" /ve /t REG_SZ /d "%JSON%" /f
reg add "HKLM\Software\Microsoft\Edge\NativeMessagingHosts\%HOSTNAME%" /ve /t REG_SZ /d "%JSON%" /f
echo Registered %HOSTNAME%

