> Historical development record. Use [current reader README](../../README.md) for supported setup; these notes are not current installation/API instructions.

# Version Display Update

**Date:** 2024-11-13  
**Feature:** Display both extension and helper versions in popup

---

## ✨ What's New

The extension popup now displays **both** version numbers:
- **Extension Version** - From the manifest
- **Helper Version** - Retrieved directly from the native helper application

### Visual Example

```
┌─────────────────────────────────────┐
│  Estonian ID Card Reader            │
│  Personal Data from Chip            │
├─────────────────────────────────────┤
│  Status: Ready to read              │
│  [Read ID Card] [Clear Cache]      │
├─────────────────────────────────────┤
│  Extension: 1.0.1                   │
│  Helper:    1.0.1                   │
│  Documentation                      │
└─────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### 1. New Command in Native Helper

Added `version` command to main.go:

```go
case "version":
    // Return version information
    logger.LogNativeMessage("request", "version", true, "")
    versionData := map[string]string{
        "version": Version,
        "name":    AppName,
    }
    _ = writeNativeMessage(map[string]interface{}{
        "ok":      true,
        "version": Version,
        "name":    AppName,
        "data":    versionData,
    })
```

**Request:**
```json
{
  "command": "version"
}
```

**Response:**
```json
{
  "ok": true,
  "version": "1.0.1",
  "name": "eid-pd-reader",
  "data": {
    "version": "1.0.1",
    "name": "eid-pd-reader"
  }
}
```

---

### 2. Extension Background Script

Added `EID_PD_VERSION` message handler:

**background.js:**
```javascript
if (request.type === 'EID_PD_VERSION') {
  // Get version from native helper
  (async () => {
    try {
      const result = await callNativeHelper({ command: 'version' });
      sendResponse(result);
    } catch (error) {
      sendResponse({ 
        ok: false, 
        error: error.message
      });
    }
  })();
  return true;
}
```

**background-firefox.js:** Same implementation for Firefox compatibility.

---

### 3. Popup UI Updates

**popup.html:**
```html
<footer>
  <div class="version-info">
    <p><strong>Extension:</strong> <span id="extensionVersion">1.0.1</span></p>
    <p><strong>Helper:</strong> <span id="helperVersion">checking...</span></p>
  </div>
  <p><a href="https://github.com/aarepuusepp/web-eid-pd" target="_blank">Documentation</a></p>
</footer>
```

**popup.js:**
```javascript
// Set extension version
extensionVersionSpan.textContent = chrome.runtime.getManifest().version;

// Get helper version
async function getHelperVersion() {
  try {
    const response = await chrome.runtime.sendMessage({ type: 'EID_PD_VERSION' });
    if (response.ok && response.version) {
      helperVersionSpan.textContent = response.version;
      helperVersionSpan.classList.remove('error');
    } else {
      helperVersionSpan.textContent = 'not available';
      helperVersionSpan.classList.add('error');
    }
  } catch (error) {
    helperVersionSpan.textContent = 'not available';
    helperVersionSpan.classList.add('error');
  }
}

// Fetch helper version
getHelperVersion();
```

**popup.css:**
```css
footer .version-info {
  display: flex;
  justify-content: space-around;
  margin-bottom: 8px;
  padding: 8px;
  background: white;
  border-radius: 4px;
  font-size: 11px;
}

footer .version-info span {
  font-family: 'Courier New', monospace;
  color: #0072CE;
  font-weight: 600;
}

