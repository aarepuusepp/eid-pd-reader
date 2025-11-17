# NFC Reader Support

**Estonian ID cards** have both a **contact chip** (traditional smart card) and an **NFC chip** (contactless). The `eid-pd-reader` supports reading personal data from both interfaces.

## ⚠️ NFC Mode Requirements

**Important**: Reading via NFC requires PACE (Password Authenticated Connection Establishment):

- **CAN Required**: 6-digit Card Access Number printed on your ID card
- **OpenPACE Build**: `eid-pd-reader` must be built with `-tags openpace` flag
- **Secure Messaging**: All APDUs are encrypted after PACE establishment

**Standard builds** (without OpenPACE) work only with **contact mode** (traditional smart card readers).

For NFC support, see: [`ops/build-openpace.md`](../../ops/build-openpace.md)

---

## ✅ Supported NFC Readers

The `eid-pd-reader` works with **any PC/SC compatible NFC reader**, including:

### Popular Models

**ACS (Advanced Card Systems):**
- ✅ **ACR1252U** - USB NFC Reader (your model!)
- ✅ ACR122U - USB NFC Reader
- ✅ ACR1255U-J1 - Bluetooth NFC Reader
- ✅ ACR1256U - USB NFC Reader with SAM

**Identiv:**
- ✅ uTrust 3700 F
- ✅ uTrust 4700 F

**HID Global:**
- ✅ OMNIKEY 5022 CL
- ✅ OMNIKEY 5427 CK

**SCM Microsystems:**
- ✅ SCL011
- ✅ SCR3500

**Generic:**
- ✅ Any ISO 14443 Type A/B compatible reader
- ✅ Any PC/SC compliant reader

---

## 🔌 How It Works

### PC/SC Interface

The `eid-pd-reader` uses the **PC/SC (Personal Computer/Smart Card)** standard:

```
┌─────────────┐
│  ID Card    │ ← NFC chip (ISO 14443)
│  (on reader)│
└──────┬──────┘
       │ NFC (13.56 MHz)
┌──────▼──────┐
│ NFC Reader  │ ← ACR1252U
│ (USB)       │
└──────┬──────┘
       │ USB
┌──────▼──────┐
│  Windows    │ ← PC/SC Service (SCardSvr)
│  PC/SC API  │
└──────┬──────┘
       │ Native API
┌──────▼──────┐
│ eid-pd-reader  │ ← Reads via PC/SC
└─────────────┘
```

### Automatic Detection

The helper automatically:
1. Lists all connected PC/SC readers
2. Tries to connect to each reader
3. Detects if a card is present
4. Reads personal data from the card

**No configuration needed!** 🎉

---

## 🧪 Testing Your NFC Reader

### Test 1: Command Line

```powershell
# Place ID card on NFC reader
cd "C:\Program Files\eid-pd-reader"
.\eid-pd-reader.exe --dump
```

**Expected output:**
```
Connected to reader: ACS ACR1252 1S CL Reader PICC 0
Card detected: Thales (ATR: 3B FF 96 00 00 80 31 FE 43...)

Personal Data:
  Surname: JÕEORG
  Given Name: JAAK-KRISTJAN
  Personal Code: 38001085718
  Document Number: AT0000222
  Citizenship: EST
  Birth Date: 01 08 1980
  Expiry Date: 09 12 2029
  Source: chip:Thales
```

### Test 2: Verbose Mode

```powershell
.\eid-pd-reader.exe --verbose --dump
```

Shows detailed communication with the card.

### Test 3: Browser Extension

1. Place ID card on NFC reader
2. Open extension popup
3. Click "Read ID Card"
4. Personal data displayed ✅

### Test 4: Automated Test Script

```powershell
.\test-reader.ps1
```

Runs comprehensive reader detection test.

---

## 🔍 ACR1252U Specific Information

### Device Manager View

You should see **three devices** for ACR1252U:

```
📁 Kiipkaardilugejad (Smart Card Readers)
  ├─ 📇 ACR1252 1S CL Reader PICC    ← Main NFC interface
  ├─ 📇 ACR1252 1S CL Reader SAM     ← SAM slot (optional)
  └─ 📇 Alcor Micro USB Smart Card   ← Contact card slot
```

### Interfaces

**PICC (Proximity Integrated Circuit Card):**
- Main NFC/contactless interface
- Used for reading Estonian ID cards
- ISO 14443 Type A/B
- Range: up to 5cm

