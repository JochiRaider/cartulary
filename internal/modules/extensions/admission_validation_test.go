package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
)

func TestAdmissionAlgorithmResultAndDeadline_Unit(t *testing.T) {
	t.Run("revocable terminal completion", testTerminalCompletionGrace)
	t.Run("nested cleanup reuses original interruption", func(t *testing.T) {
		ctx, cancel := context.WithCancel(extensiondeadline.WithCancellationTracking(t.Context()))
		if extensiondeadline.GraceExhausted(ctx) || extensiondeadline.GraceExhausted(t.Context()) {
			t.Fatal("read manufactured exhaustion")
		}
		cancel()
		fatal := 0
		p := extensiondeadline.InvocationPolicy{Grace: 5 * time.Millisecond, Exhausted: func(error) { fatal++ }}
		if extensiondeadline.Sample(ctx, p) == nil {
			t.Fatal("missing interruption")
		}
		time.Sleep(10 * time.Millisecond)
		nested := extensiondeadline.ReuseCancellationTracking(context.WithValue(ctx, struct{}{}, "nested"))
		_, err := extensiondeadline.Complete(nested, p, func(context.Context) (struct{}, error) { return struct{}{}, nil })
		if !errors.Is(err, extensiondeadline.ErrGraceExhausted) || fatal != 1 || !extensiondeadline.GraceExhausted(nested) || !extensiondeadline.GraceExhausted(ctx) || extensiondeadline.GraceExhausted(t.Context()) {
			t.Fatal("cleanup obtained fresh grace", err, fatal)
		}
	})
	policy := extensiondeadline.InvocationPolicy{Grace: time.Second, Exhausted: func(err error) { t.Errorf("unexpected fatal invocation: %v", err) }}
	t.Run("canceled callback retains invocation until completion", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		entered := make(chan struct{})
		release := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			_, err := invokeAdmissionAlgorithm(ctx, func(context.Context, AdmissionValidationContext) (AdmissionValidationResult, error) {
				close(entered)
				<-release
				return AdmissionValidationResult{}, nil
			}, AdmissionValidationContext{}, policy)
			done <- err
		}()
		<-entered
		cancel()
		select {
		case err := <-done:
			close(release)
			t.Fatalf("invocation released while callback was active: %v", err)
		case <-time.After(20 * time.Millisecond):
		}
		close(release)
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("late callback result accepted: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("completed callback was not joined")
		}
	})
	for _, grace := range []time.Duration{0, 10 * time.Millisecond} {
		t.Run("grace_"+grace.String(), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			entered, release := make(chan struct{}), make(chan struct{})
			fatal := make(chan error, 1)
			done := make(chan error, 1)
			go func() {
				_, err := invokeAdmissionAlgorithm(ctx, func(context.Context, AdmissionValidationContext) (AdmissionValidationResult, error) {
					close(entered)
					<-release
					return AdmissionValidationResult{}, nil
				}, AdmissionValidationContext{}, extensiondeadline.InvocationPolicy{Grace: grace, Exhausted: func(err error) { fatal <- err }})
				done <- err
			}()
			<-entered
			cancel()
			select {
			case err := <-fatal:
				if !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
					t.Errorf("fatal cause = %v", err)
				}
			case <-time.After(time.Second):
				close(release)
				<-done
				t.Fatal("grace did not signal fatal termination")
			}
			select {
			case err := <-done:
				close(release)
				t.Fatalf("fatal released active callback: %v", err)
			default:
			}
			close(release)
			if err := <-done; !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Fatalf("late result replaced fatal outcome: %v", err)
			}
		})
	}

	t.Run("earlier cancellation remains cancellation after grace wait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		entered, sampled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
		var once sync.Once
		var now atomic.Int64
		now.Store(500)
		bound := extensiondeadline.Deadline{MonotonicNS: 1000}
		policy := extensiondeadline.InvocationPolicy{Grace: time.Second, Deadline: &bound, MonotonicNowNS: func() int64 {
			value := now.Load()
			if ctx.Err() != nil {
				once.Do(func() { close(sampled) })
			}
			return value
		}, Exhausted: func(err error) { t.Errorf("unexpected fatal signal: %v", err) }}
		done := make(chan error, 1)
		go func() {
			_, err := invokeAdmissionAlgorithm(ctx, func(context.Context, AdmissionValidationContext) (AdmissionValidationResult, error) {
				close(entered)
				<-release
				return AdmissionValidationResult{}, nil
			}, AdmissionValidationContext{}, policy)
			done <- err
		}()
		<-entered
		cancel()
		select {
		case <-sampled:
		case <-time.After(time.Second):
			close(release)
			t.Fatal("cancellation was not sampled")
		}
		now.Store(2000)
		close(release)
		if err := <-done; !errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("earlier cancellation changed after grace: %v", err)
		}
	})

	t.Run("transaction completion retains an earlier boundary sample", func(t *testing.T) {
		ctx, cancel := context.WithCancel(extensiondeadline.WithCancellationTracking(context.Background()))
		defer cancel()
		var now atomic.Int64
		now.Store(500)
		bound := extensiondeadline.Deadline{MonotonicNS: 1000}
		policy := extensiondeadline.InvocationPolicy{Grace: time.Second, Deadline: &bound, MonotonicNowNS: now.Load, Exhausted: func(err error) { t.Errorf("unexpected fatal: %v", err) }}
		cancel()
		if err := extensiondeadline.Sample(ctx, policy); !errors.Is(err, context.Canceled) {
			t.Fatalf("first boundary: %v", err)
		}
		now.Store(2000)
		result, err := extensiondeadline.Complete(ctx, policy, func(context.Context) (string, error) { return "absent", nil })
		if err != nil || result.Value != "absent" || !errors.Is(result.Interruption, context.Canceled) || errors.Is(result.Interruption, context.DeadlineExceeded) {
			t.Fatalf("completion changed earlier sample: %#v, %v", result, err)
		}
	})

	input := AdmissionValidationContext{Phase: "preflight"}
	for _, test := range []struct {
		name      string
		algorithm AdmissionAlgorithm
	}{
		{"missing", nil},
		{"panic", func(context.Context, AdmissionValidationContext) (AdmissionValidationResult, error) { panic("private") }},
		{"error", func(context.Context, AdmissionValidationContext) (AdmissionValidationResult, error) {
			return AdmissionValidationResult{}, errors.New("private")
		}},
		{"deadline", func(ctx context.Context, _ AdmissionValidationContext) (AdmissionValidationResult, error) {
			<-ctx.Done()
			return AdmissionValidationResult{}, ctx.Err()
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			if _, err := invokeAdmissionAlgorithm(ctx, test.algorithm, input, policy); err == nil {
				t.Fatal("invalid invocation accepted")
			}
		})
	}
}

