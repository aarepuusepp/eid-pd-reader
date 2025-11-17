# Security Notes — WebEID Personal Data Fallback

## Threat Model Highlights

### Attack Surfaces

1. **Native Messaging Channel**: Extension ↔ Native Helper (stdio)
2. **Content Script Bridge**: Page ↔ Extension (postMessage)
3. **PC/SC Interface**: Native Helper ↔ Smart Card (APDU)
4. **Host Manifest Registration**: OS-level configuration

### Threats & Mitigations

#### 1. Origin Pinning

**Threat**: Malicious website impersonates legitimate app to read PD.

**Mitigation**:
- `extension/manifest.json` restricts `externally_connectable.matches` to specific origin(s)
- `host_permissions` limited to same origin(s)
- Content script validates `event.source === window` (same-origin check)
- Host manifest `allowed_origins` contains only legitimate extension ID

**Action Items**:
- Replace `YOUR-APP-ORIGIN` with actual HTTPS origin(s)
- Use specific paths if needed (e.g., `https://app.example.com/secure/*`)
- Do not use wildcards (`*`) in production

#### 2. Native Host Impersonation

**Threat**: Malicious binary registers as `ee.aarepuusepp.eid_helper` and intercepts PD.

**Mitigation**:
- Host manifest path must be absolute and point to trusted binary
- Registry/file permissions restrict who can modify host manifests
- Users must explicitly install/approve host registration
- Consider code signing for production binaries

**Action Items**:
- Use absolute paths in host manifests
- Document installation process clearly
- Consider signing `eid-pd-reader.exe` with code signing certificate

#### 3. Data Exposure

**Threat**: PD data leaked to unauthorized parties or persisted.

**Mitigation**:
- No persistent storage in extension (no `chrome.storage` usage)
- PD returned only to page that initiated request
- No logging of PD contents (only status codes)
- Native helper exits after each request (no state retention)

**Action Items**:
- Audit extension code for storage APIs
- Verify native helper does not log PD fields
- Ensure no telemetry includes PII

#### 4. APDU Injection

**Threat**: Page injects malicious APDUs to manipulate card.

**Mitigation**:
- Page never sees raw APDUs
- Native helper only accepts predefined commands (`readPersonalData`)
- FID overrides limited to hex strings (validated)
- Unknown commands rejected

**Action Items**:
- Keep command set minimal
- Validate FID format (4 hex chars)
- Do not expose `transmit()` or raw APDU building to page

#### 5. Man-in-the-Middle (Native Messaging)

**Threat**: Intercept/modify messages between extension and helper.

**Mitigation**:
- Native Messaging uses stdio (local only, no network)
- Messages are JSON (no binary blobs that could hide exploits)
- Length prefix prevents buffer overflows (max 1MB)

**Action Items**:
- Keep message size limits reasonable
- Validate JSON structure before processing
- Consider message signing for production (if needed)

## Logging Policy

### What to Log

- **Status only**: Success/failure, error codes (SW classes), operation duration
- **Anonymized metrics**: Field lengths (not contents), card presence, reader count
- **Diagnostics**: PC/SC errors, connection failures (no card data)

### What NOT to Log

- **Personal Data**: Never log `surname`, `givenName`, `personalCode` values
- **Raw APDUs**: Avoid logging full command/response bytes (use summaries)
- **Card serials**: Do not log ATR or card identifiers
- **User context**: No IP addresses, user agents, or session IDs

### Example Logging

```go
// GOOD: Log status only
log.Printf("PD read: success, surname_len=%d", len(pd.Surname))

// BAD: Log actual data
log.Printf("PD read: surname=%s", pd.Surname) // NEVER DO THIS
```

## Permissions Minimization

### Extension Permissions

- `nativeMessaging`: Required for Native Messaging host communication
- `scripting`: Required for dynamic content script injection
- `host_permissions`: Limited to specific origin(s) only

### Native Helper Permissions

- **Windows**: No special permissions required (uses PC/SC service)
- **Linux**: May need `pcscd` access (usually via group membership)
- **macOS**: Uses system PCSC framework (no special permissions)

## Code Signing (Production)

For production deployments:

1. **Sign native binary**:
   - Windows: Use Authenticode signing
   - Linux/macOS: Consider GPG signing or package signing

2. **Sign extension**:
   - Chrome Web Store: Automatic signing
   - Firefox Add-ons: Automatic signing
   - Self-distributed: Consider extension signing

3. **Verify host manifest**:
   - Ensure path points to signed binary
   - Consider manifest integrity checks

## NFC Considerations (Future)

When implementing NFC backend:

- **CAN handling**: Prompt in native dialog (not web page)
- **PACE keys**: Never expose to page or extension
- **Secure Messaging**: Keep SM keys isolated in native helper
- **NFC reader access**: May require additional OS permissions

## Incident Response

If security issue discovered:

1. **Immediate**: Disable feature flag (`PD_FALLBACK_ENABLED = false`)
2. **Investigation**: Review logs (status only, no PII)
3. **Mitigation**: Update extension/host manifests, revoke compromised IDs
4. **Communication**: Notify users if data exposure confirmed

## Compliance Notes

- **GDPR**: PD is personal data; ensure lawful basis for processing
- **Estonian eID**: Follow EstEID middleware security guidelines
- **Browser policies**: Comply with Chrome/Firefox extension policies

## Security Checklist

- [ ] Origin pinning configured in `manifest.json`
- [ ] Host manifest `allowed_origins` contains actual extension ID
- [ ] No PD data logged or persisted
- [ ] Native helper validates all inputs
- [ ] Extension does not use `chrome.storage` for PD
- [ ] Error messages do not leak sensitive information
- [ ] Code signing implemented (production)
- [ ] Security review completed before deployment

