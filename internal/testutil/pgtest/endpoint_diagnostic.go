package pgtest

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/jackc/pgx/v5"
	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

type endpointTraceKey struct{}
type endpointTrace struct {
	mu                      sync.Mutex
	output                  *suiteservices.PostgresEndpointDiagnostic
	start                   time.Time
	host, port, containerID string
	references              map[string]string
	resolved                map[string]bool
	initial                 []string
	listener                func(context.Context, testcontainers.Container) string
	dial                    func(context.Context, string) (net.Conn, error)
}

func newEndpointTrace(output *suiteservices.PostgresEndpointDiagnostic, start time.Time) *endpointTrace {
	*output = suiteservices.PostgresEndpointDiagnostic{InitialStatus: "not_collected", FailureStatus: "not_collected", InitialBindings: []suiteservices.PostgresBindingObservation{}, FailureBindings: []suiteservices.PostgresBindingObservation{}, Lookups: []suiteservices.PostgresLookupObservation{}, Dials: []suiteservices.PostgresDialObservation{}, InternalListener: "not_collected"}
	return &endpointTrace{output: output, start: start, listener: readInternalPostgresListener, dial: dialPublishedPostgres, references: map[string]string{}, resolved: map[string]bool{}}
}
func endpointCategory(host string) (string, string) {
	address, err := netip.ParseAddr(host)
	if err != nil {
		if host == "" {
			return "unknown", "unknown"
		}
		return "hostname", "name"
	}
	family := "ipv6"
	if address.Is4() {
		family = "ipv4"
	}
	address = address.Unmap()
	switch {
	case address.IsLoopback():
		return family, "loopback"
	case address.IsUnspecified():
		return family, "unspecified"
	case address.IsPrivate():
		return family, "private"
	case address.IsLinkLocalUnicast():
		return family, "link_local"
	default:
		return family, "public"
	}
}
func (e *endpointTrace) endpoint(host, port string) suiteservices.PostgresEndpoint {
	family, category := endpointCategory(host)
	key := net.JoinHostPort(host, port)
	ref := e.references[key]
	if ref == "" {
		if len(e.references) >= 32 || len(key) > 512 {
			e.output.Truncated = true
			ref = "omitted"
		} else {
			ref = fmt.Sprintf("endpoint-%d", len(e.references)+1)
			e.references[key] = ref
		}
	}
	return suiteservices.PostgresEndpoint{Ref: ref, Family: family, Category: category}
}
func (e *endpointTrace) selectEndpoint(ctx context.Context, c testcontainers.Container, host, port string) {
	e.host, e.port, e.containerID = host, port, c.GetContainerID()
	selected := e.endpoint(host, port)
	e.output.Selected = &selected
	e.inspect(ctx, c, true)
}
func (e *endpointTrace) inspect(parent context.Context, c testcontainers.Container, initial bool) {
	ctx, cancel := context.WithTimeout(parent, 250*time.Millisecond)
	defer cancel()
	response, err := c.Inspect(ctx)
	e.mu.Lock()
	defer e.mu.Unlock()
	status := "unavailable"
	bindings := []suiteservices.PostgresBindingObservation{}
	keys := []string{}
	if err == nil && response != nil {
		matched := response.ID == e.containerID && e.containerID != ""
		if !initial {
			e.output.ContainerIdentityMatches = &matched
		}
		if matched && response.NetworkSettings != nil {
			status = "observed"
			for port, values := range response.NetworkSettings.Ports {
				if port.String() != "5432/tcp" {
					continue
				}
				for _, b := range values {
					if len(bindings) >= 32 {
						status = "truncated"
						e.output.Truncated = true
						break
					}
					number, parseErr := strconv.Atoi(b.HostPort)
					if parseErr != nil || number < 1 || number > 65535 || !b.HostIP.IsValid() {
						status = "unavailable"
						break
					}
					keys = append(keys, net.JoinHostPort(b.HostIP.String(), b.HostPort))
					bindings = append(bindings, suiteservices.PostgresBindingObservation{Endpoint: e.endpoint(b.HostIP.String(), b.HostPort), MatchesSelectedPort: b.HostPort == e.port})
				}
			}
		}
	}
	slices.Sort(keys)
	if initial {
		e.output.InitialStatus = status
		e.output.InitialBindings = bindings
		e.initial = keys
		return
	}
	e.output.FailureStatus = status
	e.output.FailureBindings = bindings
	if status == "observed" {
		published := slices.ContainsFunc(bindings, func(b suiteservices.PostgresBindingObservation) bool { return b.MatchesSelectedPort })
		e.output.SelectedPortPublished = &published
		if e.output.InitialStatus == "observed" {
			changed := !slices.Equal(keys, e.initial)
			e.output.BindingChanged = &changed
		}
	}
}
func (e *endpointTrace) observeConfig(config *pgx.ConnConfig) {
	lookup, dial := config.LookupFunc, config.DialFunc
	config.LookupFunc = func(ctx context.Context, host string) ([]string, error) {
		addresses, err := lookup(ctx, host)
		e.mu.Lock()
		defer e.mu.Unlock()
		if len(e.output.Lookups) >= 128 {
			e.output.Truncated = true
			return addresses, err
		}
		result := suiteservices.PostgresLookupObservation{ElapsedMS: time.Since(e.start).Milliseconds(), SelectedHost: host == e.host, Outcome: "resolved", Addresses: []suiteservices.PostgresEndpoint{}}
		if err != nil {
			result.Outcome, _ = postgresProbeCause(err)
		}
		for _, address := range addresses {
			if len(result.Addresses) >= 32 {
				e.output.Truncated = true
				break
			}
			result.Addresses = append(result.Addresses, e.endpoint(address, e.port))
			if host == e.host && len(e.resolved) < 32 {
				e.resolved[address] = true
			}
		}
		e.output.Lookups = append(e.output.Lookups, result)
		return addresses, err
	}
	config.DialFunc = func(ctx context.Context, network, address string) (net.Conn, error) {
		connection, err := dial(ctx, network, address)
		e.mu.Lock()
		defer e.mu.Unlock()
		if len(e.output.Dials) < 128 {
			e.output.Dials = append(e.output.Dials, e.dialObservation(address, connection, err))
		} else {
			e.output.Truncated = true
		}
		return connection, err
	}
}
func (e *endpointTrace) dialObservation(address string, connection net.Conn, err error) suiteservices.PostgresDialObservation {
	host, port, _ := net.SplitHostPort(address)
	result := suiteservices.PostgresDialObservation{ElapsedMS: time.Since(e.start).Milliseconds(), Endpoint: e.endpoint(host, port), MatchesSelectedPort: port == e.port, FromLookup: e.resolved[host], Outcome: "connected"}
	if err != nil {
		result.Outcome, _ = postgresProbeCause(err)
	}
	if connection != nil && connection.RemoteAddr() != nil {
		host, port, _ := net.SplitHostPort(connection.RemoteAddr().String())
		peer := e.endpoint(host, port)
		result.Peer = &peer
	}
	return result
}

