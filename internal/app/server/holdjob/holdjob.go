// Package holdjob is the TH-HARNESS-REQ-450 hold-job test control. It supplies
// one harness-only job kind whose handler holds until it is cancelled or
// released, so a server-harness process can be stopped with an active job
// (Core 04 AC-604). Only the harness server profile imports it; production
// binaries never link it.
package holdjob

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

const (
	JobKind        = "harness.hold_job_v1"
	WorkerKind     = "harness.hold_job_worker_v1"
	progressUnitID = "harness.hold_job.steps.v1"
	controlSchema  = "cartulary.test.hold_job_control.v1"
	payloadSchema  = "cartulary.test.hold_job_payload.v1"
	routePath      = "/api/v1/test/runtime/hold-jobs"

	// ModeRespectCancellation holds until the handler context is cancelled.
	ModeRespectCancellation = "respect_cancellation"
	// ModeIgnoreCancellation holds until released, whatever the context.
	ModeIgnoreCancellation = "ignore_cancellation_until_released"
)

// Control is process-scoped: it retains no state across processes.
type Control struct {
	mu      sync.Mutex
	ports   *jobs.ContributionPorts
	release chan struct{}
	once    sync.Once
	running atomic.Int64
}

func NewControl() *Control {
	return &Control{release: make(chan struct{})}
}

// Contribution adds the hold job to the process's Jobs catalog.
func (c *Control) Contribution() jobs.Contribution {
	return jobs.Contribution{
		Definition: jobs.Definition{
			JobKind:        JobKind,
			ProgressUnitID: progressUnitID,
			HandlerName:    WorkerKind,
			Terminal:       jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"hold_released"}},
		},
		Worker: jobs.WorkerRuntimeContract{ProfileID: "base", WorkerKind: WorkerKind, JobKinds: []string{JobKind}, MaxActiveAttemptsPerProcess: 8},
		Bind:   c.bind,
	}
}

func (c *Control) bind(ports jobs.ContributionPorts) jobs.HandlerFunc {
	c.mu.Lock()
	c.ports = &ports
	c.mu.Unlock()
	return func(ctx context.Context, execution jobs.Execution) error {
		raw, err := ports.Payload(ctx, execution)
		if err != nil {
			return err
		}
		var payload struct {
			SchemaID string `json:"schema_id"`
			Mode     string `json:"mode"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil || payload.SchemaID != payloadSchema {
			return errors.New("holdjob: invalid payload")
		}
		c.running.Add(1)
		defer c.running.Add(-1)
		switch payload.Mode {
		case ModeRespectCancellation:
			<-ctx.Done()
			return ctx.Err()
		case ModeIgnoreCancellation:
			<-c.release
			return errors.New("holdjob: released")
		default:
			return errors.New("holdjob: unknown mode")
		}
	}
}

// Routes registers the guarded control route when test routes are enabled.
func (c *Control) Routes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return fmt.Errorf("register hold-job route: %w", err)
		}
		// BindTestRoute applies the guard, then the enveloped 405 for any
		// other method (TH-HARNESS-REQ-450).
		return httpapi.BindTestRoute(mux, deps, guard, http.MethodPost, routePath, c.serve)
	}
}

type controlRequest struct {
	SchemaID          string     `json:"schema_id"`
	Operation         string     `json:"operation"`
	Mode              string     `json:"mode,omitempty"`
	SubmittedByUserID *uuid.UUID `json:"submitted_by_user_id,omitempty"`
}

func (c *Control) serve(w http.ResponseWriter, r *http.Request) {
	var request controlRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil || decoder.More() || request.SchemaID != controlSchema {
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_hold_job_control", "invalid hold-job control", nil)
		return
	}
	switch request.Operation {
	case "enqueue":
		if (request.Mode != ModeRespectCancellation && request.Mode != ModeIgnoreCancellation) || request.SubmittedByUserID == nil {
			_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_hold_job_control", "invalid hold-job enqueue", nil)
			return
		}
		c.mu.Lock()
		ports := c.ports
		c.mu.Unlock()
		if ports == nil {
			_ = httpapi.WriteError(w, r, http.StatusServiceUnavailable, "hold_job_unavailable", "hold job not assembled", nil)
			return
		}
		payload, _ := json.Marshal(map[string]string{"schema_id": payloadSchema, "mode": request.Mode})
		jobID, err := ports.Enqueue(r.Context(), *request.SubmittedByUserID, payload)
		if err != nil {
			_ = httpapi.WriteError(w, r, http.StatusInternalServerError, "hold_job_enqueue_failed", "hold-job enqueue failed", nil)
			return
		}
		writeJSON(w, map[string]any{"job_id": jobID.String()})
	case "release":
		if request.Mode != "" || request.SubmittedByUserID != nil {
			_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_hold_job_control", "invalid hold-job release", nil)
			return
		}
		c.once.Do(func() { close(c.release) })
		writeJSON(w, map[string]any{})
	case "status":
		if request.Mode != "" || request.SubmittedByUserID != nil {
			_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_hold_job_control", "invalid hold-job status", nil)
			return
		}
		writeJSON(w, map[string]any{"running": c.running.Load()})
	default:
		_ = httpapi.WriteError(w, r, http.StatusBadRequest, "invalid_hold_job_control", "unknown hold-job operation", nil)
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(value)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body.Bytes())
}
