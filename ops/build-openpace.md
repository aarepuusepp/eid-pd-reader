# Building eid-pd-reader with NFC/PACE Support

This guide explains how to build `eid-pd-reader` with NFC (PACE + Secure Messaging) support using OpenPACE library.

## Overview

By default, `eid-pd-reader` is built without NFC support and only works with contact-based smart card readers. To enable NFC reading (required for contactless Estonian ID cards), you need to:

1. Install OpenPACE library
2. Build `eid-pd-reader` with `-tags openpace` flag
3. Ensure CGO is enabled

## Platform-Specific Instructions

### Windows

#### Prerequisites

1. **Install MSYS2** (if not already installed):
   - Download from: https://www.msys2.org/
   - Install to `C:\msys64` (default location)
   - Update MSYS2: `pacman -Syu`

2. **Install MinGW-w64 toolchain and build tools**:
   ```bash
   # Open MSYS2 MinGW 64-bit terminal
   pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-make mingw-w64-x86_64-libtool
   ```

3. **Install OpenSSL**:
   ```bash
   pacman -S mingw-w64-x86_64-openssl
   ```

#### Build OpenPACE

1. **Clone OpenPACE release** (with pre-generated configure script):
   ```bash
   cd /c/tmp
   
   # Remove old clone if exists
   rm -rf openpace
   
   # Clone stable release (1.1.3 has pre-generated configure)
   git clone --depth 1 --branch 1.1.3 https://github.com/frankmorgner/openpace.git
   cd openpace
   ```
   
   **Note**: We use the stable release tag (1.1.3) which includes a pre-generated `configure` script, avoiding the need for `autoreconf` and complex autotools setup.

2. **Configure and build**:
   ```bash
   # Configure (no autoreconf needed!)
   ./configure --prefix=/mingw64 \
       --enable-openssl-install \
       --disable-go \
       --disable-java \
       --disable-python \
       --disable-ruby \
       CRYPTO_CFLAGS="-I/mingw64/include" \
       CRYPTO_LIBS="-L/mingw64/lib -lcrypto"
   
   # Build
   make
   
   # Install
   make install
   ```

3. **Verify installation**:
   ```bash
   ls /mingw64/lib/libeac*
   ls /mingw64/include/eac
   ```

#### Build eid-pd-reader with NFC support

1. **Set up Go environment**:
   ```powershell
   # In PowerShell (not MSYS2)
   $env:CGO_ENABLED = "1"
   $env:CC = "C:\msys64\mingw64\bin\gcc.exe"
   $env:PKG_CONFIG_PATH = "C:\msys64\mingw64\lib\pkgconfig"
   ```

2. **Build**:
   ```powershell
   cd C:\path\to\web-eid-pd
   go build -tags openpace -o eid-pd-reader.exe .
   ```

3. **Test**:
   ```powershell
   .\eid-pd-reader.exe --version
   .\eid-pd-reader.exe --nfc --can 123456 --dump
   ```

#### Troubleshooting (Windows)

- **Error: "target not found: mingw-w64-x86_64-gengetopt" or "mingw-w64-x86_64-help2man"**:
  - These packages are optional and not available in MSYS2 MinGW repos
  - They're only needed for documentation generation
  - Skip them and proceed with OpenPACE build - it will work fine

- **Error: "gcc: not found"**:
  - Add `C:\msys64\mingw64\bin` to your PATH
  - Or set `CC` environment variable as shown above

- **Error: "cannot find -leac"**:
  - Verify OpenPACE is installed: `ls C:\msys64\mingw64\lib\libeac*`
  - Check PKG_CONFIG_PATH is set correctly

- **Configure warnings about gengetopt/help2man**:
  - Safe to ignore - these are optional tools for documentation
  - The library will build successfully without them

- **DLL missing at runtime**:
  - Copy required DLLs from `C:\msys64\mingw64\bin` to the same directory as `eid-pd-reader.exe`:
    - `libeac-1.dll`
    - `libcrypto-3-x64.dll`
    - `libgcc_s_seh-1.dll`
    - `libwinpthread-1.dll`

