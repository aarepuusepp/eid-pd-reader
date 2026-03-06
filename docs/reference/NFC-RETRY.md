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

- `--retry N`: Total number of read attempts (default: 3)
  - `N` is the total number of attempts
  - `--retry 1` = single attempt (no retry)

### Examples

#### Default (3 attempts)
```bash
# Uses default 3 attempts
eid-pd-reader.exe --dump --nfc --can 775484
```
- Total attempts: **3**
- Delays: 1s, 2s between attempts
- Connection layer: waits for card presence + retries power-on

#### Single Attempt
```bash
# No retry, fail immediately
eid-pd-reader.exe --dump --nfc --can 775484 --retry 1
```
- Total attempts: **1**
- Behavior: Fails immediately if first attempt doesn't work

#### Five Attempts
```bash
# Maximum reliability
eid-pd-reader.exe --dump --nfc --can 775484 --retry 5
```
- Total attempts: **5**
- Delays: 1s, 2s, 3s, 3s between attempts

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
- **Connection layer**: Uses `SCardGetStatusChange` to wait up to 5s for card presence before connecting
- **Power-on retry**: Each connection attempt retries the power-on sequence up to 4 times (0ms, 300ms, 500ms, 1000ms delays)
- **Read-level retry**: Progressive delays between full read attempts (1s, 2s, 3s)

### What Gets Retried
Each retry attempt includes:
1. Reconnecting to the NFC reader
2. Establishing PACE protocol
3. Reading all personal data fields

### When to Use Retries

#### Recommended: default (3 attempts) or `--retry 5`
- **ACR122U readers**: Often need 2-3 attempts
- **ACR1252U readers**: Usually work on first attempt, but default 3 is safe
- **Unstable environments**: Use `--retry 5` for maximum reliability

#### Not Recommended: `--retry 1`
- Only use if you want immediate failure
- Useful for testing or when you know the reader is reliable

#### Not Recommended: `--retry 10+`
- Excessive retries waste time
- If 5 attempts fail, the issue is likely not transient
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
- **1 attempt**: ~2-3 seconds (if successful)
- **Default (3)**: ~3-8 seconds (if first fails)
- **5 attempts**: ~5-15 seconds (if multiple fail)

### Best Practices
1. **Use default (3 attempts)** for most use cases — no flag needed
2. **Use `--retry 5`** for production environments where reliability is critical
3. **Use `--retry 1`** only for testing or known-good readers
4. **Avoid `--retry 10+`** as it provides diminishing returns

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

### Two-Level Retry Architecture

**Level 1 — Connection layer (`nfc.ConnectAny`):**
- Uses `SCardGetStatusChange` to wait for card presence (5s timeout)
- Retries `SCardConnect` power-on up to 4 times with progressive delays
- Handles "smart card not responding to reset" errors automatically

**Level 2 — Read layer (`ReadPersonalDataNFCWithRetry`):**
- Retries the full connection + PACE + read cycle
- Progressive delays: 1s, 2s, 3s between attempts
- Disconnects with `ResetCard` disposition between attempts

## See Also

- [NFC Readers Guide](NFC-READERS.md) - Compatible readers and setup
- [PACE Protocol](../implementation/PACE-PROTOCOL.md) - Technical details
- [Troubleshooting](../guides/TROUBLESHOOTING.md) - Common issues and solutions

