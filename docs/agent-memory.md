# Agent Memory

### Mistake: ReadPersonalData (Contact mode) fails when multiple readers are connected
**Problem**: `ReadPersonalData` connected to the FIRST reader that accepted a connection
and tried all strategies only on that reader. If that reader's card failed (due to strategy
order differences), it gave up without trying other readers. Meanwhile `ReadPersonalDataFromAllReaders`
(`readPersonalDataFromCard`) successfully read the same card because it has a more comprehensive
fallback chain (tries ALL manufacturer strategies after standard approach fails).

**Fix**: Refactored `ReadPersonalData` to iterate ALL contact readers. For each reader that
connects, it calls `readPersonalDataFromCard` (same function used by All Readers mode).
If reading fails, it moves to the next reader. Only returns error after all readers are exhausted.

### Mistake: ReadPersonalDataFromAllReaders hangs with multiple readers
**Problem**: When using `--all-readers` flag, the program hangs indefinitely because:
1. NFC/contactless readers (PICC/SAM) are not skipped — `Connect` or APDU transmit can block without PACE
2. No timeout protection exists, so a stuck reader blocks the entire operation
3. Missing error code checks for `SCARD_W_UNRESPONSIVE_CARD` and `error(80100069)`

**Fix**: Applied three-part fix in `reader_multi.go`:
1. Skip NFC readers (PICC/SAM) — consistent with `ReadPersonalData` single-reader behavior
2. Added `readFromReaderWithTimeout` using goroutine + `time.After` (15s per reader)
3. Added missing SCard error codes for unresponsive cards

### Pattern: Go PC/SC multi-reader iteration
When iterating over PC/SC readers on Windows:
- Always filter out NFC/contactless interfaces (PICC, SAM) unless NFC mode is explicitly enabled
- Always add per-reader timeouts — `SCardConnect` can block indefinitely on certain reader states
- Check all relevant SCard error codes: `SCARD_E_NO_SMARTCARD`, `SCARD_W_REMOVED_CARD`, `SCARD_W_UNRESPONSIVE_CARD`, `error(8010000C)`, `error(80100069)`

### Mistake: PACE protocol OID constants were shifted by one category
**Problem**: In `libopenpace-go/pace/types.go`, the OID constant labels and values were mismatched:
- `OID_PACE_ECDH_GM_*` had `.1.x` values (actually DH-GM per BSI TR-03110)
- `OID_PACE_ECDH_IM_*` had `.2.x` values (actually ECDH-GM)
- `OID_PACE_DH_GM_*` had `.3.x` values (actually DH-IM)
- `OID_PACE_DH_IM_*` had `.4.x` values (actually ECDH-IM)

**Fix**: Per BSI TR-03110: id-PACE = 0.4.0.127.0.7.2.2.4, sub-components: .1=DH-GM, .2=ECDH-GM, .3=DH-IM, .4=ECDH-IM. Corrected all constant values.

### Mistake: PACE General Authenticate CLA byte was 0x00 instead of 0x10
**Problem**: GA Steps 1-3 used CLA=0x00. Per BSI TR-03110, GA Steps 1-3 require CLA=0x10 (command chaining). Only GA Step 4 (final) uses CLA=0x00.

**Fix**: Changed GA Steps 1-3 to CLA=0x10, kept Step 4 at CLA=0x00.

### Mistake: Generic Mapping not implemented for ECDH
**Problem**: GA Step 2 sent no mapping data (nil) and didn't generate a mapping keypair. Generic Mapping requires the terminal to generate an ephemeral mapping keypair and exchange public keys with the card.

**Fix**: Added `generateMappingKeypair()` in protocol.go. GA Step 2 now sends terminal's mapping public key (Tag 0x81) and parses card's response (Tag 0x82). `mapGeneratorGM_ECDH` computes G' = s*G + H where H = ECDH(mappingPriv, cardMappingPub).

