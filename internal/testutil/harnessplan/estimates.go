package harnessplan

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// TH-HARNESS-REQ-865: observed work estimates. The planner's estimate of a
// row or operation is the median of its retained observations; an authored
// estimate (a row's reviewed estimate, its evidence class's estimate, or an
// operation's policy estimate) is the prior used until it is observed.
// Observations come only from passing retained runs, through the
// estimate-refresh maintenance command, and land as a reviewed source change:
// the source tree stays the plan's only input.
const (
	estimatesPath   = "tools/harness_work_estimates.json"
	EstimatesSchema = "cartulary.harness_work_estimates.v2"
	// maxObservations bounds each key's history; the estimate is its median.
	maxObservations = 3
	// maxFoldedRuns bounds the provenance list that refuses a run twice.
	maxFoldedRuns = 20
)

// Estimates is the observed-estimates document. Generations holds, per row,
// the memory footprints in MiB of the unit generations that ran it
// (TH-HARNESS-REQ-865, G65).
type Estimates struct {
	SchemaID    string             `json:"schema_id"`
	Runs        []FoldedRun        `json:"runs"`
	Rows        map[string][]int64 `json:"rows"`
	Operations  map[string][]int64 `json:"operations"`
	Generations map[string][]int64 `json:"generations"`
}

// Generation claims are sized in these steps, with this floor, and a margin
// of three halves over the largest observed footprint.
const (
	generationStepMB  = 256
	generationFloorMB = 768
)

// generationClaims are a unit generation's resident claims. When every row
// the unit runs has an observed footprint, memory is three halves of the
// largest, rounded up to the step and held between the floor and the
// topology's unit-generation claim, whose remaining claims it keeps.
// Otherwise the unit takes that claim whole. Rows rather than units key the
// observations because packing changes unit identities; a generation's peak
// follows its largest test and its bounded WAL, so the rows' maximum carries
// over when rows are repacked.
func (in *Inputs) generationClaims(rowIDs []string) map[string]int64 {
	claims := maps.Clone(in.Topology.PhysicalResourceClaims["postgres"])
	if len(rowIDs) == 0 {
		return claims
	}
	var largest int64
	for _, id := range rowIDs {
		observed := in.Estimates.Generations[id]
		if len(observed) == 0 {
			return claims
		}
		largest = max(largest, slices.Max(observed))
	}
	sized := (largest*3/2 + generationStepMB - 1) / generationStepMB * generationStepMB
	claims["memory_mb"] = min(claims["memory_mb"], max(generationFloorMB, sized))
	return claims
}

// FoldedRun is the provenance of one run whose observations were folded in.
type FoldedRun struct {
	RunID        string `json:"run_id"`
	Target       string `json:"target"`
	SourceDigest string `json:"source_digest"`
}

func median(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := slices.Sorted(slices.Values(values))
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle] + 1) / 2
}

// rowEstimate is a row's planning estimate: its observed median, otherwise
// its reviewed estimate, otherwise its evidence class's.
func (in *Inputs) rowEstimate(row Row) int64 {
	if observed := median(in.Estimates.Rows[row.RowID]); observed > 0 {
		return observed
	}
	if row.EstimatedWorkMS > 0 {
		return row.EstimatedWorkMS
	}
	// A Playwright row's class estimate is per test it selects.
	if row.Runner == RunnerPlaywright {
		return in.Owner.EvidenceEstimatesMS[row.EvidenceClass] * int64(max(1, len(row.Selector.Titles)))
	}
	return in.Owner.EvidenceEstimatesMS[row.EvidenceClass]
}

// operationEstimate is an operation's planning estimate: its observed
// median, otherwise its policy estimate.
func (in *Inputs) operationEstimate(id string) int64 {
	if observed := median(in.Estimates.Operations[id]); observed > 0 {
		return observed
	}
	return in.Owner.PolicyUnits[id].EstimatedWorkMS
}

// validateEstimates checks the document's shape. Keys of rows or operations
// that no longer exist are ignored by planning and pruned by the next
// refresh, so removing a row never breaks the plan.
func (e *Estimates) validate() error {
	for kind, entries := range map[string]map[string][]int64{"rows": e.Rows, "operations": e.Operations, "generations": e.Generations} {
		for key, values := range entries {
			if key == "" || len(values) == 0 || len(values) > maxObservations {
				return fmt.Errorf("%s %q: between 1 and %d observations required", kind, key, maxObservations)
			}
			for _, value := range values {
				if value <= 0 {
					return fmt.Errorf("%s %q: observations must be positive", kind, key)
				}
			}
		}
	}
	if len(e.Runs) > maxFoldedRuns {
		return fmt.Errorf("at most %d folded runs", maxFoldedRuns)
	}
	return nil
}

