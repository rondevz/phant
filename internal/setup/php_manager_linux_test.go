package setup

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestParsePHPVersionFromOutput(t *testing.T) {
	t.Parallel()

	got := parsePHPVersionFromOutput("PHP 8.3.4 (cli) (built: Jan 01 2026 00:00:00)\n")
	if got != "8.3" {
		t.Fatalf("parsePHPVersionFromOutput(...) = %q, want %q", got, "8.3")
	}
}

func TestParsePHPVersionsFromPackageList(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"php8.3-cli - command-line interpreter",
		"php8.2-cli - command-line interpreter",
		"php8.3-cli - duplicate",
		"php-cli - default cli",
	}, "\n")

	versions := parsePHPVersionsFromPackageList(input)
	if len(versions) != 3 {
		t.Fatalf("parsePHPVersionsFromPackageList(...) length = %d, want %d", len(versions), 3)
	}
}

func TestUniqueSortedVersions(t *testing.T) {
	t.Parallel()

	input := []string{"8.1", "8.3", "8.2", "8.3"}
	got := uniqueSortedVersions(input)
	want := []string{"8.3", "8.2", "8.1"}

	if len(got) != len(want) {
		t.Fatalf("uniqueSortedVersions(...) length = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("uniqueSortedVersions(...)[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestInstallPHPVersionLinux_RequiresVersionFormat(t *testing.T) {
	originalGOOS := phpManagerGOOS
	phpManagerGOOS = "linux"
	t.Cleanup(func() {
		phpManagerGOOS = originalGOOS
	})

	result := installPHPVersionLinux(context.Background(), "8")
	if result.Error == "" {
		t.Fatalf("installPHPVersionLinux(...) error = empty, want validation error")
	}
}

func TestInstallPHPVersionLinux_ReportsPrivilegeHints(t *testing.T) {
	originalGOOS := phpManagerGOOS
	originalRun := phpManagerRunCommand
	originalLookPath := phpManagerLookPath

	phpManagerGOOS = "linux"
	phpManagerLookPath = func(file string) (string, error) {
		if file == "apt-get" {
			return "/usr/bin/apt-get", nil
		}
		return "", errors.New("missing")
	}
	phpManagerRunCommand = func(_ context.Context, name string, args ...string) (string, error) {
		switch name {
		case "dpkg-query":
			if len(args) > 0 && args[0] == "--version" {
				return "Debian dpkg-query", nil
			}
			if len(args) > 2 && args[0] == "-W" && args[2] == "php8.3-cli" {
				return "", errors.New("dpkg-query: no packages found")
			}
			return "", nil
		case "apt-get":
			return "", errors.New("E: Could not open lock file /var/lib/dpkg/lock-frontend - open (13: Permission denied)")
		default:
			return "", fmt.Errorf("unexpected command: %s", name)
		}
	}

	t.Cleanup(func() {
		phpManagerGOOS = originalGOOS
		phpManagerRunCommand = originalRun
		phpManagerLookPath = originalLookPath
	})

	result := installPHPVersionLinux(context.Background(), "8.3")
	if !result.RequiresPrivilege {
		t.Fatalf("installPHPVersionLinux(...).RequiresPrivilege = false, want true")
	}
	if len(result.SuggestedCommands) == 0 {
		t.Fatalf("installPHPVersionLinux(...).SuggestedCommands = empty, want at least one command")
	}
}

func TestSwitchPHPVersionLinux_RequiresVersionFormat(t *testing.T) {
	originalGOOS := phpManagerGOOS
	phpManagerGOOS = "linux"
	t.Cleanup(func() {
		phpManagerGOOS = originalGOOS
	})

	result := switchPHPVersionLinux(context.Background(), "php8.3")
	if result.Error == "" {
		t.Fatalf("switchPHPVersionLinux(...) error = empty, want validation error")
	}
}

func TestRequiresRootPrivileges(t *testing.T) {
	t.Parallel()

	err := errors.New("Could not open lock file /var/lib/dpkg/lock-frontend - open (13: Permission denied)")
	if !requiresRootPrivileges(err) {
		t.Fatalf("requiresRootPrivileges(...) = false, want true")
	}
}

func TestParsePHPIniSettingsOutput(t *testing.T) {
	t.Parallel()

	output := strings.Join([]string{
		"upload_max_filesize=64M",
		"post_max_size=64M",
		"memory_limit=512M",
		"max_execution_time=120",
	}, "\n")

	settings := parsePHPIniSettingsOutput(output)
	if settings.UploadMaxFilesize != "64M" {
		t.Fatalf("UploadMaxFilesize = %q, want %q", settings.UploadMaxFilesize, "64M")
	}
	if settings.PostMaxSize != "64M" {
		t.Fatalf("PostMaxSize = %q, want %q", settings.PostMaxSize, "64M")
	}
	if settings.MemoryLimit != "512M" {
		t.Fatalf("MemoryLimit = %q, want %q", settings.MemoryLimit, "512M")
	}
	if settings.MaxExecutionTime != "120" {
		t.Fatalf("MaxExecutionTime = %q, want %q", settings.MaxExecutionTime, "120")
	}
}

func TestNormalizeExtensionName(t *testing.T) {
	t.Parallel()

	got := normalizeExtensionName(" Xdebug.ini ")
	if got != "xdebug" {
		t.Fatalf("normalizeExtensionName(...) = %q, want %q", got, "xdebug")
	}
}

func TestBuildManagedPHPSettingsINI_RequiresAtLeastOneValue(t *testing.T) {
	t.Parallel()

	_, err := buildManagedPHPSettingsINI(PHPIniSettingsUpdateRequest{})
	if err == nil {
		t.Fatalf("buildManagedPHPSettingsINI(...) error = nil, want error")
	}
}

func TestBuildManagedPHPSettingsINI_BuildsContent(t *testing.T) {
	t.Parallel()

	content, err := buildManagedPHPSettingsINI(PHPIniSettingsUpdateRequest{
		UploadMaxFilesize: "64M",
		MemoryLimit:       "512M",
	})
	if err != nil {
		t.Fatalf("buildManagedPHPSettingsINI(...) error = %v, want nil", err)
	}
	if !strings.Contains(content, "upload_max_filesize = 64M") {
		t.Fatalf("buildManagedPHPSettingsINI(...) missing upload_max_filesize line")
	}
	if !strings.Contains(content, "memory_limit = 512M") {
		t.Fatalf("buildManagedPHPSettingsINI(...) missing memory_limit line")
	}
}
