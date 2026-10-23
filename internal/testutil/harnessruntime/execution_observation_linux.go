//go:build linux

package harnessruntime

import (
	"encoding/json"
	"io"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"sync"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesscollect"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

const executionRecordSchema = "cartulary.harness_execution_record.v2"

type executionSession struct {
	mu                                               sync.Mutex
	root, runRoot                                    string
	manifest                                         map[string]any
	policy                                           map[string]any
	enabled, captured, published, partial, truncated bool
	omissions, bytes                                 int
	records                                          map[string]map[string]any
	missing                                          map[string]bool
	lifecycle                                        allocationLifecycleSection
	proof                                            func(map[string]any, map[string]any)
}

func newExecutionSession(runtime *harnesssession.Runtime, runRoot string, manifest, policy map[string]any, proof func(map[string]any, map[string]any)) *executionSession {
	root, pathErr := runtime.PrivatePath("execution-observations")
	s := &executionSession{lifecycle: newAllocationLifecycleSection(), root: root, runRoot: runRoot, manifest: manifest, policy: policy, records: map[string]map[string]any{}, missing: map[string]bool{}, proof: proof, enabled: manifest["instrumentation"].(map[string]any)["mode"] == "basic", partial: pathErr != nil}
	if pathErr != nil {
		return s
	}
	if s.enabled {
		dir, err := harnesslocal.PrivateDirectory(s.root)
		if err == nil {
			dir.Close()
			b, _ := json.Marshal(policy)
			err = harnesslocal.PublishNew(filepath.Join(s.root, "limits.json"), b)
		}
		if err == nil {
			err = harnesslocal.PublishNew(filepath.Join(s.root, "open"), []byte("1\n"))
		}
		s.partial = err != nil
	}
	return s
}
func (s *executionSession) limit(name string) int { return int(s.policy[name].(float64)) }
func (s *executionSession) record(v map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordLocked(v, false)
}
func (s *executionSession) recordLocked(v map[string]any, launchUpdate bool) {
	if !s.enabled || s.published {
		return
	}
	b, err := documentBytes(executionRecordSchema, v)
	if err != nil || len(b) > s.limit("maximum_record_bytes") {
		s.partial = true
		s.omissions++
		return
	}
	// Normalize and copy at ingress. Neither a producer nor a later callback can
	// mutate accepted facts, and all schema predicates see the same JSON domain.
	normalized, _ := harnessschema.Decode(b)
	v = normalized.(map[string]any)
	ref := v["ref"].(string)
	if old := s.records[ref]; old != nil {
		previous, _ := json.Marshal(old)
		if string(previous) == string(b) {
			return
		}
		allocation := old["kind"] == "allocation" && v["kind"] == "allocation" && old["allocation_ref"] == v["allocation_ref"] && old["unit_id"] == v["unit_id"] && old["capability"] == v["capability"]
		launch := launchUpdate && old["kind"] == "launch" && v["kind"] == "launch" && old["invocation_id"] == v["invocation_id"]
		if old["outcome"] != "incomplete" || (!allocation && !launch) {
			s.partial = true
			s.omissions++
			return
		}
		s.bytes -= len(previous)
		delete(s.records, ref)
	}
	if len(s.records) >= s.limit("maximum_execution_records") || s.bytes+len(b) > s.limit("maximum_metadata_bytes")-512*1024 {
		s.truncated = true
		s.omissions++
		if len(s.missing) < s.limit("maximum_execution_records") {
			s.missing[ref] = true
		}
		return
	}
	s.records[ref] = v
	s.bytes += len(b)
}

var executionSlot = regexp.MustCompile(`^slot-[0-9]+$`)

func (s *executionSession) capture() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled || s.captured {
		return
	}
	s.captured = true
	if harnesslocal.ReplaceFile(filepath.Join(s.root, "open"), []byte("0\n")) != nil {
		s.partial = true
		return
	}
	dir, err := harnesslocal.OpenDirectory(s.root, true)
	if err != nil {
		s.partial = true
		return
	}
	defer dir.Close()
	names, err := dir.Readdirnames(s.limit("maximum_execution_records") + 4)
	if err != nil && err != io.EOF || len(names) > s.limit("maximum_execution_records")+3 {
		s.partial = true
		s.truncated = true
		return
	}
	slices.Sort(names)
	for _, name := range names {
		if name == "truncated" {
			s.truncated = true
			continue
		}
		if name == "limits.json" || name == "open" {
			continue
		}
		if !executionSlot.MatchString(name) {
			s.partial = true
			s.omissions++
			continue
		}
		b, err := harnesslocal.ReadFileAt(dir, name, int64(s.limit("maximum_record_bytes")), false)
		var raw any
		if err == nil {
			raw, err = harnessschema.Decode(b)
		}
		v, ok := raw.(map[string]any)
		if err != nil || !ok {
			s.partial = true
			s.omissions++
			continue
		}
		if proof, present := v["process_proof"]; present {
			if harnessschema.Validate("cartulary.harness_private_launch_observation.v2", v) != nil {
				s.partial = true
				s.omissions++
				continue
			}
			record := v["record"].(map[string]any)
			p := proof.(map[string]any)
			if record["kind"] != "launch" || record["launched"] != true || p["start"] != p["stat"].(map[string]any)["start"] {
				s.partial = true
				s.omissions++
				continue
			}
			s.recordLocked(record, false)
			if s.proof != nil {
				optionalObservation(func() { s.proof(p, record) })
			}
		} else {
			s.recordLocked(v, false)
		}
	}
}
func (s *executionSession) publish(clock any) error {
	s.capture()
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled || s.published {
		return nil
	}
	s.published = true
	records := []map[string]any{}
	for _, ref := range slices.Sorted(maps.Keys(s.records)) {
		records = append(records, s.records[ref])
		delete(s.missing, ref)
	}
	for _, r := range records {
		dependencies := []string{}
		if p, ok := r["parent_invocation_id"].(string); ok {
			dependencies = append(dependencies, "launch:"+p)
		}
		if p, ok := r["invocation_id"].(string); ok && r["kind"] != "launch" {
			dependencies = append(dependencies, "launch:"+p)
		}
		if p, ok := r["allocation_ref"].(string); ok && r["kind"] != "allocation" {
			dependencies = append(dependencies, p)
		}
		for _, ref := range dependencies {
			if s.records[ref] == nil {
				s.missing[ref] = true
			}
		}
		if r["outcome"] == "incomplete" {
			s.partial = true
		}
	}
	if len(s.missing) > s.limit("maximum_execution_records") {
		return harnesslocal.ErrUnsafe
	}
	if len(s.missing) > 0 {
		s.partial = true
		s.omissions = max(s.omissions, len(s.missing))
	}
	status := "complete"
	var omissions any = s.omissions
	if s.partial {
		status = "partial"
	}
	if s.truncated {
		status = "truncated"
		omissions = nil
	}
	b, _ := json.Marshal(records)
	index := map[string]any{"schema_id": "cartulary.harness_execution_index.v3", "run_id": s.manifest["run_id"], "source_digest": s.manifest["source_digest"], "graph_digest": s.manifest["graph_digest"], "policy_digest": s.manifest["instrumentation"].(map[string]any)["policy_digest"], "status": status, "omitted_records": omissions, "graph_clock": clock, "omitted_refs": append([]string{}, slices.Sorted(maps.Keys(s.missing))...), "records": records, "records_digest": bytesDigest(b), "allocation_lifecycle": s.lifecycle.document()}
	encoded, err := documentBytes("cartulary.harness_execution_index.v3", index)
	if err != nil {
		return err
	}
	if len(encoded)+1 > s.limit("maximum_metadata_bytes")+s.limit("maximum_allocation_lifecycle_bytes") {
		return harnesslocal.ErrUnsafe
	}
	return writeEvidence(s.runRoot, "diagnostics/execution-index.json", "cartulary.harness_execution_index.v3", index, false)
}
func (s *executionSession) fixture(v harnessfixtures.Observation) {
	if v.Kind == "allocation_lifecycle" {
		s.recordLifecycle(v.Lifecycle)
		return
	}
	if v.Kind == "allocation" && v.Capability == "browser_stack" && v.Outcome == "incomplete" {
		s.beginLifecycle(v.AllocationRef)
	}
	r := map[string]any{"kind": v.Kind, "ref": v.Ref, "allocation_ref": v.AllocationRef, "unit_id": nullableString(v.UnitID), "capability": v.Capability, "ownership": nullableString(v.Ownership)}
	if v.Kind == "lease" {
		r["ref"] = "lease:" + v.LeaseRef
		r["lease_ref"] = v.LeaseRef
	} else {
		r["outcome"] = v.Outcome
	}
	s.record(r)
}

type launchObservation struct {
	session  *executionSession
	record   map[string]any
	register func(map[string]any, map[string]any)
	settled  bool
}

func (s *executionSession) launch(identity map[string]any) processObservation {
	if !s.enabled {
		return nil
	}
	r := maps.Clone(identity)
	delete(r, "schema_id")
	r["kind"] = "launch"
	r["ref"] = "launch:" + identity["invocation_id"].(string)
	r["launched"] = false
	r["outcome"] = "incomplete"
	s.record(r)
	return &launchObservation{session: s, record: r, register: s.proof}
}
func (o *launchObservation) Spawned(pid int) {
	if o.settled {
		return
	}
	o.record["launched"] = true
	o.session.mu.Lock()
	o.session.recordLocked(o.record, true)
	o.session.mu.Unlock()
	if o.register != nil {
		p, err := harnesscollect.ProcessIdentity(pid)
		if err == nil {
			o.register(p, o.record)
		}
	}
}
func (o *launchObservation) Settled(outcome string) {
	if o.settled {
		return
	}
	o.settled = true
	o.record["outcome"] = outcome
	o.session.mu.Lock()
	defer o.session.mu.Unlock()
	o.session.recordLocked(o.record, true)
}
