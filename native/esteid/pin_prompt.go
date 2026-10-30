package esteid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var ErrPINTimeout = errors.New("PIN entry timed out after 60 seconds")

// PINContext is sent only over stdin to the installed native dialog.
type PINContext struct {
	FileName     string `json:"fileName"`
	Name         string `json:"name"`
	PersonalCode string `json:"personalCode"`
	Origin       string `json:"origin"`
}

func PINErrorCode(err error) string {
	if errors.Is(err, ErrPINTimeout) {
		return "ERR_EIDPD_PIN_TIMEOUT"
	}
	return ""
}

// PromptPIN1 shows a native PIN1 dialog or secure console prompt.
func PromptPIN1(title string) (string, error) {
	return promptPIN1WithContext(title, PINContext{})
}

func promptPIN1WithContext(title string, details PINContext) (string, error) {
	if runtime.GOOS == "windows" {
		return promptPINWindows(details)
	}
	if pin, err := promptPINPinentry(title); err == nil && pin != "" {
		return pin, nil
	}
	return promptPINConsole(title)
}

func promptPINWindows(details PINContext) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot locate EID-PD installation")
	}
	// Resolve only next to the installed reader, never through PATH or the
	// current directory. The installer protects this directory under Program Files.
	helper := filepath.Join(filepath.Dir(executable), "pin-ui", "EidPd.PinPrompt.exe")
	if _, err := os.Stat(helper); err != nil {
		return "", fmt.Errorf("PIN dialog is missing; repair the EID-PD installation")
	}
	// The UI enforces 60 seconds from Loaded using a monotonic clock. This
	// watchdog additionally covers startup or a stalled UI process.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	metadata, err := json.Marshal(details)
	if err != nil {
		return "", fmt.Errorf("cannot prepare PIN dialog")
	}
	cmd := exec.CommandContext(ctx, helper)
	cmd.Stdin = strings.NewReader(string(metadata))
	out, err := cmd.Output()
	defer clear(out)
	if ctx.Err() != nil {
		return "", ErrPINTimeout
	}
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 3 {
			return "", ErrPINTimeout
		}
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 2 {
			return "", fmt.Errorf("PIN entry cancelled")
		}
		return "", fmt.Errorf("PIN dialog failed; repair the EID-PD installation")
	}
	pin := strings.TrimSpace(string(out))
	if pin == "" {
		return "", fmt.Errorf("PIN entry cancelled")
	}
	return pin, nil
}

func promptPINPinentry(title string) (string, error) {
	path, err := exec.LookPath("pinentry")
	if err != nil {
		return "", err
	}
	cmd := exec.Command(path)
	cmd.Stdin = strings.NewReader("SETDESC " + title + "\nSETPROMPT ID-card PIN1:\nGETPIN\n")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "D ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "D ")), nil
		}
	}
	return "", fmt.Errorf("pinentry returned no PIN")
}

func promptPINConsole(title string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s\nEnter ID-card PIN1: ", title)
	var pin string
	_, err := fmt.Fscanln(os.Stdin, &pin)
	if err != nil {
		return "", fmt.Errorf("PIN entry cancelled")
	}
	if strings.TrimSpace(pin) == "" {
		return "", fmt.Errorf("PIN is required")
	}
	return pin, nil
}
