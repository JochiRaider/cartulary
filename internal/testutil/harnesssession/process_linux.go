//go:build linux

package harnesssession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
)

// ProcessProof distinguishes a process group from the containing session.
// The physical owner never infers absence from a missing/reused leader alone.
type ProcessProof struct {
	Boot  string `json:"boot"`
	PID   int    `json:"pid"`
	Start string `json:"start"`
	Group bool   `json:"group"`
	PGID  int    `json:"pgid"`
	SID   int    `json:"sid"`
}
type processState struct {
	ProcessProof
	parent int
	state  string
}

func readProcess(pid int) (processState, error) {
	var p processState
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return p, err
	}
	i := strings.LastIndex(string(data), ") ")
	if i < 0 {
		return p, harnesslocal.ErrUnsafe
	}
	f := strings.Fields(string(data)[i+2:])
	if len(f) < 20 {
		return p, harnesslocal.ErrUnsafe
	}
	p.PID, p.Start, p.state = pid, f[19], f[0]
	p.parent, err = strconv.Atoi(f[1])
	if err != nil {
		return p, err
	}
	p.PGID, err = strconv.Atoi(f[2])
	if err != nil {
		return p, err
	}
	p.SID, err = strconv.Atoi(f[3])
	if err != nil {
		return p, err
	}
	p.Group = p.PGID == pid
	return p, nil
}
func CaptureProcess(pid int) (ProcessProof, error) {
	p, err := readProcess(pid)
	if err != nil {
		return ProcessProof{}, err
	}
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	p.Boot = strings.TrimSpace(string(boot))
	return p.ProcessProof, err
}
func sameProcess(a, b ProcessProof) bool {
	return a.PID == b.PID && a.Start == b.Start && a.PGID == b.PGID && a.SID == b.SID
}
func processGone(err error) bool { return errors.Is(err, os.ErrNotExist) || errors.Is(err, unix.ESRCH) }

func processMembers(p ProcessProof) ([]processState, error) {
	if !p.Group {
		v, err := readProcess(p.PID)
		if processGone(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if !sameProcess(v.ProcessProof, p) {
			return nil, AtCleanupStage("process_identity", ErrUnresolved)
		}
		return []processState{v}, nil
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	if len(entries) > 65536 {
		return nil, ErrUnresolved
	}
	var members []processState
	for _, entry := range entries {
		pid, e := strconv.Atoi(entry.Name())
		if e != nil {
			continue
		}
		v, e := readProcess(pid)
		if processGone(e) {
			continue
		}
		if e != nil {
			return nil, e
		}
		if v.PGID == p.PGID {
			if v.SID != p.SID {
				return nil, AtCleanupStage("process_identity", ErrUnresolved)
			}
			members = append(members, v)
		}
	}
	return members, nil
}

// StopProcess uses pidfds for each verified member, never a reusable negative
// PID. An unreaped leader or an already observed surviving member anchors each
// subsequent discovery sweep. Unanchored groups are never signaled; a bounded
// attempt may observe their disappearance, otherwise settlement fails closed.
func StopProcess(ctx context.Context, p ProcessProof, grace time.Duration) error {
	if p.PID < 2 || p.PGID < 1 || p.SID < 1 || !regexpDigits.MatchString(p.Start) || p.Group != (p.PGID == p.PID) {
		return harnesslocal.ErrUnsafe
	}
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(boot)) != p.Boot {
		return ErrUnresolved
	}
	return settleProcess(ctx, p, grace, processMembers, signalProcess)
}

func signalProcess(m processState, sig unix.Signal) error {
	fd, err := unix.PidfdOpen(m.PID, 0)
	if err != nil {
		// A reaped group leader can still have a kernel PID object held by its
		// surviving group. pidfd_open then reports EINVAL, not ESRCH. Only a
		// fresh absence observation permits the enclosing discovery to continue;
		// EINVAL itself never proves absence or authorizes signaling another PID.
		if errors.Is(err, unix.EINVAL) {
			if _, observed := readProcess(m.PID); processGone(observed) {
				return observed
			}
		}
		return AtCleanupStage("process_handle", err)
	}
	defer unix.Close(fd)
	current, err := readProcess(m.PID)
	if err != nil {
		return AtCleanupStage("process_identity", err)
	}
	if !sameProcess(current.ProcessProof, m.ProcessProof) {
		return AtCleanupStage("process_identity", ErrUnresolved)
	}
	return unix.PidfdSendSignal(fd, sig, nil, 0)
}

func settleProcess(ctx context.Context, p ProcessProof, grace time.Duration, inspect func(ProcessProof) ([]processState, error), signal func(processState, unix.Signal) error) error {
	known := map[int]ProcessProof{p.PID: p}
	escalate := time.Now().Add(grace)
	for {
		members, err := inspect(p)
		if err != nil {
			return AtCleanupStage("process_discovery", err)
		}
		if len(members) == 0 {
			return nil
		}
		anchored, live := false, false
		for _, m := range members {
			old, ok := known[m.PID]
			anchored = anchored || ok && sameProcess(old, m.ProcessProof)
			live = live || (m.state != "Z" && m.state != "X")
		}
		var pending error
		if !anchored {
			pending = AtCleanupStage("process_group_anchor", ErrUnresolved)
			if _, bounded := ctx.Deadline(); !bounded {
				return pending
			}
		} else if !live {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return errors.Join(pending, err)
		}
		if anchored {
			sig := unix.SIGTERM
			if !time.Now().Before(escalate) {
				sig = unix.SIGKILL
			}
			for _, m := range members {
				known[m.PID] = m.ProcessProof
				if m.state == "Z" || m.state == "X" {
					continue
				}
				e := signal(m, sig)
				if e != nil && !processGone(e) {
					return AtCleanupStage("process_signal", e)
				}
			}
		}
		// Unanchored observations neither signal nor enter the known set.
		// Reobserve in this attempt; keep its original deadline and grace clock.
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(pending, ctx.Err())
		case <-timer.C:
		}
	}
}

