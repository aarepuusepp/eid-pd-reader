# Browser support for the Windows workstation package

The development distribution targets Chrome, Edge and Firefox on Windows x64.
Package/reader: 2.1.10. Extension: 2.1.6. Website SDK: 1.0.2.
Use installer-generated Chromium/Firefox directories and native registrations.
[Current setup](../../README.md#installation-and-update).

Firefox temporary add-ons are removed on browser restart. Reload the application
tab after any extension change. Public browser-store delivery/signing is a separate
release gate. Safari/Linux/macOS/Windows ARM64 are not accepted by this package.
Historical browser version tables describe prior tests, not the current browser
market or a guaranteed minimum. Real card and mocked browser acceptance have
separate scope in the versioned release and NFC evidence documents.
