# WebEID Personal Data Fallback (Go native helper + Extension API)

A minimal **native helper** (Go, PC/SC) + **browser extension API** (Manifest V3) to read **Estonian ID‑card Personal Data** files (surname, given name, personal code) directly from the chip **even when certificates are missing**, and expose them to your web app via a safe background bridge.

> The helper reads Personal Data EFs with default EstEID FIDs (Surname=0x5001, GivenName=0x5002, PersonalCode=0x5006). These defaults can be overridden via message payload.

---

## 1) Project layout

```
webeid-pd-fallback/
├─ native/
│  ├─ cmd/eid-pd-reader/main.go
│  ├─ internal/apdu/apdu.go
│  ├─ internal/esteid/reader.go
│  ├─ go.mod
│  └─ go.sum
├─ host-manifests/
│  ├─ ee.aarepuusepp.eid_helper.json          # Chrome/Edge on Linux/macOS
│  ├─ ee.aarepuusepp.eid_helper.firefox.json  # Firefox (all platforms)
│  └─ windows/registry-install.cmd         # Registers Chrome/Edge host on Windows
└─ extension/
   ├─ manifest.json
   ├─ background.js
   ├─ content-bridge.js
   └─ page-api.js
```

---

## 2) Native helper (Go, PC/SC)

### 2.1 `native/go.mod`
```go
module com.example.eid-pd-reader

go 1.22

require (
	github.com/ebfe/scard v0.0.0-20230418175058-5a341b9e398f
)
```

### 2.2 `native/internal/apdu/apdu.go`
```go
package apdu

import (
	fmt
)

// APDU is a simple container for command APDUs
// CLA INS P1 P2 Lc Data Le (Le optional)
type APDU struct {
	CLA  byte
	INS  byte
	P1   byte
	P2   byte
	Data []byte
	Le   int // -1 => no Le
}

func (a APDU) Bytes() []byte {
	b := []byte{a.CLA, a.INS, a.P1, a.P2}
	if len(a.Data) > 0 {
		b = append(b, byte(len(a.Data)))
		b = append(b, a.Data...)
	}
	if a.Le >= 0 {
		b = append(b, byte(a.Le))
	}
	return b
}

// StatusWord splits trailer into SW1/SW2
func StatusWord(resp []byte) (byte, byte) {
	if len(resp) < 2 {
		return 0x6F, 0x00 // generic error
	}
	return resp[len(resp)-2], resp[len(resp)-1]
}

// DataPart returns response data without SW1SW2
func DataPart(resp []byte) []byte {
	if len(resp) < 2 {
		return nil
	}
	return resp[:len(resp)-2]
}

func CheckOK(sw1, sw2 byte) error {
	if sw1 == 0x90 && sw2 == 0x00 {
		return nil
	}
	return fmt.Errorf("card returned error SW=%02X%02X", sw1, sw2)
}
```

