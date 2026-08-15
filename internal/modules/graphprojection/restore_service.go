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
)

const restoreRequestedBy = "recovery_restore"

// RestoreStagedProjection is a fully admitted and derived projection that may
// be published. It contains no source capability and is created before the
// publication transaction begins.
type RestoreStagedProjection struct {
	SourceRegistrationID string
	CandidateID          string
	Run                  ProjectionRun
}

// RestorePublicationPlan is the complete, immutable input to the narrow Graph
// restore writer. Implementations replace all five derived tables atomically.
type RestorePublicationPlan struct {
	RestoreOperationID  uuid.UUID
	TargetGenerationID  uuid.UUID
	ClearedTableIDs     []string
	Projections         []RestoreStagedProjection
	RebuiltViews        []RestoreRebuiltView
	SkippedCandidates   []RestoreSkippedCandidate
	PostconditionSHA256 string
}

// RestorePublicationProof is returned only after commit and a committed-state
// reread prove that the publication plan is present.
type RestorePublicationProof struct {
	RebuiltViews        []RestoreRebuiltView
	PostconditionSHA256 string
}

type RestorePublisher interface {
	ReplaceAll(context.Context, RestorePublicationPlan) (RestorePublicationProof, error)
}

// RestorePublicationError distinguishes a known rollback from a commit whose
// outcome cannot be proved. Cause is intentionally not exposed by Graph's
// participant error or result.
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
	NewNonce          func() (string, error)
	SupportedBindings []RestoreImplementationBindingRef
}

