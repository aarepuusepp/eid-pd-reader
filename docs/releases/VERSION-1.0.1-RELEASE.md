# Version 1.0.1 Release Notes

**Release Date:** 2024-11-13  
**Type:** Maintenance Release

---

## 📦 What's New

### Version Information
- **Extension Version:** 1.0.1
- **Native Helper Version:** 1.0.1
- **API Version:** 1.0.1

---

## ✨ New Features

### 1. Version Flag
The native helper now supports a `--version` flag:

```bash
eid-pd-reader.exe --version
```

**Output:**
```
eid-pd-reader version 1.0.1
Estonian ID Card Personal Data Reader
https://github.com/aarepuusepp/web-eid-pd
```

### 2. Configuration System
Added comprehensive YAML-based configuration:

```yaml
# config.yaml
logging:
  enabled: true
  level: info
  file: ""
  
card:
  timeout_seconds: 30
  retry_attempts: 3
  
cache:
  enabled: true
  duration_seconds: 300
```

**Features:**
- Flexible logging configuration
- Card reader timeout and retry settings
- Cache control
- Security settings
- Custom FID mappings

**Documentation:** See [CONFIG-GUIDE.md](CONFIG-GUIDE.md)

---

## 🔄 Changes

### Naming Updates
- **Host name:** `com.example.eid_helper` → `ee.aarepuusepp.eid_helper`
- **Email:** `esteid-pd@example.com` → `aare.puusepp@gmail.com`
- **GitHub:** `github.com/yourusername/web-eid-pd` → `github.com/aarepuusepp/web-eid-pd`

### Documentation Reorganization
New structure for better navigation:

```
docs/
├── guides/          # Integration guides
├── installation/    # Installation guides
├── reference/       # Reference documentation
├── development/     # Developer documentation
└── specs/           # Technical specifications
```

### Go Module Update
- Module name: `ee.aarepuusepp.eid-pd-reader`
- New dependency: `gopkg.in/yaml.v3` for configuration

---

## 🐛 Bug Fixes

- Fixed all naming inconsistencies across codebase
- Updated all internal documentation links
- Corrected host manifest references in installation scripts

---

## 📝 Documentation Updates

### New Documents
1. **CONFIG-GUIDE.md** - Complete configuration guide
2. **HOUSEKEEPING-CHANGES.md** - Detailed change log
3. **VERIFICATION-COMPLETE.md** - Verification checklist

### Updated Documents
- **README.md** - Updated links and changelog
- **docs/README.md** - New documentation index
- All integration guides updated with new naming

---

## 🔧 Technical Details

### Files Changed
- **Extension manifests:** Version updated to 1.0.1
- **main.go:** Added version constant and flag
- **page-api.js:** Version updated to 1.0.1
- **go.mod:** Module renamed, yaml dependency added

### New Files
- `config.yaml` - Configuration template
- `native/config/config.go` - Configuration loader
- `host-manifests/ee.aarepuusepp.eid_helper.json`
- `host-manifests/ee.aarepuusepp.eid_helper.firefox.json`

### Removed Files
- `host-manifests/com.example.eid_helper.json` (old)
- `host-manifests/com.example.eid_helper.firefox.json` (old)

---

## 📦 Installation

### New Installation
Follow the standard installation process:

```bash
# Windows
.\install.cmd

# Linux/macOS
./ops/install-host.sh
```

### Upgrade from 1.0.0

**Important:** The host name has changed. You must:

1. **Uninstall old version:**
   ```bash
   # Windows
   .\uninstall.cmd
   
   # Linux/macOS
   rm -rf ~/.config/google-chrome/NativeMessagingHosts/com.example.eid_helper.json
   rm -rf ~/.mozilla/native-messaging-hosts/com.example.eid_helper.json
   ```

2. **Install new version:**
   ```bash
   # Windows
   .\install.cmd
   
   # Linux/macOS
   ./ops/install-host.sh
   ```

3. **Update extension:**
   - Remove old extension from browser
   - Load new extension (version 1.0.1)
   - Update extension ID in host manifest

4. **Restart browser**

---

## 🧪 Testing

### Verify Installation

**Check native helper version:**
```bash
eid-pd-reader.exe --version
```

**Test reading:**
```bash
eid-pd-reader.exe --dump
```

**Test configuration:**
```bash
eid-pd-reader.exe --config config.yaml --dump
```

**Check extension:**
```javascript
// In browser console
window.EstEID_PD.version
// Should return: "1.0.1"
```

---

## 🔒 Security

No security vulnerabilities fixed in this release.

**Security features:**
- Origin validation maintained
- Host manifest security unchanged
- Configuration file security added

---

## 📊 Compatibility

### Browser Support
- ✅ Chrome 88+
- ✅ Edge 88+
- ✅ Firefox 78+