func dialPublishedPostgres(ctx context.Context, address string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, "tcp", address)
}

func (e *endpointTrace) observeFailure(ctx context.Context, c testcontainers.Container) {
	if e.host == "" || e.port == "" {
		return
	}
	e.inspect(ctx, c, false)
	if e.output.ContainerIdentityMatches == nil || !*e.output.ContainerIdentityMatches {
		return
	}
	internal, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	e.output.InternalListener = e.listener(internal, c)
	cancel()
	if ctx.Err() != nil {
		return
	}
	probe, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	address := net.JoinHostPort(e.host, e.port)
	connection, err := e.dial(probe, address)
	e.mu.Lock()
	observation := e.dialObservation(address, connection, err)
	e.output.PublishedDial = &observation
	e.mu.Unlock()
	if connection != nil {
		connection.Close()
	}
	cancel()
}

func readInternalPostgresListener(ctx context.Context, c testcontainers.Container) string {
	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		return "unavailable"
	}
	defer cli.Close()
	process, err := cli.ExecCreate(ctx, c.GetContainerID(), client.ExecCreateOptions{Cmd: []string{"/bin/busybox", "timeout", "1", "/bin/cat", "/proc/net/tcp", "/proc/net/tcp6"}, AttachStdout: true})
	if err != nil {
		return "unavailable"
	}
	stream, err := cli.ExecAttach(ctx, process.ID, client.ExecAttachOptions{})
	if err != nil {
		return "unavailable"
	}
	defer stream.Close()
	if deadline, ok := ctx.Deadline(); ok {
		if err := stream.Conn.SetReadDeadline(deadline); err != nil {
			return "unavailable"
		}
	}
	data, truncated, err := readPostgresLogFrames(stream.Reader)
	if err != nil {
		return "unavailable"
	}
	if truncated {
		return "truncated"
	}
	return postgresListenerFromProc(string(data))
}
func postgresListenerFromProc(data string) string {
	recognized := false
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if fields[1] == "local_address" {
			recognized = true
		}
		if strings.HasSuffix(fields[1], ":1538") && fields[3] == "0A" {
			return "present"
		}
	}
	if recognized {
		return "absent"
	}
	return "unavailable"
}
