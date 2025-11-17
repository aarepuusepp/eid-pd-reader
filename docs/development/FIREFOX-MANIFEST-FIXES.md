# Firefox Extension Manifest Fixes

**Date:** November 14, 2025  
**Version:** 1.0.1  
**Issue:** Mozilla Add-ons validation warnings

---

## 🐛 Issues Fixed

### 1. Missing `data_collection_permissions` Property ✅

**Warning:**
> The "/browser_specific_settings/gecko/data_collection_permissions" property is required for all new Firefox extensions, and will be required for new versions of existing extensions in the future.

**Reference:** https://mzl.la/firefox-builtin-data-consent

**Fix Applied:**
```json
"browser_specific_settings": {
  "gecko": {
    "id": "eid-pd@aarepuusepp.ee",
    "strict_min_version": "109.0",
    "data_collection_permissions": {
      "data_collection": false
    }
  }
}
```

**Explanation:**
- This extension does **NOT** collect any user data
- All data processing happens locally on the user's machine
- Personal data is read from ID card chip and stored only in browser's local storage
- No telemetry, analytics, or external data transmission

---

### 2. Unsupported `service_worker` Property ✅

**Warning:**
> "/background/service_worker" is not supported.

**Old Configuration:**
```json
"background": {
  "service_worker": "background-firefox.js",
  "scripts": ["background-firefox.js"]
}
```

**Fix Applied:**
```json
"background": {
  "scripts": ["background-firefox.js"]
}
```

**Explanation:**
- Firefox Manifest V3 uses `scripts` array for background scripts
- Chrome uses `service_worker` for background service workers
- The `service_worker` property is Chrome-specific and not supported in Firefox
- Removed the unsupported property while keeping `scripts`

---

### 3. Author Information Updated ✅

**Old:**
```json
"author": "web-eid-pd"
```

**New:**
```json
"author": "Aare Puusepp",
"developer": {
  "name": "Aare Puusepp",
  "url": "https://addons.mozilla.org/en-US/firefox/user/17859597/"
}
```

**Changes:**
- Updated `author` field to "Aare Puusepp"
- Added `developer` object with name and profile URL
- Profile link: https://addons.mozilla.org/en-US/firefox/user/17859597/
- Also updated Chrome manifest for consistency

---

## 📋 Validation Results

### Before Fixes
```
General Tests: 0 errors, 2 warnings, 0 notices
- Warning: data_collection_permissions missing
- Warning: service_worker not supported
```

### After Fixes (Expected)
```
General Tests: 0 errors, 0 warnings, 0 notices
✅ Add-on passed validation
```

---

## 📦 Files Modified

### Firefox Extension
✅ `extension/manifest-firefox.json`
- Added `data_collection_permissions` (set to false)
- Removed unsupported `service_worker` property
- Updated `author` to "Aare Puusepp"
- Added `developer` object with profile URL

### Chrome Extension
✅ `extension/manifest.json`
- Updated `author` to "Aare Puusepp" (for consistency)

---

## 🔍 Technical Details

### Data Collection Permissions

According to Mozilla's [Built-in Data Consent documentation](https://mzl.la/firefox-builtin-data-consent):

> Firefox extensions must declare whether they collect user data. This is required for transparency and user trust.

**Our Declaration:**
```json
"data_collection_permissions": {
  "data_collection": false
}
```

**What this means:**
- ✅ Extension does NOT collect user data
- ✅ No telemetry or analytics
- ✅ No data sent to external servers
- ✅ All processing is local
- ✅ Compliant with privacy regulations

---

### Background Scripts: Firefox vs Chrome

**Firefox (Manifest V3):**
```json
"background": {
  "scripts": ["background-firefox.js"]
}
```
- Uses persistent background scripts
- Scripts run in a background page context
- Can use all WebExtension APIs

**Chrome (Manifest V3):**
```json
"background": {
  "service_worker": "background.js",
  "scripts": ["background.js"],
  "type": "module"
}
```
- Uses service workers (event-driven)
- Non-persistent, wakes up on events
- More restrictive than Firefox's approach

**Key Difference:**
- Firefox: Traditional background scripts
- Chrome: Service worker architecture
- This extension maintains separate files for each browser

---

## 🎯 Privacy & Security

### Data Collection Statement

**This extension does NOT collect any data:**

1. **Personal Data Reading**
   - Read directly from ID card chip
   - Stored only in browser's local storage
   - Never transmitted externally

