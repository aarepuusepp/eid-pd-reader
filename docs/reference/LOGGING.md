# Logging Guide

## Overview

The `eid-pd-reader` application automatically logs all card read operations to a log file. This helps with:
- Auditing card usage
- Troubleshooting issues
- Monitoring application activity
- Compliance and security tracking

## Log File Location

### Windows
```
%LOCALAPPDATA%\eid-pd-reader\eid-pd-reader.log
```

**Example:**
```
C:\Users\YourUsername\AppData\Local\eid-pd-reader\eid-pd-reader.log
```

### Quick Access
```powershell
# Open log directory
explorer "$env:LOCALAPPDATA\eid-pd-reader"

# View log file
notepad "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log"

# Tail log file (PowerShell)
Get-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log" -Wait -Tail 10
```

## What Gets Logged

### Card Read Operations
Every successful card read logs:
- **Timestamp** - When the card was read
- **Personal Code** - Estonian personal identification code
- **Name** - Surname and given name
- **Document Number** - ID card number
- **Source** - Which card type/method was used (e.g., `chip:Thales`, `chip:IDEMIA`)

**Example:**
```
[2025-11-13 15:02:22.519] [INFO] Card read successful | personalCode=38001085718 surname=JÕEORG givenName=JAAK-KRISTJAN documentNr=AT0000222 source=chip:Thales
```

### Failed Read Attempts
Failed card reads are also logged:
```
[2025-11-13 15:03:45.123] [ERROR] Card read failed | error=CARD_NOT_FOUND
```

### Native Messaging Activity
When used as a browser extension helper:
```
[2025-11-13 15:04:10.456] [INFO] Native messaging host started | pid=12345
[2025-11-13 15:04:11.789] [INFO] Native message | direction=request command=getPersonalData success=true
[2025-11-13 15:04:12.012] [INFO] Native message | direction=response command=getPersonalData success=true
```

### Application Lifecycle
```
[2025-11-13 15:02:21.875] [INFO] Logger initialized | logPath=C:\Users\...\eid-pd-reader.log pid=32340
[2025-11-13 15:05:30.123] [INFO] Native messaging host stopped | reason=EOF
```

## Log Format

Each log entry follows this format:
```
[TIMESTAMP] [LEVEL] MESSAGE | key1=value1 key2=value2 ...
```

**Components:**
- **TIMESTAMP**: `YYYY-MM-DD HH:MM:SS.mmm` (24-hour format with milliseconds)
- **LEVEL**: `INFO`, `WARN`, `ERROR`, or `DEBUG`
- **MESSAGE**: Human-readable description
- **KEY=VALUE**: Additional structured data

## Command-Line Options

### Custom Log Path
```bash
# Specify custom log file location
eid-pd-reader.exe --log "C:\MyLogs\custom.log" --dump
```

### Disable Logging
```bash
# Run without logging
eid-pd-reader.exe --nolog --dump
```

### View Help
```bash
eid-pd-reader.exe --help
```

**Available flags:**
```
  --dump          Dump personal data to stdout as JSON and exit
  --debug         Debug mode: show card info and try different FIDs
  --verbose       Verbose mode: show detailed APDU commands and responses
  --log string    Log file path (default: %LOCALAPPDATA%\eid-pd-reader\eid-pd-reader.log)
  --nolog         Disable logging
```

## Log Management

### Viewing Logs

#### PowerShell
```powershell
# View entire log
Get-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log"

# View last 20 lines
Get-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log" -Tail 20

# Follow log in real-time
Get-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log" -Wait -Tail 10

# Search for specific personal code
Get-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log" | Select-String "38001085718"

# Show only errors
Get-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log" | Select-String "\[ERROR\]"

# Show today's entries
Get-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log" | Select-String (Get-Date -Format "yyyy-MM-dd")
```

#### Command Prompt
```cmd
REM View log file
type "%LOCALAPPDATA%\eid-pd-reader\eid-pd-reader.log"

REM View last 10 lines
powershell -Command "Get-Content '%LOCALAPPDATA%\eid-pd-reader\eid-pd-reader.log' -Tail 10"
```

### Log Rotation

The log file grows with each card read. To manage size:

#### Manual Rotation
```powershell
# Archive old log
$logPath = "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log"
$archivePath = "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"
Move-Item $logPath $archivePath

# Compress old logs
Compress-Archive -Path "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader-*.log" -DestinationPath "$env:LOCALAPPDATA\eid-pd-reader\archive-$(Get-Date -Format 'yyyyMM').zip"
```

#### Automated Rotation Script
Create `rotate-logs.ps1`:
```powershell
$logDir = "$env:LOCALAPPDATA\eid-pd-reader"
$logFile = "$logDir\eid-pd-reader.log"
$maxSizeMB = 10

if (Test-Path $logFile) {
    $size = (Get-Item $logFile).Length / 1MB
    if ($size -gt $maxSizeMB) {
        $timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
        Move-Item $logFile "$logDir\eid-pd-reader-$timestamp.log"
        Write-Host "Log rotated: eid-pd-reader-$timestamp.log"
    }
}
```

Run via Task Scheduler daily/weekly.