### 2.3 `native/internal/esteid/reader.go`
```go
package esteid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ebfe/scard"

	"com.example.eid-pd-reader/internal/apdu"
)

// Default file IDs (FID) for EstEID Personal Data (transparent EFs)
var (
	FID_DF_PERSONAL = [2]byte{0x50, 0x00} // DF '5000'
	FID_SURNAME     = [2]byte{0x50, 0x01} // EF '5001'
	FID_GIVENNAME   = [2]byte{0x50, 0x02} // EF '5002'
	FID_PERSONALCODE= [2]byte{0x50, 0x06} // EF '5006'
)

// PersonalData holds the fields we export
type PersonalData struct {
	Surname      string `json:"surname"`
	GivenName    string `json:"givenName"`
	PersonalCode string `json:"personalCode"`
	Source       string `json:"source"` // "chip:EF500x"
}

// ReadPersonalData reads PD EFs via PC/SC from the first available EstEID card.
// Overrides optionalFIDs can provide custom FIDs in hex like ["5000","5001",...].
func ReadPersonalData(optionalFIDs map[string]string) (*PersonalData, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, fmt.Errorf("pcsc: establish context: %w", err)
	}
	defer ctx.Release()

	rdrs, err := ctx.ListReaders()
	if err != nil {
		return nil, fmt.Errorf("pcsc: list readers: %w", err)
	}
	if len(rdrs) == 0 {
		return nil, fmt.Errorf("pcsc: no readers found")
	}

	var card *scard.Card
	for _, r := range rdrs {
		c, err := ctx.Connect(r, scard.ShareShared, scard.ProtocolAny)
		if err == nil {
			card = c
			break
		}
	}
	if card == nil {
		return nil, fmt.Errorf("pcsc: could not connect to any reader")
	}
	defer card.Disconnect(scard.LeaveCard)

	// SELECT MF 3F00
	if err := selectByFID(card, [2]byte{0x3F, 0x00}); err != nil {
		return nil, fmt.Errorf("select MF: %w", err)
	}

	// DF Personal (5000) — allow override
	pdfid := FID_DF_PERSONAL
	if s, ok := optionalFIDs["df_personal"]; ok {
		if fid, e := parseFID(s); e == nil { pdfid = fid }
	}
	if err := selectDF(card, pdfid); err != nil {
		return nil, fmt.Errorf("select DF 0x%02X%02X: %w", pdfid[0], pdfid[1], err)
	}

	// Read EFs
	surnameFID := FID_SURNAME
	if s, ok := optionalFIDs["ef_surname"]; ok { if fid, e := parseFID(s); e == nil { surnameFID = fid } }
	givenFID := FID_GIVENNAME
	if s, ok := optionalFIDs["ef_given"]; ok { if fid, e := parseFID(s); e == nil { givenFID = fid } }
	codeFID := FID_PERSONALCODE
	if s, ok := optionalFIDs["ef_code"]; ok { if fid, e := parseFID(s); e == nil { codeFID = fid } }

	surname, err := readTransparentEF(card, surnameFID, 256)
	if err != nil { return nil, fmt.Errorf("read EF %02X%02X: %w", surnameFID[0], surnameFID[1], err) }
	given, err := readTransparentEF(card, givenFID, 256)
	if err != nil { return nil, fmt.Errorf("read EF %02X%02X: %w", givenFID[0], givenFID[1], err) }
	code, err := readTransparentEF(card, codeFID, 256)
	if err != nil { return nil, fmt.Errorf("read EF %02X%02X: %w", codeFID[0], codeFID[1], err) }

	pd := &PersonalData{
		Surname:      sanitizeText(surname),
		GivenName:    sanitizeText(given),
		PersonalCode: sanitizeText(code),
		Source:       "chip:EF500x",
	}
	return pd, nil
}

func parseFID(s string) ([2]byte, error) {
	s = strings.TrimPrefix(strings.ToLower(s), "0x")
	if len(s) != 4 { return [2]byte{}, fmt.Errorf("fid must be 4 hex chars") }
	b, err := hex.DecodeString(s)
	if err != nil { return [2]byte{}, err }
	return [2]byte{b[0], b[1]}, nil
}

func selectByFID(card *scard.Card, fid [2]byte) error {
	// 00 A4 00 0C 02 FID
	cmd := apdu.APDU{CLA:0x00, INS:0xA4, P1:0x00, P2:0x0C, Data: []byte{fid[0], fid[1]}, Le:-1}
	resp, err := transmit(card, cmd)
	if err != nil { return err }
	sw1, sw2 := apdu.StatusWord(resp)
	return apdu.CheckOK(sw1, sw2)
}

func selectDF(card *scard.Card, fid [2]byte) error {
	// 00 A4 01 0C 02 FID (select DF by id)
	cmd := apdu.APDU{CLA:0x00, INS:0xA4, P1:0x01, P2:0x0C, Data: []byte{fid[0], fid[1]}, Le:-1}
	resp, err := transmit(card, cmd)
	if err != nil { return err }
	sw1, sw2 := apdu.StatusWord(resp)
	return apdu.CheckOK(sw1, sw2)
}

func selectEF(card *scard.Card, fid [2]byte) error {
	// 00 A4 02 0C 02 FID (select EF by id)
	cmd := apdu.APDU{CLA:0x00, INS:0xA4, P1:0x02, P2:0x0C, Data: []byte{fid[0], fid[1]}, Le:-1}
	resp, err := transmit(card, cmd)
	if err != nil { return err }
	sw1, sw2 := apdu.StatusWord(resp)
	return apdu.CheckOK(sw1, sw2)
}

func readTransparentEF(card *scard.Card, fid [2]byte, max int) ([]byte, error) {
	if err := selectEF(card, fid); err != nil { return nil, err }
	// Read up to max bytes using READ BINARY (00 B0 offHi offLo Le). Most PD fields are short.
	cmd := apdu.APDU{CLA:0x00, INS:0xB0, P1:0x00, P2:0x00, Data:nil, Le: max}
	resp, err := transmit(card, cmd)
	if err != nil { return nil, err }
	sw1, sw2 := apdu.StatusWord(resp)
	if err := apdu.CheckOK(sw1, sw2); err != nil { return nil, err }
	return bytes.TrimRight(apdu.DataPart(resp), "\x00\x20\n\r"), nil
}

func transmit(card *scard.Card, cmd apdu.APDU) ([]byte, error) {
	resp, err := card.Transmit(cmd.Bytes())
	if err != nil {
		return nil, fmt.Errorf("pcsc: transmit % X: %w", cmd.Bytes(), err)
	}
	return resp, nil
}

func sanitizeText(b []byte) string {
	// EstEID PD tends to be ASCII/UTF-8. Remove trailing control chars.
	out := string(bytes.TrimSpace(b))
	// Collapse internal NULs (some cards may pad with 0x00)
	out = strings.ReplaceAll(out, "\x00", "")
	return out
}
```

