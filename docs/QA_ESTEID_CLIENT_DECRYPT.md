# EstEID Client CDOC Decrypt — Manual QA Matrix

> **Full install/configure/publish guide:** [`docs/CLIENT_DECRYPT_INSTALLATION_GUIDE.md`](NFC-CRYPTO.md) (workspace root)

Scope: CDOC **1.0** and **1.1** only.

## Environment matrix

| OS | Browser | OpenSC | EstEID card |
|----|---------|--------|-------------|
| Windows 10/11 | Chrome / Edge | opensc-pkcs11.dll | Physical card + reader |
| Ubuntu 22.04+ | Chrome / Firefox | opensc-pkcs11.so | Physical card + reader |

## Pre-checks

- [ ] `web-eid-decrypt.js` `status()` returns extension + nativeApp versions (no `ERR_EIDPD_*`)
- [ ] digisign-encrypt-ui shows **Decrypt with ID card** when recipient status is `CLIENT_KEY`
- [ ] No PIN field in web UI; native PIN1 dialog appears

## Test cases

| # | Container | Recipient key | Expected |
|---|-----------|---------------|----------|
| 1 | CDOC 1.0 | RSA (EstEID auth cert) | Job `DECRYPTED`, files downloadable |
| 2 | CDOC 1.1 | RSA | Job `DECRYPTED` |
| 3 | CDOC 1.1 | EC (P-384) | Job `DECRYPTED`; verify EC digest matches container |
| 4 | Multi-recipient | Actor matches one recipient only | Only matching cert unwraps; wrong card → `ERR_EIDPD_WRONG_KEY` |
| 5 | No card inserted | — | `ERR_EIDPD_CARD_NOT_FOUND` before PIN |
| 6 | User cancels PIN | — | `ERR_EIDPD_USER_CANCELLED`, job `FAILED` via client-fail |
| 7 | Challenge TTL (>5 min idle) | — | `challenge_expired` on finish |
| 8 | CDOC 2.0 container | — | Prepare returns `unsupported_container` (out of scope) |

## UX parity with web-eid signing

- [ ] Single action button starts flow
- [ ] Card insert prompt when reader empty
- [ ] Native permission/PIN dialogs only
- [ ] Container never downloaded to browser (network tab: only challenge + CEK JSON)

## Regression

- [ ] Server-side PKCS#11 / PKCS#12 decrypt unchanged
- [ ] `EstEID_PD` personal-data read unchanged
- [ ] `EstEID_PD.decryptContainer` still works in demo/CLI (deprecated path)
