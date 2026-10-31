# NFC reader setup

Use a Windows PC/SC contactless reader and its matching vendor driver. The ACR1252
was used in recorded tests. Keep the card stable and supply its own CAN through
the application's NFC field. Do not reuse a CAN for another card.

PACE is built into libopenpace-go; no OpenPACE C build or localhost server is needed.
Reader/card presence is not proof of personalized fields or decryption capability.
Use [current setup](../../README.md), [NFC crypto](../NFC-CRYPTO.md) and
[direct developer diagnostics](../DIRECT-NFC-TEST.md). The initial ACR1252 double
beep was also reproduced by a plain PC/SC connection and is not claimed fixed.
