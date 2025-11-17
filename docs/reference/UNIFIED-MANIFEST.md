# Unified Manifest V3 for Chrome and Firefox

## Overview

Since **Chrome 121** (January 2024) and **Firefox 121** (December 2023), you can use a **single manifest** that works in both browsers by including both `service_worker` and `scripts` properties.

## The Modern Approach

### Single Manifest for All Browsers

```json
{
  "manifest_version": 3,
  "name": "Estonian ID Card Personal Data Reader",
  "version": "1.0.0",
  
  "background": {
    "service_worker": "background.js",
    "scripts": ["background.js"],
    "type": "module"
  }
}
```

**How it works:**
- ✅ **Chrome 121+** uses `service_worker`, ignores `scripts`
- ✅ **Firefox 121+** uses `scripts`, ignores `service_worker`
- ✅ **Edge 121+** uses `service_worker`, ignores `scripts`
- ✅ **One manifest, all browsers!**

## Browser Version Requirements

| Browser | Version | Release Date | Behavior |
|---------|---------|--------------|----------|
| **Chrome 120 and earlier** | < 121 | Before Jan 2024 | ❌ Refuses to load if `scripts` present |
| **Chrome 121+** | ≥ 121 | Jan 2024+ | ✅ Ignores `scripts`, uses `service_worker` |
| **Firefox 120 and earlier** | < 121 | Before Dec 2023 | ❌ Doesn't start if `service_worker` present |
| **Firefox 121+** | ≥ 121 | Dec 2023+ | ✅ Ignores `service_worker`, uses `scripts` |
| **Edge 121+** | ≥ 121 | Jan 2024+ | ✅ Ignores `scripts`, uses `service_worker` |

## Current Browser Versions (November 2025)

| Browser | Current Version | Supports Unified Manifest |
|---------|-----------------|---------------------------|
| Chrome | 142.0 | ✅ Yes (since 121) |
| Edge | 142.0 | ✅ Yes (since 121) |
| Firefox | 145.0 | ✅ Yes (since 121) |

**All current browsers support the unified manifest approach!**

## Implementation in This Project

### Our Manifests

We maintain two manifests for clarity and Firefox-specific settings:

#### `manifest.json` (Chrome/Edge/Firefox)
```json
{
  "manifest_version": 3,
  "background": {
    "service_worker": "background.js",
    "scripts": ["background.js"],
    "type": "module"
  },
  "browser_specific_settings": {
    "gecko": {
      "id": "aare.puusepp@gmail.com"
    }
  }
}
```

#### `manifest-firefox.json` (Firefox-specific)
```json
{
  "manifest_version": 3,
  "background": {
    "service_worker": "background-firefox.js",
    "scripts": ["background-firefox.js"]
  },
  "browser_specific_settings": {
    "gecko": {
      "id": "aare.puusepp@gmail.com",
      "strict_min_version": "109.0"
    }
  }
}
```

**Why two manifests?**
- Different background scripts (`background.js` vs `background-firefox.js`)
- Firefox uses `browser.*` API with Chrome fallback
- Easier to maintain browser-specific configurations
- Clearer separation for development

**Could we use one?** Yes! With Chrome 121+ and Firefox 121+, you could use a single `manifest.json` with the same background script for both browsers.

## Migration Guide

### If You Have Separate Manifests (Old Approach)

**Before (Chrome):**
```json
"background": {
  "service_worker": "background.js"
}
```

**Before (Firefox):**
```json
"background": {
  "scripts": ["background.js"]
}
```

**After (Unified):**
```json
"background": {
  "service_worker": "background.js",
  "scripts": ["background.js"]
}
```

### Testing the Unified Manifest

1. **Update manifest.json:**
   ```json
   "background": {
     "service_worker": "background.js",
     "scripts": ["background.js"],
     "type": "module"
   }
   ```

2. **Test in Chrome:**
   ```
   chrome://extensions/ → Load unpacked → Select extension folder
   ```

3. **Test in Firefox:**
   ```
   about:debugging → Load Temporary Add-on → Select manifest.json
   ```

