# Version 1.0.2 Release Notes

**Release Date:** November 14, 2025  
**Type:** Minor Update - Extension Improvements

---

## 🎯 Overview

Version 1.0.2 focuses on improving the browser extension with better version visibility and Mozilla Add-ons compliance.

---

## ✨ What's New

### 1. **Version Display in Extension Popup** 🆕

The extension popup now shows both component versions:

```
┌─────────────────────────────────┐
│ Extension: 1.0.2                │
│ Helper:    1.0.1                │
└─────────────────────────────────┘
```

**Features:**
- ✅ Extension version from manifest
- ✅ Helper version fetched dynamically from running application
- ✅ Error state shows "not available" if helper isn't installed
- ✅ Professional styling with monospace fonts
- ✅ Color-coded (blue for normal, red for errors)

**Benefits:**
- Easy troubleshooting - users can report both versions
- Installation verification - see if helper is properly installed
- Version mismatch detection - spot outdated components
- Professional look - clean, informative UI

---

### 2. **Firefox Manifest Compliance** 🦊

Fixed all Mozilla Add-ons validation warnings:

**Added Data Collection Declaration:**
```json
"data_collection_permissions": {
  "data_collection": false
}
```
- Declares that extension does NOT collect user data
- Required by Mozilla for transparency
- All processing is local, no external data transmission

**Removed Unsupported Property:**
- Removed `service_worker` from Firefox manifest (Chrome-only)
- Firefox uses `scripts` array for background scripts

**Updated Author Information:**
```json
"author": "Aare Puusepp",
"developer": {
  "name": "Aare Puusepp",
  "url": "https://addons.mozilla.org/en-US/firefox/user/17859597/"
}
```

**Validation Result:**
- Before: 0 errors, **2 warnings**, 0 notices
- After: 0 errors, **0 warnings**, 0 notices ✅

---

## 📦 Changes Summary

### Extension Updates

**Chrome Extension (`manifest.json`):**
- ✅ Version updated to 1.0.2
- ✅ Author updated to "Aare Puusepp"

**Firefox Extension (`manifest-firefox.json`):**
- ✅ Version updated to 1.0.2
- ✅ Author updated to "Aare Puusepp"
- ✅ Added `data_collection_permissions` declaration
- ✅ Removed unsupported `service_worker` property
- ✅ Added `developer` object with profile URL

**Background Scripts:**
- ✅ Added `EID_PD_VERSION` message handler
- ✅ Fetches version from native helper on demand

**Popup UI:**
- ✅ Updated HTML to show both versions
- ✅ Added JavaScript to fetch helper version
- ✅ Added CSS styling for version display
- ✅ Error handling for unavailable helper

**Page API:**
- ✅ Version updated to 1.0.2

### Native Helper Updates

**main.go:**
- ✅ Added `version` command handler
- ✅ Returns version info via native messaging

---

## 🔧 Technical Details

### New Version Command

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

### New Message Type: EID_PD_VERSION

**Usage:**
```javascript
const response = await chrome.runtime.sendMessage({ 
  type: 'EID_PD_VERSION' 
});

if (response.ok) {
  console.log('Helper version:', response.version);
}
```

---

## 📥 Installation

### For End Users

**No action required** - This is a browser extension update only. The native helper (1.0.1) remains compatible.

**Chrome/Edge:**
1. Extension will auto-update from Chrome Web Store (when published)
2. Or manually install from `dist/chrome/`

**Firefox:**
1. Extension will auto-update from Mozilla Add-ons (when published)
2. Or manually install from `dist/firefox/`

### For Developers

**Build Extension:**
```powershell
# Chrome/Edge
.\ops\build-chrome-extension.ps1

# Firefox
.\ops\sign-firefox-extension.ps1
```

**Output:**
- Chrome: `dist/chrome/eid-pd-1.0.2.zip`
- Firefox: `dist/firefox/eid-pd-1.0.2.xpi`

---

## 🧪 Testing

### Test Version Display

1. **Install/Update Extension**
   - Load unpacked extension from `extension/` folder
   - Or install built package

2. **Open Extension Popup**
   - Click extension icon in browser toolbar

3. **Verify Version Display**
   - Should show: Extension: 1.0.2
   - Should show: Helper: 1.0.1 (or version installed)

4. **Test Error State**
   - Stop/uninstall native helper
   - Reload extension popup
   - Should show: Helper: not available (in red)

### Test Firefox Validation

```bash
# Validate with web-ext
npx web-ext lint --source-dir=extension

# Expected output:
# Validation Summary:
#   Errors: 0
#   Notices: 0
#   Warnings: 0
```

---

## 🔄 Migration from 1.0.1

### No Breaking Changes

This is a **minor update** with no breaking changes:

- ✅ Native helper 1.0.1 remains compatible
- ✅ All existing APIs work unchanged
- ✅ Configuration files unchanged
- ✅ Host manifests unchanged

### Update Steps

**Option 1: Automatic Update (Recommended)**
- Wait for browser to auto-update extension
- No user action required

