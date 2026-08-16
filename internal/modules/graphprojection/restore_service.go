package graphprojection

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

type RestoreStagedProjection struct {
	SourceRegistrationID string
	CandidateID          string
	Result               CompletedResultV2
}

type RestorePublicationPlan struct {
	RestoreOperationID  uuid.UUID
	TargetGenerationID  uuid.UUID
	ClearedTableIDs     []string
	Projections         []RestoreStagedProjection
	RebuiltViews        []RestoreRebuiltView
	PostconditionSHA256 string
}

type RestorePublicationProof struct {
	RebuiltViews                  []RestoreRebuiltView
	ReconciledNonterminalJobCount int
	ReconciledLeaseCount          int
	PostconditionSHA256           string
}

type RestorePublisher interface {
	ReplaceAll(context.Context, RestorePublicationPlan) (RestorePublicationProof, error)
}

type RestorePublicationError struct {
	Indeterminate bool
	Cause         error
}

func (err *RestorePublicationError) Error() string {
	if err == nil {
		return "graphprojection restore publication failed"
	}
	if err.Indeterminate {
		return "graphprojection restore publication outcome is indeterminate"
	}
	return "graphprojection restore publication failed"
}

func (err *RestorePublicationError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

type RestoreServiceOptions struct {
	Now               func() time.Time
	SupportedBindings []RestoreImplementationBindingRef
	Engine            *EngineV2
}

type RestoreService struct {
	publisher RestorePublisher
	registry  *RestoreSourceRegistry
	now       func() time.Time
	engine    *EngineV2
	bindings  map[string]RestoreImplementationBindingRef
}

var _ RestoreParticipant = (*RestoreService)(nil)

func NewRestoreService(publisher RestorePublisher, registry *RestoreSourceRegistry, options RestoreServiceOptions) (*RestoreService, error) {
	if publisher == nil || registry == nil {
		return nil, NewRestoreError(RestoreErrorInvalidRequest)
	}
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	engine := options.Engine
	if engine == nil {
		engine = NewEngineV2()
	}
	supported := options.SupportedBindings
	if len(supported) == 0 {
		supported = []RestoreImplementationBindingRef{CurrentRestoreImplementationBinding()}
	}
	bindings := make(map[string]RestoreImplementationBindingRef, len(supported))
	for _, binding := range supported {
		if !wellFormedRestoreImplementationBinding(binding) {
			return nil, NewRestoreError(RestoreErrorBindingUnavailable)
		}
		if _, duplicate := bindings[binding.SHA256]; duplicate {
			return nil, NewRestoreError(RestoreErrorBindingUnavailable)
		}
		bindings[binding.SHA256] = binding
	}
	return &RestoreService{publisher: publisher, registry: registry, now: now, engine: engine, bindings: bindings}, nil
}

func (service *RestoreService) Rebuild(ctx context.Context, request RestoreRebuildRequest) (RestoreRebuildResult, error) {
	base := RestoreRebuildResult{
		SchemaID: RestoreRebuildResultSchemaID, RestoreOperationID: request.RestoreOperationID.String(),
		TargetGenerationID: request.TargetGenerationID.String(), Status: RestoreStatusFailed,
		ReadinessOutcome: RestoreReadinessIncomplete, AlgorithmID: RestoreAlgorithmID,
		ImplementationBindingSHA256: request.ImplementationBinding.SHA256,
		SourceRegistrySHA256:        request.SourceRegistry.SHA256,
		ClearedTableIDs:             []string{}, RebuiltViews: []RestoreRebuiltView{}, Warnings: []RestoreSafeMessage{}, Errors: []RestoreSafeMessage{},
	}
	if err := service.validateRequest(ctx, request); err != nil {
		code := restoreCodeOr(err, RestoreErrorInvalidRequest)
		return restoreFailedResult(base, code), err
	}

	staged, rebuilt, err := service.preflight(ctx, request)
	if err != nil {
		code := restoreCodeOr(err, RestoreErrorInvalidCandidate)
		return restoreFailedResult(base, code), err
	}
	postconditionSHA256, err := restorePostconditionDigest(request, rebuilt)
	if err != nil {
		return restoreFailedResult(base, RestoreErrorPostconditionFailed), NewRestoreError(RestoreErrorPostconditionFailed)
	}
	plan := RestorePublicationPlan{
		RestoreOperationID: request.RestoreOperationID, TargetGenerationID: request.TargetGenerationID,
		ClearedTableIDs: RestoreGraphTableIDs(), Projections: staged, RebuiltViews: rebuilt,
		PostconditionSHA256: postconditionSHA256,
	}
	proof, err := service.publisher.ReplaceAll(ctx, plan)
	if err != nil {
		code := RestoreErrorPublicationFailed
		var publicationError *RestorePublicationError
		if errors.As(err, &publicationError) && publicationError.Indeterminate {
			code = RestoreErrorOutcomeIndeterminate
		}
		return restoreFailedResult(base, code), NewRestoreError(code)
	}
	if proof.PostconditionSHA256 != postconditionSHA256 || !equalRestoreRebuiltViews(proof.RebuiltViews, rebuilt) ||
		proof.ReconciledNonterminalJobCount < 0 || proof.ReconciledLeaseCount < 0 {
		return restoreFailedResult(base, RestoreErrorPostconditionFailed), NewRestoreError(RestoreErrorPostconditionFailed)
	}

	result := base
	result.Status = RestoreStatusSucceeded
	result.ReadinessOutcome = RestoreReadinessReady
	result.ClearedTableIDs = RestoreGraphTableIDs()
	result.RebuiltViews = append([]RestoreRebuiltView{}, rebuilt...)
	result.ReconciledNonterminalJobCount = proof.ReconciledNonterminalJobCount
	result.ReconciledLeaseCount = proof.ReconciledLeaseCount
	result.PostconditionSHA256 = &postconditionSHA256
	result.Errors = []RestoreSafeMessage{}
	if err := result.Validate(); err != nil {
		return restoreFailedResult(base, RestoreErrorPostconditionFailed), NewRestoreError(RestoreErrorPostconditionFailed)
	}
	return result, nil
}

func (service *RestoreService) validateRequest(ctx context.Context, request RestoreRebuildRequest) error {
	if service == nil || service.publisher == nil || service.registry == nil || ctx == nil || request.Context == nil ||
		request.RestoreOperationID == uuid.Nil || request.RestoredSourceState == nil || request.BackupSetID == uuid.Nil ||
		request.ConsistencyPointAt.IsZero() || request.TargetGenerationID == uuid.Nil || ctx.Err() != nil || request.Context.Err() != nil {
		return NewRestoreError(RestoreErrorInvalidRequest)
	}
	binding, admitted := service.bindings[request.ImplementationBinding.SHA256]
	if !admitted || !sameRestoreBinding(binding, request.ImplementationBinding) {
		return NewRestoreError(RestoreErrorBindingUnavailable)
	}
	if request.RecoveryStateCatalog.AlgorithmID != RestoreAlgorithmID ||
		!equalStrings(request.RecoveryStateCatalog.GraphTableIDs, restoreGraphTableIDs) {
		return NewRestoreError(RestoreErrorRecoveryCatalogMismatch)
	}
	if request.RecoveryStateCatalog.DigestSHA256 != binding.Binding.RecoveryStateCatalogSHA256 ||
		!equalStrings(binding.Binding.GraphTableIDs, restoreGraphTableIDs) {
		return NewRestoreError(RestoreErrorRecoveryCatalogMismatch)
	}
	if request.SourceRegistry.Registry == nil ||
		request.SourceRegistry.SHA256 != service.registry.DigestSHA256() ||
		request.SourceRegistry.Registry.DigestSHA256() != service.registry.DigestSHA256() ||
		request.SourceRegistry.SHA256 != binding.Binding.SourceRegistrySHA256 {
		return NewRestoreError(RestoreErrorSourceRegistryMismatch)
	}
	return nil
}

func wellFormedRestoreImplementationBinding(ref RestoreImplementationBindingRef) bool {
	if !validSHA256String(ref.SHA256) {
		return false
	}
	body, err := restoreCanonicalJSON(ref.Binding)
	if err != nil || sha256Hex(body) != ref.SHA256 {
		return false
	}
	return ref.SHA256 == contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256 &&
		string(body) == contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON &&
		ref.Binding.SchemaID == RestoreImplementationBindingSchemaID && ref.Binding.AlgorithmID == RestoreAlgorithmID &&
		equalStrings(ref.Binding.GraphTableIDs, restoreGraphTableIDs)
}

func sameRestoreBinding(left, right RestoreImplementationBindingRef) bool {
	if left.SHA256 != right.SHA256 {
		return false
	}
	leftJSON, leftErr := restoreCanonicalJSON(left.Binding)
	rightJSON, rightErr := restoreCanonicalJSON(right.Binding)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func (service *RestoreService) preflight(ctx context.Context, request RestoreRebuildRequest) ([]RestoreStagedProjection, []RestoreRebuiltView, error) {
	registrations := service.registry.Registrations()
	if len(registrations) == 0 || len(registrations) > RestoreMaximumSourceRegistrations {
		return nil, nil, NewRestoreError(RestoreErrorSourceRegistryMismatch)
	}
	staged := make([]RestoreStagedProjection, 0)
	rebuilt := make([]RestoreRebuiltView, 0)
	seenCandidates := make(map[string]struct{})
	seenViews := make(map[string]struct{})
	usage := restoreResourceUsage{sourceRegistrations: len(registrations)}
	for _, registration := range registrations {
		if restoreContextsErr(ctx, request.Context) != nil {
			return nil, nil, NewRestoreError(RestoreErrorSourceEnumeration)
		}
		candidates, err := registration.Enumerate(ctx, request.RestoredSourceState, request.ConsistencyPointAt.UTC())
		if err != nil || !usage.addCandidates(len(candidates)) {
			if err != nil {
				return nil, nil, NewRestoreError(RestoreErrorSourceEnumeration)
			}
			return nil, nil, NewRestoreError(RestoreErrorResourceOverflow)
		}
		for index, candidate := range candidates {
			semanticInput := candidate.SemanticInput
			if strings.TrimSpace(candidate.CandidateID) == "" || strings.TrimSpace(candidate.GraphViewID) == "" || len(semanticInput) == 0 ||
				(index > 0 && candidates[index-1].CandidateID >= candidate.CandidateID) ||
				candidate.ExpectedBinding.GraphViewID != candidate.GraphViewID ||
				candidate.ExpectedBinding.SourceOwnerID != registration.Entry.SourceOwnerID {
				return nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			candidateKey := registration.Entry.SourceRegistrationID + "\x00" + candidate.CandidateID
			if _, duplicate := seenCandidates[candidateKey]; duplicate {
				return nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			if _, duplicate := seenViews[candidate.GraphViewID]; duplicate {
				return nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			seenCandidates[candidateKey] = struct{}{}
			seenViews[candidate.GraphViewID] = struct{}{}
			if !usage.addNormalizedInputBytes(int64(len(semanticInput))) {
				return nil, nil, NewRestoreError(RestoreErrorResourceOverflow)
			}
			if registration.EvaluateValidity != nil {
				validity, validityErr := registration.EvaluateValidity(ctx, candidate, request.ConsistencyPointAt.UTC())
				if validityErr != nil {
					return nil, nil, NewRestoreError(RestoreErrorSourceEnumeration)
				}
				if !validity.Eligible {
					return nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
				}
			}
			projection, err := service.engine.Project(ctx, InvocationContextV2{
				GraphViewID: candidate.GraphViewID, SourceOwnerID: registration.Entry.SourceOwnerID,
				CancellationCheck: func(context.Context) error { return restoreContextsErr(ctx, request.Context) },
			}, semanticInput)
			if err != nil || projection.ResultBindingV2() != candidate.ExpectedBinding {
				return nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			completed, err := projection.CompletedResult()
			if err != nil || !usage.addGraphSize(len(completed.Vertices), len(completed.Edges)) {
				return nil, nil, NewRestoreError(RestoreErrorResourceOverflow)
			}
			completed.PublishedAt = request.ConsistencyPointAt.UTC()
			staged = append(staged, RestoreStagedProjection{SourceRegistrationID: registration.Entry.SourceRegistrationID, CandidateID: candidate.CandidateID, Result: completed})
			binding := completed.Binding
			rebuilt = append(rebuilt, RestoreRebuiltView{
				SourceRegistrationID: registration.Entry.SourceRegistrationID, CandidateID: candidate.CandidateID,
				GraphViewID: binding.GraphViewID, ProjectionResultID: binding.ProjectionResultID,
				SourceSnapshotID: binding.SourceSnapshotID, ProjectionVersion: binding.ProjectionVersion,
				NormalizedConfigurationSHA256: binding.NormalizedConfigurationSHA256,
				NormalizedSourceSHA256:        binding.NormalizedSourceSHA256,
				VertexCount:                   len(completed.Vertices), EdgeCount: len(completed.Edges), CanonicalOutputSHA256: binding.CanonicalOutputSHA256,
			})
		}
	}
	sort.Slice(rebuilt, func(i, j int) bool {
		if rebuilt[i].SourceRegistrationID != rebuilt[j].SourceRegistrationID {
			return rebuilt[i].SourceRegistrationID < rebuilt[j].SourceRegistrationID
		}
		return rebuilt[i].CandidateID < rebuilt[j].CandidateID
	})
	return staged, rebuilt, nil
}

type restoreResourceUsage struct {
	sourceRegistrations  int
	candidates           int
	normalizedInputBytes int64
	vertices             int
	edges                int
}

func (usage *restoreResourceUsage) valid() bool {
	return usage != nil && usage.sourceRegistrations >= 0 && usage.sourceRegistrations <= RestoreMaximumSourceRegistrations &&
		usage.candidates >= 0 && usage.candidates <= RestoreMaximumCandidates && usage.normalizedInputBytes >= 0 &&
		usage.normalizedInputBytes <= RestoreMaximumNormalizedInputBytes && usage.vertices >= 0 && usage.vertices <= RestoreMaximumVertices &&
		usage.edges >= 0 && usage.edges <= RestoreMaximumEdges
}

func (usage *restoreResourceUsage) addCandidates(count int) bool {
	if usage == nil || count < 0 || !usage.valid() || count > RestoreMaximumCandidates-usage.candidates {
		return false
	}
	usage.candidates += count
	return true
}
func (usage *restoreResourceUsage) addNormalizedInputBytes(count int64) bool {
	if usage == nil || count < 0 || !usage.valid() || count > RestoreMaximumNormalizedInputBytes-usage.normalizedInputBytes {
		return false
	}
	usage.normalizedInputBytes += count
	return true
}
func (usage *restoreResourceUsage) addGraphSize(vertices, edges int) bool {
	if usage == nil || vertices < 0 || edges < 0 || !usage.valid() || vertices > RestoreMaximumVertices-usage.vertices || edges > RestoreMaximumEdges-usage.edges {
		return false
	}
	usage.vertices += vertices
	usage.edges += edges
	return true
}

func restoreContextsErr(contexts ...context.Context) error {
	for _, ctx := range contexts {
		if ctx == nil {
			return context.Canceled
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func restorePostconditionDigest(request RestoreRebuildRequest, rebuilt []RestoreRebuiltView) (string, error) {
	body, err := restoreCanonicalJSON(map[string]any{
		"algorithm_id": RestoreAlgorithmID, "cleared_table_ids": restoreGraphTableIDs, "rebuilt_views": rebuilt,
		"restore_operation_id": request.RestoreOperationID.String(), "target_generation_id": request.TargetGenerationID.String(),
	})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(body)
	return hex.EncodeToString(digest[:]), nil
}

func restoreFailedResult(base RestoreRebuildResult, code RestoreErrorCode) RestoreRebuildResult {
	base.Status = RestoreStatusFailed
	base.ReadinessOutcome = RestoreReadinessIncomplete
	base.ClearedTableIDs = []string{}
	base.RebuiltViews = []RestoreRebuiltView{}
	base.ReconciledNonterminalJobCount = 0
	base.ReconciledLeaseCount = 0
	base.PostconditionSHA256 = nil
	base.Errors = []RestoreSafeMessage{{Code: code}}
	return base
}

func restoreCodeOr(err error, fallback RestoreErrorCode) RestoreErrorCode {
	if code, ok := RestoreErrorCodeOf(err); ok {
		return code
	}
	return fallback
}

func equalRestoreRebuiltViews(left, right []RestoreRebuiltView) bool {
	leftJSON, leftErr := restoreCanonicalJSON(left)
	rightJSON, rightErr := restoreCanonicalJSON(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func restoreCanonicalJSON(value any) ([]byte, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var generic any
	if err := decoder.Decode(&generic); err != nil {
		return nil, err
	}
	return canonicalJSON(generic)
}
