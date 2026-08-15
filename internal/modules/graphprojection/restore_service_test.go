package graphprojection

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type restoreTestSourceState struct{}

func (restoreTestSourceState) GraphProjectionRestoreSourceState() {}

type recordingRestorePublisher struct {
	plans []RestorePublicationPlan
	err   error
}

func (publisher *recordingRestorePublisher) ReplaceAll(_ context.Context, plan RestorePublicationPlan) (RestorePublicationProof, error) {
	publisher.plans = append(publisher.plans, plan)
	if publisher.err != nil {
		return RestorePublicationProof{}, publisher.err
	}
	return RestorePublicationProof{RebuiltViews: plan.RebuiltViews, PostconditionSHA256: plan.PostconditionSHA256}, nil
}

func TestGraphRestoreAcceptanceGPRA01EmptyRegistryClearOnly(t *testing.T) {
	publisher := &recordingRestorePublisher{}
	service, err := NewRestoreService(publisher, CurrentRestoreSourceRegistry(), RestoreServiceOptions{})
	if err != nil {
		t.Fatalf("construct restore service: %v", err)
	}
	request := currentRestoreTestRequest(context.Background())
	result, err := service.Rebuild(request.Context, request)
	if err != nil {
		t.Fatalf("clear-only Graph restore: %v result=%#v plans=%#v", err, result, publisher.plans)
	}
	if !result.ReadinessSatisfied() || len(result.RebuiltViews) != 0 || len(result.SkippedCandidates) != 0 || len(publisher.plans) != 1 {
		t.Fatalf("clear-only result or publication mismatch: result=%#v plans=%d", result, len(publisher.plans))
	}
	if len(publisher.plans[0].Projections) != 0 || !equalStrings(publisher.plans[0].ClearedTableIDs, RestoreGraphTableIDs()) {
		t.Fatalf("clear-only plan mismatch: %#v", publisher.plans[0])
	}
}

