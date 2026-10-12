//go:build linux

package harnessfixtures

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

const sessionSchema = "cartulary.harness_fixture_session.v1"
const maximumSessionFrame = 1048576

type sessionRequest struct {
	CloseScope   *string  `json:"close_scope"`
	Schema       string   `json:"schema_id"`
	Kind         string   `json:"kind"`
	ID           int      `json:"request_id"`
	RunID        string   `json:"run_id"`
	RuntimeLease string   `json:"runtime_lease_id"`
	Operation    string   `json:"operation"`
	Request      *Request `json:"request"`
	LeaseID      *string  `json:"lease_id"`
	Healthy      bool     `json:"healthy"`
	Retain       bool     `json:"retain"`
}
type sessionCleanup struct {
	Operation string        `json:"operation"`
	UnitID    string        `json:"unit_id"`
	LeaseID   string        `json:"lease_id"`
	Failures  []wireFailure `json:"failures"`
	Published bool          `json:"published"`
}
type sessionResult struct {
	Cleanup              []sessionCleanup  `json:"cleanup"`
	AllocationRef        string            `json:"allocation_ref"`
	ObservationAvailable bool              `json:"observation_available"`
	Lease                *LeaseRecord      `json:"lease"`
	Environment          map[string]string `json:"environment"`
	Retained             bool              `json:"retained"`
	Unresolved           bool              `json:"unresolved"`
	Secrets              []string          `json:"secrets"`
	Observations         []Observation     `json:"observations"`
	ContainerProof       map[string]any    `json:"container_proof"`
}
type wireFailure struct {
	Class  string   `json:"failure_class"`
	Reason string   `json:"failure_reason"`
	Refs   []string `json:"artifact_refs"`
}
type sessionResponse struct {
	Schema       string         `json:"schema_id"`
	Kind         string         `json:"kind"`
	ID           int            `json:"request_id"`
	RunID        string         `json:"run_id"`
	RuntimeLease string         `json:"runtime_lease_id"`
	Operation    string         `json:"operation"`
	OK           bool           `json:"ok"`
	Result       *sessionResult `json:"result"`
	Failures     []wireFailure  `json:"failures"`
}
type Session struct {
	Runtime           *harnesssession.Runtime
	Broker            *Broker
	Close             func(context.Context) error
	Recover           func(context.Context) error
	mu                sync.Mutex
	observations      []Observation
	cleanup           []sessionCleanup
	container         map[string]any
	observationFailed bool
}

func (s *Session) Observe(v Observation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.observations) >= 4096 {
		s.observationFailed = true
		return
	}
	s.observations = append(s.observations, v)
}
func (s *Session) Cleanup(v CleanupRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanup = append(s.cleanup, sessionCleanup{v.Operation, v.UnitID, v.LeaseID, sessionFailures(v.Err), v.Published})
}
func (s *Session) Container(v map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.container = maps.Clone(v)
}
func (s *Session) facts() (*sessionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	secrets, err := s.Runtime.ForbiddenValues()
	v := &sessionResult{Cleanup: append([]sessionCleanup{}, s.cleanup...), Environment: map[string]string{}, Secrets: append([]string{}, secrets...), Observations: append([]Observation{}, s.observations...), ContainerProof: maps.Clone(s.container), Unresolved: s.Broker.Unresolved(), ObservationAvailable: !s.observationFailed}
	s.observations = nil
	s.cleanup = nil
	return v, err
}
func sessionFailures(err error) []wireFailure {
	result := []wireFailure{}
	if err == nil {
		return result
	}
	var visit func(error)
	visit = func(err error) {
		if err == nil || len(result) >= 64 {
			return
		}
		if many, ok := err.(interface{ Unwrap() []error }); ok {
			for _, child := range many.Unwrap() {
				visit(child)
			}
			return
		}
		var f *Failure
		if errors.As(err, &f) {
			result = append(result, wireFailure{f.Class, f.Reason, append([]string{}, f.Refs...)})
			return
		}
		if errors.Is(err, context.Canceled) {
			result = append(result, wireFailure{"interrupted", "cancelled_or_interrupted", []string{}})
		} else if errors.Is(err, context.DeadlineExceeded) {
			result = append(result, wireFailure{"timing", "timeout_failure", []string{}})
		} else {
			result = append(result, wireFailure{"harness", "fixture_error", []string{}})
		}
	}
	visit(err)
	return result
}

