# Reader 2.1.4

Updates the pinned libopenpace-go snapshot to 1.1.2:

- Correct AES-128 SHA-1 KDF and secret/nonce/counter ordering.
- Reject incomplete PACE Integrated Mapping, 3DES and unknown protocol OIDs,
  including variants advertised by a card before authentication.
- Preserve AES-128/192/256 Generic Mapping and AES secure messaging.
- Include fixed independent KDF vectors and protocol rejection tests.

The native messaging protocol remains version 1. Browser extensions 2.1.3
remain compatible. This is a reader source release; it does not replace the
already installed workstation package 2.1.3. A new installer must be built
and validated separately before deploying this reader.

Run both the reader tests and the nested source snapshot tests; Go does not
include nested modules in the root ./... test pattern:

    go test -count=1 ./...
    go vet ./...
    cd third_party/libopenpace-go
    go test -count=1 ./...
    go vet ./...

Real NFC/PACE interoperability is not established by contact-card reading or
CDOC v1 PKCS#11 decryption. This patch has not been tested on an NFC card.

## Validation on 2026-10-30

Passed on Windows x64 with Go 1.27.3 and MSYS2 UCRT64 GCC:
- Reader: go test -mod=readonly -p 2 -count=1 ./... and go vet -mod=readonly ./...
- Nested libopenpace-go snapshot: full tests and go vet.
- Native Windows CGO build with -mod=readonly -trimpath; --version reports 2.1.4.
- All 42 copied source, test, fixture and documentation files match the pinned
  upstream commit byte for byte (README.snapshot.md is reader-specific metadata).
- Upstream library: race tests, 20 independent KDF vectors verified with Python
  and OpenSSL, old-defect mutation checks. See the copied VALIDATION-1.1.2.md.

No hardware card operation or browser decryption was initiated during these tests.
