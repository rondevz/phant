package setup

type PHPVersion struct {
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
	Active    bool   `json:"active"`
}

type PHPIniSettings struct {
	UploadMaxFilesize string `json:"uploadMaxFilesize"`
	PostMaxSize       string `json:"postMaxSize"`
	MemoryLimit       string `json:"memoryLimit"`
	MaxExecutionTime  string `json:"maxExecutionTime"`
}

type PHPExtension struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Scope     string `json:"scope"`
	INIPath   string `json:"iniPath"`
	INIExists bool   `json:"iniExists"`
}

type PHPManagerSnapshot struct {
	GeneratedAt   string         `json:"generatedAt"`
	Supported     bool           `json:"supported"`
	Platform      string         `json:"platform"`
	ActiveVersion string         `json:"activeVersion"`
	Versions      []PHPVersion   `json:"versions"`
	Settings      PHPIniSettings `json:"settings"`
	Extensions    []PHPExtension `json:"extensions"`
	Warnings      []string       `json:"warnings"`
	LastError     string         `json:"lastError"`
}

type PHPIniSettingsUpdateRequest struct {
	UploadMaxFilesize string `json:"uploadMaxFilesize"`
	PostMaxSize       string `json:"postMaxSize"`
	MemoryLimit       string `json:"memoryLimit"`
	MaxExecutionTime  string `json:"maxExecutionTime"`
}

type PHPExtensionToggleRequest struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type PHPActionResult struct {
	Success           bool     `json:"success"`
	Supported         bool     `json:"supported"`
	Version           string   `json:"version"`
	Command           string   `json:"command"`
	RequiresPrivilege bool     `json:"requiresPrivilege"`
	SuggestedCommands []string `json:"suggestedCommands"`
	Message           string   `json:"message"`
	Error             string   `json:"error"`
}
