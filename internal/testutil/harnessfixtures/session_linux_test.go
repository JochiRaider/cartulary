//go:build linux

package harnessfixtures

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func sessionFixture(t *testing.T, provider Provider) (*Session, net.Conn, <-chan error) {
	t.Helper()
	runtime, _, _ := runtimeFixture(t)
	broker := New(Options{Providers: map[string]Provider{"browser_stack": provider}})
	session := &Session{Runtime: runtime, Broker: broker, Close: broker.Close}
	a, b := net.Pipe()
	done := make(chan error, 1)
	go func() { done <- session.Serve(context.Background(), a) }()
	t.Cleanup(func() {
		b.Close()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("session failed to drain")
		}
	})
	return session, b, done
}
func fixtureRequest(s *Session, id int, op string) sessionRequest {
	v := sessionRequest{Schema: sessionSchema, Kind: "request", ID: id, RunID: s.Runtime.RunID, RuntimeLease: s.Runtime.LeaseID, Operation: op}
	if op == "close" {
		scope := "session"
		v.CloseScope = &scope
	}
	return v
}
func exchange(t *testing.T, c net.Conn, request sessionRequest) sessionResponse {
	t.Helper()
	data, _ := json.Marshal(request)
	if _, err := c.Write(append(data, '\n')); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(c).ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	var response sessionResponse
	if err := json.Unmarshal(line, &response); err != nil {
		t.Fatal(err)
	}
	if response.ID != request.ID || response.RunID != request.RunID || response.RuntimeLease != request.RuntimeLease {
		t.Fatal("forged response")
	}
	return response
}
func TestFixtureSession(t *testing.T) {
	t.Run("inherited_channel_requires_exact_parent_socket", func(t *testing.T) {
		proof, err := harnessadmission.Identity(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		pair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer unix.Close(pair[0])
		defer unix.Close(pair[1])
		if err := verifySessionPeer(pair[1], proof); err != nil {
			t.Fatal(err)
		}
		proof.Start = "0"
		if verifySessionPeer(pair[1], proof) == nil {
			t.Fatal("forged process admitted")
		}
		file, err := os.CreateTemp(t.TempDir(), "channel")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if verifySessionPeer(int(file.Fd()), proof) == nil {
			t.Fatal("regular file admitted")
		}
	})
	t.Run("closed_operations_and_repeated_release", func(t *testing.T) {
		var acquired, closed atomic.Int32
		s, c, _ := sessionFixture(t, &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
			acquired.Add(1)
			return &Allocation{Ownership: "owned", ResourceIDs: []string{"browser:a"}, Environment: map[string]string{"PORT": "1"}, Release: func(context.Context, bool) error { closed.Add(1); return nil }}, nil
		}})
		request := fixtureRequest(s, 1, "acquire")
		request.Request = &Request{Capability: "browser_stack", UnitID: "unit"}
		response := exchange(t, c, request)
		if !response.OK || response.Result.Environment["PORT"] != "1" {
			t.Fatal(response)
		}
		id := response.Result.Lease.ID
		request = fixtureRequest(s, 2, "release")
		request.LeaseID = &id
		request.Healthy = true
		response = exchange(t, c, request)
		if !response.OK {
			t.Fatal(response)
		}
		request.ID = 3
		if !exchange(t, c, request).OK {
			t.Fatal("idempotent release failed")
		}
		request = fixtureRequest(s, 4, "close")
		if !exchange(t, c, request).OK {
			t.Fatal("close failed")
		}
		if acquired.Load() != 1 || closed.Load() != 1 {
			t.Fatal(acquired.Load(), closed.Load())
		}
	})
	for _, bad := range []string{"duplicate_key", "forged_binding", "unknown_operation", "incomplete", "oversize", "late_recovery"} {
		t.Run(bad, func(t *testing.T) {
			var calls atomic.Int32
			s, c, _ := sessionFixture(t, &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
				calls.Add(1)
				return nil, errors.New("unexpected")
			}})
			request := fixtureRequest(s, 1, "acquire")
			request.Request = &Request{Capability: "browser_stack", UnitID: "unit"}
			data, _ := json.Marshal(request)
			switch bad {
			case "duplicate_key":
				data = []byte(strings.Replace(string(data), `"kind":"request"`, `"kind":"request","kind":"request"`, 1))
			case "forged_binding":
				data = []byte(strings.Replace(string(data), s.Runtime.LeaseID, strings.Repeat("a", 36), 1))
			case "unknown_operation":
				data = []byte(strings.Replace(string(data), `"operation":"acquire"`, `"operation":"run"`, 1))
			case "oversize":
				data = []byte(strings.Repeat(" ", maximumSessionFrame+1))
			case "late_recovery":
				request = fixtureRequest(s, 2, "recover")
				data, _ = json.Marshal(request)
			}
			if bad != "incomplete" {
				data = append(data, '\n')
			}
			c.SetWriteDeadline(time.Now().Add(time.Second))
			_, _ = c.Write(data)
			c.Close()
			time.Sleep(10 * time.Millisecond)
			if calls.Load() != 0 {
				t.Fatal("invalid request caused a side effect")
			}
		})
	}
	t.Run("owner_loss_cancels_inflight_acquisition", func(t *testing.T) {
		entered := make(chan struct{})
		cancelled := make(chan struct{})
		s, c, _ := sessionFixture(t, &ProviderFuncs{AcquireFunc: func(ctx context.Context, _ Request) (*Allocation, error) {
			close(entered)
			<-ctx.Done()
			close(cancelled)
			return nil, ctx.Err()
		}})
		request := fixtureRequest(s, 1, "acquire")
		request.Request = &Request{Capability: "browser_stack", UnitID: "unit"}
		data, _ := json.Marshal(request)
		c.Write(append(data, '\n'))
		<-entered
		c.Close()
		select {
		case <-cancelled:
		case <-time.After(time.Second):
			t.Fatal("owner loss did not cancel acquisition")
		}
	})
	t.Run("close_cancels_before_waiting_for_acquisition", func(t *testing.T) {
		entered := make(chan struct{})
		s, c, _ := sessionFixture(t, &ProviderFuncs{AcquireFunc: func(ctx context.Context, _ Request) (*Allocation, error) {
			close(entered)
			<-ctx.Done()
			return nil, ctx.Err()
		}})
		request := fixtureRequest(s, 1, "acquire")
		request.Request = &Request{Capability: "browser_stack", UnitID: "unit"}
		data, _ := json.Marshal(request)
		c.Write(append(data, '\n'))
		<-entered
		request = fixtureRequest(s, 2, "close")
		data, _ = json.Marshal(request)
		c.Write(append(data, '\n'))
		c.SetReadDeadline(time.Now().Add(3 * time.Second))
		reader := bufio.NewReader(c)
		for range 2 {
			if _, err := reader.ReadBytes('\n'); err != nil {
				t.Fatal(err)
			}
		}
	})
}
