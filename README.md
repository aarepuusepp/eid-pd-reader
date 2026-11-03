# EID-PD reader 2.1.11

Windows x64 native host: `ee.aarepuusepp.eid_pd_reader`. Package/reader 2.1.11,
extension 2.1.6, web SDK 1.0.2. Dependencies: published libcdoc-go 2.0.8 and
vendored libopenpace-go 1.1.3. Reads ID-card data and performs local CDOC1/CDOC2
private-key operations over contact or supported NFC sessions.

## Installation and update

Use the [eid-pd-installer](https://bitbucket.org/aarepuusepp/eid-pd-installer)
Windows EXE. It replaces managed files in `C:\Program Files\eid-pd-reader`,
removes obsolete payloads after success and generates/registers host manifests.
Admin configuration/policy lives in `%ProgramData%\eid-pd-reader`. No versions
subdirectory is used. The generated installation.html gives exact browser paths.

Chrome/Edge load the generated chromium directory in developer mode. Firefox
loads firefox/manifest.json through about:debugging; temporary add-ons must be
loaded again after a Firefox restart. Reload application tabs after updates.
Users do not edit extension IDs/manifests/registry. Public signing/store delivery
is separate from development installation.

Windows PC/SC and matching reader drivers are required. Contact crypto needs a
compatible x64 PKCS#11 module. Personal-data and crypto readiness are separate.
Linux/macOS code exists, but this package and acceptance cover Windows x64 only.

## Build and test

Go 1.27.3+, Windows x64 C compiler and PC/SC. Adjust compiler path as needed:

```powershell
$env:GOWORK = 'off'
$env:CGO_ENABLED = '1'
$env:CC = 'C:\msys64-2\ucrt64\bin\gcc.exe'
$env:PATH = (Split-Path $env:CC) + ';' + $env:PATH
New-Item -ItemType Directory -Force .local | Out-Null
go mod verify
go test -race ./...
go vet ./...
go build -o .local/eid-pd-reader.exe ./cmd/eid-pd-reader
```

For the complete WinUI PIN dialog/maintenance/browser package, clone reader,
extension and installer as siblings and run `pwsh -File build-installer.ps1`
from installer. That builder also needs Node.js, .NET 10 SDK and NSIS.
Legacy install-reader/build-release scripts are not the current workstation
packaging path. No sibling libcdoc checkout is required.

## Use and diagnostics

Applications bundle web-eid-pd.js and use status, readPersonalData, readAllReaders,
getDecryptCertificate and unwrapTransportKey. The extension enforces top-level
HTTPS origin and machine/site policy on each operation. No shared cross-site
personal-data cache is used. Native messaging uses length-prefixed JSON and is
not a localhost HTTP service.

PIN1 belongs only in the native WinUI prompt (60-second deadline). Web callers
must not send pin, PKCS#11 paths or arbitrary native commands. NFC CAN stays only
for the operation and must not reach the backend. status reads no personal data
and asks for no PIN. Legacy APIs share the same policy checks.

```powershell
& 'C:\Program Files\eid-pd-reader\eid-pd-reader.exe' --version
& 'C:\Program Files\eid-pd-reader\eid-pd-reader.exe' --help
# Explicitly reads personal data to stdout; use only for an intended local test:
& 'C:\Program Files\eid-pd-reader\eid-pd-reader.exe' --dump
```

Without diagnostic flags, the executable waits for native messages; it does not
print a menu. For direct NFC developer tests use --nfc --can-env with an ephemeral
environment variable. Normal users enter CAN in the application. Verbose/debug
output can contain card data and must stay local.

## Configuration and access policy

HKLM machine policy takes precedence over administrator-managed ProgramData
policy.json. Blocked origin wins, then allowed origin, then defaultAction
(ask/allow/deny). Malformed policy denies operations. Personal-data/decrypt user
grants are separate. See installer DEPLOYMENT.md for registry values, ACLs, silent
install, repair, rollback and uninstall. Runtime config.yaml sets card/logging/
PKCS#11 defaults; it does not grant website access or accept a stored PIN.

## Evidence and documentation

Contact CDOC1 and IDEMIA NFC data/CDOC1/CDOC2 paths have separately recorded
card tests. Do not infer all Thales/test-card profiles from these results.
The user confirmed standalone CDOC2 roundtrip in Digisign and reading in Visitor
with installed package 2.1.10; browser/transport was unspecified in that report.

- [Release notes](RELEASE-2.1.11.md)
- [NFC crypto and evidence](docs/NFC-CRYPTO.md)
- [Direct NFC tests](docs/DIRECT-NFC-TEST.md)
- [Documentation index](docs/README.md)
- [PIN dialog build](pin-ui/README.md)

MIT license: LICENSE. Current installation authority is the installer README;
SDK API authority is the extension repository's web-eid-pd/README.md.
