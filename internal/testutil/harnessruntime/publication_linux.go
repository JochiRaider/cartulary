//go:build linux

package harnessruntime

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

type graphPublication struct {
	mu        sync.Mutex
	plan      *Plan
	root      string
	facts     *canonicalFacts
	events    *harnesslocal.Stream
	Forbidden func() ([]string, error)
	closed    bool
}

func newGraphPublication(plan *Plan, runRoot string) (*graphPublication, error) {
	directory, err := harnesslocal.PrivateDirectory(runRoot)
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	manifest, err := documentBytes("cartulary.harness_run_manifest.v2", plan.facts.Manifest)
	if err != nil {
		return nil, err
	}
	if err := harnesslocal.PublishNew(filepath.Join(runRoot, "run-manifest.json"), append(manifest, '\n')); err != nil {
		return nil, err
	}
	if err := writeEvidence(runRoot, "execution-context.json", "cartulary.harness_execution_context.v1", plan.context, false); err != nil {
		return nil, err
	}
	stream, err := harnesslocal.NewStream(filepath.Join(runRoot, "unit-events.ndjson"), 64*1024)
	if err != nil {
		return nil, err
	}
	return &graphPublication{plan: plan, root: runRoot, facts: newCanonicalFacts(plan), events: stream}, nil
}
func (p *graphPublication) Event(record map[string]any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return harnesslocal.ErrUnsafe
	}
	if err := p.facts.consume(record); err != nil {
		return err
	}
	bytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return p.events.Write(append(bytes, '\n'))
}
func (p *graphPublication) Prefix() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return "", harnesslocal.ErrUnsafe
	}
	prefix, err := p.events.Prefix()
	if err != nil {
		return "", err
	}
	data, err := documentBytes("cartulary.harness_unit_event_prefix.v1", map[string]any{
		"schema_id": "cartulary.harness_unit_event_prefix.v1", "run_id": p.plan.facts.Manifest["run_id"],
		"file": prefix.File, "device": prefix.Device, "inode": prefix.Inode,
		"byte_length": prefix.ByteLength, "last_seq": p.facts.seq,
	})
	return string(data), err
}
func writeEvidence(root, relative, schema string, value any, replace bool) error {
	if !safeRelative(relative) {
		return harnesslocal.ErrUnsafe
	}
	directory, err := harnesslocal.PrivateDirectory(filepath.Dir(filepath.Join(root, relative)))
	if err != nil {
		return err
	}
	directory.Close()
	bytes, err := documentBytes(schema, value)
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	if replace {
		return harnesslocal.ReplaceFile(filepath.Join(root, relative), bytes)
	}
	return harnesslocal.PublishNew(filepath.Join(root, relative), bytes)
}

var safeUnitName = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)
var secretAssignment = regexp.MustCompile(`(?i)((?:PASSWORD|SECRET|TOKEN|COOKIE|SESSION|DSN|ACCESS_KEY|PRIVATE_KEY)=)[^\s]+`)
var secretPostgres = regexp.MustCompile(`(?i)((?:postgres|postgresql)://[^\s/:]+:)[^\s/@]+(@)`)

