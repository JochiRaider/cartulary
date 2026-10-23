//go:build linux

package harnessadmission

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func testQuantities(cpu float64) Quantities {
	q := Quantities{}
	for _, key := range resources {
		q[key] = 0
	}
	q["cpu"] = cpu
	return q
}
func request(t *testing.T, root, token string, claim, capacity float64) Request {
	t.Helper()
	owner, err := Identity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	return Request{Root: root, Operation: "admit", Token: strings.Repeat(token, 32), Owner: owner, Mode: "shared", Claims: testQuantities(claim), Capacities: testQuantities(capacity)}
}
func apply(t *testing.T, r Request, expected bool) {
	t.Helper()
	value, err := Transact(context.Background(), r)
	if err != nil || value != expected {
		t.Fatalf("operation %s token %s: %v %v, expected %v", r.Operation, r.Token, value, err, expected)
	}
}
func TestAdmissionWeightedTree(t *testing.T) {
	t.Run("exact_release_retains_replacements_and_rejects_stale_sets", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		owner := request(t, root, "a", 1, 4)
		owner.Capacities["memory_mb"] = 1024
		first := Resident{Token: strings.Repeat("b", 32), Reference: filepath.Join(root, "first.json"), Claims: testQuantities(0), Capacities: owner.Capacities}
		first.Claims["memory_mb"] = 300
		second := first
		second.Token = strings.Repeat("c", 32)
		second.Reference = filepath.Join(root, "second.json")
		owner.Residents = []Resident{first, second}
		apply(t, owner, true)
		apply(t, Request{Root: root, Operation: "release", Token: owner.Token}, false)
		pending := request(t, root, "d", 1, 4)
		pending.Capacities = owner.Capacities
		pending.Claims["memory_mb"] = 500
		pending.Residents = []Resident{first}
		apply(t, pending, false)
		query := Request{Root: root, Operation: "retirement_required", Token: pending.Token, Residents: []Resident{first}, retirementDiagnostic: &RetirementDecision{}}
		apply(t, query, false)
		if query.retirementDiagnostic.Before["memory_mb"] != 76 || query.retirementDiagnostic.After["memory_mb"] != 76 {
			t.Fatal(query.retirementDiagnostic)
		}
		query.Residents = []Resident{second}
		apply(t, query, true)
		if len(query.retirementDiagnostic.After) != 0 {
			t.Fatal(query.retirementDiagnostic)
		}
		stale := second
		stale.Token = strings.Repeat("e", 32)
		for _, residents := range [][]Resident{nil, {stale}, {second, second}} {
			query.Residents = residents
			if _, err := Transact(t.Context(), query); !errors.Is(err, ErrConflict) {
				t.Fatal("invalid set accepted", err)
			}
		}
		query.Residents = []Resident{second}
		query.Claims = second.Claims
		if _, err := Transact(t.Context(), query); !errors.Is(err, ErrConflict) {
			t.Fatal("quantity query accepted", err)
		}
		if err := ReleaseResident(t.Context(), root, second); err != nil {
			t.Fatal(err)
		}
		query.Claims = nil
		if _, err := Transact(t.Context(), query); !errors.Is(err, ErrConflict) {
			t.Fatal("settled generation accepted", err)
		}
		apply(t, pending, true)
	})
	t.Run("ancestor_envelope_masks_apparent_retirement_savings", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		parent := request(t, root, "a", 1, 4)
		parent.Claims["memory_mb"] = 800
		parent.Capacities["memory_mb"] = 1200
		apply(t, parent, true)
		child := request(t, root, "b", 1, 4)
		child.Parent = &parent.Token
		child.Capacities = parent.Capacities
		nested := Resident{Token: strings.Repeat("c", 32), Reference: filepath.Join(root, "nested.json"), Claims: testQuantities(0), Capacities: parent.Capacities, Parent: &parent.Token}
		nested.Claims["memory_mb"] = 200
		child.Residents = []Resident{nested}
		apply(t, child, true)
		unrelated := request(t, root, "d", 1, 4)
		unrelated.Capacities = parent.Capacities
		outside := Resident{Token: strings.Repeat("e", 32), Reference: filepath.Join(root, "outside.json"), Claims: testQuantities(0), Capacities: parent.Capacities}
		outside.Claims["memory_mb"] = 300
		unrelated.Residents = []Resident{outside}
		apply(t, unrelated, true)
		apply(t, Request{Root: root, Operation: "release", Token: child.Token}, false)
		apply(t, Request{Root: root, Operation: "release", Token: unrelated.Token}, false)
		pending := request(t, root, "f", 1, 4)
		pending.Parent = &parent.Token
		pending.Capacities = testQuantities(4)
		pending.Capacities["memory_mb"] = 1000
		pending.diagnostic = &Decision{}
		apply(t, pending, false)
		if pending.diagnostic.PersistentDeficits["memory_mb"] != 100 {
			t.Fatal(pending.diagnostic)
		}
		// The parent envelope masks all released memory; overlap alone is not relief.
		apply(t, Request{Root: root, Operation: "retirement_required", Token: pending.Token, Residents: []Resident{nested}}, false)
	})
	t.Run("diagnostic_facts_come_from_admission_and_do_not_disclose_capabilities", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		holder := request(t, root, "a", 2, 2)
		holder.Capacities["memory_mb"] = 512
		apply(t, holder, true)
		candidate := request(t, root, "b", 1, 2)
		candidate.Capacities = holder.Capacities
		var records []Observation
		clock := float64(0)
		ctx := WithObservation(t.Context(), "unit", func() float64 { clock++; return clock }, func(o Observation) { records = append(records, o) })
		granted, err := observedTransaction(ctx, candidate)
		if err != nil || granted || len(records) != 1 {
			t.Fatal(granted, err, records)
		}
		d := records[0].Decision
		if d == nil || d.Deficits["cpu"] != 1 || len(d.PersistentDeficits) != 0 || len(d.Holders) != 1 || d.QueuePosition != 1 {
			t.Fatalf("wrong transient facts: %+v", d)
		}
		encoded, _ := json.Marshal(records)
		if strings.Contains(string(encoded), holder.Token) || strings.Contains(string(encoded), root) {
			t.Fatal("private identity leaked")
		}
		apply(t, Request{Root: root, Operation: "release", Token: holder.Token}, false)
		granted, err = observedTransaction(ctx, candidate)
		if err != nil || !granted || !records[1].Decision.Granted || len(records[1].Decision.Deficits) != 0 {
			t.Fatal(granted, err, records)
		}
		end := Phase(ctx, "admission_wait")
		end()
		refs := []Resident{{Token: holder.Token}}
		Retirement(ctx, refs, "capacity_pressure")(errors.New("private recovery path"))
		if records[2].DurationMS != 1 || records[3].Outcome != "unresolved" || records[3].DurationMS != 1 {
			t.Fatal(records)
		}
		apply(t, Request{Root: root, Operation: "release", Token: candidate.Token}, false)
	})
	t.Run("independent_retirement_invalidates_broad_snapshot_not_scoped_cleanup", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		initial := request(t, root, "a", 1, 2)
		initial.Capacities["memory_mb"] = 512
		first := Resident{Token: strings.Repeat("b", 32), Reference: filepath.Join(root, "first.json"), Claims: testQuantities(0), Capacities: initial.Capacities}
		first.Claims["memory_mb"] = 80
		second := first
		second.Token, second.Reference = strings.Repeat("c", 32), filepath.Join(root, "second.json")
		initial.Residents = []Resident{first, second}
		apply(t, initial, true)
		apply(t, Request{Root: root, Operation: "release", Token: initial.Token}, false)
		// Capture both exact identities, then settle an independent allocation before
		// cleanup admission. This is the interleaving a whole-invocation scan admits.
		stale := request(t, root, "d", 1, 2)
		stale.Capacities, stale.Mode, stale.Residents = initial.Capacities, "drain", []Resident{first, second}
		if err := ReleaseResident(t.Context(), root, second); err != nil {
			t.Fatal(err)
		}
		if _, err := Transact(t.Context(), stale); !errors.Is(err, ErrConflict) {
			t.Fatal("vanished identity accepted", err)
		}
		scoped := stale
		scoped.Token, scoped.Residents = strings.Repeat("e", 32), []Resident{first}
		apply(t, scoped, true)
		if err := ReleaseResident(t.Context(), root, first); err != nil {
			t.Fatal(err)
		}
		apply(t, Request{Root: root, Operation: "release", Token: scoped.Token}, false)
	})

	t.Run("prepared_waiter_keeps_priority_and_refreshes_only_pending_residents", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		holder := request(t, root, "a", 2, 4)
		holder.Capacities["memory_mb"] = 512
		apply(t, holder, true)
		later := request(t, root, "b", 1, 4)
		later.Capacities = holder.Capacities
		old := Resident{Token: strings.Repeat("c", 32), Reference: filepath.Join(root, "old.json"), Claims: testQuantities(0), Capacities: holder.Capacities}
		old.Claims["memory_mb"] = 80
		fresh := old
		fresh.Token, fresh.Reference = strings.Repeat("d", 32), filepath.Join(root, "fresh.json")
		preparations, misses := 0, 0
		var pendingToken string
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()
		lease, err := AcquireWithPreparedResidents(ctx, root, testQuantities(3), holder.Capacities, nil, false, func(context.Context) ([]Resident, error) {
			preparations++
			if preparations == 1 {
				return []Resident{old}, nil
			}
			return []Resident{fresh}, nil
		}, func(_ context.Context, pending *Lease) error {
			misses++
			pendingToken = pending.Token
			apply(t, later, false) // A fitting new arrival cannot overtake us.
			apply(t, Request{Root: root, Operation: "release", Token: holder.Token}, false)
			apply(t, later, false) // The same priority survives before reacquisition.
			return nil
		})
		if err != nil || lease == nil || lease.Token != pendingToken || preparations != 2 || misses != 1 {
			t.Fatal("pending preparation lost identity or failed to refresh", lease, preparations, misses, err)
		}
		apply(t, Request{Root: root, Operation: "resident_granted", Token: old.Token, Reference: old.Reference}, false)
		apply(t, Request{Root: root, Operation: "resident_granted", Token: fresh.Token, Reference: fresh.Reference}, true)
		changed := request(t, root, "e", 3, 4)
		changed.Token, changed.Capacities, changed.Residents = lease.Token, holder.Capacities, []Resident{old}
		if _, err := Transact(t.Context(), changed); !errors.Is(err, ErrConflict) {
			t.Fatal("granted resident identity changed", err)
		}
		apply(t, later, true) // Only the granted request consumes active capacity.
		if err := errors.Join(lease.Release(t.Context()), ReleaseResident(t.Context(), root, fresh)); err != nil {
			t.Fatal(err)
		}
		apply(t, Request{Root: root, Operation: "release", Token: later.Token}, false)
	})
	t.Run("impossible_startup_rejects_instead_of_waiting_on_parent", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		parent := request(t, root, "a", 1, 4)
		parent.Capacities["memory_mb"] = 1024
		parent.Claims["memory_mb"] = 800
		apply(t, parent, true)
		child := request(t, root, "b", 1, 4)
		child.Parent = &parent.Token
		child.Capacities = parent.Capacities
		resident := Resident{Token: strings.Repeat("c", 32), Reference: filepath.Join(root, "resident.json"), Claims: testQuantities(0), Capacities: child.Capacities}
		resident.Claims["memory_mb"] = 256
		child.Residents = []Resident{resident}
		if _, err := Transact(t.Context(), child); !errors.Is(err, ErrConflict) {
			t.Fatal("startup waited on enclosing reservation", err)
		}
		child.Parent = nil
		child.Claims["memory_mb"] = 800
		if _, err := Transact(t.Context(), child); !errors.Is(err, ErrConflict) {
			t.Fatal("impossible startup and residency accepted", err)
		}
	})

	t.Run("multiple_clusters_retain_volume_capability_and_add_memory", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		first := request(t, root, "a", 1, 4)
		first.Capacities["memory_mb"], first.Capacities["volume"] = 1024, 1
		for _, token := range []string{"b", "c"} {
			resident := Resident{Token: strings.Repeat(token, 32), Reference: filepath.Join(root, token+".json"), Claims: testQuantities(0), Capacities: first.Capacities}
			resident.Claims["memory_mb"], resident.Claims["volume"] = 256, 1
			first.Residents = append(first.Residents, resident)
		}
		apply(t, first, true)
		work := request(t, root, "d", 1, 4)
		work.Capacities = first.Capacities
		work.Claims["memory_mb"], work.Claims["volume"] = 600, 1
		work.diagnostic = &Decision{}
		apply(t, work, false)
		if work.diagnostic.Deficits["memory_mb"] != 88 || work.diagnostic.PersistentDeficits["memory_mb"] != 88 || work.diagnostic.Deficits["volume"] != 0 || len(work.diagnostic.Residents) != 2 {
			t.Fatalf("resident decision facts: %+v", work.diagnostic)
		}
		evaluation := &RetirementDecision{}
		apply(t, Request{Root: root, Operation: "retirement_required", Token: work.Token, Residents: []Resident{first.Residents[0]}, retirementDiagnostic: evaluation}, true)
		if evaluation.Before["volume"] != 0 || evaluation.After["volume"] != 0 || evaluation.Before["memory_mb"] != 88 || evaluation.After["memory_mb"] != 0 {
			t.Fatal("capability resource became additive", evaluation)
		}
		if err := ReleaseResident(t.Context(), root, first.Residents[0]); err != nil {
			t.Fatal(err)
		}
		apply(t, work, true)
		unavailable := request(t, filepath.Join(t.TempDir(), "admission"), "e", 1, 4)
		unavailable.Claims["volume"] = 1
		if _, err := Transact(t.Context(), unavailable); !errors.Is(err, ErrConflict) {
			t.Fatal("unproven volume admitted", err)
		}
	})

	for _, nested := range []bool{false, true} {
		t.Run("transient_pressure_preserves_waiter_priority_"+strconv.FormatBool(nested), func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "admission")
			holder := request(t, root, "a", 2, 4)
			apply(t, holder, true)
			parent := request(t, root, "b", 1, 4)
			if nested {
				apply(t, parent, true)
			}
			older := request(t, root, "c", 3, 4)
			if nested {
				older.Parent = &parent.Token
			}
			apply(t, older, false)
			later := request(t, root, "d", 1, 4)
			later.diagnostic = &Decision{}
			apply(t, later, false) // Fits now, but must not starve the older request.
			if len(later.diagnostic.Deficits) != 0 || len(later.diagnostic.PriorityBlockers) != 1 || len(later.diagnostic.ExclusiveBlockers) != 0 {
				t.Fatalf("priority confused with capacity: %+v", later.diagnostic)
			}
			apply(t, Request{Root: root, Operation: "release", Token: holder.Token}, false)
			apply(t, later, false) // Priority persists until the waiter takes admission.
			apply(t, older, true)
			apply(t, later, true) // Remaining capacity is usable after that grant.
		})
	}
	t.Run("retirement_preserves_required_ancestor_claims", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		parent := request(t, root, "a", 1, 2)
		parent.Claims["memory_mb"] = 500
		parent.Capacities["memory_mb"] = 1024
		resident := Resident{Token: strings.Repeat("b", 32), Reference: filepath.Join(root, "retained.json"), Claims: testQuantities(0), Capacities: parent.Capacities}
		resident.Claims["memory_mb"] = 80
		parent.Residents = []Resident{resident}
		apply(t, parent, true)
		child := request(t, root, "c", 1, 2)
		child.Parent = &parent.Token
		child.Claims["memory_mb"] = 120
		child.Capacities["memory_mb"] = 512
		child.diagnostic = &Decision{}
		apply(t, child, false)
		if child.diagnostic.PersistentDeficits["memory_mb"] != 68 || len(child.diagnostic.Ancestors) != 1 || child.diagnostic.Ancestors[0].Claims["memory_mb"] != 500 {
			t.Fatalf("ancestor envelope omitted: %+v", child.diagnostic)
		}
		apply(t, Request{Root: root, Operation: "retirement_required", Token: child.Token, Residents: []Resident{resident}}, true)
		if err := ReleaseResident(t.Context(), root, resident); err != nil {
			t.Fatal(err)
		}
		apply(t, child, true)
		apply(t, Request{Root: root, Operation: "release", Token: child.Token}, false)
		apply(t, Request{Root: root, Operation: "release", Token: parent.Token}, false)
	})
	t.Run("retirement_distinguishes_transient_work_from_resident_pressure", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		owner := request(t, root, "a", 1, 2)
		owner.Capacities["memory_mb"] = 512
		owner.Capacities["port_lane"] = 16
		owner.Claims["memory_mb"] = 400
		resident := Resident{Token: strings.Repeat("b", 32), Reference: filepath.Join(root, "retained.json"), Claims: testQuantities(0), Capacities: owner.Capacities}
		resident.Claims["memory_mb"], resident.Claims["port_lane"] = 80, 14
		unrelated := Resident{Token: strings.Repeat("e", 32), Reference: filepath.Join(root, "unrelated.json"), Claims: testQuantities(0), Capacities: owner.Capacities}
		unrelated.Claims["memory_mb"] = 16
		owner.Residents = []Resident{resident, unrelated}
		apply(t, owner, true)
		pending := request(t, root, "c", 1, 2)
		pending.Capacities = owner.Capacities
		pending.Claims["memory_mb"] = 120
		apply(t, pending, false)
		query := Request{Root: root, Operation: "retirement_required", Token: pending.Token, Residents: []Resident{resident}}
		apply(t, query, false) // Completing the active work frees enough memory.
		apply(t, Request{Root: root, Operation: "release", Token: pending.Token}, false)
		pending.Token = strings.Repeat("d", 32)
		pending.Claims["port_lane"] = 3
		apply(t, pending, false)
		query.Token = pending.Token
		apply(t, query, true) // The physical port shortage survives active work.
		query.Residents = []Resident{unrelated}
		apply(t, query, false) // Retiring unrelated memory cannot free a port.
		apply(t, Request{Root: root, Operation: "resident_granted", Token: resident.Token, Reference: resident.Reference}, true)
		apply(t, Request{Root: root, Operation: "release", Token: owner.Token}, false)
		if err := ReleaseResident(t.Context(), root, resident); err != nil {
			t.Fatal(err)
		}
		query.Residents = []Resident{resident}
		if _, err := Transact(t.Context(), query); !errors.Is(err, ErrConflict) {
			t.Fatal("stale retirement identity accepted", err)
		}
		apply(t, pending, true)
		if _, err := (&Lease{root: root, Token: pending.Token}).NeedsRetirement(t.Context(), []Resident{resident}); !errors.Is(err, ErrConflict) {
			t.Fatal("granted activity used as a pending eviction decision", err)
		}
	})
	t.Run("resident_pressure_allows_fitting_consumers_but_preserves_exclusive_priority", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		owner := request(t, root, "a", 1, 2)
		owner.Capacities["port_lane"] = 16
		resident := Resident{Token: strings.Repeat("b", 32), Reference: filepath.Join(root, "retained.json"), Claims: testQuantities(0), Capacities: owner.Capacities}
		resident.Claims["port_lane"] = 14
		owner.Residents = []Resident{resident}
		apply(t, owner, true)
		apply(t, Request{Root: root, Operation: "release", Token: owner.Token}, false)
		probe := testQuantities(1)
		probe["port_lane"] = 3
		interrupted := errors.New("preparation interrupted")
		if lease, err := AcquireWithResidents(t.Context(), root, probe, owner.Capacities, nil, false, nil, func(context.Context, *Lease) error { return interrupted }); !errors.Is(err, interrupted) || lease != nil {
			t.Fatal("interrupted queue retained an active lease", lease, err)
		}
		probeQuiet := request(t, root, "9", 1, 2)
		probeQuiet.Capacities = owner.Capacities
		probeQuiet.Mode = "exclusive"
		apply(t, probeQuiet, true) // Interrupted preparation left no queued reservation.
		apply(t, Request{Root: root, Operation: "release", Token: probeQuiet.Token}, false)
		large := request(t, root, "c", 1, 2)
		large.Capacities = owner.Capacities
		large.Claims["port_lane"] = 3
		apply(t, large, false)
		consumer := request(t, root, "d", 1, 2)
		consumer.Capacities = owner.Capacities
		consumer.Claims["port_lane"] = 1
		consumer.Residents = []Resident{resident}
		apply(t, consumer, true)
		quiet := request(t, root, "e", 1, 2)
		quiet.Capacities = owner.Capacities
		quiet.Mode = "exclusive"
		apply(t, quiet, false)
		later := request(t, root, "f", 0, 2)
		later.Capacities = owner.Capacities
		apply(t, later, false)
		if err := ReleaseResident(t.Context(), root, resident); err != nil {
			t.Fatal(err)
		}
		apply(t, Request{Root: root, Operation: "release", Token: consumer.Token}, false)
		apply(t, large, true)
		apply(t, Request{Root: root, Operation: "release", Token: large.Token}, false)
		apply(t, quiet, true)
		apply(t, later, false)
		apply(t, Request{Root: root, Operation: "release", Token: quiet.Token}, false)
		apply(t, later, true)
	})
	t.Run("dead_owner_retains_physical_capacity_and_allows_bounded_drain", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		process := exec.Command("/bin/sh", "-c", "read value")
		input, err := process.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := process.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { input.Close(); process.Process.Kill(); process.Wait() })
		parent := request(t, root, "a", 1, 2)
		parent.Owner, err = Identity(process.Process.Pid)
		if err != nil {
			t.Fatal(err)
		}
		parent.Capacities["memory_mb"] = 256
		parent.Capacities["process"] = 2
		apply(t, parent, true)
		work := request(t, root, "b", 1, 2)
		work.Parent = &parent.Token
		work.Capacities = parent.Capacities
		allocation := Resident{Token: strings.Repeat("c", 32), Reference: filepath.Join(root, "allocation.json"), Claims: testQuantities(0), Capacities: parent.Capacities, Parent: &parent.Token}
		allocation.Claims["memory_mb"] = 80
		work.Residents = []Resident{allocation}
		apply(t, work, true)
		apply(t, Request{Root: root, Operation: "release", Token: work.Token}, false)
		input.Close()
		process.Wait()
		contender := request(t, root, "d", 1, 2)
		contender.Capacities = parent.Capacities
		contender.Claims["memory_mb"] = 177
		apply(t, contender, false)
		drain := request(t, root, "e", 1, 2)
		drain.Capacities = parent.Capacities
		drain.Parent = &parent.Token
		drain.Mode = "drain"
		drain.Residents = []Resident{allocation}
		drain.Claims["memory_mb"] = 64
		drain.Claims["process"] = 1
		apply(t, drain, true) // The queued contender cannot obstruct cleanup.
		if err := ReleaseResident(t.Context(), root, allocation); err != nil {
			t.Fatal(err)
		}
		apply(t, Request{Root: root, Operation: "release", Token: drain.Token}, false)
		apply(t, contender, true)
	})
	t.Run("resident_ownership_is_atomic_durable_and_not_activity", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		first := request(t, root, "a", 1, 1)
		first.Capacities["memory_mb"] = 200
		allocation := Resident{Token: strings.Repeat("c", 32), Reference: filepath.Join(root, "owned-allocation.json"), Claims: testQuantities(0), Capacities: first.Capacities}
		allocation.Claims["memory_mb"] = 80
		first.Residents = []Resident{allocation}
		apply(t, first, true)
		second := request(t, root, "b", 1, 1)
		second.Capacities = first.Capacities
		second.Residents = first.Residents
		apply(t, second, false)
		apply(t, Request{Root: root, Operation: "release", Token: first.Token}, false)
		apply(t, second, true) // Reborrowing charges the physical allocation once.
		apply(t, Request{Root: root, Operation: "release", Token: second.Token}, false)
		quiet := request(t, root, "d", 1, 1)
		quiet.Capacities = first.Capacities
		quiet.Mode = "exclusive"
		quiet.Claims["memory_mb"] = 121
		apply(t, quiet, false) // Idle infrastructure remains resident.
		apply(t, Request{Root: root, Operation: "release", Token: quiet.Token}, false)
		quiet.Token = strings.Repeat("e", 32)
		quiet.Claims["memory_mb"] = 120
		apply(t, quiet, true) // Residents alone hold no activity fence.
		apply(t, Request{Root: root, Operation: "release", Token: quiet.Token}, false)
		if err := ReleaseResident(context.Background(), root, Resident{Token: allocation.Token, Reference: filepath.Join(root, "foreign"), Claims: allocation.Claims, Capacities: allocation.Capacities}); !errors.Is(err, ErrConflict) {
			t.Fatal("foreign release accepted", err)
		}
		for range 2 {
			if err := ReleaseResident(context.Background(), root, allocation); err != nil {
				t.Fatal(err)
			}
		}
		last := request(t, root, "f", 1, 1)
		last.Capacities = first.Capacities
		last.Claims["memory_mb"] = 200
		apply(t, last, true)
	})
	t.Run("terminal child binding retains exact proof and reservation", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		parent := request(t, root, "a", 1, 1)
		apply(t, parent, true)
		child := exec.Command("/bin/sh", "-c", "read value")
		input, err := child.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := child.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = input.Close(); _ = child.Process.Kill(); _ = child.Wait() })
		proof, err := Identity(child.Process.Pid)
		if err != nil {
			t.Fatal(err)
		}
		if live, err := Alive(proof); err != nil || !live {
			t.Fatalf("child was not live before the transition: %v", err)
		}
		// End the child after proof capture, but pin its identity until the bind
		// transaction completes. No sleep or probabilistic race is involved.
		if err := input.Close(); err != nil {
			t.Fatal(err)
		}
		for {
			var info unix.Siginfo
			err := unix.Waitid(unix.P_PID, child.Process.Pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
			if err == unix.EINTR {
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			break
		}
		binding := Request{Root: root, Operation: "bind", Token: parent.Token, Guard: proof}
		apply(t, binding, true)
		apply(t, binding, true)
		for _, change := range []func(*Request){
			func(r *Request) { r.Guard.Start += "0" },
			func(r *Request) { r.Guard.Boot += "-foreign" },
			func(r *Request) { r.Token = strings.Repeat("f", 32) },
		} {
			invalid := binding
			change(&invalid)
			if _, err := Transact(context.Background(), invalid); !errors.Is(err, ErrConflict) {
				t.Fatalf("invalid terminal binding did not reject: %v", err)
			}
		}
		waiter := request(t, root, "b", 1, 1)
		apply(t, waiter, false)
		apply(t, Request{Root: root, Operation: "release", Token: parent.Token}, false)
		lease := &Lease{Token: parent.Token, root: root}
		if err := lease.Bind(child.Process.Pid); !errors.Is(err, ErrConflict) {
			t.Fatalf("terminal child bypassed released lease: %v", err)
		}
		apply(t, waiter, true)
	})
	t.Run("admitted descendants drain ahead of queued ancestor siblings", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		parent := request(t, root, "a", 2, 2)
		apply(t, parent, true)
		holder := request(t, root, "b", 2, 2)
		holder.Parent = &parent.Token
		apply(t, holder, true)
		waiter := request(t, root, "c", 1, 2)
		waiter.Parent = &parent.Token
		apply(t, waiter, false)
		child := request(t, root, "d", 2, 2)
		child.Parent = &holder.Token
		apply(t, child, true)
		queued := request(t, root, "e", 2, 2)
		queued.Parent = &holder.Token
		apply(t, queued, false)
		later := request(t, root, "f", 0, 2)
		later.Parent = &holder.Token
		apply(t, later, false) // An older sibling keeps priority while active work drains.
		grandchild := request(t, root, "1", 1, 2)
		grandchild.Parent = &child.Token
		apply(t, grandchild, true)
		apply(t, Request{Root: root, Operation: "release", Token: grandchild.Token}, false)
		apply(t, Request{Root: root, Operation: "release", Token: child.Token}, false)
		apply(t, queued, true)
		apply(t, later, true)
		apply(t, waiter, false) // The admitted holder still owns its reservation.
		apply(t, Request{Root: root, Operation: "release", Token: queued.Token}, false)
		apply(t, Request{Root: root, Operation: "release", Token: later.Token}, false)
		apply(t, Request{Root: root, Operation: "release", Token: holder.Token}, false)
		apply(t, waiter, true)
	})
	t.Run("independent admitted branches preserve credit and capacity", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		parent := request(t, root, "a", 2, 2)
		apply(t, parent, true)
		first := request(t, root, "b", 1, 2)
		first.Parent = &parent.Token
		apply(t, first, true)
		second := request(t, root, "c", 1, 2)
		second.Parent = &parent.Token
		apply(t, second, true)
		growing := request(t, root, "d", 2, 2)
		growing.Parent = &first.Token
		apply(t, growing, false)
		draining := request(t, root, "e", 1, 2)
		draining.Parent = &second.Token
		apply(t, draining, true)
		apply(t, growing, false)
		apply(t, Request{Root: root, Operation: "release", Token: draining.Token}, false)
		apply(t, Request{Root: root, Operation: "release", Token: second.Token}, false)
		apply(t, growing, true)
	})
	root := filepath.Join(t.TempDir(), "admission")
	parent := request(t, root, "a", 2, 2)
	apply(t, parent, true)
	first := request(t, root, "b", 1, 2)
	first.Parent = &parent.Token
	apply(t, first, true)
	second := request(t, root, "c", 1, 2)
	second.Parent = &parent.Token
	apply(t, second, true)
	third := request(t, root, "d", 1, 2)
	third.Parent = &parent.Token
	apply(t, third, false)
	// Parent credit is max(2, 1+1), not 2+1+1, and cannot be borrowed a third time.
	apply(t, Request{Root: root, Operation: "release", Token: first.Token}, false)
	apply(t, third, true)
	changed := third
	changed.Claims = testQuantities(0)
	if _, err := Transact(context.Background(), changed); err == nil {
		t.Fatal("conflicting token replay admitted")
	}
}
func TestAdmissionQuietOrdering(t *testing.T) {
	t.Run("bounded observation does not reorder workload or bypass quietness", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "admission")
		holder := request(t, root, "a", 2, 2)
		apply(t, holder, true)
		waiter := request(t, root, "b", 2, 2)
		apply(t, waiter, false)
		later := request(t, root, "c", 0, 2)
		apply(t, later, false)
		observer := request(t, root, "d", 0, 0)
		observer.Mode = "observation"
		apply(t, observer, true)
		apply(t, later, false)
		// Passive capacity values cannot lower the workload capacity ceiling.
		apply(t, Request{Root: root, Operation: "release", Token: holder.Token}, false)
		apply(t, waiter, true)
		apply(t, later, true)
		exclusive := request(t, root, "e", 0, 2)
		exclusive.Mode = "exclusive"
		apply(t, exclusive, false)
		next := request(t, root, "f", 0, 2)
		next.Mode = "observation"
		apply(t, next, false)
		apply(t, Request{Root: root, Operation: "release", Token: waiter.Token}, false)
		apply(t, Request{Root: root, Operation: "release", Token: later.Token}, false)
		apply(t, exclusive, false) // The admitted observation still owns its fence.
		apply(t, Request{Root: root, Operation: "release", Token: observer.Token}, false)
		apply(t, exclusive, true)
		apply(t, next, false)
		apply(t, Request{Root: root, Operation: "release", Token: exclusive.Token}, false)
		apply(t, next, true)
		for _, change := range []func(*Request){
			func(r *Request) { r.Claims = testQuantities(1) },
			func(r *Request) { r.Parent = &next.Token },
			func(r *Request) { r.Mode = "shared"; r.Parent = &next.Token },
			func(r *Request) { r.Operation = "bind"; r.Token = next.Token; r.Guard = next.Owner },
		} {
			invalid := request(t, root, "9", 0, 2)
			invalid.Mode = "observation"
			change(&invalid)
			if _, err := Transact(context.Background(), invalid); err == nil {
				t.Fatal("observation admitted workload, nesting or an inherited guard")
			}
		}
	})
	root := filepath.Join(t.TempDir(), "admission")
	first := request(t, root, "a", 0, 2)
	apply(t, first, true)
	quiet := request(t, root, "b", 0, 2)
	quiet.Mode = "exclusive"
	apply(t, quiet, false)
	later := request(t, root, "c", 0, 2)
	apply(t, later, false)
	apply(t, Request{Root: root, Operation: "release", Token: first.Token}, false)
	apply(t, quiet, true)
	apply(t, later, false)
	apply(t, Request{Root: root, Operation: "release", Token: quiet.Token}, false)
	apply(t, later, true)
}
func TestAdmissionStateRejection(t *testing.T) {
	root := filepath.Join(t.TempDir(), "admission")
	r := request(t, root, "a", 0, 2)
	apply(t, r, true)
	file := filepath.Join(root, "state.json")
	original, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{`{"sequence":0,"sequence":0,"leases":[]}`, `{"sequence":0,"leases":[],"extra":0}`, strings.Replace(string(original), `"granted":true`, `"granted":null`, 1), strings.Replace(string(original), `"cpu":0`, `"cpu":null`, 1), strings.Replace(string(original), `"guards":[]`, `"guards":null`, 1)} {
		if err := os.WriteFile(file, []byte(invalid), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Transact(context.Background(), r); err == nil {
			t.Fatalf("accepted invalid state: %s", invalid)
		}
	}
	if err := os.WriteFile(file, original, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Transact(ctx, r); err == nil {
		t.Fatal("ignored cancellation")
	}
	after, _ := os.ReadFile(file)
	if string(after) != string(original) {
		t.Fatal("cancelled transaction changed state")
	}
	ctx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if _, err := Transact(ctx, Request{Root: root, Operation: "release", Token: r.Token}); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(r)
	if _, err := ReadRequest(data); err != nil {
		t.Fatal(err)
	}
	invalid := strings.Replace(string(data), `"cpu":0`, `"cpu":null`, 1)
	if _, err := ReadRequest([]byte(invalid)); err == nil {
		t.Fatal("invented zero for null request claim")
	}
}
