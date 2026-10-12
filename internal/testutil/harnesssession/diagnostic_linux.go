//go:build linux

package harnesssession

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
)

type diagnosticScope struct {
	Boot  string `json:"boot"`
	Start string `json:"start"`
	Token string `json:"token"`
	UID   int    `json:"uid"`
}

type DiagnosticInspection struct {
	Processes []ProcessProof `json:"processes"`
	RSSBytes  int64          `json:"rss_bytes"`
}

// CreateDiagnosticScope persists the native admission boundary before returning
// the nonce to a producer. Linux /proc starttime uses USER_HZ (100 ticks/sec).
func (r *Runtime) CreateDiagnosticScope(dependencies []string) (diagnosticScope, error) {
	scope, err := newProcessScope()
	if err != nil {
		return scope, err
	}
	return scope, r.RecordDependentResource("diagnostic_scope", scope, "pending", dependencies)
}

func newProcessScope() (diagnosticScope, error) {
	var scope diagnosticScope
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return scope, err
	}
	var now unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &now); err != nil {
		return scope, err
	}
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return scope, err
	}
	scope = diagnosticScope{Boot: strings.TrimSpace(string(boot)), Start: strconv.FormatInt(now.Nano()/10000000, 10), UID: os.Getuid(), Token: hex.EncodeToString(nonce[:])}
	return scope, nil
}

// authorizedScope runs under the runtime lock. No unrecorded nonce, process ID,
// command, or output location can grant diagnostic cleanup authority.
func (r *Runtime) authorizedScope(raw json.RawMessage) (diagnosticScope, error) {
	var scope diagnosticScope
	resource := Resource{Kind: "diagnostic_scope", Target: raw, State: "acquired", RunID: r.RunID, LeaseID: r.LeaseID}
	if r.checkResource(resource) != nil || json.Unmarshal(raw, &scope) != nil {
		return scope, harnesslocal.ErrUnsafe
	}
	resources, err := r.resources(4096)
	if err != nil {
		return scope, err
	}
	for _, resource := range resources {
		var recorded diagnosticScope
		if resource.Kind == "diagnostic_scope" && json.Unmarshal(resource.Target, &recorded) == nil && recorded == scope {
			boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
			if err != nil {
				return scope, err
			}
			if strings.TrimSpace(string(boot)) != scope.Boot {
				return scope, ErrUnresolved
			}
			return scope, nil
		}
	}
	return scope, harnesslocal.ErrUnsafe
}

// diagnosticMember bounds discovery and rechecks identity around private proof
// reads. Inaccessible proof for a same-UID process born within the scope is
// uncertainty, never evidence that the process is unrelated.
func diagnosticMember(scope diagnosticScope, pid int) (processState, bool, int64, error) {
	return scopedMember(scope, pid, "CARTULARY_DIAGNOSTIC_TOKEN")
}

// Procfs can deny environ access briefly while exec/exit replaces the mm.
// Reobserve the exact identity inside a bounded proof-read window. This never
// accepts inaccessible proof, changes the diagnostic sampling interval, or
// carries discovery authority across an identity change.
func scopedMember(scope diagnosticScope, pid int, marker string) (processState, bool, int64, error) {
	anchor, err := readProcess(pid)
	if err != nil {
		return processState{}, false, 0, err
	}
	deadline := time.Now().Add(10 * time.Millisecond)
	for {
		member, marked, rss, err := readScopedMember(scope, pid, marker)
		if !errors.Is(err, os.ErrPermission) || !time.Now().Before(deadline) {
			return member, marked, rss, err
		}
		time.Sleep(time.Millisecond)
		current, check := readProcess(pid)
		if processGone(check) {
			return processState{}, false, 0, nil
		}
		if check != nil || !sameProcess(anchor.ProcessProof, current.ProcessProof) {
			return processState{}, false, 0, errors.Join(check, ErrUnresolved)
		}
	}
}

func readScopedMember(scope diagnosticScope, pid int, marker string) (found processState, marked bool, rss int64, resultErr error) {
	var empty processState
	member, err := readProcess(pid)
	if err != nil {
		return empty, false, 0, err
	}
	defer func() {
		if resultErr != nil && diagnosticReadSettled(member) {
			found, marked, rss, resultErr = empty, false, 0, nil
		}
	}()
	start, err := strconv.ParseUint(member.Start, 10, 64)
	boundary, boundaryErr := strconv.ParseUint(scope.Start, 10, 64)
	if err != nil || boundaryErr != nil {
		return empty, false, 0, ErrUnresolved
	}
	if start < boundary || member.state == "Z" || member.state == "X" {
		return empty, false, 0, nil
	}
	credentials, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return empty, false, 0, err
	}
	var ids []string
	for _, line := range strings.Split(string(credentials), "\n") {
		if strings.HasPrefix(line, "Uid:") {
			ids = strings.Fields(line)[1:]
			break
		}
	}
	if len(ids) != 4 {
		return empty, false, 0, ErrUnresolved
	}
	// procfs inode ownership can change when dumpability changes. Actual task
	// credentials, not the directory owner, decide whether proof is relevant.
	matching := 0
	for _, id := range ids {
		uid, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return empty, false, 0, ErrUnresolved
		}
		if uid == uint64(scope.UID) {
			matching++
		}
	}
	if matching == 0 {
		return empty, false, 0, nil
	}
	if matching != len(ids) {
		return empty, false, 0, ErrUnresolved
	}
	file, err := os.Open(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		return empty, false, 0, err
	}
	environment, readErr := io.ReadAll(io.LimitReader(file, 1048577))
	err = errors.Join(readErr, file.Close())
	if err != nil || len(environment) > 1048576 {
		return empty, false, 0, errors.Join(err, ErrUnresolved)
	}
	marked = false
	for _, value := range strings.Split(string(environment), "\x00") {
		if value == marker+"="+scope.Token {
			marked = true
		}
	}
	if !marked {
		return empty, false, 0, nil
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", pid))
	if err != nil {
		return empty, false, 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return empty, false, 0, ErrUnresolved
	}
	pages, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || pages < 0 || pages > (1<<52)/int64(os.Getpagesize()) {
		return empty, false, 0, ErrUnresolved
	}
	current, err := readProcess(pid)
	if err != nil {
		return empty, false, 0, err
	}
	if !sameProcess(member.ProcessProof, current.ProcessProof) {
		return empty, false, 0, ErrUnresolved
	}
	member.Boot = scope.Boot
	return member, true, pages * int64(os.Getpagesize()), nil
}

