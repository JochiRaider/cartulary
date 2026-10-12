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
	alive, err := Alive(proof)
	if err != nil || !alive {
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
		for _, k := range []string{"browser_stack", "cpu", "io", "memory_mb", "process"} {
			out[k] = input[k]
		}
		return out
	}
	mode := "shared"
	if exclusive {
		mode = "exclusive"
	}
	request := Request{Root: root, Operation: "admit", Token: lease.Token, Owner: owner, Mode: mode, Claims: quantities(claims), Capacities: quantities(capacities), Parent: parent}
	for {
		admitted, err := Transact(ctx, request)
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