footer .version-info span.error {
  color: #dc3545;
  font-style: italic;
}
```

---

## 📊 Files Modified

### Native Helper
- ✅ `main.go` - Added version command handler

### Extension
- ✅ `extension/background.js` - Added EID_PD_VERSION handler
- ✅ `extension/background-firefox.js` - Added EID_PD_VERSION handler
- ✅ `extension/popup/popup.html` - Updated footer with version info
- ✅ `extension/popup/popup.js` - Added helper version fetching
- ✅ `extension/popup/popup.css` - Styled version display

---

## 🎯 Benefits

### 1. **Version Verification**
Users can now verify that both components are installed and up-to-date:
- Extension version from manifest
- Helper version from actual running application

### 2. **Troubleshooting**
When reporting issues, users can easily provide both version numbers:
```
Extension: 1.0.1
Helper: 1.0.1
```

### 3. **Version Mismatch Detection**
If extension and helper versions don't match, it's immediately visible:
```
Extension: 1.0.1
Helper: 1.0.0  ← Outdated!
```

### 4. **Installation Verification**
If helper shows "not available", user knows the native helper isn't properly installed:
```
Extension: 1.0.1
Helper: not available  ← Not installed!
```

---

## 🧪 Testing

### Test Helper Version Command

**Command line:**
```bash
# Test version command (not available via CLI, only native messaging)
# Use --version flag instead
eid-pd-reader.exe --version
```

**Output:**
```
eid-pd-reader version 1.0.1
Estonian ID Card Personal Data Reader
https://github.com/aarepuusepp/web-eid-pd
```

### Test Extension Popup

1. **Open extension popup**
2. **Check footer** - Should show:
   - Extension: 1.0.1
   - Helper: 1.0.1 (or "checking..." then version)

### Test Error State

1. **Stop native helper** (uninstall or rename executable)
2. **Open extension popup**
3. **Check footer** - Should show:
   - Extension: 1.0.1
   - Helper: not available (in red, italic)

---

## 🔍 Error Handling

### Helper Not Available

**Display:**
```
Extension: 1.0.1
Helper: not available
```

**CSS Class:** `error` (red, italic)

**When this happens:**
- Native helper not installed
- Native helper not registered
- Native messaging connection failed
- Helper crashed or not responding

### Version Fetch Failed

**Display:**
```
Extension: 1.0.1
Helper: not available
```

**When this happens:**
- Native messaging error
- Helper doesn't support version command (old version)
- Communication timeout

---

## 📝 API Documentation

### New Message Type: EID_PD_VERSION

**Request:**
```javascript
chrome.runtime.sendMessage({ type: 'EID_PD_VERSION' })
```

**Success Response:**
```javascript
{
  ok: true,
  version: "1.0.1",
  name: "eid-pd-reader",
  data: {
    version: "1.0.1",
    name: "eid-pd-reader"
  }
}
```

**Error Response:**
```javascript
{
  ok: false,
  error: "Native helper not available"
}
```

---

## 🎨 UI Design

### Version Info Box

**Layout:**
- Horizontal flex layout
- Two columns (Extension | Helper)
- White background with rounded corners
- Monospace font for version numbers
- Blue color for normal versions
- Red italic for errors

**Spacing:**
- 8px padding
- 4px gap between label and version
- 8px margin bottom

**Typography:**
- Labels: 11px, bold, gray
- Versions: 11px, monospace, blue/red
- Error state: italic, red

---

## 🚀 Future Enhancements

### Potential Improvements

1. **Version Comparison**
   ```javascript
   if (extensionVersion !== helperVersion) {
     showWarning('Version mismatch detected');
   }
   ```

2. **Update Notification**
   ```javascript
   if (helperVersion < latestVersion) {
     showUpdatePrompt('New version available');
   }
   ```

3. **Detailed Version Info**
   - Build date
   - Git commit hash
   - Platform information

4. **Click to Copy**
   ```javascript
   versionSpan.addEventListener('click', () => {
     navigator.clipboard.writeText(version);
     showToast('Version copied to clipboard');
   });
   ```

---

## 📚 Related Documentation

- [VERSION-1.0.1-RELEASE.md](VERSION-1.0.1-RELEASE.md) - Release notes
- [README.md](README.md) - Main documentation
- [docs/guides/INTEGRATION-QUICK-REFERENCE.md](docs/guides/INTEGRATION-QUICK-REFERENCE.md) - API reference

---

## ✅ Verification Checklist

- [x] Version command added to main.go
- [x] Background scripts handle EID_PD_VERSION
- [x] Popup HTML updated with version display
- [x] Popup JS fetches helper version
- [x] CSS styled for version info
- [x] Error states handled gracefully
- [x] Both browsers supported (Chrome/Firefox)
- [x] Documentation updated

---

## 🎉 Summary

The extension popup now provides complete version information:
- **Extension version** from manifest (static)
- **Helper version** from running application (dynamic)

This makes troubleshooting easier and helps users verify their installation is complete and up-to-date.

**Status:** ✅ Complete and tested

