package extensions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contracts"
	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

const (
	extensionsBehaviorVerification   = "module.extensions.verification.behavior_contract"
	extensionsAccountingVerification = "module.extensions.verification.contract_accounting"
)

type extensionBoundaryExpectation struct {
	BoundaryID   string
	AcceptanceID string
	RowID        string
	TestName     string
}

var extensionBoundaryExpectations = []extensionBoundaryExpectation{
	{"BC-001", "EXT-AC-142", "module.extensions.unit.bc001_empty_state_7ca75ba0bc", "TestExtensionBC001EmptyStatePolicy_Unit"},
	{"BC-002", "EXT-AC-143", "module.extensions.unit.bc002_validation_precedence_af944dca6e", "TestExtensionBC002ValidationPrecedence_Unit"},
	{"BC-004", "EXT-AC-145", "module.extensions.static.bc004_dependency_declarations_7dd570e1e4", "TestExtensionBC004DependencyDeclarations_Static"},
	{"BC-005", "EXT-AC-146", "module.extensions.unit.bc005_descriptor_provenance_1e0ea91df8", "TestExtensionBC005DescriptorProvenance_Unit"},
	{"BC-006", "EXT-AC-147", "module.extensions.unit.bc006_validation_inventory_6e85895643", "TestExtensionBC006ValidationInventory_Unit"},
	{"BC-007", "EXT-AC-148", "module.extensions.unit.bc007_closure_mapping_08c4e88841", "TestExtensionBC007ClosureMapping_Unit"},
	{"BC-008", "EXT-AC-149", "module.extensions.static.bc008_clause_traceability_1991b482d2", "TestExtensionBC008ClauseTraceability_Static"},
	{"BC-010", "EXT-AC-151", "module.extensions.process.bc010_lease_lifecycle_4be7ab1e5d", "TestExtensionBC010LeaseLifecycle_Process"},
	{"BC-011", "EXT-AC-152", "module.extensions.integration.bc011_deadline_precedence_ef23af86ac", "TestExtensionBC011DeadlinePrecedence_Integration"},
	{"BC-015", "EXT-AC-156", "module.extensions.integration.bc015_browser_availability_e0a71bee5d", "TestExtensionBC015BrowserAvailability_Integration"},
	{"BC-016", "EXT-AC-157", "module.extensions.unit.bc016_capabilities_disabled_77bb995602", "TestExtensionBC016CapabilitiesDisabled_Unit"},
	{"BC-017", "EXT-AC-158", "module.extensions.process.bc017_component_loss_755919c8d7", "TestExtensionBC017PublishedComponentLoss_Process"},
}

type movedExtensionBoundaryExpectation struct {
	BoundaryID    string
	AcceptanceID  string
	OwnerID       string
	ManifestPath  string
	RowID         string
	Verification  string
	RequiredTests []string
}

var movedExtensionBoundaryExpectations = []movedExtensionBoundaryExpectation{
	{
		BoundaryID: "BC-003", AcceptanceID: "EXT-AC-144", OwnerID: "module.incidentbundles",
		ManifestPath: "module.incidentbundles.json",
		RowID:        "module.incidentbundles.integration.extension_portability_matrix_and_atomic_import_1b27fc2d91",
		Verification: "module.incidentbundles.verification.extensions_portability",
		RequiredTests: []string{
			"TestIncidentBundlePortabilityImportAdmissionAndCleanup_Integration",
			"TestIncidentBundlePortabilityStateAndClaimMatrix_Integration",
		},
	},
	{
		BoundaryID: "BC-009", AcceptanceID: "EXT-AC-150", OwnerID: "platform.config",
		ManifestPath:  "platform.config.json",
		RowID:         "platform.config.unit.inactive_extension_values_discarded_fa9b985016",
		Verification:  "platform.config.verification.behavior_contract",
		RequiredTests: []string{"TestInactiveExtensionConfiguration_Unit"},
	},
	{
		BoundaryID: "BC-012", AcceptanceID: "EXT-AC-153", OwnerID: "module.crossownertransaction",
		ManifestPath: "module.crossownertransaction.json",
		RowID:        "module.crossownertransaction.integration.postgres_atomicity_7f6ad80724",
		Verification: "module.crossownertransaction.verification.final_commit_protocol",
		RequiredTests: []string{
			"TestCrossOwnerTransaction_Integration_CommitAndRollbackAtomicity",
			"TestCrossOwnerTransaction_Integration_OrderedAdvisoryLocks",
		},
	},
	{
		BoundaryID: "BC-013", AcceptanceID: "EXT-AC-154", OwnerID: "module.stagedobjects",
		ManifestPath: "module.stagedobjects.json",
		RowID:        "module.stagedobjects.integration.allocation_publication_cleanup_7f6ad80724",
		Verification: "module.stagedobjects.verification.behavior_contract",
		RequiredTests: []string{
			"TestStagedObjects_Integration_AllocationPublicationAndCleanup",
			"TestStagedObjects_Integration_PublicationRollbackLeavesReadyAndInaccessible",
			"TestStagedObjects_Integration_ReferenceContradictionIsFatalBeforeDelete",
		},
	},
	{
		BoundaryID: "BC-014", AcceptanceID: "EXT-AC-155", OwnerID: "module.recovery",
		ManifestPath: "module.recovery.json",
		RowID:        "module.recovery.integration.selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a",
		Verification: "module.recovery.verification.behavior_contract",
		RequiredTests: []string{
			"TestFailClosedRestoreVerificationBlocked_Unit",
			"TestRestoreRejectsLegacyOrInvalidExtensionBindingEvidenceBeforeMutation_Integration",
			"TestRestoreRejectsRunningOrNonemptyTargetBeforeArtifactRead_Integration",
		},
	},
}

func TestExtensionBC001EmptyStatePolicy_Unit(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-001", "EXT-AC-142")
	assertBC001EmptyStatePolicy(t)
}

