> Historical development record. Use [current reader README](../../README.md) for supported setup; these notes are not current installation/API instructions.

# Manifest V3 Migration Guide

## Overview

This extension uses **Manifest V3** for all supported browsers (Chrome, Edge, Firefox).

## Browser Support Timeline

### Chrome
- **MV3 Available:** Version 88 (January 2021)
- **MV2 Deprecated:** June 2024
- **MV2 Removed:** Chrome Web Store no longer accepts MV2 extensions
- **Current:** Chrome 142+ (November 2025) - MV3 only

### Edge
- **MV3 Available:** Version 88 (January 2021)
- **MV2 Deprecated:** June 2024
- **MV2 Removed:** Edge Add-ons no longer accepts MV2 extensions
- **Current:** Edge 142+ (November 2025) - MV3 only

### Firefox
- **MV3 Available:** Version 109 (January 2023)
- **MV2 Still Supported:** Yes (until 2026)
- **MV3 Recommended:** Yes
- **Current:** Firefox 145+ (November 2025) - Full MV3 support

## Why Manifest V3?

### Benefits

1. **Better Performance**
   - Service workers instead of persistent background pages
   - Lower memory usage
   - Faster startup

2. **Improved Security**
   - More restrictive permissions model
   - Better isolation
   - Enhanced privacy

3. **Future-Proof**
   - MV2 is deprecated/removed in Chrome/Edge
   - MV3 is the future standard
   - Better long-term support

4. **Cross-Browser Compatibility**
   - Firefox 109+ fully supports MV3
   - Same code works across all browsers
   - Easier maintenance

## Key Changes from MV2 to MV3

### 1. Background Scripts → Service Workers

**MV2:**
```json
"background": {
  "scripts": ["background.js"],
  "persistent": false
}
```

**MV3:**
```json
"background": {
  "service_worker": "background.js",
  "type": "module"
}
```

### 2. Browser Action → Action

**MV2:**
```json
"browser_action": {
  "default_popup": "popup.html"
}
```

**MV3:**
```json
"action": {
  "default_popup": "popup.html"
}
```

### 3. Web Accessible Resources

**MV2:**
```json
"web_accessible_resources": [
  "page-api.js"
]
```

**MV3:**
```json
"web_accessible_resources": [
  {
    "resources": ["page-api.js"],
    "matches": ["<all_urls>"]
  }
]
```

### 4. Host Permissions

**MV2:**
```json
"permissions": [
  "https://example.com/*"
]
```

**MV3:**
```json
"permissions": [],
"host_permissions": [
  "https://example.com/*"
]
```

## Our Implementation

### Manifest V3 Features Used

1. **Service Worker Background**
   - Non-persistent background script
   - Event-driven architecture
   - Automatic lifecycle management

2. **Native Messaging**
   - `nativeMessaging` permission
   - Communication with native helper
   - Works identically in MV3

3. **Content Scripts**
   - Injected into all pages
   - Same API as MV2
   - No changes needed

4. **Storage API**
   - `storage` permission
   - For extension settings
   - Same API as MV2

### Code Compatibility

Our extension is designed to work across all browsers:

```javascript
// background.js works in Chrome, Edge, and Firefox
const NATIVE_HOST = "ee.aarepuusepp.eid_helper";

chrome.runtime.sendNativeMessage(NATIVE_HOST, message, (response) => {
  // Handle response
});
```

Firefox supports both `chrome.*` and `browser.*` APIs in MV3:

```javascript
// This works in Firefox MV3
const browserAPI = typeof browser !== 'undefined' ? browser : chrome;
browserAPI.runtime.sendNativeMessage(...);
```

## Testing Across Browsers

### Chrome 142+ (MV3)
```powershell
# Load extension
chrome://extensions/ → Load unpacked → Select extension folder

# Test
.\eid-pd-reader.exe --dump
```

### Edge 142+ (MV3)
```powershell
# Load extension
edge://extensions/ → Load unpacked → Select extension folder

# Test
.\eid-pd-reader.exe --dump
```

### Firefox 145+ (MV3)
```powershell
# Load extension
about:debugging → Load Temporary Add-on → Select manifest.json

# Test
.\eid-pd-reader.exe --dump
```

## Minimum Version Requirements

To use this extension, users need:

