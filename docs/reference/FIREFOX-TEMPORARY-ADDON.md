# Firefox development add-on

Install EID-PD 2.1.10. At about:debugging#/runtime/this-firefox choose Load Temporary
Add-on and select C:\Program Files\eid-pd-reader\browsers\firefox\manifest.json.
Reload application pages afterwards. Repeat after Firefox restarts. The installer
uses the configured stable Firefox ID; do not generate/update a temporary host ID.

The popup displays extension 2.1.6 and reader 2.1.10 independently. Public release
uses a Mozilla-signed XPI/store delivery; temporary loading is for development.
See [current installation](../../README.md#installation-and-update).