func TestGraphRestoreAcceptanceGPRA02Through05CandidateAndSkipSemantics(t *testing.T) {
	consistencyPoint := fixedTime()
	validInput := minimalInput(t, "restore_active")
	validInput["requested_at"] = FormatLifecycleTimestamp(consistencyPoint)
	validInput["requested_by"] = restoreRequestedBy
	validGraphViewID := validInput["graph_view_id"].(string)
	introducedAfter := consistencyPoint.Add(time.Hour)
	retiredAt := consistencyPoint.Add(-time.Hour)
	registrations := []RestoreSourceRegistration{
		{
			Entry: restoreTestRegistryEntry("active", consistencyPoint.Add(-time.Hour), nil),
			Enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
				return []RestoreCandidate{{CandidateID: "candidate_active", GraphViewID: validGraphViewID, NormalizedInput: mustJSON(t, validInput)}}, nil
			},
			EvaluateValidity: func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error) {
				return RestoreCandidateValidity{Eligible: true}, nil
			},
		},
		{
			Entry: restoreTestRegistryEntry("expired", consistencyPoint.Add(-time.Hour), nil),
			Enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
				return []RestoreCandidate{{CandidateID: "candidate_expired", GraphViewID: "gv_expired", NormalizedInput: []byte(`{}`)}}, nil
			},
			EvaluateValidity: func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error) {
				return RestoreCandidateValidity{SkipReason: RestoreSkipDeclarationExpired}, nil
			},
		},
		{
			Entry: restoreTestRegistryEntry("future", introducedAfter, nil),
			Enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
				return []RestoreCandidate{{CandidateID: "candidate_future", GraphViewID: "gv_future", NormalizedInput: []byte(`{}`)}}, nil
			},
			EvaluateValidity: func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error) {
				t.Fatal("future registration validity evaluator must not run")
				return RestoreCandidateValidity{}, nil
			},
		},
		{
			Entry: restoreTestRegistryEntry("retired", consistencyPoint.Add(-2*time.Hour), &retiredAt),
			Enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
				return []RestoreCandidate{{CandidateID: "candidate_retired", GraphViewID: "gv_retired", NormalizedInput: []byte(`{}`)}}, nil
			},
			EvaluateValidity: func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error) {
				t.Fatal("retired registration validity evaluator must not run")
				return RestoreCandidateValidity{}, nil
			},
		},
	}
	registry, err := NewRestoreSourceRegistry(registrations...)
	if err != nil {
		t.Fatalf("construct test restore registry: %v", err)
	}
	binding := restoreTestBinding(t, registry)
	publisher := &recordingRestorePublisher{}
	service, err := NewRestoreService(publisher, registry, RestoreServiceOptions{
		Now:               func() time.Time { return consistencyPoint.Add(2 * time.Hour) },
		NewNonce:          func() (string, error) { return "restore-nonce-1", nil },
		SupportedBindings: []RestoreImplementationBindingRef{binding},
	})
	if err != nil {
		t.Fatalf("construct restore service: %v", err)
	}
	request := restoreTestRequest(context.Background(), registry, binding, consistencyPoint)
	result, err := service.Rebuild(request.Context, request)
	if err != nil {
		t.Fatalf("restore candidates: %v", err)
	}
	if len(result.RebuiltViews) != 1 || result.RebuiltViews[0].GraphViewID != validGraphViewID {
		t.Fatalf("rebuilt views mismatch: %#v", result.RebuiltViews)
	}
	wantReasons := []RestoreSkipReason{RestoreSkipDeclarationExpired, RestoreSkipRegistrationNotYetActive, RestoreSkipRegistrationRetired}
	if len(result.SkippedCandidates) != len(wantReasons) {
		t.Fatalf("skipped candidates got %#v", result.SkippedCandidates)
	}
	for index, reason := range wantReasons {
		if result.SkippedCandidates[index].Reason != reason {
			t.Fatalf("skip[%d] got %q want %q", index, result.SkippedCandidates[index].Reason, reason)
		}
	}
	if publisher.plans[0].Projections[0].Run.Request.RequestedBy != restoreRequestedBy ||
		publisher.plans[0].Projections[0].Run.AcceptedAt.Equal(consistencyPoint) {
		t.Fatalf("caller/server timestamp ownership mismatch: %#v", publisher.plans[0].Projections[0].Run)
	}
}

func TestGraphRestoreAcceptanceGPRA06Through11PreflightFailsBeforeMutation(t *testing.T) {
	base := currentRestoreTestRequest(context.Background())
	tests := []struct {
		name   string
		mutate func(*RestoreRebuildRequest)
		code   RestoreErrorCode
	}{
		{name: "registry digest", mutate: func(request *RestoreRebuildRequest) { request.SourceRegistry.SHA256 = strings.Repeat("a", 64) }, code: RestoreErrorSourceRegistryMismatch},
		{name: "binding unavailable", mutate: func(request *RestoreRebuildRequest) { request.ImplementationBinding.SHA256 = strings.Repeat("b", 64) }, code: RestoreErrorBindingUnavailable},
		{name: "catalog digest", mutate: func(request *RestoreRebuildRequest) {
			request.RecoveryStateCatalog.DigestSHA256 = strings.Repeat("c", 64)
		}, code: RestoreErrorRecoveryCatalogMismatch},
		{name: "algorithm", mutate: func(request *RestoreRebuildRequest) { request.RecoveryStateCatalog.AlgorithmID = "unknown" }, code: RestoreErrorRecoveryCatalogMismatch},
		{name: "tables", mutate: func(request *RestoreRebuildRequest) {
			request.RecoveryStateCatalog.GraphTableIDs = request.RecoveryStateCatalog.GraphTableIDs[:4]
		}, code: RestoreErrorRecoveryCatalogMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publisher := &recordingRestorePublisher{}
			service, err := NewRestoreService(publisher, CurrentRestoreSourceRegistry(), RestoreServiceOptions{})
			if err != nil {
				t.Fatalf("construct restore service: %v", err)
			}
			request := base
			request.RecoveryStateCatalog.GraphTableIDs = append([]string(nil), base.RecoveryStateCatalog.GraphTableIDs...)
			test.mutate(&request)
			result, err := service.Rebuild(request.Context, request)
			if code, ok := RestoreErrorCodeOf(err); !ok || code != test.code {
				t.Fatalf("error code got %q,%v want %q", code, ok, test.code)
			}
			if result.SchemaID != "" || len(publisher.plans) != 0 {
				t.Fatalf("pre-admission failure returned result or mutated: result=%#v plans=%d", result, len(publisher.plans))
			}
		})
	}
}

