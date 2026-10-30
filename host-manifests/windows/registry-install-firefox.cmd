@echo off
REM Register native messaging host for Firefox on Windows
REM Run this script as Administrator

echo.
echo ========================================
echo Firefox Native Messaging Host Installer
echo ========================================
echo.

REM Check for admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator
    echo Right-click and select "Run as Administrator"
    pause
    exit /b 1
)

REM Set the path to the manifest file
set MANIFEST_PATH=C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.firefox.json

REM Check if manifest file exists
if not exist "%MANIFEST_PATH%" (
    echo ERROR: Manifest file not found at:
    echo %MANIFEST_PATH%
    echo.
    echo Please make sure the native helper is installed first.
    pause
    exit /b 1
)

echo Installing for Firefox...
echo Manifest: %MANIFEST_PATH%
echo.

REM Register for Firefox
reg add "HKLM\Software\Mozilla\NativeMessagingHosts\ee.aarepuusepp.eid_helper" /ve /t REG_SZ /d "%MANIFEST_PATH%" /f

if %errorLevel% equ 0 (
    echo.
    echo ========================================
    echo SUCCESS: Native host registered!
    echo ========================================
    echo.
    echo Next steps:
    echo 1. Restart Firefox
    echo 2. Install the extension
    echo 3. Test with demo page
    echo.
) else (
    echo.
    echo ERROR: Failed to register native host
    echo Please check the error message above
    echo.
)

pause

