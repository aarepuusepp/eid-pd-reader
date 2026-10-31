# Reader 2.1.4

Updates the pinned libopenpace-go snapshot to 1.1.2:

- Correct AES-128 SHA-1 KDF and secret/nonce/counter ordering.
- Reject incomplete PACE Integrated Mapping, 3DES and unknown protocol OIDs,
  including variants advertised by a card before authentication.
- Preserve AES-128/192/256 Generic Mapping and AES secure messaging.
- Include fixed independent KDF vectors and protocol rejection tests.

The native messaging protocol remains version 1. Browser extensions 2.1.3
remain compatible. This is a reader source release; it does not replace the
already installed workstation package 2.1.3. A new installer must be built
and validated separately before deploying this reader.

Run both the reader tests and the nested source snapshot tests; Go does not
include nested modules in the root ./... test pattern:

    go test -count=1 ./...
    go vet ./...
    cd third_party/libopenpace-go
    go test -count=1 ./...
    go vet ./...

Real NFC/PACE interoperability is not established by contact-card reading or
CDOC v1 PKCS#11 decryption. This patch has not been tested on an NFC card.

## Validation on 2026-10-30

Passed on Windows x64 with Go 1.27.3 and MSYS2 UCRT64 GCC:
- Reader: go test -mod=readonly -p 2 -count=1 ./... and go vet -mod=readonly ./...
- Nested libopenpace-go snapshot: full tests and go vet.
- Native Windows CGO build with -mod=readonly -trimpath; --version reports 2.1.4.
- All 42 copied source, test, fixture and documentation files match the pinned
  upstream commit byte for byte (README.snapshot.md is reader-specific metadata).
- Upstream library: race tests, 20 independent KDF vectors verified with Python
  and OpenSSL, old-defect mutation checks. See the copied VALIDATION-1.1.2.md.

No hardware card operation or browser decryption was initiated during these tests.

## Local NFC development follow-up

The subsequent 2.1.4 development rebuild holds a PC/SC transaction across PACE
and all secure-messaging reads. This prevents other card clients from changing
the selected application or authentication state between APDUs. End the
transaction with UnpowerCard, then disconnect with LeaveCard; ResetCard at
transaction end failed on the local ACR1252 driver.

Four regression tests cover lock failure, protected reads, cleanup failure and
panic cleanup. A real ACR1252 PICC / test-card run completed PACE and protected
file reads with AES-256 / brainpoolP384r1. The card returned empty personal-data
fields, so this is not proof of populated Visitor form fields. An incorrect CAN
was separately rejected at final PACE authentication with SW=6300.

This local rebuild does not change the pinned libopenpace-go 1.1.2 snapshot.
The rebuilt native host was installed as 2.1.4-ff0fd34dae02. Chrome / Visitor
completed NFC reading and displayed the empty-test-card notice. Contact reading
populated the personal and document fields from the personalized card. NFC
reading of populated personal data and CDOC1 revalidation remain separate checks.

## NFC read optimization (local development rebuild)

The user confirmed that a personalized NFC test card returned personal data in
Visitor; a separate unpersonalized test card correctly returned empty fields.
Two beeps were reported together at the beginning of a read. Their exact hardware
cause has not been established; fewer software operations do not prove one beep.

- One SCardConnect per present reader per attempt; failed connections no longer
  fall through into a second insertion wait/connect loop. StateChanged is cleared
  when acknowledging PC/SC status events. The ACR1252 SAM slot is excluded.
- Only connection establishment failures may retry. PACE/authentication, secure
  messaging, read and cleanup failures return immediately, without repeating CAN
  authentication. The existing 500 ms stabilization delay remains.
- Read the nine documented personal-data files directly after selecting their
  directory. Remove six speculative GET DATA commands. Reject an unselected
  personal-data directory instead of reading overlapping PKCS-15 file IDs.
- Preserve the transaction across PACE and all reads, and UnpowerCard cleanup.
  Do not alter the reader's global buzzer or polling settings.
- Remove CAN/session-key values from PACE verbose output and personal field values
  from the NFC read debug messages.

Validation: reader full tests and go vet passed on Windows x64 / Go 1.27.3.
Eight additional regression tests cover connection counts, insertion wait, SAM
exclusion, retry boundaries, invalid CAN, populated/empty cards and directory
isolation. The simulated IDEMIA/Thales read uses 19/21 protected exchanges instead
of 25/27 (six fewer, excluding PACE and any length-correction responses). These
are deterministic test counts, not measured hardware latency. Hardware timing,
beep count and the optimized personalized-card read still need user confirmation.

The user confirmed the optimized build still fills the Visitor form, but two
beeps remain at the start. A CAN-free diagnostic containing a single SCardConnect
(and no Transmit calls) failed before BeginTransaction on the local ACR1252.
Windows Smartcard-Server event 610 recorded three IOCTL POWER reset failures and
a GET_ATTRIBUTE failure for that one connection. The same result occurred with
and without the existing 500 ms stabilization delay. This establishes a local
reader/driver connection issue independent of PACE, but does not by itself prove
which hardware event produces each audible beep.

