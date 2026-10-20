//go:build linux

// Package harnessadmission owns the one host-wide weighted admission tree.
// Native and Node callers use this same transaction owner and mutex.
package harnessadmission

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

var ErrConflict = errors.New("host admission unavailable")
var resources = []string{"browser_stack", "cpu", "io", "memory_mb", "object_store", "port_lane", "postgres", "process", "service_stack", "volume"}

const maxInteger = 9007199254740991

var digits = regexp.MustCompile(`^[0-9]+$`)
var tokenPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

type Proof struct {
	Boot  string `json:"boot"`
	PID   int    `json:"pid"`
	Start string `json:"start"`
}
type Quantities map[string]float64

// Resident is a durable physical allocation, not permission to perform work.
// Reference binds it to the native resource journal until proven cleanup.
type Resident struct {
	Token      string     `json:"token"`
	Reference  string     `json:"reference"`
	Claims     Quantities `json:"claims"`
	Capacities Quantities `json:"capacities"`
	Parent     *string    `json:"parent"`
}
type Request struct {
	Residents  []Resident `json:"residents,omitempty"`
	Reference  string     `json:"reference,omitempty"`
	Root       string     `json:"root"`
	Operation  string     `json:"operation"`
	Token      string     `json:"token"`
	Owner      Proof      `json:"owner"`
	Mode       string     `json:"mode"`
	Claims     Quantities `json:"claims"`
	Capacities Quantities `json:"capacities"`
	Parent     *string    `json:"parent"`
	Guard      Proof      `json:"guard"`
}
type lease struct {
	Residents  []Resident `json:"residents"`
	Token      string     `json:"token"`
	Process    Proof      `json:"process"`
	Guards     []Proof    `json:"guards"`
	Sequence   float64    `json:"sequence"`
	Mode       string     `json:"mode"`
	Claims     Quantities `json:"claims"`
	Capacities Quantities `json:"capacities"`
	Parent     *string    `json:"parent"`
	Granted    bool       `json:"granted"`
	Released   bool       `json:"released"`
}
type state struct {
	Generation int        `json:"generation"`
	Residents  []Resident `json:"residents"`
	Sequence   float64    `json:"sequence"`
	Leases     []*lease   `json:"leases"`
}

func integer(n float64) bool { return n >= 0 && n <= maxInteger && math.Trunc(n) == n }
func quantities(q Quantities) bool {
	if len(q) != len(resources) {
		return false
	}
	for _, r := range resources {
		n, ok := q[r]
		if !ok || !integer(n) {
			return false
		}
	}
	return true
}
func validMode(mode string, claims Quantities, parent *string) bool {
	if mode == "shared" || mode == "exclusive" || mode == "drain" {
		return true
	}
	if mode != "observation" || parent != nil {
		return false
	}
	for _, value := range claims {
		if value != 0 {
			return false
		}
	}
	return true
}
func proofShape(p Proof) bool { return p.Boot != "" && p.PID > 1 && digits.MatchString(p.Start) }
func Identity(pid int) (Proof, error) {
	proof, _, err := processIdentity(pid)
	return proof, err
}

