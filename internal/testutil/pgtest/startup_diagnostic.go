package pgtest

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

// The primary acquisition error and the single cleanup result remain distinct.
// A failed cleanup returns the partial Harness; Close returns the same receipt.
type postgresStartupFailure struct {
	primary, cleanup error
	diagnostic       *startupDiagnostic
}

func (e *postgresStartupFailure) Error() string {
	return "postgres acquisition failed: " + e.diagnostic.Outcome
}
func (e *postgresStartupFailure) Unwrap() []error { return []error{e.primary, e.cleanup} }
func (e *postgresStartupFailure) RetryForbidden() bool {
	return e.cleanup != nil || e.diagnostic.Stage == "sql_connectivity" || e.diagnostic.Stage == "security_capability" || e.diagnostic.Stage == "generation_initialization"
}

type startupDiagnostic struct {
	suiteservices.PostgresReadinessDiagnostic
	ctx               context.Context
	phaseContext      context.Context
	start, stageStart time.Time
	endpoint          *endpointTrace
}

func newStartupDiagnostic(ctx context.Context) *startupDiagnostic {
	now := time.Now()
	d := &startupDiagnostic{PostgresReadinessDiagnostic: suiteservices.PostgresReadinessDiagnostic{
		SchemaID: "cartulary.postgres_readiness_diagnostic.v4", SQLWindowMS: postgresClientReadyTimeout.Milliseconds(), PollMS: postgresHealthPollInterval.Milliseconds(), ProbeCapMS: postgresClientAttemptTimeout.Milliseconds(), ContainerState: "unavailable", Phases: []suiteservices.PostgresReadinessPhase{}, ProbeFailures: map[string]int{}, StartupLogStatus: "not_collected", StartupLogSignals: []string{}, StartupLogClock: "docker_wall", StartupLogEvents: []suiteservices.PostgresStartupLogEvent{}, ProbeObservations: []suiteservices.PostgresProbeObservation{},
	}, ctx: ctx, start: now, stageStart: now}
	d.endpoint = newEndpointTrace(&d.Endpoint, now)
	return d
}

var sqlStatePattern = regexp.MustCompile(`^[0-9A-Z]{5}$`)

// Classification never publishes Error(), which can contain endpoints, paths,
// credentials, certificate identities or server-controlled text.
func postgresProbeCause(err error) (string, string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		state := pgErr.Code
		if !sqlStatePattern.MatchString(state) {
			state = ""
		}
		switch state {
		case "57P03":
			return "server_starting", state
		case "53300":
			return "connection_limit", state
		case "28P01", "28000":
			return "authentication_rejected", state
		case "42501":
			return "capability_rejected", state
		default:
			return "server_error", state
		}
	}
	var verification *tls.CertificateVerificationError
	var unknownCA x509.UnknownAuthorityError
	var invalidCert x509.CertificateInvalidError
	var hostname x509.HostnameError
	var record tls.RecordHeaderError
	var config *pgconn.ParseConfigError
	var file *os.PathError
	var network net.Error
	switch {
	case errors.As(err, &verification), errors.As(err, &unknownCA), errors.As(err, &invalidCert), errors.As(err, &hostname), errors.As(err, &record):
		return "tls_rejected", ""
	case errors.As(err, &config), errors.As(err, &file):
		return "local_configuration", ""
	case errors.Is(err, syscall.ECONNREFUSED):
		return "connection_refused", ""
	case errors.Is(err, syscall.ECONNRESET), errors.Is(err, syscall.EPIPE):
		return "connection_reset", ""
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return "connection_closed", ""
	case errors.Is(err, context.Canceled):
		return "cancelled", ""
	case errors.Is(err, context.DeadlineExceeded):
		return "probe_timeout", ""
	case errors.As(err, &network):
		if network.Timeout() {
			return "probe_timeout", ""
		}
		return "transport_unreachable", ""
	default:
		return "unknown", ""
	}
}

func (d *startupDiagnostic) recordProbeFailure(err error) {
	cause, state := postgresProbeCause(err)
	d.ProbeFailures[cause]++
	d.LastProbeFailure, d.LastSQLState = cause, state
	if len(d.ProbeObservations) < 128 {
		d.ProbeObservations = append(d.ProbeObservations, suiteservices.PostgresProbeObservation{ElapsedMS: time.Since(d.start).Milliseconds(), Cause: cause})
	} else {
		d.ProbeObservationsTruncated = true
	}
}

var postgresStartupLogsFn = readPostgresStartupLogs

// Request timestamps from the same Docker endpoint used by the container owner.
// Only the exact acquired container is read, never a name search or foreign PID.
func readPostgresStartupLogs(ctx context.Context, container testcontainers.Container) ([]postgresLogChunk, bool, error) {
	id := container.GetContainerID()
	if id == "" {
		return nil, false, errors.New("missing container identity")
	}
	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		return nil, false, err
	}
	defer cli.Close()
	logs, err := cli.ContainerLogs(ctx, id, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Timestamps: true})
	if err != nil {
		return nil, false, err
	}
	defer logs.Close()
	return readPostgresLogStreams(logs)
}

