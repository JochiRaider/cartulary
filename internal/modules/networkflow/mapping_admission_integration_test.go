package networkflow_test

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMappingCandidatePublicAdmission(t *testing.T) {
	for _, profileID := range []string{"cisco_sna_netflow_csv_v1", "cisco_sna_flow_search_csv_v1"} {
		for _, name := range []string{"alias_only", "alias_first", "alias_last", "unicode_alias", "missing_target", "null_parser", "null_display_name", "missing_transform", "missing_combinability", "policy_alias_first", "policy_alias_last"} {
			t.Run(profileID+"/"+name, func(t *testing.T) {
				assertCiscoMappingRejection(t, profileID, 0, nil, func(mapping map[string]any) json.RawMessage {
					encode := func(value any) json.RawMessage {
						raw, err := json.Marshal(value)
						if err != nil {
							t.Fatal(err)
						}
						return raw
					}
					field := mapping["field_mappings"].([]any)[0].(map[string]any)
					switch name {
					case "alias_only":
						mapping["TARGET_KIND"] = mapping["target_kind"]
						delete(mapping, "target_kind")
					case "alias_first", "alias_last":
						raw := encode(mapping)
						if name == "alias_first" {
							return json.RawMessage(`{"TARGET_KIND":null,` + string(raw[1:]))
						}
						return json.RawMessage(string(raw[:len(raw)-1]) + `,"TARGET_KIND":null}`)
					case "unicode_alias":
						mapping["ſource_profile_id"] = mapping["source_profile_id"]
						delete(mapping, "source_profile_id")
					case "missing_target":
						delete(mapping, "target_kind")
					case "null_parser":
						mapping["parser_profile_id"] = nil
					case "null_display_name":
						mapping["display_name_override"] = nil
					case "missing_transform":
						delete(field, "transform_id")
					case "missing_combinability":
						delete(field, "combinability")
					case "policy_alias_first", "policy_alias_last":
						field["empty_value_policy"] = nil
						pair := `"empty_value_policy":null,"EMPTY_VALUE_POLICY":"profile_default"`
						if name == "policy_alias_first" {
							pair = `"EMPTY_VALUE_POLICY":"profile_default","empty_value_policy":null`
						}
						return bytes.Replace(encode(mapping), []byte(`"empty_value_policy":null`), []byte(pair), 1)
					}
					return encode(mapping)
				})
			})
		}
	}
}
