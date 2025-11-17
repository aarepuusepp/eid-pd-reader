# Troubleshooting Guide — WebEID Personal Data Fallback

## Common Issues

### SW=6A86 (Wrong P1P2) when selecting DF 5000

**Symptom:**
```
{"ok":false,"error":"select DF 0x5000: card returned error SW=6A86"}
```

**Possible Causes:**
- Card profile doesn't use DF 5000
- SELECT command parameters incorrect for this card type
- Card requires different SELECT method

**Solutions:**
1. The code now tries multiple SELECT methods automatically
2. If all fail, it attempts to read EFs directly from MF
3. Check card profile documentation for correct file structure

### SW=6A82 (File not found) when reading EF 5001

**Symptom:**
```
{"ok":false,"error":"read EF 5001: card returned error SW=6A82"}
```

**Possible Causes:**
- EF 5001 doesn't exist on this card profile
- File structure is different (e.g., record file EF 5044)
- EFs are under a different DF
- Card requires authentication first

**Solutions:**

1. **Check card profile:**
   - Older Estonian ID cards (EstEID 3.5.x) use record file EF 5044
   - Newer ID1 cards (Cosmo v8.x) use transparent EFs 5001-500F
   - Thales cards might use different structure

2. **Try alternative FIDs:**
   ```powershell
   # Test with custom FIDs (if your card uses different structure)
   # You can modify the code to try EF 5044 (record file) or other FIDs
   ```

3. **Verify card is readable:**
   - Ensure card is properly inserted
   - Check PC/SC service is running
   - Verify card drivers are installed

4. **Check if authentication required:**
   - Some cards might require PIN or CAN before reading PD
   - Contact interface should allow reading PD without PIN
   - If card requires authentication, this is unexpected

### Card Not Detected

**Symptom:**
```
{"ok":false,"error":"pcsc: no readers found"}
```
or
```
{"ok":false,"error":"pcsc: could not connect to any reader"}
```

**Solutions:**
1. **Windows:**
   ```powershell
   # Check PC/SC service
   Get-Service SCardSvr
   
   # Start if stopped
   Start-Service SCardSvr
   ```

2. **Linux:**
   ```bash
   # Check pcscd service
   systemctl status pcscd
   
   # Start if stopped
   sudo systemctl start pcscd
   ```

3. **Verify card reader:**
   - Check reader is connected and powered
   - Verify card is inserted correctly
   - Try removing and reinserting card

4. **Check drivers:**
   - Install Estonian ID-card middleware/drivers
   - Verify reader drivers are installed
   - Check Windows Device Manager for reader

### Extension Not Communicating with Native Helper

**Symptom:**
- Extension installed but no response from native helper
- Error: "Native host not found"

**Solutions:**
1. **Verify host manifest:**
   - Check `host-manifests/ee.aarepuusepp.eid_helper.json` exists
   - Verify `path` field points to absolute path of `eid-pd-reader.exe`
   - Ensure extension ID matches `allowed_origins`

2. **Windows registry:**
   ```powershell
   # Check registry entry exists
   reg query "HKCU\Software\Google\Chrome\NativeMessagingHosts\ee.aarepuusepp.eid_helper"
   
   # Re-run install script if missing
   .\ops\install-host.ps1
   ```

3. **Test native helper directly:**
   ```powershell
   .\eid-pd-reader.exe --dump
   ```
   Should return JSON response

4. **Check extension permissions:**
   - Verify `nativeMessaging` permission in manifest
   - Check `externally_connectable` matches your origin

## Card-Specific Issues

### Thales Estonian ID-Card

**Known Issues:**
- May use different file structure than standard EstEID
- FIDs might differ from documented values
- May require different SELECT parameters

**Debugging Steps:**
1. Test with `--dump` flag to see exact error
2. Check card profile documentation
3. Try alternative FIDs if known
4. Consider using Estonian ID-card middleware tools to inspect card structure

### Older Card Profiles (EstEID 3.5.x)

**Characteristics:**
- Uses record file EF 5044 instead of transparent EFs
- Requires READ RECORD instead of READ BINARY
- Different file structure

**Solution:**
- Code currently targets transparent EF structure
- Would need modification to support record files
- Consider FID override feature for custom implementations

## Diagnostic Commands

### Test Native Helper
```powershell
# Windows
.\eid-pd-reader.exe --dump

# Should return JSON with PD or error message
```

### Check PC/SC Service
```powershell
# Windows
Get-Service SCardSvr | Format-List *

# Linux
systemctl status pcscd
```

### Verify Extension Installation
1. Open Chrome: `chrome://extensions/`
2. Enable Developer mode
3. Load unpacked extension from `extension/` directory
4. Note extension ID
5. Update host manifest with extension ID

### Test Extension → Native Communication
1. Open DevTools on a page matching your origin
2. Run in console:
   ```javascript
   chrome.runtime.sendNativeMessage("ee.aarepuusepp.eid_helper", 
     {cmd:"readPersonalData", fids:{}}, 
     console.log
   );
   ```

## Getting Help

If issues persist:

1. **Collect diagnostic information:**
   - Error message from `--dump` test
   - Card type/manufacturer
   - Card profile version (if known)
   - PC/SC service status
   - Extension ID and host manifest path

2. **Check logs:**
   - Browser console for extension errors
   - PC/SC logs (if available)
   - Windows Event Viewer for PC/SC service errors

3. **Verify prerequisites:**
   - Go 1.25.4 installed
   - PC/SC stack running
   - Card reader drivers installed
   - Estonian ID-card middleware installed (if required)

## Next Steps

If your Thales card uses a different file structure:

1. **Identify correct FIDs:**
   - Use Estonian ID-card middleware tools
   - Check card documentation
   - Inspect card with PC/SC tools

2. **Update FID overrides:**
   - Modify `webapp/config/pd-fallback.config.js`
   - Use `fids` parameter in API calls
   - Test with custom FIDs

3. **Extend code if needed:**
   - Add support for record files (EF 5044)
   - Implement alternative SELECT methods
   - Add card profile detection