func TestClaimAdmissionPreflightBeforeState_ServiceBacked(t *testing.T) {
	assertAdmissionPhaseOverflow(t)
	t.Run("dependency execution", verifyDependencyExecution)
	fixture := newNetworkFlowStateFixture(t, nil)
	calls := 0
	runtime := newNetworkFlowStateRuntime(t, fixture.store, &calls)
	coordinator := requireGeneratedCoordinator(t)
	resolution, err := coordinator.ResolveClaims([]string{"import", "network_flow_activity"})
	if err != nil {
		t.Fatal(err)
	}
	view, err := coordinator.ConfigurationView("network_flow_activity", []ProfileConfigurationValue{{Key: "network_flow_activity.key_ring_manifest_path", Source: "explicit", Value: map[string]any{"kind": "regular_file_ref", "key": "network_flow_activity.key_ring_manifest_path"}}})
	if err != nil {
		t.Fatal(err)
	}
	views := map[string]ProfileConfigurationView{"network_flow_activity": view}
	algorithmID := "network_flow_activity.incident_portability_cutover_v1"
	invocations := 0
	algorithm := func(_ context.Context, input AdmissionValidationContext) (AdmissionValidationResult, error) {
		invocations++
		raw, _ := json.Marshal(input)
		var members map[string]any
		_ = json.Unmarshal(raw, &members)
		if len(members) != 11 || input.StatePresent || input.StateMetadata != nil || len(input.MigrationDefinitions) != 0 || len(input.Configuration.Values) != 2 || input.Configuration.Values[1].Source != "default" || strings.Contains(string(raw), "/home/") {
			t.Errorf("incomplete or unsafe admission context: %s", raw)
		}
		return ValidAdmissionResult(input, algorithmID), nil
	}
	algorithms := map[string]AdmissionAlgorithm{algorithmID: algorithm}
	t.Run("production classification precedence", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			count       int
			schema      string
			invocation  bool
			nilFindings bool
			want        string
		}{
			{"empty", 0, "cartulary.extension_admission_validation_result.v1", false, false, ""},
			{"nil findings", 0, "cartulary.extension_admission_validation_result.v1", false, true, "extension_validation_result_invalid"},
			{"wrong identity", 0, "wrong", false, false, "extension_validation_result_invalid"},
			{"256", 256, "cartulary.extension_admission_validation_result.v1", false, false, "extension_admission_validation_failed"},
			{"257", 257, "cartulary.extension_admission_validation_result.v1", false, false, "extension_admission_validation_failed"},
			{"4096", 4096, "cartulary.extension_admission_validation_result.v1", false, false, "extension_admission_validation_failed"},
			{"overflow before identity and elements", 4097, "wrong", false, false, "extension_diagnostic_overflow"},
			{"invocation before overflow", 4097, "wrong", true, false, "extension_admission_validation_failed"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				fn := func(_ context.Context, input AdmissionValidationContext) (AdmissionValidationResult, error) {
					result := ValidAdmissionResult(input, algorithmID)
					result.SchemaID = tc.schema
					for i := 0; i < tc.count; i++ {
						f := admissionFailure(input, algorithmID, false)
						if tc.schema == "wrong" {
							f.Message = "private callback value"
						}
						result.Findings = append(result.Findings, f)
					}
					if tc.nilFindings {
						result.Findings = nil
					}
					if tc.invocation {
						return result, errors.New("private callback value")
					}
					return result, nil
				}
				err := runtime.ValidateClaimAdmission(context.Background(), coordinator, resolution, "preflight", views, map[string]AdmissionAlgorithm{algorithmID: fn})
				if tc.want == "" {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				var failure *StartupValidationError
				if !errors.As(err, &failure) || len(failure.Findings) == 0 || failure.Findings[0].ReasonCode != tc.want {
					t.Fatalf("classification = %#v / %v, want %s", failure, err, tc.want)
				}
				encoded, _ := json.Marshal(failure.Findings)
				if strings.Contains(string(encoded), "private callback value") {
					t.Fatal("callback value escaped")
				}
				if tc.want == "extension_diagnostic_overflow" && (len(failure.Findings) != 1 || failure.Findings[0].Details["actual"] != 4097) {
					t.Fatal("overflow did not replace findings at the first excess")
				}
			})
		}
	})
	if err := runtime.ValidateClaimAdmission(context.Background(), coordinator, resolution, "preflight", views, algorithms); err != nil {
		t.Fatal(err)
	}
	if invocations != 1 || calls != 0 {
		t.Fatalf("invocation count=%d state validators=%d", invocations, calls)
	}
	var count int
	if err := fixture.pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_state_metadata`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("preflight initialized state: %d %v", count, err)
	}
	for _, mode := range []string{"wrong phase", "wrong identity", "nil findings", "failure"} {
		t.Run(mode, func(t *testing.T) {
			algorithms[algorithmID] = func(_ context.Context, input AdmissionValidationContext) (AdmissionValidationResult, error) {
				result := ValidAdmissionResult(input, algorithmID)
				switch mode {
				case "wrong phase":
					result.Phase = "post_migration"
				case "wrong identity":
					result.AlgorithmID = "other"
				case "nil findings":
					result.Findings = nil
				case "failure":
					return result, errors.New("private retained data")
				}
				return result, nil
			}
			err := runtime.ValidateClaimAdmission(context.Background(), coordinator, resolution, "preflight", views, algorithms)
			var failure *StartupValidationError
			if !errors.As(err, &failure) || len(failure.Findings) != 1 || strings.Contains(err.Error(), "private") {
				t.Fatalf("wrong admission failure: %v", err)
			}
			if err := fixture.pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_state_metadata`).Scan(&count); err != nil || count != 0 {
				t.Fatal("failed preflight changed state")
			}
		})
	}
}

