# Supported Estonian ID Cards

`eid-pd-reader` supports all current Estonian ID card types with automatic card detection and optimized reading strategies.

## Supported Card Types

### 1. **IDEMIA Cosmo 8.1/8.2**
- **ATR**: `3BDB960080B1FE451F830012233F536549440F9000F1`
- **Manufacturer**: IDEMIA (formerly Oberthur Technologies)
- **Used in**: Estonian ID cards issued from ~2018-2021
- **Reading Strategy**: Path-based selection (MF/DF 5000)
- **Source Tag**: `chip:IDEMIA`

### 2. **IDEMIA Cosmo X**
- **ATR**: `3BDC960080B1FE451F830012233F54654944320F9000C3`
- **Manufacturer**: IDEMIA
- **Used in**: Estonian ID cards issued from ~2021 onwards
- **Reading Strategy**: Path-based selection (MF/DF 5000)
- **Special Note**: Uses limited Le (0xC0) for certificate reading
- **Source Tag**: `chip:IDEMIA-CosmoX`

### 3. **Thales (Gemalto)**
- **ATR**: `3BFF9600008031FE438031B85365494464B085051012233F1D`
- **Manufacturer**: Thales (formerly Gemalto)
- **Used in**: Estonian ID cards issued ~2024 onwards
- **Reading Strategy**: PKCS-15 AID selection + DF DFDD
- **Special Note**: Requires different file structure (DFDD instead of 5000)
- **Source Tag**: `chip:Thales`

## Card Detection

The helper automatically detects card type using the ATR (Answer To Reset) and applies the optimal reading strategy:

1. **ATR-based detection**: Identifies card manufacturer from ATR bytes
2. **Optimized strategy first**: Uses manufacturer-specific commands for fast reading
3. **Fallback strategies**: If optimized strategy fails, tries generic methods
4. **Universal compatibility**: Works with unknown cards using standard ISO 7816 commands

## Reading Strategies

### IDEMIA Cards (Cosmo 8.1/8.2, Cosmo X)
```
1. SELECT MF + DF 5000 (path-based: P1=0x09)
2. SELECT EF 5001-5009 (P1=0x02)
3. READ BINARY for each field
```

### Thales Cards
```
1. SELECT PKCS-15 AID (A000000063504B43532D3135)
2. SELECT DF DFDD (path-based: P1=0x08)
3. SELECT EF 5001-5009 (P1=0x02)
4. READ BINARY for each field
```

### Generic Fallback (Unknown Cards)
```
Strategy 1: GET DATA commands (BER-TLV tags)
Strategy 2: Standard MF->DF->EF navigation
Strategy 3: Direct EF selection without DF
Strategy 4: Full path-based selection
```

## Personal Data Fields

All card types return the following fields:

| Field | FID | Description | Example |
|-------|-----|-------------|---------|
| Surname | 5001 | Last name | "JÕEORG" |
| Given Name | 5002 | First name(s) | "JAAK-KRISTJAN" |
| Sex | 5003 | Gender | "M" or "F" |
| Citizenship | 5004 | Nationality | "EST" |
| Birth Date | 5005 | Date and place of birth | "01 08 1980" |
| Personal Code | 5006 | National ID number | "38001085718" |
| Document Number | 5007 | Card number | "AT0000222" |
| Expiry Date | 5008 | Card expiration | "09 12 2029" |
| Issue Date | 5009 | Card issue date | "09 12 2024" |

## Testing

### Test with your card:
```bash
# Basic test
eid-pd-reader.exe --dump

# Verbose mode (shows APDU commands)
eid-pd-reader.exe --dump --verbose

# Debug mode (tests all file structures)
eid-pd-reader.exe --debug
```

### Expected output:
```json
{
  "ok": true,
  "data": {
    "surname": "JÕEORG",
    "givenName": "JAAK-KRISTJAN",
    "sex": "M",
    "citizenship": "EST",
    "birthDate": "01 08 1980",
    "personalCode": "38001085718",
    "documentNr": "AT0000222",
    "expiryDate": "09 12 2029",
    "issueDate": "09 12 2024",
    "source": "chip:Thales"
  }
}
```

## Compatibility Notes

### ✅ Fully Supported
- All production Estonian ID cards (IDEMIA, Thales)
- Both contact and contactless interfaces
- Windows, Linux, macOS (via PC/SC)
- Special characters (UTF-8 encoding: Õ, Ä, Ö, Ü)

### ⚠️ Limitations
- **Test cards**: Some test cards may have non-standard provisioning
- **NFC**: Contactless reading requires PACE protocol (not yet implemented)
- **Certificates**: Certificate reading not yet implemented (planned)

## Implementation Details

Based on official RIA (Estonian Information System Authority) repositories:
- **DigiDoc4-Client**: Desktop application reference implementation
- **nfc-Android-lib**: Android NFC library
- **ESTEID-PersoFile**: Card personalization file reader

All APDU commands follow:
- ISO/IEC 7816-4 (smart card standard)
- PKCS#15 (cryptographic token information format)
- Estonian ID card specifications

## Future Enhancements

Planned features:
- [ ] NFC support with PACE protocol
- [ ] Certificate reading and parsing
- [ ] PIN verification and change
- [ ] Digital signature operations
- [ ] Additional card types (eResident, Diplomatic ID)

## References

- [Estonian ID Card Developer Guide](https://www.id.ee/en/article/for-developers/)
- [RIA GitHub Repositories](https://github.com/open-eid)
- [ISO/IEC 7816-4 Standard](https://www.iso.org/standard/77180.html)
- [PKCS#15 Specification](http://docs.oasis-open.org/pkcs11/pkcs11-curr/v2.40/pkcs11-curr-v2.40.html)

