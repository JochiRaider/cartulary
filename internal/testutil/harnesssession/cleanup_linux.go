//go:build linux

package harnesssession

import (
	"context"
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

type cleanupAttempt struct {
	done chan struct{}
	err  error
}

// AttemptCleanup applies each physical owner's deadline once in this runtime.
// A recovery in a new process receives a new ledger. An expired sibling or a
// cancelled acquisition cannot consume another resource's cleanup budget.
func (r *Runtime) AttemptCleanup(ctx context.Context, kind string, target any, fn func(context.Context) error) error {
	maximum, known := cleanupDeadlinesMS[kind]
	if !known {
		return harnesslocal.ErrUnsafe
	}
	data, err := json.Marshal(target)
	if err != nil {
		return err
	}
	resource := Resource{Kind: kind, Target: data, State: "acquired", RunID: r.RunID, LeaseID: r.LeaseID}
	if err := r.checkResource(resource); err != nil {
		return err
	}
	// Attempt and ready receipts are the same proxy instance's authority.
	if kind == "object_store_proxy" {
		file, err := resource.Path()
		if err != nil {
			return err
		}
		data, _ = json.Marshal(filepath.Dir(file))
	}
	// Equivalent structured proofs must share the same attempt regardless of
	// the field ordering used by their producer.
	var canonical any
	if err := json.Unmarshal(data, &canonical); err != nil {
		return err
	}
	data, err = json.Marshal(canonical)
	if err != nil {
		return err
	}
	key := kind + ":" + string(data)
	r.cleanupMu.Lock()
	if prior := r.cleanupAttempts[key]; prior != nil {
		r.cleanupMu.Unlock()
		<-prior.done
		return prior.err
	}
	if r.cleanupAttempts == nil {
		r.cleanupAttempts = map[string]*cleanupAttempt{}
	}
	attempt := &cleanupAttempt{done: make(chan struct{})}
	r.cleanupAttempts[key] = attempt
	r.cleanupMu.Unlock()
	step, done := context.WithTimeout(context.WithoutCancel(ctx), time.Duration(maximum)*time.Millisecond)
	defer done()
	defer close(attempt.done)
	attempt.err = fn(step)
	if attempt.err == nil {
		attempt.err = step.Err()
	}
	return attempt.err
}
