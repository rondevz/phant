package setup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

var (
	phpManagerRunCommand = runCommand
	phpManagerLookPath   = exec.LookPath
	phpManagerGOOS       = runtime.GOOS
)

var phpVersionPattern = regexp.MustCompile(`^\d+\.\d+$`)

func discoverPHPVersions(ctx context.Context) (string, []PHPVersion, error) {
	if phpManagerGOOS != "linux" {
		return "", nil, nil
	}

	versionOutput, err := phpManagerRunCommand(ctx, "php", "-v")
	if err != nil {
		return "", nil, fmt.Errorf("php -v failed: %w", err)
	}

	activeVersion := parsePHPVersionFromOutput(versionOutput)
	installed, installErr := discoverInstalledPHPVersions(ctx)
	if installErr != nil {
		if activeVersion == "" {
			return "", nil, installErr
		}
		return activeVersion, []PHPVersion{{
			Version:   activeVersion,
			Installed: true,
			Active:    true,
		}}, nil
	}

	available, availableErr := discoverAvailablePHPVersions(ctx)
	if availableErr != nil {
		available = nil
	}

	allVersions := make(map[string]struct{}, len(installed)+len(available)+1)
	for _, version := range installed {
		allVersions[version] = struct{}{}
	}
	for _, version := range available {
		allVersions[version] = struct{}{}
	}
	if activeVersion != "" {
		allVersions[activeVersion] = struct{}{}
	}

	if len(allVersions) == 0 {
		return activeVersion, nil, nil
	}

	installedSet := make(map[string]struct{}, len(installed))
	for _, version := range installed {
		installedSet[version] = struct{}{}
	}

	versions := make([]string, 0, len(allVersions))
	for version := range allVersions {
		versions = append(versions, version)
	}
	sort.Slice(versions, func(i, j int) bool {
		return comparePHPVersions(versions[i], versions[j]) > 0
	})

	result := make([]PHPVersion, 0, len(versions))
	for _, version := range versions {
		_, isInstalled := installedSet[version]
		result = append(result, PHPVersion{
			Version:   version,
			Installed: isInstalled,
			Active:    version == activeVersion,
		})
	}

	return activeVersion, result, nil
}

func discoverPHPIniSettings(ctx context.Context) (PHPIniSettings, error) {
	output, err := phpManagerRunCommand(ctx, "php", "-r", phpSettingsReadScript())
	if err != nil {
		return PHPIniSettings{}, fmt.Errorf("failed to read php.ini settings: %w", err)
	}

	return parsePHPIniSettingsOutput(output), nil
}

func discoverPHPExtensions(ctx context.Context) ([]PHPExtension, error) {
	enabledOutput, err := phpManagerRunCommand(ctx, "php", "-m")
	if err != nil {
		return nil, fmt.Errorf("php -m failed: %w", err)
	}

	enabledSet := parsePHPExtensionsOutput(enabledOutput)
	availableMap, discoverErr := discoverAvailableExtensionINIFiles()
	if discoverErr != nil {
		return nil, discoverErr
	}

	namesSet := make(map[string]struct{}, len(enabledSet)+len(availableMap))
	for name := range enabledSet {
		namesSet[name] = struct{}{}
	}
	for name := range availableMap {
		namesSet[name] = struct{}{}
	}

	names := make([]string, 0, len(namesSet))
	for name := range namesSet {
		names = append(names, name)
	}
	sort.Strings(names)

	extensions := make([]PHPExtension, 0, len(names))
	for _, name := range names {
		iniPath, exists := availableMap[name]
		_, enabled := enabledSet[name]
		extensions = append(extensions, PHPExtension{
			Name:      name,
			Enabled:   enabled,
			Scope:     "linux",
			INIPath:   iniPath,
			INIExists: exists,
		})
	}

	return extensions, nil
}