// Read lifetime and terminal state from one stat snapshot. In particular, an
// unreaped child retains an exact identity even after it stops being live.
func processIdentity(pid int) (Proof, bool, error) {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return Proof{}, false, err
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return Proof{}, false, err
	}
	end := strings.LastIndex(string(data), ") ")
	if end < 0 {
		return Proof{}, false, ErrConflict
	}
	f := strings.Fields(string(data)[end+2:])
	if len(f) < 20 || !digits.MatchString(f[19]) {
		return Proof{}, false, ErrConflict
	}
	return Proof{strings.TrimSpace(string(boot)), pid, f[19]}, f[0] != "Z" && f[0] != "X", nil
}
func Alive(p Proof) (bool, error) {
	if !proofShape(p) {
		return false, nil
	}
	actual, live, err := processIdentity(p.PID)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return actual == p && live, nil
}
func strict(data []byte, out any) error {
	if _, err := canonicaljson.DecodeStrict(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}
func quantityShape(raw json.RawMessage) bool {
	var q map[string]json.RawMessage
	if json.Unmarshal(raw, &q) != nil || len(q) != len(resources) {
		return false
	}
	for _, key := range resources {
		value, ok := q[key]
		if !ok || string(value) == "null" {
			return false
		}
	}
	return true
}
func residentShape(raw json.RawMessage) bool {
	var values []map[string]json.RawMessage
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return false
	}
	for _, value := range values {
		if len(value) != 5 || value["parent"] == nil || !quantityShape(value["claims"]) || !quantityShape(value["capacities"]) {
			return false
		}
	}
	return true
}
func ReadRequest(data []byte) (Request, error) {
	var r Request
	if len(data) > 65536 {
		return r, ErrConflict
	}
	err := strict(data, &r)
	if err == nil && r.Operation == "admit" {
		var shape map[string]json.RawMessage
		if json.Unmarshal(data, &shape) != nil || !quantityShape(shape["claims"]) || !quantityShape(shape["capacities"]) {
			err = ErrConflict
		}
		if value, present := shape["residents"]; present && !residentShape(value) {
			err = ErrConflict
		}
	}
	return r, err
}
func readState(directory *os.File, mutable bool) (*state, error) {
	data, err := harnesslocal.ReadFileAt(directory, "state.json", 4*1024*1024, mutable)
	if errors.Is(err, os.ErrNotExist) {
		return &state{Generation: 2, Residents: []Resident{}, Leases: []*lease{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var s state
	if strict(data, &s) != nil || s.Generation != 2 || s.Residents == nil || len(s.Residents) > 4096 || !integer(s.Sequence) || s.Leases == nil || len(s.Leases) > 4096 {
		return nil, ErrConflict
	}
	// Structural required-field checks distinguish omitted false/zero from facts.
	var shape struct {
		Residents json.RawMessage              `json:"residents"`
		Sequence  *float64                     `json:"sequence"`
		Leases    []map[string]json.RawMessage `json:"leases"`
	}
	if json.Unmarshal(data, &shape) != nil || shape.Sequence == nil || !residentShape(shape.Residents) {
		return nil, ErrConflict
	}
	tokens := map[string]*lease{}
	for i, l := range s.Leases {
		if l == nil || len(shape.Leases[i]) != 11 || l.Residents == nil || !proofShape(l.Process) || l.Guards == nil || len(l.Guards) > 64 || !tokenPattern.MatchString(l.Token) || tokens[l.Token] != nil || !validMode(l.Mode, l.Claims, l.Parent) || !integer(l.Sequence) || l.Sequence < 1 || l.Sequence > s.Sequence || !quantities(l.Claims) || !quantities(l.Capacities) || (l.Parent != nil && (tokens[*l.Parent] == nil || tokens[*l.Parent].Mode == "observation")) || (l.Mode == "observation" && len(l.Guards) != 0) {
			return nil, ErrConflict
		}
		for _, p := range l.Guards {
			if !proofShape(p) {
				return nil, ErrConflict
			}
		}
		for _, key := range []string{"granted", "released"} {
			text := string(shape.Leases[i][key])
			if text != "true" && text != "false" {
				return nil, ErrConflict
			}
		}
		if !quantityShape(shape.Leases[i]["claims"]) || !quantityShape(shape.Leases[i]["capacities"]) || !residentShape(shape.Leases[i]["residents"]) {
			return nil, ErrConflict
		}
		tokens[l.Token] = l
	}
	for _, allocation := range s.Residents {
		if !validResident(allocation) || allocation.Parent != nil && tokens[*allocation.Parent] == nil {
			return nil, ErrConflict
		}
	}
	seen := map[string]bool{}
	for _, allocation := range s.Residents {
		if seen[allocation.Token] {
			return nil, ErrConflict
		}
		seen[allocation.Token] = true
	}
	for _, l := range s.Leases {
		for _, allocation := range l.Residents {
			if !validResident(allocation) {
				return nil, ErrConflict
			}
		}
	}
	return &s, nil
}

// Quiet uses the same strict state reader and lifecycle proofs as admission.
// An ungranted exclusive request also pauses optional observations so the
// observer cannot starve an earlier measurement request.
func Quiet(root string) (bool, error) {
	directory, err := harnesslocal.OpenDirectory(root, true)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	defer directory.Close()
	s, err := readState(directory, true)
	if err != nil {
		return true, err
	}
	for _, l := range s.Leases {
		if l.Mode == "exclusive" && !l.Released {
			alive, err := Alive(l.Process)
			if err != nil {
				return true, err
			}
			if alive {
				return true, nil
			}
		}
	}
	return false, nil
}
func mutate(s *state, r Request) (bool, error) {
	alive := map[Proof]bool{}
	live := func(p Proof) bool { return alive[p] }
	for _, l := range s.Leases {
		for _, p := range append(append([]Proof{}, l.Guards...), l.Process) {
			if _, known := alive[p]; known {
				continue
			}
			value, err := Alive(p)
			if err != nil {
				return false, err
			}
			alive[p] = value
		}
	}
	guarded := func(l *lease) bool {
		for _, p := range l.Guards {
			if live(p) {
				return true
			}
		}
		return false
	}
	// Durable allocations keep ancestor identities, never a dead controller's
	// activity entitlement. Only proven live descendants retain active claims.
	for _, l := range s.Leases {
		if !live(l.Process) && !guarded(l) {
			l.Released = true
		}
	}
	retain := map[string]bool{}
	for _, l := range s.Leases {
		if !l.Released && live(l.Process) || guarded(l) {
			retain[l.Token] = true
		}
	}
	for _, allocation := range s.Residents {
		if allocation.Parent != nil {
			retain[*allocation.Parent] = true
		}
	}
	for i := len(s.Leases) - 1; i >= 0; i-- {
		l := s.Leases[i]
		if retain[l.Token] && l.Parent != nil {
			retain[*l.Parent] = true
		}
	}
	filtered := []*lease{}
	for _, l := range s.Leases {
		if retain[l.Token] {
			filtered = append(filtered, l)
		}
	}
	s.Leases = filtered
	find := func(token string) *lease {
		for _, l := range s.Leases {
			if l.Token == token {
				return l
			}
		}
		return nil
	}
	capacityFor := func(candidate *lease, active []*lease) (Quantities, []Resident) {
		included := map[string]bool{candidate.Token: true}
		charged := map[string]bool{candidate.Token: true}
		for _, entry := range active {
			included[entry.Token] = true
			charged[entry.Token] = true
		}
		for _, entry := range s.Leases {
			if included[entry.Token] {
				for parent := entry.Parent; parent != nil; parent = find(*parent).Parent {
					included[*parent] = true
				}
			}
		}
		allocations := slices.Clone(s.Residents)
		for _, candidate := range candidate.Residents {
			found := false
			for _, existing := range allocations {
				found = found || existing.Token == candidate.Token
			}
			if !found {
				allocations = append(allocations, candidate)
			}
		}
		// Include retained ancestors for weighted nesting, without granting them
		// activity or charging a released ancestor's former active claim.
		for _, allocation := range allocations {
			for parent := allocation.Parent; parent != nil; parent = find(*parent).Parent {
				included[*parent] = true
			}
		}
		var total func(*lease, string) float64
		total = func(entry *lease, key string) float64 {
			sum := float64(0)
			for _, child := range s.Leases {
				if child.Parent != nil && *child.Parent == entry.Token && included[child.Token] {
					sum += total(child, key)
				}
			}
			for _, allocation := range allocations {
				if allocation.Parent != nil && *allocation.Parent == entry.Token {
					sum += allocation.Claims[key]
				}
			}
			claim := entry.Claims[key]
			if !charged[entry.Token] {
				claim = 0
			}
			return math.Max(claim, sum)
		}
		deficits := Quantities{}
		for _, key := range resources {
			if candidate.Mode == "observation" {
				break
			}
			capacity := candidate.Capacities[key]
			for _, entry := range active {
				if entry.Mode != "observation" {
					capacity = math.Min(capacity, entry.Capacities[key])
				}
			}
			sum := float64(0)
			residentTotal := float64(0)
			for _, allocation := range allocations {
				capacity = math.Min(capacity, allocation.Capacities[key])
				if key == "volume" {
					residentTotal = math.Max(residentTotal, allocation.Claims[key])
				} else {
					residentTotal += allocation.Claims[key]
				}
				if allocation.Parent == nil {
					sum += allocation.Claims[key]
				}
			}
			for _, entry := range s.Leases {
				if entry.Parent == nil && included[entry.Token] {
					sum += total(entry, key)
				}
			}
			// Volume is a writable-root capability, not a consumable disk slot.
			// Every resident retains the requirement; sharing never invents
			// additional storage capacity or promises device isolation.
			if key == "volume" {
				sum = 0
				for _, allocation := range allocations {
					sum = math.Max(sum, allocation.Claims[key])
				}
				for _, entry := range s.Leases {
					if charged[entry.Token] {
						sum = math.Max(sum, entry.Claims[key])
					}
				}
			}
			if sum > capacity {
				deficits[key] = sum - capacity
			}
			// Keep bounded cleanup possible after all active work drains. Idle
			// physical resources cannot consume the final cleanup process/memory.
			if residentTotal > 0 && residentTotal+cleanupHeadroom(key) > capacity {
				deficits[key] = math.Max(deficits[key], residentTotal+cleanupHeadroom(key)-capacity)
			}
		}
		return deficits, allocations
	}
	// Independent active work can drain; a waiting descendant must still
	// account for the ancestors whose entitlement it needs to execute.
	requiredAncestors := func(candidate *lease) []*lease {
		ancestors := []*lease{}
		for parent := candidate.Parent; parent != nil; {
			ancestor := find(*parent)
			if ancestor.Granted && (!ancestor.Released || guarded(ancestor)) {
				ancestors = append(ancestors, ancestor)
			}
			parent = ancestor.Parent
		}
		return ancestors
	}
	switch r.Operation {
	case "retirement_required":
		pending := find(r.Token)
		if pending == nil || pending.Granted || pending.Released || !quantities(r.Claims) {
			return false, ErrConflict
		}
		// Ancestors retain their activity entitlement while this request waits;
		// independent active work can finish without retiring infrastructure.
		deficits, _ := capacityFor(pending, requiredAncestors(pending))
		for _, key := range resources {
			if r.Claims[key] > 0 && deficits[key] > 0 {
				return true, nil
			}
		}
		return false, nil
	case "resident_granted":
		for _, allocation := range s.Residents {
			if allocation.Token == r.Token {
				return allocation.Reference == r.Reference, nil
			}
		}
		return false, nil
	case "release_resident":
		for i, a := range s.Residents {
			if a.Token == r.Token {
				if a.Reference != r.Reference {
					return false, ErrConflict
				}
				s.Residents = append(s.Residents[:i], s.Residents[i+1:]...)
				return false, nil
			}
		}
		return false, nil
	case "release":
		if l := find(r.Token); l != nil {
			l.Released = true
		}
		for i := len(s.Leases) - 1; i >= 0; i-- {
			l := s.Leases[i]
			hasChild := false
			for _, child := range s.Leases {
				if child.Parent != nil && *child.Parent == l.Token {
					hasChild = true
					break
				}
			}
			for _, allocation := range s.Residents {
				if allocation.Parent != nil && *allocation.Parent == l.Token {
					hasChild = true
				}
			}
			if l.Released && !guarded(l) && !hasChild {
				s.Leases = append(s.Leases[:i], s.Leases[i+1:]...)
			}
		}
		return false, nil
	case "bind":
		l := find(r.Token)
		if l == nil || !l.Granted || l.Released || l.Mode == "observation" || !proofShape(r.Guard) {
			return false, ErrConflict
		}
		actual, liveGuard, err := processIdentity(r.Guard.PID)
		if err != nil {
			return false, err
		}
		if actual != r.Guard {
			return false, ErrConflict
		}
		// A pinned terminal child needs no additional liveness guard. Its
		// granted lease still holds capacity until the owner settles the group
		// and releases it. Absence or a reused identity is never this case.
		if !liveGuard {
			return true, nil
		}
		guards := []Proof{}
		found := false
		for _, p := range l.Guards {
			if live(p) {
				guards = append(guards, p)
				found = found || p == r.Guard
			}
		}
		if !found {
			guards = append(guards, r.Guard)
		}
		if len(guards) > 64 {
			return false, ErrConflict
		}
		l.Guards = guards
		return true, nil
	case "admit":
		if r.Residents == nil {
			r.Residents = []Resident{}
		}
		if len(r.Residents) > 64 || r.Mode == "observation" && len(r.Residents) != 0 {
			return false, ErrConflict
		}
		seen := map[string]bool{}
		for _, allocation := range r.Residents {
			if !validResident(allocation) || seen[allocation.Token] || allocation.Parent != nil && !reflect.DeepEqual(allocation.Parent, r.Parent) {
				return false, ErrConflict
			}
			seen[allocation.Token] = true
			for _, existing := range s.Residents {
				if existing.Token == allocation.Token && !reflect.DeepEqual(existing, allocation) {
					return false, ErrConflict
				}
			}
		}
		if r.Mode == "drain" {
			if len(r.Residents) == 0 {
				return false, ErrConflict
			}
			for _, allocation := range r.Residents {
				found := false
				for _, existing := range s.Residents {
					found = found || reflect.DeepEqual(existing, allocation)
				}
				if !found {
					return false, ErrConflict
				}
			}
		}
		if !quantities(r.Claims) || !quantities(r.Capacities) || !validMode(r.Mode, r.Claims, r.Parent) {
			return false, ErrConflict
		}
		for _, key := range resources {
			if r.Claims[key] > r.Capacities[key] {
				return false, ErrConflict
			}
		}
		l := find(r.Token)
		if l == nil {
			if len(s.Leases) >= 4096 || s.Sequence >= maxInteger {
				return false, ErrConflict
			}
			value, err := Alive(r.Owner)
			if err != nil {
				return false, err
			}
			if !value {
				return false, ErrConflict
			}
			if r.Parent != nil {
				ancestor := find(*r.Parent)
				if ancestor == nil || !ancestor.Granted || ancestor.Released && r.Mode != "drain" || ancestor.Mode == "observation" {
					return false, ErrConflict
				}
			}
			s.Sequence++
			l = &lease{Residents: slices.Clone(r.Residents), Token: r.Token, Process: r.Owner, Guards: []Proof{}, Sequence: s.Sequence, Mode: r.Mode, Claims: r.Claims, Capacities: r.Capacities, Parent: r.Parent}
			s.Leases = append(s.Leases, l)
		} else if l.Granted && !reflect.DeepEqual(l.Residents, r.Residents) || l.Released || l.Process != r.Owner || l.Mode != r.Mode || !reflect.DeepEqual(l.Claims, r.Claims) || !reflect.DeepEqual(l.Capacities, r.Capacities) || !reflect.DeepEqual(l.Parent, r.Parent) {
			return false, ErrConflict
		}
		// Pending borrowers yield logical ownership while queued. Retirement
		// may replace a generation before the next attempt, but must not move
		// the borrower to the back of the queue. No capacity belongs to these
		// references until the request is granted atomically below.
		if !l.Granted {
			l.Residents = slices.Clone(r.Residents)
		}
		// An invocation cannot wait for capacity consumed by its own
		// indispensable startup and resident claims. Reject that shape before
		// launching physical owners, while ordinary contention remains queued.
		for _, key := range resources {
			resident, independent := float64(0), float64(0)
			for _, allocation := range l.Residents {
				if key == "volume" {
					resident = math.Max(resident, allocation.Claims[key])
				} else {
					resident += allocation.Claims[key]
				}
				if allocation.Parent == nil {
					independent += allocation.Claims[key]
				}
			}
			own := l.Claims[key] + resident
			if key == "volume" {
				own = math.Max(l.Claims[key], resident)
			}
			if own > l.Capacities[key] || resident > 0 && resident+cleanupHeadroom(key) > l.Capacities[key] {
				return false, ErrConflict
			}
			ancestorClaim := float64(0)
			for parent := l.Parent; parent != nil; parent = find(*parent).Parent {
				ancestorClaim = math.Max(ancestorClaim, find(*parent).Claims[key])
			}
			if key != "volume" && ancestorClaim+independent > l.Capacities[key] {
				return false, ErrConflict
			}
		}
		rootOf := func(l *lease) *lease {
			for l.Parent != nil {
				l = find(*l.Parent)
			}
			return l
		}
		root := rootOf(l)
		active, predecessors, roots := []*lease{}, []*lease{}, []*lease{}
		for _, entry := range s.Leases {
			if entry.Granted && (!entry.Released || guarded(entry)) {
				active = append(active, entry)
				if entry.Parent == nil {
					roots = append(roots, entry)
				}
			} else if !entry.Released && entry.Sequence < l.Sequence {
				predecessors = append(predecessors, entry)
			}
		}
		ancestors := map[string]bool{}
		for entry := l; entry.Parent != nil; entry = find(*entry.Parent) {
			ancestors[*entry.Parent] = true
		}
		// Preserve priority while independent active work drains. Only durable
		// resident pressure permits backfill: its consumers may need to run
		// before those allocations can release capacity.
		fitsFor := func(candidate *lease) (bool, []Resident) {
			deficits, allocations := capacityFor(candidate, active)
			return len(deficits) == 0, allocations
		}
		blocks := func(entry *lease) bool {
			if entry.Mode == "exclusive" {
				return true
			}
			if entry.Mode == "observation" {
				return false
			}
			deficits, _ := capacityFor(entry, requiredAncestors(entry))
			return len(deficits) == 0
		}
		compatible := true
		if l.Parent != nil {
			for _, entry := range predecessors {
				// An admitted parent retains its entitlement to drain. A
				// waiter in another branch must not block the descendants
				// needed to release that branch's existing reservation.
				// Siblings preserve transient wait priority; the weighted tree
				// still checks every request against host capacity.
				if r.Mode != "drain" && entry.Parent != nil && *entry.Parent == *l.Parent && blocks(entry) {
					compatible = false
				}
			}
			for _, entry := range active {
				if r.Mode == "exclusive" && !ancestors[entry.Token] || (r.Mode == "shared" || r.Mode == "drain") && entry.Mode == "exclusive" && !ancestors[entry.Token] {
					compatible = false
				}
			}
		} else if r.Mode == "observation" {
			// A bounded observer reserves no workload capacity and must be able
			// to inspect resource holders while ordinary work is queued. Any
			// pending exclusive request fences even an earlier observer retry.
			for _, entry := range s.Leases {
				if entry.Mode == "exclusive" && !entry.Released {
					compatible = false
				}
			}
		} else if r.Mode == "exclusive" {
			compatible = len(roots) == 0 && len(predecessors) == 0
		} else {
			for _, entry := range active {
				if entry.Mode == "exclusive" {
					compatible = false
				}
			}
			for _, entry := range predecessors {
				if r.Mode != "drain" && (entry.Parent == nil || rootOf(entry) != root) && blocks(entry) {
					compatible = false
				}
			}
		}
		fits, allocations := fitsFor(l)
		if fits && compatible {
			l.Granted = true
			s.Residents = allocations
		}
		return l.Granted, nil
	default:
		return false, ErrConflict
	}
}

func cleanupHeadroom(resource string) float64 {
	switch resource {
	case "memory_mb":
		return 64
	case "process":
		return 1
	default:
		return 0
	}
}

// Transact owns both the kernel mutex and the complete state transition. It
// never waits for a parent event loop while holding the lock.
func Transact(ctx context.Context, r Request) (bool, error) {
	if !filepath.IsAbs(r.Root) || filepath.Clean(r.Root) != r.Root || !tokenPattern.MatchString(r.Token) {
		return false, ErrConflict
	}
	directory, err := harnesslocal.PrivateDirectory(r.Root)
	if err != nil {
		return false, err
	}
	defer directory.Close()
	fd, err := syscall.Openat(int(directory.Fd()), "mutex", syscall.O_RDONLY|syscall.O_CREAT|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return false, err
	}
	lock := os.NewFile(uintptr(fd), "admission-mutex")
	defer lock.Close()
	var stat syscall.Stat_t
	if err := syscall.Fstat(fd, &stat); err != nil {
		return false, err
	}
	if stat.Mode&syscall.S_IFMT != syscall.S_IFREG || stat.Mode&0777 != 0600 || stat.Uid != uint32(os.Getuid()) || stat.Nlink != 1 || stat.Size != 0 {
		return false, ErrConflict
	}
	for {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			return false, err
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	current, err := os.Lstat(fmt.Sprintf("/proc/self/fd/%d/mutex", directory.Fd()))
	if err != nil {
		return false, err
	}
	opened, err := lock.Stat()
	if err != nil || !os.SameFile(current, opened) {
		return false, ErrConflict
	}
	if err := harnesslocal.CheckDirectory(directory, r.Root); err != nil {
		return false, err
	}
	s, err := readState(directory, false)
	if err != nil {
		return false, err
	}
	value, err := mutate(s, r)
	if err != nil {
		return false, err
	}
	data, err := json.Marshal(s)
	if err != nil || len(data)+1 > 4*1024*1024 {
		return false, ErrConflict
	}
	data = append(data, '\n')
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if err := harnesslocal.ReplaceFileAt(directory, "state.json", data); err != nil {
		return false, err
	}
	if err := harnesslocal.CheckDirectory(directory, r.Root); err != nil {
		return false, err
	}
	return value, nil
}
func Main(args []string, out io.Writer) int {
	if len(args) != 0 {
		return 2
	}
	request, err := ReadRequest([]byte(os.Getenv("CARTULARY_HOST_ADMISSION_REQUEST")))
	if err != nil {
		fmt.Fprintln(out, `{"ok":false}`)
		return 1
	}
	timeout := 5000
	if raw := os.Getenv("CARTULARY_HOST_ADMISSION_TIMEOUT_MS"); raw != "" {
		n, e := strconv.Atoi(raw)
		if e != nil || n < 1 || n > 2147483647 {
			return 2
		}
		timeout = n
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()
	value, err := Transact(ctx, request)
	if err != nil {
		fmt.Fprintln(out, `{"ok":false}`)
		return 1
	}
	fmt.Fprintf(out, "{\"ok\":true,\"value\":%t}\n", value)
	return 0
}
