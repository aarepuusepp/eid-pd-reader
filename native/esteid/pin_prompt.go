package esteid

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// PromptPIN1 shows a native PIN1 dialog or secure console prompt.
func PromptPIN1(title string) (string, error) {
	if runtime.GOOS == "windows" {
		if pin := promptPINWindows(title); pin != "" {
			return pin, nil
		}
		return "", fmt.Errorf("PIN entry cancelled")
	}
	if pin := promptPINWindows(title); pin != "" {
		return pin, nil
	}
	if pin, err := promptPINPinentry(title); err == nil && pin != "" {
		return pin, nil
	}
	return promptPINConsole(title)
}

func promptPINWindows(title string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	script := fmt.Sprintf(
		`Add-Type -AssemblyName System.Windows.Forms; $form = New-Object System.Windows.Forms.Form; $form.Text = '%s'; $form.Size = New-Object System.Drawing.Size(360,160); $form.StartPosition = 'CenterScreen'; $label = New-Object System.Windows.Forms.Label; $label.Text = 'Enter ID-card PIN1:'; $label.AutoSize = $true; $label.Location = New-Object System.Drawing.Point(10,20); $form.Controls.Add($label); $box = New-Object System.Windows.Forms.TextBox; $box.UseSystemPasswordChar = $true; $box.Location = New-Object System.Drawing.Point(10,50); $box.Size = New-Object System.Drawing.Size(320,20); $form.Controls.Add($box); $ok = New-Object System.Windows.Forms.Button; $ok.Text = 'OK'; $ok.Location = New-Object System.Drawing.Point(250,85); $ok.Add_Click({$form.Tag = $box.Text; $form.Close()}); $form.Controls.Add($ok); $form.Add_Shown({$box.Focus()}); [void]$form.ShowDialog(); if ($form.Tag) { Write-Output $form.Tag }`,
		strings.ReplaceAll(title, "'", "''"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-STA", "-Command", script).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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