func redactDiagnostic(bytes []byte, forbidden []string) []byte {
	if len(bytes) > 64*1024 {
		bytes = bytes[len(bytes)-64*1024:]
	}
	text := string(bytes)
	for _, secret := range forbidden {
		if len(secret) >= 8 {
			text = strings.ReplaceAll(text, secret, "<redacted>")
		}
	}
	text = secretAssignment.ReplaceAllString(text, "${1}<redacted>")
	text = secretPostgres.ReplaceAllString(text, "${1}<redacted>${2}")
	return []byte(text)
}
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func unitDocument(u unit, r unitResult, evidence []string, executed bool) map[string]any {
	var class, reason, code any
	if r.Failure != nil {
		class = r.Failure.Class
		reason = r.Failure.Reason
	}
	if executed && r.Signal == "" {
		code = r.ExitCode
	}
	missing := r.MissingOutputs
	if missing == nil {
		missing = []string{}
	}
	return map[string]any{"schema_id": "cartulary.harness_unit_result.v1", "unit_id": u.ID, "semantic_digest": u.Digest, "status": r.Status, "exit_code": code, "signal": nullableString(r.Signal), "failure_class": class, "failure_reason": reason, "evidence_outputs": evidence, "missing_outputs": missing}
}
func (p *graphPublication) Terminal(u unit, r *unitResult) error {
	if p.closed {
		return harnesslocal.ErrUnsafe
	}
	forbidden := []string{}
	if p.Forbidden != nil {
		var err error
		forbidden, err = p.Forbidden()
		if err != nil {
			return err
		}
	}
	logRoot := filepath.Join(p.root, "unit-logs", safeUnitName.ReplaceAllString(u.ID, "-"))
	directory, err := harnesslocal.PrivateDirectory(logRoot)
	if err != nil {
		return err
	}
	directory.Close()
	for name, bytes := range map[string][]byte{"stdout.log": r.Stdout, "stderr.log": r.Stderr} {
		if err := harnesslocal.ReplaceFile(filepath.Join(logRoot, name), redactDiagnostic(bytes, forbidden)); err != nil {
			return err
		}
	}
	// Raw output tails are no longer needed once bounded redacted logs exist.
	r.Stdout = nil
	r.Stderr = nil
	if r.Status == "passed" {
		for _, relative := range u.Evidence {
			if strings.HasPrefix(relative, "unit-results/") {
				continue
			}
			if !safeRelative(relative) {
				return harnesslocal.ErrUnsafe
			}
			if err := checkRegularEvidence(filepath.Join(p.root, relative)); err != nil {
				r.MissingOutputs = append(r.MissingOutputs, relative)
			}
		}
		if len(r.MissingOutputs) > 0 {
			r.Status = "failed"
			r.Failure = failed("artifact", "artifact_error")
			r.ExitCode = 11
		}
	}
	return p.writeUnit(u, *r, true)
}
func (p *graphPublication) writeUnit(u unit, r unitResult, executed bool) error {
	evidence := slices.Clone(u.Evidence)
	evidence = append(evidence, serviceReferences(r)...)
	if ids := rowIDs(u); len(ids) > 0 {
		for _, name := range []string{"runner.json", "vitest-failure-details.json"} {
			relative := "unit-logs/row-" + ids[0] + "/" + name
			if info, err := os.Lstat(filepath.Join(p.root, relative)); err == nil && info.Mode().IsRegular() {
				evidence = append(evidence, relative)
			}
		}
	}
	slices.Sort(evidence)
	evidence = slices.Compact(evidence)
	for _, relative := range u.Evidence {
		if strings.HasPrefix(relative, "unit-results/") {
			return writeEvidence(p.root, relative, "cartulary.harness_unit_result.v1", unitDocument(u, r, evidence, executed), true)
		}
	}
	return errCanonical
}

var serviceScopeReference = regexp.MustCompile(`^_shared/test-services/[A-Za-z0-9][A-Za-z0-9._-]*/service-scope\.json$`)

