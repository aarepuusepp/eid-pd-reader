# NFC retries

Reader --retry selects total NFC read attempts (default 3). Retries do not justify
reusing another card's CAN or retrying an authentication failure with weaker
cryptography. Keep the current card stable and distinguish no-card, unpersonalized
card, unsupported profile and authentication errors. Clear operation CAN afterwards.

A PIN timeout/cancel must be retried by explicit user action; no plaintext is
released. See [NFC flow](../NFC-CRYPTO.md), [direct tests](../DIRECT-NFC-TEST.md)
and [troubleshooting](TROUBLESHOOTING.md).
