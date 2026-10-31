> Historical browser investigation. For current Windows package installation, identities and distribution use [reader README](../../README.md) and installer deployment instructions. Browser versions below are recorded test versions, not latest releases.

# Firefox Extension Signing - Quick Start

## Goal
Create a permanently installable Firefox extension that can be distributed with your installer.

## 5-Minute Setup

### Step 1: Get Mozilla API Credentials (One-time)

1. **Create Mozilla account:** https://addons.mozilla.org/developers/
2. **Get API credentials:** https://addons.mozilla.org/developers/addon/api/key/
3. **Click "Generate new credentials"**
4. **Save these values:**
   - API Key (JWT issuer): `user:12345:67`
   - API Secret (JWT secret): `abc123def456...`

⚠️ **Keep these secret!** Don't commit them to Git.

### Step 2: Install Node.js (If Not Installed)

Download from: https://nodejs.org/
- Choose LTS version
- Install with default options

### Step 3: Sign the Extension

```powershell
# Option A: Pass credentials as parameters
.\ops\sign-firefox-extension.ps1 `
    -ApiKey "user:12345:67" `
    -ApiSecret "abc123def456..."

# Option B: Use environment variables
$env:AMO_API_KEY = "user:12345:67"
$env:AMO_API_SECRET = "abc123def456..."
.\ops\sign-firefox-extension.ps1
```

**What happens:**
1. Script installs `web-ext` tool (if needed)
2. Packages your extension
3. Uploads to Mozilla for signing
4. Downloads signed `.xpi` file
5. Saves to `dist/esteid-pd-extension-signed.xpi`

⏱️ **Time:** 2-5 minutes (mostly waiting for Mozilla)

### Step 4: Test the Signed Extension

1. **Open Firefox**
2. **Drag `dist/esteid-pd-extension-signed.xpi` into Firefox**
3. **Click "Add"**
4. **Test card reading**
5. **Restart Firefox** - extension should still be there! ✅

### Step 5: Include in Installer

Copy the signed extension to your installer package:

```powershell
# Copy signed extension
Copy-Item "dist\esteid-pd-extension-signed.xpi" "installer\esteid-pd-extension.xpi"
```

Update `install.cmd` to install the extension:

```cmd
REM Open Firefox with extension file
if exist "esteid-pd-extension.xpi" (
    start firefox.exe "%CD%\esteid-pd-extension.xpi"
    echo Please click "Add" in Firefox to install the extension
)
```

---

## Complete Installation Flow

```
User runs install.cmd
  ↓
Installs eid-pd-reader.exe + native host
  ↓
Opens Firefox with .xpi file
  ↓
User clicks "Add"
  ↓
Extension permanently installed! ✅
```

---

## Troubleshooting

### "npm not found"
**Solution:** Install Node.js from https://nodejs.org/

### "Invalid API credentials"
**Solution:** 
- Check your API key and secret
- Make sure you copied them correctly
- Generate new credentials if needed

### "Extension validation failed"
**Solution:**
- Check `extension/manifest-firefox.json` syntax
- Ensure `id` is set: `"id": "aare.puusepp@gmail.com"`
- Check that all required files are present

### "Signing timed out"
**Solution:**
- Check internet connection
- Try again (Mozilla servers may be busy)
- Wait a few minutes and retry

---

## Updating the Extension

When you make changes:

1. **Update version** in `extension/manifest-firefox.json`:
   ```json
   "version": "1.0.1"
   ```

2. **Sign again:**
   ```powershell
   .\ops\sign-firefox-extension.ps1
   ```

3. **Distribute new `.xpi` file**

4. **Users install new version** (Firefox auto-updates)

---

## Security & Privacy

### What Mozilla Checks
- ✅ No malicious code
- ✅ Follows extension policies
- ✅ Proper permissions
- ✅ Valid manifest

### What Gets Signed
- ✅ Your extension code
- ✅ Manifest
- ✅ All resources

### What's NOT Sent to Mozilla
- ❌ User data
- ❌ Card data
- ❌ Private keys
- ❌ Logs

---

## FAQ

### Q: Do I need to list my extension publicly?
**A:** No! Choose "On your own" (self-distributed) during signing.

### Q: Is signing free?
**A:** Yes, completely free.

### Q: How long does signing take?
**A:** Usually 2-5 minutes. Automated review is fast.

### Q: Can I sign without Mozilla account?
**A:** No, but accounts are free and instant.

### Q: Will my extension work in all Firefox versions?
**A:** Yes, in Firefox 121+ (your minimum version).

### Q: Can I automate this in CI/CD?
**A:** Yes! Store API credentials as secrets and run the script.

### Q: What if Mozilla rejects my extension?
**A:** Very rare for self-distributed extensions. Fix issues and resubmit.

### Q: Can I revoke a signed extension?
**A:** Yes, through your Mozilla developer dashboard.

---

## Alternative: Enterprise Deployment

If you're deploying in a corporate environment, you can use Firefox policies instead:

**`C:\Program Files\Mozilla Firefox\distribution\policies.json`:**
```json
{
  "policies": {
    "ExtensionSettings": {
      "aare.puusepp@gmail.com": {
        "installation_mode": "force_installed",
        "install_url": "file:///C:/Program%20Files/eid-pd-reader/esteid-pd-extension.xpi"
      }
    }
  }
}
```

This allows unsigned extensions in enterprise environments.

---

## Summary

### ✅ What You Get
- Permanently installable Firefox extension
- Stable extension ID (`aare.puusepp@gmail.com`)
- Works in all Firefox versions
- No public listing required
- Full control over distribution

### 📦 What You Need
- Mozilla Developer Account (free)
- Node.js (for web-ext tool)
- 5 minutes for signing

### 🎯 Result
Users can install your extension like any other Firefox add-on, and it will persist across browser restarts.

---

## Next Steps

1. ✅ Get Mozilla API credentials
2. ✅ Run signing script
3. ✅ Test signed extension
4. ✅ Update installer
5. ✅ Distribute to users

**See also:**
- `FIREFOX-DISTRIBUTION.md` - Complete distribution guide
- `ops/sign-firefox-extension.ps1` - Signing script
- Mozilla docs: https://extensionworkshop.com/

