# EID-PD reader 2.1.10

- Pin published libcdoc-go v2.0.8, including the DigiDoc4-compatible default
  CDOC2 recipient label correction.
- Keep native reader and workstation package versions synchronized at 2.1.10.
  The separately released browser extension remains 2.1.6.
- Existing contact/NFC decryption APIs and consent/PIN boundaries are unchanged.

The corrected library's standalone P-384 container was successfully decrypted
with a real ID card in DigiDoc4 4.11.1.5434; the user verified the plaintext.
This is interoperability evidence for that path, not all recipient/card types.


## Installed-release acceptance (2026-10-31)

After installing EID-PD 2.1.10, the user confirmed successful encryption and
subsequent decryption of a CDOC 2.0 standalone container in Digisign, and
successful ID-card personal-data reading in Visitor. The release set is:
reader/package 2.1.10, extension 2.1.6, web SDK 1.0.2, libcdoc-go 2.0.8,
libopenpace-go 1.1.3, Digisign backend 2.0.7/UI 0.0.8, Visitor Vue 2.1.6.

This is user-reported real-environment acceptance. The latest report does not
identify the browser, card profile or contact/NFC transport, so it does not
establish a new browser/card/transport matrix. Earlier separately confirmed
NFC and DigiDoc4 results retain their original scope. Central transfer-server
and Smart-ID/Mobile-ID acceptance are separate from standalone CDOC2.