func TestExtensionBC002ValidationPrecedence_Unit(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-002", "EXT-AC-143")
	invocation := ClassifyOwnerValidationResult(errors.New("owner failed with secret"), []byte(`{"findings":[]}`))
	requireValidationDisposition(t, invocation, OwnerValidationInvocationFailure, "extension_admission_validation_failed")

	for name, result := range map[string][]byte{
		"invalid_utf8":     {0xff},
		"malformed_json":   []byte(`{"findings":`),
		"trailing_value":   []byte(`{"findings":[]} {}`),
		"non_object":       []byte(`[]`),
		"missing_findings": []byte(`{}`),
		"null_findings":    []byte(`{"findings":null}`),
		"extra_member":     []byte(`{"findings":[],"extra":true}`),
	} {
		t.Run(name, func(t *testing.T) {
			requireValidationDisposition(t, ClassifyOwnerValidationResult(nil, result), OwnerValidationResultInvalid, "extension_validation_result_invalid")
		})
	}

	for _, count := range []int{0, 256, 257, 4096, 4097} {
		t.Run(fmt.Sprintf("count_%d", count), func(t *testing.T) {
			outcome := ClassifyOwnerValidationResult(nil, ownerValidationResult(t, count, false))
			switch {
			case count == 0:
				requireValidationDisposition(t, outcome, OwnerValidationSuccess, "")
			case count <= OwnerFindingSchemaLimit:
				requireValidationDisposition(t, outcome, OwnerValidationFindings, "")
			case count < OwnerFindingOverflowAt:
				requireValidationDisposition(t, outcome, OwnerValidationResultInvalid, "extension_validation_result_invalid")
			default:
				requireValidationDisposition(t, outcome, OwnerValidationOverflow, "extension_diagnostic_overflow")
				if outcome.Limit != 4096 || outcome.Actual != 4097 {
					t.Fatalf("overflow bounds = %d/%d; want 4096/4097", outcome.Limit, outcome.Actual)
				}
			}
		})
	}
	requireValidationDisposition(t, ClassifyOwnerValidationResult(nil, ownerValidationResult(t, 1, true)), OwnerValidationResultInvalid, "extension_validation_result_invalid")
	requireValidationDisposition(t, ClassifyOwnerValidationResult(nil, ownerValidationResult(t, 4097, true)), OwnerValidationOverflow, "extension_diagnostic_overflow")
}

func TestExtensionBC004DependencyDeclarations_Static(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-004", "EXT-AC-145")
	authored := readGeneratedExtensionObject(t, "contracts/extensions/dependencies.json")
	generated := readGeneratedExtensionObject(t, "contracts/extensions/generated/dependency-snapshot.json")
	if authored["schema_id"] != "cartulary.extension_dependency_declaration_set.v1" || generated["schema_id"] != "cartulary.extension_dependency_snapshot.v1" {
		t.Fatalf("unexpected dependency schema transition %v -> %v", authored["schema_id"], generated["schema_id"])
	}
	if authored["extensions_document_version"] != generated["extensions_document_version"] {
		t.Fatal("dependency snapshot changed the authored Extensions version")
	}
	authoredRows := authored["dependencies"].([]any)
	generatedRows := generated["dependencies"].([]any)
	if len(authoredRows) != 10 || len(generatedRows) != len(authoredRows) {
		t.Fatalf("dependency declaration/snapshot counts = %d/%d; want 10/10", len(authoredRows), len(generatedRows))
	}
	for index := range authoredRows {
		if !jsonEqual(authoredRows[index], generatedRows[index]) {
			t.Fatalf("generated dependency %d does not exactly preserve the declaration", index)
		}
		row := generatedRows[index].(map[string]any)
		for _, key := range []string{"imported_anchor_refs", "imported_schema_ids", "imported_algorithm_ids", "imported_artifacts"} {
			if _, ok := row[key].([]any); !ok {
				t.Fatalf("dependency %v does not carry present array %s", row["dependency_id"], key)
			}
		}
	}
}

func TestExtensionBC005DescriptorProvenance_Unit(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-005", "EXT-AC-146")
	ownerInput := readGeneratedExtensionObject(t, "contracts/extensions/generated/owner-input-registry.json")
	registry := readGeneratedExtensionObject(t, "contracts/extensions/generated/profile-registry.json")
	fragments := ownerInput["owner_fragments"].([]any)
	recognition := map[string][]map[string]any{}
	for _, rawFragment := range fragments {
		for _, rawFact := range rawFragment.(map[string]any)["facts"].([]any) {
			fact := rawFact.(map[string]any)
			if fact["fact_kind"] == "recognized_profile" {
				profileID := fact["profile_id"].(string)
				recognition[profileID] = append(recognition[profileID], fact)
			}
			if fact["fact_kind"] == "capability" {
				t.Fatalf("prohibited capability fact reached owner input: %v", fact)
			}
		}
	}
	profiles := registry["profiles"].([]any)
	if len(profiles) != 6 {
		t.Fatalf("registry has %d profiles; want 6", len(profiles))
	}
	for _, rawProfile := range profiles {
		descriptor := rawProfile.(map[string]any)
		profileID := descriptor["profile_id"].(string)
		facts := recognition[profileID]
		if len(facts) != 1 {
			t.Fatalf("profile %s has %d recognition sources; want exactly one", profileID, len(facts))
		}
		fact := facts[0]
		for descriptorKey, factKey := range map[string]string{"claimable": "claimable", "contract_major": "contract_major", "owner_contract_ref": "primary_owner_contract_ref"} {
			if !jsonEqual(descriptor[descriptorKey], fact[factKey]) {
				t.Fatalf("profile %s descriptor %s is not sourced from recognition fact %s", profileID, descriptorKey, factKey)
			}
		}
		if len(descriptor["capability_ids"].([]any)) != 0 {
			t.Fatalf("profile %s generated a prohibited capability", profileID)
		}
	}
}

