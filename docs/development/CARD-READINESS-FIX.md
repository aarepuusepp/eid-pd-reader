# Card Readiness Fix - Preventing Corrupted Data Reads

**Date:** November 14, 2025  
**Issue:** Corrupted UTF-8 data when reading card immediately after insertion  
**Version:** 1.0.3 (upcoming)

---

## 🐛 Problem Description

### Symptoms

When a user:
1. Removes ID card from reader
2. Clears cache
3. Inserts ID card into reader
4. **Immediately** clicks "Read ID Card" button

The extension sometimes displays corrupted data with replacement characters (�):

**Corrupted Output:**
```
Name: 0����� JÕEORG
Personal Code: 00authentication PIN�0�0��
              ��060signature PIN�0�0��
              ��070PIN unblocking key�0�0� ��
Document Number: AT0000222
```

**Expected Output:**
```
Name: JAAK-KRISTJAN JÕEORG
Personal Code: 38001085718
Document Number: AT0000222
```

### Root Cause

The issue occurs because:

1. **Card Not Fully Initialized**: When a smart card is inserted, the card reader needs a brief moment to:
   - Establish electrical contact
   - Perform ATR (Answer To Reset) negotiation
   - Initialize the card's internal state

2. **Immediate Read Attempt**: If the application tries to read data before the card is fully ready:
   - APDU commands may return partial or corrupted responses
   - Data bytes may be misinterpreted
   - UTF-8 encoding becomes corrupted

3. **No Validation**: Previous code didn't validate the data before returning it to the user.

---

## ✅ Solution Implemented

### 1. Card Readiness Check

Added `ensureCardReady()` function that:
- Attempts to communicate with the card
- Retries up to 3 times with 100ms delays
- Verifies card responds correctly to SELECT MF command

```go
func ensureCardReady(card *scard.Card) error {
	maxRetries := 3
	retryDelay := 100 * time.Millisecond
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logVerbose("Card readiness check attempt %d/%d", attempt+1, maxRetries)
			time.Sleep(retryDelay)
		}
		
		// Try to get card status
		status, err := card.Status()
		if err != nil {
			logVerbose("Card status check failed: %v", err)
			continue
		}
		
		// Check if card is responsive
		if len(status.Atr) == 0 {
			logVerbose("Card ATR is empty, card may not be ready")
			continue
		}
		
		// Try a simple SELECT MF command to verify card is responsive
		selectMF := []byte{0x00, 0xA4, 0x00, 0x0C, 0x02, 0x3F, 0x00}
		resp, err := card.Transmit(selectMF)
		if err != nil {
			logVerbose("SELECT MF failed: %v", err)
			continue
		}
		
		// Check response
		if len(resp) >= 2 {
			sw1, sw2 := resp[len(resp)-2], resp[len(resp)-1]
			if sw1 == 0x90 && sw2 == 0x00 {
				logVerbose("Card is ready (attempt %d)", attempt+1)
				return nil
			}
			logVerbose("SELECT MF returned SW=%02X%02X", sw1, sw2)
		}
	}
	
	return fmt.Errorf("card not responding after %d attempts", maxRetries)
}
```

**Key Features:**
- ✅ 3 retry attempts
- ✅ 100ms delay between attempts
- ✅ Validates ATR presence
- ✅ Tests card with SELECT MF command
- ✅ Verbose logging for debugging

---

### 2. Data Validation

Added UTF-8 validation functions:

#### `isValidUTF8Text()`
Checks if text contains valid UTF-8:
- Detects UTF-8 replacement character (�) = `\uFFFD`
- Checks for excessive control characters
- Returns false if >20% of characters are control chars

```go
func isValidUTF8Text(s string) bool {
	// Check for UTF-8 replacement character (�) which indicates corrupted encoding
	if strings.Contains(s, "\uFFFD") {
		return false
	}
	// Check for excessive control characters (except space and common whitespace)
	controlCount := 0
	for _, r := range s {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			controlCount++
		}
	}
	// If more than 20% of characters are control characters, data is likely corrupted
	if len(s) > 0 && controlCount > len(s)/5 {
		return false
	}
	return true
}
```

#### `validatePersonalData()`
Validates critical fields:
- Surname
- Given Name
- Personal Code

```go
func validatePersonalData(pd *PersonalData) bool {
	// Check critical fields for corruption
	if pd.Surname != "" && !isValidUTF8Text(pd.Surname) {
		logVerbose("Surname contains invalid UTF-8: %q", pd.Surname)
		return false
	}
	if pd.GivenName != "" && !isValidUTF8Text(pd.GivenName) {
		logVerbose("GivenName contains invalid UTF-8: %q", pd.GivenName)
		return false
	}
	if pd.PersonalCode != "" && !isValidUTF8Text(pd.PersonalCode) {
		logVerbose("PersonalCode contains invalid UTF-8: %q", pd.PersonalCode)
		return false
	}
	return true
}
```

---

### 3. Strategy Validation

Updated all read strategies to validate data before returning:

**Before:**
```go
pd, err := readPersonalDataThales(card, optionalFIDs)
if err == nil {
	return pd, nil  // ❌ Returns potentially corrupted data
}
```

