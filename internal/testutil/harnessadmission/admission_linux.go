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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

var ErrConflict = errors.New("host admission unavailable")
var resources = []string{"browser_stack", "cpu", "io", "memory_mb", "process"}

const maxInteger = 9007199254740991

var digits = regexp.MustCompile(`^[0-9]+$`)
var tokenPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

type Proof struct {
	Boot  string `json:"boot"`
	PID   int    `json:"pid"`
	Start string `json:"start"`
}
type Quantities map[string]float64
type Request struct {
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
	Sequence float64  `json:"sequence"`
	Leases   []*lease `json:"leases"`
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
func proofShape(p Proof) bool { return p.Boot != "" && p.PID > 1 && digits.MatchString(p.Start) }
func Identity(pid int) (Proof, error) {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return Proof{}, err
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return Proof{}, err
	}
	end := strings.LastIndex(string(data), ") ")
	if end < 0 {
		return Proof{}, ErrConflict
	}
	f := strings.Fields(string(data)[end+2:])
	if len(f) < 20 || !digits.MatchString(f[19]) {
		return Proof{}, ErrConflict
	}
	return Proof{strings.TrimSpace(string(boot)), pid, f[19]}, nil
}
func Alive(p Proof) (bool, error) {
	if !proofShape(p) {
		return false, nil
	}
	actual, err := Identity(p.PID)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if actual != p {
		return false, nil
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", p.PID))
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	end := strings.LastIndex(string(data), ") ")
	if end < 0 {
		return false, ErrConflict
	}
	fields := strings.Fields(string(data)[end+2:])
	return len(fields) > 19 && fields[0] != "Z" && fields[0] != "X" && fields[19] == p.Start, nil
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
func ReadRequest(data []byte) (Request, error) {
	var r Request
	if len(data) > 8192 {
		return r, ErrConflict
	}
	err := strict(data, &r)
	if err == nil && r.Operation == "admit" {
		var shape map[string]json.RawMessage
		if json.Unmarshal(data, &shape) != nil || !quantityShape(shape["claims"]) || !quantityShape(shape["capacities"]) {
			err = ErrConflict
		}
	}
	return r, err
}
func readState(directory *os.File, mutable bool) (*state, error) {
	data, err := harnesslocal.ReadFileAt(directory, "state.json", 4*1024*1024, mutable)
	if errors.Is(err, os.ErrNotExist) {
		return &state{Leases: []*lease{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var s state
	if strict(data, &s) != nil || !integer(s.Sequence) || s.Leases == nil || len(s.Leases) > 4096 {
		return nil, ErrConflict
	}
	// Structural required-field checks distinguish omitted false/zero from facts.
	var shape struct {
		Sequence *float64                     `json:"sequence"`
		Leases   []map[string]json.RawMessage `json:"leases"`
	}
	if json.Unmarshal(data, &shape) != nil || shape.Sequence == nil {
		return nil, ErrConflict
	}
	tokens := map[string]bool{}
	for i, l := range s.Leases {
		if l == nil || len(shape.Leases[i]) != 10 || !proofShape(l.Process) || l.Guards == nil || len(l.Guards) > 64 || !tokenPattern.MatchString(l.Token) || tokens[l.Token] || (l.Mode != "shared" && l.Mode != "exclusive") || !integer(l.Sequence) || l.Sequence < 1 || l.Sequence > s.Sequence || !quantities(l.Claims) || !quantities(l.Capacities) || (l.Parent != nil && !tokens[*l.Parent]) {
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
		if !quantityShape(shape.Leases[i]["claims"]) || !quantityShape(shape.Leases[i]["capacities"]) {
			return nil, ErrConflict
		}
		tokens[l.Token] = true
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
	retain := map[string]bool{}
	for _, l := range s.Leases {
		if !l.Released && live(l.Process) || guarded(l) {
			retain[l.Token] = true
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
	switch r.Operation {
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
			if l.Released && !guarded(l) && !hasChild {
				s.Leases = append(s.Leases[:i], s.Leases[i+1:]...)
			}
		}
		return false, nil
	case "bind":
		l := find(r.Token)
		value, err := Alive(r.Guard)
		if err != nil {
			return false, err
		}
		if l == nil || !l.Granted || l.Released || !value {
			return false, ErrConflict
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
		if !quantities(r.Claims) || !quantities(r.Capacities) || (r.Mode != "shared" && r.Mode != "exclusive") {
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
				if ancestor == nil || !ancestor.Granted || ancestor.Released {
					return false, ErrConflict
				}
			}
			s.Sequence++
			l = &lease{r.Token, r.Owner, []Proof{}, s.Sequence, r.Mode, r.Claims, r.Capacities, r.Parent, false, false}
			s.Leases = append(s.Leases, l)
		} else if l.Released || l.Process != r.Owner || l.Mode != r.Mode || !reflect.DeepEqual(l.Claims, r.Claims) || !reflect.DeepEqual(l.Capacities, r.Capacities) || !reflect.DeepEqual(l.Parent, r.Parent) {
			return false, ErrConflict
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
			if entry.Granted {
				active = append(active, entry)
				if entry.Parent == nil {
					roots = append(roots, entry)
				}
			} else if entry.Sequence < l.Sequence {
				predecessors = append(predecessors, entry)
			}
		}
		ancestors := map[string]bool{}
		for entry := l; entry.Parent != nil; entry = find(*entry.Parent) {
			ancestors[*entry.Parent] = true
		}
		compatible := true
		if l.Parent != nil {
			for _, entry := range predecessors {
				if entry.Parent != nil && rootOf(entry) == root {
					compatible = false
				}
			}
			for _, entry := range active {
				if r.Mode == "exclusive" && !ancestors[entry.Token] || r.Mode == "shared" && entry.Mode == "exclusive" && !ancestors[entry.Token] {
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
				if entry.Parent == nil || rootOf(entry) != root {
					compatible = false
				}
			}
		}
		included := map[string]bool{l.Token: true}
		for _, entry := range active {
			included[entry.Token] = true
		}
		var total func(*lease, string) float64
		total = func(entry *lease, key string) float64 {
			sum := float64(0)
			for _, child := range s.Leases {
				if child.Parent != nil && *child.Parent == entry.Token && included[child.Token] {
					sum += total(child, key)
				}
			}
			return math.Max(entry.Claims[key], sum)
		}
		fits := true
		for _, key := range resources {
			capacity := r.Capacities[key]
			for _, entry := range active {
				capacity = math.Min(capacity, entry.Capacities[key])
			}
			sum := float64(0)
			for _, entry := range s.Leases {
				if entry.Parent == nil && included[entry.Token] {
					sum += total(entry, key)
				}
			}
			if sum > capacity {
				fits = false
			}
		}
		if fits && compatible {
			l.Granted = true
		}
		return l.Granted, nil
	default:
		return false, ErrConflict
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
