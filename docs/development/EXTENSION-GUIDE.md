# Browser Extension Developer Guide

Complete guide for developing and integrating the Estonian ID Card Personal Data Reader browser extension.

## Table of Contents

- [Architecture](#architecture)
- [API Reference](#api-reference)
- [Integration Examples](#integration-examples)
- [Development Setup](#development-setup)
- [Testing](#testing)
- [Deployment](#deployment)

## Architecture

### Components Overview

```
┌─────────────────┐
│   Web Page      │  Uses window.EstEID_PD API
│   (Your App)    │
└────────┬────────┘
         │ postMessage
         ▼
┌─────────────────┐
│ Content Script  │  Bridges page ↔ extension
│ (Injected)      │
└────────┬────────┘
         │ chrome.runtime.sendMessage
         ▼
┌─────────────────┐
│ Service Worker  │  Manages state & caching
│ (Background)    │
└────────┬────────┘
         │ chrome.runtime.sendNativeMessage
         ▼
┌─────────────────┐
│ Native Helper   │  Reads ID card via PC/SC
│ (eid-pd-reader.exe)│
└─────────────────┘
```

### File Structure

```
extension/
├── manifest.json              # Extension configuration
├── background.js              # Service worker (Manifest V3)
├── content-bridge.js          # Content script (injected)
├── page-api.js                # Page-level API (injected)
├── web-eid-fallback.js        # web-eid.js wrapper (optional)
├── popup/
│   ├── popup.html            # Extension popup UI
│   ├── popup.css             # Popup styles
│   └── popup.js              # Popup logic
└── icons/
    ├── icon16.png            # Toolbar icon
    ├── icon32.png            # Windows icon
    ├── icon48.png            # Extension management
    └── icon128.png           # Chrome Web Store
```

## API Reference

### Page-Level API (`window.EstEID_PD`)

The extension injects a global `EstEID_PD` object into every web page.

#### `readPersonalData(options)`

Read personal data from Estonian ID card.

**Parameters:**
- `options` (Object, optional)
  - `fids` (Object): Custom FID mapping (advanced)
  - `useCache` (Boolean): Use cached data if available (default: `true`)
  - `timeout` (Number): Request timeout in milliseconds (default: `30000`)

**Returns:** `Promise<Object>`

**Success Response:**
```javascript
{
  ok: true,
  data: {
    surname: "JÕEORG",
    givenName: "JAAK-KRISTJAN",
    sex: "M",
    citizenship: "EST",
    birthDate: "01 08 1980",
    personalCode: "38001085718",
    documentNr: "AT0000222",
    expiryDate: "09 12 2029",
    issueDate: "09 12 2024",
    source: "chip:Thales"
  },
  cached: false,      // true if from cache
  cacheAge: 0         // milliseconds since cached
}
```

**Error Response:**
```javascript
{
  ok: false,
  error: "Native helper not installed. Please run the installer."
}
```

**Example:**
```javascript
const result = await EstEID_PD.readPersonalData();
if (result.ok) {
  console.log('Name:', result.data.givenName, result.data.surname);
  console.log('Personal Code:', result.data.personalCode);
} else {
  console.error('Error:', result.error);
}
```

#### `checkStatus()`

Check if extension and native helper are available.

**Returns:** `Promise<Object>`

**Response:**
```javascript
{
  ok: true,
  available: true,    // Native helper is available
  cached: false,      // Cached data exists
  cacheAge: null      // Age of cached data (ms)
}
```

**Example:**
```javascript
const status = await EstEID_PD.checkStatus();
if (status.ok && status.available) {
  // Show "Read ID Card" button
  document.getElementById('readBtn').disabled = false;
} else {
  // Show installation instructions
  showInstallationGuide();
}
```

#### `clearCache()`

Clear cached personal data.

**Returns:** `Promise<Object>`

**Example:**
```javascript
await EstEID_PD.clearCache();
console.log('Cache cleared');
```

#### `isAvailable()`

Synchronously check if extension is loaded.

**Returns:** `Boolean`

**Example:**
```javascript
if (EstEID_PD.isAvailable()) {
  console.log('Extension is ready');
}
```

#### `version`

Extension API version (string).

**Example:**
```javascript
console.log('EstEID_PD version:', EstEID_PD.version);
```

### Events

#### `esteid-pd-ready`

Fired when the extension API is initialized.

**Example:**
```javascript
window.addEventListener('esteid-pd-ready', (event) => {
  console.log('EstEID_PD ready, version:', event.detail.version);
  initializeApp();
});
```

## Integration Examples

### Example 1: Simple Form Auto-Fill

```html
<!DOCTYPE html>
<html>
<head>
  <title>Registration Form</title>
</head>
<body>
  <form id="registrationForm">
    <label>First Name: <input id="firstName" required></label><br>
    <label>Last Name: <input id="lastName" required></label><br>
    <label>Personal Code: <input id="personalCode" required></label><br>
    <button type="button" id="readCardBtn">📇 Read from ID Card</button>
    <button type="submit">Submit</button>
  </form>

  <script>
    document.getElementById('readCardBtn').addEventListener('click', async () => {
      if (!window.EstEID_PD) {
        alert('Please install the EstEID extension');
        return;
      }

      const result = await EstEID_PD.readPersonalData();
      if (result.ok) {
        document.getElementById('firstName').value = result.data.givenName;
        document.getElementById('lastName').value = result.data.surname;
        document.getElementById('personalCode').value = result.data.personalCode;
      } else {
        alert('Error: ' + result.error);
      }
    });
  </script>
</body>
</html>
```

### Example 2: Authentication with Fallback

```javascript
async function authenticateUser() {
  // Check if extension is available
  if (!window.EstEID_PD) {
    return await traditionalLogin();
  }

  const status = await EstEID_PD.checkStatus();
  if (!status.available) {
    return await traditionalLogin();
  }

  // Try to read ID card
  const result = await EstEID_PD.readPersonalData();
  if (result.ok) {
    // Send to backend for verification
    const response = await fetch('/api/auth/id-card', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        personalCode: result.data.personalCode,
        givenName: result.data.givenName,
        surname: result.data.surname,
        source: result.data.source
      })
    });

    if (response.ok) {
      const session = await response.json();
      return session;
    }
  }

  // Fallback to traditional login
  return await traditionalLogin();
}
```

### Example 3: React Integration

```jsx
import React, { useState, useEffect } from 'react';

function IDCardReader() {
  const [available, setAvailable] = useState(false);
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    // Check if extension is available
    const checkExtension = async () => {
      if (window.EstEID_PD) {
        const status = await window.EstEID_PD.checkStatus();
        setAvailable(status.ok && status.available);
      }
    };

    // Listen for extension ready event
    window.addEventListener('esteid-pd-ready', checkExtension);
    checkExtension();

    return () => {
      window.removeEventListener('esteid-pd-ready', checkExtension);
    };
  }, []);

  const readCard = async () => {
    setLoading(true);
    setError(null);

    try {
      const result = await window.EstEID_PD.readPersonalData();
      if (result.ok) {
        setData(result.data);
      } else {
        setError(result.error);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  if (!available) {
    return <div>Extension not available. Please install it.</div>;
  }

  return (
    <div>
      <button onClick={readCard} disabled={loading}>
        {loading ? 'Reading...' : 'Read ID Card'}
      </button>

      {error && <div className="error">{error}</div>}

      {data && (
        <div className="data">
          <p>Name: {data.givenName} {data.surname}</p>
          <p>Personal Code: {data.personalCode}</p>
          <p>Document: {data.documentNr}</p>
        </div>
      )}
    </div>
  );
}

export default IDCardReader;
```

### Example 4: Vue.js Integration

```vue
<template>
  <div class="id-card-reader">
    <button @click="readCard" :disabled="!available || loading">
      {{ loading ? 'Reading...' : 'Read ID Card' }}
    </button>

    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="data" class="data">
      <p>Name: {{ data.givenName }} {{ data.surname }}</p>
      <p>Personal Code: {{ data.personalCode }}</p>
      <p>Document: {{ data.documentNr }}</p>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      available: false,
      loading: false,
      data: null,
      error: null
    };
  },

  mounted() {
    this.checkExtension();
    window.addEventListener('esteid-pd-ready', this.checkExtension);
  },

  beforeUnmount() {
    window.removeEventListener('esteid-pd-ready', this.checkExtension);
  },

  methods: {
    async checkExtension() {
      if (window.EstEID_PD) {
        const status = await window.EstEID_PD.checkStatus();
        this.available = status.ok && status.available;
      }
    },

    async readCard() {
      this.loading = true;
      this.error = null;

      try {
        const result = await window.EstEID_PD.readPersonalData();
        if (result.ok) {
          this.data = result.data;
        } else {
          this.error = result.error;
        }
      } catch (err) {
        this.error = err.message;
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>
```

## Development Setup

### 1. Install Dependencies

```bash
# No npm dependencies needed!
# Extension uses vanilla JavaScript
```

### 2. Load Extension in Browser

```bash
# Chrome/Edge
1. Open chrome://extensions/
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select the "extension" folder
```

### 3. Enable Console Logging

The extension logs all operations to the console:

```javascript
// Background service worker console
chrome://extensions/ → "Inspect views: service worker"

// Content script console
F12 on any page → Console tab → Filter: "[EstEID-PD"

// Page API console
F12 on any page → Console tab
```

### 4. Hot Reload

Changes to most files require reloading the extension:

```bash
chrome://extensions/ → Click reload icon
# OR
Ctrl+R on the extensions page
```

**Note:** Changes to `manifest.json` require removing and re-adding the extension.

## Testing

### Unit Testing (Manual)

Test individual components in the browser console:

```javascript
// Test status check
await chrome.runtime.sendMessage({ type: 'EID_PD_STATUS' });

// Test reading
await chrome.runtime.sendMessage({ 
  type: 'EID_PD_READ',
  options: { useCache: false }
});

// Test cache clear
await chrome.runtime.sendMessage({ type: 'EID_PD_CLEAR_CACHE' });
```

### Integration Testing

Use the demo page:

```bash
cd demo
python -m http.server 8000
# Open http://localhost:8000
```

### Automated Testing (Future)

Consider adding:
- Jest for JavaScript unit tests
- Puppeteer for browser automation
- Selenium for cross-browser testing

## Deployment

### Development Build

```powershell
# Package for testing
.\ops\package-extension.ps1
```

### Production Build

1. **Update version** in `manifest.json`

2. **Create icons** (if not done):
   ```powershell
   cd extension/icons
   .\generate-icons.ps1
   # Then create PNGs from the SVG
   ```

3. **Package**:
   ```powershell
   .\ops\package-extension.ps1 -IncludeIcons
   ```

4. **Test the package**:
   - Load the .zip in Chrome
   - Test all functionality
   - Check console for errors

### Chrome Web Store

1. Create a developer account: https://chrome.google.com/webstore/devconsole

2. Prepare store listing:
   - Screenshots (1280x800 or 640x400)
   - Promotional images
   - Description
   - Privacy policy

3. Upload the .zip file

4. Submit for review

### Self-Hosting

For enterprise deployment:

1. Host the .crx file on your server
2. Create an update manifest XML
3. Configure via Group Policy (Windows) or MDM

## Best Practices

### Security

1. **Validate all data** from the native helper
2. **Use HTTPS** for any network communication
3. **Don't log sensitive data** in production
4. **Clear cache** when user logs out

### Performance

1. **Use caching** for repeated reads (5-minute default)
2. **Debounce** rapid requests
3. **Show loading states** during card reading

### User Experience

1. **Check availability** before showing ID card options
2. **Provide clear error messages**
3. **Show installation instructions** if extension missing
4. **Handle card removal** gracefully

### Error Handling

```javascript
async function safeReadCard() {
  try {
    // Check extension
    if (!window.EstEID_PD) {
      throw new Error('Extension not installed');
    }

    // Check status
    const status = await EstEID_PD.checkStatus();
    if (!status.available) {
      throw new Error('Native helper not available');
    }

    // Read card
    const result = await EstEID_PD.readPersonalData();
    if (!result.ok) {
      throw new Error(result.error);
    }

    return result.data;
  } catch (error) {
    console.error('ID card reading failed:', error);
    // Show user-friendly message
    showError('Failed to read ID card. Please check that:\n' +
              '1. Card is inserted\n' +
              '2. Extension is installed\n' +
              '3. Native helper is installed');
    return null;
  }
}
```

## Troubleshooting

See [EXTENSION-INSTALL.md](EXTENSION-INSTALL.md#troubleshooting) for common issues and solutions.

## Contributing

Contributions welcome! Please:

1. Test thoroughly in multiple browsers
2. Follow existing code style
3. Add console logging for debugging
4. Update documentation
5. Submit pull request

## License

[Your license here]