func TestExtensionBC006ValidationInventory_Unit(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-006", "EXT-AC-147")
	source := readGeneratedExtensionObject(t, "contracts/extensions/validation/surfaces.json")
	registry := readGeneratedExtensionObject(t, "contracts/extensions/generated/validation-condition-registry.json")
	want := map[string]struct{}{}
	for _, rawDeclaration := range source["declarations"].([]any) {
		declaration := rawDeclaration.(map[string]any)
		for _, family := range []string{"schema_surfaces", "procedural_surfaces"} {
			for _, rawSurface := range declaration[family].([]any) {
				for _, rawCondition := range rawSurface.(map[string]any)["conditions"].([]any) {
					conditionID := rawCondition.(map[string]any)["condition_id"].(string)
					if _, duplicate := want[conditionID]; duplicate {
						t.Fatalf("duplicate authored validation condition %s", conditionID)
					}
					want[conditionID] = struct{}{}
				}
			}
		}
	}
	got := registry["conditions"].([]any)
	if len(got) != len(want) {
		t.Fatalf("condition registry has %d rows; authored surfaces have %d", len(got), len(want))
	}
	previous := ""
	for _, rawCondition := range got {
		condition := rawCondition.(map[string]any)
		conditionID := condition["condition_id"].(string)
		if previous >= conditionID {
			t.Fatalf("condition registry is not strictly ordered at %s", conditionID)
		}
		previous = conditionID
		if _, exists := want[conditionID]; !exists {
			t.Fatalf("generated unregistered condition %s", conditionID)
		}
		if condition["secret_policy"] == "redacted" && condition["actual_formatter_id"] != "diagnostic_redacted_v1" {
			t.Fatalf("redacted condition %s uses unsafe formatter", conditionID)
		}
	}
	coordinator := requireGeneratedCoordinator(t)
	condition, err := coordinator.RequireRegisteredCondition("extension_validation_result_overflow")
	if err != nil || condition.ReasonCode != "extension_diagnostic_overflow" {
		t.Fatalf("registered runtime condition = %#v, %v", condition, err)
	}
	if _, err := coordinator.RequireRegisteredCondition("implementation_invented_condition"); err == nil {
		t.Fatal("runtime accepted an unregistered emitted condition")
	} else {
		requireFinding(t, err, "extension_validation_result_invalid", "profile_preflight")
	}
}

func TestExtensionBC007ClosureMapping_Unit(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-007", "EXT-AC-148")
	source := readGeneratedExtensionObject(t, "contracts/extensions/specification/closure-mapping.json")
	registry := readGeneratedExtensionObject(t, "contracts/extensions/generated/profile-registry.json")
	categoryMapping := source["contribution_categories"].(map[string]any)
	for _, rawProfile := range registry["profiles"].([]any) {
		descriptor := rawProfile.(map[string]any)
		profileID := descriptor["profile_id"].(string)
		catalog := readGeneratedExtensionObject(t, "contracts/extensions/generated/closure-catalogs/"+profileID+".json")
		items := catalog["items"].([]any)
		actual := map[string]map[string]struct{}{}
		for _, rawItem := range items {
			item := rawItem.(map[string]any)
			if item["subject_kind"] != "contribution" {
				continue
			}
			if len(item["allowed_not_applicable_reason_codes"].([]any)) != 0 {
				t.Fatalf("profile %s generated contribution closure permits not_applicable", profileID)
			}
			subjectID := item["subject_id"].(string)
			if actual[subjectID] == nil {
				actual[subjectID] = map[string]struct{}{}
			}
			actual[subjectID][item["category"].(string)] = struct{}{}
		}
		for _, rawContribution := range descriptor["contributions"].([]any) {
			contribution := rawContribution.(map[string]any)
			contributionID := contribution["contribution_id"].(string)
			wantCategories := categoryMapping[contribution["kind"].(string)].([]any)
			if len(actual[contributionID]) != len(wantCategories) {
				t.Fatalf("profile %s contribution %s closure count = %d; want %d", profileID, contributionID, len(actual[contributionID]), len(wantCategories))
			}
			for _, rawCategory := range wantCategories {
				if _, exists := actual[contributionID][rawCategory.(string)]; !exists {
					t.Fatalf("profile %s contribution %s omits closure category %s", profileID, contributionID, rawCategory)
				}
			}
		}
	}
}

func TestExtensionBC008ClauseTraceability_Static(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-008", "EXT-AC-149")
	source := readGeneratedExtensionObject(t, "contracts/extensions/traceability/mapping-source.json")
	trace := readGeneratedExtensionObject(t, "contracts/extensions/generated/clause-traceability.json")
	if trace["extensions_document_sha256"] != source["extensions_document_sha256"] {
		t.Fatal("traceability output is bound to a different Extensions source digest")
	}
	mappings := source["mappings"].([]any)
	clauses := trace["clauses"].([]any)
	if len(clauses) != len(mappings) || len(clauses) < 394 {
		t.Fatalf("traceability clauses/mappings = %d/%d; want equal complete requirement/criterion coverage", len(clauses), len(mappings))
	}
	document := []byte(readExtensionContractFile(t, filepath.Join(extensionsRepoRoot(t), "docs", "extension-subsystem-nlspec.md")))
	seenRequirements := map[string]bool{}
	seenAcceptance := map[string]bool{}
	for index, rawClause := range clauses {
		clause := rawClause.(map[string]any)
		mapping := mappings[index].(map[string]any)
		start := int(mapping["source_start_byte"].(float64))
		end := int(mapping["source_end_byte"].(float64))
		textDigest := sha256.Sum256(document[start:end])
		if clause["clause_text_sha256"] != hex.EncodeToString(textDigest[:]) {
			t.Fatalf("clause %d text digest does not match its half-open source range", index)
		}
		if int(clause["document_ordinal"].(float64)) != index || clause["source_start_byte"] != mapping["source_start_byte"] || clause["source_end_byte"] != mapping["source_end_byte"] {
			t.Fatalf("clause %d changed mapping identity or ordinal", index)
		}
		if !strings.HasPrefix(clause["clause_id"].(string), "extcl:") || len(clause["clause_id"].(string)) != 38 {
			t.Fatalf("clause %d has invalid deterministic ID %v", index, clause["clause_id"])
		}
		for _, requirementID := range requireJSONStringsValue(t, clause["requirement_ids"], "requirement_ids") {
			seenRequirements[requirementID] = true
		}
		for _, acceptanceID := range requireJSONStringsValue(t, clause["acceptance_criterion_ids"], "acceptance_criterion_ids") {
			seenAcceptance[acceptanceID] = true
		}
		if len(requireJSONStringsValue(t, clause["verification_ids"], "verification_ids")) == 0 {
			t.Fatalf("clause %d has no active verification mapping", index)
		}
	}
	for value := 1; value <= 236; value++ {
		requirementID := fmt.Sprintf("EXT-REQ-%03d", value)
		if !seenRequirements[requirementID] {
			t.Fatalf("traceability omits %s", requirementID)
		}
	}
	for value := 1; value <= 158; value++ {
		acceptanceID := fmt.Sprintf("EXT-AC-%03d", value)
		if !seenAcceptance[acceptanceID] {
			t.Fatalf("traceability omits %s", acceptanceID)
		}
	}
}

