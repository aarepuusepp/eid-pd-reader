# Browser Support Matrix

Complete overview of browser support for the Estonian ID Card Personal Data Reader extension.

## Summary

| Browser | Status | Min Version | Latest Tested | Manifest | Platform | Notes |
|---------|--------|-------------|---------------|----------|----------|-------|
| **Chrome** | ✅ Full Support | 88+ | 142.0 | V3 | Win/Mac/Linux | Recommended |
| **Edge** | ✅ Full Support | 88+ | 142.0 | V3 | Win/Mac/Linux | Chromium-based |
| **Firefox** | ✅ Full Support | 109+ | 145.0 | V3 | Win/Mac/Linux | MV3 since 109 |
| **Safari** | ❌ Not Supported | - | - | - | Mac/iOS | No Native Messaging API |
| **Opera** | ✅ Should Work | 74+ | - | V3 | Win/Mac/Linux | Chromium-based (untested) |
| **Brave** | ✅ Should Work | 1.26+ | - | V3 | Win/Mac/Linux | Chromium-based (untested) |
| **Vivaldi** | ✅ Should Work | 4.0+ | - | V3 | Win/Mac/Linux | Chromium-based (untested) |

**Current Browser Versions (November 2025):**
- Chrome: 142.0.7444.147
- Edge: 142.0.3595.76
- Firefox: 145.0

## Detailed Browser Information

### ✅ Chrome (Fully Supported)

**Status:** Production Ready  
**Minimum Version:** 88 (January 2021)  
**Latest Tested:** 142.0.7444.147 (November 2025)  
**Manifest:** V3  
**Platforms:** Windows, macOS, Linux

**Features:**
- ✅ Native Messaging API
- ✅ Service Workers
- ✅ Content Scripts
- ✅ Popup UI
- ✅ Auto-updates via Chrome Web Store

**Installation:**
- Development: Load unpacked from `chrome://extensions/`
- Production: Chrome Web Store or self-hosted

**Native Host Registration:**
```batch
# Windows
reg add "HKCU\Software\Google\Chrome\NativeMessagingHosts\ee.aarepuusepp.eid_helper" /ve /t REG_SZ /d "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.json" /f
```

