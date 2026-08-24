package admission

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// DecodeCreateRequest admits the Workbook create wire shape into the Indicator
// owner's command vocabulary. Indicator field policy must remain here rather
// than in Workbook's generic transport boundary.
func DecodeCreateRequest(reader io.Reader) (indicators.CreateCommand, *indicators.IndicatorCreateValidationError) {
	schema, ok := viewschema.Lookup(indicators.ViewSchemaID)
	if !ok {
		return indicators.CreateCommand{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return indicators.CreateCommand{}, invalidMutationPayload("", "request_not_object")
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

	if len(schema.MinimumCreateFieldSets) > 0 && !minimumCreateFieldsSatisfied(schema.MinimumCreateFieldSets, values) {
		return indicators.CreateCommand{}, invalidMutationPayload("payload", "at_least_one_value_required")
	}
	command = createCommandFromValues(command.ClientTxnID, values)
	if err := indicators.ValidateCreateCommand(command); err != nil {
		var validation *indicators.IndicatorCreateValidationError
		if errors.As(err, &validation) {
			return indicators.CreateCommand{}, invalidMutationPayload(validation.Field, validation.ReasonCode)
		}
		return indicators.CreateCommand{}, invalidMutationPayload("payload", "invalid_value")
	}
	return command, nil
}

func minimumCreateFieldsSatisfied(fieldSets [][]string, values map[string]string) bool {
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

func createCommandFromValues(clientTxnID string, values map[string]string) indicators.CreateCommand {
	return indicators.CreateCommand{
		ClientTxnID:     clientTxnID,
		IndicatorType:   values["indicator.indicator_type"],
		ValueKind:       values["indicator.value_kind"],
		DisplayValue:    values["indicator.display_value"],
		NormalizedValue: optionalValue(values, "indicator.normalized_value"),
		DefangedValue:   optionalValue(values, "indicator.defanged_value"),
		HashAlgorithm:   optionalValue(values, "indicator.hash_algorithm"),
		HashValue:       optionalValue(values, "indicator.hash_value"),
		STIXPattern:     optionalValue(values, "indicator.stix_pattern"),
	}
}

func optionalValue(values map[string]string, key string) *string {
	value, present := values[key]
	if !present {
		return nil
	}
	return &value
}

func invalidMutationPayload(field string, reasonCode string) *indicators.IndicatorCreateValidationError {
	return &indicators.IndicatorCreateValidationError{Field: field, ReasonCode: reasonCode}
}
