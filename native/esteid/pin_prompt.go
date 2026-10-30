package esteid

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// PromptPIN1 shows a native PIN1 dialog or secure console prompt.
func PromptPIN1(title string) (string, error) {
	if runtime.GOOS == "windows" {
		return promptPINWindows()
	}
	if pin, err := promptPINPinentry(title); err == nil && pin != "" {
		return pin, nil
	}
	return promptPINConsole(title)
}

func promptPINWindows() (string, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, helper).Output()
	defer clear(out)
	if ctx.Err() != nil {
		return "", fmt.Errorf("PIN entry timed out")
	}
	if err != nil {
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
