> Historical browser investigation. For current Windows package installation, identities and distribution use [reader README](../../README.md) and installer deployment instructions. Browser versions below are recorded test versions, not latest releases.

# Firefox Extension Distribution Guide

## Overview

Firefox requires all extensions to be **signed by Mozilla** for security. However, you have several options for distributing your extension **without listing it publicly** on addons.mozilla.org (AMO).

## Distribution Options

### ✅ Option 1: Self-Distributed Signed Extension (Recommended)

**Best for:** General distribution to end users

**Process:**
1. Sign the extension with Mozilla (free, automated)
2. Get a signed `.xpi` file
3. Distribute the `.xpi` file yourself
4. Users install it like any other extension

**Advantages:**
- ✅ Works in all Firefox versions
- ✅ No public listing required
- ✅ Free
- ✅ Automated process
- ✅ Extension gets a permanent ID

**Disadvantages:**
- ⚠️ Requires Mozilla Developer Account
- ⚠️ Must follow Mozilla's extension policies

### ✅ Option 2: Enterprise Policy Distribution

**Best for:** Corporate/organizational deployment

**Process:**
1. Deploy unsigned extension via Group Policy
2. Firefox allows it based on enterprise policy

**Advantages:**
- ✅ No Mozilla account needed
- ✅ Full control
- ✅ Automated deployment

**Disadvantages:**
- ⚠️ Only works in enterprise environments
- ⚠️ Requires Active Directory or policy management

### ⚠️ Option 3: Firefox Developer/Nightly (Not Recommended)

**Best for:** Development only

**Process:**
1. Users install Firefox Developer Edition
2. Disable signature verification
3. Install unsigned extension

**Disadvantages:**
- ❌ Not suitable for production
- ❌ Requires special Firefox version
- ❌ Security risk

---

## Recommended Solution: Self-Distributed Signed Extension

This is the best option for your use case. Here's how to do it:

### Step 1: Prepare Extension for Signing

Your extension needs a permanent ID in the manifest:

**`extension/manifest-firefox.json`:**
```json
{
  "manifest_version": 3,
  "name": "Estonian ID Card Personal Data Reader",
  "version": "1.0.0",
  "description": "Read personal data from Estonian ID card chip",
  
  "browser_specific_settings": {
    "gecko": {
      "id": "aare.puusepp@gmail.com",
      "strict_min_version": "121.0"
    }
  },
  
  // ... rest of manifest
}
```

✅ **Your extension already has this!** (`aare.puusepp@gmail.com`)

### Step 2: Get Mozilla Developer Account

1. Go to https://addons.mozilla.org
2. Click "Register" or "Log in"
3. Create a free account
4. Verify your email

### Step 3: Get API Credentials

1. Go to https://addons.mozilla.org/developers/addon/api/key/
2. Click "Generate new credentials"
3. Save your:
   - **API Key** (JWT issuer)
   - **API Secret** (JWT secret)

⚠️ **Keep these secret!** They're like passwords.

### Step 4: Sign the Extension

#### Method A: Using web-ext (Automated)

Install web-ext:
```bash
npm install -g web-ext
```

Sign the extension:
```bash
cd extension
web-ext sign \
  --api-key=YOUR_API_KEY \
  --api-secret=YOUR_API_SECRET \
  --channel=unlisted
```

This will:
- Upload your extension to Mozilla
- Mozilla automatically reviews and signs it (usually < 10 minutes)
- Download the signed `.xpi` file
- Save it to `web-ext-artifacts/`

#### Method B: Manual Upload

1. Package your extension as ZIP:
   ```powershell
   .\ops\package-extension-firefox.ps1
   ```

2. Go to https://addons.mozilla.org/developers/addon/submit/distribution
3. Select **"On your own"** (self-distributed)
4. Upload the ZIP file
5. Wait for automated review (usually < 10 minutes)
6. Download the signed `.xpi` file