func TestGraphRestoreAcceptanceGPRA06And10SourceFailuresBeforeMutation(t *testing.T) {
	consistencyPoint := fixedTime()
	tests := []struct {
		name      string
		enumerate RestoreCandidateEnumerator
		code      RestoreErrorCode
	}{
		{
			name: "required source input missing",
			enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
				return nil, errors.New("SECRET unavailable source capability")
			},
			code: RestoreErrorSourceEnumeration,
		},
		{
			name: "normalized projection input invalid",
			enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
				return []RestoreCandidate{{CandidateID: "candidate_invalid", GraphViewID: "gv_invalid", NormalizedInput: []byte(`{}`)}}, nil
			},
			code: RestoreErrorInvalidCandidate,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registration := RestoreSourceRegistration{
				Entry:     restoreTestRegistryEntry("active", consistencyPoint.Add(-time.Hour), nil),
				Enumerate: test.enumerate,
				EvaluateValidity: func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error) {
					return RestoreCandidateValidity{Eligible: true}, nil
				},
			}
			registry, err := NewRestoreSourceRegistry(registration)
			if err != nil {
				t.Fatalf("construct source registry: %v", err)
			}
			binding := restoreTestBinding(t, registry)
			publisher := &recordingRestorePublisher{}
			service, err := NewRestoreService(publisher, registry, RestoreServiceOptions{SupportedBindings: []RestoreImplementationBindingRef{binding}})
			if err != nil {
				t.Fatalf("construct restore service: %v", err)
			}
			request := restoreTestRequest(context.Background(), registry, binding, consistencyPoint)
			result, err := service.Rebuild(request.Context, request)
			if code, ok := RestoreErrorCodeOf(err); !ok || code != test.code {
				t.Fatalf("error code got %q,%v want %q", code, ok, test.code)
			}
			if len(publisher.plans) != 0 || result.ReadinessOutcome != RestoreReadinessIncomplete ||
				len(result.Errors) != 1 || result.Errors[0].Code != test.code {
				t.Fatalf("source failure mutated or returned unsafe result: result=%#v plans=%d", result, len(publisher.plans))
			}
			body, marshalErr := json.Marshal(result)
			if marshalErr != nil || strings.Contains(string(body), "SECRET") || strings.Contains(err.Error(), "SECRET") {
				t.Fatalf("source failure leaked detail: result=%s error=%v", body, err)
			}
		})
	}
}

