# libopenpace-go source snapshot

Source: https://bitbucket.org/aarepuusepp/libopenpace-go
Mirror / Go module path: https://github.com/aarepuusepp/libopenpace-go
Version: v1.1.3
Commit: 032fd6b37ba4e401e6631c5d0ce58e0497d7ed05

Pinned module sources (apdu, asn1, crypto, kdf, pace, sm), tests, reference
fixtures, module files, license and release documentation. The examples and
repository tooling are not copied. All module source files match this commit.

This snapshot makes the reader build independent of a sibling checkout or
module-proxy publication delay. Update the module requirement, source files,
fixtures and this revision together. The original LICENSE is retained.

Run go test ./... and go vet ./... inside this directory as well as at the
reader root: Go deliberately excludes nested modules from the root pattern.
