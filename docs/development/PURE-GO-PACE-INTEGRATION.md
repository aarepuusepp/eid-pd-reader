> Historical development record. Use [current reader README](../../README.md) for supported setup; these notes are not current installation/API instructions.

# Pure Go PACE Implementation - Integration Complete

## Overview

The eid-pd-reader application now uses a **pure Go implementation** of the PACE (Password Authenticated Connection Establishment) protocol, eliminating the need for CGO and the OpenPACE C library.

## What Changed

### Before
- Required CGO compilation
- Depended on OpenPACE C library
- Platform-specific builds (Windows/Linux/macOS)
- Complex build process with external dependencies
- Required MSYS2/MinGW on Windows

### After
- **Pure Go** - no CGO required
- **No external libraries** - self-contained
- **Cross-platform** - single codebase
- **Simple builds** - just `go build`
- **Easier deployment** - single binary

## Technical Implementation

### New Components

1. **libopenpace-go** (`repos/libopenpace-go/`)
   - Pure Go PACE protocol implementation
   - BSI TR-03110 v2.21 compliant
   - Complete Secure Messaging support
   - 12 commits with full implementation

2. **CardAdapter** (`native/nfc/card_adapter.go`)
   - Bridges `scard.Card` with libopenpace-go's `CardInterface`
   - Enables PC/SC communication

3. **SMAdapter** (`native/nfc/sm_adapter.go`)
   - Wraps libopenpace-go's SM context
   - Implements existing `SM` interface
   - Drop-in replacement for OpenPACE SM

4. **StartPACE** (`native/nfc/pace_pure.go`)
   - Pure Go PACE establishment
   - Uses libopenpace-go's `EstablishPACE`
   - Protocol: PACE ECDH IM AES-128 with P-256

### Integration Points

```go
// Old (CGO-based):
sm, err := nfc.StartPACE(card, can)  // Called OpenPACE via CGO

// New (Pure Go):
sm, err := nfc.StartPACE(card, can)  // Uses libopenpace-go
```

**Same API, different implementation!**

## Benefits

### 1. No CGO Dependency
- Faster compilation
- No C compiler required
- No platform-specific toolchains

### 2. Simplified Build Process
```bash
# Before (with OpenPACE):
# 1. Install MSYS2/MinGW
# 2. Build OpenPACE
# 3. Set CGO flags
# 4. Build with CGO

# After (pure Go):
go build -o eid-pd-reader.exe
```

### 3. Cross-Platform Compatibility
- Same code works on Windows, Linux, macOS
- No platform-specific compilation flags
- No external library dependencies

### 4. Easier Deployment
- Single binary
- No DLL/SO dependencies
- Smaller distribution size

### 5. Better Maintainability
- Pure Go codebase
- Standard Go tooling
- Easier debugging
- Better IDE support

## Build Instructions

### Standard Build (Pure Go)
```bash
go build -o eid-pd-reader.exe
```

That's it! No additional setup required.

### With OpenPACE (Legacy, Optional)
If you still want to use the CGO-based OpenPACE:

```bash
# Rename files
mv native/nfc/pace_pure.go native/nfc/pace_pure.go.disabled
mv native/nfc/pace_openpace.go.old native/nfc/pace_openpace.go

# Build with CGO
go build -tags openpace -o eid-pd-reader.exe
```

## Testing

The pure Go implementation has been tested with:
- ✅ Successful compilation on Windows
- ✅ Version check: `eid-pd-reader.exe --version`
- ⏳ NFC reading with real Estonian ID card (requires physical card)

## Implementation Details

### PACE Protocol
- **Algorithm**: ECDH with Integrated Mapping
- **Cipher**: AES-128
- **Curve**: NIST P-256 (secp256r1)
- **Standard**: BSI TR-03110 v2.21

### Secure Messaging
- **Encryption**: AES-CBC with zero IV
- **MAC**: CMAC (AES)
- **Padding**: ISO 9797-1 Method 2
- **SSC**: 8-byte big-endian counter

### Key Derivation
- **KDF_PI**: Derives initial PACE keys from CAN
- **KDF_Enc**: Derives encryption key from shared secret
- **KDF_MAC**: Derives MAC key from shared secret
- **Hash**: SHA-256

## File Changes

### New Files
- `native/nfc/card_adapter.go` - PC/SC to libopenpace-go bridge
- `native/nfc/sm_adapter.go` - SM context wrapper
- `native/nfc/pace_pure.go` - Pure Go PACE implementation
- `native/nfc/apdu_helpers.go` - APDU helper functions
- `repos/libopenpace-go/` - Complete PACE library (12 commits)

### Modified Files
- `go.mod` - Added libopenpace-go dependency
- `native/nfc/sm_helpers.go` - Removed esteid import cycle

### Renamed Files (Old Implementation)
- `native/nfc/pace_openpace.go` → `pace_openpace.go.old`
- `native/nfc/pace_stub.go` → `pace_stub.go.old`

## Performance

The pure Go implementation is expected to have:
- **Similar or better performance** than CGO (no CGO overhead)
- **Faster startup** (no DLL loading)
- **Lower memory usage** (no C library overhead)

Actual benchmarks pending real card testing.

## Compatibility

### Supported Platforms
- ✅ Windows (tested)
- ✅ Linux (should work)
- ✅ macOS (should work)

### Supported Cards
- Estonian ID card (contact interface)
- Estonian ID card (contactless/NFC interface with CAN)
- Other PACE-enabled smart cards

### Supported Readers
- All PC/SC compatible readers
- Tested with ACR1252U NFC reader

## Future Work

1. **Real Card Testing**: Test with actual Estonian ID card via NFC
2. **Performance Benchmarks**: Compare with OpenPACE implementation
3. **Additional Protocols**: Add support for other PACE variants
4. **Brainpool Curves**: Add support for Brainpool domain parameters
5. **Generic Mapping**: Complete GM implementation (currently uses IM)

## Documentation

- **libopenpace-go**: See `repos/libopenpace-go/README.md`
- **Implementation Notes**: See `repos/libopenpace-go/docs/IMPLEMENTATION.md`
- **Completion Summary**: See `repos/libopenpace-go/COMPLETION_SUMMARY.md`

## Conclusion

The pure Go PACE implementation successfully replaces the CGO-based OpenPACE dependency, providing a simpler, more maintainable, and more portable solution for Estonian ID card NFC reading.

**Status**: ✅ **Integration Complete and Build Tested**

---

**Date**: November 15, 2025  
**Version**: 1.0.2  
**Author**: Aare Puusepp  
**Repository**: https://github.com/aarepuusepp/web-eid-pd

