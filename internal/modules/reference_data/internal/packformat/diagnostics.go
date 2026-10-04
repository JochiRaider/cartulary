package packformat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type Check struct {
	ID            string `json:"check_id"`
	Applicability string `json:"applicability"`
	Rank          int    `json:"rank"`
	Phase         string `json:"phase"`
	Code          string `json:"code"`
	Severity      string `json:"severity"`
}

var checks = loadChecks()

func loadChecks() []Check {
	var catalog struct {
		SchemaID string  `json:"schema_id"`
		Checks   []Check `json:"checks"`
	}
	decoder := json.NewDecoder(bytes.NewReader(projection("checks.v1.json")))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil || catalog.SchemaID != "cartulary.reference_pack_check_registry.v1" || len(catalog.Checks) == 0 {
		panic("invalid reference pack check registry")
	}
	seen := map[string]bool{}
	for i, c := range catalog.Checks {
		if c.Rank != i+1 || c.ID == "" || seen[c.ID] || c.Severity != "error" || !slices.Contains([]string{"attempt_bytes", "available_container", "admitted_archive", "present_bounded_manifest", "present_manifest", "admitted_manifest", "admitted_type_registry"}, c.Applicability) {
			panic("invalid reference pack check order")
		}
		seen[c.ID] = true
	}
	return catalog.Checks
}
func Checks() []Check { return slices.Clone(checks) }

type SafeDetails struct {
	LimitID            *string `json:"limit_id"`
	ExpectedToken      *string `json:"expected_token"`
	ActualToken        *string `json:"actual_token"`
	RelatedPackKey     *string `json:"related_pack_key"`
	RelatedPackVersion *string `json:"related_pack_version"`
}

// LimitFinding reports only a registered identifier and numeric measurements.
// For a streaming guard, observed is the smallest measured count proving the
// breach; callers must not continue parsing to determine the final total.
func LimitFinding(path, id string, maximum, observed int64) Finding {
	return unsignedLimitFinding(path, id, uint64(maximum), uint64(observed))
}

func unsignedLimitFinding(path, id string, maximum, observed uint64) Finding {
	expected, actual := strconv.FormatUint(maximum, 10), strconv.FormatUint(observed, 10)
	return Finding{Path: path, Details: SafeDetails{LimitID: &id, ExpectedToken: &expected, ActualToken: &actual}}
}

type Finding struct {
	Path    string
	EntryID *string
	Details SafeDetails
}
type Issue struct {
	ID         string      `json:"issue_id"`
	CheckID    string      `json:"check_id"`
	Severity   string      `json:"severity"`
	Phase      string      `json:"phase"`
	Code       string      `json:"code"`
	ReasonCode *string     `json:"reason_code"`
	Path       string      `json:"path"`
	EntryID    *string     `json:"entry_id"`
	Details    SafeDetails `json:"safe_details"`
}
type ValidationSummary struct {
	SchemaID       string  `json:"schema_id"`
	Result         string  `json:"result"`
	PrimaryIssueID *string `json:"primary_issue_id"`
	Truncated      bool    `json:"issues_truncated"`
	Total          int     `json:"total_issue_count"`
	Retained       int     `json:"retained_issue_count"`
	Issues         []Issue `json:"issues"`
}
type FindingSink func(Finding) error
type CheckFunc func(context.Context, FindingSink) error

// RunChecks executes a complete declared check program. Applicability belongs
// to the check implementation, not to optional caller callbacks. Missing, nil,
// or unknown checks are implementation errors. Operational errors produce no
// content verdict.
func RunChecks(ctx context.Context, scratch DiagnosticScratch, program map[string]CheckFunc) (*ValidationSummary, error) {
	if len(program) != len(checks) {
		return nil, errors.New("reference pack: incomplete check program")
	}
	for _, check := range checks {
		if run, exists := program[check.ID]; !exists || run == nil {
			return nil, errors.New("reference pack: unknown check program")
		}
	}
	for _, check := range checks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		run := program[check.ID]
		collector := newDiagnosticCollector(ctx, check, scratch)
		if err := run(ctx, collector.add); err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		summary, err := collector.finish()
		if err != nil || summary.Result == "failed" {
			return summary, err
		}
	}
	return successSummary(), nil
}

func successSummary() *ValidationSummary {
	return &ValidationSummary{SchemaID: "cartulary.reference_pack_validation_summary.v1", Result: "succeeded", Issues: []Issue{}}
}