### Mistake: PACE auth token format was wrong
**Problem**: Auth token was computed as MAC(K_mac, ownPK || peerPK). Per BSI TR-03110, it should be MAC_8(K_mac, 7F49{06 <OID> 86 <peerPK>}) — a TLV-encoded structure with the PACE OID and the peer's public key only.

**Fix**: Implemented BSI TR-03110 auth token format in step3cd.go with proper TLV encoding and 8-byte MAC truncation.

### Mistake: Wrong ParamID → curve mapping (ParamID 12 was brainpoolP256r1, should be secp256r1)
**Problem**: `setupDomainParameters` mapped ParamID 12 to brainpoolP256r1. Per BSI TR-03110 Part 3,
standardized domain parameters: ParamID 12 = secp256r1 (NIST P-256), ParamID 13 = brainpoolP256r1.
This caused GA Step 2 to fail (SW=6A80) because the mapping public key was on the wrong curve.

**Fix**: Corrected all ParamID mappings per BSI TR-03110 Part 3, Table D.1:
 8=secp192r1, 9=brainpoolP192r1, 10=secp224r1, 11=brainpoolP224r1,
 12=secp256r1, 13=brainpoolP256r1, 14=brainpoolP320r1,
 15=secp384r1, 16=brainpoolP384r1, 17=brainpoolP512r1, 18=secp521r1.
Also switched all curves to use `elliptic.Curve` interface consistently (not `ecdh.Curve`).

### Pattern: Estonian ID card PACE parameters
- **IDEMIA cards**: ECDH-GM-AES256 (OID 0.4.0.127.0.7.2.2.4.2.4), secp256r1/NIST P-256 (ParamID=12)
- **Thales cards**: ECDH-GM-AES256 (same OID), brainpoolP384r1 (ParamID=16), 16-byte nonce (not 32)
- PerformPACE reads EF.CardAccess and overrides defaults, so the initial OID/ParamID are fallbacks only

### Mistake: Thales NFC read returned empty files after PACE+SM succeeded
**Problem**: On Thales cards, personal data files (5001-5009) are NOT under DF 5000.
DF 5000 does not exist on Thales cards (SELECT with P1=01 returns 6A86, P1=00 returns 6A82).
The personal data files are under `DF.DocumentData` (FID DFDD), which is a subdirectory
of the eID application (PKCS-15 AID). The eID application root contains PKCS-15 directory
files with OVERLAPPING FIDs (5001=PrKD, 5003=CD, 5005=DCOD, 5006=AOD). Selecting PKCS-15
AID without then selecting DFDD reads these PKCS-15 objects instead of personal data.

**Fix (initial, incorrect)**: Tried always selecting DF 5000 first. This broke Thales
because DF 5000 doesn't exist on Thales. The code fell through to MF context where
files 5001-5009 also don't exist (6A82).

**Fix (correct)**: Two-path approach in `reader_nfc.go`:
1. Try SELECT DF 5000 (P1=01) → works for IDEMIA
2. If fails, SELECT PKCS-15 AID + SELECT DFDD by path (P1=08) → works for Thales
   Fallback: SELECT DF.DocumentData by AID "Document Data" (446F63756D656E742044617461)

### Pattern: Estonian ID card file structure for NFC personal data
- **IDEMIA cards**: Personal data files (5001-5009) are children of **DF 5000** under MF
  - SELECT DF 5000 (P1=01, P2=0C) → then SELECT EF 500x (P1=02) + READ BINARY
- **Thales cards**: Personal data files (5001-5009) are under **DF.DocumentData** (FID DFDD)
  - DF.DocumentData is a subdirectory of the eID application (PKCS-15 AID)
  - AID of DF.DocumentData: "Document Data" = 446F63756D656E742044617461
  - Navigation: SELECT PKCS-15 AID (P1=04) → SELECT DFDD (P1=08, data=DFDD) → SELECT EF 500x (P1=02)
- **DF 5000 does NOT exist on Thales cards** (returns 6A86 or 6A82)
- **PKCS-15 root has overlapping FIDs**: 5001=PrKD, 5003=CD — NOT personal data
- Thales P1 restrictions: P1=01 not supported (6A86), P1=00/02/04/08 work