### Step 5: Distribute the Signed Extension

The signed `.xpi` file can be:
- Downloaded from your website
- Included in your installer
- Sent via email
- Distributed on internal network

**Users install it by:**
1. Downloading the `.xpi` file
2. Opening Firefox
3. Dragging the `.xpi` file into Firefox
4. Clicking "Add" when prompted

✅ **The extension is now permanently installed!**

---

## Automated Signing Script

I'll create a script to automate the signing process:

**`ops/sign-firefox-extension.ps1`:**
```powershell
# Sign Firefox extension using web-ext
param(
    [Parameter(Mandatory=$true)]
    [string]$ApiKey,
    
    [Parameter(Mandatory=$true)]
    [string]$ApiSecret,
    
    [string]$OutputDir = "dist"
)

# Check if web-ext is installed
if (-not (Get-Command web-ext -ErrorAction SilentlyContinue)) {
    Write-Host "Installing web-ext..." -ForegroundColor Yellow
    npm install -g web-ext
}

# Create output directory
New-Item -Path $OutputDir -ItemType Directory -Force | Out-Null

# Sign the extension
Write-Host "Signing Firefox extension..." -ForegroundColor Cyan
cd extension
web-ext sign `
    --api-key=$ApiKey `
    --api-secret=$ApiSecret `
    --channel=unlisted `
    --artifacts-dir="../$OutputDir"

cd ..

Write-Host "`nSigned extension saved to: $OutputDir" -ForegroundColor Green
Write-Host "Distribute the .xpi file to your users." -ForegroundColor Yellow
```

---

## Integration with Installer

Update your `install.cmd` to include the Firefox extension:

### Step 1: Include Signed Extension in Installer

Place the signed `.xpi` file in your installer package:
```
installer/
├── install.cmd
├── eid-pd-reader.exe
├── esteid-pd-extension.xpi  ← Signed extension
└── host-manifests/
```

### Step 2: Install Extension Automatically

**Option A: Silent Install (Requires Admin)**

```cmd
REM Install Firefox extension system-wide
copy "esteid-pd-extension.xpi" "%ProgramFiles%\Mozilla Firefox\browser\extensions\aare.puusepp@gmail.com.xpi"
```

**Option B: User Install (Recommended)**

```cmd
REM Copy extension to user's profile
set FIREFOX_PROFILE=%APPDATA%\Mozilla\Firefox\Profiles\
for /d %%p in ("%FIREFOX_PROFILE%\*.default-release") do (
    copy "esteid-pd-extension.xpi" "%%p\extensions\aare.puusepp@gmail.com.xpi"
)
```

**Option C: Prompt User to Install**

```cmd
REM Open Firefox with extension file
start firefox.exe "esteid-pd-extension.xpi"
```

---

## Enterprise Deployment (Alternative)

If you're deploying in an enterprise environment, you can use Firefox policies to allow unsigned extensions.

### Create Policy File

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

This forces the extension to be installed for all users.

### Group Policy (Windows Domain)

1. Download Firefox ADMX templates
2. Import into Group Policy
3. Configure extension installation

---

## Complete Installation Flow

### For End Users (Recommended)

```
1. Run install.cmd (as Administrator)
   ├─ Installs eid-pd-reader.exe
   ├─ Registers native host
   └─ Opens Firefox with extension file

2. User clicks "Add" in Firefox
   └─ Extension is permanently installed

3. Done! ✅
```

### For Enterprise (Automated)

```
1. Run install.cmd (as Administrator)
   ├─ Installs eid-pd-reader.exe
   ├─ Registers native host
   ├─ Copies extension to Firefox directory
   └─ Creates enterprise policy

2. User opens Firefox
   └─ Extension is automatically installed

3. Done! ✅
```

---

## Updated Installer

I'll create an updated installer that handles Firefox extension installation:

**`install-with-firefox-extension.cmd`:**
```cmd
@echo off
REM ... (existing installer code)

