package extensions

import (
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func checkParticipantContributionReferences(t *testing.T) {
	for _, name := range []string{"missing reference", "retired reference", "wrong contract kind", "wrong owner"} {
		t.Run(name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if object["participant_id"] != "network_flow_activity.reporting_graph_source_v2" {
					return
				}
				if object["kind"] == "snapshot_reporting_participant" {
					switch name {
					case "missing reference":
						delete(object, "participant_contract_ref")
					case "retired reference":
						object["participant_contract_ref"] = "cartulary.extension_participant_contract.v1"
					case "wrong contract kind":
						object["participant_contract_ref"] = "cartulary.extension_transaction_participant_contract.v3"
					}
				}
				if name == "wrong owner" && object["schema_id"] == "cartulary.extension_reporting_source_contract.v1" {
					object["profile_id"] = "snapshot_reporting"
				}
			})
			if _, err := NewCoordinator(source); err == nil {
				t.Fatal("unresolved participant contribution was admitted")
			}
		})
	}
}

func checkParticipantSchemaBindings(t *testing.T) {
	for _, name := range []string{"transaction input schema", "retired transaction input", "retired transaction algorithm", "transaction prepare algorithm", "transaction state family", "state presence identity"} {
		t.Run(name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if object["schema_id"] == "cartulary.extension_state_presence_manifest.v1" && object["profile_id"] == "network_flow_activity" && name == "state presence identity" {
					object["state_presence_manifest_id"] = "reference_pack.state_presence.v1"
				}
				if object["schema_id"] != "cartulary.extension_transaction_participant_contract.v3" || object["participant_id"] != "network_flow_activity.indicator_link_v1" {
					return
				}
				switch name {
				case "transaction input schema":
					object["participant_input_schema_id"] = "cartulary.unresolved_input.v1"
				case "retired transaction input":
					object["participant_input_schema_id"] = "cartulary.network_flow_activity.indicator_link_transaction_input.v1"
				case "retired transaction algorithm":
					object["prepare_algorithm_id"] = "network_flow_activity.indicator_link_prepare_v1"
				case "transaction prepare algorithm":
					object["prepare_algorithm_id"] = "network_flow_activity.unresolved_prepare_v1"
				case "transaction state family":
					object["owned_state_family_ids"] = []any{"network_flow_activity.unresolved_family"}
				}
			})
			if _, err := NewCoordinator(source); err == nil {
				t.Fatal("inconsistent participant or state identity was admitted")
			}
		})
	}
}

func checkNestedMetadataContracts(t *testing.T) {
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	want := TransactionLimits{MinimumParticipants: 1, Participants: 16384, ParticipantInputBytes: 67108864, AggregateInputBytes: 67108864, PrepareResultBytes: 67108864, AggregatePrepareResultBytes: 67108864, SerializationKeysPerParticipant: 1024, AggregateSerializationKeys: 4096, ResultBytes: 1048576, ValidationFindings: 256}
	if got := coordinator.TransactionLimits(); got != want {
		t.Fatalf("admitted transaction limits = %+v", got)
	}
	copy := coordinator.TransactionLimits()
	copy.ParticipantInputBytes = 1
	if coordinator.TransactionLimits() != want {
		t.Fatal("returned limits mutated the admitted catalog")
	}

	for _, name := range []string{"unknown backup row", "unknown participant row", "unknown state row", "wrong participant limits", "wrong shared limit", "wrong encoding", "wrong byte admission", "retired shared protocol", "retired prepare schema", "retired validation schema", "retired write schema", "missing binding schema", "wrong condition multiplicity"} {
		t.Run(name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				switch name {
				case "unknown backup row", "unknown state row":
					id := "cartulary.extension_backup_registry.v1"
					if name == "unknown state row" {
						id = "cartulary.extension_state_registry.v1"
					}
					if rows, ok := object["profiles"].([]any); ok && object["schema_id"] == id && len(rows) > 0 {
						rows[0].(map[string]any)["extra"] = true
					}
				case "unknown participant row":
					if rows, ok := object["participants"].([]any); ok && object["schema_id"] == "cartulary.extension_participant_registry.v1" && len(rows) > 0 {
						rows[0].(map[string]any)["extra"] = true
					}
				case "wrong participant limits":
					if limits, ok := object["transaction_participant_limits"].(map[string]any); ok {
						limits["validation_findings"] = nil
					}
				case "wrong shared limit":
					if object["schema_id"] == "cartulary.extension_shared_protocol_contract_set.v5" {
						if limits, ok := object["transaction"].(map[string]any); ok {
							limits["per_participant_input_max_bytes"] = 16777216
						}
					}
				case "wrong encoding", "wrong byte admission", "retired shared protocol":
					if object["schema_id"] == "cartulary.extension_shared_protocol_contract_set.v5" {
						tx, ok := object["transaction"].(map[string]any)
						if !ok {
							return
						}
						switch name {
						case "wrong encoding":
							tx["canonical_encoding_id"] = "ordinary_json"
						case "wrong byte admission":
							tx["input_byte_admission"] = "normalize"
						case "retired shared protocol":
							object["schema_id"] = "cartulary.extension_shared_protocol_contract_set.v4"
						}
					}
				case "retired prepare schema", "retired validation schema", "retired write schema":
					kind := "prepare"
					if name == "retired validation schema" {
						kind = "validation"
					}
					if name == "retired write schema" {
						kind = "write"
					}
					if object["$id"] == "cartulary.extension_transaction_participant_"+kind+"_result.v2" {
						object["$id"] = "cartulary.extension_transaction_participant_" + kind + "_result.v1"
					}
				case "missing binding schema":
					if object["$id"] == "cartulary.extension_implementation_binding.v2" {
						object["$id"] = "cartulary.unused_schema.v1"
					}
				case "wrong condition multiplicity":
					if object["condition_id"] == "extension_validation_result_overflow" {
						object["multiplicity"] = "unbounded"
					}
				}
			})
			if _, err := NewCoordinator(source); err == nil {
				t.Fatal("digest-valid malformed nested metadata was admitted")
			}
		})
	}
}