The installed ACS driver was 4.2.8.0 (2018). ACS currently publishes installer
4.4.2.1, containing Windows x64 driver 4.4.2.0; the package explicitly supports
ACR1252U-A1/M1 and A2/M2. Its Setup.exe and x64 MSI Authenticode signatures were
verified as valid, signed by Advanced Card Systems Ltd. The old driver package
was exported under .local/acs-driver-backup-4.2.8.0 before a local upgrade.
Source: https://www.acs.com.hk/en/driver/342/acr1252u-usb-nfc-reader-iii-nfc-forum-certified-reader/
The vendor ReadMe requires unplugging the reader for installation and reconnecting
it afterward. This driver is a workstation prerequisite, not bundled in EID-PD.

Post-upgrade verification: both ACR1252 interfaces now use driver 4.4.2.0,
INF oem19.inf. Installation returned 0 and did not request a reboot. Read-only
reader controls returned LED/buzzer behavior 0x7F and PICC polling 0x8B (vendor
defaults); neither setting was changed. The POWER/connection failure remained.

In the final CAN-free lifecycle probe at 23:34:56-23:35:09 local time, the first
SCardConnect failed, one recovery attempt in the same PC/SC context succeeded,
and no APDU was transmitted. The user confirmed the same two beeps immediately
at the start. Thus duplicate personal-data reading or repeated PACE is not
required to reproduce the sound. The later cleanup stage could not be assessed:
Windows reported SCARD_W_REMOVED_CARD after the 12-second pause. No claim is made
that the exact firmware beep event or the initial POWER failure is fixed.
The reader's production behavior remains the tested optimized 2.1.4 build;
no buzzer/polling changes or PACE cleanup weakening were made to suppress sound.

## Manual acceptance update - 2026-10-31

The user confirmed personalized NFC reading in Visitor and contact-reader CDOC v1
 decryption in Digisign with Chrome after the ACS driver upgrade. The remaining
 two starting beeps are not evidence of duplicate data reading.

At that acceptance checkpoint, NFC support was personal-data reading with CAN/PACE. The decrypt command
handlers use GetDecryptCertificate / UnwrapTransportKey with PKCS#11 configuration;
 they do not establish a CAN/PACE session. NFC decryption is not implemented in
 this flow and was not accepted as supported. A manual NFC decrypt attempt returned
 CARD_NOT_FOUND from the card-presence check. Generic diagnostics only report
 presence and library availability, not operation-specific NFC decrypt readiness.

Results are user-reported hardware acceptance, not an automated end-to-end test.

## NFC crypto development and SELECT fix - 2026-10-31

The subsequent development build adds direct PACE/secure-messaging certificate
reading and CDOC key unwrapping; see [NFC crypto](docs/NFC-CRYPTO.md).
The first real-card attempt failed before PIN verification with SELECT SW=6A86.
The IAS-ECC application selection requested unsupported FCI (P2=00); the file
selection also used Thales absolute-path semantics on IDEMIA.

The correction uses P2=0C for IAS-ECC AID selection and explicit MF, child DF and
child EF selection, including reselecting the authentication DF before VERIFY.
Thales retains its own absolute-path profile and now includes the documented Le
on application SELECT. Strict, complete certificate-selection tests supplement
the token round-trip tests that previously skipped application/file selection.
All reader tests and go vet pass. Real-card NFC crypto acceptance remains pending.

The following real-card retest returned 6A82 at the hard-coded IAS-ECC MF AID.
That AID prerequisite has now been removed. The IDEMIA branch follows the
certificate-read sequence common to the 2018 and 2025 ID1 guides: case-1 SELECT
MF (`00 A4 00 0C`), child ADF1, child EF 3401. The same MF/ADF sequence is used
before PIN verification. Selection failures now identify the MF, DF or EF step.
Regression tests reject an extra AID probe and ensure selection/security errors
stop before PIN verification. NFC personal-data reading remains unchanged.

## Direct NFC hardware acceptance - 2026-10-31

The next SELECT MF failed with SW=6988, reproduced without a browser. The root
cause was in libopenpace-go secure messaging: a case-1 APDU's already padded
header was padded a second time despite having no DO87/DO97. The shared library
and its reader snapshot now use the case-1 MAC construction also used by OpenSC.
Independent OpenSSL AES-128 and AES-256 vectors exercise the exact protected
SELECT MF bytes. Data-bearing commands retain their existing MAC construction.

With that correction, the user's personalized IDEMIA NFC card successfully read
its certificate without PIN and decrypted a CDOC v1 fixture into two files using
the native PIN dialog. The new `cmd/eid-pd-card-test` tool ran directly against
PC/SC, without a browser or extension; only byte counts and SHA-256 hashes were
printed and plaintext was not saved. This is real-card validation of the direct
flow, not a browser or all-card/all-format acceptance result.
