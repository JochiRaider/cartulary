//go:build linux

package harnessruntime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/openapioperations"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

func contractRecord(kind, operation string, status int, code string, pid int) string {
	return fmt.Sprintf(`{"schema_id":"cartulary.harness_contract_evidence.v1","kind":%q,"operation_id":%q,"status":%d,"code":%q,"process_kind":"in_process","work_unit_id":null,"pid":%d}`, kind, operation, status, code, pid)
}

func fatalRecord(reason, process string, pid int) string {
	return fmt.Sprintf(`{"schema_id":"cartulary.harness_contract_evidence.v1","kind":"server_fatal","reason_code":%q,"process_kind":%q,"work_unit_id":null,"pid":%d}`, reason, process, pid)
}

func writeContractStream(t *testing.T, root string, files map[string]string) {
	t.Helper()
	directory := filepath.Join(root, "contract-evidence")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func privateRunRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestRuntimeContractEvidenceGate covers TH-HARNESS-REQ-860, REQ-863, AC-154
// and AC-158: a violation fails the run as product/contract_violation and is
// listed, a server fatal fails it as product/server_fatal and is listed first,
// a malformed record fails as artifact, and an absent, empty or coverage-only
// stream passes.
func TestRuntimeContractEvidenceGate(t *testing.T) {
	violation := contractRecord("undeclared_error_code", "createIncident", 409, "record_locked", 41)
	emission := contractRecord("error_emission", "createIncident", 404, "not_found", 41)
	fatal := fatalRecord("application_process_lease_lost", "in_process", 41)
	for _, tc := range []struct {
		name      string
		files     map[string]string
		absent    bool
		class     string
		reason    string
		emissions int
		fatals    int
		listed    int
		omitted   int
		malformed int
	}{
		{name: "absent stream passes", absent: true},
		{name: "empty stream passes", files: map[string]string{}},
		{name: "empty process file passes", files: map[string]string{"41.jsonl": ""}},
		{name: "coverage records are not violations", files: map[string]string{"41.jsonl": emission + "\n" + emission + "\n"}, emissions: 2},
		{name: "violation fails and is listed once", files: map[string]string{"41.jsonl": violation + "\n" + emission + "\n" + violation + "\n"}, class: "product", reason: "contract_violation", emissions: 1, listed: 1},
		{name: "malformed json fails as artifact", files: map[string]string{"41.jsonl": "{not json}\n"}, class: "artifact", reason: "artifact_error", malformed: 1},
		{name: "schema-invalid record fails as artifact", files: map[string]string{"41.jsonl": strings.Replace(emission, `"pid":41`, `"pid":41,"authorization":"Bearer sentinel"`, 1) + "\n"}, class: "artifact", reason: "artifact_error", malformed: 1},
		{name: "foreign process identity fails as artifact", files: map[string]string{"42.jsonl": emission + "\n"}, class: "artifact", reason: "artifact_error", malformed: 1},
		{name: "torn append fails as artifact", files: map[string]string{"41.jsonl": emission + "\n" + emission[:20]}, class: "artifact", reason: "artifact_error", emissions: 1, malformed: 1},
		{name: "unexpected stream entry fails as artifact", files: map[string]string{"notes.txt": "x\n"}, class: "artifact", reason: "artifact_error", malformed: 1},
		{name: "violation outranks malformed evidence", files: map[string]string{"41.jsonl": violation + "\n{}\n"}, class: "product", reason: "contract_violation", listed: 1, malformed: 1},
		{name: "server fatal fails and is listed once", files: map[string]string{"41.jsonl": emission + "\n" + fatal + "\n", "42.jsonl": fatalRecord("application_process_lease_lost", "in_process", 42) + "\n"}, class: "product", reason: "server_fatal", emissions: 1, fatals: 2, listed: 1},
		{name: "server fatals of each process kind are listed apart", files: map[string]string{"41.jsonl": fatal + "\n", "42.jsonl": fatalRecord("application_process_lease_lost", "server_harness", 42) + "\n"}, class: "product", reason: "server_fatal", fatals: 2, listed: 2},
		{name: "server fatal outranks a contract violation", files: map[string]string{"41.jsonl": violation + "\n" + fatal + "\n"}, class: "product", reason: "server_fatal", fatals: 1, listed: 2},
		{name: "server fatal outranks malformed evidence", files: map[string]string{"41.jsonl": fatal + "\n{}\n"}, class: "product", reason: "server_fatal", fatals: 1, listed: 1, malformed: 1},
		{name: "server fatal without its reason fails as artifact", files: map[string]string{"41.jsonl": strings.Replace(fatal, `"reason_code":"application_process_lease_lost",`, "", 1) + "\n"}, class: "artifact", reason: "artifact_error", malformed: 1},
		{name: "server fatal with free text fails as artifact", files: map[string]string{"41.jsonl": fatalRecord("lease lost for Bearer sentinel", "in_process", 41) + "\n"}, class: "artifact", reason: "artifact_error", malformed: 1},
		{name: "server fatal naming an operation fails as artifact", files: map[string]string{"41.jsonl": strings.Replace(fatal, `"kind"`, `"operation_id":"createIncident","status":503,"kind"`, 1) + "\n"}, class: "artifact", reason: "artifact_error", malformed: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := privateRunRoot(t)
			if !tc.absent {
				writeContractStream(t, root, tc.files)
			}
			section, gate := collateContractEvidence(root)
			if (gate == nil) != (tc.class == "") || gate != nil && (gate.Class != tc.class || gate.Reason != tc.reason) {
				t.Fatalf("gate = %v, want %s/%s", gate, tc.class, tc.reason)
			}
			counts := section["record_counts"].(map[string]int)
			if counts["error_emission"] != tc.emissions || counts["server_fatal"] != tc.fatals || len(section["violations"].([]any)) != tc.listed || section["omitted_violation_count"] != tc.omitted || section["malformed_record_count"] != tc.malformed {
				t.Fatalf("section = %#v", section)
			}
			if listed := section["violations"].([]any); tc.fatals > 0 && len(listed) > 0 {
				first := listed[0].(map[string]any)
				if first["kind"] != "server_fatal" || first["reason_code"] != "application_process_lease_lost" || first["operation_id"] != nil {
					t.Fatalf("server fatal is not listed first: %#v", listed)
				}
			}
		})
	}
	t.Run("distinct violations are bounded", func(t *testing.T) {
		root := privateRunRoot(t)
		var lines strings.Builder
		for i := range contractEvidenceViolationLimit + 8 {
			lines.WriteString(contractRecord("undeclared_error_code", fmt.Sprintf("operation%02d", i), 500, "internal_error", 7) + "\n")
		}
		writeContractStream(t, root, map[string]string{"7.jsonl": lines.String()})
		section, gate := collateContractEvidence(root)
		if gate == nil || gate.Reason != "contract_violation" || len(section["violations"].([]any)) != contractEvidenceViolationLimit || section["omitted_violation_count"] != 8 {
			t.Fatalf("unbounded violation list: %v %#v", gate, section)
		}
	})
	t.Run("a server fatal is never omitted by the bound", func(t *testing.T) {
		root := privateRunRoot(t)
		var lines strings.Builder
		for i := range contractEvidenceViolationLimit + 8 {
			lines.WriteString(contractRecord("undeclared_error_code", fmt.Sprintf("operation%02d", i), 500, "internal_error", 7) + "\n")
		}
		lines.WriteString(fatalRecord("published_component_lost", "server_harness", 7) + "\n")
		writeContractStream(t, root, map[string]string{"7.jsonl": lines.String()})
		section, gate := collateContractEvidence(root)
		listed := section["violations"].([]any)
		if gate == nil || gate.Reason != "server_fatal" || len(listed) != contractEvidenceViolationLimit || section["omitted_violation_count"] != 9 || listed[0].(map[string]any)["reason_code"] != "published_component_lost" {
			t.Fatalf("bounded list omitted the server fatal: %v %#v", gate, section)
		}
	})
	t.Run("unsafe stream directory fails as artifact", func(t *testing.T) {
		root := privateRunRoot(t)
		writeContractStream(t, root, map[string]string{"41.jsonl": violation + "\n"})
		if err := os.Chmod(filepath.Join(root, "contract-evidence"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, gate := collateContractEvidence(root); gate == nil || gate.Class != "artifact" {
			t.Fatalf("unsafe directory accepted: %v", gate)
		}
	})
	t.Run("run gate fails the published run and lists the violation", func(t *testing.T) {
		p := planFixture(t)
		root := privateRunRoot(t)
		writeContractStream(t, root, map[string]string{"41.jsonl": violation + "\n" + violation + "\n"})
		publisher, err := newGraphPublication(p, root)
		if err != nil {
			t.Fatal(err)
		}
		defer publisher.Abort()
		result, err := executeGraph(context.Background(), p, enginePorts{Event: publisher.Event, Terminal: publisher.Terminal,
			Execute:          func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} },
			ContractEvidence: func() (map[string]any, *failure) { return collateContractEvidence(root) }})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != "failed" || result.Failure == nil || result.Failure.Class != "product" || result.Failure.Reason != "contract_violation" || failureExitCode(result.Failure, "") != 10 {
			t.Fatalf("violation did not fail the run: %#v", result.Failure)
		}
		summary, err := publisher.Finish(result)
		if err != nil || summary["status"] != "fail" || summary["failure_reason"] != "contract_violation" {
			t.Fatalf("published summary = %v, %v", summary, err)
		}
		data, err := os.ReadFile(filepath.Join(root, "run-summary.json"))
		if err != nil {
			t.Fatal(err)
		}
		value, err := harnessschema.Decode(data)
		if err != nil || harnessschema.Validate("cartulary.harness_run_summary.v1", value) != nil {
			t.Fatal("invalid run summary", err)
		}
		listed := value.(map[string]any)["contract_evidence"].(map[string]any)["violations"].([]any)
		entry := listed[0].(map[string]any)
		if len(listed) != 1 || entry["operation_id"] != "createIncident" || entry["code"] != "record_locked" || entry["occurrences"] != float64(2) {
			t.Fatalf("violation list = %#v", listed)
		}
	})
	t.Run("server fatal fails the run while the failing test keeps its own failure", func(t *testing.T) {
		p := planFixture(t)
		root := privateRunRoot(t)
		writeContractStream(t, root, map[string]string{"41.jsonl": violation + "\n" + fatal + "\n"})
		publisher, err := newGraphPublication(p, root)
		if err != nil {
			t.Fatal(err)
		}
		defer publisher.Abort()
		// The unit fails on the 503 the fatal server sent.
		failing := p.facts.Graph.Units[0].ID
		result, err := executeGraph(context.Background(), p, enginePorts{Event: publisher.Event, Terminal: publisher.Terminal,
			Execute: func(_ context.Context, request executionRequest) unitResult {
				if request.Unit.ID == failing {
					return unitResult{Status: "failed", Failure: failed("product", "test_assertion_failure")}
				}
				return unitResult{Status: "passed"}
			},
			ContractEvidence: func() (map[string]any, *failure) { return collateContractEvidence(root) }})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != "failed" || result.Failure == nil || result.Failure.Class != "product" || result.Failure.Reason != "server_fatal" || failureExitCode(result.Failure, "") != 10 {
			t.Fatalf("server fatal did not fail the run: %#v", result.Failure)
		}
		if unit := result.Units[failing]; unit.Failure == nil || unit.Failure.Reason != "test_assertion_failure" {
			t.Fatalf("failing test lost its own failure: %#v", unit.Failure)
		}
		summary, err := publisher.Finish(result)
		if err != nil || summary["status"] != "fail" || summary["failure_reason"] != "server_fatal" {
			t.Fatalf("published summary = %v, %v", summary, err)
		}
		data, err := os.ReadFile(filepath.Join(root, "run-summary.json"))
		if err != nil {
			t.Fatal(err)
		}
		value, err := harnessschema.Decode(data)
		if err != nil || harnessschema.Validate("cartulary.harness_run_summary.v1", value) != nil {
			t.Fatal("invalid run summary", err)
		}
		evidence := value.(map[string]any)["contract_evidence"].(map[string]any)
		listed := evidence["violations"].([]any)
		entry := listed[0].(map[string]any)
		if len(listed) != 2 || entry["kind"] != "server_fatal" || entry["reason_code"] != "application_process_lease_lost" || entry["process_kind"] != "in_process" || entry["occurrences"] != float64(1) || evidence["record_counts"].(map[string]any)["server_fatal"] != float64(1) {
			t.Fatalf("server fatal list = %#v", evidence)
		}
	})
	t.Run("a run without a fatal is unaffected", func(t *testing.T) {
		root := privateRunRoot(t)
		writeContractStream(t, root, map[string]string{"41.jsonl": emission + "\n"})
		section, gate := collateContractEvidence(root)
		if gate != nil || section["record_counts"].(map[string]int)["server_fatal"] != 0 {
			t.Fatalf("run without a fatal failed: %v %#v", gate, section)
		}
	})
}

