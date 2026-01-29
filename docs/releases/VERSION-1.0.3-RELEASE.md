# Version 1.0.3 Release Notes

**Release Date:** November 17, 2025

## Overview

Version 1.0.3 represents a major repository restructuring and naming standardization release. The project has been reorganized into three distinct components, each prepared to become a separate repository, with consistent naming throughout the codebase.

## Major Changes

### Repository Restructuring

The monorepo has been reorganized into three main components:

1. **eid-pd-reader/** - Native helper application (formerly in root)
2. **eid-pd-extension/** - Browser extension (formerly `extension/`)
3. **libopenpace-go/** - PACE protocol library (moved from `repos/`)

This structure prepares the project for splitting into three independent repositories while maintaining the ability to build all components together.

### Complete Naming Standardization

All references to the old `eid-helper` name have been replaced with `eid-pd-reader`:

- **Executable:** `eid-helper.exe` → `eid-pd-reader.exe`
- **Go Module:** `ee.aarepuusepp.eid-helper` → `ee.aarepuusepp.eid-pd-reader`
- **Host Name:** `ee.aarepuusepp.eid_helper` → `ee.aarepuusepp.eid_pd_reader`
- **Installation Path:** `C:\Program Files\eid-helper\` → `C:\Program Files\eid-pd-reader\`

### Documentation Reorganization

Documentation has been split and made component-specific:

- **Reader-specific docs** → `eid-pd-reader/docs/`
- **Extension-specific docs** → `eid-pd-extension/docs/`
- **Library docs** → `libopenpace-go/`
- **Root README** → Simplified overview with links to component READMEs

### Build System Updates

- **Root-level build:** `ops/build-all.ps1` - Builds all components
- **Component builds:** Each component has its own build scripts in `ops/` folders
- **GitHub Actions:** Updated workflows to work with new structure
- **Installation scripts:** Separate install/uninstall for each component, plus master scripts

### NFC/PACE Implementation

- Confirmed pure Go implementation using `libopenpace-go` v1.0.0
- Removed obsolete C-based build scripts (`build-nfc.ps1`, `build-nfc.sh`)
- NFC support is now built-in by default (no special build steps required)
- Updated documentation to reflect pure Go architecture
- `libopenpace-go` is now a versioned dependency (v1.0.0)

## Breaking Changes

### File Paths

- **Executable location changed:**
  - Old: `C:\Program Files\eid-helper\eid-helper.exe`
  - New: `C:\Program Files\eid-pd-reader\eid-pd-reader.exe`

- **Host manifest names changed:**
  - Old: `ee.aarepuusepp.eid_helper.json`
  - New: `ee.aarepuusepp.eid_pd_reader.json`

### Native Messaging

- **Host name changed:** Extensions must connect to `ee.aarepuusepp.eid_pd_reader`
- **Registry keys changed:** Windows registry paths updated to use new name

### Installation

Users must:
1. Uninstall old version (`eid-helper`)
2. Install new version (`eid-pd-reader`)
3. Update browser extension to version 1.0.3+
4. Update extension ID in host manifest (if using unpacked extension)

## Migration Guide

### For End Users

1. **Uninstall old version:**
   ```cmd
   cd <old-installation-directory>
   uninstall-all.cmd
   ```

2. **Install new version:**
   ```cmd
   cd eid-pd-reader
   install-reader.cmd
   ```

3. **Update browser extension:**
   - Remove old extension from browser
   - Install new version from `eid-pd-extension/`

4. **Verify installation:**
   - Open extension popup
   - Check that both extension and native helper versions show 1.0.3

### For Developers

1. **Update Git repository:**
   ```bash
   git pull
   ```

2. **Rebuild components:**
   ```powershell
   # Build all
   .\ops\build-all.ps1
   
   # Or build individually
   cd eid-pd-reader
   go mod tidy
   go build -o eid-pd-reader.exe ./cmd/eid-pd-reader
   ```

3. **Update import paths** (if you have custom code):
   ```go
   // Old
   import "ee.aarepuusepp.eid-helper/native/esteid"
   
   // New
   import "ee.aarepuusepp.eid-pd-reader/native/esteid"
   ```

4. **Update host manifest:**
   - See [Extension ID Configuration Guide](../guides/EXTENSION-ID-CONFIGURATION.md)

## Technical Details

### Changed Files (62 files total)

**Code Files:**
- `eid-pd-reader/main.go` - Module path and app name
- `eid-pd-reader/go.mod` - Module name, versioned libopenpace-go dependency (v1.0.0)
- `eid-pd-reader/native/**/*.go` - Import paths
- `eid-pd-extension/extension/background*.js` - Host name
- `libopenpace-go/README.md` - Version 1.0.0 documentation

**Configuration Files:**
- `eid-pd-reader/config.yaml` - Log paths
- `eid-pd-reader/host-manifests/**/*.json` - Host name and paths
- `eid-pd-extension/extension/manifest*.json` - Host name

**Scripts:**
- `eid-pd-reader/ops/*.ps1` - Executable name and paths
- `eid-pd-reader/ops/*.sh` - Executable name and paths
- `eid-pd-extension/ops/*.ps1` - Host name references
- `ops/*.ps1` - Component paths and names

**Registry Files:**
- `eid-pd-extension/ops/register-*.reg` - Registry paths and host name

**Documentation:**
- All `.md` files in `eid-pd-reader/docs/` (39 files)
- All `.md` files in `eid-pd-extension/docs/` (10 files)
- Root and component `README.md` files

### New Files

- `README.md` (root) - Monorepo overview
- `eid-pd-reader/README.md` - Reader-specific documentation
- `eid-pd-extension/README.md` - Extension-specific documentation
- `eid-pd-reader/docs/guides/EXTENSION-ID-CONFIGURATION.md` - Configuration guide
- `ops/install-all.cmd` - Master installation script
- `ops/uninstall-all.cmd` - Master uninstallation script
- `.gitignore` files for each component
- `libopenpace-go/VERSION` - Version file (1.0.0)
- `libopenpace-go/CHANGELOG.md` - Version history and release notes

### Removed Files

- `ops/build-nfc.ps1` - No longer needed (NFC built-in)
- `ops/build-nfc.sh` - No longer needed (NFC built-in)
- `ops/build-openpace.md` - Obsolete C-based instructions

## Compatibility

### Supported Platforms

- **Windows:** 10/11 (64-bit)
- **Linux:** Ubuntu 20.04+, Debian 11+, Fedora 35+
- **macOS:** 10.15+ (Catalina or later)

### Browser Support

- **Chrome/Edge:** Version 88+ (Manifest V3)
- **Firefox:** Version 109+ (Manifest V2)

### Card Support

No changes to card support:
- IDEMIA Cosmo 8.1/8.2
- IDEMIA Cosmo X
- Thales (Gemalto) cards
- Standard Estonian ID cards

## Known Issues

None specific to this release.

## Testing

All components have been tested on:
- ✅ Windows 11 with Chrome 120
- ✅ Windows 11 with Edge 120
- ✅ Windows 11 with Firefox 121
- ✅ Contact card readers
- ✅ NFC card readers with PACE

## Upgrade Path

### From 1.0.2 → 1.0.3

**Required steps:**
1. Uninstall version 1.0.2
2. Install version 1.0.3
3. Update browser extension
4. Restart browser

**Data migration:** None required (no configuration format changes)

### From 1.0.1 → 1.0.3

Same as 1.0.2 → 1.0.3

### From earlier versions

Not supported. Please upgrade to 1.0.2 first, then to 1.0.3.

## Future Plans

### Version 1.1.0 (Planned)

- Split into three separate repositories
- Individual release cycles for each component
- Improved error messages
- Enhanced logging options

### Version 2.0.0 (Future)

- Manifest V3 for Firefox (when supported)
- Additional card types
- Performance optimizations

## Contributors

- Aare Puusepp (@aarepuusepp) - Repository restructuring, naming standardization, documentation updates

## Support

- **Issues:** [GitHub Issues](https://github.com/aarepuusepp/web-eid-pd/issues)
- **Email:** aare.puusepp@gmail.com
- **Documentation:** See component-specific README files

## License

MIT License - No changes

---

**Full Changelog:** [v1.0.2...v1.0.3](https://github.com/aarepuusepp/web-eid-pd/compare/v1.0.2...v1.0.3)