// validateEstimateBounds holds planning estimates to the packing rules of
// TH-HARNESS-REQ-865: a non-measurement Go row, with the Go startup and
// cleanup allowance, fits the hard packing bound, and every row estimate is
// below its runner timeout.
func (in *Inputs) validateEstimateBounds() error {
	policy := in.Owner.InvocationPolicies[RunnerGo]
	for _, row := range in.Catalog.Rows {
		estimate := in.rowEstimate(row)
		if estimate >= in.timeoutFor(row.ResourceProfileID) {
			return fmt.Errorf("row %s: estimate %d ms is not below its timeout", row.RowID, estimate)
		}
		if row.Runner == RunnerGo && row.EvidenceClass != "measurement" && estimate+policy.StartupMS+policy.CleanupMS > in.Owner.GoUnitTargetMS {
			return fmt.Errorf("row %s: estimate %d ms with the Go startup and cleanup allowance exceeds the packing bound %d ms; split the row or make it a measurement row",
				row.RowID, estimate, in.Owner.GoUnitTargetMS)
		}
	}
	return nil
}

// timeoutFor is a resource profile's runner timeout, or the owner's default.
func (in *Inputs) timeoutFor(profileID string) int64 {
	index := slices.IndexFunc(in.Topology.ResourceProfiles, func(p ResourceProfile) bool { return p.ID == profileID })
	if index >= 0 && in.Topology.ResourceProfiles[index].RunnerTimeoutMS > 0 {
		return in.Topology.ResourceProfiles[index].RunnerTimeoutMS
	}
	return in.Owner.DefaultTimeoutMS
}

// ErrRefresh reports a run that cannot be folded into the estimates.
var ErrRefresh = errors.New("estimate refresh rejected")

// RefreshReport summarizes one refresh.
type RefreshReport struct {
	RunID                         string
	Rows, Operations, Generations int
	PrunedRows, PrunedOperation   int
}

// retainedPlan is the part of a retained plan a refresh reads.
type retainedPlan struct {
	SchemaID string `json:"schema_id"`
	Graph    struct {
		Units []struct {
			ID          string       `json:"unit_id"`
			Kind        string       `json:"kind"`
			RowBindings []RowBinding `json:"row_bindings"`
		} `json:"units"`
	} `json:"graph"`
	Manifest struct {
		RunID        string `json:"run_id"`
		Target       string `json:"target"`
		SourceDigest string `json:"source_digest"`
	} `json:"manifest"`
}

// RefreshEstimates folds one passing retained run's observations into the
// estimates file under root and rewrites it canonically. A row's observation
// is its share of its unit's runner time beyond the runner's startup and
// cleanup allowance, in proportion to the durations its runner reported; an
// operation's is its unit's runner time. Units served from cache have no
// runner time and contribute nothing.
func RefreshEstimates(root, runDir string) (RefreshReport, error) {
	in, err := Load(root)
	if err != nil {
		return RefreshReport{}, err
	}
	estimates, report, err := in.foldRun(runDir)
	if err != nil {
		return RefreshReport{}, err
	}
	data, err := canonicalEstimates(estimates)
	if err != nil {
		return RefreshReport{}, err
	}
	return report, writeAtomic(filepath.Join(root, estimatesPath), data)
}

