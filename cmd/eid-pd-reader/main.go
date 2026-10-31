package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"ee.aarepuusepp.eid-pd-reader/native/config"
	"ee.aarepuusepp.eid-pd-reader/native/esteid"
	"ee.aarepuusepp.eid-pd-reader/native/logger"
	"ee.aarepuusepp.eid-pd-reader/native/nfc"
	"ee.aarepuusepp.eid-pd-reader/native/policy"
)

// Version information
const (
	Version = "2.1.9"
	AppName = "eid-pd-reader"
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
	Origin          string            `json:"origin"`
	Command         string            `json:"command"` // New field name for consistency
	Cmd             string            `json:"cmd"`     // Legacy field name
	FIDs            map[string]string `json:"fids"`
	CAN             string            `json:"can"`             // Card Access Number for NFC/PACE
	Mode            string            `json:"mode"`            // "contact" or "nfc"
	Retry           int               `json:"retry"`           // NFC retry count (0 = default 3)
	ContainerBase64 string            `json:"containerBase64"` // CDOC bytes encoded as base64
	PIN             string            `json:"pin"`             // optional; native prompt used when empty
	PKCS11          pkcs11Request     `json:"pkcs11"`
	Certificate     string            `json:"certificate"`
	Challenge       json.RawMessage   `json:"challenge"`
	Lang            string            `json:"lang"`
}

type pkcs11Request struct {
	LibraryPath string `json:"libraryPath"`
	Slot        *int   `json:"slot"`
	Label       string `json:"label"`
	KeyIDHex    string `json:"keyIdHex"`
}

type response struct {
	OK        bool                 `json:"ok"`
	Timestamp string               `json:"timestamp"`
	Duration  string               `json:"duration"`
	Error     string               `json:"error,omitempty"`
	Data      *esteid.PersonalData `json:"data,omitempty"`
}

func makeResponse(ok bool, startTime time.Time, data *esteid.PersonalData, errMsg string) response {
	return response{
		OK:        ok,
		Timestamp: startTime.Format("2006-01-02 15:04:05.00000"),
		Duration:  esteid.FormatDuration(time.Since(startTime)),
		Error:     errMsg,
		Data:      data,
	}
}