### 2.4 `native/cmd/eid-pd-reader/main.go`
```go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"com.example.eid-pd-reader/internal/esteid"
)

// Native Messaging protocol: length-prefixed JSON messages over stdin/stdout.
// Request JSON:
// {
//   "cmd": "readPersonalData",
//   "fids": {"df_personal":"5000","ef_surname":"5001","ef_given":"5002","ef_code":"5006"}
// }
//
// Response JSON:
// {"ok":true, "data": {surname, givenName, personalCode, source}}
// or {"ok":false, "error":"..."}

type request struct {
	Cmd  string            `json:"cmd"`
	FIDs map[string]string `json:"fids"`
}

type response struct {
	OK    bool                    `json:"ok"`
	Error string                  `json:"error,omitempty"`
	Data  *esteid.PersonalData    `json:"data,omitempty"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, err := readNativeMessage(reader)
		if err != nil {
			return // EOF or fatal
		}
		var req request
		if err := json.Unmarshal(msg, &req); err != nil {
			_ = writeNativeMessage(response{OK:false, Error:fmt.Sprintf("bad request: %v", err)})
			continue
		}
		switch req.Cmd {
		case "readPersonalData":
			pd, err := esteid.ReadPersonalData(req.FIDs)
			if err != nil {
				_ = writeNativeMessage(response{OK:false, Error:err.Error()})
				continue
			}
			_ = writeNativeMessage(response{OK:true, Data:pd})
		default:
			_ = writeNativeMessage(response{OK:false, Error:"unknown cmd"})
		}
	}
}

// --- Native messaging framing (4-byte little endian length) ---
func readNativeMessage(r *bufio.Reader) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := r.Read(lenBuf); err != nil { return nil, err }
	length := int(uint32(lenBuf[0]) | uint32(lenBuf[1])<<8 | uint32(lenBuf[2])<<16 | uint32(lenBuf[3])<<24)
	if length <= 0 || length > 1<<20 { return nil, fmt.Errorf("invalid message length: %d", length) }
	payload := make([]byte, length)
	if _, err := r.Read(payload); err != nil { return nil, err }
	return payload, nil
}

