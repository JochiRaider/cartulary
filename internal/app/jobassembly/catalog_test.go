package jobassembly

import (
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/gen/contractjobs"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestBaseJobsRemainSelectedWithoutOptionalClaims(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims(nil)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	published, err := extensionassembly.NewPublicationCatalog(plan, coordinator.ParticipantContracts())
	if err != nil {
		t.Fatal(err)
	}
	selected, err := SelectedDefinitions(published)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].JobKind != "evidence.verify_blob_v1" || selected[0].Extension != nil {
		t.Fatalf("unexpected base selection: %#v", selected)
	}
	workers, err := WorkerContracts(published)
	if err != nil {
		t.Fatal(err)
	}
	if len(workers) != 1 || workers[0].ProfileID != "base" || workers[0].WorkerKind != "evidence.blob_verification_worker_v1" || workers[0].MaxActiveAttemptsPerProcess != 8 {
		t.Fatalf("unexpected base worker: %#v", workers)
	}
	recognized, err := RecognizedDefinitions(coordinator.JobKindContracts(), coordinator.WorkerRuntimeContracts())
	if err != nil {
		t.Fatal(err)
	}
	if len(recognized) <= len(selected) {
		t.Fatal("inactive optional definitions disappeared from recognition")
	}
	t.Run("current owner outcomes", func(t *testing.T) { checkCurrentOwnerOutcomes(t, recognized) })
	var discovery jobs.Definition
	for _, definition := range recognized {
		if definition.JobKind == "import.discovery_v1" {
			discovery = definition
		}
	}
	if discovery.JobKind == "" {
		t.Fatal("current import discovery owner is missing")
	}
	valid := jobs.ResultSummary{
		Code: "import_session_discovered", Message: "Import discovered.",
		ResourceRefs: []jobs.ResourceRef{{Kind: "import_session", ID: "12345678-1234-4234-8234-123456789abc", Route: "/api/v1/import-sessions/12345678-1234-4234-8234-123456789abc"}},
	}
	if _, _, _, _, err := jobs.CanonicalTerminalSuccess(discovery, &valid); err != nil {
		t.Fatalf("current owner success rejected: %v", err)
	}
	for _, name := range []string{"unknown code", "wrong owner code", "cancellation as success", "missing required resource"} {
		t.Run(name, func(t *testing.T) {
			invalid := valid
			switch name {
			case "unknown code":
				invalid.Code = "unregistered_success"
			case "wrong owner code":
				invalid.Code = "release_created"
			case "cancellation as success":
				invalid.Code = "job_canceled"
			case "missing required resource":
				invalid.ResourceRefs = nil
			}
			if _, _, _, _, err := jobs.CanonicalTerminalSuccess(discovery, &invalid); err == nil {
				t.Fatal("selected owner contract accepted an invalid terminal success")
			}
		})
	}
	catalog, err := jobs.NewCatalog(recognized)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = jobs.NewRuntimeSelection(catalog, extensionassembly.JobKinds(selected), workers); err != nil {
		t.Fatal(err)
	}
}

func TestBaseJobContractRejectsUnknownAndInconsistentInputs(t *testing.T) {
	raw := contractjobs.Index["contracts/jobs/base.v2.json"].JSON
	if _, _, err := decodeBaseDefinitions(raw); err != nil {
		t.Fatal(err)
	}
	for name, invalid := range map[string]string{
		"unknown field":         strings.Replace(raw, "{", "{\"compatibility\":true,", 1),
		"duplicate field":       strings.Replace(raw, "{", "{\"schema_id\":\"cartulary.base_job_catalog.v2\",", 1),
		"retired identity":      strings.ReplaceAll(raw, "cartulary.base_job_catalog.v2", "cartulary.base_job_catalog.v1"),
		"missing success codes": strings.Replace(raw, ",\"terminal_success_codes\":[\"evidence_blob_verified\"]", "", 1),
		"empty success codes":   strings.Replace(raw, "[\"evidence_blob_verified\"]", "[]", 1),
		"cancellation success":  strings.ReplaceAll(raw, "evidence_blob_verified", "job_canceled"),
		"extension base worker": strings.ReplaceAll(raw, "\"profile_id\":\"base\"", "\"profile_id\":\"import\""),
		"over capacity":         strings.ReplaceAll(raw, "\"max_active_attempts_per_process\":8", "\"max_active_attempts_per_process\":9"),
		"unassigned handler":    strings.Replace(raw, "evidence.blob_verification_worker_v1", "evidence.unassigned_worker_v1", 1),
		"missing jobs":          "{\"schema_id\":\"cartulary.base_job_catalog.v2\",\"workers\":[]}",
		"trailing value":        raw + " {}",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := decodeBaseDefinitions(invalid); err == nil {
				t.Fatal("invalid base catalog was accepted")
			}
		})
	}
}