### Operating Systems
- ✅ Windows 10/11 (64-bit)
- ✅ Linux (PC/SC Lite required)
- ✅ macOS 10.15+ (Catalina or later)

### Card Support
- ✅ IDEMIA Cosmo 8.1/8.2
- ✅ IDEMIA Cosmo X
- ✅ Thales (Gemalto)

---

## 🚀 Migration Guide

### For Developers

**Update your code:**
```javascript
// Old (still works)
if (window.EstEID_PD) {
  const result = await EstEID_PD.readPersonalData();
}

// New (recommended)
if (window.EstEID_PD && window.EstEID_PD.version === '1.0.1') {
  const result = await EstEID_PD.readPersonalData();
}
```

**Update documentation links:**
- Old: `APPLICATION-INTEGRATION-GUIDE.md`
- New: `docs/guides/APPLICATION-INTEGRATION-GUIDE.md`

### For System Administrators

**Update host manifests:**
```json
{
  "name": "ee.aarepuusepp.eid_helper",
  "description": "Estonian ID Card Personal Data Reader",
  "path": "C:\\Program Files\\eid-pd-reader\\eid-pd-reader.exe",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://YOUR_EXTENSION_ID/"
  ]
}
```

**Update registry keys (Windows):**
```powershell
# Remove old
reg delete "HKLM\Software\Google\Chrome\NativeMessagingHosts\com.example.eid_helper" /f

# Add new
reg add "HKLM\Software\Google\Chrome\NativeMessagingHosts\ee.aarepuusepp.eid_helper" /ve /t REG_SZ /d "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_helper.json" /f
```

---

## 📚 Documentation

### Main Documentation
- [README.md](README.md) - Project overview
- [CONFIG-GUIDE.md](CONFIG-GUIDE.md) - Configuration guide
- [docs/guides/APPLICATION-INTEGRATION-GUIDE.md](docs/guides/APPLICATION-INTEGRATION-GUIDE.md) - Integration guide
- [docs/guides/INTEGRATION-QUICK-REFERENCE.md](docs/guides/INTEGRATION-QUICK-REFERENCE.md) - Quick reference

### Installation
- [docs/installation/QUICK-START.md](docs/installation/QUICK-START.md) - Quick start
- [docs/installation/INSTALLATION.md](docs/installation/INSTALLATION.md) - Detailed installation

### Reference
- [docs/reference/TROUBLESHOOTING.md](docs/reference/TROUBLESHOOTING.md) - Troubleshooting
- [docs/reference/SECURITY.md](docs/reference/SECURITY.md) - Security
- [docs/reference/SUPPORTED-CARDS.md](docs/reference/SUPPORTED-CARDS.md) - Card compatibility

---

## 🆘 Support

### Getting Help
1. Check [docs/reference/TROUBLESHOOTING.md](docs/reference/TROUBLESHOOTING.md)
2. Review [HOUSEKEEPING-CHANGES.md](HOUSEKEEPING-CHANGES.md)
3. Search [GitHub Issues](https://github.com/aarepuusepp/web-eid-pd/issues)
4. Open a new issue
5. Email: aare.puusepp@gmail.com

### Known Issues
None at this time.

### Reporting Bugs
Please include:
- Version: 1.0.1
- Operating system
- Browser and version
- Steps to reproduce
- Error messages
- Output of `eid-pd-reader.exe --version`

---

## 🔮 What's Next

### Version 1.1.0 (Planned)
- [ ] Enhanced logging features
- [ ] Performance improvements
- [ ] Additional card type support
- [ ] Configuration validation

### Version 2.0.0 (Future)
- [ ] NFC support with PACE
- [ ] Certificate reading
- [ ] PIN operations
- [ ] GUI application

---

## 🙏 Acknowledgments

- All contributors and testers
- RIA (Estonian Information System Authority)
- Open-source community

---

## 📄 License

Same license as version 1.0.0.

---

## 📊 Statistics

### Changes in 1.0.1
- **Files changed:** 45+
- **New files:** 7
- **Documentation pages:** 24+
- **Lines of code:** 200+ changes
- **Configuration options:** 15+

### Download Size
- **Native helper:** ~3-5 MB (unchanged)
- **Extension:** ~50 KB (unchanged)

---

## ✅ Verification

**Verify your installation:**

```bash
# Check version
eid-pd-reader.exe --version
# Expected: eid-pd-reader version 1.0.1

# Test reading
eid-pd-reader.exe --dump
# Expected: JSON with personal data

# Check extension
# In browser console:
window.EstEID_PD.version
# Expected: "1.0.1"
```

---

**Release Date:** 2024-11-13  
**Release Type:** Maintenance  
**Status:** ✅ Stable  

**Download:** https://github.com/aarepuusepp/web-eid-pd/releases/tag/v1.0.1

---

**Thank you for using web-eid-pd!** 🎉

