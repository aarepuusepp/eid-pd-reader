@echo off
REM ============================================================================
REM Estonian ID Card Personal Data Reader - Uninstaller
REM ============================================================================
REM This script removes the native helper and all registry entries
REM
REM Run this script as Administrator
REM ============================================================================

echo.
echo ============================================================================
echo Estonian ID Card Personal Data Reader - Uninstaller
echo ============================================================================
echo.

REM Check for admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator
    echo.
    echo Right-click this file and select "Run as Administrator"
    echo.
    pause
    exit /b 1
)

echo WARNING: This will remove:
echo   - Native helper from C:\Program Files\eid-pd-reader
echo   - Registry entries for Chrome, Edge, and Firefox
echo.
echo Browser extensions will need to be removed manually.
echo.

set /p CONFIRM="Are you sure you want to continue? (Y/N): "
if /i not "%CONFIRM%"=="Y" (
    echo.
    echo Uninstallation cancelled.
    echo.
    pause
    exit /b 0
)

echo.
echo ============================================================================
echo Removing installation...
echo ============================================================================
echo.

set INSTALL_DIR=C:\Program Files\eid-pd-reader

REM ============================================================================
echo [1/4] Removing Chrome registry entry...
echo.

reg delete "HKLM\Software\Google\Chrome\NativeMessagingHosts\ee.aarepuusepp.eid_pd_reader" /f >nul 2>&1

if %errorLevel% equ 0 (
    echo     [OK] Chrome registry entry removed
) else (
    echo     [SKIP] Chrome registry entry not found
)

echo.

REM ============================================================================
echo [2/4] Removing Edge registry entry...
echo.

reg delete "HKLM\Software\Microsoft\Edge\NativeMessagingHosts\ee.aarepuusepp.eid_pd_reader" /f >nul 2>&1

if %errorLevel% equ 0 (
    echo     [OK] Edge registry entry removed
) else (
    echo     [SKIP] Edge registry entry not found
)

echo.

REM ============================================================================
echo [3/4] Removing Firefox registry entry...
echo.

reg delete "HKLM\Software\Mozilla\NativeMessagingHosts\ee.aarepuusepp.eid_pd_reader" /f >nul 2>&1

if %errorLevel% equ 0 (
    echo     [OK] Firefox registry entry removed
) else (
    echo     [SKIP] Firefox registry entry not found
)

echo.

REM ============================================================================
echo [4/4] Removing installation directory...
echo.

if exist "%INSTALL_DIR%" (
    rmdir /s /q "%INSTALL_DIR%"
    if %errorLevel% equ 0 (
        echo     [OK] Removed: %INSTALL_DIR%
    ) else (
        echo     [ERROR] Failed to remove directory
        echo     You may need to remove it manually
    )
) else (
    echo     [SKIP] Installation directory not found
)

echo.

REM ============================================================================
echo ============================================================================
echo Uninstallation Complete!
echo ============================================================================
echo.
echo The native helper has been removed.
echo.
echo To complete the removal:
echo.
echo 1. Remove browser extensions manually:
echo.
echo    Chrome/Edge:
echo      - Open chrome://extensions/ or edge://extensions/
echo      - Find "Estonian ID Card Personal Data Reader"
echo      - Click "Remove"
echo.
echo    Firefox:
echo      - Open about:addons
echo      - Find "Estonian ID Card Personal Data Reader"
echo      - Click "Remove"
echo.
echo 2. Restart your browsers
echo.
echo ============================================================================
echo.

pause

