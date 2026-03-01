package dump

import "encoding/json"

const SchemaVersion = 1

type Event struct {
	SchemaVersion int             `json:"schemaVersion"`
	ID            string          `json:"id"`
	Timestamp     string          `json:"timestamp"`
	SourceType    string          `json:"sourceType"`
	ProjectRoot   string          `json:"projectRoot"`
	PHPSAPI       string          `json:"phpSapi"`
	RequestID     *string         `json:"requestId"`
	HTTP          *HTTPMeta       `json:"http,omitempty"`
	Command       *CommandMeta    `json:"command,omitempty"`
	IsDD          bool            `json:"isDd"`
	PayloadFormat string          `json:"payloadFormat"`
	Payload       json.RawMessage `json:"payload"`
	Trace         []TraceFrame    `json:"trace"`
	Host          HostMeta        `json:"host"`
}

type HTTPMeta struct {
	Method     string `json:"method"`
	Scheme     string `json:"scheme"`
	Host       string `json:"host"`
	Path       string `json:"path"`
	Query      string `json:"query,omitempty"`
	StatusCode *int   `json:"statusCode,omitempty"`
	ClientIP   string `json:"clientIp,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
}

type CommandMeta struct {
	Name string   `json:"name"`
	Args []string `json:"args,omitempty"`
	Cwd  string   `json:"cwd,omitempty"`
}

type TraceFrame struct {
	File string `json:"file,omitempty"`
	Line int    `json:"line,omitempty"`
	Func string `json:"func,omitempty"`
}

type HostMeta struct {
	Hostname string `json:"hostname"`
	PID      int    `json:"pid"`
}