func checkCurrentOwnerOutcomes(t *testing.T, definitions []jobs.Definition) {
	t.Helper()
	codes := map[string][]string{
		"evidence.verify_blob_v1":                         {"evidence_blob_verified"},
		"import.discovery_v1":                             {"import_session_discovered"},
		"import.apply_v1":                                 {"import_session_applied", "import_session_partially_applied"},
		"incident_portability.export_v2":                  {"incident_bundle_exported"},
		"incident_portability.import_v2":                  {"incident_bundle_imported"},
		"reference_pack.import_v2":                        {"reference_pack_imported"},
		"reference_pack.reverify_v2":                      {"reference_pack_reverified"},
		"reference_pack.refresh_v2":                       {"reference_packs_refreshed"},
		"network_flow_activity.graph_view_materialize_v1": {"network_flow_graph_view_materialized"},
		"snapshot_reporting.composition_preview_v4":       {"composition_preview_rendered"},
		"snapshot_reporting.snapshot_create_v4":           {"snapshot_created"},
		"snapshot_reporting.release_create_v4":            {"release_created"},
	}
	primary := map[string]string{
		"import.discovery_v1": "import_session", "import.apply_v1": "import_session",
		"incident_portability.export_v2": "incident_bundle", "incident_portability.import_v2": "incident",
		"reference_pack.import_v2": "reference_pack_version", "reference_pack.reverify_v2": "reference_pack_version",
		"network_flow_activity.graph_view_materialize_v1": "network_flow_graph_view",
		"snapshot_reporting.snapshot_create_v4":           "snapshot", "snapshot_reporting.release_create_v4": "release",
	}
	id := "11111111-1111-4111-8111-111111111111"
	ref := func(kind string) jobs.ResourceRef {
		value := jobs.ResourceRef{Kind: kind, ID: id}
		switch kind {
		case "incident":
			value.Route = "/api/v1/incidents/" + id
		case "import_session":
			value.Route = "/api/v1/import-sessions/" + id
		case "incident_bundle":
			value.Route = "/api/v1/incident-bundles/" + id
		case "snapshot":
			value.Route = "/api/v1/snapshots/" + id
		case "release":
			value.Route = "/api/v1/releases/" + id
		case "network_flow_table":
			value.ID = "nft_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			value.Route = "/api/v1/incidents/" + id + "/network-flow/tables/" + value.ID
		case "network_flow_graph_view":
			value.ID = "nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			value.Route = "/api/v1/incidents/" + id + "/network-flow/graph-views/" + value.ID
		case "reference_pack_version":
			value.ID = "/api/v1/reference-packs/example.pack/2026.11+build"
			value.Route = value.ID
		default:
			t.Fatal("uncovered current resource kind", kind)
		}
		return value
	}
	if len(definitions) != len(codes) {
		t.Fatal("current owner matrix is incomplete")
	}
	for _, definition := range definitions {
		t.Run(definition.JobKind, func(t *testing.T) {
			expected, ok := codes[definition.JobKind]
			if !ok || !slices.Equal(expected, definition.Terminal.SuccessCodes) {
				t.Fatal("current owner success code projection changed")
			}
			for _, code := range expected {
				result := jobs.ResultSummary{Code: code, Message: "Completed."}
				if kind := primary[definition.JobKind]; kind != "" {
					if _, _, _, _, err := jobs.CanonicalTerminalSuccess(definition, &result); err == nil {
						t.Fatal("required primary resource omitted")
					}
					result.ResourceRefs = []jobs.ResourceRef{ref(kind)}
				}
				if _, _, _, _, err := jobs.CanonicalTerminalSuccess(definition, &result); err != nil {
					t.Fatal("current owner outcome rejected", code, err)
				}
				for _, resource := range definition.Terminal.Resources {
					if resource.Kind != primary[definition.JobKind] {
						result.ResourceRefs = append(result.ResourceRefs, ref(resource.Kind))
					}
				}
				if _, _, _, _, err := jobs.CanonicalTerminalSuccess(definition, &result); err != nil {
					t.Fatal("optional output rejected", code, err)
				}
			}
		})
	}
}