func assertAdmissionPhaseOverflow(t *testing.T) {
	t.Helper()
	fixture := newSyntheticStateFixture(t, nil)
	registered := requireGeneratedCoordinator(t)
	c := dependencyCoordinator(map[string][]string{"a": {}, "b": {}, "c": {}})
	c.validateFinding = registered.validateFinding
	c.validationConditions = registered.validationConditions
	resolution, err := c.ResolveClaims(c.orderedProfileIDs)
	if err != nil {
		t.Fatal(err)
	}
	views := map[string]ProfileConfigurationView{}
	algorithms := map[string]AdmissionAlgorithm{}
	calls := []string{}
	for _, id := range c.orderedProfileIDs {
		view, err := c.ConfigurationView(id, nil)
		if err != nil {
			t.Fatal(err)
		}
		views[id] = view
		algorithms[id+".admit"] = func(_ context.Context, input AdmissionValidationContext) (AdmissionValidationResult, error) {
			calls = append(calls, input.ProfileID)
			result := ValidAdmissionResult(input, input.ProfileID+".admit")
			count := 1
			if input.ProfileID == "a" {
				count = 4096
			}
			for i := 0; i < count; i++ {
				result.Findings = append(result.Findings, admissionFailure(input, result.AlgorithmID, false))
			}
			return result, nil
		}
	}
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{validationTimeout: 10 * time.Second})
	err = runtime.ValidateClaimAdmission(context.Background(), c, resolution, "preflight", views, algorithms)
	var failure *StartupValidationError
	if !errors.As(err, &failure) || len(failure.Findings) != 1 || failure.Findings[0].ReasonCode != "extension_diagnostic_overflow" || strings.Join(calls, ",") != "a,b" {
		t.Fatalf("phase overflow = %v, calls=%v", err, calls)
	}
	if failure.Findings[0].Details["actual"] != 4097 {
		t.Fatal("phase overflow did not stop at the first excess finding")
	}
}