func installPHPVersionLinux(ctx context.Context, version string) PHPActionResult {
	if phpManagerGOOS != "linux" {
		return unsupportedPHPActionResult("install", version)
	}

	requested := strings.TrimSpace(version)
	if !isValidPHPVersion(requested) {
		return PHPActionResult{
			Supported: true,
			Version:   requested,
			Error:     "version must use major.minor format, for example 8.3",
		}
	}

	if !isAptBasedLinux(ctx) {
		return PHPActionResult{
			Supported: false,
			Version:   requested,
			Message:   "automatic install currently supports apt-based Linux distributions only",
		}
	}

	alreadyInstalled, checkErr := isPHPVersionInstalled(ctx, requested)
	if checkErr == nil && alreadyInstalled {
		return PHPActionResult{
			Success:   true,
			Supported: true,
			Version:   requested,
			Message:   fmt.Sprintf("PHP %s is already installed", requested),
		}
	}

	packages := []string{
		fmt.Sprintf("php%s", requested),
		fmt.Sprintf("php%s-cli", requested),
		fmt.Sprintf("php%s-fpm", requested),
		fmt.Sprintf("php%s-common", requested),
	}
	commandArgs := append([]string{"install", "-y"}, packages...)
	commandText := "apt-get " + strings.Join(commandArgs, " ")

	_, err := phpManagerRunCommand(ctx, "apt-get", commandArgs...)
	if err != nil {
		result := PHPActionResult{
			Supported: true,
			Version:   requested,
			Command:   commandText,
			Error:     fmt.Sprintf("PHP install failed: %v", err),
		}
		if requiresRootPrivileges(err) {
			result.RequiresPrivilege = true
			result.SuggestedCommands = []string{
				"sudo apt-get update",
				"sudo " + commandText,
			}
		}
		return result
	}

	return PHPActionResult{
		Success:   true,
		Supported: true,
		Version:   requested,
		Command:   commandText,
		Message:   fmt.Sprintf("PHP %s installed successfully", requested),
	}
}

func switchPHPVersionLinux(ctx context.Context, version string) PHPActionResult {
	if phpManagerGOOS != "linux" {
		return unsupportedPHPActionResult("switch", version)
	}

	requested := strings.TrimSpace(version)
	if !isValidPHPVersion(requested) {
		return PHPActionResult{
			Supported: true,
			Version:   requested,
			Error:     "version must use major.minor format, for example 8.3",
		}
	}

	binaryPath := filepath.Join("/usr/bin", "php"+requested)
	if _, statErr := os.Stat(binaryPath); statErr != nil {
		return PHPActionResult{
			Supported: true,
			Version:   requested,
			Error:     fmt.Sprintf("PHP %s binary was not found at %s", requested, binaryPath),
		}
	}

	args := []string{"--set", "php", binaryPath}
	commandText := "update-alternatives " + strings.Join(args, " ")
	_, err := phpManagerRunCommand(ctx, "update-alternatives", args...)
	if err != nil {
		result := PHPActionResult{
			Supported: true,
			Version:   requested,
			Command:   commandText,
			Error:     fmt.Sprintf("failed to switch CLI PHP: %v", err),
		}
		if requiresRootPrivileges(err) {
			result.RequiresPrivilege = true
			result.SuggestedCommands = []string{"sudo " + commandText}
		}
		return result
	}

	if _, valetErr := phpManagerLookPath("valet"); valetErr == nil {
		valetArgs := []string{"use", "php@" + requested}
		valetCommand := "valet " + strings.Join(valetArgs, " ")
		if _, runErr := phpManagerRunCommand(ctx, "valet", valetArgs...); runErr != nil {
			return PHPActionResult{
				Supported: true,
				Version:   requested,
				Command:   valetCommand,
				Error:     fmt.Sprintf("CLI PHP switched, but Valet switch failed: %v", runErr),
				Message:   "PHP CLI switched, but Valet failed to update",
			}
		}
	}

	return PHPActionResult{
		Success:   true,
		Supported: true,
		Version:   requested,
		Command:   commandText,
		Message:   fmt.Sprintf("PHP %s is now active", requested),
	}
}

