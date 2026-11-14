package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	contractws "github.com/JochiRaider/cartulary/internal/gen/contractws"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

var replayablePayloadSchemas = sync.OnceValues(func() (map[string]*jsonschema.Schema, error) {
	var index struct {
		Definitions map[string]any `json:"$defs"`
		Properties  struct {
			Messages struct {
				Defaults []struct {
					Type    string         `json:"type"`
					Payload map[string]any `json:"payload_schema"`
				} `json:"default"`
			} `json:"messages"`
		} `json:"properties"`
	}
	if err := json.Unmarshal([]byte(contractws.Index["contracts/ws/index.schema.json"].JSON), &index); err != nil {
		return nil, errors.New("embedded collaboration contract is invalid")
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	const base = "https://collaboration.invalid/payload/"
	for _, message := range index.Properties.Messages.Defaults {
		if !IsReplayableMessageType(message.Type) {
			continue
		}
		if message.Payload == nil {
			return nil, errors.New("embedded collaboration payload schema is missing")
		}
		message.Payload["$defs"] = index.Definitions
		if err := compiler.AddResource(base+message.Type, message.Payload); err != nil {
			return nil, fmt.Errorf("compile collaboration payload schema: %w", err)
		}
	}
	schemas := make(map[string]*jsonschema.Schema, 3)
	for _, family := range []string{"record_changed", "job_progress", "extension_resource_changed"} {
		schema, err := compiler.Compile(base + family)
		if err != nil {
			return nil, fmt.Errorf("compile collaboration payload schema: %w", err)
		}
		schemas[family] = schema
	}
	return schemas, nil
})

// Validate the embedded wire projection without loading a file or remote URI.
// Semantic checks below remain independent of structural schema validation.
func validateReplayablePayloadShape(family string, payload json.RawMessage) error {
	if !IsReplayableMessageType(family) {
		return fmt.Errorf("unsupported replayable message family %q", family)
	}
	if err := validateJSONMembers(payload); err != nil {
		return errors.New("replayable payload JSON is invalid")
	}
	schemas, err := replayablePayloadSchemas()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return errors.New("replayable payload JSON is invalid")
	}
	if err := schemas[family].Validate(value); err != nil {
		// Schema diagnostics may contain payload values; keep them out of errors.
		return fmt.Errorf("invalid %s payload shape", family)
	}
	return nil
}

type payloadField struct {
	name   string
	target any
}

// The wire shape has already been checked. Project exact names so additive
// case variants cannot substitute for, overwrite or invalidate known members.
func decodeKnownPayloadFields(raw json.RawMessage, fields ...payloadField) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return errors.New("payload projection requires an object")
	}
	for _, field := range fields {
		if value, present := object[field.name]; present {
			if err := json.Unmarshal(value, field.target); err != nil {
				return fmt.Errorf("invalid payload member %s", field.name)
			}
		}
	}
	return nil
}
