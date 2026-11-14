package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ClientPayload is the known semantic projection of a client message. Unknown
// additive members are ignored, including at nested presence boundaries.
type ClientPayload struct {
	ClientInstanceID  string
	ResumeToken       string
	LastSeenStreamSeq int64
	Presence          PresenceInput
}

func DecodeClientPayload(message Message) (ClientPayload, error) {
	var result ClientPayload
	object, err := clientObject(message.Payload)
	if err != nil {
		return result, err
	}
	switch message.Type {
	case "pong":
		return result, nil
	case "hello", "resume":
		if err := clientMember(object, "client_instance_id", &result.ClientInstanceID); err != nil {
			return result, err
		}
		if strings.TrimSpace(result.ClientInstanceID) == "" {
			return result, fmt.Errorf("client_instance_id is required")
		}
		if message.Type == "resume" {
			if err := clientMember(object, "resume_token", &result.ResumeToken); err != nil {
				return result, err
			}
			if strings.TrimSpace(result.ResumeToken) == "" {
				return result, fmt.Errorf("resume_token is required")
			}
			if err := clientMember(object, "last_seen_stream_seq", &result.LastSeenStreamSeq); err != nil {
				return result, err
			}
			if result.LastSeenStreamSeq < 0 {
				return result, fmt.Errorf("last_seen_stream_seq must be non-negative")
			}
		}
	case "presence_update":
	default:
		return result, fmt.Errorf("unknown client message type")
	}
	presence, err := clientObject(object["presence"])
	if err != nil {
		return result, fmt.Errorf("presence: %w", err)
	}
	if err := clientMember(presence, "mode", &result.Presence.Mode); err != nil {
		return result, err
	}
	for _, field := range []struct {
		name   string
		target **string
	}{{"record_id", &result.Presence.RecordID}, {"field_key", &result.Presence.FieldKey}} {
		if _, present := presence[field.name]; present {
			var value string
			if err := clientMember(presence, field.name, &value); err != nil {
				return result, err
			}
			*field.target = &value
		}
	}
	sheet, err := clientObject(presence["sheet_ref"])
	if err != nil {
		return result, fmt.Errorf("sheet_ref: %w", err)
	}
	var kind string
	if err := clientMember(sheet, "kind", &kind); err != nil {
		return result, err
	}
	result.Presence.SheetRef = map[string]string{"kind": kind}
	keys := []string{"id"}
	if kind == "extension_workspace" {
		if _, present := sheet["id"]; present {
			return result, fmt.Errorf("extension workspace id is forbidden")
		}
		keys = []string{"extension_profile_id", "workspace_key"}
	}
	for _, key := range keys {
		var value string
		if err := clientMember(sheet, key, &value); err != nil {
			return result, err
		}
		result.Presence.SheetRef[key] = value
	}
	return result, ValidatePresenceInput(result.Presence)
}

func clientObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return nil, fmt.Errorf("client payload member must be an object")
	}
	return object, nil
}

func clientMember(object map[string]json.RawMessage, name string, target any) error {
	raw, present := object[name]
	if !present || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("%s is required and cannot be null", name)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("invalid %s: %w", name, err)
	}
	return nil
}
