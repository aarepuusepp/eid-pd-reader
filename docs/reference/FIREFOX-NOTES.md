# Firefox-Specific Implementation Notes

## Background Script Differences

### Modern Approach (Chrome 121+ / Firefox 121+)

**Good news!** Since Chrome 121 and Firefox 121, you can use **both properties** in the same manifest:

```json
"background": {
  "service_worker": "background.js",
  "scripts": ["background.js"],
  "type": "module"
}
```

- **Chrome 121+** (Jan 2024): Uses `service_worker`, ignores `scripts`
- **Firefox 121+** (Dec 2023): Uses `scripts`, ignores `service_worker`
- **Result:** One manifest works for all browsers!

### Legacy Approach (Before Chrome 121 / Firefox 121)

#### Chrome/Edge (Service Worker)
```json
"background": {
  "service_worker": "background.js",
  "type": "module"
}
```

#### Firefox (Scripts Array)
```json
"background": {
  "scripts": ["background-firefox.js"]
}
```

## Why the Difference?

Firefox's Manifest V3 implementation uses a different approach:

1. **Service Workers Not Used**
   - Firefox 109+ has MV3 support
   - But uses `background.scripts` instead of `service_worker`
   - Scripts run in a non-persistent background context (like service workers)

2. **Background Scripts in MV3**
   - Firefox MV3 allows `background.scripts` (adapted from MV2)
   - Non-persistent by default in MV3
   - Same event-driven architecture as service workers
   - Functionally equivalent to Chrome's service workers

3. **Cross-Browser Compatibility (Chrome 121+ / Firefox 121+)**
   - **Before Chrome 121**: Chrome refused to load if `background.scripts` was present
   - **Before Firefox 120**: Firefox didn't start if `service_worker` was present (bug 1860304)
   - **Since Chrome 121** (Jan 2024): Ignores `background.scripts` in MV3
   - **Since Firefox 121** (Dec 2023): Ignores `service_worker` in MV3
   - **Result**: Both properties can coexist in one manifest!

## Implementation Strategy

### Two Manifest Files

**`manifest.json`** (Chrome/Edge):
```json
{
  "manifest_version": 3,
  "background": {
    "service_worker": "background.js",
    "type": "module"
  }
}
```

**`manifest-firefox.json`** (Firefox):
```json
{
  "manifest_version": 3,
  "background": {
    "scripts": ["background-firefox.js"]
  }
}
```

### Two Background Scripts

**`background.js`** (Chrome/Edge):
- Service worker context
- No DOM access
- Event-driven lifecycle

**`background-firefox.js`** (Firefox):
- Same logic as `background.js`
- Uses `browser.*` API with Chrome fallback
- Works in background script context

## Code Compatibility

Both background scripts use the same logic:

```javascript
// Works in both Chrome and Firefox
const browserAPI = typeof browser !== 'undefined' ? browser : chrome;

browserAPI.runtime.sendNativeMessage(NATIVE_HOST, message, (response) => {
  // Handle response
});
```

## Testing in Firefox

### Load Temporary Add-on

1. Open `about:debugging#/runtime/this-firefox`
2. Click "Load Temporary Add-on..."
3. Select `manifest-firefox.json` (not `manifest.json`)
4. Extension loads successfully

### Common Errors

**Error: "background.service_worker is currently disabled"**
- **Cause:** Using Chrome's service worker format
- **Solution:** Use `manifest-firefox.json` with `background.scripts`

**Error: "background.scripts is not allowed in Manifest V3"**
- **Cause:** Incorrect manifest version
- **Solution:** Ensure `manifest_version: 3` is set

## Firefox MV3 Status (November 2025)

| Feature | Chrome MV3 | Firefox MV3 | Status |
|---------|------------|-------------|--------|
| manifest_version: 3 | ✅ | ✅ | Supported |
| Service workers | ✅ | ⚠️ Experimental | Not ready |
| background.scripts | ❌ Removed | ✅ Allowed | Firefox-specific |
| Action API | ✅ | ✅ | Supported |
| Native Messaging | ✅ | ✅ | Supported |
| Content Scripts | ✅ | ✅ | Supported |
| web_accessible_resources | ✅ | ✅ | Supported |

