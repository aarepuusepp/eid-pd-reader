@echo off
REM ============================================================================
REM Estonian ID Card Personal Data Reader - Complete Installer
REM ============================================================================
REM This script installs the native helper and configures it for:
REM - Chrome
REM - Edge  
REM - Firefox
REM
REM Run this script as Administrator
REM ============================================================================

echo.
echo ============================================================================
echo Estonian ID Card Personal Data Reader - Installer
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

REM Get the directory where this script is located
set SCRIPT_DIR=%~dp0
REM Remove trailing backslash
set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

echo [1/6] Checking prerequisites...
echo.
echo     Script directory: %SCRIPT_DIR%
echo.

REM Check if eid-pd-reader.exe exists (check both project root and cmd subdirectory)
set EXE_PATH=
if exist "%SCRIPT_DIR%\eid-pd-reader.exe" (
    set EXE_PATH=%SCRIPT_DIR%\eid-pd-reader.exe
) else if exist "%SCRIPT_DIR%\cmd\eid-pd-reader\eid-pd-reader.exe" (
    set EXE_PATH=%SCRIPT_DIR%\cmd\eid-pd-reader\eid-pd-reader.exe
)

if "%EXE_PATH%"=="" (
    echo ERROR: eid-pd-reader.exe not found
    echo.
    echo Searched locations:
    echo     %SCRIPT_DIR%\eid-pd-reader.exe
    echo     %SCRIPT_DIR%\cmd\eid-pd-reader\eid-pd-reader.exe
    echo.
    echo To build all installable files, run:
    echo     build-release.cmd
    echo.
    echo Then install from the build folder:
    echo     cd build
    echo     install-reader.cmd
    echo.
    pause
    exit /b 1
)

echo     [OK] Found: %EXE_PATH%
echo.

REM ============================================================================
echo [2/6] Creating installation directory...
echo.

set INSTALL_DIR=C:\Program Files\eid-pd-reader

if exist "%INSTALL_DIR%" (
    echo     Directory already exists, cleaning up...
    rmdir /s /q "%INSTALL_DIR%" 2>nul
)

mkdir "%INSTALL_DIR%"
if %errorLevel% neq 0 (
    echo ERROR: Failed to create directory: %INSTALL_DIR%
    pause
    exit /b 1
)

echo     [OK] Created: %INSTALL_DIR%
echo.

REM ============================================================================
echo [3/6] Copying files...
echo.

REM Copy native helper executable
copy /Y "%EXE_PATH%" "%INSTALL_DIR%\eid-pd-reader.exe" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy eid-pd-reader.exe
    echo Source: %EXE_PATH%
    echo Dest:   %INSTALL_DIR%\eid-pd-reader.exe
    pause
    exit /b 1
)
echo     [OK] Copied: eid-pd-reader.exe

REM Find and copy Chrome/Edge host manifest (check build folder first, then host-manifests)
set CHROME_MANIFEST=
if exist "%SCRIPT_DIR%\ee.aarepuusepp.eid_pd_reader.json" (
    set CHROME_MANIFEST=%SCRIPT_DIR%\ee.aarepuusepp.eid_pd_reader.json
) else if exist "%SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.json" (
    set CHROME_MANIFEST=%SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.json
)

if "%CHROME_MANIFEST%"=="" (
    echo ERROR: Chrome/Edge manifest not found
    echo Searched: %SCRIPT_DIR%\ee.aarepuusepp.eid_pd_reader.json
    echo Searched: %SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.json
    pause
    exit /b 1
)

copy /Y "%CHROME_MANIFEST%" "%INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.json" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy Chrome/Edge manifest
    pause
    exit /b 1
)
echo     [OK] Copied: ee.aarepuusepp.eid_pd_reader.json (Chrome/Edge)