func updatePHPIniSettingsLinux(ctx context.Context, request PHPIniSettingsUpdateRequest) PHPActionResult {
	if phpManagerGOOS != "linux" {
		return unsupportedPHPActionResult("update-settings", "")
	}

	content, contentErr := buildManagedPHPSettingsINI(request)
	if contentErr != nil {
		return PHPActionResult{
			Supported: true,
			Error:     contentErr.Error(),
		}
	}

	cliINIOutput, err := phpManagerRunCommand(ctx, "php", "--ini")
	if err != nil {
		return PHPActionResult{
			Supported: true,
			Error:     fmt.Sprintf("php --ini failed: %v", err),
		}
	}

	cliConfDPath := parseAdditionalINIPath(cliINIOutput)
	if cliConfDPath == "" || strings.EqualFold(cliConfDPath, "(none)") {
		return PHPActionResult{
			Supported: true,
			Error:     "unable to detect CLI conf.d directory",
		}
	}

	type settingsTarget struct {
		path        string
		serviceName string
	}

	targets := []settingsTarget{{
		path: filepath.Join(cliConfDPath, "99-phant-settings.ini"),
	}}

	services, discoverErr := discoverFPMServices(ctx)
	if discoverErr == nil {
		for _, service := range services {
			targets = append(targets, settingsTarget{
				path:        filepath.Join(service.ConfDPath, "99-phant-settings.ini"),
				serviceName: service.ServiceName,
			})
		}
	}

	hasWriteFailure := false
	requiresPrivilege := false
	suggested := make([]string, 0, len(targets))

	for _, target := range targets {
		if writeErr := writeHookINI(ctx, target.path, content); writeErr != nil {
			hasWriteFailure = true
			if isPermissionError(writeErr) || strings.Contains(strings.ToLower(writeErr.Error()), "pkexec") {
				requiresPrivilege = true
				suggested = append(suggested, buildLinuxWriteManagedINICommand(target.path, content))
			}
		}
	}

	if hasWriteFailure {
		return PHPActionResult{
			Supported:         true,
			RequiresPrivilege: requiresPrivilege,
			SuggestedCommands: uniqueStrings(suggested),
			Message:           "one or more php.ini targets failed to update",
			Error:             "failed to apply PHP settings to all targets",
		}
	}

	for _, service := range services {
		if restartErr := restartFPMService(ctx, service.ServiceName); restartErr != nil {
			return PHPActionResult{
				Supported:         true,
				RequiresPrivilege: true,
				SuggestedCommands: []string{service.RestartCommand},
				Error:             fmt.Sprintf("settings updated, but failed to restart %s: %v", service.ServiceName, restartErr),
				Message:           "settings updated, but one or more PHP-FPM services require manual restart",
			}
		}
	}

	return PHPActionResult{
		Success:   true,
		Supported: true,
		Message:   "PHP settings updated for CLI and discovered FPM services",
	}
}

func setPHPExtensionStateLinux(ctx context.Context, request PHPExtensionToggleRequest) PHPActionResult {
	if phpManagerGOOS != "linux" {
		return unsupportedPHPActionResult("toggle-extension", request.Name)
	}

	extensionName := normalizeExtensionName(request.Name)
	if extensionName == "" {
		return PHPActionResult{Supported: true, Error: "extension name is required"}
	}

	commandName := "phpenmod"
	verb := "enabled"
	if !request.Enabled {
		commandName = "phpdismod"
		verb = "disabled"
	}

	if _, err := phpManagerLookPath(commandName); err != nil {
		return PHPActionResult{
			Supported: false,
			Message:   fmt.Sprintf("%s is not available on this Linux distribution", commandName),
		}
	}

	versions, versionErr := discoverInstalledPHPVersions(ctx)
	if versionErr != nil {
		versions = []string{}
	}
	if len(versions) == 0 {
		activeVersion, _, discoverErr := discoverPHPVersions(ctx)
		if discoverErr == nil && activeVersion != "" {
			versions = []string{activeVersion}
		}
	}

	commands := make([]string, 0, len(versions))
	for _, version := range versions {
		args := []string{"-v", version, "-s", "ALL", extensionName}
		commands = append(commands, commandName+" "+strings.Join(args, " "))
		if _, err := phpManagerRunCommand(ctx, commandName, args...); err != nil {
			result := PHPActionResult{
				Supported: true,
				Command:   commandName + " " + strings.Join(args, " "),
				Error:     fmt.Sprintf("failed to update extension %s for PHP %s: %v", extensionName, version, err),
			}
			if requiresRootPrivileges(err) {
				result.RequiresPrivilege = true
				result.SuggestedCommands = []string{"sudo " + commandName + " " + strings.Join(args, " ")}
			}
			return result
		}
	}

	services, discoverErr := discoverFPMServices(ctx)
	if discoverErr == nil {
		for _, service := range services {
			if restartErr := restartFPMService(ctx, service.ServiceName); restartErr != nil {
				return PHPActionResult{
					Supported:         true,
					RequiresPrivilege: true,
					SuggestedCommands: []string{service.RestartCommand},
					Error:             fmt.Sprintf("extension updated, but failed to restart %s: %v", service.ServiceName, restartErr),
					Message:           "extension updated, but one or more PHP-FPM services require manual restart",
				}
			}
		}
	}

	return PHPActionResult{
		Success:   true,
		Supported: true,
		Command:   strings.Join(commands, " && "),
		Message:   fmt.Sprintf("extension %s %s successfully", extensionName, verb),
	}
}