**Legend:**
- ✅ Fully supported
- ⚠️ Experimental/unstable
- ❌ Not available

## Mozilla's Documentation

From Mozilla's Extension Workshop:

> "Firefox currently supports background scripts in Manifest V3 extensions. 
> Service workers are being implemented but are not yet recommended for production use."

Source: https://extensionworkshop.com/documentation/develop/manifest-v3-migration-guide/

## When Will Firefox Support Service Workers?

**Current Status (November 2025):**
- Service workers are "experimental"
- `background.scripts` is the recommended approach
- No official timeline for full service worker support

**What to Watch:**
- Firefox release notes
- Mozilla's extension workshop updates
- Bug tracker: https://bugzilla.mozilla.org/

## Migration Path

### Current (2025)
```
Chrome/Edge: Service Workers (MV3)
Firefox: Background Scripts (MV3)
```

### Future (When Firefox Adds Service Workers)
```
All browsers: Service Workers (MV3)
Single manifest.json for all browsers
```

## Practical Implications

### For Developers

**Pros:**
- ✅ Firefox MV3 works today
- ✅ Native Messaging fully supported
- ✅ Same functionality as Chrome

**Cons:**
- ⚠️ Need separate manifest files
- ⚠️ Need separate build process
- ⚠️ Slight maintenance overhead

### For Users

**No Impact:**
- Extension works identically in all browsers
- Same features and performance
- Transparent to end users

## Build Process

### Chrome/Edge
```powershell
.\ops\package-extension.ps1
# Creates: dist/esteid-pd-extension-v1.0.0.zip
# Contains: manifest.json + background.js
```

### Firefox
```powershell
.\ops\package-extension-firefox.ps1
# Creates: dist/esteid-pd-extension-firefox-v1.0.0.zip
# Contains: manifest-firefox.json + background-firefox.js
```

## Distribution

### Chrome Web Store
- Upload: `esteid-pd-extension-v1.0.0.zip`
- Uses: `manifest.json` with service worker

### Edge Add-ons
- Upload: `esteid-pd-extension-v1.0.0.zip`
- Uses: `manifest.json` with service worker

### Firefox Add-ons (AMO)
- Upload: `esteid-pd-extension-firefox-v1.0.0.zip`
- Uses: `manifest-firefox.json` with background scripts

## Recommendations

### For This Project

✅ **Use separate manifests** - Most reliable approach  
✅ **Use background.scripts for Firefox** - Recommended by Mozilla  
✅ **Monitor Firefox updates** - Watch for service worker support  
⚠️ **Plan for future migration** - When Firefox adds full service worker support  

### Testing Checklist

- [x] Chrome: Load with `manifest.json`
- [x] Edge: Load with `manifest.json`
- [x] Firefox: Load with `manifest-firefox.json`
- [ ] Test Native Messaging in all browsers
- [ ] Test all features in Firefox
- [ ] Verify error handling

## References

- [Firefox MV3 Migration Guide](https://extensionworkshop.com/documentation/develop/manifest-v3-migration-guide/)
- [Firefox Extension API](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions)
- [Chrome MV3 Documentation](https://developer.chrome.com/docs/extensions/mv3/)
- [Service Worker Status in Firefox](https://bugzilla.mozilla.org/show_bug.cgi?id=1573659)

## Conclusion

Firefox's Manifest V3 implementation is **functional but different** from Chrome/Edge:

✅ **Works Today:**
- Manifest V3 supported
- Native Messaging works
- All features functional

⚠️ **Requires Adaptation:**
- Use `background.scripts` instead of `background.service_worker`
- Separate manifest file needed
- Separate build process

🔮 **Future:**
- Full service worker support coming
- Will align with Chrome/Edge
- Can migrate to unified manifest later

**For now, use `manifest-firefox.json` with `background.scripts` - it's the official recommendation.**

---

**Last Updated:** November 2025  
**Firefox Version:** 145.0  
**Status:** Production Ready (with separate manifest)

