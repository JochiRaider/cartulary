//go:build linux

package harnessadmission

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testQuantities(cpu float64) Quantities {
	return Quantities{"browser_stack": 0, "cpu": cpu, "io": 0, "memory_mb": 0, "process": 0}
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
