# NFC certificate reading and CDOC decryption

This development implementation adds `mode: "nfc"` and a six-digit `can` to
`GET_DECRYPT_CERTIFICATE` and `UNWRAP_TRANSPORT_KEY`. Update the native reader,
browser extension and web SDK together. Native status advertises
`nfcCryptoSupported`; the extension additionally advertises `nfcCryptoBridge`.
The SDK rejects missing capabilities instead of silently using a contact reader.

## Supported profiles

- IDEMIA IAS-ECC: authentication certificate ADF1/3401, PIN1 reference 01.
- Thales PKCS#15: authentication certificate ADF1/3411, PIN1 reference 81;
  the active authentication key reference is read from PrKD, not guessed.
- Both require a P-384 ECDH authentication certificate with key-agreement usage.
  RSA, brainpool and unknown/ambiguous key profiles are rejected.
- Existing CDOC 1.1 and CDOC 2.0 client key-unwrapping code is reused. This does
  not add new CDOC2 server, Mobile-ID or Smart-ID contracts.
- NFC uses PC/SC, PACE and secure messaging directly; it does not require a
  PKCS#11 module. Contact operations retain the existing PKCS#11 implementation.

The APDU profiles follow the [IDEMIA ID1 Developer Guide 2018](https://www.id.ee/wp-content/uploads/2020/10/id1developerguide.pdf),
[IDEMIA ID1 Developer Guide 2025](https://www.id.ee/wp-content/uploads/2026/03/id1developerguide2025.pdf)
and [Thales developer guide release 1.33](https://www.id.ee/wp-content/uploads/2025/11/ged24035_015_tdc_est_eid_developer_guide.pdf).

## Credentials and card sessions

Reading the public recipient certificate requires CAN/PACE, but no PIN.
Encryption uses that public certificate through the existing application flow;
the card does not encrypt the file itself.

For decryption, the native reader verifies the actual card certificate against
the challenge and constructs the PIN dialog identity from that certificate.
The existing native 60-second PIN1 dialog is used. It is opened outside the
PC/SC transaction. After confirmation, the reader opens a new PACE session and
checks the certificate again before the single PIN verification/private-key
operation. Keep the same card on the reader until completion. This deliberately
uses separate card sessions; reader activation sounds can occur more than once.

PIN1 never travels through the web page. CAN is retained only for the active
operation and is not sent to the Digisign backend or saved as a user preference.
Existing origin policy and per-operation extension consent also apply to NFC.
Card swaps, unsupported points/profiles and PIN cancellation/timeout fail the
operation; PIN and PACE failures are not automatically retried. Raw secure
messaging APDUs and response bytes are not written to diagnostic logs.

## Validation and manual acceptance

Run with `GOWORK=off`: `go test ./...` and `go vet ./...` (CGO/PCSC build tools
are required). Tests cover both profiles, independent Go ECDH round trips for
CDOC 1.1/2.0, APDU/PIN encoding, key selection, certificate swaps, PIN timeout,
wrong-PIN no-retry behavior and bounded secure-message certificate reads.
These tests use simulated cards and do not establish hardware compatibility.

Direct real-card certificate reading and CDOC v1 decryption passed on an IDEMIA
NFC card on 2026-10-31; see [direct test instructions](DIRECT-NFC-TEST.md).
The user subsequently confirmed NFC CDOC v1 decryption in Digisign with the
installed fix. Other card profiles, CDOC2 and the encryption round trip still
require the acceptance below:

1. Update the native package and reload the matching extension; reload Digisign.
2. In encryption, choose **Lisa saaja ID-kaardilt > NFC**, enter that card's CAN
   and read its certificate. No PIN prompt should appear. Encrypt a small test
   file for this recipient using the application's supported format.
3. In decryption, select that container, choose **NFC**, enter CAN and start.
   Confirm the site permission when requested, then enter PIN1 in the native
   dialog. Expect one PIN prompt and verify the resulting file contents.
4. Repeat with a pre-existing CDOC v1 container for the same card, then confirm
   contact decryption still works. Check cancellation and the 60-second timeout.
5. Repeat in Chrome and Firefox. Do not deliberately submit an incorrect PIN;
   automated tests cover that case without consuming real card attempts.

Release package: EID-PD 2.1.5, extension 2.1.5, web SDK 1.0.2.
The reader pins libopenpace-go 1.1.3. Its case-1 MAC padding correction resolves
SELECT MF SW=6988; profile-specific MF/file selection also replaces the earlier
incorrect SELECT parameters and hard-coded AID prerequisite (SW=6A86/6A82).
Independent OpenSSL AES-128/256 vectors cover the MAC correction.

This is a development package. Store publication and code signing are separate.
Real-card Thales, CDOC2 and the full NFC encryption round trip remain pending.

## Acceptance update 2026-10-31

The user confirmed successful local CDOC2 NFC decryption in Digisign with workstation package 2.1.8 (reader 2.1.6, extension 2.1.6, libcdoc-go 2.0.7). This supersedes the generic CDOC2-pending statements above for this tested path only; it does not establish other card profiles or the encryption round trip.

The current workflow establishes three NFC sessions: certificate discovery, card identity verification before the native PIN prompt, and card revalidation plus key operation after PIN entry. Status checks do not connect to the card. The observed three pairs of beeps are consistent with this flow; they are not three PIN attempts. Fewer sessions have not been implemented or validated. Card-swap, PIN-timeout and no-PIN-retry regression tests pass.