### Pattern: CosmoX vs CosmoV8 SELECT command differences
- CosmoV8 (older Thales): Did not check P1 parameter strictly; P1=01 for EF selection worked
- CosmoX (newer Thales/IDEMIA): Checks P1 strictly per IAS-ECC standards
- Always use correct P1: 00=by FID, 01=child DF, 02=EF under current DF, 04=by AID, 08=by path from MF

### Pattern: Brainpool curves in Go
- Go's `crypto/ecdh` package only supports NIST curves (P-256, P-384, P-521)
- Brainpool curves must use `crypto/elliptic.Curve` interface via `go.gnd.pw/crypto/brainpool`
- Key operations (ScalarMult, ScalarBaseMult, Add) work via elliptic.Curve
- Private keys are `[]byte` (not `*ecdh.PrivateKey`), public keys are uncompressed `[]byte` (04||X||Y)
- When using a mapped generator G', compute PK = priv * G' via `ScalarMult(G'.X, G'.Y, priv)` — NOT `ScalarBaseMult`

### Mistake: PACE KDF used SHA-512 instead of SHA-256 for AES-256
**Problem**: `GetHashAlgorithm(32)` returned `HashSHA512` for key lengths > 24 bytes.
Per BSI TR-03110 Part 3, Table A.1: AES-256 (≤32 bytes) requires SHA-256, not SHA-512.
This caused `KDF_PI` to derive the wrong key, making nonce decryption produce garbage.

**Fix**: Changed `GetHashAlgorithm` to return `HashSHA256` for key lengths ≤ 32.

### Mistake: PACE nonce decryption applied PKCS#7 unpadding
**Problem**: `AESDecryptECB` always applied PKCS#7 unpadding after decryption.
PACE nonce encryption is raw AES without any padding.

**Fix**: Added `AESDecryptECBRaw` function that decrypts without unpadding.

### Mistake: PACE nonce decryption used AES-ECB instead of AES-CBC
**Problem**: Nonce decryption used AES-ECB mode. Per OpenPACE reference implementation
(`cipher_no_pad` with `EVP_aes_256_cbc()`), the cipher mode is AES-CBC with zero IV.
For a single 16-byte block (AES-128/192), ECB and CBC with zero IV are equivalent.
But IDEMIA cards send a 32-byte nonce for AES-256 (key-length per BSI TR-03110
Section A.3.4: "s MUST be at least key_length bits"). With ECB, the second block
decrypts incorrectly (CBC XORs each block with the previous ciphertext block).
Previous "truncate to block size" fix also failed because it discarded the second
half of the nonce — the card uses the full 32-byte nonce in its computations.

**Fix**: Added `AESDecryptCBCRaw(key, iv, ciphertext)` for CBC decryption without
padding removal. Updated step2.go to use CBC with zero IV and use the full
decrypted nonce without truncation.

### Pattern: PACE cipher mode is CBC, not ECB
OpenPACE uses AES-CBC (not ECB) for all PACE symmetric operations. The IV is zero
(NULL) initially. For single-block operations (16 bytes), CBC with zero IV equals ECB.
For multi-block operations (e.g. 32-byte nonce for AES-256), CBC chaining is required.
Always use `AESDecryptCBCRaw(key, zeroIV, ciphertext)` for PACE nonce decryption.

