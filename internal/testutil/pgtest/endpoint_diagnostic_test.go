package pgtest

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"net/netip"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/jackc/pgx/v5"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func testEndpointDiagnostic(t *testing.T) {
	binding := func(port string) *dockercontainer.InspectResponse {
		return &dockercontainer.InspectResponse{ID: "fixture-container", NetworkSettings: &dockercontainer.NetworkSettings{Ports: network.PortMap{network.MustParsePort("5432/tcp"): {{HostIP: netip.MustParseAddr("0.0.0.0"), HostPort: port}}}}}
	}
	t.Run("driver_observation_preserves_selection_and_redacts_identity", func(t *testing.T) {
		d := newStartupDiagnostic(t.Context())
		e := d.endpoint
		c := fakePostgresContainer{inspect: func(context.Context) (*dockercontainer.InspectResponse, error) { return binding("45678"), nil }}
		e.selectEndpoint(t.Context(), c, "sensitive-host.example", "45678")
		cfg, err := pgx.ParseConfig("postgres://user:password@sensitive-host.example:45678/postgres?sslmode=require")
		if err != nil {
			t.Fatal(err)
		}
		originalTLS := &tls.Config{MinVersion: tls.VersionTLS13, ServerName: "sensitive-host.example"}
		cfg.TLSConfig = originalTLS
		addresses := []string{"192.0.2.11", "2001:db8::11"}
		lookupCalls, dialCalls := 0, 0
		cfg.LookupFunc = func(ctx context.Context, host string) ([]string, error) {
			lookupCalls++
			if host != "sensitive-host.example" {
				t.Fatal(host)
			}
			return addresses, nil
		}
		cfg.DialFunc = func(ctx context.Context, network, address string) (net.Conn, error) {
			dialCalls++
			if network != "tcp" || address != "[2001:db8::11]:45678" {
				t.Fatal(network, address)
			}
			return nil, &net.OpError{Op: "dial", Net: network, Err: syscall.ECONNREFUSED}
		}
		e.observeConfig(cfg)
		got, err := cfg.LookupFunc(t.Context(), e.host)
		if err != nil || !reflect.DeepEqual(got, addresses) {
			t.Fatal(got, err)
		}
		_, err = cfg.DialFunc(t.Context(), "tcp", "[2001:db8::11]:45678")
		if !errors.Is(err, syscall.ECONNREFUSED) || lookupCalls != 1 || dialCalls != 1 || cfg.TLSConfig != originalTLS {
			t.Fatal("driver behavior changed")
		}
		o := d.Endpoint
		if o.Lookups[0].Addresses[0].Family != "ipv4" || o.Lookups[0].Addresses[1].Family != "ipv6" || o.Dials[0].Endpoint.Ref != o.Lookups[0].Addresses[1].Ref || !o.Dials[0].FromLookup || !o.Dials[0].MatchesSelectedPort || o.Dials[0].Outcome != "connection_refused" {
			t.Fatal(o)
		}
		d.finish("sql_connectivity", context.DeadlineExceeded)
		d.CleanupOutcome = "completed"
		payload, err := json.Marshal(d.PostgresReadinessDiagnostic)
		if err != nil {
			t.Fatal(err)
		}
		value, err := harnessschema.Decode(payload)
		if err != nil {
			t.Fatal(err)
		}
		if err := harnessschema.Validate("cartulary.postgres_readiness_diagnostic.v4", value); err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{"sensitive-host", "45678", "192.0.2.11", "2001:db8", "password", "0.0.0.0"} {
			if strings.Contains(string(payload), secret) {
				t.Fatal("retained endpoint identity", secret)
			}
		}
		for range 130 {
			_, _ = cfg.LookupFunc(t.Context(), e.host)
		}
		if len(d.Endpoint.Lookups) != 128 || !d.Endpoint.Truncated {
			t.Fatal("lookup bound lost")
		}
	})
	t.Run("binding_comparison_requires_complete_exact_identity", func(t *testing.T) {
		for _, tc := range []struct {
			name, port, id     string
			unavailable        bool
			changed, published *bool
		}{
			{"unchanged", "45678", "fixture-container", false, boolPointer(false), boolPointer(true)},
			{"changed", "45679", "fixture-container", false, boolPointer(true), boolPointer(false)},
			{"foreign", "45678", "foreign-container", true, nil, nil},
			{"malformed", "bad-port", "fixture-container", true, nil, nil},
			{"missing_inspection", "45678", "", true, nil, nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				d := newStartupDiagnostic(t.Context())
				current := binding("45678")
				c := fakePostgresContainer{inspect: func(ctx context.Context) (*dockercontainer.InspectResponse, error) {
					deadline, ok := ctx.Deadline()
					if !ok || time.Until(deadline) > 250*time.Millisecond {
						t.Fatal("inspection unbounded")
					}
					if current.ID == "" {
						return nil, errors.New("provider secret")
					}
					return current, nil
				}}
				d.endpoint.selectEndpoint(t.Context(), c, "127.0.0.1", "45678")
				current = binding(tc.port)
				current.ID = tc.id
				d.endpoint.inspect(t.Context(), c, false)
				o := d.Endpoint
				if !reflect.DeepEqual(o.BindingChanged, tc.changed) || !reflect.DeepEqual(o.SelectedPortPublished, tc.published) || (o.FailureStatus == "unavailable") != tc.unavailable {
					t.Fatal(o)
				}
			})
		}
		d := newStartupDiagnostic(t.Context())
		many := binding("45678")
		port := network.MustParsePort("5432/tcp")
		for range 33 {
			many.NetworkSettings.Ports[port] = append(many.NetworkSettings.Ports[port], many.NetworkSettings.Ports[port][0])
		}
		c := fakePostgresContainer{inspect: func(context.Context) (*dockercontainer.InspectResponse, error) { return many, nil }}
		d.endpoint.selectEndpoint(t.Context(), c, "127.0.0.1", "45678")
		d.endpoint.inspect(t.Context(), c, false)
		if d.Endpoint.FailureStatus != "truncated" || d.Endpoint.BindingChanged != nil || d.Endpoint.SelectedPortPublished != nil || len(d.Endpoint.FailureBindings) != 32 {
			t.Fatal(d.Endpoint)
		}
	})
	t.Run("internal_listener_and_published_transport_are_independent", func(t *testing.T) {
		oldListener, oldDial := postgresInternalListenerFn, postgresPublishedDialFn
		t.Cleanup(func() { postgresInternalListenerFn, postgresPublishedDialFn = oldListener, oldDial })
		d := newStartupDiagnostic(t.Context())
		c := fakePostgresContainer{inspect: func(context.Context) (*dockercontainer.InspectResponse, error) { return binding("45678"), nil }}
		d.endpoint.selectEndpoint(t.Context(), c, "127.0.0.1", "45678")
		postgresInternalListenerFn = func(ctx context.Context, _ testcontainers.Container) string {
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > 500*time.Millisecond {
				t.Fatal("listener budget")
			}
			return "present"
		}
		postgresPublishedDialFn = func(ctx context.Context, address string) (net.Conn, error) {
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > 250*time.Millisecond || address != "127.0.0.1:45678" {
				t.Fatal("published probe changed selection or budget")
			}
			return nil, syscall.ECONNREFUSED
		}
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()
		d.endpoint.observeFailure(ctx, c)
		if d.Endpoint.InternalListener != "present" || d.Endpoint.PublishedDial.Outcome != "connection_refused" || d.Endpoint.BindingChanged == nil || *d.Endpoint.BindingChanged {
			t.Fatal(d.Endpoint)
		}
		// An exhausted shared budget cannot start another network observation.
		d.Endpoint.PublishedDial = nil
		postgresInternalListenerFn = func(ctx context.Context, _ testcontainers.Container) string {
			if ctx.Err() == nil {
				t.Fatal("budget reset")
			}
			return "unavailable"
		}
		postgresPublishedDialFn = func(context.Context, string) (net.Conn, error) {
			t.Fatal("dial survived exhausted budget")
			return nil, nil
		}
		cancel()
		d.endpoint.observeFailure(ctx, c)
		if d.Endpoint.PublishedDial != nil {
			t.Fatal("late observation")
		}
	})
	t.Run("listener_parser_and_stream_order", func(t *testing.T) {
		for _, tc := range []struct{ data, want string }{{"secret malformed", "unavailable"}, {" sl local_address rem_address st\n 0: 00000000:1538 00000000:0000 0A", "present"}, {" sl local_address rem_address st\n 0: 00000000:1538 00000000:0000 01", "absent"}} {
			if got := postgresListenerFromProc(tc.data); got != tc.want {
				t.Fatal(got, tc.want)
			}
		}
		d := newStartupDiagnostic(t.Context())
		started := "2026-10-20T00:00:00Z"
		d.ContainerStartedAt = &started
		d.observeStartupLogStreams([]postgresLogChunk{
			{stream: "stdout", data: []byte("2026-10-20T00:00:10Z PostgreSQL init process complete\n")},
			{stream: "stderr", data: []byte("2026-10-20T00:00:09Z listening on IPv4 address secret\n2026-10-20T00:00:11Z database system is ready")},
			{stream: "stdout", data: []byte("2026-10-20T00:00:12Z ignored\n")},
			{stream: "stderr", data: []byte(" to accept connections\n2026-10-20T00:00:08Z FATAL: secret\n")},
		}, true)
		if len(d.StartupLogEvents) != 4 {
			t.Fatal(d.StartupLogEvents)
		}
		for i, event := range d.StartupLogEvents {
			if i < 3 && event.ContainerElapsedMS == nil {
				t.Fatal("merged stream order treated as clock regression", event)
			}
			if i == 3 && event.ContainerElapsedMS != nil {
				t.Fatal("within-stream regression not recognized")
			}
		}
		if d.StartupLogEvents[1].Stream != "stderr" || d.StartupLogEvents[2].StreamSequence != 2 {
			t.Fatal("stream provenance lost")
		}
		d = newStartupDiagnostic(t.Context())
		d.observeStartupLogStreams([]postgresLogChunk{{stream: "stderr", data: []byte("FATAL: incomplete")}}, false)
		if len(d.StartupLogEvents) != 0 {
			t.Fatal("truncated partial line interpreted")
		}
	})
}
func boolPointer(value bool) *bool { return &value }
