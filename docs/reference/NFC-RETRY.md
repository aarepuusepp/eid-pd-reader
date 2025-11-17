# NFC Reading Retry Configuration

## Overview

NFC card readers can sometimes fail to establish a connection on the first attempt due to:
- Reader initialization delays
- Card positioning
- RF interference
- Timing issues

The `--retry` flag allows you to configure how many times the application should retry NFC reading if the initial attempt fails.

## Usage

```bash
eid-pd-reader.exe --dump --nfc --can <CAN> [--retry N]
```

### Parameters

- `--retry N`: Number of retry attempts (default: 0 = no retry)
  - `N` is the number of **additional** attempts after the first one
  - Total attempts = N + 1

### Examples

#### No Retry (Default)
```bash
# Single attempt only
eid-pd-reader.exe --dump --nfc --can 775484
```
- Total attempts: **1**
- Behavior: Fails immediately if first attempt doesn't work

#### One Retry
```bash
# Initial attempt + 1 retry = 2 total attempts
eid-pd-reader.exe --dump --nfc --can 775484 --retry 1
```
- Total attempts: **2**
- Behavior: Retries once if first attempt fails
- Delay between attempts: 500ms

#### Two Retries
```bash
# Initial attempt + 2 retries = 3 total attempts
eid-pd-reader.exe --dump --nfc --can 775484 --retry 2
```
- Total attempts: **3**
- Behavior: Retries up to 2 times if needed
- Delay between attempts: 500ms

#### Verbose Mode
```bash
# See detailed retry information
eid-pd-reader.exe --verbose --dump --retry 2 --nfc --can 775484
```

Output will show:
```
Connected to NFC reader: ACS ACR1252 1S CL Reader PICC 0 (attempt 1)
Attempt 1 failed: ...
Retrying NFC connection (attempt 2/3)...
Connected to NFC reader: ACS ACR1252 1S CL Reader PICC 0 (attempt 2)
...
```

## Retry Behavior

### Timing
- **First attempt**: 300ms initial delay (for card detection)
- **Subsequent attempts**: 500ms delay between retries

### What Gets Retried
Each retry attempt includes:
1. Reconnecting to the NFC reader
2. Establishing PACE protocol
3. Reading all personal data fields

### When to Use Retries

#### Recommended: `--retry 1` or `--retry 2`
- **ACR122U readers**: Often need 1-2 retries
- **ACR1252U readers**: Usually work on first attempt, but 1 retry is safe
- **Unstable environments**: Use 2 retries for maximum reliability

#### Not Recommended: `--retry 0` (default)
- Only use if you want immediate failure
- Useful for testing or when you know the reader is reliable

#### Not Recommended: `--retry 5+`
- Excessive retries waste time
- If 3 attempts fail, the issue is likely not transient
- Better to check card positioning or reader connection

## Error Messages

### With Retries
```json
{
  "ok": false,
  "error": "all NFC read attempts failed: PACE failed: ..."
}
```

### Without Retries
```json
{
  "ok": false,
  "error": "PACE failed: ..."
}
```

## Integration with Web Extension

The web extension can specify retry behavior in the native messaging request:

```javascript
// Example: Request with retry
const response = await chrome.runtime.sendNativeMessage(
  'ee.aarepuusepp.eid_helper',
  {
    command: 'readPersonalData',
    mode: 'nfc',
    can: '775484',
    retry: 2  // Optional: number of retries
  }
);
```

**Note**: Native messaging integration for retry parameter is not yet implemented. Currently, retry is only available via command-line flag.

## Performance Considerations

### Time Impact
- **No retry**: ~2-3 seconds (if successful)
- **1 retry**: ~3-5 seconds (if first fails, second succeeds)
- **2 retries**: ~4-7 seconds (if first two fail, third succeeds)

### Best Practices
1. **Start with `--retry 1`** for most use cases
2. **Use `--retry 2`** for production environments where reliability is critical
3. **Use `--retry 0`** (default) only for testing or known-good readers
4. **Avoid `--retry 3+`** as it provides diminishing returns

## Troubleshooting

### Reader Never Connects
If all retry attempts fail:
1. Check card is properly placed on reader
2. Verify CAN is correct (6 digits)
3. Check reader is connected (should appear in Device Manager)
4. Try contact reading mode (without `--nfc`) to verify card works
5. Check for RF interference (move away from other NFC devices)

### Inconsistent Results
If sometimes works, sometimes doesn't:
1. Increase retry count to `--retry 2`
2. Check card positioning (center it on the reader)
3. Ensure card stays in place during entire read operation
4. Check reader firmware is up to date

### Always Needs Multiple Retries
If first attempt always fails:
1. May be normal for some reader models (ACR122U)
2. Consider updating reader firmware
3. Check USB connection quality
4. Try different USB port (preferably USB 2.0, not 3.0)

## Technical Details

### Retry Logic
```go
func ReadPersonalDataNFCWithRetry(can string, maxAttempts int) (*PersonalData, error) {
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        if attempt > 1 {
            time.Sleep(500 * time.Millisecond)
        }
        
        pd, err := attemptNFCRead(can, attempt)
        if err == nil {
            return pd, nil  // Success!
        }
        
        // Log and continue to next attempt
    }
    
    return nil, fmt.Errorf("all attempts failed")
}
```

### First Attempt Delay
```go
func attemptNFCRead(can string, attemptNum int) (*PersonalData, error) {
    if attemptNum == 1 {
        time.Sleep(300 * time.Millisecond)  // Initial card detection
    }
    // ... proceed with PACE and reading
}
```

## See Also

- [NFC Readers Guide](NFC-READERS.md) - Compatible readers and setup
- [PACE Protocol](../implementation/PACE-PROTOCOL.md) - Technical details
- [Troubleshooting](../guides/TROUBLESHOOTING.md) - Common issues and solutions