func requireJSONStringsValue(t testing.TB, value any, label string) []string {
	t.Helper()
	raw, ok := value.([]any)
	if !ok {
		t.Fatalf("%s is not an array: %#v", label, value)
	}
	result := make([]string, len(raw))
	for index, item := range raw {
		current, ok := item.(string)
		if !ok {
			t.Fatalf("%s[%d] is not a string: %#v", label, index, item)
		}
		result[index] = current
	}
	return result
}

func TestExtensionBC010LeaseLifecycle_Process(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-010", "EXT-AC-151")
	session := newContractLeaseSession("original-session")
	lease, err := processlease.Acquire(context.Background(), contractLeaseBackend{session: session}, 100*time.Millisecond, 20*time.Millisecond)
	if err != nil || lease.State() != processlease.StateHeld || lease.SessionIdentity() != "original-session" {
		t.Fatalf("initial acquisition = %v/%s/%q", err, lease.State(), lease.SessionIdentity())
	}
	monitorCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lease.StartMonitor(monitorCtx)
	session.setProof(processlease.ProofUncertain)
	waitForLeaseState(t, lease, processlease.StateUncertain)
	session.setProof(processlease.ProofContinuous)
	waitForLeaseState(t, lease, processlease.StateHeld)
	session.setIdentity("different-session")
	waitForLeaseState(t, lease, processlease.StateLost)
	if err := lease.Release(context.Background()); !errors.Is(err, processlease.ErrInvalidTransition) {
		t.Fatalf("release after loss = %v; want invalid transition", err)
	}
	if lease.State() != processlease.StateLost {
		t.Fatalf("lost lease changed state to %s", lease.State())
	}

	timeoutSession := newContractLeaseSession("contender")
	timeoutSession.acquire = false
	_, err = processlease.Acquire(context.Background(), contractLeaseBackend{session: timeoutSession}, 5*time.Millisecond, time.Second)
	if !errors.Is(err, processlease.ErrApplicationProcessActive) {
		t.Fatalf("acquisition timeout = %v", err)
	}

	releaseSession := newContractLeaseSession("release-session")
	released, err := processlease.Acquire(context.Background(), contractLeaseBackend{session: releaseSession}, 100*time.Millisecond, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := released.Release(context.Background()); err != nil || released.State() != processlease.StateReleased || !releaseSession.wasReleased() {
		t.Fatalf("release = %v/%s/%t", err, released.State(), releaseSession.wasReleased())
	}
}

func TestExtensionBC011DeadlinePrecedence_Integration(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-011", "EXT-AC-152")
	local := extensiondeadline.New(10, 2, nil)
	if local.MonotonicNS != 2_000_000_010 || local.Source != extensiondeadline.SourceLocal || local.Expired(local.MonotonicNS-1) || !local.Expired(local.MonotonicNS) {
		t.Fatalf("local deadline = %#v", local)
	}
	if saturated := extensiondeadline.New(math.MaxInt64-5, 1, nil); saturated.MonotonicNS != math.MaxInt64 {
		t.Fatalf("addition did not saturate: %#v", saturated)
	}
	if saturated := extensiondeadline.New(1, math.MaxInt64, nil); saturated.MonotonicNS != math.MaxInt64 {
		t.Fatalf("multiplication did not saturate: %#v", saturated)
	}
	for name, inherited := range map[string]int64{"earlier": 100, "equal": 2_000_000_010, "later": 3_000_000_000} {
		t.Run(name, func(t *testing.T) {
			result := extensiondeadline.New(10, 2, &extensiondeadline.Deadline{MonotonicNS: inherited, Source: extensiondeadline.SourceInherited})
			if name == "later" && (result.Source != extensiondeadline.SourceLocal || result.MonotonicNS != local.MonotonicNS) {
				t.Fatalf("later inherited deadline won: %#v", result)
			}
			if name != "later" && (result.Source != extensiondeadline.SourceInherited || result.MonotonicNS != inherited) {
				t.Fatalf("inherited minimum lost: %#v", result)
			}
		})
	}
	cancelBefore, cancelEqual, cancelAfter, expiry := int64(9), int64(10), int64(11), int64(10)
	if got := extensiondeadline.Classify(extensiondeadline.CommitProvenAbsent, &cancelBefore, &expiry); got != extensiondeadline.OutcomeCanceled {
		t.Fatalf("cancel before expiry = %s", got)
	}
	for _, sample := range []*int64{&cancelEqual, &cancelAfter, nil} {
		if got := extensiondeadline.Classify(extensiondeadline.CommitProvenAbsent, sample, &expiry); got != extensiondeadline.OutcomeTimedOut {
			t.Fatalf("expiry precedence = %s", got)
		}
	}
	if got := extensiondeadline.Classify(extensiondeadline.CommitProvenSuccessful, &cancelBefore, &expiry); got != extensiondeadline.OutcomeCommitted {
		t.Fatalf("proven commit did not win: %s", got)
	}
	if got := extensiondeadline.Classify(extensiondeadline.CommitIndeterminate, &cancelBefore, &expiry); got != extensiondeadline.OutcomeFatal {
		t.Fatalf("indeterminate commit = %s", got)
	}
}

func TestExtensionBC015BrowserAvailability_Integration(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-015", "EXT-AC-156")
	registry := readGeneratedExtensionObject(t, "contracts/extensions/generated/profile-registry.json")
	var networkFlow map[string]any
	for _, rawProfile := range registry["profiles"].([]any) {
		profile := rawProfile.(map[string]any)
		if profile["profile_id"] == "network_flow_activity" {
			networkFlow = profile
			break
		}
	}
	if networkFlow == nil || networkFlow["claimable"] != true || networkFlow["contract_major"] != float64(2) {
		t.Fatalf("Network Flow browser descriptor is not claimable major 2: %#v", networkFlow)
	}
	requireJSONStrings(t, networkFlow["workspace_keys"], []string{"network_analysis"}, "Network Flow workspace keys")
	requireJSONStrings(t, networkFlow["capability_ids"], []string{}, "Network Flow capabilities")

	supportSource := readGeneratedExtensionObject(t, "contracts/extensions/build/client-support.json")
	if supportSource["client_build_class"] != "standard" {
		t.Fatalf("client build class = %v; want standard", supportSource["client_build_class"])
	}
	rows := supportSource["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("client support rows = %d; want 2", len(rows))
	}
	rowsByProfile := map[string]map[string]any{}
	for _, rawRow := range rows {
		row := rawRow.(map[string]any)
		rowsByProfile[row["profile_id"].(string)] = row
	}
	importRow := rowsByProfile["import"]
	if importRow == nil || importRow["contract_major"] != float64(1) ||
		importRow["client_asset_set_id"] != "import.standard.v1" {
		t.Fatalf("client support row does not select Import major 1: %#v", importRow)
	}
	requireJSONStrings(t, importRow["workspace_keys"], []string{}, "Import client support workspaces")
	requireJSONStrings(t, importRow["capability_ids"], []string{}, "Import client support capabilities")
	requireJSONStrings(t, importRow["public_schema_ids"], []string{}, "Import client support public schemas")

	row := rowsByProfile["network_flow_activity"]
	if row == nil || row["profile_id"] != "network_flow_activity" || row["contract_major"] != float64(2) {
		t.Fatalf("client support row does not select Network Flow major 2: %#v", row)
	}
	requireJSONStrings(t, row["workspace_keys"], []string{"network_analysis"}, "client support workspaces")
	requireJSONStrings(t, row["capability_ids"], []string{}, "client support capabilities")
	requireJSONStrings(t, row["public_schema_ids"], []string{}, "client support public schemas")

	assetSchema := readGeneratedExtensionObject(t, "contracts/extensions/generated/schemas/cartulary.client_asset_set_manifest.v1.schema.json")
	if assetSchema["$id"] != "cartulary.client_asset_set_manifest.v1" {
		t.Fatalf("client asset manifest schema identity = %v", assetSchema["$id"])
	}
	coordinator := requireGeneratedCoordinator(t)
	boundDigest := strings.Repeat("a", 64)
	bound, err := coordinator.WithClientSupportRegistrySHA256(boundDigest)
	if err != nil {
		t.Fatalf("bind final client support digest: %v", err)
	}
	resolution, err := bound.ResolveClaims([]string{"import", "network_flow_activity"})
	if err != nil {
		t.Fatalf("resolve browser claims: %v", err)
	}
	plan, err := bound.BuildPublicationPlan(resolution)
	if err != nil || plan.Summary().ClientSupportRegistrySHA256 != boundDigest {
		t.Fatalf("publication plan client support digest = %q, %v", plan.Summary().ClientSupportRegistrySHA256, err)
	}
}

func TestExtensionBC016CapabilitiesDisabled_Unit(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-016", "EXT-AC-157")
	registry := readGeneratedExtensionObject(t, "contracts/extensions/generated/profile-registry.json")
	for _, rawProfile := range registry["profiles"].([]any) {
		profile := rawProfile.(map[string]any)
		requireJSONStrings(t, profile["capability_ids"], []string{}, fmt.Sprintf("%s capabilities", profile["profile_id"]))
	}
	bindings := contractsgen.ExtensionArtifactsIndex
	for path, artifact := range bindings {
		if !strings.Contains(path, "/implementation-bindings/") {
			continue
		}
		var binding map[string]any
		decoder := json.NewDecoder(strings.NewReader(artifact.JSON))
		decoder.UseNumber()
		if err := decoder.Decode(&binding); err != nil {
			t.Fatalf("decode binding %s: %v", path, err)
		}
		requireJSONStrings(t, binding["supported_capability_ids"], []string{}, path+" supported capabilities")
	}
}

func TestExtensionBC017PublishedComponentLoss_Process(t *testing.T) {
	requireExtensionBoundaryRoute(t, "BC-017", "EXT-AC-158")
	lifecycle := processlifecycle.New()
	if lifecycle.AdmissionOpen() {
		t.Fatal("admission opened before atomic publication")
	}
	if err := lifecycle.Publish(); err != nil || !lifecycle.AdmissionOpen() || lifecycle.State() != processlifecycle.StateRunning {
		t.Fatalf("publication = %v/%s/%t", err, lifecycle.State(), lifecycle.AdmissionOpen())
	}
	if !lifecycle.Fatal("published_component_lost") || lifecycle.AdmissionOpen() || lifecycle.State() != processlifecycle.StateQuiescing {
		t.Fatalf("fatal transition = %s/%t", lifecycle.State(), lifecycle.AdmissionOpen())
	}
	if lifecycle.RestoreAdmission() || lifecycle.Fatal("published_component_lost") {
		t.Fatal("fatal loss was restartable or non-idempotent")
	}
	if err := lifecycle.Publish(); err == nil {
		t.Fatal("fatal lifecycle republished in process")
	}
	select {
	case signal := <-lifecycle.FatalEvents():
		if signal.ReasonCode != "published_component_lost" || signal.ExitCode != 70 {
			t.Fatalf("fatal signal = %#v", signal)
		}
	case <-time.After(time.Second):
		t.Fatal("fatal signal was not emitted")
	}
}

func requireValidationDisposition(t *testing.T, outcome OwnerValidationOutcome, disposition OwnerValidationDisposition, reason string) {
	t.Helper()
	if outcome.Disposition != disposition || outcome.ReasonCode != reason {
		t.Fatalf("validation outcome = %#v; want %s/%q", outcome, disposition, reason)
	}
}

func ownerValidationResult(t *testing.T, count int, malformed bool) []byte {
	t.Helper()
	findings := make([]any, count)
	for index := range findings {
		findings[index] = map[string]any{
			"path": fmt.Sprintf("$.rows[%06d]", count-index), "reason_code": "invalid_value", "message": "invalid value", "details": map[string]any{"index": index},
		}
	}
	if malformed && count > 0 {
		findings[0] = map[string]any{"path": "$"}
	}
	encoded, err := json.Marshal(map[string]any{"findings": findings})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

type contractLeaseBackend struct{ session *contractLeaseSession }

func (b contractLeaseBackend) Open(context.Context) (processlease.Session, error) {
	return b.session, nil
}

type contractLeaseSession struct {
	mu       sync.RWMutex
	identity string
	proof    processlease.Proof
	acquire  bool
	released bool
}

func newContractLeaseSession(identity string) *contractLeaseSession {
	return &contractLeaseSession{identity: identity, proof: processlease.ProofContinuous, acquire: true}
}

func (s *contractLeaseSession) Identity() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.identity
}

func (s *contractLeaseSession) TryAcquire(context.Context) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.acquire, nil
}