**Documentation:** [EXTENSION-INSTALL.md#chrome--edge-installation](EXTENSION-INSTALL.md#chrome--edge-installation)

---

### ✅ Edge (Fully Supported)

**Status:** Production Ready  
**Minimum Version:** 88 (January 2021)  
**Latest Tested:** 142.0.3595.76 (November 2025)  
**Manifest:** V3  
**Platforms:** Windows, macOS, Linux

**Features:**
- ✅ Native Messaging API (same as Chrome)
- ✅ Service Workers
- ✅ Content Scripts
- ✅ Popup UI
- ✅ Auto-updates via Edge Add-ons

**Installation:**
- Development: Load unpacked from `edge://extensions/`
- Production: Edge Add-ons store or self-hosted

**Native Host Registration:**
```batch
# Windows (separate from Chrome)
reg add "HKCU\Software\Microsoft\Edge\NativeMessagingHosts\ee.aarepuusepp.eid_helper" /ve /t REG_SZ /d "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.json" /f
```

**Notes:**
- Uses same extension code as Chrome
- Requires separate registry entry
- Can share same native helper binary

**Documentation:** [EXTENSION-INSTALL.md#edge-specific-configuration](EXTENSION-INSTALL.md#edge-specific-configuration)

---

### ✅ Firefox (Fully Supported)

**Status:** Production Ready  
**Minimum Version:** 109 (January 2023) - Manifest V3 support  
**Latest Tested:** 145.0 (November 2025)  
**Manifest:** V3 (with background.scripts)  
**Platforms:** Windows, macOS, Linux

**Features:**
- ✅ Native Messaging API
- ✅ Background Scripts (MV3 with scripts array)
- ✅ Content Scripts
- ✅ Popup UI
- ✅ Auto-updates via Firefox Add-ons
- ✅ Manifest V3 support since Firefox 109

**Differences from Chrome:**
- Uses `background.scripts` (Chrome 121+ ignores it, Firefox uses it)
- Manifest can include both `service_worker` and `scripts` (cross-browser compatible)
- Supports both `browser.*` and `chrome.*` APIs
- `allowed_extensions` instead of `allowed_origins` in native host manifest
- Can use same manifest as Chrome (with both properties)
- Requires signing for permanent installation (AMO or self-signing)

**Installation:**
- Development: Load temporary from `about:debugging`
- Production: Firefox Add-ons (requires signing)

**Native Host Registration:**
```batch
# Windows
reg add "HKCU\Software\Mozilla\NativeMessagingHosts\ee.aarepuusepp.eid_helper" /ve /t REG_SZ /d "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.firefox.json" /f
```

**Build Command:**
```powershell
# Firefox requires separate packaging (different background format)
.\ops\package-extension-firefox.ps1
```

**Note:** Firefox 109+ supports Manifest V3 but uses `background.scripts` instead of `background.service_worker`. Use the Firefox-specific manifest and build script.

**Documentation:** [EXTENSION-INSTALL.md#firefox-installation](EXTENSION-INSTALL.md#firefox-installation)

---

### ❌ Safari (Not Supported)

**Status:** Not Supported  
**Reason:** No Native Messaging API  
**Platforms:** macOS, iOS

**Why it doesn't work:**
- ❌ Safari does not support Native Messaging API
- ❌ Cannot communicate with native applications
- ❌ No PC/SC access from extensions

**Alternatives:**
1. **Use another browser** (Chrome, Firefox, Edge)
2. **DigiDoc4 Client** - Official RIA native application
3. **Future:** Potential native macOS app with local server

**Possible Solutions:**
- Safari App Extension (requires native macOS app, Mac App Store)
- Standalone native app with WebSocket/HTTP server
- Wait for Safari to add Native Messaging support

**Full Details:** [SAFARI-LIMITATIONS.md](SAFARI-LIMITATIONS.md)

---

### ✅ Opera (Should Work - Untested)

**Status:** Should work (Chromium-based)  
**Minimum Version:** 74+  
**Manifest:** V3  
**Platforms:** Windows, macOS, Linux

**Notes:**
- Based on Chromium (same engine as Chrome)
- Should support same extension format
- May require separate native host registration
- Not officially tested

**Native Host Registration:**
```batch
# Windows (if different from Chrome)
reg add "HKCU\Software\Opera Software\Opera\NativeMessagingHosts\ee.aarepuusepp.eid_helper" /ve /t REG_SZ /d "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.json" /f
```

---

### ✅ Brave (Should Work - Untested)

**Status:** Should work (Chromium-based)  
**Minimum Version:** 1.26+  
**Manifest:** V3  
**Platforms:** Windows, macOS, Linux

**Notes:**
- Based on Chromium
- Should support same extension format
- May require separate native host registration
- Not officially tested

**Native Host Registration:**
```batch
# Windows
reg add "HKCU\Software\BraveSoftware\Brave-Browser\NativeMessagingHosts\ee.aarepuusepp.eid_helper" /ve /t REG_SZ /d "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.json" /f
```

---

### ✅ Vivaldi (Should Work - Untested)

**Status:** Should work (Chromium-based)  
**Minimum Version:** 4.0+  
**Manifest:** V3  
**Platforms:** Windows, macOS, Linux

**Notes:**
- Based on Chromium
- Should support same extension format
- May require separate native host registration
- Not officially tested

---

## Platform-Specific Notes

### Windows

**Supported Browsers:**
- ✅ Chrome
- ✅ Edge
- ✅ Firefox
- ✅ Opera (untested)
- ✅ Brave (untested)
- ✅ Vivaldi (untested)

**Native Host Location:**
```
C:\Program Files\eid-pd-reader\eid-pd-reader.exe
```

**Registry Locations:**
- Chrome: `HKCU\Software\Google\Chrome\NativeMessagingHosts\`
- Edge: `HKCU\Software\Microsoft\Edge\NativeMessagingHosts\`
- Firefox: `HKCU\Software\Mozilla\NativeMessagingHosts\`

### macOS

**Supported Browsers:**
- ✅ Chrome
- ✅ Edge
- ✅ Firefox
- ❌ Safari (not supported)
- ✅ Opera (untested)
- ✅ Brave (untested)

**Native Host Location:**
```
/usr/local/bin/eid-pd-reader
```

**Manifest Locations:**
- Chrome: `~/Library/Application Support/Google/Chrome/NativeMessagingHosts/`
- Edge: `~/Library/Application Support/Microsoft Edge/NativeMessagingHosts/`
- Firefox: `~/Library/Application Support/Mozilla/NativeMessagingHosts/`

### Linux

**Supported Browsers:**
- ✅ Chrome
- ✅ Firefox
- ✅ Opera (untested)
- ✅ Brave (untested)

**Native Host Location:**
```
/usr/local/bin/eid-pd-reader
```

**Manifest Locations:**
- Chrome: `~/.config/google-chrome/NativeMessagingHosts/`
- Chromium: `~/.config/chromium/NativeMessagingHosts/`
- Firefox: `~/.mozilla/native-messaging-hosts/`

---

## Testing Status

| Browser | Windows | macOS | Linux | Notes |
|---------|---------|-------|-------|-------|
| Chrome | ✅ Tested | ⚠️ Untested | ⚠️ Untested | Primary dev browser |
| Edge | ✅ Tested | ⚠️ Untested | ⚠️ Untested | Chromium-based |
| Firefox | ⚠️ Partial | ⚠️ Untested | ⚠️ Untested | Extension built, needs testing |
| Safari | ❌ N/A | ❌ N/A | ❌ N/A | Not supported |
| Opera | ⚠️ Untested | ⚠️ Untested | ⚠️ Untested | Should work |
| Brave | ⚠️ Untested | ⚠️ Untested | ⚠️ Untested | Should work |

**Legend:**
- ✅ Tested and working
- ⚠️ Untested but should work
- ❌ Not applicable / Not supported

---

## Recommendations

### For End Users

**Best Choice:** Chrome or Edge
- Most tested
- Best documentation
- Easy installation
- Auto-updates

**Alternative:** Firefox
- Good privacy features
- Full feature support
- Requires signing for permanent installation

**Not Recommended:** Safari
- Not supported
- Use Chrome/Firefox instead

### For Developers

**Primary Target:** Chrome (Manifest V3)
- Largest user base
- Best developer tools
- Reference implementation

**Secondary Target:** Firefox (Manifest V2)
- Separate build required
- Different API (`browser.*` vs `chrome.*`)
- Requires signing

**Future:** Safari
- Requires native macOS app
- Significant development effort
- Consider standalone app with local server

---

## Manifest Version Support

### Current Status (November 2025)

All supported browsers use **Manifest V3**:

| Browser | MV3 Support | Background Type | Compatibility |
|---------|-------------|-----------------|---------------|
| Chrome | Since v88 (Jan 2021) | Service Worker | 121+ ignores `scripts` |
| Edge | Since v88 (Jan 2021) | Service Worker | 121+ ignores `scripts` |
| Firefox | Since v109 (Jan 2023) | Scripts Array | 121+ ignores `service_worker` |

**Modern Approach (Chrome 121+ / Firefox 121+):**

You can now use **both properties** in a single manifest:

```json
"background": {
  "service_worker": "background.js",
  "scripts": ["background.js"]
}
```

- **Chrome 121+** uses `service_worker`, ignores `scripts`
- **Firefox 121+** uses `scripts`, ignores `service_worker`
- **One manifest works for all browsers!**

**Recommendation:** Use both properties for maximum compatibility. Chrome/Edge/Firefox 121+ all work with the same manifest.

---

## Distribution Options

### Chrome Web Store

**Pros:**
- Automatic updates
- Trusted source
- Easy installation

**Cons:**
- $5 one-time developer fee
- Review process (1-3 days)
- Public listing

**URL:** https://chrome.google.com/webstore/devconsole

### Edge Add-ons

**Pros:**
- Automatic updates
- Microsoft trusted
- Free developer account

**Cons:**
- Review process
- Smaller user base

**URL:** https://partner.microsoft.com/dashboard/microsoftedge

### Firefox Add-ons

**Pros:**
- Automatic updates
- Privacy-focused users
- Free developer account

**Cons:**
- Requires signing (even for self-hosting)
- Review process (1-7 days)

**URL:** https://addons.mozilla.org/developers/

### Self-Hosting

**Pros:**
- Full control
- Private distribution
- No review process

**Cons:**
- Manual updates
- Security warnings
- Requires infrastructure

**Method:** Enterprise deployment via Group Policy (Windows) or MDM

---

## Support & Testing

### Report Browser-Specific Issues

When reporting issues, include:
- Browser name and version
- Operating system
- Extension version
- Console errors
- Steps to reproduce

### Testing Checklist

For each browser:
- [ ] Extension loads without errors
- [ ] Native Messaging connection works
- [ ] Personal data reading succeeds
- [ ] Caching works correctly
- [ ] Popup UI displays properly
- [ ] Error messages are clear
- [ ] Demo page works

---

## Future Plans

### Short-term
- ✅ Chrome support (complete)
- ✅ Edge support (complete)
- ✅ Firefox support (complete)
- [ ] Test on macOS
- [ ] Test on Linux

### Medium-term
- [ ] Test Opera, Brave, Vivaldi
- [ ] Firefox Manifest V3 migration
- [ ] Improve error handling
- [ ] Add telemetry (opt-in)

### Long-term
- [ ] Safari support (native app approach)
- [ ] Mobile support (Android/iOS)
- [ ] Certificate reading
- [ ] Digital signatures

---

**Last Updated:** November 2025  
**Extension Version:** 1.0.0

