//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

const nestedContextKey = "CARTULARY_HARNESS_NESTED_LIFECYCLE"
const nestedSettlementBudget = 5 * time.Minute

// An abstract socket is private invocation authority, not retained evidence or
// a second resource owner. Kernel peer credentials and exact launch ancestry
// bind registrations; the unpredictable endpoint is never sufficient proof.
type nestedContext struct {
	Version  int                         `json:"version"`
	Endpoint string                      `json:"endpoint"`
	Owner    harnesssession.ProcessProof `json:"owner"`
}
type nestedScope struct {
	listener    *net.UnixListener
	context     nestedContext
	mu          sync.Mutex
	launch      harnesssession.ProcessProof
	bound       chan struct{}
	peers       map[*net.UnixConn]bool
	cancelled   bool
	drained     chan struct{}
	drainedOnce sync.Once
	err         error
	wg          sync.WaitGroup
}

func newNestedScope() (*nestedScope, error) {
	owner, err := harnesssession.CaptureProcess(os.Getpid())
	if err != nil {
		return nil, err
	}
	endpoint := "@cartulary-nested-" + uuid()
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: endpoint, Net: "unix"})
	if err != nil {
		return nil, err
	}
	s := &nestedScope{listener: listener, context: nestedContext{1, endpoint, owner}, bound: make(chan struct{}), peers: map[*net.UnixConn]bool{}, drained: make(chan struct{})}
	s.wg.Add(1)
	go s.accept()
	return s, nil
}
func (s *nestedScope) environment() string                    { data, _ := json.Marshal(s.context); return string(data) }
func (s *nestedScope) bind(proof harnesssession.ProcessProof) { s.launch = proof; close(s.bound) }
func peerIdentity(c *net.UnixConn) (harnesssession.ProcessProof, error) {
	raw, err := c.SyscallConn()
	if err != nil {
		return harnesssession.ProcessProof{}, err
	}
	var cred *unix.Ucred
	var credentialErr error
	err = raw.Control(func(fd uintptr) {
		cred, credentialErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	if err != nil || credentialErr != nil || cred == nil || int(cred.Uid) != os.Getuid() {
		return harnesssession.ProcessProof{}, errors.New("invalid nested peer")
	}
	return harnesssession.CaptureProcess(int(cred.Pid))
}
func (s *nestedScope) accept() {
	defer s.wg.Done()
	for {
		c, err := s.listener.AcceptUnix()
		if err != nil {
			return
		}
		// No peer data is consumed before authentication. Registration cannot
		// cross cancellation, and rejected peers never acquire graph resources.
		<-s.bound
		peer, err := peerIdentity(c)
		s.mu.Lock()
		if err != nil || !harnesssession.DescendsFrom(peer.PID, s.launch) || len(s.peers) >= 64 {
			s.mu.Unlock()
			c.Close()
			continue
		}
		if s.cancelled {
			c.SetWriteDeadline(time.Now().Add(time.Second))
			_, _ = c.Write([]byte{2})
			s.mu.Unlock()
			c.Close()
			continue
		}
		s.peers[c] = true
		c.SetWriteDeadline(time.Now().Add(time.Second))
		_, err = c.Write([]byte{1})
		if err != nil {
			delete(s.peers, c)
			s.err = errors.New("nested registration failed")
			s.mu.Unlock()
			c.Close()
			continue
		}
		s.wg.Add(1)
		s.mu.Unlock()
		go func() {
			defer s.wg.Done()
			var done [1]byte
			_, err := io.ReadFull(c, done[:])
			s.mu.Lock()
			if err != nil || done[0] != 3 {
				s.err = errors.New("nested graph did not acknowledge settlement")
			}
			delete(s.peers, c)
			s.drainLocked()
			s.mu.Unlock()
			c.Close()
		}()
	}
}
func (s *nestedScope) drainLocked() {
	if s.cancelled && len(s.peers) == 0 {
		s.drainedOnce.Do(func() { close(s.drained) })
	}
}
func (s *nestedScope) cancel() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.cancelled {
		s.cancelled = true
		deadline := time.Now().Add(time.Second)
		for c := range s.peers {
			c.SetWriteDeadline(deadline)
			// A concurrent completed child may already have closed its write
			// side. Its acknowledgement reader determines settlement, not EPIPE.
			_, _ = c.Write([]byte{2})
		}
		s.drainLocked()
	}
	return s.drained
}
func (s *nestedScope) close() error {
	s.listener.Close()
	// Unblock accept even when spawning or capturing the launch failed.
	select {
	case <-s.bound:
	default:
		s.bind(harnesssession.ProcessProof{})
	}
	s.mu.Lock()
	for c := range s.peers {
		c.Close()
	}
	s.mu.Unlock()
	s.wg.Wait()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Registration happens before Prepare and any resource acquisition. The parent
// observes settlement only after Run's finalizers and terminal scan return.
// The inherited channel is consumed here; each launched unit gets a new scope.
func joinNested(ctx context.Context, value string) (context.Context, func(bool), error) {
	if value == "" {
		return ctx, func(bool) {}, nil
	}
	if len(value) > 4096 {
		return nil, nil, errors.New("invalid nested context")
	}
	if _, err := harnessschema.Decode([]byte(value)); err != nil {
		return nil, nil, errors.New("invalid nested context")
	}
	var binding nestedContext
	d := json.NewDecoder(strings.NewReader(value))
	d.DisallowUnknownFields()
	if err := d.Decode(&binding); err != nil || binding.Version != 1 || !strings.HasPrefix(binding.Endpoint, "@cartulary-nested-") {
		return nil, nil, errors.New("invalid nested context")
	}
	if d.Decode(new(any)) != io.EOF {
		return nil, nil, errors.New("invalid nested context")
	}
	dialer := net.Dialer{Timeout: time.Second}
	connection, err := dialer.DialContext(ctx, "unix", binding.Endpoint)
	if err != nil {
		return nil, nil, err
	}
	c := connection.(*net.UnixConn)
	peer, err := peerIdentity(c)
	if err != nil || peer != binding.Owner {
		c.Close()
		return nil, nil, errors.New("foreign nested owner")
	}
	c.SetReadDeadline(time.Now().Add(time.Second))
	var ack [1]byte
	if _, err = io.ReadFull(c, ack[:]); err != nil || (ack[0] != 1 && ack[0] != 2) {
		c.Close()
		return nil, nil, errors.New("nested registration failed")
	}
	if ack[0] == 2 {
		c.Close()
		return nil, nil, context.Canceled
	}
	c.SetReadDeadline(time.Time{})
	child, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); var signal [1]byte; _, _ = io.ReadFull(c, signal[:]); cancel() }()
	return child, func(clean bool) {
		c.SetWriteDeadline(time.Now().Add(time.Second))
		ack := byte(3)
		if !clean {
			ack = 4
		}
		_, _ = c.Write([]byte{ack})
		c.Close()
		<-done
		cancel()
	}, nil
}
