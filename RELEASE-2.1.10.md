# EID-PD reader 2.1.10

- Pin published libcdoc-go v2.0.8, including the DigiDoc4-compatible default
  CDOC2 recipient label correction.
- Keep native reader and workstation package versions synchronized at 2.1.10.
  The separately released browser extension remains 2.1.6.
- Existing contact/NFC decryption APIs and consent/PIN boundaries are unchanged.

The corrected library's standalone P-384 container was successfully decrypted
with a real ID card in DigiDoc4 4.11.1.5434; the user verified the plaintext.
This is interoperability evidence for that path, not all recipient/card types.