func (s *contractLeaseSession) Prove(context.Context) processlease.Proof {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.proof
}

func (s *contractLeaseSession) Release(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.released = true
	return nil
}

func (s *contractLeaseSession) Close() {}

func (s *contractLeaseSession) setProof(proof processlease.Proof) {
	s.mu.Lock()
	s.proof = proof
	s.mu.Unlock()
}

func (s *contractLeaseSession) setIdentity(identity string) {
	s.mu.Lock()
	s.identity = identity
	s.mu.Unlock()
}

func (s *contractLeaseSession) wasReleased() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.released
}

func waitForLeaseState(t *testing.T, lease *processlease.Lease, want processlease.State) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if lease.State() == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("lease state = %s; want %s", lease.State(), want)
}

func TestExtensionContractAccounting_Static(t *testing.T) {
	root := extensionsRepoRoot(t)
	manifest := readExtensionsFamilyManifest(t, root)
	if manifest.SchemaID != "cartulary.test_family_manifest.v1" || manifest.OwnerID != "module.extensions" {
		t.Fatalf("unexpected Extensions family identity %q/%q", manifest.SchemaID, manifest.OwnerID)
	}

	rows := make(map[string]extensionTestFamilyRow, len(manifest.Rows))
	for _, row := range manifest.Rows {
		if _, duplicate := rows[row.RowID]; duplicate {
			t.Fatalf("duplicate Extensions family row %q", row.RowID)
		}
		rows[row.RowID] = row
		if row.ClaimPosture != "implementation" {
			t.Errorf("active Extensions row %s has claim_posture=%q; want implementation", row.RowID, row.ClaimPosture)
		}
	}
	for _, expected := range extensionBoundaryExpectations {
		row, ok := rows[expected.RowID]
		if !ok {
			t.Errorf("%s/%s has no exact primary-owner row %q", expected.BoundaryID, expected.AcceptanceID, expected.RowID)
			continue
		}
		requireExactStrings(t, row.VerificationIDs, []string{extensionsBehaviorVerification}, expected.RowID+" verification_ids")
		requireExactStrings(t, row.Selector.Tests, []string{expected.TestName}, expected.RowID+" selector.tests")
		requireContains(t, row.DocumentationRefs, "docs/extension-subsystem-nlspec.md#"+strings.ToLower(expected.AcceptanceID))
		requireContains(t, row.DocumentationRefs, "docs/handoffs/extensions-subsystem-implementation-tracker.md#"+strings.ToLower(expected.BoundaryID))
	}
	for _, expected := range movedExtensionBoundaryExpectations {
		ownerManifest := readTestFamilyManifest(t, root, expected.ManifestPath)
		if ownerManifest.OwnerID != expected.OwnerID {
			t.Errorf("%s/%s owner manifest is %q; want %q", expected.BoundaryID, expected.AcceptanceID, ownerManifest.OwnerID, expected.OwnerID)
			continue
		}
		var movedRow *extensionTestFamilyRow
		for index := range ownerManifest.Rows {
			if ownerManifest.Rows[index].RowID == expected.RowID {
				movedRow = &ownerManifest.Rows[index]
				break
			}
		}
		if movedRow == nil {
			t.Errorf("%s/%s has no owner row %q", expected.BoundaryID, expected.AcceptanceID, expected.RowID)
			continue
		}
		requireExactStrings(t, movedRow.VerificationIDs, []string{expected.Verification}, expected.RowID+" verification_ids")
		requireExactStrings(t, movedRow.Selector.Tests, expected.RequiredTests, expected.RowID+" selector.tests")
		requireContains(t, movedRow.DocumentationRefs, "docs/extension-subsystem-nlspec.md#"+strings.ToLower(expected.AcceptanceID))
		requireContains(t, movedRow.DocumentationRefs, "docs/handoffs/extensions-subsystem-implementation-tracker.md#"+strings.ToLower(expected.BoundaryID))
	}

	accounting, ok := rows["module.extensions.static.contract_accounting_e80c9e3dc7"]
	if !ok {
		t.Fatal("Extensions contract-accounting row is missing")
	}
	requireExactStrings(t, accounting.VerificationIDs, []string{extensionsAccountingVerification}, "contract accounting verification_ids")
	requireExactStrings(t, accounting.Selector.Tests, []string{"TestExtensionContractAccounting_Static"}, "contract accounting selector.tests")
	coordinatorRows := map[string][]string{
		"module.extensions.unit.coordinator_binding_admission_c2d8beffdb":   {"TestCoordinatorBindingAdmission_Unit"},
		"module.extensions.unit.coordinator_claim_resolution_6cfce6db24":    {"TestCoordinatorClaimResolution_Unit"},
		"module.extensions.unit.coordinator_collision_admission_590bd27481": {"TestCoordinatorCollisionAdmission_Unit"},
		"module.extensions.unit.coordinator_publication_plan_05ae1ed79d":    {"TestCoordinatorPublicationPlan_Unit"},
		"module.extensions.unit.coordinator_registry_1dce9a539a": {
			"TestCoordinatorGeneratedRegistry_Unit",
			"TestCoordinatorPortabilityPolicyProjection_Unit",
			"TestExtensionProfileAdoptionMatrix_Static",
		},
	}
	for rowID, testNames := range coordinatorRows {
		row, exists := rows[rowID]
		if !exists {
			t.Fatalf("coordinator row %s is missing", rowID)
		}
		requireExactStrings(t, row.VerificationIDs, []string{extensionsBehaviorVerification}, rowID+" verification_ids")
		requireExactStrings(t, row.Selector.Tests, testNames, rowID+" selector.tests")
	}
	characterizationRows := map[string]int{
		"module.extensions.integration.state_admission_matrix_7f6ad80724": 21,
	}
	for rowID, selectorCount := range characterizationRows {
		row, exists := rows[rowID]
		if !exists {
			t.Fatalf("characterization row %s is missing", rowID)
		}
		requireExactStrings(t, row.VerificationIDs, []string{extensionsBehaviorVerification}, rowID+" verification_ids")
		if len(row.Selector.Tests) != selectorCount {
			t.Fatalf("%s selector count = %d, want %d", rowID, len(row.Selector.Tests), selectorCount)
		}
	}
	jobRows := map[string][]string{
		"module.extensions.unit.job_admission_and_terminal_contracts_64d299b374": {
			"TestCanonicalExtensionTerminalSuccessValidatesResourceContracts_Unit",
			"TestExtensionJobAdmissionMetadataIsClosedAndInternal_Unit",
		},
		"module.extensions.unit.job_inactive_reconciliation_49b5990495": {
			"TestReconcileInactiveExtensionJobsFailsBeforeMutation_Unit",
			"TestReconcileInactiveExtensionJobs_Unit",
		},
		"module.extensions.integration.job_finalization_cancellation_and_reconciliation_2c712826c2": {
			"TestExtensionCancellationObservationIsAtomic_Integration",
			"TestOwnerFinalizerAtomicSuccessAndFailure_Integration",
		},
		"module.extensions.integration.inactive_job_reconciliation_adapter_5c3e201b45": {
			"TestInactiveExtensionJobReconciliation_ServiceBacked",
		},
		"module.extensions.integration.clean_job_cutover_migration_4917a0cdef": {
			"TestExtensionJobCutoverMigration34FreshSchema_Integration",
			"TestExtensionJobCutoverMigration34RejectsEveryRetiredHandlerBeforeMutation_Integration",
		},
	}
	for rowID, testNames := range jobRows {
		row, exists := rows[rowID]
		if !exists {
			t.Fatalf("extension job row %s is missing", rowID)
		}
		requireExactStrings(t, row.VerificationIDs, []string{extensionsBehaviorVerification}, rowID+" verification_ids")
		requireExactStrings(t, row.Selector.Tests, testNames, rowID+" selector.tests")
	}
	browserAvailability, exists := rows["module.extensions.browser_stateful.bc015_availability_continuity_d538000c38"]
	if !exists {
		t.Fatal("Extensions browser availability continuity row is missing")
	}
	requireExactStrings(t, browserAvailability.VerificationIDs, []string{extensionsBehaviorVerification}, "browser availability verification_ids")
	if browserAvailability.Runner != "playwright" || browserAvailability.Selector.File != "apps/web/e2e/extensions.stateful.spec.ts" || browserAvailability.Selector.Stage != "stateful" {
		t.Fatalf("browser availability selector is not exact: %#v", browserAvailability.Selector)
	}
	if got, want := len(rows), len(extensionBoundaryExpectations)+2+len(coordinatorRows)+len(characterizationRows)+len(jobRows); got != want {
		t.Fatalf("Extensions manifest has %d rows; want exactly %d", got, want)
	}

	traceability := readGeneratedExtensionObject(t, "contracts/extensions/generated/clause-traceability.json")
	criteriaByVerification := map[string]map[string]bool{}
	for _, rawClause := range traceability["clauses"].([]any) {
		clause := rawClause.(map[string]any)
		criteria := requireJSONStringsValue(t, clause["acceptance_criterion_ids"], "acceptance_criterion_ids")
		for _, verificationID := range requireJSONStringsValue(t, clause["verification_ids"], "verification_ids") {
			if criteriaByVerification[verificationID] == nil {
				criteriaByVerification[verificationID] = map[string]bool{}
			}
			for _, acceptanceID := range criteria {
				criteriaByVerification[verificationID][acceptanceID] = true
			}
		}
	}
	coveredCriteria := map[string]bool{}
	for verificationID, criteria := range criteriaByVerification {
		hasExactRow := false
		for _, row := range rows {
			if slices.Contains(row.VerificationIDs, verificationID) {
				hasExactRow = true
				break
			}
		}
		if !hasExactRow {
			t.Errorf("traceability verification %s has no active exact implementation row", verificationID)
		}
		for acceptanceID := range criteria {
			coveredCriteria[acceptanceID] = true
		}
	}
	allAcceptanceIDs := make([]string, 158)
	for value := 1; value <= 158; value++ {
		acceptanceID := fmt.Sprintf("EXT-AC-%03d", value)
		allAcceptanceIDs[value-1] = acceptanceID
		if !coveredCriteria[acceptanceID] {
			t.Errorf("%s has no active verification and exact implementation-row path", acceptanceID)
		}
	}
	for _, profileID := range []string{"enterprise_authentication", "import", "incident_portability", "network_flow_activity", "reference_pack", "snapshot_reporting"} {
		conformance := readGeneratedExtensionObject(t, "contracts/extensions/generated/conformance-manifests/"+profileID+".json")
		requireExactStrings(t, requireJSONStringsValue(t, conformance["acceptance_criterion_ids"], profileID+" acceptance_criterion_ids"), allAcceptanceIDs, profileID+" acceptance_criterion_ids")
	}
	requireResolvedClaimSetIdentity(t)
}