func unsupportedPHPActionResult(action string, version string) PHPActionResult {
	return PHPActionResult{
		Supported: false,
		Version:   version,
		Message:   fmt.Sprintf("PHP manager action %q is currently supported on Linux only.", action),
	}
}

func parsePHPVersionFromOutput(output string) string {
	line := firstLine(output)
	if line == "" {
		return ""
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ""
	}

	version := strings.TrimSpace(fields[1])
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return version
	}

	return parts[0] + "." + parts[1]
}

func discoverInstalledPHPVersions(ctx context.Context) ([]string, error) {
	output, err := phpManagerRunCommand(ctx, "dpkg-query", "-W", "-f=${Package}\\n")
	if err != nil {
		return nil, fmt.Errorf("dpkg-query failed: %w", err)
	}

	versions := parsePHPVersionsFromPackageList(output)
	return uniqueSortedVersions(versions), nil
}

func discoverAvailablePHPVersions(ctx context.Context) ([]string, error) {
	output, err := phpManagerRunCommand(ctx, "apt-cache", "search", "--names-only", "^php[0-9]\\.[0-9]-cli$")
	if err != nil {
		return nil, fmt.Errorf("apt-cache search failed: %w", err)
	}

	versions := parsePHPVersionsFromPackageList(output)
	return uniqueSortedVersions(versions), nil
}

func parsePHPVersionsFromPackageList(output string) []string {
	versions := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		name := trimmed
		if idx := strings.Index(trimmed, " "); idx > 0 {
			name = trimmed[:idx]
		}

		if !strings.HasPrefix(name, "php") || !strings.HasSuffix(name, "-cli") {
			continue
		}

		version := strings.TrimSuffix(strings.TrimPrefix(name, "php"), "-cli")
		if isValidPHPVersion(version) {
			versions = append(versions, version)
		}
	}

	return versions
}

func uniqueSortedVersions(versions []string) []string {
	seen := make(map[string]struct{}, len(versions))
	unique := make([]string, 0, len(versions))
	for _, version := range versions {
		if _, ok := seen[version]; ok {
			continue
		}
		seen[version] = struct{}{}
		unique = append(unique, version)
	}

	sort.Slice(unique, func(i, j int) bool {
		return comparePHPVersions(unique[i], unique[j]) > 0
	})

	return unique
}

func comparePHPVersions(left string, right string) int {
	leftMajor, leftMinor := parsePHPVersionParts(left)
	rightMajor, rightMinor := parsePHPVersionParts(right)

	if leftMajor != rightMajor {
		if leftMajor > rightMajor {
			return 1
		}
		return -1
	}

	if leftMinor != rightMinor {
		if leftMinor > rightMinor {
			return 1
		}
		return -1
	}

	return 0
}

func parsePHPVersionParts(version string) (int, int) {
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return 0, 0
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		major = 0
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		minor = 0
	}

	return major, minor
}

