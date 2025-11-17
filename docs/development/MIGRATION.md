# Migration Guide — Integrating WebEID PD Fallback

This guide explains how to integrate the WebEID Personal Data fallback into an existing web application that uses `web-eid.js` for certificate-based identity verification.

## Overview

The fallback adds a client-side enhancement that automatically retrieves Personal Data directly from the Estonian ID-card chip when certificates are missing or the certificate-based flow fails. **No server-side API changes are required.**

## Integration Steps

### 1. Install Native Helper

**Windows:**
```powershell
# Build helper
.\ops\build-all.ps1

# Install host manifest
.\ops\install-host.ps1
```

**Linux/macOS:**
```bash
# Build helper
./ops/build-all.sh

# Install host manifest
./ops/install-host.sh
```

### 2. Install Browser Extension

1. Load unpacked extension from `extension/` directory in Chrome/Edge
2. Note the extension ID (chrome://extensions/)
3. Update `host-manifests/ee.aarepuusepp.eid_helper.json`:
   - Replace `__EXTENSION_ID__` with actual extension ID
4. Update `extension/manifest.json`:
   - Replace `YOUR-APP-ORIGIN` with your app's HTTPS origin
5. Reload extension

### 3. Copy Web App Files

Copy these files to your web application:

- `extension/page-api.js` → Your web app (e.g., `/static/js/page-api.js`)
- `webapp/config/pd-fallback.config.js` → Your web app (e.g., `/static/js/config/pd-fallback.config.js`)
- `webapp/page-api-auto.js` → Your web app (e.g., `/static/js/page-api-auto.js`)

### 4. Wire Web-EID Adapter

Edit `webapp/page-api-auto.js` and implement `getPersonFromCert()`:

```javascript
// Replace the stub with your actual web-eid.js integration
async function getPersonFromCert() {
  try {
    // Example: Get certificate via web-eid.js
    const cert = await webEid.getCertificate();
    
    // Parse certificate subject to extract PD
    const subject = parseCertificateSubject(cert);
    
    // Return in expected format
    return {
      ok: true,
      data: {
        givenName: subject.givenName,
        surname: subject.surname,
        personalCode: subject.serialNumber // or appropriate field
      }
    };
  } catch (error) {
    // Return failure (will trigger fallback)
    return {
      ok: false,
      error: error.message
    };
  }
}
```

**Important**: Your adapter should:
- Return `{ok:true, data:{givenName, surname, personalCode}}` on success
- Return `{ok:false, error:"..."}` or throw on failure
- Complete within `PD_FALLBACK_TIMEOUT_MS` (default 5s) or fallback will trigger

### 5. Update Your UI

Replace existing web-eid.js calls with `getIdentityAuto()`:

**Before:**
```javascript
async function identifyUser() {
  const cert = await webEid.getCertificate();
  const subject = parseCertificateSubject(cert);
  // Use subject data...
}
```

**After:**
```javascript
import { getIdentityAuto } from '/static/js/page-api-auto.js';

async function identifyUser() {
  const res = await getIdentityAuto();
  if (!res?.ok) {
    alert('Identity verification failed: ' + (res?.error || 'unknown error'));
    return;
  }
  const { givenName, surname, personalCode, source } = res.data;
  
  // source will be either:
  // - "cert" (from web-eid.js)
  // - "fallback:chip" (from native helper)
  
  // Use data...
}
```

### 6. Configure Feature Flag

Edit `webapp/config/pd-fallback.config.js`:

```javascript
export const PD_FALLBACK_ENABLED = true;  // Enable fallback
export const PD_FALLBACK_TIMEOUT_MS = 5000; // 5s timeout
export const PD_FIELDS = {
  df_personal: "5000",
  ef_surname:  "5001",
  ef_given:    "5002",
  ef_code:     "5006"
};
```

**Rollout Strategy:**
- Start with `PD_FALLBACK_ENABLED = false` (cert-only)
- Test with small user group
- Enable gradually (`PD_FALLBACK_ENABLED = true`)
- Monitor telemetry

## API Compatibility

### Request/Response Shape

The fallback maintains compatibility with your existing API:

**Input**: None (reads from card directly)

**Output**: Consistent shape
```typescript
{
  ok: boolean;
  data?: {
    givenName: string;
    surname: string;
    personalCode: string;
    source: "cert" | "fallback:chip";
  };
  error?: string;
}
```

### No Server Changes Required

- All processing is client-side
- No new API endpoints needed
- No database schema changes
- Existing authentication/authorization logic unchanged

## Testing Checklist

- [ ] Certificate path works (when certs present)
- [ ] Fallback path works (when certs missing)
- [ ] Timeout handling (cert path times out → fallback)
- [ ] Error handling (no reader, no card)
- [ ] UI shows appropriate messages
- [ ] Source indicator (`source` field) tracked correctly

## Rollback Plan

If issues arise:

1. **Immediate**: Set `PD_FALLBACK_ENABLED = false` in config
2. **Deploy**: Push config change (no code changes needed)
3. **Verify**: System falls back to cert-only mode
4. **Investigate**: Review logs and telemetry

## User Experience

### When Certificates Present

- User clicks "Identify" button
- `getIdentityAuto()` calls `getPersonFromCert()`
- Certificate read succeeds → returns PD immediately
- Fallback not triggered
- User sees normal flow

### When Certificates Missing

- User clicks "Identify" button
- `getIdentityAuto()` calls `getPersonFromCert()`
- Certificate read fails or times out
- Fallback triggers → native helper reads from chip
- PD returned from chip
- User sees same result (may take slightly longer)

### Error Cases

- **No reader**: Clear message "Card reader not found"
- **No card**: Clear message "Please insert your ID-card"
- **Card busy**: Retry message or timeout
- **Extension not installed**: Graceful degradation (cert-only)

## Telemetry Integration

Track fallback usage (anonymized):

```javascript
async function identifyUser() {
  const res = await getIdentityAuto();
  
  // Log source (for analytics)
  if (res?.ok) {
    trackEvent('identity_verified', {
      source: res.data.source, // "cert" or "fallback:chip"
      duration: res.duration // if tracked
    });
  } else {
    trackEvent('identity_failed', {
      error_type: res?.error?.substring(0, 50) // truncated, no PII
    });
  }
}
```

**Important**: Do not log PD values (surname, givenName, personalCode).

## Troubleshooting

### Extension Not Working

- Verify extension is installed and enabled
- Check extension ID matches host manifest `allowed_origins`
- Verify origin matches `externally_connectable.matches`
- Check browser console for errors

### Native Helper Not Responding

- Verify host manifest is registered (Windows: check registry)
- Verify `eid-pd-reader.exe` path is absolute and correct
- Test CLI mode: `.\eid-pd-reader.exe --dump`
- Check PC/SC service is running

### Fallback Not Triggering

- Verify `PD_FALLBACK_ENABLED = true`
- Check `getPersonFromCert()` throws or returns `{ok:false}`
- Verify timeout is sufficient (`PD_FALLBACK_TIMEOUT_MS`)
- Check browser console for errors

## Support

For issues:
1. Check `ops/QA-checklist.md` for common problems
2. Review browser console and extension logs
3. Test native helper directly (`--dump` mode)
4. Verify PC/SC service and card reader drivers

