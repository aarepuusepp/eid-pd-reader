# eid-pd-reader - Estonian ID Card Personal Data Reader

Native helper application for reading personal data from Estonian ID cards.

## Overview

The eid-pd-reader is a Go-based native application that communicates with PC/SC card readers to extract personal data from Estonian ID cards. It supports both contact and contactless (NFC) reading modes.

## Features

- **Contact Reading:** Direct reading via contact card readers
- **NFC Reading:** Built-in contactless reading with pure Go PACE protocol implementation
- **Multiple Card Types:** Support for IDEMIA, Thales, and other Estonian ID cards
- **Configurable:** YAML-based configuration
- **Cross-Platform:** Windows, Linux, macOS
- **Pure Go:** No C dependencies except PC/SC (uses libopenpace-go for PACE protocol)

## Installation

### Windows

```cmd
install-reader.cmd
```

This will:
- Copy the executable to `C:\Program Files\eid-pd-reader\`
- Install host manifests for Chrome/Edge and Firefox
- Register with Windows Registry

### Linux

```bash
./ops/install-host.sh
```

### macOS

```bash
./ops/install-host.sh
```

## Configuration

Create a `config.yaml` file:

```yaml
logging:
  enabled: true
  level: info
  file: ""

card:
  timeout_seconds: 30
  retry_attempts: 3

cache:
  enabled: true
  duration_seconds: 300
```

See [Configuration Guide](./docs/guides/CONFIG-GUIDE.md) for details.

## Usage

### Command Line

```bash
# Read card data
eid-pd-reader.exe --dump

# With NFC
eid-pd-reader.exe --dump --nfc --can 123456

# Verbose mode
eid-pd-reader.exe --dump --verbose

# Show version
eid-pd-reader.exe --version
```

### Native Messaging

The reader communicates with the browser extension via native messaging protocol.

Request:
```json
{
  "command": "readPersonalData",
  "mode": "contact"
}
```

Response:
```json
{
    "ok": true,
    "data": {
        "surname": "JÕEORG",
        "givenName": "JAAK-KRISTJAN",
        "sex": "M",
        "citizenship": "EST",
        "birthDate": "01 08 1980",
        "personalCode": "38001085718",
        "documentNr": "AT0000222",
        "expiryDate": "09 12 2029",
        "issueDate": "09 12 2024",
        "source": "chip:Thales"
    }
}
```

## Building

NFC support is built-in using the pure Go `libopenpace-go` library. No special build steps required.

```bash
# Standard build (includes NFC support)
go mod tidy
go build -o eid-pd-reader.exe .

# Or use the build script
.\build.ps1

# Run tests
go test ./...
```

The reader automatically detects if a card requires PACE (NFC) and uses the appropriate protocol.

## Extension ID Configuration

After installing the browser extension, you need to configure the extension ID in the host manifest:

1. Get your extension ID from browser
2. Update `host-manifests/ee.aarepuusepp.eid_pd_reader.json`
3. Replace `YOUR_EXTENSION_ID` with actual ID

See [Extension ID Configuration Guide](./docs/guides/EXTENSION-ID-CONFIGURATION.md) for details.

## Supported Cards

- IDEMIA Cosmo 8.1/8.2
- IDEMIA Cosmo X
- Thales (Gemalto) cards
- Standard Estonian ID cards

See [Supported Cards](./docs/reference/SUPPORTED-CARDS.md) for details.

## Documentation

- [Configuration Guide](./docs/guides/CONFIG-GUIDE.md)
- [Supported Cards](./docs/reference/SUPPORTED-CARDS.md)
- [Troubleshooting](./docs/reference/TROUBLESHOOTING.md)
- [Security](./docs/reference/SECURITY.md)
- [NFC Implementation](./docs/development/NFC-IMPLEMENTATION-SUMMARY.md)
- [Development Docs](./docs/development/)

## Requirements

### Windows
- Windows 10/11 (64-bit)
- PC/SC smart card service (built-in)

### Linux
- PC/SC Lite: `sudo apt-get install pcscd libpcsclite-dev`
- Go 1.25+ (for building from source)

### macOS
- macOS 10.15+ (Catalina or later)
- PC/SC framework (built-in)

## Troubleshooting

### Card Not Detected
- Ensure card reader is connected
- Check PC/SC service is running
- Try `--verbose` flag for detailed logs

### NFC Reading Fails
- Verify CAN (Card Access Number) is correct
- Ensure card is properly positioned on reader
- Check reader supports NFC/contactless

See [Troubleshooting Guide](./docs/reference/TROUBLESHOOTING.md) for more.

## Related Components

- [eid-pd-extension](../eid-pd-extension/) - Browser extension component
- [libopenpace-go](../libopenpace-go/) - PACE protocol library

## License

MIT License

## Support

- **Issues:** [GitHub Issues](https://github.com/aarepuusepp/web-eid-pd/issues)
- **Email:** aare.puusepp@gmail.com

## Version

Current version: 1.0.3

See [Release Notes](./docs/releases/) for changelog.

## Technical Details

### NFC/PACE Support

The reader uses the pure Go `libopenpace-go` library for PACE (Password Authenticated Connection Establishment) protocol support. This means:

- **No C dependencies** for PACE protocol (previously required OpenPACE C library)
- **Easier builds** across all platforms
- **Better integration** with Go toolchain
- **Automatic protocol detection** - the reader detects if PACE is needed and applies it transparently

The PACE implementation supports:
- Brainpool curves (BP256R1, BP320R1, BP384R1, BP512R1)
- AES encryption (AES-128, AES-192, AES-256)
- Automatic CAN (Card Access Number) handling

See [libopenpace-go](../libopenpace-go/) for more details on the PACE implementation.

