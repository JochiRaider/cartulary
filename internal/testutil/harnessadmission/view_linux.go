//go:build linux

package harnessadmission

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// View is one locked reading of the host ledger (G30, TH-HARNESS-REQ-803
// amended 2026-11-25). A graph scheduler evaluates its ready work against one
// view per decision without writing, and transacts only for work the view
// admits. Work therefore holds nothing, not even a queue position, until every
// claim and resident it needs is granted together.
type View struct {
	state *state
	owner Proof
	alive map[Proof]bool
}

// Evaluation is a prospective request's outcome against a view.
type Evaluation struct {
	Admissible bool
	// Deficits are the resources the request would exceed; empty when only
	// queue priority or an exclusive fence denies it.
	Deficits Quantities
}

// ReadView reads the ledger once under its mutex.
func ReadView(ctx context.Context, root string) (*View, error) {
	owner, err := Identity(os.Getpid())
	if err != nil {
		return nil, err
	}
	view := &View{owner: owner, alive: map[Proof]bool{}}
	err = withState(ctx, root, false, func(s *state) error {
		data, err := json.Marshal(s)
		if err != nil {
			return err
		}
		view.state = &state{}
		return json.Unmarshal(data, view.state)
	})
	return view, err
}

// Evaluate decides a prospective admission against the view exactly as the
// ledger would decide it now, without registering it. Residents not yet
// charged are prospective and need not exist as records.
func (v *View) Evaluate(claims, capacities Quantities, parent *string, exclusive bool, residents []Resident) (Evaluation, error) {
	data, err := json.Marshal(v.state)
	if err != nil {
		return Evaluation{}, err
	}
	clone := &state{}
	if err := json.Unmarshal(data, clone); err != nil {
		return Evaluation{}, err
	}
	token, err := freshToken()
	if err != nil {
		return Evaluation{}, err
	}
	mode := "shared"
	if exclusive {
		mode = "exclusive"
	}
	decision := &Decision{}
	request := Request{Operation: "admit", Token: token, Owner: v.owner, Mode: mode, Claims: complete(claims), Capacities: complete(capacities),
		Parent: parent, Residents: residents, alive: v.alive, diagnostic: decision}
	granted, err := mutate(clone, request)
	if errors.Is(err, ErrConflict) {
		return Evaluation{}, err
	}
	return Evaluation{Admissible: granted, Deficits: decision.Deficits}, err
}

// ProspectiveResident names an allocation that does not exist yet, for
// evaluation only; the ledger never sees its reference.
func ProspectiveResident(claims, capacities Quantities, parent *string) (Resident, error) {
	token, err := freshToken()
	if err != nil {
		return Resident{}, err
	}
	return Resident{Token: token, Reference: "/prospective/" + token, Claims: complete(claims), Capacities: complete(capacities), Parent: parent}, nil
}

// TryAcquireWithPreparedResidents makes one admission attempt. A denied
// request is removed at once, so it keeps no queue position and its next
// attempt competes afresh; ranking belongs to the caller's scheduler.
func TryAcquireWithPreparedResidents(ctx context.Context, root string, claims, capacities Quantities, parent *string, exclusive bool, prepare func(context.Context) ([]Resident, error)) (*Lease, bool, error) {
	if root == "" {
		root = DefaultRoot()
	}
	owner, err := Identity(os.Getpid())
	if err != nil {
		return nil, false, err
	}
	token, err := freshToken()
	if err != nil {
		return nil, false, err
	}
	mode := "shared"
	if exclusive {
		mode = "exclusive"
	}
	lease := &Lease{Token: token, root: root}
	request := Request{Root: root, Operation: "admit", Token: token, Owner: owner, Mode: mode, Claims: complete(claims), Capacities: complete(capacities), Parent: parent}
	if prepare != nil {
		done := Phase(ctx, "logical_slot")
		request.Residents, err = prepare(ctx)
		done()
		if err != nil {
			return nil, false, err
		}
	}
	done := Phase(ctx, "admission_transaction")
	granted, err := observedTransaction(ctx, request)
	done()
	if err == nil && granted {
		return lease, true, nil
	}
	cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if releaseErr := lease.Release(cleanup); releaseErr != nil {
		return nil, false, errors.Join(err, ErrRelease, releaseErr)
	}
	return nil, false, err
}

// Watch reports changes of the ledger under root: one coalesced signal per
// burst of state replacements, by any process. It stops when ctx ends. A
// host without inotify falls back to a periodic signal.
func Watch(ctx context.Context, root string) (<-chan struct{}, error) {
	changes := make(chan struct{}, 1)
	signal := func() {
		select {
		case changes <- struct{}{}:
		default:
		}
	}
	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC | unix.IN_NONBLOCK)
	if err == nil {
		if _, err = unix.InotifyAddWatch(fd, root, unix.IN_MOVED_TO|unix.IN_CLOSE_WRITE|unix.IN_CREATE); err != nil {
			unix.Close(fd)
		}
	}
	if err != nil {
		go func() {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					signal()
				}
			}
		}()
		return changes, nil
	}
	go func() {
		defer unix.Close(fd)
		buffer := make([]byte, 4096)
		poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		for ctx.Err() == nil {
			n, err := unix.Poll(poll, 250)
			if err != nil && err != unix.EINTR {
				return
			}
			if n <= 0 {
				continue
			}
			for {
				if _, err := unix.Read(fd, buffer); err != nil {
					break
				}
			}
			signal()
		}
	}()
	return changes, nil
}

func freshToken() (string, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(nonce[:]), nil
}

func complete(input Quantities) Quantities {
	out := Quantities{}
	for _, key := range resources {
		out[key] = input[key]
	}
	return out
}
