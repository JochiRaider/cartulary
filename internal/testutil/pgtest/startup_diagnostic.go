package pgtest

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/jackc/pgx/v5/pgconn"
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
}

func newStartupDiagnostic(ctx context.Context) *startupDiagnostic {
	now := time.Now()
	return &startupDiagnostic{PostgresReadinessDiagnostic: suiteservices.PostgresReadinessDiagnostic{
		SchemaID: "cartulary.postgres_readiness_diagnostic.v1", SQLWindowMS: postgresClientReadyTimeout.Milliseconds(), PollMS: postgresHealthPollInterval.Milliseconds(), ProbeCapMS: postgresClientAttemptTimeout.Milliseconds(), ContainerState: "unavailable", Phases: []suiteservices.PostgresReadinessPhase{},
	}, ctx: ctx, start: now, stageStart: now}
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
