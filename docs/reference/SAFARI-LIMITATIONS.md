# Safari Extension Limitations

## Overview

Safari browser has significant limitations for implementing the Estonian ID Card Personal Data Reader extension due to Apple's restrictions on native messaging and extension capabilities.

## Current Status: ❌ Not Supported

Safari extensions **cannot** currently access native applications via Native Messaging API, which is the core requirement for this extension.

## Technical Limitations

### 1. No Native Messaging API

Safari does not support the Chrome/Firefox Native Messaging API:
- ❌ No `browser.runtime.sendNativeMessage()`
- ❌ No communication with native applications
- ❌ No PC/SC access from extensions

**Impact**: Cannot communicate with `eid-pd-reader.exe` to read ID cards.

### 2. Safari App Extensions (macOS only)

Safari 14+ introduced "Safari App Extensions" which are part of a native macOS app:

**Pros:**
- ✅ Can include native code in the app
- ✅ Can access system APIs (including PC/SC)

**Cons:**
- ❌ Requires complete rewrite as a native macOS app
- ❌ Must be distributed via Mac App Store ($99/year)
- ❌ Requires Swift/Objective-C development
- ❌ macOS only (no Windows/Linux)
- ❌ Complex development and deployment

### 3. Safari Web Extensions (iOS/macOS)

Safari 15+ supports Web Extensions API (similar to Chrome):

**Pros:**
- ✅ Cross-platform (iOS + macOS)
- ✅ Similar to Chrome/Firefox extensions

**Cons:**
- ❌ Still no Native Messaging API
- ❌ Cannot communicate with native apps
- ❌ Limited to web APIs only

## Alternative Approaches for Safari

### Option 1: Safari App Extension (macOS only)

**Architecture:**
```
Safari Extension (JavaScript)
    ↓ XPC (Inter-Process Communication)
Native macOS App (Swift/Objective-C)
    ↓ PC/SC Framework
ID Card Reader
```

**Requirements:**
- Rewrite as native macOS application
- Implement PC/SC communication in Swift/Objective-C
- Package as .app bundle with embedded extension
- Sign with Apple Developer certificate
- Distribute via Mac App Store or notarization

**Estimated Effort:** 2-4 weeks of development

### Option 2: Standalone Native App (Recommended)

Instead of a browser extension, create a native macOS application:

**Architecture:**
```
Web Page
    ↓ WebSocket or HTTP
Native macOS App (Local Server)
    ↓ PC/SC Framework
ID Card Reader
```

**Pros:**
- ✅ Works in any browser (Safari, Chrome, Firefox)
- ✅ Simpler development
- ✅ No browser extension needed
- ✅ Can use existing Go code (via CGO)

**Cons:**
- ⚠️ Requires local server running
- ⚠️ Need to handle CORS/security
- ⚠️ User must start the app

### Option 3: Web Crypto API (Certificate-based only)

Safari supports Web Crypto API for certificate operations:

**Pros:**
- ✅ Built-in browser support
- ✅ No extension needed

**Cons:**
- ❌ Only works with certificates (not chip reading)
- ❌ Cannot read personal data fields directly
- ❌ Requires valid certificates

## Comparison with Other Browsers

| Feature | Chrome | Firefox | Edge | Safari |
|---------|--------|---------|------|--------|
| Native Messaging | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No |
| Web Extensions | ✅ MV3 | ✅ MV2/3 | ✅ MV3 | ⚠️ Limited |
| PC/SC Access | ✅ Via Native | ✅ Via Native | ✅ Via Native | ❌ No |
| Cross-Platform | ✅ Win/Mac/Linux | ✅ Win/Mac/Linux | ✅ Win/Mac | ⚠️ Mac only |
| Distribution | ✅ Web Store | ✅ Add-ons | ✅ Web Store | ⚠️ App Store |

## Recommendations

### For Safari Users

**Option A: Use Chrome/Firefox/Edge**
- Recommended for best experience
- Full feature support
- Easy installation

**Option B: Use DigiDoc4 Client**
- Official RIA application
- Native macOS support
- Full ID card functionality

**Option C: Wait for Native App**
- We may develop a standalone macOS app in the future
- Would work with Safari and all browsers
- Subscribe to project updates

### For Developers

If you need Safari support:

1. **Short-term**: Direct users to Chrome/Firefox/Edge
2. **Medium-term**: Develop standalone native macOS app with local server
3. **Long-term**: Consider Safari App Extension if Mac App Store distribution is acceptable

## Technical Details for Safari App Extension

If you decide to implement Safari App Extension, here's what's needed:

### 1. Create Xcode Project

```swift
// Swift code for PC/SC communication
import CryptoTokenKit

class CardReader {
    func readPersonalData() -> PersonalData? {
        // Connect to card reader
        // Send APDU commands
        // Parse responses
        // Return data
    }
}
```

### 2. Implement XPC Service

```swift
// Communication between extension and native app
@objc protocol CardReaderProtocol {
    func readCard(reply: @escaping (PersonalData?, Error?) -> Void)
}
```

### 3. Safari Extension JavaScript

```javascript
// Extension calls native code
browser.runtime.sendMessage({
    type: "readCard"
}, (response) => {
    // Handle response
});
```

### 4. Build and Sign

```bash
# Build the app
xcodebuild -project EstEID.xcodeproj -scheme EstEID

# Sign with Developer ID
codesign --deep --force --verify --verbose \
  --sign "Developer ID Application: Your Name" \
  EstEID.app

# Notarize for distribution
xcrun altool --notarize-app --file EstEID.zip
```

## Resources

### Apple Documentation
- [Safari App Extensions](https://developer.apple.com/documentation/safariservices/safari_app_extensions)
- [Safari Web Extensions](https://developer.apple.com/documentation/safariservices/safari_web_extensions)
- [CryptoTokenKit Framework](https://developer.apple.com/documentation/cryptotokenkit)

### Alternative Solutions
- [DigiDoc4 Client](https://github.com/open-eid/DigiDoc4-Client) - Official RIA app with macOS support
- [Web eID](https://github.com/web-eid) - Certificate-based authentication (works in Safari)

## Conclusion

**Safari support is not currently feasible** for this extension due to lack of Native Messaging API.

**Recommended approach:**
1. Use Chrome/Firefox/Edge for full functionality
2. Consider developing a standalone native macOS app for Safari users
3. Monitor Safari Web Extensions API for future Native Messaging support

## Questions?

If you have questions about Safari support or need help implementing alternatives:

1. Check [Apple Developer Forums](https://developer.apple.com/forums/)
2. Review [Web eID project](https://github.com/web-eid) for certificate-based approach
3. Open an issue on our GitHub repository

---

**Last Updated:** November 2025  
**Status:** Safari not supported, alternatives available