REM ============================================================================
echo [7/7] Installing Firefox extension...
echo.

REM Check if Firefox is installed
reg query "HKLM\SOFTWARE\Mozilla\Mozilla Firefox" >nul 2>&1
if %errorLevel% neq 0 (
    echo     [SKIP] Firefox not installed
    goto :END
)

REM Check if signed extension exists
if not exist "esteid-pd-extension.xpi" (
    echo     [SKIP] Signed extension not found
    echo     Please sign the extension first using ops\sign-firefox-extension.ps1
    goto :END
)

REM Option 1: Copy to Firefox extensions directory (requires admin)
set FIREFOX_EXT_DIR=%ProgramFiles%\Mozilla Firefox\browser\extensions
if exist "%FIREFOX_EXT_DIR%" (
    copy /Y "esteid-pd-extension.xpi" "%FIREFOX_EXT_DIR%\aare.puusepp@gmail.com.xpi" >nul
    if %errorLevel% equ 0 (
        echo     [OK] Extension installed system-wide
    ) else (
        echo     [WARN] Failed to install system-wide
    )
)

REM Option 2: Open in Firefox for user installation
echo     [INFO] Opening extension in Firefox for user installation...
start firefox.exe "%CD%\esteid-pd-extension.xpi"
echo     [INFO] Click "Add" in Firefox to complete installation

:END
echo.
echo ============================================================================
echo Installation Complete!
echo ============================================================================
```

---

## Testing

### Test Signed Extension

1. Sign the extension
2. Install it in Firefox
3. Test card reading
4. Check that extension persists after Firefox restart

### Verify Permanent Installation

```powershell
# Check installed extensions
Get-Content "$env:APPDATA\Mozilla\Firefox\Profiles\*.default-release\extensions.json" | Select-String "esteid-pd"
```

---

## Troubleshooting

### Extension Not Signing

**Error:** "Validation failed"
- Check manifest.json syntax
- Ensure all required fields are present
- Check that `id` is set in `browser_specific_settings`

### Extension Not Installing

**Error:** "This add-on could not be installed because it appears to be corrupt"
- Re-download the signed `.xpi` file
- Don't unzip or modify the signed file

### Extension Disappears After Restart

**Cause:** Extension is not signed or signature is invalid
- Use a properly signed `.xpi` file
- Don't use temporary add-on for production

---

## Security Considerations

### Code Signing

- ✅ Mozilla's signing ensures integrity
- ✅ Users can trust the extension
- ✅ Prevents tampering

### Privacy

- ✅ Extension only requests necessary permissions
- ✅ No data sent to external servers
- ✅ All processing is local

### Updates

To update the extension:
1. Increment version in manifest
2. Sign new version
3. Distribute new `.xpi` file
4. Firefox automatically updates when users install new version

---

## Summary

### ✅ Recommended Approach

1. **Sign extension with Mozilla** (free, automated)
2. **Distribute signed `.xpi` file** with your installer
3. **Users install by opening the file** in Firefox
4. **Extension is permanently installed** with stable ID

### 📦 What You Need

- Mozilla Developer Account (free)
- API credentials (from AMO)
- `web-ext` tool (npm package)
- Signed `.xpi` file

### 🎯 Result

- ✅ Extension works in all Firefox versions
- ✅ Permanent installation (survives restarts)
- ✅ Stable extension ID (`aare.puusepp@gmail.com`)
- ✅ No public listing required
- ✅ Full control over distribution

---

## Next Steps

1. Create Mozilla Developer Account
2. Get API credentials
3. Run signing script
4. Test signed extension
5. Update installer to include `.xpi` file
6. Distribute to users

See also:
- `ops/sign-firefox-extension.ps1` - Automated signing script
- `install-with-firefox-extension.cmd` - Updated installer
- Mozilla documentation: https://extensionworkshop.com/documentation/publish/