func writeNativeMessage(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil { return err }
	lenBuf := []byte{byte(len(b)), byte(len(b)>>8), byte(len(b)>>16), byte(len(b)>>24)}
	if _, err := os.Stdout.Write(lenBuf); err != nil { return err }
	_, err = os.Stdout.Write(b)
	return err
}
```

### 2.5 Build

```bash
cd native
go mod tidy
# All platforms (output name derived from directory)
go build ./cmd/eid-pd-reader
```

> Requires a PC/SC stack: **Windows Smart Card Service**, **pcscd** (Linux), or **PCSC** (macOS). Ensure your ID-card driver/middleware is installed and a **contact** reader is used.

---

## 3) Native Messaging host manifests

> Replace `/absolute/path/to/eid-pd-reader` with your built binary.

### 3.1 Chrome/Edge (Linux/macOS) – `host-manifests/ee.aarepuusepp.eid_helper.json`
```json
{
  "name": "ee.aarepuusepp.eid_helper",
  "description": "EstEID Personal Data helper",
  "path": "/absolute/path/to/eid-pd-reader",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://__EXTENSION_ID__/"
  ]
}
```

**Install**
- Linux: place under `/etc/opt/chrome/native-messaging-hosts/` (Chrome) or `/etc/chromium/native-messaging-hosts/` (Chromium).
- macOS: `~/Library/Application Support/Google/Chrome/NativeMessagingHosts/` or system-wide under `/Library/Google/Chrome/NativeMessagingHosts/`.

### 3.2 Firefox – `host-manifests/ee.aarepuusepp.eid_helper.firefox.json`
```json
{
  "name": "ee.aarepuusepp.eid_helper",
  "description": "EstEID Personal Data helper",
  "path": "/absolute/path/to/eid-pd-reader",
  "type": "stdio",
  "allowed_extensions": [ "eid-pd@aarepuusepp.ee" ]
}
```

**Install**
- Linux/macOS: `~/.mozilla/native-messaging-hosts/` or `/usr/lib/mozilla/native-messaging-hosts/`.
- Windows: `%APPDATA%\Mozilla\NativeMessagingHosts\`.

### 3.3 Windows (Chrome/Edge) registry – `host-manifests/windows/registry-install.cmd`
```bat
@echo off
REM Adjust the paths below
set HOSTNAME=ee.aarepuusepp.eid_helper
set JSON=%~dp0..\ee.aarepuusepp.eid_helper.json
reg add "HKLM\Software\Google\Chrome\NativeMessagingHosts\%HOSTNAME%" /ve /t REG_SZ /d "%JSON%" /f
reg add "HKLM\Software\Microsoft\Edge\NativeMessagingHosts\%HOSTNAME%" /ve /t REG_SZ /d "%JSON%" /f
echo Registered %HOSTNAME%
```

---

## 4) Browser extension (Manifest V3)

### 4.1 `extension/manifest.json`
```json
{
  "manifest_version": 3,
  "name": "WebEID PD Fallback",
  "version": "0.1.0",
  "description": "Reads Personal Data from EstEID chip via native helper.",
  "background": {"service_worker": "background.js"},
  "host_permissions": ["https://YOUR-APP-ORIGIN/*"],
  "permissions": ["nativeMessaging", "scripting"],
  "externally_connectable": {
    "matches": ["https://YOUR-APP-ORIGIN/*"]
  }
}
```

### 4.2 `extension/background.js`
```js
const HOST = "ee.aarepuusepp.eid_helper";

async function callNative(msg) {
  return new Promise((resolve) => {
    chrome.runtime.sendNativeMessage(HOST, msg, (response) => {
      if (chrome.runtime.lastError) {
        resolve({ ok: false, error: chrome.runtime.lastError.message });
      } else {
        resolve(response || { ok: false, error: "no response" });
      }
    });
  });
}

// Accept messages from the page via content script
chrome.runtime.onMessage.addListener((req, _sender, sendResponse) => {
  if (req && req.type === "EID_PD_READ") {
    (async () => {
      const payload = { cmd: "readPersonalData", fids: req.fids || {} };
      const res = await callNative(payload);
      sendResponse(res);
    })();
    return true; // keep channel open
  }
});
```

### 4.3 `extension/content-bridge.js`
```js
// Bridge between page and extension background.

// Only allow trusted origins (configured in manifest). Optionally, add origin checks here.
window.addEventListener("message", async (event) => {
  if (event.source !== window) return;
  const msg = event.data || {};
  if (msg.target !== "webEID-PD-FALLBACK" || msg.type !== "READ") return;

  chrome.runtime.sendMessage({ type: "EID_PD_READ", fids: msg.fids || {} }, (resp) => {
    window.postMessage({ target: "webEID-PD-FALLBACK", type: "READ_RESULT", payload: resp }, "*");
  });
});
```

### 4.4 `extension/page-api.js`
```js
// Inject this into the page (via <script> or bundler) to expose a simple API.

