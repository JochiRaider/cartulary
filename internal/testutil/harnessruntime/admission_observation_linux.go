//go:build linux

package harnessruntime

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
)

type admissionSection struct {
	records            []harnessadmission.Observation
	last               map[string]int
	bytes, omissions   int
	partial, truncated bool
}

func (s *executionSession) admission(value harnessadmission.Observation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		if recover() != nil {
			s.hostAdmission.partial = true
			s.hostAdmission.omissions++
		}
	}()
	if !s.enabled || s.published {
		return
	}
	a := &s.hostAdmission
	if a.last == nil {
		a.last = map[string]int{}
	}
	encoded, err := documentBytes("cartulary.harness_host_admission_record.v4", value)
	if err != nil || len(encoded) > s.limit("maximum_record_bytes") {
		a.partial = true
		a.omissions++
		return
	}
	// Copy at ingress so transaction-owned maps cannot rewrite accepted facts.
	var next harnessadmission.Observation
	if json.Unmarshal(encoded, &next) != nil {
		a.partial = true
		a.omissions++
		return
	}
	compactDecision(next.Decision)
	key := value.UnitID + ":" + value.Kind
	if value.Kind == "phase" {
		key += ":" + value.Phase
	}
	if (value.Kind == "decision" || value.Kind == "retirement_evaluation") && value.RequestRef != nil {
		key += ":" + *value.RequestRef
		if value.Evaluation != nil {
			refs, _ := json.Marshal(value.Evaluation.Residents)
			key += ":" + string(refs)
		}
	}
	index := -1
	if prior, ok := a.last[key]; ok && value.Kind != "retirement" {
		old := a.records[prior]
		same := value.Kind == "phase"
		if value.Kind == "decision" {
			same = decisionIdentity(old) == decisionIdentity(next)
		} else if !same {
			left, _ := json.Marshal([]any{old.Decision, old.Evaluation})
			right, _ := json.Marshal([]any{next.Decision, next.Evaluation})
			same = string(left) == string(right) && old.Outcome == next.Outcome
		}
		if same {
			if next.StartMS < old.EndMS {
				a.partial = true
				a.omissions++
				return
			}
			index = prior
			next.StartMS = old.StartMS
			next.Count += old.Count
			next.DurationMS += old.DurationMS
		}
	}
	encoded, _ = json.Marshal(next)
	if len(encoded) > s.limit("maximum_record_bytes") {
		a.partial = true
		a.omissions++
		return
	}
	delta := len(encoded) + 1
	if index >= 0 {
		old, _ := json.Marshal(a.records[index])
		delta -= len(old) + 1
	}
	if index < 0 && len(a.records) >= s.limit("maximum_host_admission_records") || a.bytes+delta > s.limit("maximum_host_admission_bytes")-4096 {
		a.truncated = true
		a.omissions++
		return
	}
	if index < 0 {
		index = len(a.records)
		a.records = append(a.records, next)
	} else {
		a.records[index] = next
	}
	a.last[key] = index
	a.bytes += delta
}

// decisionIdentity is the material content of an admission decision
// (TH-HARNESS-REQ-857). Re-evaluating a waiting request changes queue position,
// deficit amounts and holders on almost every transaction; only a change in this
// identity starts a new record, so record volume follows decision transitions
// rather than retry frequency.
func decisionIdentity(o harnessadmission.Observation) string {
	d := o.Decision
	if d == nil {
		return ""
	}
	deficitSet := func(q harnessadmission.Quantities) []string {
		keys := []string{}
		for k, v := range q {
			if v != 0 {
				keys = append(keys, k)
			}
		}
		slices.Sort(keys)
		return keys
	}
	b, _ := json.Marshal(struct {
		Outcome          string                            `json:"outcome"`
		Mode             string                            `json:"mode"`
		Granted          bool                              `json:"granted"`
		Requested        harnessadmission.Quantities       `json:"requested"`
		Capacities       harnessadmission.Quantities       `json:"capacities"`
		Ancestors        []harnessadmission.ObservedHolder `json:"ancestors"`
		Current          []string                          `json:"current"`
		Persistent       []string                          `json:"persistent"`
		PriorityBlocked  bool                              `json:"priority_blocked"`
		ExclusiveBlocked bool                              `json:"exclusive_blocked"`
	}{o.Outcome, d.Mode, d.Granted, d.RequestedClaims, d.Capacities, d.Ancestors, deficitSet(d.Deficits), deficitSet(d.PersistentDeficits), len(d.PriorityBlockers) > 0, len(d.ExclusiveBlockers) > 0})
	return string(b)
}

// compactDecision omits zero claim and deficit quantities. Effective capacities
// stay complete because a zero capacity is itself a fact.
func compactDecision(d *harnessadmission.Decision) {
	if d == nil {
		return
	}
	prune := func(q harnessadmission.Quantities) {
		for k, v := range q {
			if v == 0 {
				delete(q, k)
			}
		}
	}
	prune(d.RequestedClaims)
	prune(d.Deficits)
	prune(d.PersistentDeficits)
	for _, holders := range [][]harnessadmission.ObservedHolder{d.Ancestors, d.Holders, d.Residents, d.Prospective} {
		for _, h := range holders {
			prune(h.Claims)
		}
	}
}

func (a *admissionSection) document() map[string]any {
	records := append([]harnessadmission.Observation{}, a.records...)
	slices.SortFunc(records, func(x, y harnessadmission.Observation) int {
		if c := strings.Compare(x.UnitID, y.UnitID); c != 0 {
			return c
		}
		if x.StartMS < y.StartMS {
			return -1
		}
		if x.StartMS > y.StartMS {
			return 1
		}
		if c := strings.Compare(x.Kind, y.Kind); c != 0 {
			return c
		}
		return strings.Compare(x.Phase, y.Phase)
	})
	status := "complete"
	if a.partial {
		status = "partial"
	}
	if a.truncated {
		status = "truncated"
	}
	b, _ := json.Marshal(records)
	return map[string]any{"status": status, "omitted_records": a.omissions, "records": records, "records_digest": bytesDigest(b)}
}
