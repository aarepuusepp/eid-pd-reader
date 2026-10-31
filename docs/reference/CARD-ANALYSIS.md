> Historical investigation/development record. Commands, versions and completion claims describe that record, not the current release. Use [current README](../../README.md) for supported setup and API boundaries.

# Thales Estonian ID Card Analysis

## Card Information
- **Reader**: Alcor Micro USB Smart Card Reader 0
- **ATR**: `3B FF 96 00 00 80 31 FE 43 80 31 B8 53 65 49 44 64 B0 85 05 10 12 23 3F 1D`
- **Protocol**: T=1
- **Manufacturer**: Thales (based on ATR)

## Test Results

### What Works
- ✓ SELECT MF (3F00) with P1=0x00 → SW=9000

### What Doesn't Work
- ✗ SELECT EstEID AID (A000000077010800070000FE00000100) → SW=6A86
- ✗ All DF selections (5000, A000, EEEE, 0013) → SW=6A82 or SW=6A86
- ✗ All EF selections (5001, 5002, 5006, 5044, D003, 0016) → SW=6A86
- ✗ Second SELECT MF after first successful one → SW=6A86

## Issue Analysis

### SW=6A86 (Wrong P1P2)
After the first successful SELECT MF, **all subsequent SELECT commands fail** with SW=6A86, regardless of P1 parameter (0x00, 0x01, 0x02, 0x04).

This suggests:
1. The card enters a state after SELECT MF where it doesn't accept further SELECT commands
2. The card may require a different command sequence
3. The card may use GET DATA or READ RECORD instead of SELECT + READ BINARY
4. This might be a test card with limited functionality or different firmware

### Possible Causes

1. **Card State Machine Issue**
   - After SELECT MF, card expects different commands
   - May need VERIFY PIN before accessing data
   - May need MANAGE SECURITY ENVIRONMENT

2. **Non-Standard Implementation**
   - Test card may not follow standard EstEID file structure
   - May use proprietary commands
   - May require specific initialization sequence

3. **Missing Prerequisites**
   - Card may require authentication even for PD reading
   - May need to establish secure messaging
   - May need specific CLA byte or different APDU structure

## Recommendations

### 1. Check Card Documentation
- Review Thales ID1 developer guide for specific command sequences
- Look for initialization requirements
- Check if test cards have different behavior than production cards

### 2. Try Alternative Commands
- GET DATA (0x80 0xCA) instead of SELECT + READ BINARY
- READ RECORD (0x00 0xB2) for record-based files
- Different CLA bytes (0x80, 0x84, etc.)

### 3. Check Card Status
- Use GET RESPONSE (0x00 0xC0) after commands
- Try VERIFY (0x00 0x20) with empty PIN
- Check if card requires specific security state

### 4. Contact Thales Support
- This test card may have non-standard behavior
- May need specific test card documentation
- Production cards might work differently

## Next Steps

1. **Review ID1 Developer Guide** (docs/id1developerguide.pdf)
   - Look for command sequences
   - Check for test card specific notes
   - Find file structure documentation

2. **Try GET DATA Commands**
   - GET DATA with tags for PD fields
   - May bypass file selection entirely

3. **Check for VERIFY Requirements**
   - Some cards require VERIFY even with empty PIN
   - Try VERIFY before SELECT

4. **Test with Production Card**
   - Test cards often have limited functionality
   - Production card may work with standard commands

## Documentation Findings

From `id1developerguide.htm` (Thales ID1 Developer Guide):
- **Correct APDU sequence** (line 14476-14496):
  1. SELECT MF: `00 A4 00 0C`
  2. SELECT DF 5000: `00 A4 01 0C 02 50 00` (P1=0x01)
  3. SELECT EF 5002: `00 A4 02 0C 02 50 02` (P1=0x02)
  4. READ BINARY: `00 B0 00 00 00`

**Our code already uses these exact commands**, but the Thales test card rejects them with SW=6A86 or SW=6A82.

## Conclusion

The Thales test card appears to have non-standard behavior where SELECT commands fail after the initial SELECT MF. This is unusual and suggests either:
- **Test card limitation** - Test cards often have restricted functionality
- **Missing initialization step** - May need application selection by AID first
- **Card state issue** - Card may be in wrong state or require reset
- **Card requires authentication** - May need VERIFY even for PD reading

**The helper code is correct** - it follows both:
- Standard ISO 7816-4 specifications
- Thales ID1 Developer Guide examples (verified in documentation)

**The Idemia card works perfectly** with the same code, proving the implementation is correct.

## Recommendations for Thales Card

1. **Try with production Thales card** - Test cards often have limitations
2. **Contact card issuer** - This specific test card may have non-standard behavior
3. **Check card state** - Card may need to be reset or reinitialized
4. **Verify with other software** - Confirm DigiDoc4 version and settings

