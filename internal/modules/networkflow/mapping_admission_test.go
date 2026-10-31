package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func mappingTestJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// Rewrite only the selected object, preserving raw member order inside it.
// A map round-trip of the mutated object would erase duplicate/escape evidence.
func rewriteMappingObject(t *testing.T, raw json.RawMessage, path []string, change func(map[string]json.RawMessage) json.RawMessage) json.RawMessage {
	t.Helper()
	if len(path) == 0 {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil {
			t.Fatal(err)
		}
		return change(object)
	}
	if raw[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			t.Fatal(err)
		}
		index, err := strconv.Atoi(path[0])
		if err != nil || index < 0 || index >= len(items) {
			t.Fatalf("invalid test path %v", path)
		}
		items[index] = rewriteMappingObject(t, items[index], path[1:], change)
		return mappingTestJSON(t, items)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	object[path[0]] = rewriteMappingObject(t, object[path[0]], path[1:], change)
	return mappingTestJSON(t, object)
}

func mappingClosedObjectMutations(t *testing.T, baseline json.RawMessage, candidate bool) []approvedVariantMutation {
	t.Helper()
	paths := [][]string{nil, {"timestamp_profile"}}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(baseline, &root); err != nil {
		t.Fatal(err)
	}
	var rows []fieldMapping
	if err := json.Unmarshal(root["field_mappings"], &rows); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for i, row := range rows {
		if !seen[row.MappingKind] {
			paths = append(paths, []string{"field_mappings", strconv.Itoa(i)})
			seen[row.MappingKind] = true
		}
	}
	if !candidate {
		paths = append(paths, []string{"source_columns", "0"}, []string{"source_columns", "0", "sample_values", "0"})
	}
	var mutations []approvedVariantMutation
	for _, path := range paths {
		var original map[string]json.RawMessage
		rewriteMappingObject(t, baseline, path, func(object map[string]json.RawMessage) json.RawMessage {
			original = object
			return mappingTestJSON(t, object)
		})
		keys := make([]string, 0, len(original))
		for key := range original {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		add := func(label, reason string, change func(map[string]json.RawMessage) json.RawMessage) {
			mutations = append(mutations, approvedVariantMutation{strings.Join(path, "/") + "/" + label, reason, rewriteMappingObject(t, baseline, path, change)})
		}
		for _, value := range []string{"null", "false", "[]", `""`} {
			add("object/"+value, "variant_member_conflict", func(map[string]json.RawMessage) json.RawMessage {
				return json.RawMessage(value)
			})
		}
		for _, key := range keys {
			optional := len(path) == 0 && key == "display_name_override" || candidate && (len(path) == 0 && (key == "parser_profile_id" || key == "unknown_column_policy") || len(path) == 1 && key == "precision" || len(path) == 2 && key == "empty_value_policy")
			if !optional {
				add(key+"/missing", "variant_member_conflict", func(object map[string]json.RawMessage) json.RawMessage {
					delete(object, key)
					return mappingTestJSON(t, object)
				})
			}
			for _, value := range []string{"null", "{}", "false"} {
				if value == "null" && (key == "timezone" || key == "timezone_ruleset_id" || key == "safe_sample") {
					continue
				}
				reason := "variant_member_conflict"
				if candidate && key == "empty_value_policy" {
					reason = "invalid_empty_value_policy"
				}
				add(key+"/"+value, reason, func(object map[string]json.RawMessage) json.RawMessage {
					object[key] = json.RawMessage(value)
					return mappingTestJSON(t, object)
				})
			}
			aliases := []string{strings.ToUpper(key)}
			if strings.Contains(key, "s") {
				aliases = append(aliases, strings.Replace(key, "s", "ſ", 1))
			}
			if strings.Contains(key, "k") {
				aliases = append(aliases, strings.Replace(key, "k", "K", 1))
			}
			for _, alias := range aliases {
				for _, form := range []string{"alone", "first", "last"} {
					add(key+"/"+alias+"/"+form, "variant_member_conflict", func(object map[string]json.RawMessage) json.RawMessage {
						value := object[key]
						if form == "alone" {
							delete(object, key)
						} else {
							// Deliberately disagree with the canonical value.
							value = json.RawMessage(`null`)
						}
						pair := strconv.Quote(alias) + ":" + string(value)
						rest := string(mappingTestJSON(t, object))
						if form == "first" {
							return json.RawMessage("{" + pair + "," + rest[1:])
						}
						return json.RawMessage(rest[:len(rest)-1] + "," + pair + "}")
					})
				}
			}
			add(key+"/escaped_duplicate", "variant_member_conflict", func(object map[string]json.RawMessage) json.RawMessage {
				rest := string(mappingTestJSON(t, object))
				escaped := fmt.Sprintf(`"\u%04x%s"`, key[0], key[1:])
				return json.RawMessage(rest[:len(rest)-1] + "," + escaped + ":" + string(object[key]) + "}")
			})
		}
		add("unknown", "variant_member_conflict", func(object map[string]json.RawMessage) json.RawMessage {
			object["unknown_member"] = json.RawMessage(`null`)
			return mappingTestJSON(t, object)
		})
		if len(path) == 2 && path[0] == "field_mappings" {
			for _, member := range []string{"field_key", "source_column_ordinal", "transform_id", "empty_value_policy", "combinability", "ignore_reason", "derivation_id"} {
				if _, allowed := original[member]; !allowed {
					add("foreign/"+member, "variant_member_conflict", func(object map[string]json.RawMessage) json.RawMessage {
						object[member] = json.RawMessage(`null`)
						return mappingTestJSON(t, object)
					})
				}
			}
		}
	}
	for label, raw := range map[string]json.RawMessage{
		"trailing":        append(bytes.Clone(baseline), []byte(` {}`)...),
		"unpaired_escape": bytes.Replace(baseline, []byte(`"target_kind"`), []byte(`"\ud800"`), 1),
		"invalid_utf8":    bytes.Replace(baseline, []byte(`"target_kind"`), []byte{'"', 0xff, '"'}, 1),
	} {
		mutations = append(mutations, approvedVariantMutation{label, "variant_member_conflict", raw})
	}
	return mutations
}

func TestMappingCandidateClosedJSONAdmission(t *testing.T) {
	t.Run("all timestamp variants", assertMappingTimestampAdmission)
	for _, profileID := range []string{sourceProfileCiscoSNANetFlowCSV, "cisco_sna_flow_search_csv_v1"} {
		t.Run(profileID, func(t *testing.T) {
			_, result, source := approvedVariantFixture(t, profileID)
			approved, err := decodeApprovedMapping(result.OwnerMapping, defaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			candidate := rewriteMappingObject(t, result.OwnerMapping, nil, func(object map[string]json.RawMessage) json.RawMessage {
				delete(object, "source_columns")
				return mappingTestJSON(t, object)
			})
			assertValid := func(raw json.RawMessage) {
				t.Helper()
				got, err := materializeApprovedMapping(raw, approved.SourceColumns)
				if err != nil || !bytes.Equal(marshalApprovedMapping(got), result.OwnerMapping) {
					t.Fatalf("positive control changed approved bytes: %v", err)
				}
				if fingerprint, err := mappingFingerprint(got, sha256Hex([]byte(source))); err != nil || fingerprint != result.MappingFingerprint {
					t.Fatalf("positive control changed fingerprint: %v", err)
				}
			}
			assertValid(candidate)
			assertValid(bytes.Replace(candidate, []byte(`"target_kind"`), []byte(`"\u0074arget_kind"`), 1))
			for _, mutation := range mappingClosedObjectMutations(t, candidate, true) {
				t.Run(mutation.name, func(t *testing.T) {
					_, err := materializeApprovedMapping(mutation.raw, approved.SourceColumns)
					var conflict *mappingValidationError
					if !errors.As(err, &conflict) || conflict.Code != "network_flow_mapping_conflict" || conflict.ReasonCode != mutation.reason {
						t.Errorf("expected %s, got %v", mutation.reason, err)
					}
				})
			}
			for _, member := range []string{"parser_profile_id", "unknown_column_policy"} {
				assertValid(rewriteMappingObject(t, candidate, nil, func(object map[string]json.RawMessage) json.RawMessage {
					delete(object, member)
					return mappingTestJSON(t, object)
				}))
				invalid := rewriteMappingObject(t, candidate, nil, func(object map[string]json.RawMessage) json.RawMessage {
					object[member] = json.RawMessage(`""`)
					return mappingTestJSON(t, object)
				})
				if _, err := materializeApprovedMapping(invalid, approved.SourceColumns); err == nil {
					t.Errorf("explicit empty %s was defaulted", member)
				}
			}
		})
	}
}

func assertMappingTimestampAdmission(t *testing.T) {
	_, result, _ := approvedVariantFixture(t, sourceProfileCiscoSNANetFlowCSV)
	baseline, err := decodeApprovedMapping(result.OwnerMapping, defaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"rfc3339", "epoch_seconds", "epoch_milliseconds", "netflow_sys_uptime_milliseconds", "cisco_sna_datetime_v1"} {
		t.Run(mode, func(t *testing.T) {
			profile := materializeTimestampProfile(timestampProfile{Mode: mode})
			if mode == "netflow_sys_uptime_milliseconds" {
				export, uptime := 9, 10
				profile.NetFlowExportTimeColumnOrdinal, profile.NetFlowExporterUptimeAtExportColumnOrdinal = &export, &uptime
				profile.NetFlowExportTimeMode = stringPtr("epoch_seconds")
			}
			mapping := baseline
			mapping.TimestampProfile = profile
			if mode == "cisco_sna_datetime_v1" {
				mapping.SourceProfileID = "cisco_sna_flow_search_csv_v1"
			}
			approved := marshalApprovedMapping(mapping)
			if _, err := decodeApprovedMapping(approved, defaultLimits()); err != nil {
				t.Fatalf("invalid approved timestamp control: %v", err)
			}
			candidate := rewriteMappingObject(t, approved, nil, func(object map[string]json.RawMessage) json.RawMessage {
				delete(object, "source_columns")
				return mappingTestJSON(t, object)
			})
			for _, form := range []string{"explicit", "omitted"} {
				raw := candidate
				if form == "omitted" {
					raw = rewriteMappingObject(t, raw, []string{"timestamp_profile"}, func(object map[string]json.RawMessage) json.RawMessage {
						delete(object, "precision")
						return mappingTestJSON(t, object)
					})
				}
				got, err := materializeApprovedMapping(raw, mapping.SourceColumns)
				if err != nil || !bytes.Equal(marshalApprovedMapping(got), approved) {
					t.Fatalf("%s precision changed approved mapping: %v", form, err)
				}
			}
			for _, approvedInput := range []bool{false, true} {
				raw := candidate
				if approvedInput {
					raw = approved
				}
				invalid := rewriteMappingObject(t, raw, []string{"timestamp_profile"}, func(object map[string]json.RawMessage) json.RawMessage {
					object["precision"] = json.RawMessage(`""`)
					return mappingTestJSON(t, object)
				})
				if approvedInput {
					_, err = decodeApprovedMapping(invalid, defaultLimits())
				} else {
					_, err = materializeApprovedMapping(invalid, mapping.SourceColumns)
				}
				if err == nil {
					t.Fatalf("explicit empty precision repaired, approved=%v", approvedInput)
				}
				for _, mutation := range mappingClosedObjectMutations(t, raw, !approvedInput) {
					if !strings.HasPrefix(mutation.name, "timestamp_profile/") {
						continue
					}
					if approvedInput {
						_, err = decodeApprovedMapping(mutation.raw, defaultLimits())
					} else {
						_, err = materializeApprovedMapping(mutation.raw, mapping.SourceColumns)
					}
					var conflict *mappingValidationError
					if !errors.As(err, &conflict) || conflict.ReasonCode != "variant_member_conflict" {
						t.Errorf("approved=%v %s: %v", approvedInput, mutation.name, err)
					}
				}
			}
		})
	}
}