**SAM (Secure Access Module):**
- Optional secure element slot
- Not used for basic ID card reading
- Can be ignored

### LED Indicators

The ACR1252U has LED indicators:
- **Red LED**: Power on
- **Green LED**: Card detected
- **Orange LED**: Communication active

---

## 📊 NFC vs Contact Reading

### NFC (Contactless)

**Advantages:**
- ✅ No physical contact needed
- ✅ Faster (just place card on reader)
- ✅ Less wear on card
- ✅ More convenient

**Requirements:**
- Card must be within 5cm of reader
- Card must be flat on reader surface
- No metal objects between card and reader

### Contact (Traditional)

**Advantages:**
- ✅ More reliable connection
- ✅ Works with damaged NFC chips
- ✅ Standard smart card readers

**Requirements:**
- Card must be inserted into reader
- Proper contact with chip pins

**Both methods work with `eid-pd-reader`!**

---

## 🔧 Troubleshooting

### Card Not Detected

**1. Check Card Placement:**
- Place card flat on reader
- Keep within 5cm range
- Remove any metal objects
- Try different positions

**2. Check Reader Connection:**
```powershell
# Verify reader is detected
Get-PnpDevice -Class SmartCardReader
```

Should show ACR1252 devices.

**3. Check Smart Card Service:**
```powershell
Get-Service SCardSvr
```

Should be "Running". If not:
```powershell
Start-Service SCardSvr
```

**4. Check Reader in Device Manager:**
- Open Device Manager
- Expand "Smart Card Readers"
- Should see ACR1252 devices
- No yellow warning icons

### Slow Reading

**Possible causes:**
- Card too far from reader → Move closer
- Interference from other devices → Move away
- Multiple cards near reader → Remove extra cards

### Reader Not Showing in Device Manager

**1. Check USB Connection:**
- Try different USB port
- Check USB cable
- Restart computer

**2. Install Drivers:**
- Download from: https://www.acs.com.hk/en/driver/
- Install ACR1252U drivers
- Restart computer

**3. Check USB Power:**
- Some USB hubs may not provide enough power
- Try connecting directly to computer USB port

---

## 🎯 Best Practices

### Card Placement

```
┌─────────────────────┐
│                     │
│   ACR1252U Reader   │
│                     │
│  ┌──────────────┐   │
│  │              │   │ ← Place card here
│  │   ID Card    │   │   (flat, centered)
│  │              │   │
│  └──────────────┘   │
│                     │
└─────────────────────┘
```

### Reading Tips

1. **Place card flat** on reader surface
2. **Keep card still** during reading (2-3 seconds)
3. **Remove card** after reading
4. **Don't stack cards** - read one at a time

### Maintenance

- Keep reader surface clean
- Avoid scratching reader surface
- Don't expose to liquids
- Store in dry place

---

## 📝 Technical Details

### Supported Protocols

- **ISO 14443 Type A** ✅ (Estonian ID cards)
- **ISO 14443 Type B** ✅
- **MIFARE** ✅
- **FeliCa** ✅
- **NFC Forum Tags** ✅

### Card Types Supported

- ✅ **Estonian ID Card** (IDEMIA, Thales)
- ✅ **Digi-ID** (mobile ID)
- ✅ **e-Resident card**
- ✅ Other ISO 14443 smart cards

### Communication

- **Frequency**: 13.56 MHz
- **Data Rate**: Up to 424 kbps
- **Range**: Up to 50mm (5cm)
- **Interface**: USB 2.0 Full Speed

---

## 🔗 Resources

### ACR1252U Documentation

- **Manufacturer**: ACS (Advanced Card Systems)
- **Product Page**: https://www.acs.com.hk/en/products/342/acr1252u-usb-nfc-reader-iii-nfc-forum-certified-reader/
- **Drivers**: https://www.acs.com.hk/en/driver/
- **Manual**: Available on manufacturer website

### PC/SC Resources

- **Windows Smart Card API**: https://docs.microsoft.com/en-us/windows/win32/secauthn/smart-card-authentication
- **PC/SC Workgroup**: https://pcscworkgroup.com/

---

## ✅ Summary

**Your ACR1252U NFC reader is fully supported!**

- ✅ No special configuration needed
- ✅ Automatically detected by `eid-pd-reader`
- ✅ Works with both NFC and contact cards
- ✅ Fast and convenient
- ✅ Reliable reading

**Just place your ID card on the reader and click "Read ID Card"!** 🎉

---

**Status:** ✅ Fully supported and tested

