//go:build linux

package harnessadmission

import (
	"context"
	"crypto/sha256"
	"fmt"
	"maps"
	"slices"
)

// Diagnostic identities are one-way names, never usable lease capabilities.
func ObservationReference(kind, token string) string {
	return fmt.Sprintf("%s:%x", kind, sha256.Sum256([]byte("harness-observation\x00"+kind+"\x00"+token)))
}

type ObservedHolder struct {
	Reference string     `json:"reference"`
	Claims    Quantities `json:"claims"`
}

type Decision struct {
	RequestRef         string           `json:"request_ref"`
	Sequence           float64          `json:"sequence"`
	QueuePosition      int              `json:"queue_position"`
	Mode               string           `json:"mode"`
	Granted            bool             `json:"granted"`
	RequestedClaims    Quantities       `json:"requested_claims"`
	Capacities         Quantities       `json:"effective_capacities"`
	Deficits           Quantities       `json:"current_deficits"`
	PersistentDeficits Quantities       `json:"persistent_deficits"`
	Ancestors          []ObservedHolder `json:"required_ancestors"`
	Holders            []ObservedHolder `json:"active_holders"`
	Residents          []ObservedHolder `json:"residents"`
	Prospective        []ObservedHolder `json:"prospective"`
	PriorityBlockers   []string         `json:"priority_blockers"`
	ExclusiveBlockers  []string         `json:"exclusive_blockers"`
}

type RetirementDecision struct {
	RequestRef       string     `json:"request_ref"`
	Residents        []string   `json:"resident_refs"`
	Before           Quantities `json:"before_deficits"`
	After            Quantities `json:"after_deficits"`
	ActiveReferences []string   `json:"active_references"`
	Eligible         bool       `json:"eligible"`
}

type Observation struct {
	Evaluation *RetirementDecision `json:"evaluation,omitempty"`
	UnitID     string              `json:"unit_id"`
	RequestRef *string             `json:"request_ref"`
	Clock      string              `json:"clock_identity"`
	Kind       string              `json:"kind"`
	StartMS    float64             `json:"start_ms"`
	EndMS      float64             `json:"end_ms"`
	DurationMS float64             `json:"duration_ms"`
	Count      int                 `json:"observations"`
	Phase      string              `json:"phase,omitempty"`
	Outcome    string              `json:"outcome,omitempty"`
	Decision   *Decision           `json:"decision,omitempty"`
	Residents  *[]string           `json:"resident_refs,omitempty"`
	Reason     string              `json:"reason,omitempty"`
}

type observationKey struct{}
type phasesSuppressedKey struct{}
type observer struct {
	unit  string
	clock func() float64
	emit  func(Observation)
}

func WithObservation(ctx context.Context, unit string, clock func() float64, emit func(Observation)) context.Context {
	if clock == nil || emit == nil {
		return ctx
	}
	return context.WithValue(ctx, observationKey{}, &observer{unit, clock, emit})
}

// Phase records disjoint operation time. Repetition may be aggregated; the
// first/last span is not a claim of continuous occupancy.
func Phase(ctx context.Context, phase string) func() {
	o, _ := ctx.Value(observationKey{}).(*observer)
	if o == nil || ctx.Value(phasesSuppressedKey{}) == true {
		return func() {}
	}
	start := o.clock()
	return func() {
		end := o.clock()
		o.emit(Observation{UnitID: o.unit, Clock: "graph-monotonic", Kind: "phase", StartMS: start, EndMS: end, DurationMS: max(0, end-start), Count: 1, Phase: phase})
	}
}

// Retirement observes the physical owner's actual cleanup, outside admission's
// mutex. Diagnostics do not release claims or decide settlement.
func Retirement(ctx context.Context, residents []Resident, reason string) func(error) {
	o, _ := ctx.Value(observationKey{}).(*observer)
	if o == nil {
		return func(error) {}
	}
	refs := []string{}
	for _, r := range residents {
		refs = append(refs, ObservationReference("resident", r.Token))
	}
	slices.Sort(refs)
	start := o.clock()
	return func(err error) {
		end := o.clock()
		outcome := "settled"
		if err != nil {
			outcome = "unresolved"
		}
		o.emit(Observation{UnitID: o.unit, Clock: "graph-monotonic", Kind: "retirement", StartMS: start, EndMS: end, DurationMS: max(0, end-start), Count: 1, Residents: &refs, Reason: reason, Outcome: outcome})
	}
}

func observedTransaction(ctx context.Context, request Request) (bool, error) {
	o, _ := ctx.Value(observationKey{}).(*observer)
	if o == nil {
		return Transact(ctx, request)
	}
	d := &Decision{}
	request.diagnostic = d
	start := o.clock()
	value, err := Transact(ctx, request)
	end := o.clock()
	if d.RequestRef != "" {
		outcome := "committed"
		if err != nil {
			outcome = "failed"
		}
		d.Granted = value && err == nil
		o.emit(Observation{UnitID: o.unit, RequestRef: &d.RequestRef, Clock: "graph-monotonic", Kind: "decision", StartMS: start, EndMS: end, Count: 1, Outcome: outcome, Decision: d})
	}
	return value, err
}

func observedLease(l *lease) ObservedHolder {
	return ObservedHolder{ObservationReference("request", l.Token), maps.Clone(l.Claims)}
}