// CheckSummary uses the registry's fixed mapping for one known failed check.
// Callers supply its entire enumerable finding stream, never a retained slice.
func CheckSummary(ctx context.Context, id string, scratch DiagnosticScratch, run CheckFunc) (*ValidationSummary, error) {
	if run == nil {
		return nil, errors.New("reference pack: missing diagnostic check")
	}
	for _, check := range checks {
		if check.ID != id {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		collector := newDiagnosticCollector(ctx, check, scratch)
		if err := run(ctx, collector.add); err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return collector.finish()
	}
	return nil, errors.New("reference pack: unknown diagnostic check")
}

func makeIssue(check Check, f Finding) (Issue, error) {
	if err := validateFinding(f); err != nil {
		return Issue{}, err
	}
	// Retained issues own their values. A streaming producer can safely reuse
	// its per-row findings without mutating the retained prefix or digest.
	f.EntryID = cloneDiagnosticString(f.EntryID)
	f.Details = SafeDetails{cloneDiagnosticString(f.Details.LimitID), cloneDiagnosticString(f.Details.ExpectedToken), cloneDiagnosticString(f.Details.ActualToken), cloneDiagnosticString(f.Details.RelatedPackKey), cloneDiagnosticString(f.Details.RelatedPackVersion)}
	preimage := struct {
		CheckID  string      `json:"check_id"`
		Severity string      `json:"severity"`
		Phase    string      `json:"phase"`
		Code     string      `json:"code"`
		Path     string      `json:"path"`
		EntryID  *string     `json:"entry_id"`
		Details  SafeDetails `json:"safe_details"`
	}{check.ID, check.Severity, check.Phase, check.Code, f.Path, f.EntryID, f.Details}
	raw, err := canonicaljson.Marshal(preimage)
	if err != nil {
		return Issue{}, err
	}
	id := "rpi_" + Digest(raw)
	reason := check.Code
	return Issue{ID: id, CheckID: check.ID, Severity: check.Severity, Phase: check.Phase, Code: check.Code, ReasonCode: &reason, Path: f.Path, EntryID: f.EntryID, Details: f.Details}, nil
}

func cloneDiagnosticString(value *string) *string {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func compareIssues(a, b Issue) int {
	if c := strings.Compare(a.Path, b.Path); c != 0 {
		return c
	}
	if a.EntryID == nil && b.EntryID != nil {
		return -1
	}
	if a.EntryID != nil && b.EntryID == nil {
		return 1
	}
	if a.EntryID != nil {
		if c := strings.Compare(*a.EntryID, *b.EntryID); c != 0 {
			return c
		}
	}
	return strings.Compare(a.ID, b.ID)
}

// runCheckRange is private to the format owner. Every check in the contiguous
// registry range must be implemented. Boolean readers and diagnostic attempts
// execute the same program; only diagnostic retention differs.
func runCheckRange(ctx context.Context, first, last string, scratch DiagnosticScratch, program map[string]CheckFunc, diagnostics bool) error {
	selected := []Check{}
	active := false
	for _, check := range checks {
		if check.ID == first {
			active = true
		}
		if active {
			selected = append(selected, check)
		}
		if check.ID == last && active {
			active = false
			break
		}
	}
	if active || len(selected) == 0 || selected[len(selected)-1].ID != last || len(selected) != len(program) {
		return errors.New("reference pack: incomplete check range")
	}
	for _, check := range selected {
		if program[check.ID] == nil {
			return errors.New("reference pack: missing range check")
		}
	}
	for _, check := range selected {
		if err := ctx.Err(); err != nil {
			return err
		}
		if diagnostics {
			summary, err := CheckSummary(ctx, check.ID, scratch, program[check.ID])
			if err != nil {
				return err
			}
			if summary.Result == "failed" {
				return &Failure{Code: check.Code, CheckID: check.ID, Summary: summary}
			}
		} else {
			err := program[check.ID](ctx, func(Finding) error { return errShapeMismatch })
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if errors.Is(err, errShapeMismatch) {
				return &Failure{Code: check.Code, CheckID: check.ID}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeferredChecks binds a complete contiguous owner range before execution while
// constructing its dependent state only when the first rank is reached. A
// missing callback is an implementation failure, never an applicability rule.
func DeferredChecks(first, last string, build func(context.Context) (map[string]CheckFunc, error)) (map[string]CheckFunc, error) {
	if build == nil {
		return nil, errors.New("reference pack: missing check builder")
	}
	selected := []string{}
	active := false
	for _, check := range checks {
		if check.ID == first {
			active = true
		}
		if active {
			selected = append(selected, check.ID)
		}
		if active && check.ID == last {
			active = false
			break
		}
	}
	if active || len(selected) == 0 || selected[len(selected)-1] != last {
		return nil, errors.New("reference pack: unknown deferred check range")
	}
	var built map[string]CheckFunc
	program := map[string]CheckFunc{}
	for index, id := range selected {
		program[id] = func(ctx context.Context, emit FindingSink) error {
			if built == nil {
				if index != 0 {
					return errors.New("reference pack: deferred range executed out of order")
				}
				var err error
				built, err = build(ctx)
				if err != nil {
					return err
				}
				if len(built) != len(selected) {
					return errors.New("reference pack: incomplete deferred check range")
				}
				for _, required := range selected {
					if built[required] == nil {
						return errors.New("reference pack: missing deferred check")
					}
				}
			}
			return built[id](ctx, emit)
		}
	}
	return program, nil
}

// ComposeChecks rejects overlapping ownership as well as missing callbacks.
// RunChecks then requires coverage of the entire registry before running it.
func ComposeChecks(parts ...map[string]CheckFunc) (map[string]CheckFunc, error) {
	result := map[string]CheckFunc{}
	for _, part := range parts {
		for id, run := range part {
			if run == nil || result[id] != nil {
				return nil, errors.New("reference pack: conflicting check ownership")
			}
			result[id] = run
		}
	}
	return result, nil
}