func TestGraphRestoreAcceptanceGPRA11OperationWideResourceLimits(t *testing.T) {
	tooManyRegistrations := make([]RestoreSourceRegistration, RestoreMaximumSourceRegistrations+1)
	if _, err := NewRestoreSourceRegistry(tooManyRegistrations...); restoreCodeOr(err, "") != RestoreErrorResourceOverflow {
		t.Fatalf("source registration limit got %v", err)
	}

	usage := restoreResourceUsage{sourceRegistrations: RestoreMaximumSourceRegistrations}
	if !usage.addCandidates(RestoreMaximumCandidates) || usage.addCandidates(1) {
		t.Fatal("candidate operation-wide limit was not enforced at the exact boundary")
	}
	usage = restoreResourceUsage{}
	if !usage.addNormalizedInputBytes(RestoreMaximumNormalizedInputBytes) || usage.addNormalizedInputBytes(1) {
		t.Fatal("normalized-input operation-wide limit was not enforced at the exact boundary")
	}
	usage = restoreResourceUsage{}
	if !usage.addGraphSize(RestoreMaximumVertices, RestoreMaximumEdges) || usage.addGraphSize(1, 0) || usage.addGraphSize(0, 1) {
		t.Fatal("Graph-size operation-wide limits were not enforced at the exact boundaries")
	}

	consistencyPoint := fixedTime()
	candidates := make([]RestoreCandidate, RestoreMaximumCandidates+1)
	registration := RestoreSourceRegistration{
		Entry: restoreTestRegistryEntry("active", consistencyPoint.Add(-time.Hour), nil),
		Enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
			return candidates, nil
		},
		EvaluateValidity: func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error) {
			return RestoreCandidateValidity{Eligible: true}, nil
		},
	}
	registry, err := NewRestoreSourceRegistry(registration)
	if err != nil {
		t.Fatalf("construct source registry: %v", err)
	}
	binding := restoreTestBinding(t, registry)
	publisher := &recordingRestorePublisher{}
	service, err := NewRestoreService(publisher, registry, RestoreServiceOptions{SupportedBindings: []RestoreImplementationBindingRef{binding}})
	if err != nil {
		t.Fatalf("construct restore service: %v", err)
	}
	request := restoreTestRequest(context.Background(), registry, binding, consistencyPoint)
	result, err := service.Rebuild(request.Context, request)
	if code, ok := RestoreErrorCodeOf(err); !ok || code != RestoreErrorResourceOverflow {
		t.Fatalf("candidate overflow got %q,%v (%v)", code, ok, err)
	}
	if len(publisher.plans) != 0 || result.ReadinessOutcome != RestoreReadinessIncomplete {
		t.Fatalf("resource overflow crossed publication boundary: result=%#v plans=%d", result, len(publisher.plans))
	}
}

func TestGraphRestoreAcceptanceGPRA15RetryPreservesDeterministicOutput(t *testing.T) {
	consistencyPoint := fixedTime()
	input := minimalInput(t, "restore_retry")
	input["requested_at"] = FormatLifecycleTimestamp(consistencyPoint)
	input["requested_by"] = restoreRequestedBy
	registration := RestoreSourceRegistration{
		Entry: restoreTestRegistryEntry("active", consistencyPoint.Add(-time.Hour), nil),
		Enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
			return []RestoreCandidate{{CandidateID: "candidate_retry", GraphViewID: input["graph_view_id"].(string), NormalizedInput: mustJSON(t, input)}}, nil
		},
		EvaluateValidity: func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error) {
			return RestoreCandidateValidity{Eligible: true}, nil
		},
	}
	registry, err := NewRestoreSourceRegistry(registration)
	if err != nil {
		t.Fatalf("construct source registry: %v", err)
	}
	binding := restoreTestBinding(t, registry)
	nonces := []string{"restore-retry-nonce-1", "restore-retry-nonce-2"}
	publisher := &recordingRestorePublisher{}
	service, err := NewRestoreService(publisher, registry, RestoreServiceOptions{
		Now: func() time.Time { return consistencyPoint.Add(time.Hour) },
		NewNonce: func() (string, error) {
			nonce := nonces[0]
			nonces = nonces[1:]
			return nonce, nil
		},
		SupportedBindings: []RestoreImplementationBindingRef{binding},
	})
	if err != nil {
		t.Fatalf("construct restore service: %v", err)
	}
	request := restoreTestRequest(context.Background(), registry, binding, consistencyPoint)
	first, err := service.Rebuild(request.Context, request)
	if err != nil {
		t.Fatalf("first restore: %v", err)
	}
	second, err := service.Rebuild(request.Context, request)
	if err != nil {
		t.Fatalf("retry restore: %v", err)
	}
	firstView := first.RebuiltViews[0]
	secondView := second.RebuiltViews[0]
	if firstView.ProjectionRunID == secondView.ProjectionRunID {
		t.Fatal("retry must use a fresh server-owned run nonce")
	}
	firstView.ProjectionRunID = ""
	secondView.ProjectionRunID = ""
	if firstView != secondView {
		t.Fatalf("retry changed deterministic identity or output: first=%#v second=%#v", firstView, secondView)
	}
}