// ProcessMain is a closed physical adapter over private stdin. It accepts no
// command or PID arguments. Authority is a live descendant of the invoking
// parent, or an exact proof in the bound runtime's resource/dependency journal.
func ProcessMain(args []string, input io.Reader, diagnostic io.Writer) int {
	var failure error
	defer func() {
		if failure != nil {
			_ = json.NewEncoder(diagnostic).Encode(DescribeCleanupFailure(failure))
		}
	}()
	if len(args) != 0 {
		return 2
	}
	data, err := io.ReadAll(io.LimitReader(input, 4097))
	if err != nil || len(data) > 4096 {
		return 11
	}
	v, err := harnessschema.Decode(data)
	m, ok := v.(map[string]any)
	if err != nil || !ok || len(m) != 3 {
		return 11
	}
	var q struct {
		Proof ProcessProof `json:"proof"`
		Grace int          `json:"grace_ms"`
		Kill  int          `json:"kill_ms"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&q) != nil || q.Grace < 0 || q.Grace > 2000 || q.Kill < 1 || q.Kill > 5000 {
		return 11
	}
	// Reuse the generated process-proof projection at the physical boundary.
	if harnessschema.Validate("cartulary.harness_runtime_resource.v2", map[string]any{"kind": "helper_process", "target": m["proof"], "state": "acquired", "run_id": "physical-process", "lease_id": "00000000-0000-0000-0000-000000000000"}) != nil {
		return 11
	}
	members, err := processMembers(q.Proof)
	if err != nil {
		failure = AtCleanupStage("process_discovery", err)
		return 12
	}
	if len(members) == 0 {
		if err := StopProcess(context.Background(), q.Proof, 0); err != nil {
			failure = AtCleanupStage("process_settlement", err)
			return 12
		}
		return 0
	}
	authorized := false
	for pid, depth := q.Proof.PID, 0; pid > 1 && depth < 128; depth++ {
		p, e := readProcess(pid)
		if e != nil {
			break
		}
		if p.parent == os.Getppid() {
			authorized = true
			break
		}
		if p.parent == pid {
			break
		}
		pid = p.parent
	}
	if !authorized {
		authorized = recordedProcess(q.Proof)
	}
	if !authorized {
		failure = AtCleanupStage("process_authority", ErrUnresolved)
		return 12
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(q.Grace+q.Kill)*time.Millisecond)
	defer cancel()
	if err := StopProcess(ctx, q.Proof, time.Duration(q.Grace)*time.Millisecond); err != nil {
		failure = AtCleanupStage("process_settlement", err)
		return 12
	}
	return 0
}

func recordedProcess(proof ProcessProof) bool {
	root, err := os.Getwd()
	if err != nil {
		return false
	}
	output := os.Getenv("CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT")
	if output == "" {
		output = filepath.Join(os.Getenv("CARTULARY_TEST_RESULTS_DIR"), os.Getenv("CARTULARY_TEST_RUN_ID"))
	}
	r, err := BorrowRuntime(root, output, os.Getenv("CARTULARY_HARNESS_SUITE_RUNTIME_ROOT"), os.Getenv("CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID"), os.Getenv("CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID"))
	if err != nil {
		return false
	}
	defer r.ReleaseHandles()
	resources, err := r.Resources(4096)
	if err != nil {
		return false
	}
	for _, resource := range resources {
		var p ProcessProof
		if strings.HasSuffix(resource.Kind, "_process") && json.Unmarshal(resource.Target, &p) == nil && p == proof {
			return true
		}
		for _, name := range resource.RecoveryFiles {
			data, e := harnesslocal.ReadFile(name, 1048576, false)
			if e != nil {
				continue
			}
			var record struct {
				Proof ProcessProof `json:"proof"`
			}
			if json.Unmarshal(data, &record) == nil && record.Proof == proof {
				return true
			}
		}
	}
	return false
}