// A process can exit between stat and environ; Linux may then return EACCES
// while dismantling its mm. Only renewed absence or the same dead identity
// resolves that read failure. A live inaccessible or replaced PID stays unsafe.
func diagnosticReadSettled(previous processState) bool {
	current, err := readProcess(previous.PID)
	return processGone(err) || err == nil && sameProcess(previous.ProcessProof, current.ProcessProof) && (current.state == "Z" || current.state == "X")
}

func inspectDiagnostic(scope diagnosticScope) (DiagnosticInspection, []processState, error) {
	return inspectScope(scope, "CARTULARY_DIAGNOSTIC_TOKEN", nil)
}

func inspectScope(scope diagnosticScope, marker string, excluded map[int]ProcessProof) (DiagnosticInspection, []processState, error) {
	result := DiagnosticInspection{Processes: []ProcessProof{}}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return result, nil, err
	}
	if len(entries) > 65536 {
		return result, nil, ErrUnresolved
	}
	var members []processState
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid < 2 {
			continue
		}
		if proof, ok := excluded[pid]; ok {
			current, err := CaptureProcess(pid)
			if processGone(err) {
				continue
			}
			if err != nil || current.Boot != proof.Boot || !sameProcess(current, proof) {
				return result, nil, ErrUnresolved
			}
			continue
		}
		member, marked, rss, err := scopedMember(scope, pid, marker)
		if processGone(err) {
			continue
		}
		if err != nil {
			return result, nil, errors.Join(ErrUnresolved, err)
		}
		if !marked {
			continue
		}
		if len(members) >= 4096 || result.RSSBytes+rss > 1<<52 {
			return result, nil, ErrUnresolved
		}
		members = append(members, member)
		result.Processes = append(result.Processes, member.ProcessProof)
		result.RSSBytes += rss
	}
	return result, members, nil
}

func (r *Runtime) InspectDiagnosticScope(raw json.RawMessage) (DiagnosticInspection, error) {
	unlock, err := r.lock()
	if err != nil {
		return DiagnosticInspection{}, err
	}
	defer unlock()
	scope, err := r.authorizedScope(raw)
	if err != nil {
		return DiagnosticInspection{}, err
	}
	result, _, err := inspectDiagnostic(scope)
	return result, err
}

func (r *Runtime) SettleDiagnosticScope(ctx context.Context, raw json.RawMessage) error {
	unlock, err := r.lock()
	if err != nil {
		return err
	}
	defer unlock()
	scope, err := r.authorizedScope(raw)
	if err != nil {
		return err
	}
	return settleScope(ctx, scope, "CARTULARY_DIAGNOSTIC_TOKEN", nil)
}

func settleScope(ctx context.Context, scope diagnosticScope, marker string, excluded map[int]ProcessProof) error {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil || strings.TrimSpace(string(boot)) != scope.Boot {
		return errors.Join(err, ErrUnresolved)
	}
	escalate := time.Now().Add(200 * time.Millisecond)
	known := map[int]ProcessProof{}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, members, err := inspectScope(scope, marker, excluded)
		if err != nil {
			return err
		}
		observed := map[int]bool{}
		for _, member := range members {
			observed[member.PID] = true
		}
		for pid, proof := range known {
			current, err := readProcess(pid)
			if processGone(err) {
				continue
			}
			if err != nil || !sameProcess(current.ProcessProof, proof) {
				return errors.Join(err, ErrUnresolved)
			}
			if current.state != "Z" && current.state != "X" && !observed[pid] {
				return ErrUnresolved
			}
		}
		if len(members) == 0 {
			return nil
		}
		sig := unix.SIGTERM
		if !time.Now().Before(escalate) {
			sig = unix.SIGKILL
		}
		for _, member := range members {
			known[member.PID] = member.ProcessProof
			current, marked, _, err := scopedMember(scope, member.PID, marker)
			if processGone(err) {
				continue
			}
			if err != nil {
				return errors.Join(ErrUnresolved, err)
			}
			if !marked {
				state, err := readProcess(member.PID)
				if processGone(err) || err == nil && sameProcess(state.ProcessProof, member.ProcessProof) && (state.state == "Z" || state.state == "X") {
					continue
				}
				return ErrUnresolved
			}
			if !sameProcess(current.ProcessProof, member.ProcessProof) {
				return ErrUnresolved
			}
			if err := signalProcess(current, sig); err != nil && !processGone(err) {
				return err
			}
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
