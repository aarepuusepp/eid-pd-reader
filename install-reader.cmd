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

REM Check if eid-pd-reader.exe exists
if not exist "%SCRIPT_DIR%\eid-pd-reader.exe" (
    echo ERROR: eid-pd-reader.exe not found in script directory
    echo.
    echo Expected location: %SCRIPT_DIR%\eid-pd-reader.exe
    echo.
    pause
    exit /b 1
)

echo     [OK] eid-pd-reader.exe found
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
copy /Y "%SCRIPT_DIR%\eid-pd-reader.exe" "%INSTALL_DIR%\eid-pd-reader.exe" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy eid-pd-reader.exe
    echo Source: %SCRIPT_DIR%\eid-pd-reader.exe
    echo Dest:   %INSTALL_DIR%\eid-pd-reader.exe
    pause
    exit /b 1
)
echo     [OK] Copied: eid-pd-reader.exe

REM Copy Chrome/Edge host manifest
copy /Y "%SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.json" "%INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.json" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy Chrome/Edge manifest
    echo Source: %SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.json
    echo Dest:   %INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.json
    pause
    exit /b 1
)
echo     [OK] Copied: ee.aarepuusepp.eid_pd_reader.json (Chrome/Edge)

REM Copy Firefox host manifest
copy /Y "%SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.firefox.json" "%INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.firefox.json" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy Firefox manifest
    echo Source: %SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.firefox.json
    echo Dest:   %INSTALL_DIR%\ee.aarepuusepp.eid_pd_reader.firefox.json
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

