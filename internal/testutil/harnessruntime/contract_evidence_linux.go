//go:build linux

package harnessruntime

import (
	"bytes"
	"cmp"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/gen/openapioperations"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/runevidence"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

// TH-HARNESS-REQ-860 bounds. A stream beyond them is not silently truncated:
// the unread remainder is malformed evidence.
const (
	contractEvidenceFileLimit      = 4096
	contractEvidenceFileBytes      = 16 << 20
	contractEvidenceViolationLimit = 32
)

var contractEvidenceFileName = regexp.MustCompile(`^[1-9][0-9]{0,9}\.jsonl$`)

// TH-HARNESS-REQ-862 declared-versus-exercised error-code coverage report,
// published in the run root whenever the stream holds an emission.
const (
	errorCodeCoverageFile     = "error-code-coverage.json"
	errorCodeCoverageSchemaID = "cartulary.harness_error_code_coverage.v1"
)

// contractViolation is one distinct finding: a guard violation keyed by its
// operation, status and fact, or a server fatal keyed by its reason code and
// process kind (TH-HARNESS-REQ-863).
type contractViolation struct {
	kind, operation, code, declared, observed string
	reason, process                           string
	status                                    int
	occurrences                               int
}

// errorEmission is one distinct operation, status and code some process of
// the run emitted (TH-HARNESS-REQ-862).
type errorEmission struct {
	operation string
	status    int
	code      string
}

// coverageCatalog is the generated operation catalog a coverage report is
// computed against.
type coverageCatalog struct {
	operations      []openapioperations.Operation
	sha256, version string
}

func generatedCoverageCatalog() coverageCatalog {
	return coverageCatalog{operations: openapioperations.All(), sha256: openapioperations.CanonicalSHA256, version: openapioperations.DocumentVersion}
}

// collateContractEvidence validates every record under the run root's contract
// evidence stream after all writers have settled. It returns the bounded
// run-summary section and the gate failure: malformed evidence is an artifact
// failure; any violation is a product contract violation; any server fatal is
// a product server_fatal failure. An absent or empty stream passes. When the stream holds error emissions it also publishes the
// error-code coverage report against the generated operation catalog; coverage
// never fails the run.
func collateContractEvidence(runRoot string) (map[string]any, *failure) {
	return collateContractEvidenceAgainst(runRoot, generatedCoverageCatalog)
}

func collateContractEvidenceAgainst(runRoot string, catalog func() coverageCatalog) (map[string]any, *failure) {
	counts := map[string]int{string(runevidence.UndeclaredErrorCode): 0, string(runevidence.SecurityMismatch): 0, string(runevidence.ServerFatal): 0, string(runevidence.ErrorEmission): 0}
	violations := map[string]*contractViolation{}
	emissions := map[errorEmission]struct{}{}
	malformed := 0
	err := readContractEvidence(filepath.Join(runRoot, runevidence.Directory), func(pid int, record map[string]any) {
		kind := record["kind"].(string)
		counts[kind]++
		if kind == string(runevidence.ErrorEmission) {
			// Each process writes a distinct emission once; processes repeat
			// one another, so the report counts each emission once per run.
			emissions[errorEmission{operation: record["operation_id"].(string), status: int(record["status"].(float64)), code: record["code"].(string)}] = struct{}{}
			return
		}
		v := contractViolation{kind: kind}
		if kind == string(runevidence.ServerFatal) {
			v.reason, v.process = record["reason_code"].(string), record["process_kind"].(string)
		} else {
			v.operation, v.status = record["operation_id"].(string), int(record["status"].(float64))
		}
		if code, ok := record["code"].(string); ok {
			v.code = code
		}
		if security, ok := record["security"].(map[string]any); ok {
			v.declared, v.observed = security["declared"].(string), security["observed"].(string)
		}
		key := v.kind + "\x00" + v.operation + "\x00" + strconv.Itoa(v.status) + "\x00" + v.code + "\x00" + v.declared + "\x00" + v.observed + "\x00" + v.reason + "\x00" + v.process
		if existing := violations[key]; existing != nil {
			existing.occurrences++
			return
		}
		v.occurrences = 1
		violations[key] = &v
	}, func() { malformed++ })
	if err != nil && malformed == 0 {
		malformed = 1
	}
	distinct := make([]*contractViolation, 0, len(violations))
	fatal := false
	for _, v := range violations {
		distinct = append(distinct, v)
		fatal = fatal || v.kind == string(runevidence.ServerFatal)
	}
	// Server fatals sort first so the bounded list never omits the cause of the
	// run's failure; the rest keep their deterministic order.
	slices.SortFunc(distinct, func(a, b *contractViolation) int {
		return cmp.Or(-cmp.Compare(boolRank(a.kind == string(runevidence.ServerFatal)), boolRank(b.kind == string(runevidence.ServerFatal))),
			cmp.Compare(a.kind, b.kind), cmp.Compare(a.reason, b.reason), cmp.Compare(a.process, b.process),
			cmp.Compare(a.operation, b.operation), cmp.Compare(a.status, b.status), cmp.Compare(a.code, b.code), cmp.Compare(a.declared, b.declared), cmp.Compare(a.observed, b.observed))
	})
	listed := []any{}
	for _, v := range distinct[:min(len(distinct), contractEvidenceViolationLimit)] {
		entry := map[string]any{"kind": v.kind, "occurrences": v.occurrences}
		switch v.kind {
		case string(runevidence.ServerFatal):
			entry["reason_code"], entry["process_kind"] = v.reason, v.process
		case string(runevidence.SecurityMismatch):
			entry["operation_id"], entry["status"] = v.operation, v.status
			entry["security"] = map[string]any{"declared": v.declared, "observed": v.observed}
		default:
			entry["operation_id"], entry["status"] = v.operation, v.status
			entry["code"] = v.code
		}
		listed = append(listed, entry)
	}
	section := map[string]any{
		"schema_id":               runevidence.SchemaID,
		"record_counts":           counts,
		"malformed_record_count":  malformed,
		"violations":              listed,
		"omitted_violation_count": len(distinct) - len(listed),
	}
	if len(emissions) > 0 {
		report, totals := errorCodeCoverage(catalog(), emissions)
		summary := map[string]any{"report": nil}
		for key, value := range totals {
			summary[key] = value
		}
		// A report that cannot be published leaves the totals without a
		// reference; coverage is evidence, never a gate (TH-HARNESS-REQ-862).
		if writeEvidence(runRoot, errorCodeCoverageFile, errorCodeCoverageSchemaID, report, true) == nil {
			summary["report"] = errorCodeCoverageFile
		}
		section["error_code_coverage"] = summary
	}
	// Precedence: a server fatal outranks a contract violation. Once a test
	// server is fatal it answers every later request with 503
	// service_unavailable, so findings recorded around it are consequences and
	// the fatal is the cause. Both stay listed. Either outranks malformed
	// evidence, which remains counted.
	var gate *failure
	if fatal {
		gate = failed("product", "server_fatal")
	} else if len(distinct) > 0 {
		gate = failed("product", "contract_violation")
	} else if malformed > 0 {
		gate = failed("artifact", "artifact_error")
		gate.Artifact = runevidence.Directory
	}
	return section, gate
}

func boolRank(value bool) int {
	if value {
		return 1
	}
	return 0
}

// errorCodeCoverage lists, for each operation and error status of the
// catalog, the declared codes, the codes the run exercised and the declared
// codes it did not exercise. An emission at a status the catalog does not
// declare for its operation gets its own entry with no declared codes. An
// exercised code the operation does not declare is also listed as
// undeclared_exercised, internal_error included: every operation declares it
// at 500 (Core 01 REQ-01-702). The guard has already failed the run for each
// such undeclared code.
func errorCodeCoverage(catalog coverageCatalog, emissions map[errorEmission]struct{}) (map[string]any, map[string]any) {
	type response struct {
		operation string
		status    int
	}
	declared := map[response][]string{}
	for _, operation := range catalog.operations {
		for status, codes := range operation.ErrorCodes {
			declared[response{operation.OperationID, status}] = codes
		}
	}
	exercised := map[response][]string{}
	for emission := range emissions {
		key := response{emission.operation, emission.status}
		exercised[key] = append(exercised[key], emission.code)
		if _, known := declared[key]; !known {
			declared[key] = nil
		}
	}
	keys := make([]response, 0, len(declared))
	for key := range declared {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b response) int {
		return cmp.Or(cmp.Compare(a.operation, b.operation), cmp.Compare(a.status, b.status))
	})
	totals := map[string]any{"operation_statuses": len(keys)}
	var declaredCount, exercisedDeclared, unexercisedCount, undeclaredCount int
	responses := make([]any, 0, len(keys))
	for _, key := range keys {
		declaredCodes := slices.Sorted(slices.Values(declared[key]))
		exercisedCodes := slices.Sorted(slices.Values(exercised[key]))
		unexercised, undeclared := []string{}, []string{}
		for _, code := range declaredCodes {
			if !slices.Contains(exercisedCodes, code) {
				unexercised = append(unexercised, code)
			}
		}
		for _, code := range exercisedCodes {
			if !slices.Contains(declaredCodes, code) {
				undeclared = append(undeclared, code)
			}
		}
		declaredCount += len(declaredCodes)
		unexercisedCount += len(unexercised)
		exercisedDeclared += len(declaredCodes) - len(unexercised)
		undeclaredCount += len(undeclared)
		responses = append(responses, map[string]any{
			"operation_id": key.operation, "status": key.status,
			"declared": append([]string{}, declaredCodes...), "exercised": append([]string{}, exercisedCodes...),
			"unexercised": unexercised, "undeclared_exercised": undeclared,
		})
	}
	totals["declared_codes"] = declaredCount
	totals["exercised_declared_codes"] = exercisedDeclared
	totals["unexercised_codes"] = unexercisedCount
	totals["undeclared_exercised_codes"] = undeclaredCount
	report := map[string]any{
		"schema_id": errorCodeCoverageSchemaID, "catalog_sha256": catalog.sha256, "document_version": catalog.version,
		"responses": responses, "totals": totals,
	}
	return report, totals
}

