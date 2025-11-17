# Documentation Index

This directory contains all technical documentation for the web-eid-pd project, organized by category.

---

## 📁 Directory Structure

```
docs/
├── guides/                    # Integration and usage guides
│   ├── APPLICATION-INTEGRATION-GUIDE.md
│   ├── INTEGRATION-QUICK-REFERENCE.md
│   ├── INTEGRATION-SUMMARY.md
│   ├── DEVELOPER-CHEAT-SHEET.md
│   ├── INTEGRATION-FLOW.md
│   └── WHEN-TO-USE.md
│
├── installation/              # Installation guides
│   ├── INSTALLATION.md
│   ├── QUICK-START.md
│   └── EXTENSION-INSTALL.md
│
├── reference/                 # Reference documentation
│   ├── TROUBLESHOOTING.md
│   ├── SECURITY.md
│   ├── SUPPORTED-CARDS.md
│   ├── BROWSER-SUPPORT.md
│   ├── BROWSER-VERSIONS-SUMMARY.md
│   ├── FIREFOX-*.md
│   ├── UNIFIED-MANIFEST.md
│   └── web_eid_personal_data_fallback_go_native_helper_extension_api.md
│
├── development/               # Developer documentation
│   ├── EXTENSION-GUIDE.md
│   ├── EXTENSION-SUMMARY.md
│   ├── MANIFEST-V3-MIGRATION.md
│   └── MIGRATION.md
│
└── specs/                     # Technical specifications (PDFs)
    ├── ged24035_015_tdc_est_eid_developer_guide.pdf
    ├── id1developerguide.pdf
    ├── RIA-EstEID-Chip-App-v3.5.8_fix_form.pdf
    ├── td-id1-chip-app-4.pdf
    └── estonia_cosmov8_cosmox_discrepancies.pdf
```

---

## 🎯 Quick Navigation

### I want to...

#### **Install the software**
- [QUICK-START.md](installation/QUICK-START.md) - 5-minute quick start
- [INSTALLATION.md](installation/INSTALLATION.md) - Detailed installation guide
- [EXTENSION-INSTALL.md](installation/EXTENSION-INSTALL.md) - Browser extension installation

#### **Integrate into my web app**
- [APPLICATION-INTEGRATION-GUIDE.md](guides/APPLICATION-INTEGRATION-GUIDE.md) - Complete integration guide
- [INTEGRATION-QUICK-REFERENCE.md](guides/INTEGRATION-QUICK-REFERENCE.md) - One-page API reference
- [DEVELOPER-CHEAT-SHEET.md](guides/DEVELOPER-CHEAT-SHEET.md) - Printable cheat sheet

#### **Understand the system**
- [INTEGRATION-FLOW.md](guides/INTEGRATION-FLOW.md) - Architecture and flow diagrams
- [WHEN-TO-USE.md](guides/WHEN-TO-USE.md) - Decision guide and use cases
- [INTEGRATION-SUMMARY.md](guides/INTEGRATION-SUMMARY.md) - Complete documentation index

#### **Fix problems**
- [TROUBLESHOOTING.md](reference/TROUBLESHOOTING.md) - Common issues and solutions
- [BROWSER-SUPPORT.md](reference/BROWSER-SUPPORT.md) - Browser compatibility
- [SUPPORTED-CARDS.md](reference/SUPPORTED-CARDS.md) - Card compatibility

#### **Develop the extension**
- [EXTENSION-GUIDE.md](development/EXTENSION-GUIDE.md) - Extension development guide
- [MANIFEST-V3-MIGRATION.md](development/MANIFEST-V3-MIGRATION.md) - Manifest V3 migration
- [EXTENSION-SUMMARY.md](development/EXTENSION-SUMMARY.md) - Extension overview

#### **Review security**
- [SECURITY.md](reference/SECURITY.md) - Security considerations and best practices

---

## 📚 Documentation by Type

### 📖 Guides (guides/)
Complete guides for integration and usage:
- **APPLICATION-INTEGRATION-GUIDE.md** - Main integration guide (~1,500 lines)
- **INTEGRATION-QUICK-REFERENCE.md** - Quick API reference (~400 lines)
- **INTEGRATION-SUMMARY.md** - Documentation hub (~600 lines)
- **DEVELOPER-CHEAT-SHEET.md** - Printable reference (~300 lines)
- **INTEGRATION-FLOW.md** - Architecture diagrams (~600 lines)
- **WHEN-TO-USE.md** - Decision guide (~800 lines)

### 🔧 Installation (installation/)
Installation and setup guides:
- **QUICK-START.md** - 5-minute quick start
- **INSTALLATION.md** - Detailed installation instructions
- **EXTENSION-INSTALL.md** - Browser extension setup

### 📋 Reference (reference/)
Reference documentation and troubleshooting:
- **TROUBLESHOOTING.md** - Common issues and solutions
- **SECURITY.md** - Security best practices
- **SUPPORTED-CARDS.md** - Card compatibility information
- **BROWSER-SUPPORT.md** - Browser compatibility matrix
- **BROWSER-VERSIONS-SUMMARY.md** - Browser version details
- **FIREFOX-*.md** - Firefox-specific documentation
- **UNIFIED-MANIFEST.md** - Manifest documentation
- **web_eid_personal_data_fallback_go_native_helper_extension_api.md** - Technical API spec

### 💻 Development (development/)
Developer documentation:
- **EXTENSION-GUIDE.md** - Extension development guide
- **EXTENSION-SUMMARY.md** - Extension overview
- **MANIFEST-V3-MIGRATION.md** - Manifest V3 migration guide
- **MIGRATION.md** - General migration guide

