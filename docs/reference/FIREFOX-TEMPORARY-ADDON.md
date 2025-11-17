# Firefox Temporary Add-on Guide

## The Problem

Firefox temporary add-ons get a **random extension ID** each time they're loaded. This causes the native messaging connection to fail because the manifest only allows specific extension IDs.

### Example IDs:
```
First load:  d0af5ecefeac510d4b67ba522e99bd4a6834a984@temporary-addon
Second load: 916cbd01756b1dece80d7adcc12ea3b40c9569b1@temporary-addon
Third load:  <another random ID>@temporary-addon
```

## Solutions

### Solution 1: Update Manifest with New ID (Quick Fix)

Each time you load the temporary addon:

1. **Load the addon** in Firefox (`about:debugging` → Load Temporary Add-on)
2. **Copy the Extension ID** from the addon details
3. **Run the update script:**

```powershell
# Run as Administrator
.\ops\update-firefox-extension-id.ps1 -ExtensionID "YOUR_EXTENSION_ID@temporary-addon"
```

4. **Restart Firefox**
5. **Reload the temporary addon**

### Solution 2: Sign the Extension (Permanent Fix)

For permanent use, sign the extension with Mozilla:

1. **Create Mozilla account** at https://addons.mozilla.org
2. **Get API credentials** from https://addons.mozilla.org/developers/addon/api/key/
3. **Sign the extension:**

```bash
npm install -g web-ext
web-ext sign --api-key=YOUR_KEY --api-secret=YOUR_SECRET
```

4. **Install the signed .xpi file**
   - The signed extension will have a permanent ID: `aare.puusepp@gmail.com`
   - This ID is already in the manifest, so it will work immediately

### Solution 3: Use Firefox Developer Edition

Firefox Developer Edition allows loading unsigned extensions permanently:

1. **Download Firefox Developer Edition**
2. **Set `xpinstall.signatures.required` to `false` in `about:config`**
3. **Install extension from .xpi file**

## Checking Current Extension ID

### Method 1: Firefox UI
1. Open `about:debugging#/runtime/this-firefox`
2. Find "Estonian ID Card Personal Data Reader"
3. Look at "Extension ID" or "Internal UUID"

### Method 2: Browser Console
```javascript
// In browser console (Ctrl+Shift+J)
browser.runtime.id
```

### Method 3: From Extension
```javascript
// In extension's background script or popup
console.log('Extension ID:', browser.runtime.id);
```

## Verifying Manifest

Check what IDs are currently allowed:

```powershell
Get-Content "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.firefox.json" | ConvertFrom-Json | Select-Object -ExpandProperty allowed_extensions
```

## Manual Update

If you prefer to update manually:

1. **Open the manifest file:**
   ```
   C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.firefox.json
   ```

2. **Add your extension ID to `allowed_extensions`:**
   ```json
   {
     "name": "ee.aarepuusepp.eid_helper",
     "description": "Estonian ID Card Personal Data Reader",
     "path": "C:\\Program Files\\eid-pd-reader\\eid-pd-reader.exe",
     "type": "stdio",
     "allowed_extensions": [
       "aare.puusepp@gmail.com",
       "YOUR_NEW_ID@temporary-addon"
     ]
   }
   ```

3. **Restart Firefox**

## Why This Happens

Firefox generates a random ID for temporary add-ons for security reasons:
- Prevents malicious extensions from impersonating legitimate ones
- Ensures each test session is isolated
- Forces developers to properly sign extensions for production

## Recommended Workflow

### For Development:
1. Load temporary addon
2. Copy extension ID
3. Run `update-firefox-extension-id.ps1`
4. Restart Firefox once
5. Continue development (no need to update again until Firefox restart)

### For Production:
1. Sign the extension with Mozilla
2. Use the permanent ID: `aare.puusepp@gmail.com`
3. No manual updates needed

## Troubleshooting

### Extension loads but can't connect to native helper

**Symptom:** Extension shows "Native helper not available"

**Cause:** Extension ID not in manifest

**Fix:**
```powershell
# Check current extension ID
# In Firefox: about:debugging → find your extension → copy ID

# Update manifest
.\ops\update-firefox-extension-id.ps1 -ExtensionID "YOUR_ID@temporary-addon"

# Restart Firefox
```

### Changes not taking effect

1. **Close Firefox completely** (all windows)
2. **Reopen Firefox**
3. **Reload temporary addon**
4. **Reload test page**

### Script fails with "Access Denied"

Run PowerShell as Administrator:
```powershell
Start-Process powershell -Verb RunAs
```

## Alternative: Use Chrome/Edge for Development

Chrome and Edge use stable extension IDs even for unpacked extensions, making development easier:

1. **Chrome:** `chrome://extensions/` → Load unpacked
2. **Edge:** `edge://extensions/` → Load unpacked
3. Extension ID remains the same across browser restarts
4. Update manifest once with the extension ID
5. No need to update again

## See Also

- `EXTENSION-INSTALL.md` - Full extension installation guide
- `BROWSER-SUPPORT.md` - Browser compatibility matrix
- `ops/README.md` - Operations scripts documentation

