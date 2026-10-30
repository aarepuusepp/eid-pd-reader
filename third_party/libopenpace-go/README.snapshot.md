# libopenpace-go source snapshot

Source: https://bitbucket.org/aarepuusepp/libopenpace-go
Mirror / Go module path: https://github.com/aarepuusepp/libopenpace-go
Version: v1.1.2
Commit: 736dfdea5e501deee953eac5bc0b8419e4c3c7c6

Pinned module sources (apdu, asn1, crypto, kdf, pace, sm), tests, reference
fixtures, module files, license and release documentation. The examples and
repository tooling are not copied. There are no reader-specific source patches.

This snapshot makes the reader build independent of a sibling checkout or
module-proxy publication delay. Update the module requirement, source files,
fixtures and this revision together. The original LICENSE is retained.

Run go test ./... inside this directory as well as at the reader root:
Go deliberately excludes nested modules from the root test pattern.