export function readPersonalData(options = {}) {
  return new Promise((resolve) => {
    function onMessage(ev) {
      const data = ev.data || {};
      if (data.target === "webEID-PD-FALLBACK" && data.type === "READ_RESULT") {
        window.removeEventListener("message", onMessage);
        resolve(data.payload);
      }
    }
    window.addEventListener("message", onMessage);
    window.postMessage({ target: "webEID-PD-FALLBACK", type: "READ", fids: options.fids || {} }, "*");
  });
}
```

**Load the content script** using `chrome.scripting` at install time (or declare as `content_scripts` if you prefer):

```js
// In background.js, on install
chrome.runtime.onInstalled.addListener(() => {
  chrome.scripting.registerContentScripts([
    {
      id: "pd-bridge",
      js: ["content-bridge.js"],
      matches: ["https://YOUR-APP-ORIGIN/*"],
      runAt: "document_start",
      allFrames: false
    }
  ]);
});
```

---

## 5) Web app usage example

```html
<script type="module">
  import { readPersonalData } from "/page-api.js";

  async function fallbackRead() {
    const res = await readPersonalData({});
    if (!res?.ok) {
      console.error("PD read failed:", res?.error);
      alert("ID-kaardi andmete lugemine ebaõnnestus: " + (res?.error || "tundmatu viga"));
      return;
    }
    const { surname, givenName, personalCode } = res.data || {};
    console.log("PD:", surname, givenName, personalCode);
    // Use data in your existing flow
  }

  document.querySelector('#fallback-btn').addEventListener('click', fallbackRead);