// Rebind every affected copy and digest so a rejection must establish contract
// validity rather than merely notice a stale package. This never reads sources.
func rewriteCoherentPackage(t *testing.T, source mutableArtifactSource, mutate func(map[string]any)) {
	t.Helper()
	objects := map[string]any{}
	var walk func(any, func(map[string]any))
	walk = func(value any, visit func(map[string]any)) {
		switch v := value.(type) {
		case map[string]any:
			visit(v)
			for _, child := range v {
				walk(child, visit)
			}
		case []any:
			for _, child := range v {
				walk(child, visit)
			}
		}
	}
	for path, artifact := range source {
		value, err := canonicaljson.DecodeExtension(artifact.JSON)
		if err != nil {
			t.Fatal(err)
		}
		walk(value, mutate)
		objects[path] = value
	}
	replacements := map[string]string{}
	var replace func(any) any
	replace = func(value any) any {
		switch v := value.(type) {
		case string:
			for replacements[v] != "" && replacements[v] != v {
				v = replacements[v]
			}
			return v
		case map[string]any:
			for key, child := range v {
				v[key] = replace(child)
			}
		case []any:
			for i, child := range v {
				v[i] = replace(child)
			}
		}
		return value
	}
	for iteration := 0; iteration < 32; iteration++ {
		changed := false
		for path, value := range objects {
			replace(value)
			body, err := canonicaljson.MarshalExtension(value)
			if err != nil {
				t.Fatal(err)
			}
			digest := canonicalDigest(value)
			old := source[path]
			if digest != old.SHA256 {
				replacements[old.SHA256] = digest
				source[path] = PackagedArtifact{JSON: body, SHA256: digest}
				changed = true
			}
		}
		if !changed {
			return
		}
	}
	t.Fatal("fixture digest graph did not converge")
}

func checkSchemaResourceConfinement(t *testing.T) {
	var requests atomic.Int64
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); _, _ = w.Write([]byte("true")) }))
	defer remote.Close()
	file := filepath.Join(t.TempDir(), "external.json")
	if err := os.WriteFile(file, []byte("true"), 0600); err != nil {
		t.Fatal(err)
	}
	fileURL := (&url.URL{Scheme: "file", Path: file}).String()
	for _, reference := range []string{remote.URL + "/external.json", fileURL} {
		source := newMutableGeneratedSource()
		rewriteCoherentPackage(t, source, func(object map[string]any) {
			if object["$id"] == "cartulary.extension_profile_descriptor.v3" {
				object["$ref"] = reference
			}
		})
		_, err := NewCoordinator(source)
		requireFinding(t, err, "extension_registry_invalid", "registry_generation")
	}
	if requests.Load() != 0 {
		t.Fatal("schema admission performed network I/O")
	}
}

func checkClientSupportJoins(t *testing.T) {
	for _, name := range []string{"wrong major", "missing workspace", "unrecognized schema", "capability", "missing required profile", "duplicate profile"} {
		t.Run(name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if object["schema_id"] != "cartulary.client_extension_support_registry.v1" {
					return
				}
				rows, ok := object["profiles"].([]any)
				if !ok {
					return
				}
				for index, raw := range rows {
					row := raw.(map[string]any)
					if row["profile_id"] != "network_flow_activity" {
						continue
					}
					switch name {
					case "wrong major":
						row["supported_contract_majors"] = []any{7}
					case "missing workspace":
						row["workspace_keys"] = []any{}
					case "unrecognized schema":
						row["public_schema_ids"] = []any{"cartulary.unrecognized.v1"}
					case "capability":
						row["capability_ids"] = []any{"unrecognized"}
					case "missing required profile":
						object["profiles"] = append(rows[:index], rows[index+1:]...)
					case "duplicate profile":
						object["profiles"] = append(rows, row)
					}
					break
				}
			})
			if _, err := NewCoordinator(source); err == nil {
				t.Fatal("inconsistent client support admitted")
			}
		})
	}
}

func checkReportingSourceContracts(t *testing.T) {
	for name, change := range map[string]func(map[string]any){
		"unknown member": func(c map[string]any) { c["extra"] = true },
		"generic context": func(c map[string]any) {
			c["shared_context_schema_id"] = "cartulary.extension_snapshot_reporting_participant_context.v1"
		},
		"generic emit":      func(c map[string]any) { c["operations"] = []any{} },
		"unresolved labels": func(c map[string]any) { c["label_schema_id"] = "cartulary.unresolved.v1" },
		"null budget":       func(c map[string]any) { c["max_label_bytes"] = nil },
		"oversize vertices": func(c map[string]any) { c["max_vertices"] = 100001 },
		"wrong family":      func(c map[string]any) { c["state_family_ids"] = []any{"snapshot_reporting.graph_views"} },
		"duplicate family": func(c map[string]any) {
			c["state_family_ids"] = []any{"network_flow_activity.graph_views", "network_flow_activity.graph_views"}
		},
		"unbound algorithm": func(c map[string]any) { c["read_algorithm_id"] = "network_flow_activity.unimplemented" },
	} {
		t.Run(name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(c map[string]any) {
				if c["schema_id"] == "cartulary.extension_reporting_source_contract.v1" && c["participant_id"] == "network_flow_activity.reporting_graph_source_v2" {
					change(c)
				}
			})
			if _, err := NewCoordinator(source); err == nil {
				t.Fatal("coherent invalid source contract admitted")
			}
		})
	}
}
