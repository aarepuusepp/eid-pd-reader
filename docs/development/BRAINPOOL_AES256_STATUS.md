> Historical development record. Use [current reader README](../../README.md) for supported setup; these notes are not current installation/API instructions.

# brainpoolP256r1 + AES-256 Implementation Status

## Summary

We've made significant progress implementing brainpool and AES-256 support, but there's one remaining issue preventing the Thales test card from working.

## ✅ Completed

### 1. brainpoolP256r1 Curve Support
- ✅ Added `go.gnd.pw/crypto/brainpool` dependency
- ✅ Added `EllipticCurve` field to `DomainParameters`
- ✅ Implemented ParamID=16 → brainpoolP256r1 mapping
- ✅ Created `GenerateEllipticKey()` helper for brainpool curves
- ✅ Created `ComputeEllipticSharedSecret()` for brainpool ECDH
- ✅ Updated `step3b.go` to handle both `ecdh.Curve` and `elliptic.Curve`

### 2. Password Reference Testing
- ✅ Added MRZ (0x01) password reference for eMRTD
- ✅ Tested all three password references (0x01, 0x02, 0x03)
- ✅ Confirmed issue is NOT password reference (all return SW=6985)

### 3. CardAccess Parsing
- ✅ Parse OID from CardAccess
- ✅ Parse ParamID from CardAccess
- ✅ Use parsed values in MSE:Set AT command

## ❌ Remaining Issue

### Protocol Initialization Mismatch

**Problem**: The PACE protocol is initialized with hardcoded values in `pace_pure.go`:
```go
smCtx, err := pace.PerformPACE(
    cardAdapter,
    secret,
    pace.OID_PACE_ECDH_IM_AES128,  // ← Hardcoded AES-128
    8,                               // ← Hardcoded P-256
)
```

But the card requires:
- **OID**: `0.4.0.127.0.7.2.2.4.2.4` (ECDH IM AES-**256**)
- **ParamID**: 16 (brainpoolP256r1)

**Impact**: 
- KDF functions generate 16-byte keys instead of 32-byte keys
- Protocol uses AES-128 instead of AES-256
- Card detects mismatch and returns SW=6985

### Solution Required

We need to:

1. **Parse CardAccess BEFORE calling PerformPACE**:
   ```go
   // Read CardAccess
   cardAccess := readCardAccess(card)
   paceInfo := parseCardAccessPACEInfos(cardAccess)[0]
   
   // Determine OID and ParamID from CardAccess
   protocol := determineProtocolFromOID(paceInfo.oidDER)
   paramID := int(paceInfo.paramID)
   
   // Call PerformPACE with correct values
   smCtx, err := pace.PerformPACE(cardAdapter, secret, protocol, paramID)
   ```

2. **Update KDF to support AES-256**:
   ```go
   // In KDF functions
   func KDF_Enc(secret []byte, protocol ProtocolOID) ([]byte, error) {
       keyLen := protocol.GetKeyLength()  // 16 for AES-128, 32 for AES-256
       // ... derive keyLen bytes
   }
   ```

3. **Update SM to support AES-256**:
   ```go
   // In Secure Messaging
   func NewContext(kEnc, kMac []byte) *Context {
       cipherSuite := determineCipherSuite(len(kEnc))  // AES-128 vs AES-256
       // ...
   }
   ```

## Test Results

### Current Behavior
```
[PACE] Protocol: ECDH IM AES-128, Curve: P-256  ← Wrong!
[PACE] Using brainpoolP256r1 curve (ParamID=16)  ← Correct!
[PACE] MSE:Set AT data: 800C060A04007F00070202040204830102840110  ← Correct!
MSE:Set AT → SW=6985  ← Crypto mismatch!
```

### Expected Behavior (after fix)
```
[PACE] Protocol: ECDH IM AES-256, Curve: brainpoolP256r1
[PACE] Using brainpoolP256r1 curve (ParamID=16)
[PACE] MSE:Set AT data: 800C060A04007F00070202040204830102840110
MSE:Set AT → SW=9000  ← Success!
[PACE] PACE established successfully
```

## Implementation Plan

### Phase 1: Parse CardAccess in pace_pure.go (2-3 hours)
1. Move CardAccess reading logic before `PerformPACE` call
2. Parse OID and determine protocol variant
3. Map OID bytes to `ProtocolOID` constant
4. Pass correct protocol and paramID to `PerformPACE`

### Phase 2: Add GetKeyLength() method (1 hour)
1. Add method to `ProtocolOID` to return key length
2. Update KDF functions to use dynamic key length
3. Test with both AES-128 and AES-256

### Phase 3: Update SM for AES-256 (1-2 hours)
1. Add cipher suite detection based on key length
2. Update encryption/decryption to handle 32-byte keys
3. CMAC already supports 256-bit keys (no changes needed)

### Phase 4: Testing (1-2 hours)
1. Test with Thales card
2. Verify MSE:Set AT returns SW=9000
3. Complete PACE protocol
4. Read personal data

**Total Estimated Time**: 5-8 hours

## Quick Fix for Testing

If you want to test quickly, you can manually change `pace_pure.go`:

```go
// Change from:
pace.OID_PACE_ECDH_IM_AES128,
8, // P-256

// To:
pace.OID_PACE_ECDH_IM_AES256,
16, // brainpoolP256r1
```

This won't fully work yet (KDF and SM still need updates), but it will show if MSE:Set AT accepts the correct parameters.

## Files Modified

### libopenpace-go
- `pace/types.go`: Added `EllipticCurve` field
- `pace/api.go`: Added brainpool import and ParamID=16 mapping
- `pace/step3b.go`: Added `generateEphemeralKeyElliptic()`
- `pace/protocol.go`: Added MRZ password reference for eMRTD
- `crypto/ecdh.go`: Added `GenerateEllipticKey()` and `ComputeEllipticSharedSecret()`
- `go.mod`: Added `go.gnd.pw/crypto` dependency

### eid-pd-reader
- None yet (changes needed in `native/nfc/pace_pure.go`)

## References

- RFC 5639: Elliptic Curve Cryptography (ECC) Brainpool Standard Curves
- BSI TR-03110 v2.21: Advanced Security Mechanisms
- go.gnd.pw/crypto/brainpool: Brainpool curve implementation

## Next Steps

1. **Option A**: Complete AES-256 implementation (5-8 hours)
2. **Option B**: Test with real Estonian ID card (may use AES-128 + P-256 and work immediately)

**Recommendation**: Try Option B first - test with a real card before spending time on AES-256.

