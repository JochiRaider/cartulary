//go:build linux

package harnessadmission

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var ErrRelease = errors.New("host admission release failed")

func DefaultRoot() string {
	return filepath.Join(os.TempDir(), "cartulary-host-admission-g"+strconv.Itoa(stateGeneration)+"-"+strconv.Itoa(os.Getuid()))
}

type Lease struct {
	Token    string
	root     string
	mu       sync.Mutex
	released bool
}

func (l *Lease) Release(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return nil
	}
	_, err := Transact(ctx, Request{Root: l.root, Operation: "release", Token: l.Token})
	if err == nil {
		l.released = true
	}
	return err
}
func (l *Lease) Bind(pid int) error {
	proof, err := Identity(pid)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = Transact(ctx, Request{Root: l.root, Operation: "bind", Token: l.Token, Guard: proof})
	return err
}

// Acquire shares queueing, cancellation and release ownership across native graph
// and review callers. Public failure mapping remains with the calling subsystem.
func Acquire(ctx context.Context, root string, claims, capacities Quantities, parent *string, exclusive bool) (*Lease, error) {
	return AcquireWithResidents(ctx, root, claims, capacities, parent, exclusive, nil, nil)
}

// AcquireWithResidents atomically admits work and any missing physical
// reservations. A queued worker holds neither active nor prospective capacity.
// While waiting, the owner may retire idle allocations outside the state mutex.
func AcquireWithResidents(ctx context.Context, root string, claims, capacities Quantities, parent *string, exclusive bool, residents []Resident, waiting Retire) (*Lease, error) {
	mode := "shared"
	if exclusive {
		mode = "exclusive"
	}
	return acquire(ctx, root, claims, capacities, parent, mode, residents, nil, waiting)
}
func AcquireDrain(ctx context.Context, root string, claims, capacities Quantities, parent *string, residents []Resident) (*Lease, error) {
	return acquire(ctx, root, claims, capacities, parent, "drain", residents, nil, nil)
}

// AcquireWithPreparedResidents keeps one queue position while the physical owner
// reacquires logical slots and refreshes resident references before each attempt.
// waiting yields those logical reservations before the next queue wait. The owner
// also settles them on return errors; a granted lease retains exact references.
func AcquireWithPreparedResidents(ctx context.Context, root string, claims, capacities Quantities, parent *string, exclusive bool, prepare func(context.Context) ([]Resident, error), waiting Retire) (*Lease, error) {
	mode := "shared"
	if exclusive {
		mode = "exclusive"
	}
	return acquire(ctx, root, claims, capacities, parent, mode, nil, prepare, waiting)
}

// NeedsRetirement evaluates the exact proposed release set. It neither grants
// cleanup activity nor releases capacity; the physical owner must pin the
// generation, settle it and then reevaluate admission.
func (l *Lease) NeedsRetirement(ctx context.Context, residents []Resident) (bool, error) {
	request := Request{Root: l.root, Operation: "retirement_required", Token: l.Token, Residents: residents}
	o, _ := ctx.Value(observationKey{}).(*observer)
	if o == nil {
		return Transact(ctx, request)
	}
	decision := &RetirementDecision{}
	request.retirementDiagnostic = decision
	start := o.clock()
	eligible, err := Transact(ctx, request)
	end := o.clock()
	if decision.RequestRef != "" {
		outcome := "committed"
		if err != nil {
			outcome = "failed"
		}
		decision.Eligible = eligible && err == nil
		o.emit(Observation{UnitID: o.unit, RequestRef: &decision.RequestRef, Clock: "graph-monotonic", Kind: "retirement_evaluation", StartMS: start, EndMS: end, Count: 1, Outcome: outcome, Evaluation: decision})
	}
	return eligible, err
}

// Retire is the owner's pressure retirement between admission attempts. It
// reports whether retirement released an allocation or deferred one only
// because another request provisionally pinned it, so a later round may still
// relieve the wait.
type Retire func(context.Context, *Lease) (relievable bool, err error)

func acquire(ctx context.Context, root string, claims, capacities Quantities, parent *string, mode string, residents []Resident, prepare func(context.Context) ([]Resident, error), waiting Retire) (*Lease, error) {
	if mode == "drain" {
		ctx = context.WithValue(ctx, phasesSuppressedKey{}, true)
	}
	if root == "" {
		root = DefaultRoot()
	}
	owner, err := Identity(os.Getpid())
	if err != nil {
		return nil, err
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	lease := &Lease{Token: hex.EncodeToString(nonce[:]), root: root}
	quantities := func(input Quantities) Quantities {
		out := Quantities{}
		for _, k := range resources {
			out[k] = input[k]
		}
		return out
	}
	request := Request{Root: root, Operation: "admit", Token: lease.Token, Owner: owner, Mode: mode, Claims: quantities(claims), Capacities: quantities(capacities), Parent: parent, Residents: residents}
	// The identity of the previous round's unsatisfiable decision, kept only
	// while that round's pressure retirement could not relieve it. Such a
	// decision unchanged on the next round can never be granted: waiting would
	// hold the queue position forever (TH-HARNESS-REQ-803).
	previous := ""
	for {
		var admitted bool
		var blocked Unsatisfiable
		request.blocked = &blocked
		err := ctx.Err()
		if err == nil && prepare != nil {
			done := Phase(ctx, "logical_slot")
			request.Residents, err = prepare(ctx)
			done()
		}
		if err == nil {
			done := Phase(ctx, "admission_transaction")
			admitted, err = observedTransaction(ctx, request)
			done()
		}
		if err == nil && !admitted && blocked.identity != "" && blocked.identity == previous {
			err = &blocked
		}
		previous = ""
		if err == nil && !admitted {
			relievable := false
			if waiting != nil {
				done := Phase(ctx, "pressure_retirement")
				relievable, err = waiting(ctx, lease)
				done()
			}
			if !relievable {
				previous = blocked.identity
			}
		}
		if err == nil && admitted && ctx.Err() == nil {
			return lease, nil
		}
		if err != nil || ctx.Err() != nil {
			cleanup, done := context.WithTimeout(context.Background(), 5*time.Second)
			releaseErr := lease.Release(cleanup)
			done()
			if releaseErr != nil {
				return nil, errors.Join(err, ctx.Err(), ErrRelease, releaseErr)
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, err
		}
		done := Phase(ctx, "admission_wait")
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-ctx.Done():
		case <-timer.C:
		}
		timer.Stop()
		done()
	}
}
