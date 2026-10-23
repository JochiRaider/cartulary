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
	return filepath.Join(os.TempDir(), "cartulary-host-admission-"+strconv.Itoa(os.Getuid()))
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
func AcquireWithResidents(ctx context.Context, root string, claims, capacities Quantities, parent *string, exclusive bool, residents []Resident, waiting func(context.Context, *Lease) error) (*Lease, error) {
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
func AcquireWithPreparedResidents(ctx context.Context, root string, claims, capacities Quantities, parent *string, exclusive bool, prepare func(context.Context) ([]Resident, error), waiting func(context.Context, *Lease) error) (*Lease, error) {
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

func acquire(ctx context.Context, root string, claims, capacities Quantities, parent *string, mode string, residents []Resident, prepare func(context.Context) ([]Resident, error), waiting func(context.Context, *Lease) error) (*Lease, error) {
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
		for _, k := range []string{"browser_stack", "cpu", "io", "memory_mb", "object_store", "port_lane", "postgres", "process", "service_stack", "volume"} {
			out[k] = input[k]
		}
		return out
	}
	request := Request{Root: root, Operation: "admit", Token: lease.Token, Owner: owner, Mode: mode, Claims: quantities(claims), Capacities: quantities(capacities), Parent: parent, Residents: residents}
	for {
		var admitted bool
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
		if err == nil && !admitted && waiting != nil {
			done := Phase(ctx, "pressure_retirement")
			err = waiting(ctx, lease)
			done()
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