func main() {
	dumpFlag := flag.Bool("dump", false, "Dump personal data to stdout as JSON and exit")
	debugFlag := flag.Bool("debug", false, "Debug mode: show card info and try different FIDs")
	verboseFlag := flag.Bool("verbose", false, "Verbose mode: show detailed APDU commands and responses")
	logPath := flag.String("log", "", "Log file path (overrides config file)")
	noLog := flag.Bool("nolog", false, "Disable logging (overrides config file)")
	configPath := flag.String("config", "", "Path to config.yaml file")
	versionFlag := flag.Bool("version", false, "Show version information and exit")
	nfcFlag := flag.Bool("nfc", false, "Use NFC mode with PACE (requires --can)")
	canFlag := flag.String("can", "", "Card Access Number (6 digits) for NFC/PACE")
	canEnv := flag.String("can-env", "", "Environment variable name containing CAN")
	retryFlag := flag.Int("retry", 0, "Total number of attempts for NFC reading (default: 3)")
	allReadersFlag := flag.Bool("all-readers", false, "Read from all available readers and return multiple results")
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("%s version %s\n", AppName, Version)
		fmt.Println("Estonian ID Card Personal Data Reader")
		fmt.Println("https://github.com/aarepuusepp/web-eid-pd")
		return
	}

	// Load configuration
	if *configPath == "" && !*dumpFlag && !*debugFlag {
		if machineDir, dirErr := policy.DataDirectory(); dirErr == nil {
			*configPath = filepath.Join(machineDir, "eid-pd-reader", "config.yaml")
		}
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		cfg = config.Default()
	}

	// Command-line flags override config file
	if *noLog {
		cfg.Logging.Enabled = false
	}
	if *logPath != "" {
		cfg.Logging.File = *logPath
	}
	if *verboseFlag {
		cfg.Advanced.VerboseAPDU = true
	}

	// Initialize logger (unless disabled or in debug mode)
	if cfg.Logging.Enabled && !*debugFlag {
		if err := logger.Init(cfg.Logging.File); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize logger: %v\n", err)
		} else {
			defer logger.Close()
		}
	}

	if *debugFlag {
		// Debug mode: try to identify card structure
		if err := esteid.DebugCardStructure(); err != nil {
			fmt.Fprintf(os.Stderr, "Debug failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *dumpFlag {
		// CLI mode: dump PD and exit
		startTime := time.Now()

		if cfg.Advanced.VerboseAPDU {
			esteid.SetVerbose(true)
		}

		// Handle --all-readers flag
		if *allReadersFlag {
			results, err := esteid.ReadPersonalDataFromAllReaders(cfg.Advanced.CustomFIDs)
			if err != nil {
				logger.LogCardReadError(err.Error())
				resp := makeResponse(false, startTime, nil, err.Error())
				json.NewEncoder(os.Stdout).Encode(resp)
				os.Exit(1)
			}

			type multiResponse struct {
				OK        bool                `json:"ok"`
				Timestamp string              `json:"timestamp"`
				Duration  string              `json:"duration"`
				Results   []esteid.CardResult `json:"results"`
			}
			resp := multiResponse{
				OK:        true,
				Timestamp: startTime.Format("2006-01-02 15:04:05.00000"),
				Duration:  esteid.FormatDuration(time.Since(startTime)),
				Results:   results,
			}
			json.NewEncoder(os.Stdout).Encode(resp)
			return
		}

		var pd *esteid.PersonalData
		var err error

		if *nfcFlag {
			// NFC mode: requires CAN
			can := *canFlag
			if can == "" && *canEnv != "" {
				can = os.Getenv(*canEnv)
			}
			if can == "" {
				resp := makeResponse(false, startTime, nil, "NFC mode requires --can or --can-env")
				json.NewEncoder(os.Stdout).Encode(resp)
				os.Exit(1)
			}

			// Set total attempts (--retry N means try N times total)
			maxAttempts := *retryFlag
			if maxAttempts == 0 {
				maxAttempts = 3 // Default: 3 attempts (NFC can be flaky)
			}

			pd, err = esteid.ReadPersonalDataNFCWithRetry(can, maxAttempts)
		} else {
			// Contact mode
			pd, err = esteid.ReadPersonalData(cfg.Advanced.CustomFIDs)
		}

		if err != nil {
			logger.LogCardReadError(err.Error())
			resp := makeResponse(false, startTime, nil, err.Error())
			json.NewEncoder(os.Stdout).Encode(resp)
			os.Exit(1)
		}
		logger.LogCardRead(pd.PersonalCode, pd.Surname, pd.GivenName, pd.DocumentNr, pd.Source)
		resp := makeResponse(true, startTime, pd, "")
		json.NewEncoder(os.Stdout).Encode(resp)
		return
	}

	// Check if running interactively without any action flags
	// If stdin is a terminal and no flags were provided, show help
	if isStdinTerminal() {
		printUsage()
		return
	}

	// Native Messaging host mode
	logger.Info("Native messaging host started", map[string]interface{}{
		"pid": os.Getpid(),
	})

	reader := bufio.NewReader(os.Stdin)
	for {
		msg, err := readNativeMessage(reader)
		if err != nil {
			logger.Info("Native messaging host stopped", map[string]interface{}{
				"reason": err.Error(),
			})
			return // EOF or fatal
		}
		startTime := time.Now()

		var req request
		if err := json.Unmarshal(msg, &req); err != nil {
			logger.LogNativeMessage("request", "unknown", false, fmt.Sprintf("bad request: %v", err))
			_ = writeNativeMessage(makeResponse(false, startTime, nil, fmt.Sprintf("bad request: %v", err)))
			continue
		}
		// Support both "command" and "cmd" field names
		command := req.Command
		if command == "" {
			command = req.Cmd
		}

		if command == "policy" {
			p, err := policy.Load()
			if err != nil {
				_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": "Invalid machine access policy", "errorCode": "ERR_EIDPD_POLICY_INVALID"})
			} else {
				_ = writeNativeMessage(map[string]interface{}{"ok": true, "data": p, "protocol": 1})
			}
			continue
		}
		if command != "status" && command != "version" && command != "ping" && command != "capabilities" {
			p, err := policy.Load()
			if err != nil {
				_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": "Invalid machine access policy", "errorCode": "ERR_EIDPD_POLICY_INVALID"})
				continue
			}
			action, err := p.Action(req.Origin)
			if err != nil || action == "deny" {
				_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": "Website access denied", "errorCode": "ERR_EIDPD_SITE_DENIED"})
				continue
			}
			// Only the trusted extension can grant 'ask'. Never accept PINs or library paths from a page.
			if req.PIN != "" || req.PKCS11.LibraryPath != "" || req.PKCS11.Slot != nil || req.PKCS11.Label != "" || req.PKCS11.KeyIDHex != "" {
				_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": "Native PIN and machine PKCS11 configuration are required", "errorCode": "ERR_EIDPD_INVALID_REQUEST"})
				continue
			}
		}

		switch command {
		case "ping":
			logger.LogNativeMessage("request", "ping", true, "")
			_ = writeNativeMessage(makeResponse(true, startTime, nil, ""))

		case "version":
			logger.LogNativeMessage("request", "version", true, "")
			versionData := map[string]string{
				"version": Version,
				"name":    AppName,
			}
			_ = writeNativeMessage(map[string]interface{}{
				"ok":        true,
				"timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				"duration":  esteid.FormatDuration(time.Since(startTime)),
				"version":   Version,
				"name":      AppName,
				"data":      versionData,
			})

		case "readPersonalData", "getPersonalData":
			logger.LogNativeMessage("request", command, true, "")
			pd, err := esteid.ReadPersonalData(req.FIDs)
			if err != nil {
				logger.LogCardReadError(err.Error())
				logger.LogNativeMessage("response", command, false, err.Error())
				_ = writeNativeMessage(makeResponse(false, startTime, nil, err.Error()))
				continue
			}
			logger.LogCardRead(pd.PersonalCode, pd.Surname, pd.GivenName, pd.DocumentNr, pd.Source)
			logger.LogNativeMessage("response", command, true, "")
			_ = writeNativeMessage(makeResponse(true, startTime, pd, ""))

		case "readPersonalDataNFC":
			logger.LogNativeMessage("request", command, true, "CAN length: "+fmt.Sprint(len(req.CAN)))
			if req.CAN == "" {
				errMsg := "NFC mode requires CAN (Card Access Number)"
				logger.LogNativeMessage("response", command, false, errMsg)
				_ = writeNativeMessage(makeResponse(false, startTime, nil, errMsg))
				continue
			}
			maxAttempts := req.Retry
			if maxAttempts <= 0 {
				maxAttempts = 3
			}
			pd, err := esteid.ReadPersonalDataNFCWithRetry(req.CAN, maxAttempts)
			if err != nil {
				logger.LogCardReadError(err.Error())
				logger.LogNativeMessage("response", command, false, err.Error())
				_ = writeNativeMessage(makeResponse(false, startTime, nil, err.Error()))
				continue
			}
			logger.LogCardRead(pd.PersonalCode, pd.Surname, pd.GivenName, pd.DocumentNr, pd.Source)
			logger.LogNativeMessage("response", command, true, "")
			_ = writeNativeMessage(makeResponse(true, startTime, pd, ""))

		case "readAllReaders":
			logger.LogNativeMessage("request", command, true, "")
			results, err := esteid.ReadPersonalDataFromAllReaders(req.FIDs)
			if err != nil {
				logger.LogCardReadError(err.Error())
				logger.LogNativeMessage("response", command, false, err.Error())
				_ = writeNativeMessage(makeResponse(false, startTime, nil, err.Error()))
				continue
			}
			logger.LogNativeMessage("response", command, true, fmt.Sprintf("readers: %d", len(results)))
			type multiResponse struct {
				OK        bool                `json:"ok"`
				Timestamp string              `json:"timestamp"`
				Duration  string              `json:"duration"`
				Results   []esteid.CardResult `json:"results"`
			}
			_ = writeNativeMessage(multiResponse{
				OK:        true,
				Timestamp: startTime.Format("2006-01-02 15:04:05.00000"),
				Duration:  esteid.FormatDuration(time.Since(startTime)),
				Results:   results,
			})

		case "status":
			logger.LogNativeMessage("request", "status", true, "")
			_ = writeNativeMessage(map[string]interface{}{
				"ok":        true,
				"timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				"duration":  esteid.FormatDuration(time.Since(startTime)),
				"data": map[string]interface{}{
					"protocol":           1,
					"diagnostics":        esteid.Diagnostics(cfg.PKCS11.LibraryPath),
					"version":            Version,
					"name":               AppName,
					"nativeApp":          Version,
					"nfcCryptoSupported": true,
					"cdocVersions":       []string{"1.0", "1.1"},
				},
			})

		case "get-decrypt-certificate", "getTokenCertificate":
			logger.LogNativeMessage("request", command, true, "")
			if req.Mode != "" && req.Mode != "contact" && req.Mode != "nfc" {
				_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": "invalid card mode"})
				continue
			}
			if req.Mode == "nfc" {
				result, err := esteid.GetDecryptCertificateNFC(req.CAN)
				if err != nil {
					_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": err.Error()})
				} else {
					_ = writeNativeMessage(map[string]interface{}{"ok": true, "data": result})
				}
				continue
			}
			p11Cfg, err := esteid.ResolvePKCS11Config(
				cfg.PKCS11.LibraryPath, cfg.PKCS11.Slot, cfg.PKCS11.Label, cfg.PKCS11.KeyIDHex,
				req.PKCS11.LibraryPath, req.PKCS11.Slot, req.PKCS11.Label, req.PKCS11.KeyIDHex,
				"", false,
			)
			if err != nil {
				_ = writeNativeMessage(map[string]interface{}{
					"ok": false, "error": err.Error(), "timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				})
				continue
			}
			result, err := esteid.GetDecryptCertificate(p11Cfg)
			if err != nil {
				_ = writeNativeMessage(map[string]interface{}{
					"ok": false, "error": err.Error(), "timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				})
				continue
			}
			_ = writeNativeMessage(map[string]interface{}{
				"ok": true, "timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				"duration": esteid.FormatDuration(time.Since(startTime)), "data": result,
			})

		case "unwrap-transport-key", "unwrapTransportKey":
			logger.LogNativeMessage("request", command, true, "")
			if req.Mode != "" && req.Mode != "contact" && req.Mode != "nfc" {
				_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": "invalid card mode"})
				continue
			}
			if req.Mode == "nfc" {
				result, err := esteid.UnwrapTransportKeyNFC(req.CAN, req.Challenge, req.Certificate, req.Origin)
				if err != nil {
					_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": err.Error(), "errorCode": esteid.PINErrorCode(err)})
				} else {
					_ = writeNativeMessage(map[string]interface{}{"ok": true, "data": result})
				}
				continue
			}
			p11Cfg, err := esteid.ResolvePKCS11Config(
				cfg.PKCS11.LibraryPath, cfg.PKCS11.Slot, cfg.PKCS11.Label, cfg.PKCS11.KeyIDHex,
				req.PKCS11.LibraryPath, req.PKCS11.Slot, req.PKCS11.Label, req.PKCS11.KeyIDHex,
				req.PIN, false,
			)
			if err != nil {
				_ = writeNativeMessage(map[string]interface{}{
					"ok": false, "error": err.Error(), "errorCode": esteid.PINErrorCode(err), "timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				})
				continue
			}
			result, err := esteid.UnwrapTransportKey(p11Cfg, req.Challenge, req.Certificate, req.Origin)
			if err != nil {
				_ = writeNativeMessage(map[string]interface{}{
					"ok": false, "error": err.Error(), "errorCode": esteid.PINErrorCode(err), "timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				})
				continue
			}
			_ = writeNativeMessage(map[string]interface{}{
				"ok": true, "timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				"duration": esteid.FormatDuration(time.Since(startTime)), "data": result,
			})

		case "decryptContainer":
			resolved, resolveErr := esteid.ResolvePKCS11Config(cfg.PKCS11.LibraryPath, cfg.PKCS11.Slot, cfg.PKCS11.Label, cfg.PKCS11.KeyIDHex, "", nil, "", "", "", true)
			if resolveErr != nil {
				_ = writeNativeMessage(map[string]interface{}{"ok": false, "error": resolveErr.Error(), "errorCode": esteid.PINErrorCode(resolveErr)})
				continue
			}
			req.PIN = resolved.PIN
			req.PKCS11.LibraryPath = resolved.LibraryPath
			logger.LogNativeMessage("request", command, true, fmt.Sprintf("containerBase64 bytes: %d", len(req.ContainerBase64)))
			slot := cfg.PKCS11.Slot
			if req.PKCS11.Slot != nil {
				slot = *req.PKCS11.Slot
			}
			libraryPath := cfg.PKCS11.LibraryPath
			if req.PKCS11.LibraryPath != "" {
				libraryPath = req.PKCS11.LibraryPath
			}
			label := cfg.PKCS11.Label
			if req.PKCS11.Label != "" {
				label = req.PKCS11.Label
			}
			keyIDHex := cfg.PKCS11.KeyIDHex
			if req.PKCS11.KeyIDHex != "" {
				keyIDHex = req.PKCS11.KeyIDHex
			}

			result, err := esteid.DecryptContainer(esteid.DecryptOptions{
				ContainerBase64: req.ContainerBase64,
				PKCS11Library:   libraryPath,
				PIN:             req.PIN,
				Slot:            slot,
				Label:           label,
				KeyIDHex:        keyIDHex,
			})
			if err != nil {
				logger.LogNativeMessage("response", command, false, err.Error())
				_ = writeNativeMessage(map[string]interface{}{
					"ok":        false,
					"timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
					"duration":  esteid.FormatDuration(time.Since(startTime)),
					"error":     err.Error(),
				})
				continue
			}
			logger.LogNativeMessage("response", command, true, fmt.Sprintf("files: %d", result.FileCount))
			_ = writeNativeMessage(map[string]interface{}{
				"ok":        true,
				"timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				"duration":  esteid.FormatDuration(time.Since(startTime)),
				"data":      result,
			})

		case "capabilities":
			logger.LogNativeMessage("request", "capabilities", true, "")
			caps := map[string]interface{}{
				"nfcSupported":          isNFCSupported(),
				"nfcCryptoSupported":    true,
				"cdocDecryptSupported":  true,
				"clientUnwrapSupported": true,
				"pkcs11RSASupported":    true,
				"pkcs11ECSupported":     true,
				"pkcs11Configured":      cfg.PKCS11.LibraryPath != "" || esteid.DefaultPKCS11LibraryPath() != "",
				"version":               Version,
				"nativeApp":             Version,
				"cdocVersions":          []string{"1.0", "1.1"},
			}
			_ = writeNativeMessage(map[string]interface{}{
				"ok":        true,
				"timestamp": startTime.Format("2006-01-02 15:04:05.00000"),
				"duration":  esteid.FormatDuration(time.Since(startTime)),
				"data":      caps,
			})

		default:
			errMsg := fmt.Sprintf("unknown command: %s", command)
			logger.LogNativeMessage("request", command, false, errMsg)
			_ = writeNativeMessage(makeResponse(false, startTime, nil, errMsg))
		}
	}
}

// isNFCSupported checks if NFC/PACE support is available in this build.
func isNFCSupported() bool {
	return nfc.PACEAvailable_Pure()
}

// isStdinTerminal checks if stdin is a terminal (interactive mode).
func isStdinTerminal() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Check if stdin is a character device (terminal) rather than a pipe or file
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// printUsage prints help information for CLI usage.
func printUsage() {
	fmt.Printf("%s version %s - Estonian ID Card Personal Data Reader\n\n", AppName, Version)
	fmt.Println("Usage:")
	fmt.Println("  eid-pd-reader --dump              Read personal data from ID card (contact mode)")
	fmt.Println("  eid-pd-reader --dump --nfc --can=XXXXXX  Read via NFC with CAN")
	fmt.Println("  eid-pd-reader --dump --all-readers       Read from all available readers")
	fmt.Println("  eid-pd-reader --debug             Debug mode: show card structure")
	fmt.Println("  eid-pd-reader --version           Show version information")
	fmt.Println("")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Native Messaging Mode:")
	fmt.Println("  When run without flags and stdin is not a terminal, the application")
	fmt.Println("  operates as a Chrome/Firefox Native Messaging host.")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  eid-pd-reader --dump")
	fmt.Println("  eid-pd-reader --dump --verbose")
	fmt.Println("  eid-pd-reader --dump --nfc --can=123456")
	fmt.Println("  eid-pd-reader --dump --nfc --can-env=EID_CAN --retry=3")
}

// --- Native messaging framing (4-byte little endian length) ---
func readNativeMessage(r *bufio.Reader) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}
	length := int(uint32(lenBuf[0]) | uint32(lenBuf[1])<<8 | uint32(lenBuf[2])<<16 | uint32(lenBuf[3])<<24)
	if length <= 0 || length > 1<<20 {
		return nil, fmt.Errorf("invalid message length: %d", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeNativeMessage(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	lenBuf := []byte{byte(len(b)), byte(len(b) >> 8), byte(len(b) >> 16), byte(len(b) >> 24)}
	if _, err := os.Stdout.Write(lenBuf); err != nil {
		return err
	}
	_, err = os.Stdout.Write(b)
	return err
}