type RestoreService struct {
	publisher RestorePublisher
	registry  *RestoreSourceRegistry
	now       func() time.Time
	newNonce  func() (string, error)
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
	newNonce := options.NewNonce
	if newNonce == nil {
		newNonce = func() (string, error) {
			value, err := uuid.NewRandom()
			return value.String(), err
		}
	}
	supported := options.SupportedBindings
	if len(supported) == 0 {
		supported = []RestoreImplementationBindingRef{
			CurrentRestoreImplementationBinding(),
			LegacyEmptyRegistryRestoreImplementationBinding(),
		}
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
	return &RestoreService{publisher: publisher, registry: registry, now: now, newNonce: newNonce, bindings: bindings}, nil
}

func (service *RestoreService) Rebuild(ctx context.Context, request RestoreRebuildRequest) (RestoreRebuildResult, error) {
	if err := service.validateRequest(ctx, request); err != nil {
		return RestoreRebuildResult{}, err
	}

	base := RestoreRebuildResult{
		SchemaID:                    RestoreRebuildResultSchemaID,
		RestoreOperationID:          request.RestoreOperationID.String(),
		TargetGenerationID:          request.TargetGenerationID.String(),
		Status:                      RestoreStatusFailed,
		ReadinessOutcome:            RestoreReadinessIncomplete,
		AlgorithmID:                 RestoreAlgorithmID,
		ImplementationBindingSHA256: request.ImplementationBinding.SHA256,
		SourceRegistrySHA256:        request.SourceRegistry.SHA256,
		ClearedTableIDs:             []string{},
		RebuiltViews:                []RestoreRebuiltView{},
		SkippedCandidates:           []RestoreSkippedCandidate{},
		Warnings:                    []RestoreSafeMessage{},
		Errors:                      []RestoreSafeMessage{},
	}

	staged, rebuilt, skipped, err := service.preflight(ctx, request)
	if err != nil {
		return restoreFailedResult(base, restoreCodeOr(err, RestoreErrorInvalidCandidate)), err
	}
	postconditionSHA256, err := restorePostconditionDigest(request, rebuilt, skipped)
	if err != nil {
		classified := NewRestoreError(RestoreErrorPostconditionFailed)
		return restoreFailedResult(base, RestoreErrorPostconditionFailed), classified
	}
	plan := RestorePublicationPlan{
		RestoreOperationID:  request.RestoreOperationID,
		TargetGenerationID:  request.TargetGenerationID,
		ClearedTableIDs:     RestoreGraphTableIDs(),
		Projections:         staged,
		RebuiltViews:        rebuilt,
		SkippedCandidates:   skipped,
		PostconditionSHA256: postconditionSHA256,
	}
	proof, err := service.publisher.ReplaceAll(ctx, plan)
	if err != nil {
		code := RestoreErrorPublicationFailed
		var publicationError *RestorePublicationError
		if errors.As(err, &publicationError) && publicationError.Indeterminate {
			code = RestoreErrorOutcomeIndeterminate
		}
		classified := NewRestoreError(code)
		return restoreFailedResult(base, code), classified
	}
	if proof.PostconditionSHA256 != postconditionSHA256 || !equalRestoreRebuiltViews(proof.RebuiltViews, rebuilt) {
		classified := NewRestoreError(RestoreErrorPostconditionFailed)
		return restoreFailedResult(base, RestoreErrorPostconditionFailed), classified
	}

	result := base
	result.Status = RestoreStatusSucceeded
	result.ReadinessOutcome = RestoreReadinessReady
	result.ClearedTableIDs = RestoreGraphTableIDs()
	result.RebuiltViews = append([]RestoreRebuiltView{}, rebuilt...)
	result.SkippedCandidates = append([]RestoreSkippedCandidate{}, skipped...)
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
		request.ConsistencyPointAt.IsZero() || request.TargetGenerationID == uuid.Nil {
		return NewRestoreError(RestoreErrorInvalidRequest)
	}
	if err := ctx.Err(); err != nil {
		return NewRestoreError(RestoreErrorInvalidRequest)
	}
	if err := request.Context.Err(); err != nil {
		return NewRestoreError(RestoreErrorInvalidRequest)
	}
	binding := request.ImplementationBinding
	if !service.supportsRestoreImplementationBinding(binding) {
		return NewRestoreError(RestoreErrorBindingUnavailable)
	}
	if request.RecoveryStateCatalog.AlgorithmID != RestoreAlgorithmID ||
		request.RecoveryStateCatalog.DigestSHA256 != binding.Binding.RecoveryStateCatalogSHA256 ||
		!equalStrings(request.RecoveryStateCatalog.GraphTableIDs, restoreGraphTableIDs) ||
		!equalStrings(binding.Binding.GraphTableIDs, restoreGraphTableIDs) {
		return NewRestoreError(RestoreErrorRecoveryCatalogMismatch)
	}
	if request.SourceRegistry.Registry == nil ||
		request.SourceRegistry.SHA256 != service.registry.DigestSHA256() ||
		request.SourceRegistry.Registry.DigestSHA256() != service.registry.DigestSHA256() ||
		request.SourceRegistry.SHA256 != binding.Binding.SourceRegistrySHA256 {
		return NewRestoreError(RestoreErrorSourceRegistryMismatch)
	}
	if binding.Legacy && len(service.registry.Registrations()) != 0 {
		return NewRestoreError(RestoreErrorBindingUnavailable)
	}
	return nil
}

func wellFormedRestoreImplementationBinding(ref RestoreImplementationBindingRef) bool {
	if !validSHA256String(ref.SHA256) {
		return false
	}
	binding := ref.Binding
	if binding.SchemaID != RestoreImplementationBindingSchemaID || binding.AlgorithmID != RestoreAlgorithmID ||
		strings.TrimSpace(binding.BindingID) == "" || strings.TrimSpace(binding.GraphProjectionContractID) == "" ||
		!validSHA256String(binding.RecoveryStateCatalogSHA256) || !validSHA256String(binding.SourceRegistrySHA256) ||
		!equalStrings(binding.GraphTableIDs, restoreGraphTableIDs) || len(binding.GraphEngineAlgorithmIDs) == 0 ||
		len(binding.GraphEngineAlgorithmIDs) != len(binding.GraphEngineAlgorithmDigests) || binding.DatabaseSchemaHead < 22 ||
		strings.TrimSpace(binding.DatabaseSchemaLineage) == "" || !validSHA256String(binding.PackagedSubjectSHA256) ||
		!validSHA256String(binding.BuildProvenanceSHA256) {
		return false
	}
	for index, algorithmID := range binding.GraphEngineAlgorithmIDs {
		if strings.TrimSpace(algorithmID) == "" || !validSHA256String(binding.GraphEngineAlgorithmDigests[index]) ||
			(index > 0 && binding.GraphEngineAlgorithmIDs[index-1] >= algorithmID) {
			return false
		}
	}
	actualJSON, err := restoreCanonicalJSON(ref.Binding)
	if err != nil {
		return false
	}
	digest := sha256.Sum256(actualJSON)
	return hex.EncodeToString(digest[:]) == ref.SHA256
}

func (service *RestoreService) supportsRestoreImplementationBinding(ref RestoreImplementationBindingRef) bool {
	if service == nil || !wellFormedRestoreImplementationBinding(ref) {
		return false
	}
	want, ok := service.bindings[ref.SHA256]
	if !ok || want.Legacy != ref.Legacy {
		return false
	}
	actualJSON, actualErr := restoreCanonicalJSON(ref.Binding)
	wantJSON, wantErr := restoreCanonicalJSON(want.Binding)
	return actualErr == nil && wantErr == nil && string(actualJSON) == string(wantJSON)
}

func (service *RestoreService) preflight(
	ctx context.Context,
	request RestoreRebuildRequest,
) ([]RestoreStagedProjection, []RestoreRebuiltView, []RestoreSkippedCandidate, error) {
	registrations := service.registry.Registrations()
	if len(registrations) > RestoreMaximumSourceRegistrations {
		return nil, nil, nil, NewRestoreError(RestoreErrorResourceOverflow)
	}
	staged := make([]RestoreStagedProjection, 0)
	rebuilt := make([]RestoreRebuiltView, 0)
	skipped := make([]RestoreSkippedCandidate, 0)
	seenCandidates := make(map[string]struct{})
	seenViews := make(map[string]struct{})
	usage := restoreResourceUsage{sourceRegistrations: len(registrations)}
	consistencyPoint := request.ConsistencyPointAt.UTC()
	lifecycleAt := service.now().UTC()

	for _, registration := range registrations {
		if err := restoreContextsErr(ctx, request.Context); err != nil {
			return nil, nil, nil, NewRestoreError(RestoreErrorSourceEnumeration)
		}
		candidates, err := registration.Enumerate(ctx, request.RestoredSourceState, consistencyPoint)
		if err != nil {
			return nil, nil, nil, NewRestoreError(RestoreErrorSourceEnumeration)
		}
		if !usage.addCandidates(len(candidates)) {
			return nil, nil, nil, NewRestoreError(RestoreErrorResourceOverflow)
		}
		for index, candidate := range candidates {
			if strings.TrimSpace(candidate.CandidateID) == "" || strings.TrimSpace(candidate.GraphViewID) == "" ||
				len(candidate.NormalizedInput) == 0 || (index > 0 && candidates[index-1].CandidateID >= candidate.CandidateID) {
				return nil, nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			candidateKey := registration.Entry.SourceRegistrationID + "\x00" + candidate.CandidateID
			if _, duplicate := seenCandidates[candidateKey]; duplicate {
				return nil, nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			seenCandidates[candidateKey] = struct{}{}
			if _, duplicate := seenViews[candidate.GraphViewID]; duplicate {
				return nil, nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			seenViews[candidate.GraphViewID] = struct{}{}
			if !usage.addNormalizedInputBytes(int64(len(candidate.NormalizedInput))) {
				return nil, nil, nil, NewRestoreError(RestoreErrorResourceOverflow)
			}

			if reason, omit := registrationSkipReason(registration.Entry, consistencyPoint); omit {
				skipped = append(skipped, RestoreSkippedCandidate{SourceRegistrationID: registration.Entry.SourceRegistrationID, CandidateID: candidate.CandidateID, Reason: reason})
				continue
			}
			validity, err := registration.EvaluateValidity(ctx, candidate, consistencyPoint)
			if err != nil {
				return nil, nil, nil, NewRestoreError(RestoreErrorSourceEnumeration)
			}
			if !validity.Eligible {
				if validity.SkipReason != RestoreSkipDeclarationExpired {
					return nil, nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
				}
				skipped = append(skipped, RestoreSkippedCandidate{SourceRegistrationID: registration.Entry.SourceRegistrationID, CandidateID: candidate.CandidateID, Reason: validity.SkipReason})
				continue
			}
			if validity.SkipReason != "" || (candidate.DeclarationExpiresAt != nil && !consistencyPoint.Before(candidate.DeclarationExpiresAt.UTC())) {
				skipped = append(skipped, RestoreSkippedCandidate{SourceRegistrationID: registration.Entry.SourceRegistrationID, CandidateID: candidate.CandidateID, Reason: RestoreSkipDeclarationExpired})
				continue
			}
			nonce, err := service.newNonce()
			if err != nil || strings.TrimSpace(nonce) == "" {
				return nil, nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			run, err := AdmitRetainedProjection(candidate.NormalizedInput, nonce, lifecycleAt)
			if err != nil || run.GraphViewID != candidate.GraphViewID ||
				run.Request.RequestedAt != FormatLifecycleTimestamp(consistencyPoint) || run.Request.RequestedBy != restoreRequestedBy {
				return nil, nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			run.State = RunStateComputing
			startedAt := lifecycleAt
			run.StartedAt = &startedAt
			run, err = ProjectAdmittedRetainedProjection(ctx, run, lifecycleAt, nil)
			if err != nil || run.State != RunStateAvailable || run.GraphView == nil || run.ValidationSummary.Status != "passed" {
				return nil, nil, nil, NewRestoreError(RestoreErrorInvalidCandidate)
			}
			if !usage.addGraphSize(len(run.GraphView.Vertices), len(run.GraphView.Edges)) {
				return nil, nil, nil, NewRestoreError(RestoreErrorResourceOverflow)
			}
			staged = append(staged, RestoreStagedProjection{SourceRegistrationID: registration.Entry.SourceRegistrationID, CandidateID: candidate.CandidateID, Run: run})
			rebuilt = append(rebuilt, RestoreRebuiltView{
				SourceRegistrationID:          registration.Entry.SourceRegistrationID,
				CandidateID:                   candidate.CandidateID,
				GraphViewID:                   run.GraphViewID,
				ProjectionRunID:               run.ProjectionRunID,
				SourceSnapshotID:              run.Request.SourceSnapshotID,
				ProjectionVersion:             run.Request.ProjectionConfig.ProjectionVersion,
				NormalizedConfigurationSHA256: run.ProjectionConfigDigest,
				NormalizedSourceSHA256:        run.ProjectionSourceDigest,
				VertexCount:                   len(run.GraphView.Vertices),
				EdgeCount:                     len(run.GraphView.Edges),
				CanonicalOutputSHA256:         restoreCanonicalOutputDigest(*run.GraphView),
			})
		}
	}
	sortRestoreOutcomes(rebuilt, skipped)
	return staged, rebuilt, skipped, nil
}

func restoreCanonicalOutputDigest(view GraphView) string {
	// projection_run_id and generated_at are server-owned lifecycle evidence.
	// They remain present in the stored canonical Graph resource but are
	// normalized out of the restore result's semantic-output digest so a known
	// rollback can be retried without changing reconstruction proof.
	view.ProjectionRunID = ""
	view.GeneratedAt = ""
	body, err := canonicalJSON(graphViewCanonicalResource(view))
	if err != nil {
		return ""
	}
	return sha256Hex(body)
}

// restoreResourceUsage centralizes the operation-wide admission accounting so
// every future source registration participates in one shared resource budget.
// Its overflow-safe methods are also independently testable without allocating
// payloads at the production limits.
type restoreResourceUsage struct {
	sourceRegistrations  int
	candidates           int
	normalizedInputBytes int64
	vertices             int
	edges                int
}

func (usage *restoreResourceUsage) valid() bool {
	return usage != nil && usage.sourceRegistrations >= 0 && usage.sourceRegistrations <= RestoreMaximumSourceRegistrations &&
		usage.candidates >= 0 && usage.candidates <= RestoreMaximumCandidates &&
		usage.normalizedInputBytes >= 0 && usage.normalizedInputBytes <= RestoreMaximumNormalizedInputBytes &&
		usage.vertices >= 0 && usage.vertices <= RestoreMaximumVertices &&
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
	if usage == nil || vertices < 0 || edges < 0 || !usage.valid() ||
		vertices > RestoreMaximumVertices-usage.vertices || edges > RestoreMaximumEdges-usage.edges {
		return false
	}
	usage.vertices += vertices
	usage.edges += edges
	return true
}

func registrationSkipReason(entry RestoreSourceRegistryEntry, consistencyPoint time.Time) (RestoreSkipReason, bool) {
	if consistencyPoint.Before(entry.IntroducedAt.UTC()) {
		return RestoreSkipRegistrationNotYetActive, true
	}
	if entry.RetiredAt != nil && !consistencyPoint.Before(entry.RetiredAt.UTC()) {
		return RestoreSkipRegistrationRetired, true
	}
	return "", false
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

func sortRestoreOutcomes(rebuilt []RestoreRebuiltView, skipped []RestoreSkippedCandidate) {
	sort.Slice(rebuilt, func(left, right int) bool {
		if rebuilt[left].SourceRegistrationID != rebuilt[right].SourceRegistrationID {
			return rebuilt[left].SourceRegistrationID < rebuilt[right].SourceRegistrationID
		}
		return rebuilt[left].CandidateID < rebuilt[right].CandidateID
	})
	sort.Slice(skipped, func(left, right int) bool {
		if skipped[left].SourceRegistrationID != skipped[right].SourceRegistrationID {
			return skipped[left].SourceRegistrationID < skipped[right].SourceRegistrationID
		}
		return skipped[left].CandidateID < skipped[right].CandidateID
	})
}

func restorePostconditionDigest(request RestoreRebuildRequest, rebuilt []RestoreRebuiltView, skipped []RestoreSkippedCandidate) (string, error) {
	body, err := restoreCanonicalJSON(map[string]any{
		"algorithm_id":         RestoreAlgorithmID,
		"cleared_table_ids":    restoreGraphTableIDs,
		"rebuilt_views":        rebuilt,
		"restore_operation_id": request.RestoreOperationID.String(),
		"skipped_candidates":   skipped,
		"target_generation_id": request.TargetGenerationID.String(),
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
	base.SkippedCandidates = []RestoreSkippedCandidate{}
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
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
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