---

### Linux (Ubuntu/Debian)

#### Prerequisites

1. **Install build tools**:
   ```bash
   sudo apt update
   sudo apt install -y build-essential autoconf automake libtool pkg-config
   ```

2. **Install OpenSSL development files**:
   ```bash
   sudo apt install -y libssl-dev
   ```

3. **Install PC/SC**:
   ```bash
   sudo apt install -y libpcsclite-dev pcscd
   ```

#### Build OpenPACE

1. **Install additional dependencies**:
   ```bash
   sudo apt install -y gengetopt help2man
   ```

2. **Clone and build OpenPACE**:
   ```bash
   cd /tmp
   git clone https://github.com/frankmorgner/openpace.git
   cd openpace
   
   autoreconf -vis
   ./configure --prefix=/usr/local
   make
   sudo make install
   sudo ldconfig
   ```

3. **Verify installation**:
   ```bash
   ls /usr/local/lib/libeac*
   ls /usr/local/include/eac
   ```

#### Build eid-pd-reader with NFC support

1. **Set up Go environment**:
   ```bash
   export CGO_ENABLED=1
   export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
   ```

2. **Build**:
   ```bash
   cd /path/to/web-eid-pd
   go build -tags openpace -o eid-pd-reader .
   ```

3. **Test**:
   ```bash
   ./eid-pd-reader --version
   ./eid-pd-reader --nfc --can 123456 --dump
   ```

#### Troubleshooting (Linux)

- **Error: "eac/eac.h: No such file or directory"**:
  - Verify OpenPACE headers: `ls /usr/local/include/eac`
  - Add to CGO_CFLAGS: `export CGO_CFLAGS="-I/usr/local/include"`

- **Error: "cannot find -leac"**:
  - Verify library: `ls /usr/local/lib/libeac*`
  - Add to CGO_LDFLAGS: `export CGO_LDFLAGS="-L/usr/local/lib"`
  - Run `sudo ldconfig`

- **Runtime error: "libeac.so.3: cannot open shared object file"**:
  - Add library path: `export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH`
  - Or add to `/etc/ld.so.conf.d/openpace.conf`:
    ```
    /usr/local/lib
    ```
  - Then run: `sudo ldconfig`

---

### macOS

#### Prerequisites

1. **Install Homebrew** (if not already installed):
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```

2. **Install build tools**:
   ```bash
   brew install autoconf automake libtool pkg-config
   ```

3. **Install OpenSSL**:
   ```bash
   brew install openssl@3
   ```

#### Build OpenPACE

1. **Install additional dependencies**:
   ```bash
   brew install gengetopt help2man
   ```

2. **Clone and build OpenPACE**:
   ```bash
   cd /tmp
   git clone https://github.com/frankmorgner/openpace.git
   cd openpace
   
   autoreconf -vis
   ./configure --prefix=/usr/local \
       CRYPTO_CFLAGS="-I$(brew --prefix openssl@3)/include" \
       CRYPTO_LIBS="-L$(brew --prefix openssl@3)/lib -lcrypto"
   make
   sudo make install
   ```

3. **Verify installation**:
   ```bash
   ls /usr/local/lib/libeac*
   ls /usr/local/include/eac
   ```

#### Build eid-pd-reader with NFC support

1. **Set up Go environment**:
   ```bash
   export CGO_ENABLED=1
   export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
   export CGO_CFLAGS="-I/usr/local/include"
   export CGO_LDFLAGS="-L/usr/local/lib"
   ```

2. **Build**:
   ```bash
   cd /path/to/web-eid-pd
   go build -tags openpace -o eid-pd-reader .
   ```

3. **Test**:
   ```bash
   ./eid-pd-reader --version
   ./eid-pd-reader --nfc --can 123456 --dump
   ```

#### Troubleshooting (macOS)

- **Error: "ld: library not found for -leac"**:
  - Verify installation: `ls /usr/local/lib/libeac*`
  - Set DYLD_LIBRARY_PATH: `export DYLD_LIBRARY_PATH=/usr/local/lib:$DYLD_LIBRARY_PATH`

- **Error: "dyld: Library not loaded: /usr/local/lib/libeac.3.dylib"**:
  - Add to `~/.zshrc` or `~/.bash_profile`:
    ```bash
    export DYLD_LIBRARY_PATH=/usr/local/lib:$DYLD_LIBRARY_PATH
    ```

---

## Automated Build Scripts

### Windows: `build-nfc.ps1`

```powershell
# Build eid-pd-reader with NFC support on Windows

