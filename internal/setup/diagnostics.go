package setup

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Diagnostics struct {
	GeneratedAt    string `json:"generatedAt"`
	PHPFound       bool   `json:"phpFound"`
	PHPVersion     string `json:"phpVersion"`
	PHPIniOutput   string `json:"phpIniOutput"`
	ServiceManager string `json:"serviceManager"`
	LastError      string `json:"lastError"`
}

func CollectDiagnostics(ctx context.Context) Diagnostics {
	report := Diagnostics{
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
		ServiceManager: detectServiceManager(ctx),
	}

	phpVersionOutput, err := runCommand(ctx, "php", "-v")
	if err != nil {
		report.LastError = fmt.Sprintf("php -v failed: %v", err)
		return report
	}

	report.PHPFound = true
	report.PHPVersion = firstLine(phpVersionOutput)

	phpIniOutput, iniErr := runCommand(ctx, "php", "--ini")
	if iniErr != nil {
		report.LastError = fmt.Sprintf("php --ini failed: %v", iniErr)
		return report
	}

	report.PHPIniOutput = strings.TrimSpace(phpIniOutput)
	return report
}

func detectServiceManager(ctx context.Context) string {
	if _, err := runCommand(ctx, "systemctl", "--version"); err == nil {
		return "systemd"
	}
	if _, err := runCommand(ctx, "service", "--status-all"); err == nil {
		return "service"
	}
	return "unknown"
}

func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func firstLine(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lines := strings.Split(trimmed, "\n")
	return strings.TrimSpace(lines[0])
}