// TestRuntimeErrorCodeCoverageReport covers TH-HARNESS-REQ-862 and
// TH-HARNESS-AC-157: the run gate publishes error-code-coverage.json, listing
// declared, exercised and unexercised codes for each operation and error
// status of the catalog, counting an emission repeated across processes once,
// listing an undeclared code as undeclared_exercised (internal_error included,
// since every operation declares it),
// and carrying the totals in the run summary. Unexercised declarations never
// fail the run, and a run without emissions publishes no report.
func TestRuntimeErrorCodeCoverageReport(t *testing.T) {
	catalog := func() coverageCatalog {
		return coverageCatalog{sha256: strings.Repeat("a", 64), version: "9.0.0", operations: []openapioperations.Operation{
			{OperationID: "createThing", ErrorCodes: map[int][]string{
				400: {"invalid_thing"}, 409: {"credential_bootstrap_rejected", "beta_conflict", "alpha_conflict"}, 500: {"internal_error"},
			}},
			{OperationID: "getThing", ErrorCodes: map[int][]string{404: {"thing_not_found"}}},
		}}
	}
	emission := func(operation string, status int, code string, pid int) string {
		return contractRecord("error_emission", operation, status, code, pid) + "\n"
	}
	readReport := func(t *testing.T, root string) map[string]any {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, "error-code-coverage.json"))
		if err != nil {
			t.Fatal(err)
		}
		value, err := harnessschema.Decode(data)
		if err != nil || harnessschema.Validate("cartulary.harness_error_code_coverage.v1", value) != nil {
			t.Fatalf("report is not cartulary.harness_error_code_coverage.v1: %s", data)
		}
		return value.(map[string]any)
	}
	t.Run("declared, exercised and unexercised codes per operation and status", func(t *testing.T) {
		root := privateRunRoot(t)
		writeContractStream(t, root, map[string]string{
			"41.jsonl": emission("createThing", 409, "alpha_conflict", 41) + emission("createThing", 400, "invalid_thing", 41) +
				emission("getThing", 500, "internal_error", 41) + "{torn",
			"42.jsonl": emission("createThing", 409, "alpha_conflict", 42) + emission("getThing", 404, "thing_not_found", 42) +
				emission("createThing", 409, "gamma_conflict", 42) + contractRecord("undeclared_error_code", "createThing", 409, "gamma_conflict", 42) + "\n" +
				emission("createThing", 422, "unknown_status", 42),
		})
		section, gate := collateContractEvidenceAgainst(root, catalog)
		if gate == nil || gate.Class != "product" || gate.Reason != "contract_violation" {
			t.Fatalf("gate = %v, want only the undeclared code violation", gate)
		}
		report := readReport(t, root)
		if report["catalog_sha256"] != strings.Repeat("a", 64) || report["document_version"] != "9.0.0" {
			t.Fatalf("catalog identity = %v %v", report["catalog_sha256"], report["document_version"])
		}
		type entry struct{ declared, exercised, unexercised, undeclared string }
		want := map[string]entry{
			"createThing 400": {"invalid_thing", "invalid_thing", "", ""},
			"createThing 409": {"alpha_conflict,beta_conflict,credential_bootstrap_rejected", "alpha_conflict,gamma_conflict", "beta_conflict,credential_bootstrap_rejected", "gamma_conflict"},
			"createThing 422": {"", "unknown_status", "", "unknown_status"},
			"createThing 500": {"internal_error", "", "internal_error", ""},
			"getThing 404":    {"thing_not_found", "thing_not_found", "", ""},
			"getThing 500":    {"", "internal_error", "", "internal_error"},
		}
		join := func(value any) string {
			codes := []string{}
			for _, code := range value.([]any) {
				codes = append(codes, code.(string))
			}
			return strings.Join(codes, ",")
		}
		responses := report["responses"].([]any)
		order := []string{}
		for _, raw := range responses {
			response := raw.(map[string]any)
			key := fmt.Sprintf("%s %v", response["operation_id"], response["status"])
			order = append(order, key)
			got := entry{join(response["declared"]), join(response["exercised"]), join(response["unexercised"]), join(response["undeclared_exercised"])}
			if got != want[key] {
				t.Fatalf("%s = %+v, want %+v", key, got, want[key])
			}
		}
		if strings.Join(order, ";") != "createThing 400;createThing 409;createThing 422;createThing 500;getThing 404;getThing 500" {
			t.Fatalf("responses are not one per operation and status in order: %v", order)
		}
		totals := map[string]any{"operation_statuses": 6, "declared_codes": 6, "exercised_declared_codes": 3, "unexercised_codes": 3, "undeclared_exercised_codes": 3}
		for key, value := range totals {
			if report["totals"].(map[string]any)[key] != float64(value.(int)) {
				t.Fatalf("report totals = %v, want %v", report["totals"], totals)
			}
		}
		summary, _ := section["error_code_coverage"].(map[string]any)
		totals["report"] = "error-code-coverage.json"
		if len(summary) != len(totals) {
			t.Fatalf("summary coverage = %v, want %v", summary, totals)
		}
		for key, value := range totals {
			if summary[key] != value {
				t.Fatalf("summary coverage = %v, want %v", summary, totals)
			}
		}
		if counts := section["record_counts"].(map[string]int); counts["error_emission"] != 7 || section["malformed_record_count"] != 1 {
			t.Fatalf("section = %#v", section)
		}
	})
	t.Run("unexercised declarations never fail the run", func(t *testing.T) {
		root := privateRunRoot(t)
		writeContractStream(t, root, map[string]string{"41.jsonl": emission("getThing", 404, "thing_not_found", 41)})
		section, gate := collateContractEvidenceAgainst(root, catalog)
		if gate != nil {
			t.Fatalf("coverage failed the run: %v", gate)
		}
		summary, _ := section["error_code_coverage"].(map[string]any)
		if readReport(t, root)["totals"].(map[string]any)["unexercised_codes"] != float64(5) || summary["unexercised_codes"] != 5 {
			t.Fatalf("section = %#v", section)
		}
	})
	t.Run("a run without emissions publishes no report", func(t *testing.T) {
		for name, files := range map[string]map[string]string{
			"absent stream":  nil,
			"violation only": {"41.jsonl": contractRecord("undeclared_error_code", "createThing", 409, "gamma_conflict", 41) + "\n"},
			"malformed only": {"41.jsonl": strings.Replace(emission("getThing", 404, "thing_not_found", 41), `"pid":41`, `"pid":41,"query":"x"`, 1)},
		} {
			t.Run(name, func(t *testing.T) {
				root := privateRunRoot(t)
				if files != nil {
					writeContractStream(t, root, files)
				}
				section, _ := collateContractEvidenceAgainst(root, catalog)
				if _, present := section["error_code_coverage"]; present {
					t.Fatalf("summary carries coverage without emissions: %#v", section)
				}
				if _, err := os.Lstat(filepath.Join(root, "error-code-coverage.json")); !os.IsNotExist(err) {
					t.Fatalf("report published without emissions: %v", err)
				}
			})
		}
	})
	t.Run("the published run summary carries the totals", func(t *testing.T) {
		p := planFixture(t)
		root := privateRunRoot(t)
		writeContractStream(t, root, map[string]string{"41.jsonl": emission("getThing", 404, "thing_not_found", 41)})
		publisher, err := newGraphPublication(p, root)
		if err != nil {
			t.Fatal(err)
		}
		defer publisher.Abort()
		result, err := executeGraph(context.Background(), p, enginePorts{Event: publisher.Event, Terminal: publisher.Terminal,
			Execute:          func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} },
			ContractEvidence: func() (map[string]any, *failure) { return collateContractEvidenceAgainst(root, catalog) }})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != "passed" || result.Failure != nil {
			t.Fatalf("coverage changed the run result: %#v", result.Failure)
		}
		if _, err := publisher.Finish(result); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(root, "run-summary.json"))
		if err != nil {
			t.Fatal(err)
		}
		value, err := harnessschema.Decode(data)
		if err != nil || harnessschema.Validate("cartulary.harness_run_summary.v1", value) != nil {
			t.Fatal("invalid run summary", err)
		}
		evidence, _ := value.(map[string]any)["contract_evidence"].(map[string]any)
		coverage, _ := evidence["error_code_coverage"].(map[string]any)
		if coverage["report"] != "error-code-coverage.json" || coverage["exercised_declared_codes"] != float64(1) || coverage["declared_codes"] != float64(6) {
			t.Fatalf("summary coverage = %#v", coverage)
		}
	})
}