// Serve accepts a single inherited private duplex descriptor. EOF cancels in
// flight acquisition and drains owned dependencies even if no close arrived.
// It has no executable/argv operation and never reads a public command plan.
func (s *Session) Serve(ctx context.Context, transport io.ReadWriteCloser) (result error) {
	life, cancel := context.WithCancel(ctx)
	defer cancel()
	defer transport.Close()
	go func() { <-life.Done(); transport.Close() }()
	var writeMu, stateMu sync.Mutex
	var workers sync.WaitGroup
	leases := map[string]*Lease{}
	inflight := make(chan struct{}, 64)
	var closeOnce sync.Once
	var closeErr error
	close := func() {
		closeOnce.Do(func() {
			cleanup, done := context.WithTimeout(context.Background(), 2*time.Minute)
			defer done()
			closeErr = s.Close(cleanup)
		})
	}
	defer func() { cancel(); workers.Wait(); close(); result = errors.Join(result, closeErr) }()
	write := func(response sessionResponse) error {
		bytes, err := validateJSON(sessionSchema, response)
		if err != nil {
			return err
		}
		if len(bytes)+1 > maximumSessionFrame {
			return harnesslocal.ErrUnsafe
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		_, err = transport.Write(append(bytes, '\n'))
		return err
	}
	scanner := bufio.NewScanner(transport)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, data[:i], nil
		}
		if atEOF && len(data) > 0 {
			return 0, nil, harnesslocal.ErrUnsafe
		}
		return 0, nil, nil
	})
	scanner.Buffer(make([]byte, 4096), maximumSessionFrame)
	last := 0
	acquired := false
	for scanner.Scan() {
		bytes := scanner.Bytes()
		value, err := harnessschema.Decode(bytes)
		if err != nil || harnessschema.Validate(sessionSchema, value) != nil {
			return artifactFailure(harnesslocal.ErrUnsafe)
		}
		var request sessionRequest
		if json.Unmarshal(bytes, &request) != nil || request.Kind != "request" || request.ID <= last || request.RunID != s.Runtime.RunID || request.RuntimeLease != s.Runtime.LeaseID {
			return artifactFailure(harnesslocal.ErrUnsafe)
		}
		last = request.ID
		if (request.Operation == "close") != (request.CloseScope != nil) {
			return artifactFailure(harnesslocal.ErrUnsafe)
		}
		if (request.Operation == "acquire") != (request.Request != nil) || (request.Operation == "release") != (request.LeaseID != nil) || request.Operation != "release" && (request.Healthy || request.Retain) {
			return artifactFailure(harnesslocal.ErrUnsafe)
		}
		if request.Operation == "recover" {
			if acquired || last != 1 {
				return artifactFailure(harnesslocal.ErrUnsafe)
			}
		}
		if request.Operation == "close" || request.Operation == "recover" {
			var err error
			if request.Operation == "recover" {
				if s.Recover == nil {
					err = artifactFailure(harnesslocal.ErrUnsafe)
				} else {
					err = s.Recover(life)
				}
			} else if *request.CloseScope == "allocations" {
				err = s.Broker.Close(life)
			} else {
				close()
				err = closeErr
			}
			workers.Wait()
			facts, factErr := s.facts()
			err = errors.Join(err, factErr)
			facts.Unresolved = facts.Unresolved || err != nil
			if e := write(sessionResponse{sessionSchema, "response", request.ID, request.RunID, request.RuntimeLease, request.Operation, err == nil, facts, sessionFailures(err)}); e != nil {
				return e
			}
			if request.Operation == "close" && *request.CloseScope == "session" {
				return err
			}
			continue
		}
		acquired = true
		select {
		case inflight <- struct{}{}:
		default:
			return artifactFailure(errors.New("fixture request bound exhausted"))
		}
		workers.Go(func() {
			defer func() { <-inflight }()
			var lease *Lease
			var retained bool
			var err error
			if request.Operation == "acquire" {
				lease, err = s.Broker.Acquire(life, *request.Request)
				if lease != nil {
					stateMu.Lock()
					leases[lease.Record().ID] = lease
					stateMu.Unlock()
				}
			} else {
				stateMu.Lock()
				lease = leases[*request.LeaseID]
				stateMu.Unlock()
				if lease == nil {
					err = artifactFailure(harnesslocal.ErrUnsafe)
				} else {
					retained, err = lease.Release(life, request.Healthy, request.Retain)
				}
			}
			facts, factErr := s.facts()
			err = errors.Join(err, factErr)
			facts.Retained = retained
			if lease != nil {
				record := lease.Record()
				facts.Lease = &record
				facts.AllocationRef = lease.allocation.request.AllocationRef
				if request.Operation == "acquire" && err == nil {
					facts.Environment = lease.Environment()
				}
			}
			if e := write(sessionResponse{sessionSchema, "response", request.ID, request.RunID, request.RuntimeLease, request.Operation, err == nil, facts, sessionFailures(err)}); e != nil {
				cancel()
			}
		})
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, os.ErrClosed) {
		return artifactFailure(err)
	}
	return nil
}

