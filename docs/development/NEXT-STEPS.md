# Next Steps - Extension Successfully Loaded! 🎉

Congratulations! The Estonian ID Card Personal Data Reader extension is now loaded in Firefox 145.

## Current Status

✅ **Extension loaded** in Firefox  
✅ **Background script running**  
✅ **SVG icon displayed**  
⚠️ **Native helper not connected** (needs registration)

## Step-by-Step Setup

### Step 1: Register Native Host for Firefox ⚠️ **DO THIS NOW**

The extension needs to communicate with the native helper (`eid-pd-reader.exe`) via Native Messaging. You need to register it in the Windows Registry.

#### Option A: Run Installer (Recommended)

1. **Right-click** on this file:
   ```
   host-manifests\windows\registry-install-firefox.cmd
   ```

2. Select **"Run as Administrator"**

3. You should see:
   ```
   SUCCESS: Native host registered!
   ```

#### Option B: Manual Registration

Run this command in PowerShell **as Administrator**:

```powershell
reg add "HKCU\Software\Mozilla\NativeMessagingHosts\ee.aarepuusepp.eid_helper" /ve /t REG_SZ /d "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.firefox.json" /f
```

#### Verify Registration

Check if the registry entry was created:

```powershell
reg query "HKCU\Software\Mozilla\NativeMessagingHosts\ee.aarepuusepp.eid_helper"
```

Should show:
```
HKEY_CURRENT_USER\Software\Mozilla\NativeMessagingHosts\ee.aarepuusepp.eid_helper
    (Default)    REG_SZ    C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.firefox.json
```

---

### Step 2: Restart Firefox

After registering the native host:

1. **Close Firefox completely** (all windows)
2. **Restart Firefox**
3. Go back to: `about:debugging#/runtime/this-firefox`
4. Your extension should still be loaded

---

### Step 3: Test the Extension

#### A. Test with Extension Popup

1. **Click the extension icon** in the Firefox toolbar (or puzzle icon → Estonian ID Card Reader)

2. The popup should show:
   - ✅ **"Ready to read"** - Native helper is connected
   - ❌ **"Native helper not available"** - Registration failed, go back to Step 1

3. **Insert your ID card** into the reader

4. Click **"Read ID Card"** button

5. Personal data should appear:
   - Surname
   - Given Name
   - Personal Code
   - Document Number
   - etc.

#### B. Test with Demo Page

1. **Open the demo page**:
   ```powershell
   cd demo
   python -m http.server 8000
   ```

2. **Open in Firefox**: `http://localhost:8000`

3. Click **"Check Status"** - should show ✅ Ready

4. Click **"Read ID Card"** - should display personal data

#### C. Test in Browser Console

1. **Open any web page** in Firefox

2. Press **F12** to open Developer Tools

3. Go to **Console** tab

4. Run these commands:

```javascript
// Check if API is available
console.log(window.EstEID_PD);

// Check status
EstEID_PD.checkStatus().then(console.log);

// Read card (insert card first!)
EstEID_PD.readPersonalData().then(console.log);
```

**Expected output:**
```javascript
{
  ok: true,
  data: {
    surname: "JÕEORG",
    givenName: "JAAK-KRISTJAN",
    personalCode: "38001085718",
    // ... more fields
  }
}
```

---

### Step 4: Test with Both Card Types

Test with both IDEMIA and Thales cards:

#### IDEMIA Card
```powershell
# Insert IDEMIA card
.\eid-pd-reader.exe --dump
```

Expected: Personal data displayed

#### Thales Card
```powershell
# Insert Thales card
.\eid-pd-reader.exe --dump
```

Expected: Personal data displayed

Both should work automatically with card detection.

---

## Troubleshooting

### Issue: "Native helper not available"

**Cause**: Native host not registered or Firefox can't find it

**Solution**:
1. Check registry entry exists (see Step 1)
2. Verify manifest file exists: `C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.firefox.json`
3. Verify helper exists: `C:\Program Files\eid-pd-reader\eid-pd-reader.exe`
4. Restart Firefox

