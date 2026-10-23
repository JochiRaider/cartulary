package pgtest

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

// Explicit-only qualification uses the existing four-Postgres resource profile.
// Each invocation has 64 fresh lifetimes; no failed lifetime is replaced.
func TestPublishedPostgresTransportQualification(t *testing.T) {
	admitted, err := strconv.Atoi(os.Getenv("CARTULARY_UNIT_CPU_TOKENS"))
	parallel, parseErr := strconv.Atoi(flag.Lookup("test.parallel").Value.String())
	if err != nil || parseErr != nil || admitted < 4 || parallel < 4 {
		t.Fatal("qualification requires four admitted PostgreSQL lanes and four native test workers")
	}
	if suiteservices.LookupEnvValue(nil, suiteservices.PGAdminDSNEnv) != "" {
		t.Fatal("self-owned qualification acquired an unused broker cluster")
	}
	verifyClient, err := testcontainers.NewDockerClientWithOpts(t.Context())
	if err != nil {
		t.Fatal("qualification inspection client unavailable")
	}
	t.Cleanup(func() { _ = verifyClient.Close() })
	seen := sync.Map{}
	for i := range 16 {
		t.Run(fmt.Sprintf("serial/%02d", i+1), func(t *testing.T) {
			qualifyPublishedPostgres(t, verifyClient, &seen, nil, nil, nil)
		})
	}
	for wave := range 8 {
		t.Run(fmt.Sprintf("concurrent/%02d", wave+1), func(t *testing.T) {
			for lane := range 4 {
				t.Run(fmt.Sprintf("lane-%d", lane+1), func(t *testing.T) {
					t.Parallel()
					qualifyPublishedPostgres(t, verifyClient, &seen, nil, nil, nil)
				})
			}
		})
	}
	for pair := range 8 {
		t.Run(fmt.Sprintf("competing/%02d", pair+1), func(t *testing.T) {
			ready, created := make(chan struct{}), make(chan struct{})
			readyOnce, createdOnce := sync.Once{}, sync.Once{}
			markReady := func() { readyOnce.Do(func() { close(ready) }) }
			markCreated := func() { createdOnce.Do(func() { close(created) }) }
			t.Run("retiring", func(t *testing.T) {
				t.Parallel()
				defer markReady()
				qualifyPublishedPostgres(t, verifyClient, &seen, nil, markReady, func() {
					select {
					case <-created:
					case <-t.Context().Done():
						t.Error("competing creation cancelled")
					}
				})
			})
			t.Run("acquiring", func(t *testing.T) {
				t.Parallel()
				defer markCreated()
				select {
				case <-ready:
				case <-t.Context().Done():
					t.Fatal("competing readiness cancelled")
				}
				qualifyPublishedPostgres(t, verifyClient, &seen, markCreated, nil, nil)
			})
		})
	}
}

