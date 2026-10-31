# Native host and extension identities

The Windows installer generates manifests and registry entries for Firefox,
Chrome and Edge. End users do not replace YOUR_EXTENSION_ID or edit manifests.
Use the [current installation path](../../README.md#installation-and-update).

Package maintainers set published IDs/URLs in installer release-config.json.
Development Chromium uses the checked-in development public key. Firefox uses
the configured stable gecko ID. Public builds validate required identities,
signing and store URLs. Native host remains ee.aarepuusepp.eid_pd_reader.

Reload the extension and application tab after an update. A mismatch should be
fixed by building/installing a coherent package, not by accepting arbitrary IDs.
