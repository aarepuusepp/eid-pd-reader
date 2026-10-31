# Current reader/browser security model

Native host: ee.aarepuusepp.eid_pd_reader. The browser launches its registered
executable and exchanges length-prefixed stdio messages. The installer owns
manifest/registry generation and administrator-only policy/configuration paths.

The extension validates browser-supplied tab/frame metadata, top-level HTTPS
origin, navigation and operation. Machine policy is checked for each operation:
blocked origin, then allowed origin, then ask/allow/deny default. User grants are
separate for personal data and decrypt. A malformed machine policy fails closed.
The page cannot authorize itself by supplying an origin string.

User consent occurs in an extension-owned window. Saved grant state is stored;
cardholder data is not shared through a cross-site cache. PIN1 is entered only
in the native WinUI dialog; webpage PIN/PKCS#11 overrides are rejected. NFC CAN
stays operation-local and is cleared by the application. Legacy adapters pass
through the same policy path. status/presence checks read no personal data.

Verbose/debug card dumps can expose personal information. Keep diagnostics,
local CDOC files, keys and config out of Git/log collection. No blanket claim is
made that arbitrary debugging modes cannot log card data. See PIN UI README for
buffer-clearing limitations and the 60-second dialog/90-second watchdog deadlines.

Browser permissions/native IDs are not website allowlists. Use installer machine
policy rather than editing extension manifests for each application. Production
signing/store publication and clean-machine acceptance remain separate gates.
[Current reader guide](../../README.md), [NFC crypto](../NFC-CRYPTO.md).
