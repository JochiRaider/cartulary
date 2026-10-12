//go:build linux

package harnessreview

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

const adapterSchema = "cartulary.harness_review_adapter.v1"

type adapterFacts struct {
	Epoch       *int                        `json:"epoch"`
	Capacities  harnessadmission.Quantities `json:"capacities"`
	Source      string                      `json:"source_digest"`
	Commit      string                      `json:"source_commit"`
	Toolchain   string                      `json:"toolchain_digest"`
	Review      *designReady                `json:"review"`
	Environment map[string]string           `json:"environment"`
}
type adapterOutcome struct {
	Epoch    *int               `json:"epoch"`
	Counts   Counts             `json:"counts"`
	Bundle   *string            `json:"bundle_id"`
	Private  []PrivateReference `json:"private_refs"`
	Failures []Failure          `json:"failures"`
	Exit     int                `json:"exit_code"`
	Stop     bool               `json:"stop"`
}
type adapterReply struct {
	Schema    string         `json:"schema_id"`
	Session   string         `json:"session_id"`
	Run       string         `json:"run_id"`
	Lease     string         `json:"runtime_lease_id"`
	ID        int            `json:"request_id"`
	Kind      string         `json:"kind"`
	Operation string         `json:"operation"`
	Facts     *adapterFacts  `json:"facts"`
	Outcome   adapterOutcome `json:"outcome"`
	Failure   Failure        `json:"failure"`
}
type pendingRequest struct {
	operation string
	response  chan adapterReply
}
type adapter struct {
	mu, writeMu sync.Mutex
	dispatch    sync.Mutex
	runtime     *harnesssession.Runtime
	session     string
	cmd         *exec.Cmd
	channel     *os.File
	done        chan struct{}
	readDone    chan struct{}
	idle        chan struct{}
	pending     map[int]pendingRequest
	sequence    int
	failure     error
	onLost      func(error)
	closing     bool
	exitError   error
}