### 📄 Specifications (specs/)
Technical specifications from card manufacturers:
- **ged24035_015_tdc_est_eid_developer_guide.pdf** - Official Estonian eID guide
- **id1developerguide.pdf** - ID-1 card developer guide
- **RIA-EstEID-Chip-App-v3.5.8_fix_form.pdf** - EstEID chip app spec
- **td-id1-chip-app-4.pdf** - Technical documentation
- **estonia_cosmov8_cosmox_discrepancies.pdf** - Card version discrepancies

---

## 📊 Documentation by Role

### Web Developer
**Start here:**
1. [guides/APPLICATION-INTEGRATION-GUIDE.md](guides/APPLICATION-INTEGRATION-GUIDE.md)
2. [guides/INTEGRATION-QUICK-REFERENCE.md](guides/INTEGRATION-QUICK-REFERENCE.md)
3. [guides/DEVELOPER-CHEAT-SHEET.md](guides/DEVELOPER-CHEAT-SHEET.md)

**Also useful:**
- [guides/INTEGRATION-FLOW.md](guides/INTEGRATION-FLOW.md)
- [reference/TROUBLESHOOTING.md](reference/TROUBLESHOOTING.md)

### System Administrator
**Start here:**
1. [installation/QUICK-START.md](installation/QUICK-START.md)
2. [installation/INSTALLATION.md](installation/INSTALLATION.md)

**Also useful:**
- [reference/TROUBLESHOOTING.md](reference/TROUBLESHOOTING.md)
- [reference/BROWSER-SUPPORT.md](reference/BROWSER-SUPPORT.md)

### Product Manager / Architect
**Start here:**
1. [guides/WHEN-TO-USE.md](guides/WHEN-TO-USE.md)
2. [guides/INTEGRATION-FLOW.md](guides/INTEGRATION-FLOW.md)
3. [guides/INTEGRATION-SUMMARY.md](guides/INTEGRATION-SUMMARY.md)

**Also useful:**
- [reference/SECURITY.md](reference/SECURITY.md)

### Security Engineer
**Start here:**
1. [reference/SECURITY.md](reference/SECURITY.md)
2. [guides/INTEGRATION-FLOW.md](guides/INTEGRATION-FLOW.md)

**Also useful:**
- [guides/APPLICATION-INTEGRATION-GUIDE.md](guides/APPLICATION-INTEGRATION-GUIDE.md) (security section)

### Extension Developer
**Start here:**
1. [development/EXTENSION-GUIDE.md](development/EXTENSION-GUIDE.md)
2. [development/EXTENSION-SUMMARY.md](development/EXTENSION-SUMMARY.md)

**Also useful:**
- [development/MANIFEST-V3-MIGRATION.md](development/MANIFEST-V3-MIGRATION.md)
- [reference/web_eid_personal_data_fallback_go_native_helper_extension_api.md](reference/web_eid_personal_data_fallback_go_native_helper_extension_api.md)

---

## 🚀 Getting Started Paths

### Beginner (Just want it to work)
1. [installation/QUICK-START.md](installation/QUICK-START.md) - 10 min
2. Try the demo: `../demo/index.html` - 5 min
3. [guides/INTEGRATION-QUICK-REFERENCE.md](guides/INTEGRATION-QUICK-REFERENCE.md) - 10 min

**Total time:** ~25 minutes

### Intermediate (Building an application)
1. [../README.md](../README.md) - 15 min
2. [guides/APPLICATION-INTEGRATION-GUIDE.md](guides/APPLICATION-INTEGRATION-GUIDE.md) - 45 min
3. [guides/INTEGRATION-FLOW.md](guides/INTEGRATION-FLOW.md) - 20 min
4. [reference/SECURITY.md](reference/SECURITY.md) - 15 min

**Total time:** ~1.5 hours

### Advanced (Enterprise deployment)
1. All Intermediate docs
2. [guides/WHEN-TO-USE.md](guides/WHEN-TO-USE.md) - 20 min
3. [installation/INSTALLATION.md](installation/INSTALLATION.md) - 30 min
4. [development/EXTENSION-GUIDE.md](development/EXTENSION-GUIDE.md) - 30 min

**Total time:** ~3 hours

---

## 🔗 External Resources

### Official Documentation
- [RIA GitHub](https://github.com/open-eid) - Open-source eID implementations
- [ID.ee](https://www.id.ee/) - Official Estonian ID card information
- [web-eid.js](https://github.com/web-eid/web-eid.js) - Certificate-based authentication library

### Standards
- ISO/IEC 7816 - Smart card standards
- ISO/IEC 14443 - Contactless card standards
- PKCS#15 - Cryptographic token information format

---

## 🆘 Need Help?

1. Check [reference/TROUBLESHOOTING.md](reference/TROUBLESHOOTING.md)
2. Review [guides/INTEGRATION-SUMMARY.md](guides/INTEGRATION-SUMMARY.md)
3. Search [GitHub Issues](https://github.com/aarepuusepp/web-eid-pd/issues)
4. Open a new issue

---

## 📝 Contributing to Documentation

Found an error or want to improve documentation?

1. Fork the repository
2. Make your changes
3. Submit a pull request
4. Tag with `documentation` label

### Documentation Guidelines

- Use clear, simple language
- Include code examples
- Add diagrams where helpful
- Link between related documents
- Test all code examples
- Keep formatting consistent

---

## 📄 License

All documentation is provided under the same license as the project.

---

**Back to:** [Main README](../README.md) | [Integration Summary](guides/INTEGRATION-SUMMARY.md)