$ErrorActionPreference = "Stop"

Write-Host "=== Building eid-pd-reader with NFC/PACE support ===" -ForegroundColor Cyan

# Check MSYS2
if (-not (Test-Path "C:\msys64\mingw64\bin\gcc.exe")) {
    Write-Host "ERROR: MSYS2 MinGW64 not found!" -ForegroundColor Red
    Write-Host "Please install MSYS2 and OpenPACE first." -ForegroundColor Yellow
    exit 1
}

# Check OpenPACE
if (-not (Test-Path "C:\msys64\mingw64\lib\libeac.a")) {
    Write-Host "ERROR: OpenPACE not found!" -ForegroundColor Red
    Write-Host "Please build and install OpenPACE first." -ForegroundColor Yellow
    exit 1
}

# Set environment
$env:CGO_ENABLED = "1"
$env:CC = "C:\msys64\mingw64\bin\gcc.exe"
$env:PKG_CONFIG_PATH = "C:\msys64\mingw64\lib\pkgconfig"

Write-Host "Building with NFC support..." -ForegroundColor Green
go build -tags openpace -o eid-pd-reader-nfc.exe .

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host ""
    .\eid-pd-reader-nfc.exe --version
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
```

### Linux/macOS: `build-nfc.sh`

```bash
#!/bin/bash
# Build eid-pd-reader with NFC support on Linux/macOS

set -e

echo "=== Building eid-pd-reader with NFC/PACE support ==="

# Check OpenPACE
if [ ! -f "/usr/local/lib/libeac.a" ] && [ ! -f "/usr/local/lib/libeac.dylib" ]; then
    echo "ERROR: OpenPACE not found!"
    echo "Please build and install OpenPACE first."
    exit 1
fi

# Set environment
export CGO_ENABLED=1
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib"

echo "Building with NFC support..."
go build -tags openpace -o eid-pd-reader-nfc .

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo ""
    ./eid-pd-reader-nfc --version
else
    echo "Build failed!"
    exit 1
fi
```

---

## Testing NFC Build

1. **Check version**:
   ```bash
   ./eid-pd-reader --version
   ```

2. **Test capabilities** (should show `nfcSupported: true`):
   ```bash
   echo '{"command":"capabilities"}' | ./eid-pd-reader
   ```

3. **Test NFC read** (with actual CAN and card):
   ```bash
   ./eid-pd-reader --nfc --can YOUR_CAN --dump
   ```

---

## Distribution

### Windows

When distributing NFC-enabled builds on Windows, include these DLLs:
- `libeac-1.dll`
- `libcrypto-3-x64.dll`
- `libgcc_s_seh-1.dll`
- `libwinpthread-1.dll`

### Linux

Use static linking or document library dependencies:
```bash
ldd ./eid-pd-reader-nfc
```

### macOS

Use `otool` to check dependencies:
```bash
otool -L ./eid-pd-reader-nfc
```

---

## Contact vs NFC Builds

You can maintain two builds:

1. **Standard build** (contact only, no OpenPACE):
   ```bash
   go build -o eid-pd-reader .
   ```

2. **NFC build** (with OpenPACE):
   ```bash
   go build -tags openpace -o eid-pd-reader-nfc .
   ```

The standard build is smaller and has no external dependencies. The NFC build requires OpenPACE but supports both contact and NFC modes.

---

## References

- OpenPACE: https://github.com/frankmorgner/openpace
- OpenPACE Documentation: https://frankmorgner.github.io/openpace/
- Estonian ID Card Specification: https://www.id.ee/en/article/estonian-id-card-specification/

