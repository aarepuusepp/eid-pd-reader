# Reader troubleshooting — 2.1.10

1. Confirm popup versions: reader/package 2.1.10 and extension 2.1.6. Different
   extension/reader version numbers are expected. Check protocol/capabilities.
2. If the extension is missing, follow the generated installation.html; Firefox
   temporary add-ons disappear after browser restart. Reload extension and page.
3. If native host is missing, repair the package. Do not manually add arbitrary
   extension IDs or bypass origin checks. Verify --version in Program Files.
4. If no reader/card, check Windows Smart Card service, USB driver and card seating.
   Presence does not prove a personalized card or valid encryption certificate.
5. Contact crypto needs a compatible x64 PKCS#11 DLL. Diagnose missing/invalid/
   wrong-architecture status separately from personal-data readiness.
6. NFC needs the correct card CAN, stable placement and a supported card profile.
   Keep CAN local; do not publish APDU/card dumps. An unpersonalized card can have
   no personal data. Initial ACR1252 double beeps also occurred in plain PC/SC
   connection diagnostics; they are not proof of duplicated application reads.
7. PIN timeout/cancellation releases no plaintext. Retry explicitly; never send PIN
   to the webpage or backend. Wrong recipient/key and unsupported profile are
   different errors from missing extension/card.

Source validation needs Go 1.27.3 and Windows x64 CGO compiler. Use the
[current README](../../README.md) and [direct NFC tests](../DIRECT-NFC-TEST.md).