func serviceReferences(r unitResult) []string {
	refs := []string{}
	for _, ref := range r.ArtifactRefs {
		if serviceScopeReference.MatchString(ref) {
			refs = append(refs, ref)
		}
	}
	return refs
}
func statusForSelection(selected []string, states map[string]string) string {
	status := "pass"
	for _, id := range selected {
		switch states[id] {
		case "cancelled":
			return "cancelled"
		case "failed":
			status = "fail"
		case "skipped":
			if status == "pass" {
				status = "skipped"
			}
		}
	}
	return status
}
func (p *graphPublication) Finish(result graphResult) (map[string]any, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, harnesslocal.ErrUnsafe
	}
	p.closed = true
	if err := p.facts.validate(); err != nil {
		_ = p.events.Abort()
		return nil, err
	}
	if p.facts.finalMS != result.DurationMS {
		_ = p.events.Abort()
		return nil, errCanonical
	}
	if err := p.events.Close(); err != nil {
		return nil, err
	}
	for _, u := range p.plan.facts.Graph.Units {
		r, ok := result.Units[u.ID]
		if !ok {
			return nil, errCanonical
		}
		_, executed := p.facts.starts[u.ID]
		if !executed {
			logRoot := filepath.Join(p.root, "unit-logs", safeUnitName.ReplaceAllString(u.ID, "-"))
			directory, err := harnesslocal.PrivateDirectory(logRoot)
			if err != nil {
				return nil, err
			}
			directory.Close()
			for _, name := range []string{"stdout.log", "stderr.log"} {
				if err := harnesslocal.PublishNew(filepath.Join(logRoot, name), nil); err != nil {
					return nil, err
				}
			}
		}
		if err := p.writeUnit(u, r, executed); err != nil {
			return nil, err
		}
	}
	cleanup := []map[string]any{}
	for i, r := range result.Cleanup {
		failures := []map[string]any{}
		var class, reason any
		if r.Failure != nil {
			class = r.Failure.Class
			reason = r.Failure.Reason
			failures = append(failures, failureFields(r.Failure))
		}
		cleanup = append(cleanup, map[string]any{"sequence": i + 1, "operation": r.Operation, "unit_id": nullableString(r.UnitID), "fixture_lease_id": nullableString(r.LeaseID), "outcome": r.Outcome, "failure_class": class, "failure_reason": reason, "failures": failures, "artifact_refs": []string{}})
	}
	if err := writeEvidence(p.root, "cleanup-results.json", "cartulary.harness_cleanup_results.v1", map[string]any{"schema_id": "cartulary.harness_cleanup_results.v1", "run_id": p.plan.facts.Manifest["run_id"], "results": cleanup}, true); err != nil {
		return nil, err
	}
	projectionNames := []string{}
	for name := range p.plan.facts.Projections {
		projectionNames = append(projectionNames, name)
	}
	slices.Sort(projectionNames)
	for _, name := range projectionNames {
		projection := p.plan.facts.Projections[name]
		selected := projection.Units
		set := map[string]bool{}
		for _, id := range selected {
			set[id] = true
		}
		children := []string{}
		childUnits := map[string]bool{}
		for _, child := range projectionNames {
			if child == name {
				continue
			}
			contained := true
			for _, id := range p.plan.facts.Projections[child].Units {
				contained = contained && set[id]
			}
			if contained {
				children = append(children, child)
				for _, id := range p.plan.facts.Projections[child].Units {
					childUnits[id] = true
				}
			}
		}
		inclusive, exclusive := []timeInterval{}, []timeInterval{}
		evidence, outputs := []string{}, []string{}
		failures := []*failure{}
		for _, id := range selected {
			inclusive = append(inclusive, p.facts.interval(id))
			if !childUnits[id] {
				exclusive = append(exclusive, p.facts.interval(id))
			}
			u := p.facts.byID[id]
			evidence = append(evidence, u.Evidence...)
			evidence = append(evidence, serviceReferences(result.Units[id])...)
			for _, output := range u.Evidence {
				if !strings.HasPrefix(output, "unit-results/") {
					outputs = append(outputs, output)
				}
			}
			if f := p.facts.terminals[id].failure; f != nil {
				failures = append(failures, f)
			}
		}
		slices.Sort(evidence)
		evidence = slices.Compact(evidence)
		slices.Sort(outputs)
		outputs = slices.Compact(outputs)
		digest, err := semanticDigest(map[string]any{"target": name, "evidence_outputs": outputs})
		if err != nil {
			return nil, err
		}
		_, critical := p.facts.criticalPath(selected)
		duration := intervalUnion(inclusive)
		var class, reason any
		if f := primaryFailure(failures); f != nil {
			class = f.Class
			reason = f.Reason
		}
		summary := map[string]any{"schema_id": "cartulary.harness_target_summary.v1", "target": name, "command_id": projection.CommandID, "status": statusForSelection(selected, result.States), "failure_class": class, "failure_reason": reason, "workload_digest": digest, "unit_ids": selected, "inclusive_wall_ms": duration, "exclusive_wall_ms": intervalUnion(exclusive), "actual_dependency_critical_path_ms": critical, "timing_accounting": p.facts.timing(selected, duration, false), "children": children, "evidence_refs": evidence}
		if err := writeEvidence(p.root, "target-summaries/"+name+".json", "cartulary.harness_target_summary.v1", summary, false); err != nil {
			return nil, err
		}
	}
	counts := map[string]int{"total": len(p.plan.facts.Graph.Units), "passed": 0, "failed": 0, "skipped": 0, "cancelled": 0}
	ids, refs := []string{}, []string{"run-manifest.json", "execution-context.json", "unit-events.ndjson", "cleanup-results.json"}
	for _, u := range p.plan.facts.Graph.Units {
		ids = append(ids, u.ID)
		counts[result.States[u.ID]]++
		refs = append(refs, serviceReferences(result.Units[u.ID])...)
	}
	for _, name := range projectionNames {
		refs = append(refs, "target-summaries/"+name+".json")
	}
	leaseRoot := filepath.Join(p.root, "_shared/fixture-leases")
	if directory, err := harnesslocal.OpenDirectory(leaseRoot, true); err == nil {
		entries, err := directory.ReadDir(-1)
		directory.Close()
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".json") {
				refs = append(refs, "_shared/fixture-leases/"+entry.Name())
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	slices.Sort(refs)
	refs = slices.Compact(refs)
	critical, duration := p.facts.criticalPath(ids)
	timing := p.facts.timing(ids, result.DurationMS, true)
	var total int64
	for _, key := range []string{"setup_ms", "fixture_ms", "execution_ms", "collation_ms", "wrapper_ms", "unattributed_ms"} {
		total += timing[key]
	}
	if total != result.DurationMS {
		return nil, errCanonical
	}
	var class, reason any
	if result.Failure != nil {
		class = result.Failure.Class
		reason = result.Failure.Reason
	}
	status := "pass"
	if result.Status != "passed" {
		status = "fail"
	}
	if counts["cancelled"] > 0 {
		status = "cancelled"
	}
	summary := map[string]any{"schema_id": "cartulary.harness_run_summary.v1", "run_id": p.plan.facts.Manifest["run_id"], "target": p.plan.facts.Manifest["target"], "status": status, "failure_class": class, "failure_reason": reason, "unit_counts": counts, "wall_duration_ms": result.DurationMS, "critical_path": critical, "actual_dependency_critical_path_ms": duration, "timing_accounting": timing, "resource_pressure": p.facts.pressure(), "cache": p.facts.cache, "artifact_refs": refs}
	if err := writeEvidence(p.root, "run-summary.json", "cartulary.harness_run_summary.v1", summary, false); err != nil {
		return nil, err
	}
	return summary, nil
}
func (p *graphPublication) Abort() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	return p.events.Abort()
}