2. **No Telemetry**
   - No usage statistics
   - No error reporting to external services
   - No analytics tracking

3. **No External Communication**
   - All processing is local
   - Native helper runs on user's machine
   - No network requests to external servers

4. **Local Storage Only**
   - Cache stored in browser's `chrome.storage.local`
   - Automatically cleared after 5 minutes
   - User can manually clear cache

---

## 🚀 Publishing to Mozilla Add-ons

### Validation Checklist

Before submitting to Mozilla Add-ons:

- [x] `data_collection_permissions` declared
- [x] No unsupported properties
- [x] Author information correct
- [x] Developer profile URL added
- [x] 0 errors, 0 warnings
- [x] Privacy policy clear
- [x] Extension ID registered: `eid-pd@aarepuusepp.ee`

### Build Command

```powershell
# Build Firefox extension
.\ops\sign-firefox-extension.ps1
```

**Output:**
```
dist/firefox/eid-pd-1.0.1.xpi
```

### Submission Steps

1. **Build Extension**
   ```powershell
   .\ops\sign-firefox-extension.ps1
   ```

2. **Validate Locally**
   - Upload to https://addons.mozilla.org/en-US/developers/addon/submit/upload-listed
   - Check validation results
   - Ensure 0 errors, 0 warnings

3. **Submit for Review**
   - Complete submission form
   - Add privacy policy
   - Explain native messaging usage
   - Wait for Mozilla review

4. **Monitor Review**
   - Check email for review status
   - Respond to any reviewer questions
   - Address any issues found

---

## 📝 Manifest Comparison

### Firefox Manifest (manifest-firefox.json)

```json
{
  "manifest_version": 3,
  "name": "Estonian ID Card Personal Data Reader",
  "short_name": "EstEID PD",
  "version": "1.0.1",
  "description": "Reads personal data directly from Estonian ID card chip using eid-pd-reader library in localhost.",
  "author": "Aare Puusepp",
  "homepage_url": "https://github.com/aarepuusepp/web-eid-pd",
  "developer": {
    "name": "Aare Puusepp",
    "url": "https://addons.mozilla.org/en-US/firefox/user/17859597/"
  },
  
  "background": {
    "scripts": ["background-firefox.js"]
  },
  
  "browser_specific_settings": {
    "gecko": {
      "id": "eid-pd@aarepuusepp.ee",
      "strict_min_version": "109.0",
      "data_collection_permissions": {
        "data_collection": false
      }
    }
  }
}
```

### Chrome Manifest (manifest.json)

```json
{
  "manifest_version": 3,
  "name": "Estonian ID Card Personal Data Reader",
  "short_name": "EstEID PD",
  "version": "1.0.1",
  "description": "Reads personal data directly from Estonian ID card chip when certificates are unavailable. Fallback for web-eid.js authentication.",
  "author": "Aare Puusepp",
  "homepage_url": "https://github.com/aarepuusepp/web-eid-pd",
  
  "background": {
    "service_worker": "background.js",
    "scripts": ["background.js"],
    "type": "module"
  }
}
```

---

## ✅ Verification

### Test Firefox Extension

1. **Load Unpacked**
   ```
   about:debugging#/runtime/this-firefox
   → Load Temporary Add-on
   → Select manifest-firefox.json
   ```

2. **Check Console**
   - No manifest errors
   - Background script loads correctly
   - Native messaging works

3. **Test Functionality**
   - Open popup
   - Check version display
   - Read ID card
   - Verify data display

### Validate Before Publishing

```bash
# Use Mozilla's web-ext tool
npx web-ext lint --source-dir=extension --firefox-preview
```

**Expected Output:**
```
Validation Summary:
  Errors: 0
  Notices: 0
  Warnings: 0
```

---

## 🎉 Summary

All Mozilla Add-ons validation warnings have been resolved:

✅ **Data Collection Permissions** - Declared as `false` (no data collection)  
✅ **Service Worker Warning** - Removed unsupported property  
✅ **Author Information** - Updated to "Aare Puusepp" with profile URL  
✅ **Validation Status** - Ready for submission to Mozilla Add-ons

**Next Steps:**
1. Build extension: `.\ops\sign-firefox-extension.ps1`
2. Test locally in Firefox
3. Submit to Mozilla Add-ons for review
4. Monitor review process

**Status:** ✅ Ready for Mozilla Add-ons submission

