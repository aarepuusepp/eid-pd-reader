# Direct NFC acceptance test (Windows)

`cmd/eid-pd-card-test` uses the same native NFC certificate and key-unwrapping
functions as the browser host. It needs no browser, extension, backend or login.
It is an explicit local diagnostic, not a replacement for web origin policy.

Build with the project's Windows Go/CGO toolchain and `GOWORK=off`:

```powershell
go build -o .local/card-test/eid-pd-card-test.exe ./cmd/eid-pd-card-test
```

For decryption, the existing native `pin-ui` directory must be next to that EXE.
In the current test checkout `.local/card-test/pin-ui` is a junction to the
installed reader version's PIN UI. Recreate it if that version is uninstalled.
Do not download or substitute an untrusted PIN helper.

Enter CAN interactively; do not put its value in shell history or a script:

```powershell
$secret = Read-Host 'CAN' -AsSecureString
$env:EIDPD_TEST_CAN = [System.Net.NetworkCredential]::new('', $secret).Password
try {
    # Without -container: certificate-only read; no PIN prompt.
    & .local/card-test/eid-pd-card-test.exe
} finally {
    Remove-Item Env:EIDPD_TEST_CAN -ErrorAction SilentlyContinue
    $secret.Dispose()
}
```

To test decryption, use the same input wrapper but add
`-container 'C:\absolute\path\example.cdoc'` to the command. Leave the card on
the reader and enter PIN1 only in the native window. The tool first checks that
the card is a recipient, then performs key unwrapping and container decryption.
It reports file count, lengths and SHA-256 hashes, not file contents or identity
data. It does not save decrypted files. CAN is removed from the child environment
before the PIN dialog is launched; PIN stays in the existing native flow.

Confirmed locally on 2026-10-31: certificate-only probe and CDOC v1 decryption
on the user's real IDEMIA NFC card; two payloads decrypted. This exposed and
verified the case-1 secure-messaging MAC padding correction. The user also confirmed successful NFC decryption in Digisign with the
updated installed reader on 2026-10-31.