### Mistake: SM wrapping had 5 critical bugs causing 6988/6985 errors
**Problem**: After PACE succeeded, all SM-protected APDUs failed:
1. **Double padding**: `ISO9797Method2Pad` + `AESEncryptCBC` (which adds PKCS#7) → encrypted data was 2x expected size
2. **Wrong IV**: Encryption used zero IV instead of `AES_ECB_Enc(K_enc, SSC)`
3. **MAC input wrong**: SSC was 8 bytes (should be block-size=16) and header was not ISO-padded
4. **Missing Le'**: SM APDU didn't append Le'=0x00 at end for card to return response
5. **SSC too short**: SSC stored as 8 bytes, but should be block-size (16 for AES) per OpenPACE `encode_ssc`

**Fix**: Rewrote SM `WrapCommand`, `encryptData`, `decryptData` in `sm/sm.go`:
- Added `AESEncryptCBCRaw` and `AESEncryptECBBlock` to `crypto/aes.go`
- IV computation: `computeIV()` → `AES_ECB_Enc(K_enc, SSC)` per OpenPACE `update_iv`
- MAC input: `CMAC(K_mac, SSC_16bytes || ISO_pad(header, 16) || DO87 || DO97)`
- Encryption: `AESEncryptCBCRaw(K_enc, computedIV, ISO_padded_data)` — no PKCS#7
- SSC stored as block-size bytes (left-padded from 8-byte initial value)
- Appended Le'=0x00 to all SM-wrapped APDUs

### Mistake: SM MAC input was NOT ISO-padded before CMAC
**Problem**: The MAC input (SSC || padded_header || DOs) was passed directly to CMAC
without final ISO 9797-1 method 2 padding. CMAC handles incomplete last blocks with K2
subkey, but ISO 7816-4 requires the entire MAC input to be padded to block boundary
before CMAC, so CMAC always sees a complete input and uses K1 subkey. These produce
different MAC values. Confirmed by OpenSCDP IsoSecureChannel test vectors showing
explicit ISO padding of MAC input before AES_CMAC.

**Fix**: Apply `ISO9797Method2Pad(macInput, blockSize)` to the full MAC input before
passing to CMAC, in both `WrapCommand` and `UnwrapResponse`.

### Pattern: BSI TR-03110 Secure Messaging structure
For SM after PACE:
- **Encryption IV** = AES_ECB_Enc(K_enc, SSC), NOT zero. SSC encoded as block-size bytes.
- **MAC input** = ISO_pad(SSC(16) || ISO_pad(header, 16) || DO87 || DO97, 16). MUST be ISO-padded before CMAC so CMAC uses K1 subkey.
- **Data encryption**: ISO 9797-1 method 2 pad, then AES-CBC with computed IV. NO PKCS#7.
- **APDU format**: CLA' INS P1 P2 Lc' [DO87] [DO97] DO8E Le'(0x00)
- **Response MAC input**: ISO_pad(SSC(16) || DOs_except_DO8E, 16). Same ISO-padding rule.

### Note: Go binary location on this machine
- Go is installed at `C:\Program Files\Go\bin\go.exe` (v1.26.0)
- Both projects use `go 1.26.0` in go.mod
- Build script (build.ps1) uses `C:\Program Files\Go` as GOROOT

### Mistake: Go modules v2+ require /v2 module path suffix
**Problem**: Tried to use `v2.0.0` version for `github.com/aarepuusepp/libopenpace-go` module,
but Go modules with major version ≥2 require a `/v2` suffix in the module path (per Go module
compatibility rules). Using `v2.0.0` without the path suffix causes `go mod tidy` to fail with:
"version v2.0.0 invalid: should be v0 or v1, not v2".

**Fix**: Use `v1.x.y` versioning for modules without the `/vN` path suffix.
If truly needed, create a new module path like `github.com/user/module/v2` with all import paths updated.

### Pattern: ACR1252 NFC reader power-on failures
The ACR1252 reader sometimes fails to power on the card on first attempt:
- Windows Event Viewer shows: "The smart card is not responding to a reset" (POWER command)
- Followed by: "The device does not recognize the command" (GET_ATTRIBUTE)
- Root cause: card RF field not fully stable when SCardConnect issues SCARD_COLD_RESET
- **Critical**: Each `SCardConnect` call generates 3-4 Event 610 entries because Windows
  internally retries the POWER command. Multiple connect retries per attempt multiply noise.

**Fix (single connect + read-level retry):**
1. **Connection layer** (`nfc.ConnectAny`): Uses `SCardGetStatusChange` to wait for card presence
   (up to 5s), then makes exactly ONE `SCardConnect` call (with 500ms stabilization delay).
   No connect-level retries — this minimizes Event 610 entries per attempt.
2. **Read layer** (`ReadPersonalDataNFCWithRetry`): Retries full connection+PACE+read cycle with
   progressive delays (1s, 2s, 3s). Default 3 attempts. Uses `UnpowerCard` disposition on disconnect
   to avoid additional POWER IOCTLs.

### Pattern: Project version locations
- **libopenpace-go version**: README.md, CHANGELOG.md, go.mod (module version in tags)
- **eid-pd-reader version**: `cmd/eid-pd-reader/main.go` const `Version`, README.md, go.mod dependency
- **eid-pd-extension version**: `extension/manifest.json`, `extension/manifest-firefox.json`, `extension/page-api.js` (version property), `extension/popup/popup.html` (footer span)
- **Build tooling**: `build.ps1` (GOROOT path), `.github/workflows/build.yml` (go-version)
- **CI actions**: checkout@v4, setup-go@v5, upload-artifact@v4

### Pattern: Native messaging host name and registry
- **Current host name**: `ee.aarepuusepp.eid_pd_reader` (changed from `ee.aarepuusepp.eid_helper`)
- **Reader manifest files**: `ee.aarepuusepp.eid_pd_reader.json` (Chrome/Edge), `ee.aarepuusepp.eid_pd_reader.firefox.json`
- **Registry keys**: `HKCU\Software\{Google\Chrome|Microsoft\Edge|Mozilla}\NativeMessagingHosts\ee.aarepuusepp.eid_pd_reader`
- **Install dir**: `C:\Program Files\eid-pd-reader\`
- **Extension ops scripts** (register-*.reg, setup-chrome-extension.ps1, etc.) must match the host name exactly

### Pattern: Native messaging commands (reader <-> extension)
| Command | Request fields | Response |
|---------|---------------|----------|
| `ping` | none | `{ok, timestamp, duration}` |
| `version` | none | `{ok, timestamp, duration, version, name, data}` |
| `getPersonalData` / `readPersonalData` | `fids` (optional) | `{ok, timestamp, duration, data: PersonalData}` |
| `readPersonalDataNFC` | `can` (required), `retry` (optional, default 3) | same as above |
| `readAllReaders` | `fids` (optional) | `{ok, timestamp, duration, results: [{reader, duration, data, error}]}` |
| `capabilities` | none | `{ok, timestamp, duration, data: {nfcSupported, version}}` |

### Mistake: Firefox AMO signing validation failures
**Problem**: Mozilla AMO validator rejected the extension with 2 errors and 1 warning:
1. `data_collection_permissions` had wrong format (`data_collection: false` boolean) — requires `required` property as array
2. `developer.url` linked to `addons.mozilla.org` which is forbidden by AMO policy
3. `innerHTML` assignment in popup.js flagged as unsafe

**Fix**:
1. Changed `data_collection_permissions` to `{ "required": ["none"] }` per Mozilla docs
2. Changed `developer.url` from AMO profile to GitHub profile URL
3. Replaced `innerHTML` construction with DOM API (`createElement`/`textContent`/`appendChild`)

### Pattern: Firefox AMO manifest requirements
- `data_collection_permissions.required` must be an array (e.g. `["none"]` or `["personallyIdentifyingInfo"]`)
- `data_collection_permissions.optional` is optional, also an array
- Never link to `addons.mozilla.org` in `homepage_url` or `developer.url`
- Avoid `innerHTML` in extension code — use DOM API (`createElement`, `textContent`) instead
- `data_collection_permissions` key requires Firefox 140+; warning is non-blocking if `strict_min_version` is lower

### Pattern: ops/ scripts must resolve paths relative to script location
- All scripts in `ops/` directory use `Push-Location`/`Pop-Location` to work from any CWD
- Pattern: `$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path; $ProjectRoot = Split-Path -Parent $ScriptDir; Push-Location $ProjectRoot`
- Always `Pop-Location` at the end of the script

### Pattern: Extension message types (popup/page <-> background)
- `EID_PD_READ` — contact mode read
- `EID_PD_READ_NFC` — NFC read (fields: `can`, `retry`)
- `EID_PD_READ_ALL` — multi-reader read
- `EID_PD_STATUS` — ping/status check
- `EID_PD_VERSION` — get version
- `EID_PD_CAPABILITIES` — get capabilities
- `EID_PD_CLEAR_CACHE` — clear cache

### Pattern: eid-pd reader response date formats
- `birthDate`: `"DD MM YYYY COUNTRY"` (e.g. `"08 01 1980 EST"`) — trailing country code
- `expiryDate`: `"DD MM YYYY"` (e.g. `"19 01 2028"`) — no suffix
- `issueDate`: `"DD MM YYYY ISSUER"` (e.g. `"19 01 2023 PPA/PBGB"`) — trailing issuer identifier
- Parser must use only the first 3 whitespace-separated tokens (DD, MM, YYYY) and ignore any trailing text

### Pattern: Vue 3 + TypeScript global type declarations
- For augmenting `Window` interface in Vue 3 + Vite projects, place `.d.ts` files in `src/types/` or root `env.d.ts`
- The `.d.ts` file must NOT have `export {}` if types need to be ambient (script-scoped)
- `interface Window { ... }` merges with the global Window type when file is ambient
- `vue-tsc --noEmit` is the authoritative type checker; Vetur/ts-plugin may show stale errors until editor reloads

### Pattern: Spring Boot 3 -> 4 migration (package relocations)
- `@MockBean` removed; replace with `@MockitoBean` from `org.springframework.test.context.bean.override.mockito`
- `@DataJpaTest` moved from `org.springframework.boot.test.autoconfigure.orm.jpa` to `org.springframework.boot.data.jpa.test.autoconfigure`
- `TestEntityManager` moved to `org.springframework.boot.jpa.test.autoconfigure` (NOT data.jpa)
- `@WebMvcTest`/`@AutoConfigureMockMvc` moved to `org.springframework.boot.webmvc.test.autoconfigure`
- `FlywayAutoConfiguration` moved to `org.springframework.boot.flyway.autoconfigure`
- `@EntityScan` moved from `org.springframework.boot.autoconfigure.domain` to `org.springframework.boot.persistence.autoconfigure`
- `TestRestTemplate` moved to `org.springframework.boot.resttestclient`
- Modular test starters required: `spring-boot-webmvc-test`, `spring-boot-data-jpa-test`, `spring-boot-flyway`, `spring-boot-resttestclient`
- Jackson 3: `com.fasterxml.jackson.databind` -> `tools.jackson.databind`, `com.fasterxml.jackson.datatype` -> `tools.jackson.datatype`, but `com.fasterxml.jackson.annotation` STAYS THE SAME
- Jackson 3: `JavaTimeModule` is auto-registered; remove explicit registration. Use `JsonMapper.builder()` instead of `new ObjectMapper()`
- Gradle 9: ben-manes versions plugin `componentSelection` API changed; use `rejectVersionIf` instead
- Gradle distribution URL uses full version: `gradle-9.4.0-bin.zip` (not `gradle-9.4-bin.zip`)
- TestContainers: keep all modules at same version (1.21.4); don't mix 1.x modules with 2.x core
- `@Temporal(TemporalType.TIMESTAMP)` is deprecated in Jakarta EE 11; remove annotation and migrate `java.util.Date` fields to `java.time.Instant`. For `Instant` fields, `@Temporal` was never needed. For `Date` fields, migrate: `new Date()` -> `Instant.now()`, `date.getTime()` -> `instant.toEpochMilli()`, `new Date(ms)` -> `Instant.ofEpochMilli(ms)`, `SimpleDateFormat` -> `DateTimeFormatter.withZone(ZoneOffset.UTC)`