| Browser | Minimum Version | Release Date | Reason |
|---------|----------------|--------------|--------|
| Chrome | 88+ | Jan 2021 | MV3 support |
| Edge | 88+ | Jan 2021 | MV3 support |
| Firefox | 109+ | Jan 2023 | Full MV3 support |

**Note:** These are minimum versions. The extension is tested with the latest versions:
- Chrome 142.0.7444.147
- Edge 142.0.3595.76
- Firefox 145.0

## Migration Checklist

If you're migrating an existing MV2 extension to MV3:

- [x] Update `manifest_version` to 3
- [x] Convert `background.scripts` to `background.service_worker`
- [x] Change `browser_action` to `action`
- [x] Update `web_accessible_resources` format
- [x] Separate `permissions` and `host_permissions`
- [x] Test service worker lifecycle
- [x] Update event listeners for service worker context
- [x] Test in all target browsers
- [x] Update documentation

## Common Issues and Solutions

### Issue 1: Service Worker Stops After Inactivity

**Problem:** Service worker terminates after 30 seconds of inactivity.

**Solution:** Use event-driven architecture. Our extension:
- Listens for `chrome.runtime.onMessage`
- Processes requests asynchronously
- Returns results via callback
- No persistent state needed

### Issue 2: Cannot Use DOM APIs in Service Worker

**Problem:** `document`, `window` not available in service worker.

**Solution:** Use content scripts for DOM access. Our extension:
- Content script (`content-bridge.js`) handles page interaction
- Service worker (`background.js`) handles Native Messaging
- Communication via `chrome.runtime.sendMessage`

### Issue 3: Firefox Compatibility

**Problem:** Firefox has some MV3 differences.

**Solution:** Use compatible APIs. Our extension:
- Uses `chrome.*` API (supported in Firefox MV3)
- Provides `browser.*` fallback if needed
- Tests in both Chrome and Firefox

## Performance Optimization

### Service Worker Best Practices

1. **Keep it Lightweight**
   ```javascript
   // Good: Minimal imports
   const NATIVE_HOST = "ee.aarepuusepp.eid_helper";
   
   // Avoid: Heavy libraries in service worker
   ```

2. **Use Caching**
   ```javascript
   // Cache data to reduce native calls
   let cachedData = null;
   let cacheTimestamp = null;
   const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes
   ```

3. **Async/Await**
   ```javascript
   // Good: Async functions
   async function readPersonalData() {
     const response = await callNativeHelper();
     return response;
   }
   ```

## Future Considerations

### Firefox MV2 Deprecation

Firefox plans to deprecate MV2 in 2026. Our extension is already MV3-ready:
- ✅ Uses MV3 manifest
- ✅ Service worker background
- ✅ Compatible APIs
- ✅ Tested in Firefox 145

### Potential MV3 Updates

Mozilla and Chrome continue to improve MV3:
- Better service worker lifecycle
- Enhanced Native Messaging
- Improved debugging tools

We'll update the extension as new features become available.

## Resources

### Official Documentation

- [Chrome MV3 Documentation](https://developer.chrome.com/docs/extensions/mv3/)
- [Firefox MV3 Documentation](https://extensionworkshop.com/documentation/develop/manifest-v3-migration-guide/)
- [Edge Extensions Documentation](https://docs.microsoft.com/en-us/microsoft-edge/extensions-chromium/)

### Migration Guides

- [Chrome MV2 to MV3 Migration](https://developer.chrome.com/docs/extensions/mv3/mv2-sunset/)
- [Firefox MV3 Migration](https://extensionworkshop.com/documentation/develop/manifest-v3-migration-guide/)

### Our Documentation

- [BROWSER-SUPPORT.md](BROWSER-SUPPORT.md) - Browser compatibility matrix
- [EXTENSION-INSTALL.md](EXTENSION-INSTALL.md) - Installation guide
- [EXTENSION-GUIDE.md](EXTENSION-GUIDE.md) - Developer guide

## Conclusion

This extension is **fully Manifest V3 compliant** and works across:
- ✅ Chrome 88+ (tested with 142.0)
- ✅ Edge 88+ (tested with 142.0)
- ✅ Firefox 109+ (tested with 145.0)

All browsers use the same MV3 codebase with minimal platform-specific adjustments.

---

**Last Updated:** November 2025  
**Manifest Version:** 3  
**Extension Version:** 1.0.0