func qualifyPublishedPostgres(t *testing.T, verifyClient client.APIClient, seen *sync.Map, acquired, ready, beforeClose func()) {
	t.Helper()
	var diagnostic *suiteservices.PostgresReadinessDiagnostic
	attempts := 0
	settlement := "not_started"
	allocation := ""
	env := map[string]string{
		"CARTULARY_TEST_RESULTS_DIR": os.Getenv("CARTULARY_TEST_RESULTS_DIR"),
		"CARTULARY_TEST_RUN_ID":      os.Getenv("CARTULARY_TEST_RUN_ID"),
		suiteservices.TargetEnv:      "test-slice",
		suiteservices.SuiteIDEnv:     "transport-" + suiteservices.ShortHash(t.Name(), 32),
	}
	record := func(event suiteservices.Event) {
		if err := suiteservices.RecordEvent(env, event); err != nil {
			t.Error("qualification journal publication failed")
		}
	}
	defer func() {
		if diagnostic != nil {
			event := suiteservices.Event{Type: suiteservices.EventPostgresReady, Service: suiteservices.ServicePostgres, Details: map[string]any{suiteservices.PostgresReadinessExtensionKey: *diagnostic}}
			if diagnostic.Outcome != "ready" {
				event.Type = suiteservices.EventFailureRecorded
				event.Details["failure_class"] = "infra"
				event.Details["failure_reason"] = "service_start_error"
				if diagnostic.Outcome == "deadline_expired" {
					event.Details["failure_reason"] = "service_readiness_timeout"
				}
			}
			record(event)
		}
		status, exit := "passed", 0
		if t.Failed() {
			status, exit = "failed", 1
		}
		physical := "unresolved"
		if settlement == "completed" {
			physical = "completed"
		}
		record(suiteservices.Event{Type: suiteservices.EventCleanupCompleted, Service: suiteservices.ServicePostgres, Name: t.Name(), Status: status, Details: map[string]any{"physical_settlement": physical, "child_exit_status": exit, "attempts": attempts, "allocation_ref": allocation}})
		if err := suiteservices.RefreshSummary(env); err != nil {
			t.Error("qualification summary publication failed")
		}
		events, present, err := suiteservices.ReadJournalEvents(env)
		if err != nil || !present || len(events) != 4 {
			t.Error("qualification lifetime evidence incomplete")
		}
	}()
	h, err := StartOwnedWithOptions(t.Context(), StartOptions{
		Observer: func(event testcontainersx.StartEvent) {
			if event.Type == testcontainersx.StartEventAttemptStart {
				attempts++
			}
		},
		Acquired: func(h *Harness) error {
			allocation = h.ContainerID()
			record(suiteservices.Event{Type: suiteservices.EventWrapperOwnedStart, Service: suiteservices.ServicePostgres, Name: t.Name(), Details: map[string]any{"postgres_allocation_ref": allocation}})
			if acquired != nil {
				acquired()
			}
			return nil
		},
		Readiness: func(value suiteservices.PostgresReadinessDiagnostic) {
			diagnostic = &value
			record(suiteservices.Event{Type: suiteservices.EventServiceStarted, Service: suiteservices.ServicePostgres})
		},
	})
	if h != nil {
		defer func() {
			if beforeClose != nil {
				beforeClose()
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if h.Close(ctx) != nil {
				settlement = "unresolved"
				t.Error("exact allocation removal failed")
				return
			}
			_, inspectErr := verifyClient.ContainerInspect(ctx, allocation, client.ContainerInspectOptions{})
			if !errdefs.IsNotFound(inspectErr) {
				settlement = "unproven"
				t.Errorf("exact allocation absence unproven (inspection error type %T)", inspectErr)
				return
			}
			settlement = "completed"
		}()
	}
	if err != nil {
		var failure *postgresStartupFailure
		if errors.As(err, &failure) {
			value := failure.diagnostic.PostgresReadinessDiagnostic
			diagnostic = &value
			settlement = value.CleanupOutcome
		}
		t.Fatal("owned PostgreSQL acquisition failed; normalized diagnostic retained")
	}
	if ready != nil {
		ready()
	}
	if attempts != 1 || allocation == "" {
		t.Error("qualification requires exactly one fresh allocation attempt")
	}
	if _, duplicate := seen.LoadOrStore(allocation, true); duplicate {
		t.Error("allocation identity was reused")
	}
	if diagnostic == nil {
		t.Fatal("readiness observation missing")
	}
	payload, err := json.Marshal(diagnostic)
	if err != nil {
		t.Fatal("readiness encoding failed")
	}
	value, err := harnessschema.Decode(payload)
	if err != nil || harnessschema.Validate(diagnostic.SchemaID, value) != nil {
		t.Fatal("readiness contract invalid")
	}
	p := diagnostic.Publication
	e := diagnostic.Endpoint
	if diagnostic.Outcome != "ready" || diagnostic.FirstSQLConnectionMS == nil || p.Status != "observed" || p.PublishAllPorts == nil || *p.PublishAllPorts || !reflect.DeepEqual(p.RequestedPorts, []string{"5432/tcp"}) || !reflect.DeepEqual(p.EffectivePorts, p.RequestedPorts) || len(p.Bindings) != 1 || p.Bindings[0].PortKind != "ephemeral" || p.Bindings[0].AddressScope != "unspecified" || p.Bindings[0].ContainerPort != "5432/tcp" {
		t.Error("publication or authenticated readiness differs from frozen control")
	}
	if e.InitialStatus != "observed" || e.Truncated || !slices.ContainsFunc(e.InitialBindings, func(binding suiteservices.PostgresBindingObservation) bool { return binding.MatchesSelectedPort }) {
		t.Error("exact endpoint binding coverage incomplete")
	}
}