// This request never enables a TTY. Bound decoded bytes, frame count and reads
// before allocating from a provider-supplied frame length; no reader goroutine
// may outlive the two-second observation context or physical cleanup.
type postgresLogChunk struct {
	stream string
	data   []byte
}

func readPostgresLogFrames(r io.Reader) ([]byte, bool, error) {
	chunks, truncated, err := readPostgresLogStreams(r)
	var output bytes.Buffer
	for _, chunk := range chunks {
		output.Write(chunk.data)
	}
	return output.Bytes(), truncated, err
}
func readPostgresLogStreams(r io.Reader) ([]postgresLogChunk, bool, error) {
	const limit = 64 * 1024
	chunks := []postgresLogChunk{}
	total := 0
	for frames := 0; frames < 1024; frames++ {
		var header [8]byte
		n, err := io.ReadFull(r, header[:])
		if err == io.EOF && n == 0 {
			return chunks, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		if (header[0] != 1 && header[0] != 2) || header[1] != 0 || header[2] != 0 || header[3] != 0 {
			return nil, false, errors.New("invalid Docker log frame")
		}
		size := int64(binary.BigEndian.Uint32(header[4:]))
		remaining := int64(limit - total)
		data := make([]byte, int(min(size, remaining)))
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, false, err
		}
		stream := "stdout"
		if header[0] == 2 {
			stream = "stderr"
		}
		chunks = append(chunks, postgresLogChunk{stream: stream, data: data})
		total += len(data)
		if size > remaining {
			return chunks, true, nil
		}
	}
	return chunks, true, nil
}

// Observation precedes termination and has its own bounded budget. Failure to
// observe is evidence absence, never a replacement failure or cleanup authority.
func (d *startupDiagnostic) observeContainer(container testcontainers.Container) {
	d.StartupLogStatus = "unavailable"
	if container == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if state, err := container.State(ctx); err == nil && state != nil {
		status := string(state.Status)
		if slices.Contains([]string{"created", "running", "paused", "restarting", "removing", "exited", "dead"}, status) {
			d.ContainerState = status
		}
		d.ContainerExitCode, d.ContainerOOMKilled = &state.ExitCode, &state.OOMKilled
		if started, err := time.Parse(time.RFC3339Nano, state.StartedAt); err == nil && !started.IsZero() {
			value := started.UTC().Format(time.RFC3339Nano)
			d.ContainerStartedAt = &value
		}
	}
	if ctx.Err() != nil {
		return
	}
	d.endpoint.observeFailure(ctx, container)
	if ctx.Err() != nil {
		return
	}
	data, truncated, err := postgresStartupLogsFn(ctx, container)
	if err != nil {
		return
	}
	d.StartupLogStatus = "complete"
	if truncated {
		d.StartupLogStatus = "truncated"
	}
	d.observeStartupLogStreams(data, !truncated)
}