func isValidPHPVersion(version string) bool {
	return phpVersionPattern.MatchString(version)
}

func isAptBasedLinux(ctx context.Context) bool {
	if _, err := phpManagerLookPath("apt-get"); err != nil {
		return false
	}

	if _, err := phpManagerRunCommand(ctx, "dpkg-query", "--version"); err != nil {
		return false
	}

	return true
}

func isPHPVersionInstalled(ctx context.Context, version string) (bool, error) {
	packageName := fmt.Sprintf("php%s-cli", version)
	_, err := phpManagerRunCommand(ctx, "dpkg-query", "-W", "-f=${Status}", packageName)
	if err != nil {
		return false, err
	}

	return true, nil
}

func requiresRootPrivileges(err error) bool {
	if err == nil {
		return false
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "permission denied") ||
		strings.Contains(lower, "are you root") ||
		strings.Contains(lower, "could not open lock file") ||
		strings.Contains(lower, "superuser")
}

func phpSettingsReadScript() string {
	return `echo "upload_max_filesize=".ini_get("upload_max_filesize").PHP_EOL;` +
		`echo "post_max_size=".ini_get("post_max_size").PHP_EOL;` +
		`echo "memory_limit=".ini_get("memory_limit").PHP_EOL;` +
		`echo "max_execution_time=".ini_get("max_execution_time").PHP_EOL;`
}

func parsePHPIniSettingsOutput(output string) PHPIniSettings {
	settings := PHPIniSettings{}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "upload_max_filesize":
			settings.UploadMaxFilesize = value
		case "post_max_size":
			settings.PostMaxSize = value
		case "memory_limit":
			settings.MemoryLimit = value
		case "max_execution_time":
			settings.MaxExecutionTime = value
		}
	}

	return settings
}

func parsePHPExtensionsOutput(output string) map[string]struct{} {
	enabled := make(map[string]struct{})
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			continue
		}
		name := normalizeExtensionName(trimmed)
		if name == "" {
			continue
		}
		enabled[name] = struct{}{}
	}

	return enabled
}

func discoverAvailableExtensionINIFiles() (map[string]string, error) {
	files, err := filepath.Glob("/etc/php/*/mods-available/*.ini")
	if err != nil {
		return nil, fmt.Errorf("failed to discover extension ini files: %w", err)
	}

	available := make(map[string]string, len(files))
	for _, file := range files {
		name := normalizeExtensionName(strings.TrimSuffix(filepath.Base(file), ".ini"))
		if name == "" {
			continue
		}
		if _, exists := available[name]; exists {
			continue
		}
		available[name] = file
	}

	return available, nil
}

func normalizeExtensionName(name string) string {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	trimmed = strings.TrimSuffix(trimmed, ".ini")
	return trimmed
}

func buildManagedPHPSettingsINI(request PHPIniSettingsUpdateRequest) (string, error) {
	settings := map[string]string{
		"upload_max_filesize": strings.TrimSpace(request.UploadMaxFilesize),
		"post_max_size":       strings.TrimSpace(request.PostMaxSize),
		"memory_limit":        strings.TrimSpace(request.MemoryLimit),
		"max_execution_time":  strings.TrimSpace(request.MaxExecutionTime),
	}

	if settings["upload_max_filesize"] == "" && settings["post_max_size"] == "" && settings["memory_limit"] == "" && settings["max_execution_time"] == "" {
		return "", fmt.Errorf("at least one php.ini setting is required")
	}

	orderedKeys := []string{"upload_max_filesize", "post_max_size", "memory_limit", "max_execution_time"}
	lines := []string{"; Managed by Phant"}
	for _, key := range orderedKeys {
		if settings[key] == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s = %s", key, settings[key]))
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func buildLinuxWriteManagedINICommand(targetPath string, content string) string {
	escapedDir := shellSingleQuote(filepath.Dir(targetPath))
	escapedTarget := shellSingleQuote(targetPath)
	escapedContent := shellSingleQuote(content)

	return fmt.Sprintf("sudo mkdir -p %s && printf '%%s' %s | sudo tee %s > /dev/null", escapedDir, escapedContent, escapedTarget)
}
