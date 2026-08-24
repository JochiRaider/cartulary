package incidentbundles_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestPartyIncidentBundleFailuresAreClosedAndAtomic_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "party-incident-bundle-v3-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "party-incident-bundle-v3-target")
	sourceAdmin, sourceAdminID := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-party-bundle-v3-source",
		"incident_key":  "PARTY-BUNDLE-V3",
		"title":         "Party bundle v3 invariant fixture",
	})
	incidentID := incident["incident_id"].(string)
	timelineRow := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-party-bundle-v3-timeline",
		"timeline.activity_synopsis_text": "Party bundle invariant fixture",
	})
	timelineRecordID := timelineRow["row"].(map[string]any)["record_id"].(string)
	seeded := seedIncidentBundlePortableState(t, sourceHarness, incidentID, timelineRecordID, sourceAdminID)
	bundle := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-party-bundle-v3-export")
	baseRows := decodeNDJSONRows(t, zipMemberBytes(t, bundle, "data/parties.ndjson"))
	if len(baseRows) != 1 {
		t.Fatalf("Party fixture rows = %d, want 1", len(baseRows))
	}

	tests := []struct {
		name      string
		txn       string
		invariant string
		hostile   string
		mutate    func([]map[string]any) []map[string]any
	}{
		{
			name: "duplicate identity", txn: "txn-party-bundle-v3-identity",
			invariant: "parties.source_identity_admitted",
			mutate: func(rows []map[string]any) []map[string]any {
				return append(rows, cloneIncidentBundleRow(rows[0]))
			},
		},
		{
			name: "closed shape", txn: "txn-party-bundle-v3-shape",
			invariant: "parties.version_shape_exact", hostile: "future_party_secret",
			mutate: func(rows []map[string]any) []map[string]any {
				rows[0]["future_party_secret"] = "must-not-escape"
				return rows
			},
		},
		{
			name: "lifecycle vocabulary", txn: "txn-party-bundle-v3-lifecycle",
			invariant: "parties.identity_lifecycle", hostile: "UNADMITTED-PARTY-KIND",
			mutate: func(rows []map[string]any) []map[string]any {
				rows[0]["party_kind"] = "UNADMITTED-PARTY-KIND"
				return rows
			},
		},
		{
			name: "normalization", txn: "txn-party-bundle-v3-normalization",
			invariant: "parties.normalization_exact", hostile: "PRIVATE-PARTY-SENTINEL",
			mutate: func(rows []map[string]any) []map[string]any {
				rows[0]["display_name"] = " PRIVATE-PARTY-SENTINEL "
				return rows
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := make([]map[string]any, len(baseRows))
			for index, row := range baseRows {
				rows[index] = cloneIncidentBundleRow(row)
			}
			invalidBundle := replaceStructuredBundleMember(
				t,
				bundle,
				"data/parties.ndjson",
				encodeNDJSONRows(t, test.mutate(rows)),
			)
			terminal := assertImportFailureLeavesState(
				t, targetHarness, targetAdmin, incidentID,
				test.txn, invalidBundle, "source_family_invalid",
			)
			errorSummary := terminal["error_summary"].(map[string]any)
			details := errorSummary["details"].(map[string]any)
			if len(details) != 3 ||
				details["reason_code"] != "source_family_invalid" ||
				details["source_family_id"] != "parties" ||
				details["invariant_id"] != test.invariant {
				t.Fatalf("Party failure details are not closed: %#v", details)
			}
			encoded, err := json.Marshal(terminal)
			if err != nil {
				t.Fatalf("encode Party failure result: %v", err)
			}
			for _, forbidden := range []string{
				test.hostile,
				seeded.PartyRecordID,
				"data/parties.ndjson",
				"party_active_key_claims",
				"constraint",
			} {
				if forbidden != "" && strings.Contains(string(encoded), forbidden) {
					t.Fatalf("Party failure exposed forbidden detail %q: %s", forbidden, encoded)
				}
			}
		})
	}
}

func cloneIncidentBundleRow(row map[string]any) map[string]any {
	cloned := make(map[string]any, len(row)+1)
	for key, value := range row {
		cloned[key] = value
	}
	return cloned
}