func (d *startupDiagnostic) observeStartupLogEvents(logs string) {
	d.observeStartupLogStreams([]postgresLogChunk{{stream: "unknown", data: []byte(logs)}}, true)
}
func (d *startupDiagnostic) observeStartupLogStreams(chunks []postgresLogChunk, complete bool) {
	type streamLine struct {
		stream, text string
		sequence     int
	}
	lines := []streamLine{}
	buffers := map[string]string{}
	sequences := map[string]int{}
	appendLine := func(stream, line string) {
		sequences[stream]++
		lines = append(lines, streamLine{stream: stream, text: line, sequence: sequences[stream]})
	}
	for _, chunk := range chunks {
		stream := chunk.stream
		buffer := buffers[stream] + string(chunk.data)
		for {
			line, rest, ok := strings.Cut(buffer, "\n")
			if !ok {
				break
			}
			appendLine(stream, line)
			buffer = rest
		}
		buffers[stream] = buffer
	}
	if complete {
		for _, stream := range []string{"stdout", "stderr", "unknown"} {
			if buffers[stream] != "" {
				appendLine(stream, buffers[stream])
			}
		}
	}

	phase := "unknown"
	previous := map[string]time.Time{}
	var started time.Time
	if d.ContainerStartedAt != nil {
		started, _ = time.Parse(time.RFC3339Nano, *d.ContainerStartedAt)
	}
	regressed := map[string]bool{}
	for _, observed := range lines {
		line := observed.text
		var timestamp *string
		var elapsed *int64
		prefix, message, hasPrefix := strings.Cut(line, " ")
		if value, err := time.Parse(time.RFC3339Nano, prefix); hasPrefix && err == nil {
			line = message
			text := value.UTC().Format(time.RFC3339Nano)
			timestamp = &text
			if !previous[observed.stream].IsZero() && value.Before(previous[observed.stream]) {
				regressed[observed.stream] = true
			}
			previous[observed.stream] = value
			if !regressed[observed.stream] && !started.IsZero() && !value.Before(started) && value.Sub(started) <= 10*time.Minute {
				ms := value.Sub(started).Milliseconds()
				elapsed = &ms
			}
		}
		for _, marker := range []struct{ text, signal string }{
			{"The files belonging to this database system will be owned by user", "initdb_started"},
			{"syncing data to disk", "initdb_sync_started"},
			{"syncing data to disk ... ok", "initdb_sync_completed"},
			{"Success. You can now start the database server", "initdb_complete"},
			{"starting temporary server", "temporary_server_start"},
			{"PostgreSQL init process complete", "initialization_complete"},
			{"Skipping initialization", "initialization_skipped"},
			{"listening on IPv4 address", "tcp_listener_started"},
			{"listening on IPv6 address", "tcp_listener_started"},
			{"database system is ready to accept connections", "server_ready"},
			{"database system is shut down", "server_shutdown"},
			{"FATAL:", "fatal"}, {"PANIC:", "panic"},
			{"No space left on device", "storage_exhausted"},
			{"Permission denied", "permission_denied"},
			{"out of memory", "memory_exhausted"},
			{"could not load private key file", "tls_file_error"},
			{"could not load server certificate file", "tls_file_error"},
		} {
			if strings.Contains(line, marker.text) {
				if len(d.StartupLogEvents) == 128 {
					d.StartupLogStatus = "truncated"
					return
				}
				switch marker.signal {
				case "initdb_started", "initdb_sync_started", "initdb_sync_completed", "initdb_complete", "temporary_server_start":
					phase = "bootstrap"
				case "initialization_complete", "initialization_skipped":
					phase = "final_server"
				}
				if !slices.Contains(d.StartupLogSignals, marker.signal) {
					d.StartupLogSignals = append(d.StartupLogSignals, marker.signal)
				}
				d.StartupLogEvents = append(d.StartupLogEvents, suiteservices.PostgresStartupLogEvent{Stream: observed.stream, StreamSequence: observed.sequence, Signal: marker.signal, Phase: phase, Timestamp: timestamp, ContainerElapsedMS: elapsed})
			}
		}
	}
}
func (d *startupDiagnostic) phase(stage string, outcome string) {
	configured := int64(0)
	switch stage {
	case "endpoint_resolution":
		configured = postgresPortMappingTimeout.Milliseconds()
	case "sql_connectivity":
		configured = postgresClientReadyTimeout.Milliseconds()
	case "security_capability":
		configured = postgresClientAttemptTimeout.Milliseconds()
	}
	effective, limiting := configured, "phase"
	phaseContext := d.phaseContext
	if phaseContext == nil {
		phaseContext = d.ctx
	}
	if deadline, ok := phaseContext.Deadline(); ok {
		remaining := max(int64(0), deadline.Sub(d.stageStart).Milliseconds())
		if configured == 0 || remaining < configured {
			effective, limiting = remaining, "parent"
		}
	} else if configured == 0 {
		limiting = "none"
	}
	d.Phases = append(d.Phases, suiteservices.PostgresReadinessPhase{Stage: stage, Outcome: outcome, ConfiguredMS: configured, EffectiveMS: effective, LimitingDeadline: limiting, ElapsedMS: max(int64(0), time.Since(d.stageStart).Milliseconds())})
	d.stageStart = time.Now()
	d.phaseContext = nil
}
func (d *startupDiagnostic) advance(stage string) { d.phase(stage, "ready") }
func (d *startupDiagnostic) finish(stage string, err error) {
	d.Stage = stage
	d.Outcome = "unavailable"
	var pgErr *pgconn.PgError
	var transport net.Error
	var ready *postgresReadinessError
	switch {
	case errors.Is(d.ctx.Err(), context.Canceled), errors.Is(err, context.Canceled):
		d.Outcome = "cancelled"
	case errors.As(err, &pgErr) && (pgErr.Code == "28P01" || pgErr.Code == "28000"):
		d.Outcome = "authentication_rejected"
	case errors.As(err, &pgErr) && pgErr.Code == "42501":
		d.Outcome = "capability_rejected"
	case errors.As(err, &ready) && ready.DeadlineExpired, errors.Is(err, context.DeadlineExceeded):
		d.Outcome = "deadline_expired"
	case errors.As(err, &transport):
		d.Outcome = "transport_unreachable"
	case stage == "security_capability":
		d.Outcome = "capability_rejected"
	case stage == "generation_initialization":
		d.Outcome = "initialization_failed"
	}
	if errors.As(err, &ready) {
		d.ProbeCount = ready.Attempts
	}
	d.ElapsedMS = max(int64(0), time.Since(d.start).Milliseconds())
	d.phase(stage, d.Outcome)
}

func ReadinessDiagnosticFromError(err error) (suiteservices.PostgresReadinessDiagnostic, bool) {
	var failure *postgresStartupFailure
	if !errors.As(err, &failure) {
		return suiteservices.PostgresReadinessDiagnostic{}, false
	}
	return failure.diagnostic.PostgresReadinessDiagnostic, true
}