func TestGraphRestoreAcceptanceGPRA13And17IndeterminateFailureIsClosed(t *testing.T) {
	secret := errors.New("database SECRET-SQL-object-key")
	publisher := &recordingRestorePublisher{err: &RestorePublicationError{Indeterminate: true, Cause: secret}}
	service, err := NewRestoreService(publisher, CurrentRestoreSourceRegistry(), RestoreServiceOptions{})
	if err != nil {
		t.Fatalf("construct restore service: %v", err)
	}
	request := currentRestoreTestRequest(context.Background())
	result, err := service.Rebuild(request.Context, request)
	if code, ok := RestoreErrorCodeOf(err); !ok || code != RestoreErrorOutcomeIndeterminate {
		t.Fatalf("indeterminate error got %q,%v (%v)", code, ok, err)
	}
	if result.ReadinessOutcome != RestoreReadinessIncomplete || len(result.Errors) != 1 || result.Errors[0].Code != RestoreErrorOutcomeIndeterminate {
		t.Fatalf("indeterminate result mismatch: %#v", result)
	}
	encoded, marshalErr := json.Marshal(result)
	if marshalErr != nil || strings.Contains(string(encoded), "SECRET") || strings.Contains(err.Error(), "SECRET") {
		t.Fatalf("safe failure leaked cause: result=%s error=%v", encoded, err)
	}
}

func restoreTestRegistryEntry(id string, introducedAt time.Time, retiredAt *time.Time) RestoreSourceRegistryEntry {
	status := "active"
	if retiredAt != nil {
		status = "retired"
	}
	return RestoreSourceRegistryEntry{
		SourceRegistrationID: id, SourceOwnerID: "module.test." + id,
		EnumeratorBindingID: "test.enumerator." + id, ValidityBindingID: "test.validity." + id,
		ProjectionInputContractID: ProjectionSchemaID, ImplementationBindingID: "test.implementation." + id,
		Status: status, IntroducedAt: introducedAt, RetiredAt: retiredAt,
	}
}

func restoreTestBinding(t *testing.T, registry *RestoreSourceRegistry) RestoreImplementationBindingRef {
	t.Helper()
	ref := CurrentRestoreImplementationBinding()
	ref.Binding.BindingID = "graphprojection.restore_rebuild.test.v1"
	ref.Binding.SourceRegistrySHA256 = registry.DigestSHA256()
	body, err := restoreCanonicalJSON(ref.Binding)
	if err != nil {
		t.Fatalf("canonicalize test binding: %v", err)
	}
	digest := sha256.Sum256(body)
	ref.SHA256 = hex.EncodeToString(digest[:])
	ref.Legacy = false
	return ref
}

func currentRestoreTestRequest(ctx context.Context) RestoreRebuildRequest {
	registry := CurrentRestoreSourceRegistry()
	return restoreTestRequest(ctx, registry, CurrentRestoreImplementationBinding(), fixedTime())
}

func restoreTestRequest(ctx context.Context, registry *RestoreSourceRegistry, binding RestoreImplementationBindingRef, consistencyPoint time.Time) RestoreRebuildRequest {
	return RestoreRebuildRequest{
		Context: ctx, RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000007001"),
		RestoredSourceState: restoreTestSourceState{}, BackupSetID: uuid.MustParse("00000000-0000-0000-0000-000000007002"),
		ConsistencyPointAt: consistencyPoint, TargetGenerationID: uuid.MustParse("00000000-0000-0000-0000-000000007003"),
		RecoveryStateCatalog: RestoreRecoveryCatalogRef{
			DigestSHA256: binding.Binding.RecoveryStateCatalogSHA256, AlgorithmID: RestoreAlgorithmID,
			GraphTableIDs: RestoreGraphTableIDs(),
		},
		SourceRegistry:        RestoreSourceRegistryRef{Registry: registry, SHA256: registry.DigestSHA256()},
		ImplementationBinding: binding,
	}
}