4. **Verify both work:**
   - Check console for errors
   - Test Native Messaging
   - Verify all features work

## Benefits of Unified Manifest

### ✅ Advantages

1. **Single Codebase**
   - One manifest file
   - Easier maintenance
   - Less duplication

2. **Simplified Build**
   - One package for all browsers
   - No browser-specific builds
   - Faster deployment

3. **Easier Testing**
   - Test same package everywhere
   - Consistent behavior
   - Fewer edge cases

4. **Future-Proof**
   - Works with current browsers
   - Compatible with future versions
   - Standard approach

### ⚠️ Considerations

1. **Minimum Version Requirements**
   - Requires Chrome 121+ (Jan 2024)
   - Requires Firefox 121+ (Dec 2023)
   - Users on older versions won't be able to install

2. **Browser-Specific Code**
   - Still need to handle API differences
   - `browser.*` vs `chrome.*`
   - Different native host manifest formats

## Browser Support Timeline

### Chrome

```
v88  (Jan 2021)  ─── MV3 available
                     service_worker only
                     
v120 (Dec 2023)  ─── Rejects if scripts present
                     
v121 (Jan 2024)  ─── Ignores scripts ✅
                     Unified manifest works
                     
v142 (Nov 2025)  ─── Current version
```

### Firefox

```
v109 (Jan 2023)  ─── MV3 available
                     scripts only
                     
v120 (Nov 2023)  ─── Doesn't start if service_worker present
                     
v121 (Dec 2023)  ─── Ignores service_worker ✅
                     Unified manifest works
                     
v145 (Nov 2025)  ─── Current version
```

## Code Example

### Unified Background Script

Since both browsers can use the same manifest, you can also use the same background script:

```javascript
// background.js - Works in both Chrome and Firefox

// Use browser API with Chrome fallback
const browserAPI = typeof browser !== 'undefined' ? browser : chrome;

const NATIVE_HOST = "ee.aarepuusepp.eid_helper";

// Native Messaging
browserAPI.runtime.sendNativeMessage(NATIVE_HOST, message, (response) => {
  // Handle response
});

// Message handling
browserAPI.runtime.onMessage.addListener((request, sender, sendResponse) => {
  // Handle messages
  return true; // Keep channel open
});
```

## Recommendations

### For New Extensions

✅ **Use unified manifest** with both properties:
```json
"background": {
  "service_worker": "background.js",
  "scripts": ["background.js"]
}
```

### For Existing Extensions

✅ **Update when ready:**
1. Verify Chrome 121+ / Firefox 121+ support
2. Add both properties to manifest
3. Test in all browsers
4. Update minimum version requirements

### For This Project

⚠️ **Current approach:**
- Keep separate manifests for clarity
- Different background scripts for each browser
- Easier to maintain browser-specific code

✅ **Could migrate to:**
- Single manifest with both properties
- Single background script with API detection
- Simpler build process

## References

### Mozilla Bug Report

- **Firefox bug 1860304**: "Background page doesn't start if service_worker is present"
- **Fixed in:** Firefox 121 (December 2023)
- **Link:** https://bugzilla.mozilla.org/show_bug.cgi?id=1860304

### Chrome Release Notes

- **Chrome 121 Release**: January 2024
- **Change:** Ignores `background.scripts` in Manifest V3
- **Impact:** Allows unified manifests with both properties

### Documentation

- [MDN: background (manifest.json)](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/manifest.json/background)
- [Chrome: Manifest V3 background](https://developer.chrome.com/docs/extensions/mv3/background_pages/)

## Conclusion

**Modern Approach (2025):**
```json
"background": {
  "service_worker": "background.js",
  "scripts": ["background.js"]
}
```

✅ **Works in:** Chrome 121+, Edge 121+, Firefox 121+  
✅ **Coverage:** 99%+ of users (all current versions)  
✅ **Benefit:** Single manifest for all browsers  

**This is now the recommended approach for new Manifest V3 extensions!**

---

**Last Updated:** November 2025  
**Browser Versions:** Chrome 142, Edge 142, Firefox 145  
**Status:** Fully supported across all major browsers