// foldRun returns the estimates with one retained run's observations folded
// in, validated against the planning bounds; the inputs are unchanged.
func (in *Inputs) foldRun(runDir string) (Estimates, RefreshReport, error) {
	reject := func(format string, args ...any) (Estimates, RefreshReport, error) {
		return Estimates{}, RefreshReport{}, fmt.Errorf("%w: "+format, append([]any{ErrRefresh}, args...)...)
	}
	var summary struct {
		RunID  string `json:"run_id"`
		Status string `json:"status"`
	}
	if err := readJSON(filepath.Join(runDir, "run-summary.json"), &summary); err != nil || summary.Status != "pass" {
		return reject("%s did not pass", runDir)
	}
	var plan retainedPlan
	if err := readJSON(filepath.Join(runDir, "plan.json"), &plan); err != nil || plan.SchemaID != PlanSchema || plan.Manifest.RunID != summary.RunID {
		return reject("%s has no current-epoch plan", runDir)
	}
	for _, folded := range in.Estimates.Runs {
		if folded.RunID == plan.Manifest.RunID && folded.SourceDigest == plan.Manifest.SourceDigest {
			return reject("run %s is already folded", plan.Manifest.RunID)
		}
	}
	runner, cached, err := runnerTimes(filepath.Join(runDir, "unit-events.ndjson"))
	if err != nil {
		return reject("%s events: %v", runDir, err)
	}
	estimates := Estimates{SchemaID: EstimatesSchema, Runs: slices.Clone(in.Estimates.Runs), Rows: maps.Clone(in.Estimates.Rows), Operations: maps.Clone(in.Estimates.Operations), Generations: maps.Clone(in.Estimates.Generations)}
	if estimates.Rows == nil {
		estimates.Rows = map[string][]int64{}
	}
	if estimates.Operations == nil {
		estimates.Operations = map[string][]int64{}
	}
	if estimates.Generations == nil {
		estimates.Generations = map[string][]int64{}
	}
	observe := func(entries map[string][]int64, key string, value int64) {
		history := append(slices.Clone(entries[key]), max(1, value))
		entries[key] = history[max(0, len(history)-maxObservations):]
	}
	report := RefreshReport{RunID: plan.Manifest.RunID}
	for _, unit := range plan.Graph.Units {
		elapsed, ran := runner[unit.ID]
		if !ran || cached[unit.ID] {
			continue
		}
		if id, ok := strings.CutPrefix(unit.ID, "operation:"); ok {
			if _, known := in.Owner.PolicyUnits[id]; known {
				observe(estimates.Operations, id, elapsed)
				report.Operations++
			}
			continue
		}
		if len(unit.RowBindings) == 0 {
			continue
		}
		runnerType := unit.RowBindings[0].Runner
		overhead, ok := in.runnerOverhead(runnerType)
		if !ok {
			continue
		}
		durations := make([]int64, len(unit.RowBindings))
		total := int64(0)
		for i, binding := range unit.RowBindings {
			var result struct {
				TerminalState string `json:"terminal_state"`
				DurationMS    int64  `json:"duration_ms"`
			}
			if readJSON(filepath.Join(runDir, "rows", binding.RowID+".json"), &result) != nil || result.TerminalState != "passed" {
				return reject("row %s has no passing result", binding.RowID)
			}
			durations[i] = max(0, result.DurationMS)
			total += durations[i]
		}
		work := max(elapsed-overhead, total)
		for i, binding := range unit.RowBindings {
			if _, known := in.rowIndex[binding.RowID]; !known {
				continue
			}
			share := work / int64(len(durations))
			if total > 0 {
				share = work * durations[i] / total
			}
			observe(estimates.Rows, binding.RowID, share)
			report.Rows++
		}
	}
	// Each unit generation's footprint is observed for every row it ran.
	footprints, err := generationFootprints(runDir)
	if err != nil {
		return reject("%s footprints: %v", runDir, err)
	}
	for _, unit := range plan.Graph.Units {
		peak, ok := footprints["postgres:unit:"+unit.ID]
		if !ok || cached[unit.ID] {
			continue
		}
		for _, binding := range unit.RowBindings {
			if _, known := in.rowIndex[binding.RowID]; known {
				observe(estimates.Generations, binding.RowID, peak)
				report.Generations++
			}
		}
	}
	for key := range estimates.Rows {
		if _, known := in.rowIndex[key]; !known {
			delete(estimates.Rows, key)
			report.PrunedRows++
		}
	}
	for key := range estimates.Generations {
		if _, known := in.rowIndex[key]; !known {
			delete(estimates.Generations, key)
		}
	}
	for key := range estimates.Operations {
		if _, known := in.Owner.PolicyUnits[key]; !known {
			delete(estimates.Operations, key)
			report.PrunedOperation++
		}
	}
	estimates.Runs = append(estimates.Runs, FoldedRun{RunID: plan.Manifest.RunID, Target: plan.Manifest.Target, SourceDigest: plan.Manifest.SourceDigest})
	estimates.Runs = estimates.Runs[max(0, len(estimates.Runs)-maxFoldedRuns):]
	if err := estimates.validate(); err != nil {
		return reject("%v", err)
	}
	// The refreshed estimates must still plan: a row observed above the
	// packing bound must be split or reclassified, not written.
	candidate := *in
	candidate.Estimates = estimates
	if err := candidate.validateEstimateBounds(); err != nil {
		return reject("%v", err)
	}
	return estimates, report, nil
}

// GenerationFootprintSchema is the record a broker generation leaves when it
// stops: its allocation and the peak memory of its container, data included.
const GenerationFootprintSchema = "cartulary.test_services.postgres_footprint.v1"

