//go:build linux

package harnessadmission

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAdmissionViewAndSingleAttempt(t *testing.T) {
	t.Run("a view decides as the ledger would, without writing", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		capacities := testQuantities(3)
		capacities["memory_mb"] = 1024
		holder := request(t, root, "a", 2, 3)
		holder.Capacities = capacities
		holder.Claims["memory_mb"] = 1000
		apply(t, holder, true)
		before, err := os.ReadFile(filepath.Join(root, "state.json"))
		if err != nil {
			t.Fatal(err)
		}
		view, err := ReadView(t.Context(), root)
		if err != nil {
			t.Fatal(err)
		}
		fits, err := view.Evaluate(testQuantities(1), capacities, nil, false, nil)
		if err != nil || !fits.Admissible {
			t.Fatal("a fitting request was refused", fits, err)
		}
		over, err := view.Evaluate(testQuantities(2), capacities, nil, false, nil)
		if err != nil || over.Admissible || over.Deficits["cpu"] != 1 {
			t.Fatal("an excess request was admitted or misreported", over, err)
		}
		resident, err := ProspectiveResident(Quantities{"memory_mb": 64}, capacities, nil)
		if err != nil {
			t.Fatal(err)
		}
		withResident, err := view.Evaluate(testQuantities(1), capacities, nil, false, []Resident{resident})
		if err != nil || withResident.Admissible || withResident.Deficits["memory_mb"] == 0 {
			t.Fatal("a prospective resident was not charged", withResident, err)
		}
		after, err := os.ReadFile(filepath.Join(root, "state.json"))
		if err != nil || string(after) != string(before) {
			t.Fatal("evaluation wrote the ledger", err)
		}
	})
	t.Run("a denied single attempt keeps no queue position", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		holder := request(t, root, "a", 2, 3)
		apply(t, holder, true)
		denied, granted, err := TryAcquireWithPreparedResidents(t.Context(), root, testQuantities(2), testQuantities(3), nil, false, nil)
		if err != nil || granted || denied != nil {
			t.Fatal("an excess attempt was granted", granted, err)
		}
		view, err := ReadView(t.Context(), root)
		if err != nil || len(view.state.Leases) != 1 {
			t.Fatal("a denied attempt stayed queued", err)
		}
		lease, granted, err := TryAcquireWithPreparedResidents(t.Context(), root, testQuantities(1), testQuantities(3), nil, false, nil)
		if err != nil || !granted {
			t.Fatal("a fitting attempt was denied", err)
		}
		if err := lease.Release(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("waiters wake on a change, and an unchanged ledger is not rewritten", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		holder := request(t, root, "a", 1, 1)
		apply(t, holder, true)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		changes, err := Watch(ctx, root)
		if err != nil {
			t.Fatal(err)
		}
		// A request the ledger already holds changes nothing.
		apply(t, holder, true)
		select {
		case <-changes:
			t.Fatal("an unchanged ledger signalled a change")
		case <-time.After(400 * time.Millisecond):
		}
		apply(t, Request{Root: root, Operation: "release", Token: holder.Token}, false)
		select {
		case <-changes:
		case <-time.After(5 * time.Second):
			t.Fatal("a release woke no waiter")
		}
	})
	t.Run("a blocked waiter is admitted promptly after a release", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		holder := request(t, root, "a", 1, 1)
		apply(t, holder, true)
		admitted := make(chan error, 1)
		go func() {
			lease, err := Acquire(context.Background(), root, testQuantities(1), testQuantities(1), nil, false)
			if err == nil {
				err = lease.Release(context.Background())
			}
			admitted <- err
		}()
		time.Sleep(200 * time.Millisecond)
		start := time.Now()
		apply(t, Request{Root: root, Operation: "release", Token: holder.Token}, false)
		select {
		case err := <-admitted:
			if err != nil {
				t.Fatal(err)
			}
			if waited := time.Since(start); waited > 900*time.Millisecond {
				t.Fatal("the waiter polled instead of waking", waited)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("the waiter was never admitted")
		}
	})
}