// readContractEvidence visits each schema-valid record. A malformed line, an
// unexpected entry, or a record whose process identity disagrees with its file
// is reported through malformed and never counted as passing evidence.
func readContractEvidence(directory string, visit func(int, map[string]any), malformed func()) error {
	parent, err := harnesslocal.OpenDirectory(directory, true)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer parent.Close()
	names, err := parent.Readdirnames(contractEvidenceFileLimit + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if len(names) > contractEvidenceFileLimit {
		return harnesslocal.ErrUnsafe
	}
	slices.Sort(names)
	for _, name := range names {
		if !contractEvidenceFileName.MatchString(name) {
			malformed()
			continue
		}
		pid, _ := strconv.Atoi(name[:len(name)-len(".jsonl")])
		data, err := harnesslocal.ReadFileAt(parent, name, contractEvidenceFileBytes, false)
		if err != nil {
			malformed()
			continue
		}
		if len(data) > 0 && data[len(data)-1] != '\n' {
			// A torn final append is unverifiable evidence, not an absent record.
			malformed()
			data = data[:bytes.LastIndexByte(data, '\n')+1]
		}
		if len(data) == 0 {
			continue
		}
		for line := range bytes.SplitSeq(bytes.TrimSuffix(data, []byte("\n")), []byte("\n")) {
			value, err := harnessschema.Decode(line)
			record, ok := value.(map[string]any)
			if err != nil || !ok || harnessschema.Validate(runevidence.SchemaID, value) != nil || record["pid"] != float64(pid) {
				malformed()
				continue
			}
			visit(pid, record)
		}
	}
	return nil
}