**Option 2: Manual Update**
1. Build new extension version
2. Remove old extension
3. Install new extension
4. Native helper remains as-is

---

## 📊 Files Modified

### Extension Files
- ✅ `extension/manifest.json` - Version 1.0.2, author updated
- ✅ `extension/manifest-firefox.json` - Version 1.0.2, Firefox compliance fixes
- ✅ `extension/background.js` - Added EID_PD_VERSION handler
- ✅ `extension/background-firefox.js` - Added EID_PD_VERSION handler
- ✅ `extension/page-api.js` - Version 1.0.2
- ✅ `extension/popup/popup.html` - Version display UI
- ✅ `extension/popup/popup.js` - Version fetching logic
- ✅ `extension/popup/popup.css` - Version display styling

### Native Helper
- ✅ `main.go` - Added version command handler (from 1.0.1)

### Documentation
- ✅ `README.md` - Updated changelog
- ✅ `VERSION-1.0.2-RELEASE.md` - This file
- ✅ `VERSION-DISPLAY-UPDATE.md` - Version display implementation
- ✅ `FIREFOX-MANIFEST-FIXES.md` - Firefox compliance fixes

---

## 🐛 Bug Fixes

None - This release focuses on improvements and compliance.

---

## 🔒 Security

### Privacy Statement

**This extension does NOT collect any user data:**

- ✅ No telemetry or analytics
- ✅ No data sent to external servers
- ✅ All processing is local
- ✅ Personal data stored only in browser's local storage
- ✅ Cache cleared automatically after 5 minutes

**Declared in Firefox manifest:**
```json
"data_collection_permissions": {
  "data_collection": false
}
```

---

## 📚 Documentation

### New Documentation
- **[VERSION-DISPLAY-UPDATE.md](VERSION-DISPLAY-UPDATE.md)** - Complete guide to version display feature
- **[FIREFOX-MANIFEST-FIXES.md](FIREFOX-MANIFEST-FIXES.md)** - Firefox compliance fixes explained

### Updated Documentation
- **[README.md](README.md)** - Changelog updated

### Existing Documentation
- **[APPLICATION-INTEGRATION-GUIDE.md](docs/guides/APPLICATION-INTEGRATION-GUIDE.md)** - Integration guide
- **[INTEGRATION-QUICK-REFERENCE.md](docs/guides/INTEGRATION-QUICK-REFERENCE.md)** - Quick reference
- **[CONFIG-GUIDE.md](CONFIG-GUIDE.md)** - Configuration guide

---

## 🎯 Compatibility

### Browser Compatibility
- ✅ Chrome 109+ (Manifest V3)
- ✅ Edge 109+ (Manifest V3)
- ✅ Firefox 109+ (Manifest V3)

### Operating Systems
- ✅ Windows 10/11
- ✅ Linux (Ubuntu 20.04+, Debian 11+)
- ✅ macOS 10.15+

### Native Helper Compatibility
- ✅ Compatible with helper version 1.0.1
- ✅ Compatible with helper version 1.0.0 (version command not available)

---

## 🚀 Publishing Status

### Chrome Web Store
- Status: Ready for submission
- Package: `dist/chrome/eid-pd-1.0.2.zip`

### Mozilla Add-ons
- Status: Ready for submission
- Package: `dist/firefox/eid-pd-1.0.2.xpi`
- Validation: ✅ 0 errors, 0 warnings, 0 notices

### Edge Add-ons
- Status: Ready for submission
- Package: `dist/chrome/eid-pd-1.0.2.zip` (same as Chrome)

---

## 🎉 Highlights

### User Experience Improvements
1. **Better Visibility** - Users can now see both component versions at a glance
2. **Easier Troubleshooting** - Version information readily available for support
3. **Installation Verification** - Clear indication if helper is installed
4. **Professional UI** - Clean, modern version display

### Developer Experience Improvements
1. **Mozilla Compliance** - No validation warnings
2. **Better Debugging** - Version command for troubleshooting
3. **Consistent Branding** - Author information updated across all files
4. **Complete Documentation** - Comprehensive guides for all new features

---

## 📝 Known Issues

None at this time.

---

## 🔮 What's Next

### Planned for 1.0.3
- [ ] Certificate reading and display
- [ ] Enhanced error messages
- [ ] Performance optimizations

### Future Features
- [ ] NFC support with PACE protocol
- [ ] PIN verification and change
- [ ] Digital signature operations
- [ ] Additional card types support

---

## 💬 Feedback

We welcome your feedback!

- **Issues**: [GitHub Issues](https://github.com/aarepuusepp/web-eid-pd/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aarepuusepp/web-eid-pd/discussions)
- **Email**: aare.puusepp@gmail.com

---

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details

---

## 🙏 Acknowledgments

- Estonian ID Card community
- Mozilla Add-ons review team
- All users and contributors

---

**Version 1.0.2 - Making version information visible and ensuring Mozilla compliance** 🎉