</script>
<button id="fallback-btn">Loe isikuandmed kaardilt</button>
```

---

## 6) Security & UX notes

- **Origin pinning**: Limit `externally_connectable.matches` and `host_permissions` in `manifest.json` to your exact app origins.
- **Consent**: Show a clear prompt before calling the helper (your UI button does this explicitly).
- **No raw APDUs in page**: APDU details remain inside the native helper; the page receives only sanitized text.
- **Contact vs NFC**: This helper targets **contact** readers. NFC (contactless) would require PACE and secure messaging; out of scope here.
- **Error handling**: Provide clear user‑facing messages for no reader, no card, card busy, or driver not installed.

---

## 7) Troubleshooting checklist

- PC/SC service is running (`sc query SCardSvr` on Windows; `systemctl status pcscd` on Linux).
- Estonian ID‑card drivers/middleware present; card visible in OS smart card list.
- Host manifest path is absolute and executable is present/launchable.
- Extension installed, and its ID placed in the host manifest `allowed_origins` (Chrome/Edge) or `allowed_extensions` (Firefox).
- For Windows Chrome/Edge, registry values exist under the *Current User* hive as shown above.

---

## 8) Extending fields

To read more EFs later, add new FIDs and small helpers (e.g., birth date, document number) and map them to strings. You may also parse TLV if a field is record‑or TLV‑structured on certain card revisions.

---

## 9) License

MIT (sample code). Replace namespaces, IDs, and origins to your organization.

---

## 10) Mida saab lugeda **ilma PIN1/PIN2 sisestamata** (kontaktliides)

> Üldpõhimõte: **lugemine (Read)** paljudest failidest on märgitud **ALW** (alati lubatud), privaatvõtmete kasutus (allkiri/Autenticate) nõuab vastavalt **PIN2/PIN1**.

**Isikuandmed (DF 5000; PD1…PD15)**  
Uutel ID1 kaartidel on PD-failid läbipaistvad EF-id FID-idega `5001…500F`. Loetavad ilma PINideta: 
- **PD1** perekonnanimi, **PD2** eesnimi
- **PD3** sugu, **PD4** kodakondsus (ISO 3166‑1 alpha‑3)
- **PD5** sünniaeg ja ‑koht
- **PD6** isikukood
- **PD7** dokumendi number, **PD8** kehtivuse lõpp
- **PD9** väljaandmise kuupäev ja asutus (varasematel profiilidel „kuupäev ja koht“)
- **PD10…PD15** märked / elamisloa väljad (kaarditüübist sõltuvalt võivad olla tühjad).  
> Kui väli puudub, eksisteerib vastav EF siiski ja sisaldab **0x00** bait– see on oodatud käitumine.

**Dokumendi number (EF D003)**  
Eraldi läbipaistev EF dokumendinumbriga, loetav ilma PINita.

**Sertifikaadid**  
- **Autentimissert** (DF.AWP) – **Read: ALW**  
- **Allkirjasert** (DF.QSCD) – **Read: ALW**  
Public‑key saad praktiliselt alati kätte sertifikaadist (või GET DATA kaudu vastavalt objekti turvaatribuutidele).

**ISO/IEC 7816‑15 metaandmed (loetavad)**  
- **EF.DIR**, **EF.CardAccess** (PACE parameetrid), **EF.CIAInfo**, **EF.DCOD** kirjeldused jmt – kõik tavaliselt **Read: ALW**.

**PIN loendurid**  
- **EF 0016** (PIN1/PIN2/PUK retry counter) kirjed on loetavad ilma PINita (abiks UX‑is „mitu katset järel“ kuvamiseks).

**Märkus tootjaprofiilide/versioonide kohta**  
- **ID1 (Cosmo v8.x, pärast 2021)** – isikuandmed on EF‑id `5001…500F` (läbipaistvad failid, READ BINARY).  
- **EstEID 3.5.x (enne ID1 ühtlustust)** – isikuandmed asusid **EF 5044** all **record‑failina** (READ RECORD).  
- Mõlemal juhul on sisu loetav ilma PINita; erineb vaid **failistruktuur ja FID‑id**.

---

## 11) NFC (kontaktivaba) lugemine – mida ja kuidas

**Eeltingimus: PACE + turvaline kanal**  
Kontaktivabal liidesel (ISO/IEC 14443) **ei saa** eID‑rakenduse andmeid otse lugeda. Enne tuleb teostada **PACE** (Password Authenticated Connection Establishment) ja avada **Secure Messaging**.  
- Eesti ID‑dokumentidel on PACE seadistatud **ECDH + AES‑CMAC**’iga; domeeniparameetrid on **BrainpoolP384r1**.  
- PACE paroolina kasutatakse **CAN**’i (Card Access Number; trükitud kaardile). Mõne teise dokumenditüübi puhul võib parooliallikas olla **MRZ** (MRTD), kuid klassikalisel ID‑kaardil kasutatakse CAN‑i.

**Pärast PACE’t (aktiivne Secure Messaging):**  
- Sama **DF.PersonalData** (PD1…PD15) ja **sertifikaadid** on **loetavad** nagu kontaktliideses; privaatvõtme operatsioonid **nõuavad jätkuvalt PINe** (PIN1 autentimiseks, PIN2 allkirjastamiseks).  
- **EF.CardAccess** (PACEInfo) on loetav – sealt saab klient enne PACE’t algoritmi/kurvi infot (tihti ka kontaktis).  
- Läbipaistvad meta‑EF‑id (nt **EF.DIR**) on samuti loetavad pärast PACE’t.

**Praktilised soovitused NFC jaoks**  
- Lugeja peab toetama **ISO 14443‑A** ja PACE’i; teegipoolselt kasuta raamistikke, mis oskavad **PACE + SM** (nt Open eID komponentide või spets NFC SDK‑de kaudu).  
- UX‑is küsi **CAN** kasutajalt ning kuva selge juhend kaardi asendamiseks lugejal.

**Piirangud NFC‑s**  
- Ilma PACE’ita (st enne turvalise kanali avamist) ei anta rakenduslikke andmeid välja.  
- **Privaatvõtmete** kasutus (Compute Signature / Internal Authenticate) **ei ole võimalik ilma vastava PINita**; NFC ei muuda seda reeglit.

---

## 12) Kuidas lisada need read olemasolevasse lahendusse

- **Kontaktliides**: juba olemas olev Go‑helper loeb PD EF‑e (5001/5002/5006 jne). Vajadusel lisa **D003** (dokumendinumber), **EF.DIR**, **EF.CardAccess**, **EF.CIAInfo** ja **EF.DCOD** lugemised ning **EF 0016** retry‑loendurite kuvamine.  
- **NFC tugi**: eraldi helper/daemon, mis teeb **PACE (CAN)**, avab **SM** ja ekspordib samad PD väljad/sertifikaadid. API jääb sinu web‑extensioni jaoks samaks; muutub ainult transpordi „back‑end“.

---

## 13) Feature‑flag Fallback: web‑eid.js → (vajadusel) native helper

**Eesmärk:** automaatne tee, mis proovib esmalt **web‑eid.js** kaudu sertifikaadipõhist isikuinfo hankimist; kui **sertifikaate pole** või kutsung **ebaõnnestub**, kasutatakse **native helperit** (kontakt või NFC back‑end).

### 13.1 Konfiguratsioon (feature‑flag)

```js
// config/pd-fallback.config.js
export const PD_FALLBACK_ENABLED = true;             // peal/maas
export const PD_FALLBACK_TIMEOUT_MS = 5000;          // kaua oodata web-eid vastust
export const PD_FIELDS = {                           // DF/EF FID override võimalus
  df_personal: "5000",
  ef_surname:  "5001",
  ef_given:    "5002",
  ef_code:     "5006"
};
```

### 13.2 Lehe poole API: esmalt web‑eid, siis helper

```js
// page-api-auto.js
import { PD_FALLBACK_ENABLED, PD_FALLBACK_TIMEOUT_MS, PD_FIELDS } from "./config/pd-fallback.config.js";
import { readPersonalData } from "./page-api.js"; // juba olemas (kutsuv sild ekstensionini)