// Main is the private composition boundary; no public selection can replace the
// repository, parent identity, runtime proof or inherited channel.
func Main(args []string) int {
	if len(args) != 0 {
		return 2
	}
	root, err := os.Getwd()
	if err != nil {
		return 2
	}
	value, err := harnessschema.Decode([]byte(os.Getenv("CARTULARY_HARNESS_FIXTURE_OWNER")))
	if err != nil {
		return 11
	}
	bytes, _ := json.Marshal(value)
	var parent harnessadmission.Proof
	if object, ok := value.(map[string]any); !ok || len(object) != 3 {
		return 11
	}
	if json.Unmarshal(bytes, &parent) != nil {
		return 11
	}
	actual, err := harnessadmission.Identity(os.Getppid())
	if err != nil || actual != parent {
		return 11
	}
	environment := map[string]string{}
	for _, entry := range os.Environ() {
		if key, value, ok := strings.Cut(entry, "="); ok {
			environment[key] = value
		}
	}
	runID := environment["CARTULARY_TEST_RUN_ID"]
	runRoot := filepath.Join(environment["CARTULARY_TEST_RESULTS_DIR"], runID)
	runtime, err := harnesssession.BorrowRuntime(root, runRoot, environment["CARTULARY_HARNESS_SUITE_RUNTIME_ROOT"], runID, environment["CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID"])
	if err != nil {
		return 11
	}
	defer runtime.Close()
	if err := verifySessionPeer(3, actual); err != nil {
		return 11
	}
	channel := os.NewFile(3, "fixture-session")
	if channel == nil {
		return 11
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	session := &Session{Runtime: runtime}
	production := NewProduction(ProductionOptions{Root: root, RunRoot: runRoot, Target: environment["CARTULARY_TEST_TARGET"], Runtime: runtime, Environment: environment, ContainerProof: session.Container}, Options{Observe: session.Observe, Cleanup: session.Cleanup})
	session.Broker = production.Broker
	session.Close = production.Close
	session.Recover = production.Options.Recover
	if err := session.Serve(ctx, channel); err != nil {
		return 12
	}
	return 0
}

// The inherited duplex socket must have been created by this exact live parent.
// A regular file, anonymous pipe, or socket from another owner is never a session.
func verifySessionPeer(fd int, parent harnessadmission.Proof) error {
	kind, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_TYPE)
	if err != nil || kind != unix.SOCK_STREAM {
		return harnesslocal.ErrUnsafe
	}
	peer, err := unix.GetsockoptUcred(fd, unix.SOL_SOCKET, unix.SO_PEERCRED)
	if err != nil || int(peer.Pid) != parent.PID || int(peer.Uid) != os.Getuid() {
		return harnesslocal.ErrUnsafe
	}
	actual, err := harnessadmission.Identity(int(peer.Pid))
	if err != nil || actual != parent {
		return harnesslocal.ErrUnsafe
	}
	return nil
}
