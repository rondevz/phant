package dump

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrUnsupportedSchemaVersion = errors.New("unsupported schemaVersion")

var requiredEventKeys = []string{
	"schemaVersion",
	"id",
	"timestamp",
	"sourceType",
	"projectRoot",
	"phpSapi",
	"requestId",
	"isDd",
	"payloadFormat",
	"payload",
	"trace",
	"host",
}

func DecodeNDJSONLine(line string) (*Event, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil, nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	if err := validateRequiredKeys(raw); err != nil {
		return nil, err
	}

	var event Event
	if err := json.Unmarshal([]byte(trimmed), &event); err != nil {
		return nil, err
	}

	if err := validateEvent(event); err != nil {
		return nil, err
	}

	return &event, nil
}

func validateRequiredKeys(raw map[string]json.RawMessage) error {
	for _, key := range requiredEventKeys {
		if _, ok := raw[key]; !ok {
			return fmt.Errorf("missing required dump event field: %s", key)
		}
	}

	requestIDRaw := raw["requestId"]
	if string(requestIDRaw) != "null" {
		var requestID string
		if err := json.Unmarshal(requestIDRaw, &requestID); err != nil {
			return errors.New("requestId must be null or string")
		}
	}

	isDDRaw := raw["isDd"]
	var isDD bool
	if err := json.Unmarshal(isDDRaw, &isDD); err != nil {
		return errors.New("isDd must be a boolean")
	}

	traceRaw := raw["trace"]
	if string(traceRaw) == "null" {
		return errors.New("trace must be an array")
	}

	var trace []json.RawMessage
	if err := json.Unmarshal(traceRaw, &trace); err != nil {
		return errors.New("trace must be an array")
	}

	return nil
}

func validateEvent(event Event) error {
	if event.SchemaVersion != SchemaVersion {
		return ErrUnsupportedSchemaVersion
	}

	if event.ID == "" || event.Timestamp == "" || event.SourceType == "" || event.ProjectRoot == "" || event.PHPSAPI == "" || event.PayloadFormat == "" || len(event.Payload) == 0 {
		return errors.New("missing required dump event fields")
	}

	if event.Host.Hostname == "" || event.Host.PID <= 0 {
		return errors.New("invalid host metadata")
	}

	timestamp, err := time.Parse(time.RFC3339Nano, event.Timestamp)
	if err != nil {
		return errors.New("timestamp must be RFC3339Nano")
	}
	if timestamp.UTC().Format(time.RFC3339Nano) != event.Timestamp {
		return errors.New("timestamp must be UTC (Z)")
	}

	switch event.SourceType {
	case "http", "cli", "worker", "cron":
	default:
		return errors.New("sourceType must be one of: http, cli, worker, cron")
	}

	if event.PayloadFormat != "json" {
		return errors.New("payloadFormat must be json for schemaVersion 1")
	}

	if event.SourceType == "http" {
		if event.HTTP == nil {
			return errors.New("http metadata is required when sourceType is http")
		}
		if event.HTTP.Method == "" || event.HTTP.Scheme == "" || event.HTTP.Host == "" || event.HTTP.Path == "" {
			return errors.New("http metadata is missing required fields")
		}
	} else {
		if event.Command == nil {
			return errors.New("command metadata is required when sourceType is cli, worker, or cron")
		}
		if event.Command.Name == "" {
			return errors.New("command metadata is missing required field: name")
		}
	}

	if !json.Valid(event.Payload) {
		return errors.New("payload must be valid JSON")
	}

	return nil
}