### Clearing Logs

```powershell
# Delete log file
Remove-Item "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log"

# Clear log content (keep file)
Clear-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log"

# Delete all old logs
Remove-Item "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader-*.log"
```

## Privacy Considerations

### What is Logged
✅ **Logged:**
- Personal code
- Name (surname, given name)
- Document number
- Timestamp
- Card type

❌ **NOT Logged:**
- Certificate data
- PIN codes
- Private keys
- Biometric data
- Any sensitive cryptographic material

### GDPR Compliance

The log file contains personal data. Consider:

1. **Access Control**: Restrict access to log files
2. **Retention**: Implement log rotation and deletion policies
3. **Purpose**: Document why logging is necessary
4. **Consent**: Inform users about logging
5. **Security**: Encrypt logs if required

### Disabling Logging

If logging is not needed:

1. **Temporary**: Use `--nolog` flag
2. **Permanent**: Modify application or use wrapper script

```powershell
# Wrapper script that always disables logging
& "C:\Program Files\eid-pd-reader\eid-pd-reader.exe" --nolog $args
```

## Troubleshooting

### Log File Not Created

**Possible causes:**
1. No write permissions to `%LOCALAPPDATA%`
2. Disk full
3. Antivirus blocking

**Solutions:**
```powershell
# Check if directory exists
Test-Path "$env:LOCALAPPDATA\eid-pd-reader"

# Create directory manually
New-Item -Path "$env:LOCALAPPDATA\eid-pd-reader" -ItemType Directory -Force

# Use custom log path
eid-pd-reader.exe --log "C:\Temp\eid.log" --dump
```

### Log File Too Large

```powershell
# Check size
(Get-Item "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log").Length / 1MB

# Rotate if > 10MB
# (See Log Rotation section above)
```

### Can't Find Log File

```powershell
# Find log file
Get-ChildItem -Path "$env:LOCALAPPDATA" -Filter "eid-pd-reader.log" -Recurse

# Check actual path
[Environment]::GetFolderPath('LocalApplicationData')
```

## Integration with Monitoring Tools

### Windows Event Log
To forward logs to Windows Event Log, use a log shipper like:
- NXLog
- Filebeat
- Splunk Universal Forwarder

### Centralized Logging
For enterprise deployments:
1. Configure log shipper on each workstation
2. Forward logs to central SIEM/log server
3. Set up alerts for suspicious activity

### Example: Filebeat Configuration
```yaml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - C:\Users\*\AppData\Local\eid-pd-reader\eid-pd-reader.log
  fields:
    application: eid-pd-reader
    environment: production

output.elasticsearch:
  hosts: ["your-elasticsearch-server:9200"]
```

## Log Analysis

### Common Queries

```powershell
$log = "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log"

# Count total card reads
(Get-Content $log | Select-String "Card read successful").Count

# List unique personal codes
Get-Content $log | Select-String "personalCode=(\d+)" | 
  ForEach-Object { $_.Matches.Groups[1].Value } | 
  Sort-Object -Unique

# Show read frequency by hour
Get-Content $log | Select-String "\[(\d{4}-\d{2}-\d{2} \d{2}):" | 
  ForEach-Object { $_.Matches.Groups[1].Value } | 
  Group-Object | 
  Sort-Object Name

# Find errors
Get-Content $log | Select-String "\[ERROR\]"
```

### Export to CSV

```powershell
# Parse log entries to CSV
Get-Content "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log" | 
  Select-String "Card read successful" | 
  ForEach-Object {
    if ($_ -match '\[(.+?)\].*personalCode=(\d+) surname=(.+?) givenName=(.+?) documentNr=(.+?) source=(.+)$') {
      [PSCustomObject]@{
        Timestamp = $matches[1]
        PersonalCode = $matches[2]
        Surname = $matches[3]
        GivenName = $matches[4]
        DocumentNr = $matches[5]
        Source = $matches[6]
      }
    }
  } | Export-Csv -Path "card-reads.csv" -NoTypeInformation
```

## Security Best Practices

1. **Restrict Access**: Set NTFS permissions on log directory
   ```powershell
   icacls "$env:LOCALAPPDATA\eid-pd-reader" /grant:r "$env:USERNAME:(OI)(CI)F" /inheritance:r
   ```

2. **Encrypt Logs**: Use EFS (Encrypting File System)
   ```powershell
   cipher /e "$env:LOCALAPPDATA\eid-pd-reader"
   ```

3. **Regular Rotation**: Implement automated log rotation

4. **Audit Access**: Monitor who accesses log files

5. **Secure Deletion**: Use secure delete tools when removing logs
   ```powershell
   # Overwrite before delete
   $logPath = "$env:LOCALAPPDATA\eid-pd-reader\eid-pd-reader.log"
   $random = New-Object byte[] (Get-Item $logPath).Length
   (New-Object Random).NextBytes($random)
   [IO.File]::WriteAllBytes($logPath, $random)
   Remove-Item $logPath
   ```

## See Also

- `README.md` - Main documentation
- `TROUBLESHOOTING.md` - Common issues
- `QUICK-START.md` - Quick setup guide