// generationFootprints reads every generation footprint a retained run left,
// as peak MiB by allocation reference.
func generationFootprints(runDir string) (map[string]int64, error) {
	paths, err := filepath.Glob(filepath.Join(runDir, "_shared", "test-services", "*", "postgres-footprint.json"))
	if err != nil {
		return nil, err
	}
	footprints := map[string]int64{}
	for _, path := range paths {
		var record struct {
			SchemaID        string `json:"schema_id"`
			AllocationRef   string `json:"allocation_ref"`
			MemoryPeakBytes int64  `json:"memory_peak_bytes"`
		}
		if err := readJSON(path, &record); err != nil || record.SchemaID != GenerationFootprintSchema || record.AllocationRef == "" || record.MemoryPeakBytes <= 0 {
			return nil, fmt.Errorf("%s is not a generation footprint", path)
		}
		footprints[record.AllocationRef] = max(footprints[record.AllocationRef], (record.MemoryPeakBytes+(1<<20)-1)>>20)
	}
	return footprints, nil
}

// runnerOverhead is the startup and cleanup allowance the planner adds to a
// runner unit's row estimates.
func (in *Inputs) runnerOverhead(runner string) (int64, bool) {
	switch runner {
	case RunnerGo, RunnerVitest:
		policy := in.Owner.InvocationPolicies[runner]
		return policy.StartupMS + policy.CleanupMS, true
	case RunnerPlaywright:
		policy := in.Browser.InvocationPolicy
		return policy.StartupMS + policy.CleanupMS, true
	}
	return 0, false
}

// runnerTimes sums each unit's runner phase intervals and notes units served
// from the cache.
func runnerTimes(path string) (map[string]int64, map[string]bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	type interval struct {
		unit  string
		start int64
	}
	open := map[string]interval{}
	runner, cached := map[string]int64{}, map[string]bool{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 16<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !bytes.Contains(line, []byte(`"phase":"runner"`)) && !bytes.Contains(line, []byte(`"cache_hit"`)) {
			continue
		}
		var event struct {
			Event      string `json:"event"`
			UnitID     string `json:"unit_id"`
			Monotonic  int64  `json:"monotonic_ms"`
			IntervalID string `json:"interval_id"`
			Phase      string `json:"phase"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, nil, err
		}
		switch {
		case event.Event == "cache_hit":
			cached[event.UnitID] = true
		case event.Event == "phase_started" && event.Phase == "runner":
			open[event.IntervalID] = interval{event.UnitID, event.Monotonic}
		case event.Event == "phase_finished" && event.Phase == "runner":
			if started, ok := open[event.IntervalID]; ok {
				runner[started.unit] += event.Monotonic - started.start
				delete(open, event.IntervalID)
			}
		}
	}
	return runner, cached, scanner.Err()
}

func readJSON(path string, into any) error {
	data, err := readBounded(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, into)
}

// canonicalEstimates writes the document with sorted keys, one entry per
// line, so a refresh reviews as a line diff.
func canonicalEstimates(e Estimates) ([]byte, error) {
	var out bytes.Buffer
	out.WriteString("{\n  \"schema_id\": \"" + EstimatesSchema + "\",\n  \"runs\": [")
	for i, run := range e.Runs {
		entry, err := json.Marshal(run)
		if err != nil {
			return nil, err
		}
		out.WriteString(map[bool]string{true: "\n    ", false: ",\n    "}[i == 0])
		out.Write(entry)
	}
	out.WriteString(map[bool]string{true: "]", false: "\n  ]"}[len(e.Runs) == 0])
	for _, section := range []struct {
		name    string
		entries map[string][]int64
	}{{"rows", e.Rows}, {"operations", e.Operations}, {"generations", e.Generations}} {
		out.WriteString(",\n  \"" + section.name + "\": {")
		keys := slices.SortedFunc(maps.Keys(section.entries), cmp.Compare[string])
		for i, key := range keys {
			name, _ := json.Marshal(key)
			values, _ := json.Marshal(section.entries[key])
			out.WriteString(map[bool]string{true: "\n    ", false: ",\n    "}[i == 0])
			out.Write(name)
			out.WriteString(": ")
			out.Write(values)
		}
		out.WriteString(map[bool]string{true: "}", false: "\n  }"}[len(keys) == 0])
	}
	out.WriteString("\n}\n")
	if _, err := canonicaljson.DecodeStrict(out.Bytes()); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func writeAtomic(path string, data []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".estimates-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), path)
}
