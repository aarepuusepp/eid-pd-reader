# Thales Test Card Support - Investigation Notes

**Date:** November 14, 2025  
**Card Type:** Thales Test ID Card (via ACR1252U NFC Reader)  
**ATR:** `3B 8F 80 01 80 31 D8 53 65 49 44 64 B0 85 05 10 12 23 3F 16`

---

## 🔍 Current Status

**Card Detection:** ✅ Working - Detected as "Thales (test/NFC)"  
**Card Reading:** ❌ Not working - All strategies failing

---

## 📊 Test Results

### What Works
- ✅ NFC reader detection (ACS ACR1252 1S CL Reader PICC)
- ✅ Card presence detection
- ✅ Card readiness check
- ✅ ATR reading
- ✅ SELECT MF (3F00) - Returns `90 00` (success)

### What Doesn't Work
- ❌ SELECT EstEID AID - Returns `6A 86` (incorrect parameters)
- ❌ SELECT PKCS-15 AID - Returns `69 82` (security status not satisfied)
- ❌ SELECT DF 5000 - Returns `6A 82` (file not found)
- ❌ All EF selection methods fail
- ❌ GET DATA commands fail

---

## 🧪 Commands Tested

### Application Selection

```
Command: SELECT EstEID AID
APDU: 00 A4 04 00 10 A0 00 00 00 77 01 08 00 07 00 00 FE 00 00 01 00
Response: 6A 86 (Incorrect parameters P1-P2)
Status: ❌ Failed
```

```
Command: SELECT PKCS-15 AID
APDU: 00 A4 04 00 0C A0 00 00 00 63 50 4B 43 53 2D 31 35
Response: 69 82 (Security status not satisfied)
Status: ❌ Failed
```

### File Selection

```
Command: SELECT MF
APDU: 00 A4 00 0C 02 3F 00
Response: 90 00 (Success)
Status: ✅ Works
```

```
Command: SELECT DF 5000
APDU: 00 A4 00 0C 02 50 00
Response: 6A 82 (File not found)
Status: ❌ Failed
```

```
Command: SELECT EF 5001 (Surname)
APDU: 00 A4 02 0C 02 50 01
Response: 6A 82 (File not found)
Status: ❌ Failed
```

### Data Reading

```
Command: GET DATA (Surname)
APDU: 00 CA 5F 0E 00
Response: 69 99 (Command not allowed)
Status: ❌ Failed
```

---

## 💡 Possible Causes

### 1. PIN Verification Required
**Likelihood:** Low  
**Reason:** Personal data should be readable without PIN  
**Test:** Try PIN verification before reading

### 2. Different File Structure
**Likelihood:** High  
**Reason:** Test cards may have different DF/EF organization  
**Solution:** Need to discover actual file structure

### 3. Application Not Activated
**Likelihood:** Medium  
**Reason:** `69 82` suggests security/activation issue  
**Solution:** May need specific activation command

### 4. NFC-Specific Limitations
**Likelihood:** Medium  
**Reason:** Contactless interface may have restrictions  
**Solution:** Try with contact reader

### 5. Test Card Configuration
**Likelihood:** High  
**Reason:** Test cards may be configured differently  
**Solution:** Need test card documentation

---

## 🔧 Next Steps to Try

### Step 1: Try Contact Reader
If you have a contact card reader, try reading the card via contact interface instead of NFC.

### Step 2: List Files
Try to list available files/applications:

```bash
# This would require adding LIST FILES command support
# APDU: 00 A4 04 00 (with various parameters)
```

### Step 3: Try PIN Verification
Even though it shouldn't be needed for personal data:

```
# PIN1 verification (if needed)
# APDU: 00 20 00 01 [PIN length] [PIN data]
```

### Step 4: Check Card Documentation
- Contact Thales support
- Check if test cards have special requirements
- Verify if test cards support personal data reading via NFC

### Step 5: Try Different P1/P2 Parameters
The `6A 86` error suggests incorrect parameters. Try variations:

```
# Different SELECT methods
P1=00, P2=00  # Select by FID
P1=01, P2=00  # Select child DF
P1=02, P2=00  # Select EF under current DF
P1=04, P2=00  # Select by AID
P1=08, P2=00  # Select by path from MF
P1=09, P2=00  # Select by path from current DF
```

---

## 📝 Error Code Reference

### SW=6A 82 - File Not Found
- The requested file (DF or EF) doesn't exist
- Possible: Different file IDs for test cards
- Solution: Discover actual file structure

### SW=6A 86 - Incorrect Parameters P1-P2
- The P1/P2 parameters are not supported
- Possible: Card doesn't support this selection method
- Solution: Try different P1/P2 combinations

### SW=69 82 - Security Status Not Satisfied
- Security condition not fulfilled
- Possible: PIN verification required
- Possible: Application not activated
- Solution: Verify PIN or activate application

### SW=69 99 - Command Not Allowed
- Command not allowed in current state
- Possible: Wrong application context
- Solution: Select correct application first

---

## 🎯 Recommendations

### For Production Use
**Use production cards** - Test cards may have limitations that don't exist in production cards.

### For Development
1. **Get production card** for testing
2. **Contact Thales** for test card documentation
3. **Try contact reader** if available
4. **Check RIA documentation** for Estonian ID card specifications

### Alternative Approach
If personal data reading via NFC test cards isn't supported:
- Use production cards for testing
- Use contact interface for test cards
- Focus on certificate-based authentication instead

---

## 📚 Resources Needed

1. **Thales Test Card Documentation**
   - File structure
   - Supported commands
   - NFC limitations
   - Activation procedures

2. **Estonian ID Card Specifications**
   - Official APDU command reference
   - File organization
   - Security requirements

3. **Contact Reader** (for comparison)
   - Test if contact interface works
   - Compare behavior

---

## ✅ What's Been Implemented

1. ✅ Thales test card ATR detection
2. ✅ Card type identification
3. ✅ Multiple selection strategies
4. ✅ Comprehensive error logging
5. ✅ Fallback mechanisms

---

## ❓ Questions for Card Issuer

1. Do test cards support personal data reading via NFC?
2. What is the correct file structure for test cards?
3. Is PIN verification required for personal data?
4. Are there any activation commands needed?
5. What are the differences between test and production cards?

---

## 🔄 Workaround

**For now, if you need to test:**

1. **Use production ID card** - Should work with current implementation
2. **Use contact reader** - May work better than NFC
3. **Focus on certificate operations** - May be more reliable for test cards

---

**Status:** 🔍 Investigation ongoing - Test card may have limitations not present in production cards.

**Recommendation:** Try with a production Estonian ID card to verify the implementation works correctly.

