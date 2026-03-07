package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"ee.aarepuusepp.eid-pd-reader/native/config"
	"ee.aarepuusepp.eid-pd-reader/native/esteid"
	"ee.aarepuusepp.eid-pd-reader/native/logger"
	"ee.aarepuusepp.eid-pd-reader/native/nfc"
)

// Version information
const (
	Version = "2.0.0"
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
	Command string            `json:"command"` // New field name for consistency
	Cmd     string            `json:"cmd"`     // Legacy field name
	FIDs    map[string]string `json:"fids"`
	CAN     string            `json:"can"`   // Card Access Number for NFC/PACE
	Mode    string            `json:"mode"`  // "contact" or "nfc"
	Retry   int               `json:"retry"` // NFC retry count (0 = default 3)
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

		case "capabilities":
			logger.LogNativeMessage("request", "capabilities", true, "")
			caps := map[string]interface{}{
				"nfcSupported": isNFCSupported(),
				"version":      Version,
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
	if _, err := r.Read(lenBuf); err != nil {
		return nil, err
	}
	length := int(uint32(lenBuf[0]) | uint32(lenBuf[1])<<8 | uint32(lenBuf[2])<<16 | uint32(lenBuf[3])<<24)
	if length <= 0 || length > 1<<20 {
		return nil, fmt.Errorf("invalid message length: %d", length)
	}
	payload := make([]byte, length)
	if _, err := r.Read(payload); err != nil {
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