// Terminal publication survives semantic cancellation only until the original
// operation grace. A delayed sink cannot leave a live publication context.
func testTerminalCompletionGrace(t *testing.T) {
	type key struct{}
	for _, canceled := range []bool{false, true} {
		t.Run(fmt.Sprint("cooperative_", canceled), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.WithValue(t.Context(), key{}, "kept"))
			defer cancel()
			if canceled {
				cancel()
			}
			policy := extensiondeadline.InvocationPolicy{Grace: time.Second, Exhausted: func(error) { t.Error("normal completion exhausted") }}
			var retained context.Context
			result, err := extensiondeadline.CompleteWithinGrace(ctx, policy, func(publication context.Context) (string, error) {
				retained = publication
				if publication.Err() != nil || publication.Value(key{}) != "kept" {
					t.Error("completion lost context values or cooperative capability")
				}
				return "committed", nil
			})
			if err != nil || result.Err != nil || result.Value != "committed" || extensiondeadline.GraceExhausted(ctx) || retained.Err() == nil {
				t.Fatal("normal completion classification or scope release", result, err)
			}
		})
	}
	for _, grace := range []time.Duration{0, 20 * time.Millisecond} {
		t.Run("shared_"+grace.String(), func(t *testing.T) {
			ctx, cancel := context.WithCancel(extensiondeadline.WithCancellationTracking(t.Context()))
			defer cancel()
			contexts := make(chan context.Context, 2)
			release, sinkRelease := make(chan struct{}), make(chan struct{})
			fatal := make(chan struct{})
			done := make(chan error, 2)
			policy := extensiondeadline.InvocationPolicy{Grace: grace, Exhausted: func(error) { close(fatal); <-sinkRelease }}
			for range 2 {
				go func() {
					_, err := extensiondeadline.CompleteWithinGrace(extensiondeadline.ReuseCancellationTracking(ctx), policy, func(publication context.Context) (struct{}, error) {
						contexts <- publication
						<-release
						return struct{}{}, nil
					})
					done <- err
				}()
			}
			first, second := <-contexts, <-contexts
			cancel()
			select {
			case <-fatal:
			case <-time.After(time.Second):
				t.Error("no fatal selection")
			}
			for _, publication := range []context.Context{first, second} {
				if !errors.Is(context.Cause(publication), extensiondeadline.ErrGraceExhausted) {
					t.Error("sink signaled before publication revocation", context.Cause(publication))
				}
			}
			select {
			case err := <-done:
				t.Error("active completion escaped", err)
			default:
			}
			close(release)
			close(sinkRelease)
			for range 2 {
				if err := <-done; !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
					t.Error("late result replaced fatal", err)
				}
			}
			_, err := extensiondeadline.CompleteWithinGrace(ctx, policy, func(context.Context) (struct{}, error) {
				t.Error("exhausted operation granted new publication")
				return struct{}{}, nil
			})
			if !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Fatal("new completion reset grace", err)
			}
		})
	}
	t.Run("expired inherited deadline admits no publication", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(extensiondeadline.WithCancellationTracking(t.Context()), time.Now().Add(-time.Second))
		defer cancel()
		policy := extensiondeadline.InvocationPolicy{Grace: time.Millisecond, Exhausted: func(error) {}}
		_, err := extensiondeadline.CompleteWithinGrace(ctx, policy, func(context.Context) (struct{}, error) {
			t.Error("expired grace admitted publication")
			return struct{}{}, nil
		})
		if !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
			t.Fatal(err)
		}
	})
}