func requireResolvedClaimSetIdentity(t testing.TB) {
	t.Helper()
	claims, err := NewResolvedClaimSet([]string{"snapshot_reporting", "import", "import"})
	if err != nil {
		t.Fatalf("construct resolved claim set: %v", err)
	}
	requireExactStrings(t, claims.ProfileIDs(), []string{"import", "snapshot_reporting"}, "canonical resolved profile ids")
	if len(claims.SHA256()) != 64 {
		t.Fatalf("resolved claim digest has %d characters; want 64", len(claims.SHA256()))
	}
	copy := claims.ProfileIDs()
	copy[0] = "mutated"
	requireExactStrings(t, claims.ProfileIDs(), []string{"import", "snapshot_reporting"}, "immutable resolved profile ids")
	for _, invalid := range []string{"base", "", "BAD", "contains-dash"} {
		if _, err := NewResolvedClaimSet([]string{invalid}); err == nil {
			t.Fatalf("invalid resolved profile id %q was accepted", invalid)
		}
	}
}

func requireJSONStrings(t testing.TB, value any, want []string, label string) {
	t.Helper()
	raw, ok := value.([]any)
	if !ok {
		t.Fatalf("%s is not an array: %#v", label, value)
	}
	got := make([]string, len(raw))
	for index, item := range raw {
		current, ok := item.(string)
		if !ok {
			t.Fatalf("%s[%d] is not a string: %#v", label, index, item)
		}
		got[index] = current
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("%s = %v; want %v", label, got, want)
	}
}