func startAdapter(root, session string, runtime *harnesssession.Runtime, environment map[string]string, onLost func(error)) (*adapter, error) {
	sockets, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	parent, child := os.NewFile(uintptr(sockets[0]), "review-parent"), os.NewFile(uintptr(sockets[1]), "review-child")
	cmd := exec.Command(environment["NODE_BIN"], filepath.Join(root, "tools/harness/browser/ui-review/adapter.mjs"), "--private-adapter")
	cmd.Dir = root
	env := maps.Clone(environment)
	maps.Copy(env, runtime.Environment())
	env["CARTULARY_REVIEW_SESSION_ID"] = session
	for _, k := range slices.Sorted(maps.Keys(env)) {
		cmd.Env = append(cmd.Env, k+"="+env[k])
	}
	cmd.ExtraFiles = []*os.File{child}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	cmd.WaitDelay = 2 * time.Second
	release, err := harnessadmission.InheritWorkspace(cmd)
	if err != nil {
		parent.Close()
		child.Close()
		return nil, err
	}
	defer release()
	if err = cmd.Start(); err != nil {
		parent.Close()
		child.Close()
		return nil, err
	}
	child.Close()
	a := &adapter{runtime: runtime, session: session, cmd: cmd, channel: parent, done: make(chan struct{}), readDone: make(chan struct{}), pending: map[int]pendingRequest{}, onLost: onLost}
	proof, proofErr := harnessadmission.Identity(cmd.Process.Pid)
	target := map[string]any{"boot": proof.Boot, "pid": proof.PID, "start": proof.Start, "group": true}
	if proofErr == nil {
		proofErr = runtime.RecordResource("helper_process", target, "acquired")
	}
	if proofErr != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	go func() {
		defer close(a.done)
		var waitErr error
		for {
			var info unix.Siginfo
			waitErr = unix.Waitid(unix.P_PID, cmd.Process.Pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
			if waitErr != unix.EINTR {
				break
			}
		}
		// The zombie leader reserves the process-group identity until the final signal.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		waitErr = errors.Join(waitErr, cmd.Wait())
		// Exit and SIGKILL are wakeups, not proof that the whole group is gone.
		// The Go dependency owner acknowledges this journal entry only after the
		// physical process adapter confirms exact descendant drain.
		select {
		case <-a.readDone:
		case <-time.After(2 * time.Second):
		}
		parent.Close()
		a.mu.Lock()
		a.exitError = waitErr
		closing := a.closing
		a.mu.Unlock()
		a.reject(waitErr)
		if !closing && a.onLost != nil {
			a.onLost(failure("session_lost"))
		}
	}()
	a.idle = make(chan struct{})
	close(a.idle)
	go a.read()
	if proofErr != nil {
		<-a.done
		return nil, proofErr
	}
	return a, nil
}
func (a *adapter) reject(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.failure == nil {
		a.failure = normalized(err, "session_lost")
	}
	if len(a.pending) > 0 {
		for id, entry := range a.pending {
			close(entry.response)
			delete(a.pending, id)
		}
		close(a.idle)
	}
}
func (a *adapter) read() {
	defer close(a.readDone)
	reader := bufio.NewReaderSize(a.channel, harnesssession.MaximumFrame)
	for {
		line, err := reader.ReadSlice('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				a.reject(err)
			}
			return
		}
		value, err := harnessschema.Decode(line)
		if err != nil || harnessschema.Validate(adapterSchema, value) != nil {
			a.reject(harnessschema.ErrInvalid)
			a.kill()
			return
		}
		var reply adapterReply
		if json.Unmarshal(line, &reply) != nil || reply.Session != a.session || reply.Run != a.runtime.RunID || reply.Lease != a.runtime.LeaseID {
			a.reject(harnesslocal.ErrUnsafe)
			a.kill()
			return
		}
		if reply.Kind == "event" {
			if reply.ID != 0 || reply.Operation != "lost" || reply.Failure.validate() != nil {
				a.reject(harnessschema.ErrInvalid)
				a.kill()
				return
			}
			if a.onLost != nil {
				a.onLost(&ReviewError{Record: reply.Failure, Exit: failurePolicy[reply.Failure.Code].Exit})
			}
			continue
		}
		if reply.Kind != "response" {
			a.reject(harnessschema.ErrInvalid)
			a.kill()
			return
		}
		for _, f := range reply.Outcome.Failures {
			if f.validate() != nil {
				a.reject(harnessschema.ErrInvalid)
				a.kill()
				return
			}
		}
		if (len(reply.Outcome.Failures) == 0) != (reply.Outcome.Exit == 0) {
			a.reject(harnessschema.ErrInvalid)
			a.kill()
			return
		}
		if reply.ID == 0 && reply.Operation == "abort" {
			continue
		}
		a.mu.Lock()
		p, ok := a.pending[reply.ID]
		if ok && p.operation == reply.Operation {
			delete(a.pending, reply.ID)
			if len(a.pending) == 0 {
				close(a.idle)
			}
			p.response <- reply
		}
		a.mu.Unlock()
		if !ok || p.operation != reply.Operation {
			a.reject(harnessschema.ErrInvalid)
			a.kill()
			return
		}
	}
}
func (a *adapter) write(id int, operation string, payload any) error {
	data, err := jsonDocument(adapterSchema, map[string]any{"schema_id": adapterSchema, "session_id": a.session, "run_id": a.runtime.RunID, "runtime_lease_id": a.runtime.LeaseID, "request_id": id, "kind": "request", "operation": operation, "payload": payload})
	if err != nil || len(data) > harnesssession.MaximumFrame {
		return harnessschema.ErrInvalid
	}
	a.writeMu.Lock()
	defer a.writeMu.Unlock()
	_ = a.channel.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = a.channel.Write(data)
	return err
}
func (a *adapter) call(ctx context.Context, operation string, payload any) (adapterReply, error) {
	a.dispatch.Lock()
	a.mu.Lock()
	if a.failure != nil {
		err := a.failure
		a.mu.Unlock()
		a.dispatch.Unlock()
		return adapterReply{}, err
	}
	if len(a.pending) == 0 {
		a.idle = make(chan struct{})
	}
	a.sequence++
	id := a.sequence
	p := pendingRequest{operation, make(chan adapterReply, 1)}
	a.pending[id] = p
	a.mu.Unlock()
	err := a.write(id, operation, payload)
	a.dispatch.Unlock()
	if err != nil {
		a.reject(err)
		a.kill()
		return adapterReply{}, err
	}
	var reply adapterReply
	var ok bool
	select {
	case reply, ok = <-p.response:
	case <-ctx.Done():
		a.interrupt()
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		select {
		case reply, ok = <-p.response:
		case <-timer.C:
			a.kill()
			<-a.done
			return adapterReply{}, failure("cleanup_failed")
		}
		if ok && len(reply.Outcome.Failures) == 0 {
			return reply, ctx.Err()
		}
	}
	if !ok {
		a.mu.Lock()
		err := a.failure
		a.mu.Unlock()
		return reply, normalized(err, "session_lost")
	}
	if len(reply.Outcome.Failures) > 0 {
		return reply, &ReviewError{Record: reply.Outcome.Failures[0], Exit: reply.Outcome.Exit, Secondary: reply.Outcome.Failures[1:]}
	}
	return reply, nil
}
func (a *adapter) interrupt() { go func() { _ = a.write(0, "abort", nil) }() }
func (a *adapter) kill() {
	a.mu.Lock()
	defer a.mu.Unlock()
	select {
	case <-a.done:
		return
	default:
	}
	_ = syscall.Kill(-a.cmd.Process.Pid, syscall.SIGKILL)
}
func (a *adapter) close(ctx context.Context) error {
	a.mu.Lock()
	a.closing = true
	a.mu.Unlock()
	_, err := a.call(ctx, "close", nil)
	if err != nil {
		a.kill()
	}
	select {
	case <-a.done:
		return errors.Join(err, a.exitError)
	case <-ctx.Done():
		a.kill()
		<-a.done
		return errors.Join(err, ctx.Err())
	}
}

func (a *adapter) drain(ctx context.Context) error {
	a.mu.Lock()
	idle := a.idle
	a.mu.Unlock()
	select {
	case <-idle:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
