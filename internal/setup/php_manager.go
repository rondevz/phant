package setup

import (
	"context"
	"runtime"
	"time"
)

func GetPHPManagerSnapshot(ctx context.Context) PHPManagerSnapshot {
	report := PHPManagerSnapshot{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Supported:   runtime.GOOS == "linux",
		Platform:    runtime.GOOS,
	}

	if runtime.GOOS != "linux" {
		report.Warnings = append(report.Warnings, "PHP manager is currently supported on Linux only.")
		return report
	}

	activeVersion, versions, versionErr := discoverPHPVersions(ctx)
	report.ActiveVersion = activeVersion
	report.Versions = versions
	if versionErr != nil {
		report.LastError = versionErr.Error()
	}

	settings, settingsErr := discoverPHPIniSettings(ctx)
	report.Settings = settings
	if settingsErr != nil {
		report.Warnings = append(report.Warnings, settingsErr.Error())
	}

	extensions, extensionsErr := discoverPHPExtensions(ctx)
	report.Extensions = extensions
	if extensionsErr != nil {
		report.Warnings = append(report.Warnings, extensionsErr.Error())
	}

	return report
}

func InstallPHPVersion(ctx context.Context, version string) PHPActionResult {
	return installPHPVersionLinux(ctx, version)
}

func SwitchPHPVersion(ctx context.Context, version string) PHPActionResult {
	return switchPHPVersionLinux(ctx, version)
}

func UpdatePHPIniSettings(ctx context.Context, request PHPIniSettingsUpdateRequest) PHPActionResult {
	return updatePHPIniSettingsLinux(ctx, request)
}

func SetPHPExtensionState(ctx context.Context, request PHPExtensionToggleRequest) PHPActionResult {
	return setPHPExtensionStateLinux(ctx, request)
}