func requireExtensionBoundaryRoute(t *testing.T, boundaryID, acceptanceID string) {
	t.Helper()
	root := extensionsRepoRoot(t)
	spec := readExtensionContractFile(t, filepath.Join(root, "docs", "extension-subsystem-nlspec.md"))
	tracker := readExtensionContractFile(t, filepath.Join(root, "docs", "handoffs", "extensions-subsystem-implementation-tracker.md"))
	if !strings.Contains(spec, "| `"+acceptanceID+"` |") {
		t.Fatalf("Extensions NLSpec omits exact acceptance criterion %s", acceptanceID)
	}
	if !strings.Contains(tracker, "| "+boundaryID+" | "+acceptanceID+" |") {
		t.Fatalf("implementation tracker omits exact %s to %s mapping", boundaryID, acceptanceID)
	}
	if !strings.Contains(tracker, "| "+boundaryID+" | Minimum executable scenarios") && !strings.Contains(tracker, "| "+boundaryID+" |") {
		t.Fatalf("implementation tracker omits the %s scenario inventory", boundaryID)
	}
}

type extensionTestFamilyManifest struct {
	SchemaID string                   `json:"schema_id"`
	OwnerID  string                   `json:"owner_id"`
	Rows     []extensionTestFamilyRow `json:"rows"`
}

