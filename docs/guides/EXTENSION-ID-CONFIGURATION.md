# Extension ID Configuration Guide

After installing the browser extension, you must configure its ID in the native reader's host manifest to enable communication between the two components.

## Why is this needed?

The native reader uses host manifests to control which browser extensions can communicate with it. This is a security feature that prevents unauthorized extensions from accessing the card reader.

## Quick Steps

### For Chrome/Edge

1. **Get Extension ID**
   - Open `chrome://extensions/` or `edge://extensions/`
   - Find "Estonian ID Card Personal Data Reader"
   - Copy the Extension ID (looks like: `abcdefghijklmnopqrstuvwxyz123456`)

2. **Update Host Manifest**
   - Open: `C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_pd_reader.json`
   - Find the line: `"chrome-extension://YOUR_EXTENSION_ID/"`
   - Replace `YOUR_EXTENSION_ID` with your actual extension ID
   - Save the file

3. **Reload Extension**
   - Go back to `chrome://extensions/`
   - Click the reload button on the extension
   - Test by opening the extension popup

### For Firefox

Firefox uses a fixed extension ID, so no configuration is needed!

- Extension ID: `eid-pd@aarepuusepp.ee`
- Host manifest: `C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_pd_reader.firefox.json`
- Already configured correctly

## Detailed Instructions

### Windows

#### Step 1: Find Your Extension ID

**Chrome/Edge:**
1. Open your browser
2. Navigate to extensions page:
   - Chrome: `chrome://extensions/`
   - Edge: `edge://extensions/`
3. Enable "Developer mode" (toggle in top right)
4. Find "Estonian ID Card Personal Data Reader" in the list
5. The Extension ID is shown under the extension name
6. Copy the entire ID (32 characters)

**Firefox:**
1. No action needed - ID is fixed as `eid-pd@aarepuusepp.ee`

#### Step 2: Locate Host Manifest File

**Chrome/Edge:**
```
C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_pd_reader.json
```

**Firefox:**
```
C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_pd_reader.firefox.json
```

#### Step 3: Edit Host Manifest

**For Chrome/Edge:**

1. Open the file in a text editor (run as Administrator)
2. Find this section:
   ```json
   "allowed_origins": [
     "chrome-extension://YOUR_EXTENSION_ID/"
   ]
   ```
3. Replace `YOUR_EXTENSION_ID` with your actual ID:
   ```json
   "allowed_origins": [
     "chrome-extension://abcdefghijklmnopqrstuvwxyz123456/"
   ]
   ```
4. Save the file

**For Firefox:**
- No changes needed - already configured

#### Step 4: Reload Extension

1. Go to extensions page
2. Find your extension
3. Click the reload/refresh button
4. Open the extension popup
5. Verify it shows "Helper: 1.0.2" (not "not available")

### Linux

#### Step 1: Find Extension ID

Same as Windows instructions above.

#### Step 2: Locate Host Manifest

**Chrome:**
```
~/.config/google-chrome/NativeMessagingHosts/ee.aarepuusepp.eid_pd_reader.json
```

**Firefox:**
```
~/.mozilla/native-messaging-hosts/ee.aarepuusepp.eid_pd_reader.json
```

#### Step 3: Edit Host Manifest

```bash
# For Chrome
nano ~/.config/google-chrome/NativeMessagingHosts/ee.aarepuusepp.eid_pd_reader.json

# For Firefox (no changes needed, but to view)
nano ~/.mozilla/native-messaging-hosts/ee.aarepuusepp.eid_pd_reader.json
```

Update the `allowed_origins` section as described in Windows Step 3.

### macOS

#### Step 1: Find Extension ID

Same as Windows instructions above.

#### Step 2: Locate Host Manifest

**Chrome:**
```
~/Library/Application Support/Google/Chrome/NativeMessagingHosts/ee.aarepuusepp.eid_pd_reader.json
```

**Firefox:**
```
~/Library/Application Support/Mozilla/NativeMessagingHosts/ee.aarepuusepp.eid_pd_reader.json
```

#### Step 3: Edit Host Manifest

```bash
# For Chrome
nano ~/Library/Application\ Support/Google/Chrome/NativeMessagingHosts/ee.aarepuusepp.eid_pd_reader.json

# For Firefox (no changes needed)
nano ~/Library/Application\ Support/Mozilla/NativeMessagingHosts/ee.aarepuusepp.eid_pd_reader.json
```

Update the `allowed_origins` section as described in Windows Step 3.

## Automated Configuration

### Using PowerShell Script (Windows)

Located in extension folder: `eid-pd-extension/ops/setup-chrome-extension.ps1`

```powershell
cd eid-pd-extension/ops
.\setup-chrome-extension.ps1 -ExtensionId "YOUR_EXTENSION_ID"
```

This script automatically updates the host manifest with your extension ID.

## Verification

### Test Connection

1. Open the extension popup
2. Check the version display:
   - **Good:** "Helper: 1.0.2"
   - **Bad:** "Helper: not available"

3. Try reading a card:
   - Insert ID card
   - Click "Read ID Card"
   - Should display personal data

### Troubleshooting

**"Native helper not available"**

Possible causes:
1. Extension ID not configured in host manifest
2. Host manifest file path incorrect
3. Reader not installed
4. Registry entry missing (Windows)

**Solution:**
1. Verify extension ID is correct
2. Check host manifest file exists and is readable
3. Reinstall reader if needed
4. Check browser console for error messages

**Extension ID keeps changing**

This happens with unpacked extensions in Chrome/Edge. Solutions:
1. Pack the extension to get a fixed ID
2. Keep the .pem file when packing
3. Use the same .pem file for future updates

## Security Notes

- Only add extension IDs you trust
- Each extension ID must be listed separately
- The host manifest controls access to the card reader
- Never share your extension's .pem file

## Multiple Extensions

To allow multiple extensions to access the reader:

```json
"allowed_origins": [
  "chrome-extension://extension_id_1/",
  "chrome-extension://extension_id_2/"
]
```

For Firefox:
```json
"allowed_extensions": [
  "eid-pd@aarepuusepp.ee",
  "other-extension@example.com"
]
```

## Related Documentation

- [Reader Installation](../installation/INSTALLATION.md)
- [Extension Installation](../../../eid-pd-extension/docs/installation/EXTENSION-INSTALL.md)
- [Troubleshooting](../reference/TROUBLESHOOTING.md)

## Support

If you encounter issues:
1. Check the [Troubleshooting Guide](../reference/TROUBLESHOOTING.md)
2. Open an issue on [GitHub](https://github.com/aarepuusepp/web-eid-pd/issues)
3. Email: aare.puusepp@gmail.com

