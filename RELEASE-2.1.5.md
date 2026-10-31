# Reader 2.1.5 (2026-10-31)

- Add NFC certificate reading and CDOC transport-key unwrapping for supported
  IDEMIA IAS-ECC and Thales PKCS#15 P-384 profiles. Contact PKCS#11 is unchanged.
- Pin the exact libopenpace-go 1.1.3 snapshot. Correct SELECT parameters and
  case-1 secure-messaging MAC padding that caused SW=6A86, SW=6A82 and SW=6988.
- Hold PC/SC transactions across NFC reads, skip SAM slots and stop retrying
  authentication failures. Remove raw secure-messaging diagnostic output.
- Keep PIN1 in the native 60-second dialog, validate the certificate before and
  after PIN entry, and perform one PIN attempt. CAN is not persisted.
- Add cmd/eid-pd-card-test for direct card diagnostics without a browser. It
  reports payload hashes and lengths; decrypted contents are not saved.

See docs/NFC-CRYPTO.md for supported profiles and docs/DIRECT-NFC-TEST.md for
direct testing. Native messaging protocol remains 1. Use extension 2.1.5 and
web-eid-pd.js 1.0.2 for the browser NFC cryptography interface.

Acceptance: Visitor NFC personal-data reading and contact decryption were
confirmed by the user. On 2026-10-31, direct IDEMIA NFC CDOC v1 decryption
returned two payloads; the user also confirmed NFC decryption in Digisign after
installing the fixed development package. The two activation beeps were
reproduced during plain PC/SC connection without PACE; they are not claimed fixed.

Real-card Thales, CDOC2 and the full NFC encryption round trip remain pending.
Simulated-card tests cover both profiles and CDOC 1.1/2.0 key derivation.
