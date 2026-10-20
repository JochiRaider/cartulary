package postgrescleanup

import (
	"context"
	"errors"
	"math"
	"regexp"
	"time"
)

// DropAttempt is shared by Go failure envelopes and browser lifecycle evidence.
// It contains no resource or backend identity and never changes the primary cause.
type DropAttempt struct {
	ServiceGeneration *string         `json:"service_generation"`
	DurationMS        int64           `json:"duration_ms"`
	AdmissionMS       int64           `json:"admission_ms"`
	AdmissionBudgetMS int64           `json:"admission_budget_ms"`
	OperationMS       int64           `json:"operation_ms"`
	OperationBudgetMS int64           `json:"operation_budget_ms"`
	Admission         string          `json:"admission"`
	Outcome           string          `json:"outcome"`
	Wait              DropObservation `json:"wait"`
}

var serviceGenerationPattern = regexp.MustCompile(`^[1-9][0-9]{0,19}$`)

func (o DropAttempt) Validate() error {
	if o.ServiceGeneration != nil && !serviceGenerationPattern.MatchString(*o.ServiceGeneration) {
		return errors.New("invalid PostgreSQL service generation")
	}
	if o.DurationMS < 0 || o.DurationMS > 86400000 || o.AdmissionMS < 0 || o.AdmissionMS > o.DurationMS ||
		o.OperationMS < 0 || o.OperationMS > o.DurationMS || o.AdmissionBudgetMS < 0 || o.AdmissionBudgetMS > 15000 || o.OperationBudgetMS < 0 || o.OperationBudgetMS > 20000 {
		return errors.New("invalid PostgreSQL cleanup timing")
	}
	switch o.Admission {
	case "acquired", "rejected", "not_attempted", "unavailable":
	default:
		return errors.New("invalid PostgreSQL admission outcome")
	}
	switch o.Outcome {
	case "completed", "failed", "cancelled", "deadline_exceeded", "unavailable":
	default:
		return errors.New("invalid PostgreSQL cleanup outcome")
	}
	if o.Admission != "acquired" && (o.OperationMS != 0 || o.OperationBudgetMS != 0 || o.Wait.Status == "sampled") {
		return errors.New("unadmitted PostgreSQL operation evidence")
	}
	return o.Wait.Validate()
}

type attemptError struct {
	err         error
	observation DropAttempt
}

func (e *attemptError) Error() string                { return e.err.Error() }
func (e *attemptError) Unwrap() error                { return e.err }
func (e *attemptError) PostgresCleanup() DropAttempt { return e.observation }

// ObservationFromError extracts only the shared validated structure. Invalid
// diagnostics are omitted without interpreting error text or changing its cause.
func ObservationFromError(err error) *DropAttempt {
	visited := 0
	var visit func(error) *DropAttempt
	visit = func(current error) *DropAttempt {
		if current == nil {
			return nil
		}
		visited++
		if visited > 32 {
			value := UnavailableAttempt()
			return &value
		}
		if observed, ok := current.(interface{ PostgresCleanup() DropAttempt }); ok {
			value := observed.PostgresCleanup()
			if value.Validate() != nil {
				value = UnavailableAttempt()
			}
			return &value
		}
		switch wrapped := current.(type) {
		case interface{ Unwrap() []error }:
			for _, child := range wrapped.Unwrap() {
				if result := visit(child); result != nil {
					return result
				}
			}
		case interface{ Unwrap() error }:
			return visit(wrapped.Unwrap())
		}
		return nil
	}
	return visit(err)
}

func UnavailableAttempt() DropAttempt {
	return DropAttempt{Admission: "unavailable", Outcome: "unavailable", Wait: DropObservation{Status: "unavailable"}}
}

func effectiveBudget(ctx context.Context, cap time.Duration) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		return max(0, min(cap, time.Until(deadline)))
	}
	return cap
}

// CheckpointInterval describes only the interval between sampled cumulative
// counters. Reset or backwards statistics cannot be reported as a valid delta.
type CheckpointInterval struct {
	Status    string  `json:"status"`
	Requested int64   `json:"requested"`
	Done      int64   `json:"done"`
	Buffers   int64   `json:"buffers"`
	WriteMS   float64 `json:"write_ms"`
	SyncMS    float64 `json:"sync_ms"`
}

func validMilliseconds(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= float64(maxDiagnosticCounter)
}
func (c CheckpointInterval) valid() bool {
	if c.Status != "sampled" {
		return (c.Status == "not_observed" || c.Status == "unavailable" || c.Status == "reset") && c.Requested == 0 && c.Done == 0 && c.Buffers == 0 && c.WriteMS == 0 && c.SyncMS == 0
	}
	return c.Requested >= 0 && c.Requested <= maxDiagnosticCounter && c.Done >= 0 && c.Done <= maxDiagnosticCounter && c.Buffers >= 0 && c.Buffers <= maxDiagnosticCounter && validMilliseconds(c.WriteMS) && validMilliseconds(c.SyncMS)
}
func checkpointInterval(first, last *DropWaitSnapshot) CheckpointInterval {
	if first == nil || last == nil {
		return CheckpointInterval{Status: "not_observed"}
	}
	if first.statsReset.IsZero() || last.statsReset.IsZero() {
		return CheckpointInterval{Status: "unavailable"}
	}
	if first.statsReset != last.statsReset || last.Requested < first.Requested || last.Done < first.Done || last.Buffers < first.Buffers || last.WriteMS < first.WriteMS || last.SyncMS < first.SyncMS {
		return CheckpointInterval{Status: "reset"}
	}
	return CheckpointInterval{Status: "sampled", Requested: last.Requested - first.Requested, Done: last.Done - first.Done, Buffers: last.Buffers - first.Buffers, WriteMS: last.WriteMS - first.WriteMS, SyncMS: last.SyncMS - first.SyncMS}
}
