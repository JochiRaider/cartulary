package main

import "fmt"

func extensionTransactionProtocol(indexed map[string]map[string]any) (map[string]any, error) {
	protocol := indexed["specification/shared-protocols.json"]
	if protocol["schema_id"] != "cartulary.extension_shared_protocol_contract_set.v5" {
		return nil, fmt.Errorf("unsupported shared protocol contract")
	}
	limits, ok := protocol["transaction"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing shared transaction protocol")
	}
	if limits["canonical_encoding_id"] != "extension_registry_canonical_json_v1" || limits["input_byte_admission"] != "require_exact_canonical_bytes" || limits["prepare_result_schema_id"] != "cartulary.extension_transaction_participant_prepare_result.v2" || limits["validation_result_schema_id"] != "cartulary.extension_transaction_participant_validation_result.v2" || limits["write_result_schema_id"] != "cartulary.extension_transaction_participant_write_result.v2" {
		return nil, fmt.Errorf("unsupported transaction byte contract")
	}
	for _, name := range []string{"minimum_participants", "maximum_participants", "per_participant_input_max_bytes", "aggregate_input_max_bytes", "per_participant_prepare_result_max_bytes", "aggregate_prepare_result_max_bytes", "serialization_keys_per_participant", "aggregate_serialization_keys", "result_max_bytes", "validation_findings"} {
		value, err := positiveJSONInt(limits[name], name)
		if err != nil || value == 0 {
			return nil, fmt.Errorf("invalid shared transaction limit %s", name)
		}
	}
	return limits, nil
}

func extensionBindingTransactionLimits(limits map[string]any) map[string]any {
	return map[string]any{
		"participant_input_bytes":            limits["per_participant_input_max_bytes"],
		"aggregate_input_bytes":              limits["aggregate_input_max_bytes"],
		"serialization_keys_per_participant": limits["serialization_keys_per_participant"],
		"aggregate_serialization_keys":       limits["aggregate_serialization_keys"],
		"result_bytes":                       limits["result_max_bytes"],
		"validation_findings":                limits["validation_findings"],
	}
}

// Schema templates name authored protocol fields; generated schemas contain
// their exact constants. Numeric limits and byte policy share one input.
func projectExtensionTransactionProtocol(value any, limits map[string]any) error {
	switch node := value.(type) {
	case map[string]any:
		for marker, keyword := range map[string]string{"x-cartulary-transaction-value": "const", "x-cartulary-transaction-max-items": "maxItems"} {
			if raw, present := node[marker]; present {
				name, ok := raw.(string)
				value, exists := limits[name]
				_, duplicate := node[keyword]
				if !ok || !exists || duplicate {
					return fmt.Errorf("unresolved transaction protocol projection")
				}
				delete(node, marker)
				node[keyword] = value
			}
		}

		for _, child := range node {
			if err := projectExtensionTransactionProtocol(child, limits); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range node {
			if err := projectExtensionTransactionProtocol(child, limits); err != nil {
				return err
			}
		}
	}
	return nil
}
