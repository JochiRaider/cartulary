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
	return acquire(ctx, root, claims, capacities, parent, mode, residents, waiting)
}
func AcquireDrain(ctx context.Context, root string, claims, capacities Quantities, parent *string, residents []Resident) (*Lease, error) {
	return acquire(ctx, root, claims, capacities, parent, "drain", residents, nil)
}

// TryAcquireWithResidents lets a physical owner coordinate a scarce logical slot
// with host admission. A miss withdraws the queued request before the caller
// releases its slot; no active or prospective reservation survives the miss.
func TryAcquireWithResidents(ctx context.Context, root string, claims, capacities Quantities, parent *string, residents []Resident, waiting func(context.Context, *Lease) error) (*Lease, error) {
	return acquire(ctx, root, claims, capacities, parent, "shared", residents, func(ctx context.Context, pending *Lease) error {
		if waiting != nil {
			if err := waiting(ctx, pending); err != nil {
				return err
			}
		}
		return errNotAdmitted
	})
}

// NeedsRetirement reports whether releasing a resident with these claims can
// relieve a capacity shortage that remains after currently active work drains.
// It neither grants activity nor releases capacity; exact cleanup still owns it.
func (l *Lease) NeedsRetirement(ctx context.Context, claims Quantities) (bool, error) {
	return Transact(ctx, Request{Root: l.root, Operation: "retirement_required", Token: l.Token, Claims: claims})
}

var errNotAdmitted = errors.New("host admission not granted")

func acquire(ctx context.Context, root string, claims, capacities Quantities, parent *string, mode string, residents []Resident, waiting func(context.Context, *Lease) error) (*Lease, error) {
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
		admitted, err := Transact(ctx, request)
		if err == nil && !admitted && waiting != nil {
			err = waiting(ctx, lease)
		}
		if err == nil && admitted && ctx.Err() == nil {
			return lease, nil
		}
		if err != nil || ctx.Err() != nil {
			cleanup, done := context.WithTimeout(context.Background(), 5*time.Second)
			releaseErr := lease.Release(cleanup)
			done()
			if releaseErr != nil {
				return nil, errors.Join(ErrRelease, releaseErr)
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if errors.Is(err, errNotAdmitted) {
				return nil, nil
			}
			return nil, err
		}
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-ctx.Done():
		case <-timer.C:
		}
		timer.Stop()
	}
}