// Adapter sinu olemasolevale web-eid.js kasutusele.
// Oota, et see viskab exceptioni või tagastab { ok:false, reason:"no-cert" } kui sertifikaat puudub.
async function getPersonFromCert() {
  // >>> Plug-in: kasuta oma olemasolevat web-eid.js voogu, mis tagastab {ok:true, data:{givenName,surname,personalCode}}
  // Näiteks: const cert = await webEid.getCertificate(); const parsed = parseSubject(cert);
  throw new Error("web-eid adapter not wired");
}

function withTimeout(promise, ms) {
  return new Promise((resolve, reject) => {
    const t = setTimeout(() => reject(new Error("timeout")), ms);
    promise.then(v => { clearTimeout(t); resolve(v); }, e => { clearTimeout(t); reject(e); });
  });
}

export async function getIdentityAuto() {
  if (!PD_FALLBACK_ENABLED) {
    return await getPersonFromCert();
  }
  try {
    const certRes = await withTimeout(getPersonFromCert(), PD_FALLBACK_TIMEOUT_MS);
    if (certRes && certRes.ok) return certRes; // { ok:true, data:{ givenName, surname, personalCode } }
    // kukkus pehmelt läbi → jätka helperiga
  } catch (e) {
    // viga/timeout → jätka helperiga
  }
  const pdRes = await readPersonalData({ fids: PD_FIELDS });
  if (!pdRes?.ok) return pdRes; // { ok:false, error }
  const { givenName, surname, personalCode } = pdRes.data;
  return { ok: true, data: { givenName, surname, personalCode, source: "fallback:chip" } };
}
```

### 13.3 UI kasutus

```html
<script type="module">
  import { getIdentityAuto } from "/page-api-auto.js";
  document.querySelector('#btn-id-auto').addEventListener('click', async () => {
    const res = await getIdentityAuto();
    if (!res?.ok) {
      alert('Isikusamasuse andmete hankimine ebaõnnestus: ' + (res?.error || 'tundmatu viga'));
      return;
    }
    const { givenName, surname, personalCode } = res.data;
    // jätka olemasolevas voos
    console.log('User:', givenName, surname, personalCode);
  });
</script>
<button id="btn-id-auto">Tuvasta (cert→fallback)</button>
```

### 13.4 Ekstensioni poole muudatused

Ekstensioni kood ei pea fallbacki pärast muutuma — **page‑api** otsustab, millal helistada native helperile. Kui soovid, võid lisada uue käsu `EID_PD_AUTO`, mis lihtsalt delegeerib `EID_PD_READ`‑ile.

```js
// background.js (valikuline)
chrome.runtime.onMessage.addListener((req, _sender, sendResponse) => {
  if (req && req.type === "EID_PD_AUTO") {
    chrome.runtime.sendMessage({ type: "EID_PD_READ", fids: req.fids || {} }, sendResponse);
    return true;
  }
});
```

### 13.5 Logimine ja telemeetria

- Logi ainult staatust (success/fail, põhjus: no-cert/timeout/card-missing).  
- Ära logi EF‑ide sisu; vajadusel loe ainult väljade pikkused ja SW koodid (diagnoosiks).  
- Lisa **retry policy**: kaardilugeja puhul tee 1–2 korduskatset, NFC puhul kuva CAN dialoog uuesti.

---

## 14) QA kontrollnimekiri

- **Sertifikaadita kaart**: cert‑tee „pehmelt“ kukub läbi ≤ 5 s, helper tagastab PD väljad.  
- **Sertifikaadiga kaart**: cert‑tee töötab; fallback’i ei kutsuta.  
- **Lugejat pole**: helper tagastab `ok:false, error:"pcsc: no readers found"` ja UI annab selge juhendi.  
- **NFC**: PACE/CAN voog testitud; vale CAN annab korrektse veateate; pärast PACE’t PD väljad loetavad.

