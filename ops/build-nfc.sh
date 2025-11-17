#!/bin/bash
# Build eid-pd-reader with NFC support on Linux/macOS
# Requires OpenPACE installed

set -e

echo "=== Building eid-pd-reader with NFC/PACE support ==="
echo ""

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macOS"
    OPENPACE_PATH="/usr/local/lib/libeac.dylib"
else
    OS="Linux"
    OPENPACE_PATH="/usr/local/lib/libeac.so"
fi

echo "Detected OS: $OS"
echo ""

# Check OpenPACE
if [ ! -f "$OPENPACE_PATH" ] && [ ! -f "/usr/local/lib/libeac.a" ]; then
    echo "ERROR: OpenPACE not found at /usr/local/lib!"
    echo ""
    echo "Please build and install OpenPACE first."
    echo "See ops/build-openpace.md for instructions."
    echo ""
    exit 1
fi

echo "✓ Found OpenPACE library"
echo ""

# Set environment
export CGO_ENABLED=1
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib"

echo "Environment:"
echo "  CGO_ENABLED = $CGO_ENABLED"
echo "  PKG_CONFIG_PATH = $PKG_CONFIG_PATH"
echo "  CGO_CFLAGS = $CGO_CFLAGS"
echo "  CGO_LDFLAGS = $CGO_LDFLAGS"
echo ""

# Clean previous builds
echo "Cleaning previous builds..."
rm -f eid-pd-reader-nfc

# Build
echo "Building with NFC support (this may take a minute)..."
go build -tags openpace -o eid-pd-reader-nfc .

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Build successful!"
    echo ""
    
    # Show version
    echo "Version information:"
    ./eid-pd-reader-nfc --version
    echo ""
    
    # Check capabilities
    echo "Checking NFC capabilities..."
    CAPS=$(echo '{"command":"capabilities"}' | ./eid-pd-reader-nfc 2>/dev/null)
    if echo "$CAPS" | grep -q '"nfcSupported":true'; then
        echo "✓ NFC/PACE support: ENABLED"
    else
        echo "⚠ NFC/PACE support: DISABLED"
    fi
    echo ""
    
    echo "Output: eid-pd-reader-nfc"
    echo ""
    echo "To test NFC reading:"
    echo "  ./eid-pd-reader-nfc --nfc --can YOUR_CAN --dump"
    
else
    echo ""
    echo "✗ Build failed!"
    exit 1
fi