REM Find and copy Firefox host manifest (check build folder first, then host-manifests)
set FIREFOX_MANIFEST=
if exist "%SCRIPT_DIR%\ee.aarepuusepp.eid_pd_reader.firefox.json" (
    set FIREFOX_MANIFEST=%SCRIPT_DIR%\ee.aarepuusepp.eid_pd_reader.firefox.json
) else if exist "%SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.firefox.json" (
    set FIREFOX_MANIFEST=%SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.firefox.json
)

if "%FIREFOX_MANIFEST%"=="" (
    echo ERROR: Firefox manifest not found
    echo Searched: %SCRIPT_DIR%\ee.aarepuusepp.eid_pd_reader.firefox.json
    echo Searched: %SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.firefox.json
    pause
    exit /b 1
)

copy /Y "%FIREFOX_MANIFEST%" "%INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.firefox.json" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy Firefox manifest
    pause
    exit /b 1
)
echo     [OK] Copied: ee.aarepuusepp.eid_pd_reader.firefox.json (Firefox)

echo.

REM ============================================================================
echo [4/6] Registering native host for Chrome...
echo.

reg add "HKCU\Software\Google\Chrome\NativeMessagingHosts\ee.aarepuusepp.eid_pd_reader" /ve /t REG_SZ /d "%INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.json" /f >nul 2>&1

if %errorLevel% equ 0 (
    echo     [OK] Chrome native host registered
) else (
    echo     [SKIP] Chrome not installed or registration failed
)

echo.

REM ============================================================================
echo [5/6] Registering native host for Edge...
echo.

reg add "HKCU\Software\Microsoft\Edge\NativeMessagingHosts\ee.aarepuusepp.eid_pd_reader" /ve /t REG_SZ /d "%INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.json" /f >nul 2>&1

if %errorLevel% equ 0 (
    echo     [OK] Edge native host registered
) else (
    echo     [SKIP] Edge not installed or registration failed
)

echo.

REM ============================================================================
echo [6/6] Registering native host for Firefox...
echo.

reg add "HKCU\Software\Mozilla\NativeMessagingHosts\ee.aarepuusepp.eid_pd_reader" /ve /t REG_SZ /d "%INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.firefox.json" /f >nul 2>&1

if %errorLevel% equ 0 (
    echo     [OK] Firefox native host registered
) else (
    echo     [SKIP] Firefox not installed or registration failed
)

echo.

REM ============================================================================
echo ============================================================================
echo Installation Complete!
echo ============================================================================
echo.
echo Installed files:
echo   %INSTALL_DIR%\eid-pd-reader.exe
echo   %INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.json
echo   %INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.firefox.json
echo.
echo Registry entries created for:
echo   - Chrome:  HKCU\Software\Google\Chrome\NativeMessagingHosts
echo   - Edge:    HKCU\Software\Microsoft\Edge\NativeMessagingHosts
echo   - Firefox: HKCU\Software\Mozilla\NativeMessagingHosts
echo.
echo ============================================================================
echo Next Steps:
echo ============================================================================
echo.
echo 1. Install the browser extension:
echo.
echo    Chrome/Edge:
echo      - Open chrome://extensions/ or edge://extensions/
echo      - Enable "Developer mode"
echo      - Click "Load unpacked"
echo      - Select the "extension" folder
echo.
echo    Firefox:
echo      - Open about:debugging#/runtime/this-firefox
echo      - Click "Load Temporary Add-on..."
echo      - Select "extension/manifest-firefox.json"
echo.
echo 2. Test the installation:
echo    %INSTALL_DIR%\eid-pd-reader.exe --dump
echo.
echo 3. Restart your browser for changes to take effect
echo.
echo 4. Insert your ID card and test the extension
echo.
echo ============================================================================
echo Documentation:
echo ============================================================================
echo.
echo   README.md              - Main documentation
echo   EXTENSION-INSTALL.md   - Extension installation guide
echo   BROWSER-SUPPORT.md     - Browser compatibility
echo   TROUBLESHOOTING.md     - Common issues and solutions
echo   NEXT-STEPS.md          - What to do after installation
echo.
echo ============================================================================
echo.

pause

