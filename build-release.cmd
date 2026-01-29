@echo off
REM ============================================================================
REM Estonian ID Card Personal Data Reader - Release Build Script
REM ============================================================================
REM This script creates a build folder with all installable components:
REM - eid-pd-reader.exe (native helper)
REM - Host manifests for Chrome/Edge and Firefox
REM - Install/uninstall scripts
REM ============================================================================

echo.
echo ============================================================================
echo Estonian ID Card Personal Data Reader - Release Build
echo ============================================================================
echo.

REM Get the directory where this script is located
set SCRIPT_DIR=%~dp0
REM Remove trailing backslash
set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

set BUILD_DIR=%SCRIPT_DIR%\build

echo [1/5] Cleaning build directory...
echo.

if exist "%BUILD_DIR%" (
    echo     Removing old build folder...
    rmdir /s /q "%BUILD_DIR%"
)

mkdir "%BUILD_DIR%"
if %errorLevel% neq 0 (
    echo ERROR: Failed to create build directory
    pause
    exit /b 1
)

echo     [OK] Created: %BUILD_DIR%
echo.

REM ============================================================================
echo [2/5] Building eid-pd-reader.exe...
echo.

REM Check if Go is available
where go >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    echo.
    echo Please install Go from https://golang.org/dl/
    echo.
    pause
    exit /b 1
)

echo     Go version:
go version
echo.

cd /d "%SCRIPT_DIR%"
go build -o "%BUILD_DIR%\eid-pd-reader.exe" ./cmd/eid-pd-reader

if %errorLevel% neq 0 (
    echo ERROR: Build failed
    pause
    exit /b 1
)

echo     [OK] Built: eid-pd-reader.exe
echo.

REM ============================================================================
echo [3/5] Copying host manifests...
echo.

REM Copy Chrome/Edge manifest
copy /Y "%SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.json" "%BUILD_DIR%\" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy Chrome/Edge manifest
    pause
    exit /b 1
)
echo     [OK] Copied: ee.aarepuusepp.eid_pd_reader.json (Chrome/Edge)

REM Copy Firefox manifest
copy /Y "%SCRIPT_DIR%\host-manifests\ee.aarepuusepp.eid_pd_reader.firefox.json" "%BUILD_DIR%\" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy Firefox manifest
    pause
    exit /b 1
)
echo     [OK] Copied: ee.aarepuusepp.eid_pd_reader.firefox.json (Firefox)

echo.

REM ============================================================================
echo [4/5] Copying install scripts...
echo.

REM Copy install script
copy /Y "%SCRIPT_DIR%\install-reader.cmd" "%BUILD_DIR%\" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy install script
    pause
    exit /b 1
)
echo     [OK] Copied: install-reader.cmd

REM Copy uninstall script
copy /Y "%SCRIPT_DIR%\uninstall-reader.cmd" "%BUILD_DIR%\" >nul
if %errorLevel% neq 0 (
    echo ERROR: Failed to copy uninstall script
    pause
    exit /b 1
)
echo     [OK] Copied: uninstall-reader.cmd

echo.

REM ============================================================================
echo [5/5] Verifying build...
echo.

echo     Checking executable...
"%BUILD_DIR%\eid-pd-reader.exe" --version
if %errorLevel% neq 0 (
    echo WARNING: Version check failed
)

echo.
echo     Build contents:
echo.
dir /b "%BUILD_DIR%"

echo.

REM ============================================================================
echo ============================================================================
echo Build Complete!
echo ============================================================================
echo.
echo Build output: %BUILD_DIR%
echo.
echo Contents:
echo   - eid-pd-reader.exe              (Native helper executable)
echo   - ee.aarepuusepp.eid_pd_reader.json         (Chrome/Edge manifest)
echo   - ee.aarepuusepp.eid_pd_reader.firefox.json (Firefox manifest)
echo   - install-reader.cmd             (Installation script)
echo   - uninstall-reader.cmd           (Uninstallation script)
echo.
echo To install, run as Administrator:
echo   cd build
echo   install-reader.cmd
echo.
echo ============================================================================
echo.

pause