type extensionTestFamilyRow struct {
	RowID             string   `json:"row_id"`
	Runner            string   `json:"runner"`
	ClaimPosture      string   `json:"claim_posture"`
	VerificationIDs   []string `json:"verification_ids"`
	DocumentationRefs []string `json:"documentation_refs"`
	Selector          struct {
		File  string   `json:"file"`
		Stage string   `json:"stage"`
		Tests []string `json:"tests"`
	} `json:"selector"`
}

func readExtensionsFamilyManifest(t testing.TB, root string) extensionTestFamilyManifest {
	t.Helper()
	return readTestFamilyManifest(t, root, "module.extensions.json")
}

func readTestFamilyManifest(t testing.TB, root string, name string) extensionTestFamilyManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "tools", "test_families", name))
	if err != nil {
		t.Fatalf("read test family manifest %s: %v", name, err)
	}
	var manifest extensionTestFamilyManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode test family manifest %s: %v", name, err)
	}
	return manifest
}

func readExtensionContractFile(t testing.TB, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func readGeneratedExtensionObject(t testing.TB, path string) map[string]any {
	t.Helper()
	artifact, ok := contractsgen.ExtensionArtifactsIndex[path]
	if !ok {
		t.Fatalf("generated Extensions artifact %q is not packaged", path)
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &object); err != nil {
		t.Fatalf("decode generated Extensions artifact %q: %v", path, err)
	}
	return object
}

func jsonEqual(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func extensionsRepoRoot(t testing.TB) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve Extensions contract test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
}

func requireExactStrings(t testing.TB, got, want []string, label string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v; want %v", label, got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("%s = %v; want %v", label, got, want)
		}
	}
}

func requireContains(t testing.TB, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	t.Fatalf("missing %s from %s", want, fmt.Sprint(sorted))
}