### Issue: "No readers found"

**Cause**: Card reader not connected or PC/SC service not running

**Solution**:
```powershell
# Check PC/SC service
Get-Service SCardSvr

# Start if stopped
Start-Service SCardSvr

# Check readers
certutil -scinfo
```

### Issue: "Could not connect to card"

**Cause**: Card not inserted or not seated properly

**Solution**:
1. Remove and reinsert card
2. Clean card contacts
3. Try different reader (if available)

### Issue: Extension shows warnings

**Cause**: Missing icon files (expected, we're using SVG)

**Solution**: Ignore warnings about missing PNG icons. The SVG icon works fine.

---

## What You Can Do Now

### 1. **Integrate into Your Web Application**

```javascript
// In your web page
if (window.EstEID_PD) {
  const result = await EstEID_PD.readPersonalData();
  if (result.ok) {
    // Use the data
    document.getElementById('name').value = 
      result.data.givenName + ' ' + result.data.surname;
    document.getElementById('personalCode').value = 
      result.data.personalCode;
  }
}
```

### 2. **Test All Features**

- ✅ Read personal data
- ✅ Cache management (5-minute cache)
- ✅ Error handling
- ✅ Multiple card types (IDEMIA, Thales)
- ✅ Status checking

### 3. **Deploy to Production**

When ready for production:

1. **Package the extension**:
   ```powershell
   .\ops\package-extension-firefox.ps1
   ```

2. **Submit to Firefox Add-ons**:
   - Go to: https://addons.mozilla.org/developers/
   - Create account
   - Submit the .zip file
   - Wait for review (1-7 days)

3. **Or self-host** for enterprise deployment

### 4. **Test in Other Browsers**

The extension also works in Chrome and Edge:

**Chrome:**
```
1. chrome://extensions/
2. Load unpacked → Select extension folder
3. Register native host (use registry-install.cmd)
```

**Edge:**
```
1. edge://extensions/
2. Load unpacked → Select extension folder
3. Register native host (separate registry entry)
```

---

## Documentation

- **Installation**: [EXTENSION-INSTALL.md](EXTENSION-INSTALL.md)
- **Browser Support**: [BROWSER-SUPPORT.md](BROWSER-SUPPORT.md)
- **Firefox Notes**: [FIREFOX-NOTES.md](FIREFOX-NOTES.md)
- **Unified Manifest**: [UNIFIED-MANIFEST.md](UNIFIED-MANIFEST.md)
- **API Guide**: [EXTENSION-GUIDE.md](EXTENSION-GUIDE.md)
- **Troubleshooting**: [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

---

## Quick Reference

### Test Commands

```powershell
# Test native helper directly
.\eid-pd-reader.exe --dump

# Test with verbose output
.\eid-pd-reader.exe --dump --verbose

# Debug card structure
.\eid-pd-reader.exe --debug

# Check PC/SC service
Get-Service SCardSvr

# List card readers
certutil -scinfo
```

### Browser Console Commands

```javascript
// Check extension
window.EstEID_PD

// Check status
await EstEID_PD.checkStatus()

// Read card
await EstEID_PD.readPersonalData()

// Clear cache
await EstEID_PD.clearCache()
```

---

## Success Criteria

You'll know everything is working when:

✅ Extension popup shows "Ready to read"  
✅ Demo page shows ✅ Ready status  
✅ `EstEID_PD.checkStatus()` returns `{ ok: true, available: true }`  
✅ `EstEID_PD.readPersonalData()` returns personal data  
✅ Both IDEMIA and Thales cards work  
✅ Data is cached for 5 minutes  

---

## Need Help?

1. Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
2. Run diagnostics: `.\ops\diagnose-card.ps1`
3. Check browser console for errors
4. Check native helper: `.\eid-pd-reader.exe --debug`

---

**You're almost done! Just register the native host and you'll be ready to read ID cards from Firefox!** 🚀

