#!/bin/bash
# Bash script to install native host manifest on Linux/macOS

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOST_MANIFEST="$SCRIPT_DIR/../host-manifests/ee.aarepuusepp.eid_helper.json"
HOSTNAME="ee.aarepuusepp.eid_helper"

if [ ! -f "$HOST_MANIFEST" ]; then
    echo "Error: Host manifest not found: $HOST_MANIFEST"
    exit 1
fi

# Get absolute path to eid-pd-reader (assume it's in repo root)
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HELPER_BIN="$REPO_ROOT/eid-pd-reader"

if [ ! -f "$HELPER_BIN" ]; then
    echo "Warning: eid-pd-reader not found at $HELPER_BIN"
    echo "Please build first using ops/build-all.sh"
    echo "Then update the 'path' field in $HOST_MANIFEST to point to the absolute path of eid-pd-reader"
fi

# Update manifest with absolute path (requires jq or manual edit)
if command -v jq &> /dev/null; then
    jq ".path = \"$HELPER_BIN\"" "$HOST_MANIFEST" > "$HOST_MANIFEST.tmp"
    mv "$HOST_MANIFEST.tmp" "$HOST_MANIFEST"
else
    echo "Note: jq not found. Please manually update 'path' in $HOST_MANIFEST to: $HELPER_BIN"
fi

# Determine OS and install location
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux: Chrome/Chromium
    CHROME_HOST_DIR="$HOME/.config/google-chrome/NativeMessagingHosts"
    CHROMIUM_HOST_DIR="$HOME/.config/chromium/NativeMessagingHosts"
    
    mkdir -p "$CHROME_HOST_DIR"
    cp "$HOST_MANIFEST" "$CHROME_HOST_DIR/$HOSTNAME.json"
    echo "Installed for Chrome: $CHROME_HOST_DIR/$HOSTNAME.json"
    
    if [ -d "$(dirname "$CHROMIUM_HOST_DIR")" ]; then
        mkdir -p "$CHROMIUM_HOST_DIR"
        cp "$HOST_MANIFEST" "$CHROMIUM_HOST_DIR/$HOSTNAME.json"
        echo "Installed for Chromium: $CHROMIUM_HOST_DIR/$HOSTNAME.json"
    fi
    
    # Firefox
    FIREFOX_HOST_DIR="$HOME/.mozilla/native-messaging-hosts"
    mkdir -p "$FIREFOX_HOST_DIR"
    FIREFOX_MANIFEST="$SCRIPT_DIR/../host-manifests/ee.aarepuusepp.eid_helper.firefox.json"
    if [ -f "$FIREFOX_MANIFEST" ]; then
        cp "$FIREFOX_MANIFEST" "$FIREFOX_HOST_DIR/$HOSTNAME.json"
        echo "Installed for Firefox: $FIREFOX_HOST_DIR/$HOSTNAME.json"
    fi
    
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS: Chrome
    CHROME_HOST_DIR="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"
    mkdir -p "$CHROME_HOST_DIR"
    cp "$HOST_MANIFEST" "$CHROME_HOST_DIR/$HOSTNAME.json"
    echo "Installed for Chrome: $CHROME_HOST_DIR/$HOSTNAME.json"
    
    # Firefox
    FIREFOX_HOST_DIR="$HOME/Library/Application Support/Mozilla/NativeMessagingHosts"
    mkdir -p "$FIREFOX_HOST_DIR"
    FIREFOX_MANIFEST="$SCRIPT_DIR/../host-manifests/ee.aarepuusepp.eid_helper.firefox.json"
    if [ -f "$FIREFOX_MANIFEST" ]; then
        cp "$FIREFOX_MANIFEST" "$FIREFOX_HOST_DIR/$HOSTNAME.json"
        echo "Installed for Firefox: $FIREFOX_HOST_DIR/$HOSTNAME.json"
    fi
else
    echo "Error: Unsupported OS: $OSTYPE"
    exit 1
fi

echo "Installation complete"