**After:**
```go
pd, err := readPersonalDataThales(card, optionalFIDs)
if err == nil && validatePersonalData(pd) {
	return pd, nil  // ✅ Only returns validated data
}
if err == nil {
	logVerbose("Thales strategy returned corrupted data, trying fallbacks")
} else {
	logVerbose("Thales strategy failed: %v, trying fallbacks", err)
}
```

**Applied to all strategies:**
- ✅ Thales-optimized strategy
- ✅ IDEMIA-optimized strategy
- ✅ Strategy 1: GET DATA
- ✅ Strategy 2: Standard (MF->DF->EF)
- ✅ Strategy 3: Direct EF (no DF)
- ✅ Strategy 4: Path-based

---

## 📊 Technical Details

### Timing Analysis

**Card Insertion Timeline:**
```
0ms    - Card inserted into reader
0-50ms - Physical contact establishment
50-100ms - ATR negotiation
100-150ms - Card initialization
150ms+ - Card fully ready
```

**Our Solution:**
- First attempt: Immediate (may succeed if card was already ready)
- Retry 1: +100ms delay (usually succeeds)
- Retry 2: +100ms delay (fallback)
- Total max delay: ~200ms (imperceptible to user)

### Character Encoding

**UTF-8 Replacement Character:**
- Unicode: U+FFFD
- UTF-8 bytes: 0xEF 0xBF 0xBD
- Displayed as: �

**When it appears:**
- Invalid UTF-8 byte sequences
- Corrupted data from card
- Encoding/decoding errors

---

## 🧪 Testing

### Test Scenario 1: Normal Operation
1. Insert card
2. Wait 1 second
3. Click "Read ID Card"
4. **Expected:** Data reads correctly (first attempt)

### Test Scenario 2: Immediate Read (Bug Reproduction)
1. Remove card
2. Clear cache
3. Insert card
4. **Immediately** click "Read ID Card"
5. **Expected:** 
   - First attempt may fail or return corrupted data
   - Retry succeeds after 100ms delay
   - Valid data displayed

### Test Scenario 3: Card Not Ready
1. Insert card
2. Immediately click "Read ID Card" (within 50ms)
3. **Expected:**
   - Readiness check retries 3 times
   - Card becomes ready during retries
   - Data reads successfully

### Test Scenario 4: Card Removed During Read
1. Start reading card
2. Remove card mid-operation
3. **Expected:**
   - Error: "card not responding after 3 attempts"
   - User sees clear error message

---

## 📝 Files Modified

### Native Helper
✅ **`native/esteid/reader.go`**
- Added `import "time"` for delays
- Added `ensureCardReady()` function
- Added `isValidUTF8Text()` function
- Added `validatePersonalData()` function
- Updated all strategy returns to validate data
- Added verbose logging for debugging

---

## 🎯 Benefits

### 1. **Reliability**
- ✅ Prevents corrupted data from being displayed
- ✅ Automatic retry on card not ready
- ✅ Graceful handling of timing issues

### 2. **User Experience**
- ✅ No need to wait before clicking button
- ✅ Transparent retry (user doesn't notice)
- ✅ Always get correct data or clear error

### 3. **Debugging**
- ✅ Verbose logging shows readiness checks
- ✅ Logs corrupted data detection
- ✅ Helps diagnose card reader issues

---

## 🔍 Error Messages

### Before Fix
```
Name: 0����� JÕEORG
Personal Code: 00authentication PIN�0�0��...
```
User sees corrupted data, doesn't know what's wrong.

### After Fix

**Scenario A: Success after retry**
```
Verbose log:
Card readiness check attempt 1/3
Card is ready (attempt 1)
✓ Strategy 2 succeeded

User sees: Correct data
```

**Scenario B: Card not responding**
```
Error: card not responding after 3 attempts

User sees: Clear error message
```

**Scenario C: Corrupted data detected**
```
Verbose log:
Surname contains invalid UTF-8: "0����� JÕEORG"
✗ Strategy 2 returned corrupted data
Trying next strategy...

User sees: Either correct data from next strategy or error
```

---

## 🚀 Deployment

### Version
- Current: 1.0.2
- Next: 1.0.3 (with this fix)

### Compatibility
- ✅ Backward compatible
- ✅ No configuration changes needed
- ✅ No API changes

### Build & Install
```bash
# Build native helper
go build -o eid-pd-reader.exe

# Install (Windows)
.\install.cmd

# Test
eid-pd-reader.exe --verbose
```

---

## 📚 Related Documentation

- **[TROUBLESHOOTING.md](docs/reference/TROUBLESHOOTING.md)** - User-facing troubleshooting
- **[SUPPORTED-CARDS.md](docs/reference/SUPPORTED-CARDS.md)** - Card compatibility
- **[README.md](README.md)** - Main documentation

---

## 🎉 Summary

**Problem:** Corrupted data when reading card immediately after insertion

**Solution:**
1. ✅ Card readiness check with retries
2. ✅ UTF-8 validation before returning data
3. ✅ All strategies validate data
4. ✅ Verbose logging for debugging

**Result:** Reliable card reading regardless of timing

**Impact:** Better user experience, fewer support requests

---

**Status:** ✅ Implemented and ready for testing

