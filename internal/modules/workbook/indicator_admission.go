package workbook

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func decodeIndicatorCreate(reader io.Reader) (indicators.CreateCommand, *httpapi.APIError) {
	schema, ok := viewschema.Lookup(indicators.ViewSchemaID)
	if !ok {
		return indicators.CreateCommand{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return indicators.CreateCommand{}, apiErr
	}

	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, allowedKey := allowed[key]; !allowedKey {
			return indicators.CreateCommand{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var command indicators.CreateCommand
	if value, present := raw["client_txn_id"]; !present {
		return indicators.CreateCommand{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &command.ClientTxnID); err != nil || strings.TrimSpace(command.ClientTxnID) == "" {
		return indicators.CreateCommand{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	values := make(map[string]string)
	for fieldKey, field := range schema.Fields() {
		value, present := raw[fieldKey]
		if !present {
			continue
		}
		if !field.Writable && !field.CreateWritable {
			return indicators.CreateCommand{}, invalidMutationPayload(fieldKey, "readonly_field")
		}
		if string(value) == "null" {
			if field.Clearable {
				continue
			}
			return indicators.CreateCommand{}, invalidMutationPayload(fieldKey, "field_not_nullable")
		}

		var rawValue string
		if err := json.Unmarshal(value, &rawValue); err != nil {
			return indicators.CreateCommand{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		normalized, ok := fieldnorm.NormalizeLine(rawValue)
		if !ok {
			return indicators.CreateCommand{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		values[fieldKey] = normalized
	}

	if len(schema.MinimumCreateFieldSets) > 0 && !indicatorCreateMinimumSatisfied(schema.MinimumCreateFieldSets, values) {
		return indicators.CreateCommand{}, invalidMutationPayload("payload", "at_least_one_value_required")
	}
	command = indicatorCreateCommandFromValues(command.ClientTxnID, values)
	if err := indicators.ValidateCreateCommand(command); err != nil {
		var validation *indicators.IndicatorCreateValidationError
		if errors.As(err, &validation) {
			return indicators.CreateCommand{}, invalidMutationPayload(validation.Field, validation.ReasonCode)
		}
		return indicators.CreateCommand{}, invalidMutationPayload("payload", "invalid_value")
	}
	return command, nil
}

func indicatorCreateMinimumSatisfied(fieldSets [][]string, values map[string]string) bool {
	for _, fieldSet := range fieldSets {
		setSatisfied := true
		for _, fieldKey := range fieldSet {
			if strings.TrimSpace(values[fieldKey]) == "" {
				setSatisfied = false
				break
			}
		}
		if setSatisfied {
			return true
		}
	}
	return false
}

func indicatorCreateCommandHash(command indicators.CreateCommand) []byte {
	payload := map[string]any{
		"view_schema_id":           indicators.ViewSchemaID,
		"client_txn_id":            command.ClientTxnID,
		"indicator.indicator_type": command.IndicatorType,
		"indicator.value_kind":     command.ValueKind,
		"indicator.display_value":  command.DisplayValue,
	}
	for key, value := range map[string]*string{
		"indicator.normalized_value": command.NormalizedValue,
		"indicator.defanged_value":   command.DefangedValue,
		"indicator.hash_algorithm":   command.HashAlgorithm,
		"indicator.hash_value":       command.HashValue,
		"indicator.stix_pattern":     command.STIXPattern,
	} {
		if value != nil {
			payload[key] = *value
		}
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func indicatorCreateCommandFromValues(clientTxnID string, values map[string]string) indicators.CreateCommand {
	return indicators.CreateCommand{
		ClientTxnID:     clientTxnID,
		IndicatorType:   values["indicator.indicator_type"],
		ValueKind:       values["indicator.value_kind"],
		DisplayValue:    values["indicator.display_value"],
		NormalizedValue: optionalIndicatorValue(values, "indicator.normalized_value"),
		DefangedValue:   optionalIndicatorValue(values, "indicator.defanged_value"),
		HashAlgorithm:   optionalIndicatorValue(values, "indicator.hash_algorithm"),
		HashValue:       optionalIndicatorValue(values, "indicator.hash_value"),
		STIXPattern:     optionalIndicatorValue(values, "indicator.stix_pattern"),
	}
}

func optionalIndicatorValue(values map[string]string, key string) *string {
	value, present := values[key]
	if !present {
		return nil
	}
	return &value
}
