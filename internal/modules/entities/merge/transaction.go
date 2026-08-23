package merge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var ErrMergeTargetNotFound = errors.New("entities: merge target not found")

type MergePreconditionError struct {
	ReasonCode string
	Details    map[string]any
}

func (e *MergePreconditionError) Error() string {
	return fmt.Sprintf("entities: merge precondition failed: %s", e.ReasonCode)
}

type MergeRowVersionConflictError struct {
	SurvivorRecordID          uuid.UUID
	LoserRecordID             uuid.UUID
	SurvivorBaseRowVersion    int64
	LoserBaseRowVersion       int64
	SurvivorCurrentRowVersion int64
	LoserCurrentRowVersion    int64
}

func (e *MergeRowVersionConflictError) Error() string {
	return fmt.Sprintf("entities: merge row version conflict for %s and %s", e.SurvivorRecordID, e.LoserRecordID)
}

type MergeRecordLockedError struct {
	RecordID uuid.UUID
}

func (e *MergeRecordLockedError) Error() string {
	return fmt.Sprintf("entities: merge record locked %s", e.RecordID)
}

type mergeTargetMeta struct {
	RecordID   uuid.UUID
	IncidentID uuid.UUID
	RecordType string
}

type mergePreservedIdentifierRecord struct {
	EntityPreservedIdentifierID uuid.UUID
	IncidentID                  uuid.UUID
	RecordID                    uuid.UUID
	EntityType                  string
	IdentifierType              string
	RawValue                    string
	NormalizedValue             string
	Classification              string
	CreatedAt                   time.Time
	DeletedAt                   *time.Time
}

type mergeAliasRecord struct {
	EntityAliasID  uuid.UUID
	IncidentID     uuid.UUID
	RecordID       uuid.UUID
	EntityType     string
	RawText        string
	NormalizedText string
	Classification string
	CreatedAt      time.Time
	DeletedAt      *time.Time
}

type mergeExactMatchCandidate struct {
	IdentifierClass string
	RawValue        string
	NormalizedValue string
	FromCanonical   bool
}

type mergeIdentifierSeed struct {
	IdentifierType  string
	RawValue        string
	NormalizedValue string
	Classification  string
}

type mergeIdentifierInsert struct {
	Seed        mergeIdentifierSeed
	MutationTag string
}

type mergeIdentifierMutation struct {
	Before map[string]any
	After  map[string]any
}

type mergeAliasMutation struct {
	Before map[string]any
	After  map[string]any
}

type mergeCarryPlan struct {
	SurvivorHost                 hostidentity.HostRecord
	SurvivorIdentity             hostidentity.IdentityRecord
	IdentifierInserts            []mergeIdentifierInsert
	ExactMatchClasses            []MergeExactMatchClassSummary
	IdentifierMutations          []mergeIdentifierMutation
	AliasMutations               []mergeAliasMutation
	AliasActions                 []hostidentity.CollectionAction
	SuggestionAliasesCopiedCount int
	SuggestionAliasDuplicateNoop int
}

type mergeMutation struct {
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
	BeforeSnapshot  *revisions.RecordSnapshot
	AfterSnapshot   *revisions.RecordSnapshot
}

type mergeCounts struct {
	RepointedMentions    int
	RepointedLinks       int
	DedupedLinks         int
	RepointedTags        int
	DedupedTags          int
	RepointedAssessments int
}

func (s *Store) GetMergeRouteIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	row := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
`, recordID)
	var incidentID uuid.UUID
	if err := row.Scan(&incidentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, ErrMergeTargetNotFound
		}
		return uuid.UUID{}, fmt.Errorf("get merge route incident: %w", err)
	}
	return incidentID, nil
}

func (s *Store) MergeEntity(ctx context.Context, actor authn.UserRecord, survivorRecordID uuid.UUID, request MergeRequest, requestHash []byte, requestID string, now time.Time) (MergeResult, error) {
	scopeKey := mergeScopeKey(survivorRecordID, request.LoserRecordID)
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    mergeRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MergeResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MergeResult{}, fmt.Errorf("decode replayed merge payload: %w", err)
		}
		return MergeResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MergeResult{}, fmt.Errorf("query merge idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MergeResult{}, fmt.Errorf("begin merge transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	protectedRecordIDs, err := s.planMergeProtectedRecordIDsTx(ctx, tx, survivorRecordID, request.LoserRecordID, now)
	if err != nil {
		return MergeResult{}, err
	}
	preLockSurvivorMeta, survivorMetaErr := loadMergeTargetMetaTx(ctx, tx, survivorRecordID)
	preLockLoserMeta, loserMetaErr := loadMergeTargetMetaTx(ctx, tx, request.LoserRecordID)
	if survivorMetaErr == nil && loserMetaErr == nil &&
		preLockSurvivorMeta.IncidentID == preLockLoserMeta.IncidentID &&
		preLockSurvivorMeta.RecordType == preLockLoserMeta.RecordType &&
		(preLockSurvivorMeta.RecordType == "host" || preLockSurvivorMeta.RecordType == "identity") {
		conflict, prepareErr := s.hostIdentity.PrepareIdentifierClaimsTx(
			ctx,
			tx,
			preLockSurvivorMeta.IncidentID,
			preLockSurvivorMeta.RecordType,
			survivorRecordID,
			request.LoserRecordID,
		)
		if prepareErr != nil {
			return MergeResult{}, prepareErr
		}
		if conflict != nil {
			return MergeResult{}, mergeIdentifierConflict(preLockSurvivorMeta.RecordType, conflict)
		}
	}
	if err := s.ports.revisions.LockDestructiveOperationRecordsNowaitTx(ctx, tx, protectedRecordIDs); err != nil {
		var locked *entityRecordLockedError
		if errors.As(err, &locked) {
			return MergeResult{}, &MergeRecordLockedError{RecordID: locked.RecordID}
		}
		if errors.Is(err, errEntityRecordEnvelopeNotFound) {
			return MergeResult{}, classifyMissingMergeTargetTx(ctx, tx, survivorRecordID, request.LoserRecordID)
		}
		return MergeResult{}, err
	}
	protectedRecordSet := uuidSet(protectedRecordIDs)
	revalidatedProtectedRecordIDs, err := s.planMergeProtectedRecordIDsTx(ctx, tx, survivorRecordID, request.LoserRecordID, now)
	if err != nil {
		return MergeResult{}, err
	}
	if !sameUUIDSet(protectedRecordIDs, revalidatedProtectedRecordIDs) {
		return MergeResult{}, &MergePreconditionError{ReasonCode: "protected_set_changed"}
	}

	survivorMeta, err := loadMergeTargetMetaTx(ctx, tx, survivorRecordID)
	if errors.Is(err, ErrMergeTargetNotFound) {
		return MergeResult{}, ErrMergeTargetNotFound
	}
	if err != nil {
		return MergeResult{}, err
	}
	loserMeta, err := loadMergeTargetMetaTx(ctx, tx, request.LoserRecordID)
	if err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return MergeResult{}, ErrMergeTargetNotFound
		}
		return MergeResult{}, err
	}
	if survivorRecordID == request.LoserRecordID {
		return MergeResult{}, &MergePreconditionError{ReasonCode: "same_record"}
	}
	if survivorMeta.RecordType != "host" && survivorMeta.RecordType != "identity" {
		return MergeResult{}, &MergePreconditionError{
			ReasonCode: "unsupported_record_type",
			Details: map[string]any{
				"record_type": survivorMeta.RecordType,
			},
		}
	}
	if loserMeta.IncidentID != survivorMeta.IncidentID {
		return MergeResult{}, ErrMergeTargetNotFound
	}
	if loserMeta.RecordType != survivorMeta.RecordType {
		return MergeResult{}, &MergePreconditionError{
			ReasonCode: "record_type_mismatch",
			Details: map[string]any{
				"survivor_record_type": survivorMeta.RecordType,
				"loser_record_type":    loserMeta.RecordType,
			},
		}
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, survivorMeta.IncidentID); err != nil {
		return MergeResult{}, err
	}

	var (
		survivorHost     hostidentity.HostRecord
		loserHost        hostidentity.HostRecord
		survivorIdentity hostidentity.IdentityRecord
		loserIdentity    hostidentity.IdentityRecord
		incidentID       = survivorMeta.IncidentID
	)
	switch survivorMeta.RecordType {
	case "host":
		survivorHost, err = s.loadHostByRecordIDTx(ctx, tx, survivorRecordID)
		if err != nil {
			if errors.Is(err, ErrMergeTargetNotFound) {
				return MergeResult{}, &MergePreconditionError{ReasonCode: "survivor_not_found"}
			}
			return MergeResult{}, err
		}
		loserHost, err = s.loadHostByRecordIDTx(ctx, tx, request.LoserRecordID)
		if err != nil {
			if errors.Is(err, ErrMergeTargetNotFound) {
				return MergeResult{}, ErrMergeTargetNotFound
			}
			return MergeResult{}, err
		}
		if err := validateHostMergePair(survivorHost, loserHost); err != nil {
			return MergeResult{}, err
		}
		if survivorHost.RowVersion != request.SurvivorBaseRowVersion || loserHost.RowVersion != request.LoserBaseRowVersion {
			return MergeResult{}, &MergeRowVersionConflictError{
				SurvivorRecordID:          survivorHost.RecordID,
				LoserRecordID:             loserHost.RecordID,
				SurvivorBaseRowVersion:    request.SurvivorBaseRowVersion,
				LoserBaseRowVersion:       request.LoserBaseRowVersion,
				SurvivorCurrentRowVersion: survivorHost.RowVersion,
				LoserCurrentRowVersion:    loserHost.RowVersion,
			}
		}
	case "identity":
		survivorIdentity, err = s.loadIdentityByRecordIDTx(ctx, tx, survivorRecordID)
		if err != nil {
			if errors.Is(err, ErrMergeTargetNotFound) {
				return MergeResult{}, &MergePreconditionError{ReasonCode: "survivor_not_found"}
			}
			return MergeResult{}, err
		}
		loserIdentity, err = s.loadIdentityByRecordIDTx(ctx, tx, request.LoserRecordID)
		if err != nil {
			if errors.Is(err, ErrMergeTargetNotFound) {
				return MergeResult{}, ErrMergeTargetNotFound
			}
			return MergeResult{}, err
		}
		if err := validateIdentityMergePair(survivorIdentity, loserIdentity); err != nil {
			return MergeResult{}, err
		}
		if survivorIdentity.RowVersion != request.SurvivorBaseRowVersion || loserIdentity.RowVersion != request.LoserBaseRowVersion {
			return MergeResult{}, &MergeRowVersionConflictError{
				SurvivorRecordID:          survivorIdentity.RecordID,
				LoserRecordID:             loserIdentity.RecordID,
				SurvivorBaseRowVersion:    request.SurvivorBaseRowVersion,
				LoserBaseRowVersion:       request.LoserBaseRowVersion,
				SurvivorCurrentRowVersion: survivorIdentity.RowVersion,
				LoserCurrentRowVersion:    loserIdentity.RowVersion,
			}
		}
	default:
		return MergeResult{}, &MergePreconditionError{ReasonCode: "unsupported_record_type"}
	}

	survivorAliases, err := loadMergeAliasesTx(ctx, tx, survivorRecordID, survivorMeta.RecordType)
	if err != nil {
		return MergeResult{}, err
	}
	loserAliases, err := loadMergeAliasesTx(ctx, tx, request.LoserRecordID, survivorMeta.RecordType)
	if err != nil {
		return MergeResult{}, err
	}
	survivorIdentifiers, err := loadMergePreservedIdentifiersTx(ctx, tx, survivorRecordID, survivorMeta.RecordType)
	if err != nil {
		return MergeResult{}, err
	}
	loserIdentifiers, err := loadMergePreservedIdentifiersTx(ctx, tx, request.LoserRecordID, survivorMeta.RecordType)
	if err != nil {
		return MergeResult{}, err
	}
	var carryPlan mergeCarryPlan
	switch survivorMeta.RecordType {
	case "host":
		carryPlan, err = s.planHostMergeCarryForwardTx(ctx, tx, incidentID, survivorHost, loserHost, survivorIdentifiers, loserIdentifiers, survivorAliases, loserAliases, actor.ID, now)
	case "identity":
		carryPlan, err = s.planIdentityMergeCarryForwardTx(ctx, tx, incidentID, survivorIdentity, loserIdentity, survivorIdentifiers, loserIdentifiers, survivorAliases, loserAliases, actor.ID, now)
	}
	if err != nil {
		return MergeResult{}, err
	}
	if _, err := tx.Exec(ctx, `SELECT public.entities_release_active_identifier_claims_v1($1)`, request.LoserRecordID); err != nil {
		return MergeResult{}, fmt.Errorf("release loser active identifier claims: %w", err)
	}
	if err := s.applyMergeCarryForwardTx(ctx, tx, incidentID, survivorMeta.RecordType, survivorRecordID, actor.ID, now, &carryPlan); err != nil {
		return MergeResult{}, err
	}

	var (
		counts    mergeCounts
		mutations []mergeMutation
	)
	survivorBeforeSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, survivorRecordID)
	if err != nil {
		return MergeResult{}, err
	}
	loserBeforeSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, request.LoserRecordID)
	if err != nil {
		return MergeResult{}, err
	}

	mentionMutations, mentionCounts, invalidatedRecords, err := s.ports.mentions.RepointMergedMentionsTx(ctx, tx, incidentID, survivorMeta.RecordType, survivorRecordID, request.LoserRecordID, actor.ID, now)
	if err != nil {
		return MergeResult{}, err
	}
	counts.RepointedMentions = mentionCounts
	mutations = append(mutations, mentionMutations...)

	linkResult, err := s.ports.links.RepointLinksTx(ctx, tx, RepointLinksCommand{
		IncidentID:       incidentID,
		SurvivorRecordID: survivorRecordID,
		LoserRecordID:    request.LoserRecordID,
		ActorUserID:      actor.ID,
		Now:              now,
	})
	if err != nil {
		return MergeResult{}, err
	}
	counts.RepointedLinks = linkResult.RepointedCount
	counts.DedupedLinks = linkResult.DedupedCount
	mutations = append(mutations, mergeMutationsFromLinkEffects(linkResult.Mutations)...)

	tagResult, err := s.ports.links.RepointTagsTx(ctx, tx, RepointTagsCommand{
		IncidentID:       incidentID,
		SurvivorRecordID: survivorRecordID,
		LoserRecordID:    request.LoserRecordID,
		ActorUserID:      actor.ID,
		Now:              now,
	})
	if err != nil {
		return MergeResult{}, err
	}
	counts.RepointedTags = tagResult.RepointedCount
	counts.DedupedTags = tagResult.DedupedCount
	mutations = append(mutations, mergeMutationsFromLinkEffects(tagResult.Mutations)...)

	assessmentResult, err := s.ports.assessments.RepointTx(ctx, tx, AssessmentRepointCommand{
		IncidentID:         incidentID,
		RecordType:         survivorMeta.RecordType,
		SurvivorRecordID:   survivorRecordID,
		LoserRecordID:      request.LoserRecordID,
		ProtectedRecordIDs: sortedUUIDSet(protectedRecordSet),
		Now:                now,
	})
	if err != nil {
		return MergeResult{}, classifyAssessmentRepointError(err)
	}
	counts.RepointedAssessments = assessmentResult.RepointedCount
	mutations = append(mutations, mergeMutationsFromAssessmentMutations(assessmentResult.Mutations)...)

	mutations = append(mutations, identifierMutationsToMergeMutations(carryPlan.IdentifierMutations)...)
	mutations = append(mutations, aliasMutationsToMergeMutations(carryPlan.AliasMutations)...)

	switch survivorMeta.RecordType {
	case "host":
		nextSurvivor := carryPlan.SurvivorHost
		nextSurvivor.RowVersion, err = s.ports.records.AdvanceVersionTx(ctx, tx, survivorHost.RecordID, actor.ID, now.UTC())
		if err != nil {
			return MergeResult{}, err
		}
		nextSurvivor.UpdatedAt = now.UTC()
		nextSurvivor.UpdatedByUser = actor.ID
		if err := s.hostIdentity.UpdateHostTx(ctx, tx, nextSurvivor); err != nil {
			return MergeResult{}, err
		}
		if err := s.ports.projections.RefreshHostTx(ctx, tx, nextSurvivor.RecordID); err != nil {
			return MergeResult{}, err
		}

		nextLoser := loserHost
		nextLoser.HostState = "merged"
		nextLoser.MergedIntoRecordID = &survivorRecordID
		nextLoser.RowVersion, err = s.ports.records.AdvanceVersionTx(ctx, tx, loserHost.RecordID, actor.ID, now.UTC())
		if err != nil {
			return MergeResult{}, err
		}
		nextLoser.UpdatedAt = now.UTC()
		nextLoser.UpdatedByUser = actor.ID
		if err := s.hostIdentity.UpdateHostTx(ctx, tx, nextLoser); err != nil {
			return MergeResult{}, err
		}
		if err := s.ports.projections.DeleteHostTx(ctx, tx, nextLoser.RecordID); err != nil {
			return MergeResult{}, err
		}

		currentAliases, loadAliasesErr := loadMergeAliasesTx(ctx, tx, nextSurvivor.RecordID, "host")
		if loadAliasesErr != nil {
			return MergeResult{}, loadAliasesErr
		}
		nextSurvivor.SuggestionOnlyAliases = aliasValuesFromRecords(currentAliases)
		nextLoser.SuggestionOnlyAliases = loserHost.SuggestionOnlyAliases
		survivorHost = nextSurvivor
		loserHost = nextLoser
	case "identity":
		nextSurvivor := carryPlan.SurvivorIdentity
		nextSurvivor.RowVersion, err = s.ports.records.AdvanceVersionTx(ctx, tx, survivorIdentity.RecordID, actor.ID, now.UTC())
		if err != nil {
			return MergeResult{}, err
		}
		nextSurvivor.UpdatedAt = now.UTC()
		nextSurvivor.UpdatedByUser = actor.ID
		if err := s.hostIdentity.UpdateIdentityTx(ctx, tx, nextSurvivor); err != nil {
			return MergeResult{}, err
		}
		if err := s.ports.projections.RefreshIdentityTx(ctx, tx, nextSurvivor.RecordID); err != nil {
			return MergeResult{}, err
		}

		nextLoser := loserIdentity
		nextLoser.IdentityState = "merged"
		nextLoser.MergedIntoRecordID = &survivorRecordID
		nextLoser.RowVersion, err = s.ports.records.AdvanceVersionTx(ctx, tx, loserIdentity.RecordID, actor.ID, now.UTC())
		if err != nil {
			return MergeResult{}, err
		}
		nextLoser.UpdatedAt = now.UTC()
		nextLoser.UpdatedByUser = actor.ID
		if err := s.hostIdentity.UpdateIdentityTx(ctx, tx, nextLoser); err != nil {
			return MergeResult{}, err
		}
		if err := s.ports.projections.DeleteIdentityTx(ctx, tx, nextLoser.RecordID); err != nil {
			return MergeResult{}, err
		}

		currentAliases, loadAliasesErr := loadMergeAliasesTx(ctx, tx, nextSurvivor.RecordID, "identity")
		if loadAliasesErr != nil {
			return MergeResult{}, loadAliasesErr
		}
		nextSurvivor.SuggestionOnlyAliases = aliasValuesFromRecords(currentAliases)
		nextLoser.SuggestionOnlyAliases = loserIdentity.SuggestionOnlyAliases
		survivorIdentity = nextSurvivor
		loserIdentity = nextLoser
	}
	survivorAfterSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, survivorRecordID)
	if err != nil {
		return MergeResult{}, err
	}
	loserAfterSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, request.LoserRecordID)
	if err != nil {
		return MergeResult{}, err
	}

	changeSetID, err := s.ports.revisions.AppendChangeSetTx(ctx, tx, entityChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      mergeRouteKey,
		Reason:      request.Reason,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MergeResult{}, err
	}

	sequenceNo := 1
	switch survivorMeta.RecordType {
	case "host":
		beforeVersionID := entityVersionID("host", survivorHost.RecordID, survivorHost.RowVersion-1)
		afterVersionID := entityVersionID("host", survivorHost.RecordID, survivorHost.RowVersion)
		if err := s.ports.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      "host",
			RecordID:        survivorHost.RecordID,
			OperationKind:   "patch",
			BeforeVersionID: &beforeVersionID,
			AfterVersionID:  &afterVersionID,
			BeforeSnapshot:  &survivorBeforeSnapshot,
			AfterSnapshot:   &survivorAfterSnapshot,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
		loserBeforeVersionID := entityVersionID("host", loserHost.RecordID, loserHost.RowVersion-1)
		loserAfterVersionID := entityVersionID("host", loserHost.RecordID, loserHost.RowVersion)
		if err := s.ports.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      "host",
			RecordID:        loserHost.RecordID,
			OperationKind:   "patch",
			BeforeVersionID: &loserBeforeVersionID,
			AfterVersionID:  &loserAfterVersionID,
			BeforeSnapshot:  &loserBeforeSnapshot,
			AfterSnapshot:   &loserAfterSnapshot,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
		if err := s.ports.revisions.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID:    changeSetID,
			RecordID:       survivorHost.RecordID,
			RowVersion:     survivorHost.RowVersion,
			BeforeSnapshot: &survivorBeforeSnapshot,
			AfterSnapshot:  &survivorAfterSnapshot,
		}); err != nil {
			return MergeResult{}, err
		}
		if err := s.ports.revisions.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID:    changeSetID,
			RecordID:       loserHost.RecordID,
			RowVersion:     loserHost.RowVersion,
			BeforeSnapshot: &loserBeforeSnapshot,
			AfterSnapshot:  &loserAfterSnapshot,
		}); err != nil {
			return MergeResult{}, err
		}
	case "identity":
		beforeVersionID := entityVersionID("identity", survivorIdentity.RecordID, survivorIdentity.RowVersion-1)
		afterVersionID := entityVersionID("identity", survivorIdentity.RecordID, survivorIdentity.RowVersion)
		if err := s.ports.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      "identity",
			RecordID:        survivorIdentity.RecordID,
			OperationKind:   "patch",
			BeforeVersionID: &beforeVersionID,
			AfterVersionID:  &afterVersionID,
			BeforeSnapshot:  &survivorBeforeSnapshot,
			AfterSnapshot:   &survivorAfterSnapshot,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
		loserBeforeVersionID := entityVersionID("identity", loserIdentity.RecordID, loserIdentity.RowVersion-1)
		loserAfterVersionID := entityVersionID("identity", loserIdentity.RecordID, loserIdentity.RowVersion)
		if err := s.ports.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      "identity",
			RecordID:        loserIdentity.RecordID,
			OperationKind:   "patch",
			BeforeVersionID: &loserBeforeVersionID,
			AfterVersionID:  &loserAfterVersionID,
			BeforeSnapshot:  &loserBeforeSnapshot,
			AfterSnapshot:   &loserAfterSnapshot,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
		if err := s.ports.revisions.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID:    changeSetID,
			RecordID:       survivorIdentity.RecordID,
			RowVersion:     survivorIdentity.RowVersion,
			BeforeSnapshot: &survivorBeforeSnapshot,
			AfterSnapshot:  &survivorAfterSnapshot,
		}); err != nil {
			return MergeResult{}, err
		}
		if err := s.ports.revisions.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID:    changeSetID,
			RecordID:       loserIdentity.RecordID,
			RowVersion:     loserIdentity.RowVersion,
			BeforeSnapshot: &loserBeforeSnapshot,
			AfterSnapshot:  &loserAfterSnapshot,
		}); err != nil {
			return MergeResult{}, err
		}
	}

	for _, mutation := range mutations {
		if mutation.BeforeSnapshot != nil || mutation.AfterSnapshot != nil {
			recordID, parseErr := uuid.Parse(mutation.TargetID)
			if parseErr != nil || recordID == uuid.Nil {
				return MergeResult{}, fmt.Errorf("append captured merge mutation: invalid record id %q", mutation.TargetID)
			}
			if err := s.ports.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
				ChangeSetID:     changeSetID,
				SequenceNo:      sequenceNo,
				TargetKind:      mutation.TargetKind,
				RecordID:        recordID,
				OperationKind:   mutation.OperationKind,
				BeforeVersionID: mutation.BeforeVersionID,
				AfterVersionID:  mutation.AfterVersionID,
				BeforeSnapshot:  mutation.BeforeSnapshot,
				AfterSnapshot:   mutation.AfterSnapshot,
			}); err != nil {
				return MergeResult{}, err
			}
			sequenceNo++
			continue
		}
		if err := s.ports.revisions.AppendMutationTx(ctx, tx, entityMutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      mutation.TargetKind,
			TargetID:        mutation.TargetID,
			OperationKind:   mutation.OperationKind,
			BeforeVersionID: mutation.BeforeVersionID,
			AfterVersionID:  mutation.AfterVersionID,
			BeforeValue:     mutation.BeforeValue,
			AfterValue:      mutation.AfterValue,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
	}

	mentionTimelineInvalidations, err := s.ports.timeline.LoadTimelineInvalidationsTx(ctx, tx, invalidatedRecords)
	if err != nil {
		return MergeResult{}, err
	}
	relationshipTimelineInvalidations, err := s.ports.timeline.LoadRelationshipInvalidationsTx(ctx, tx, linkResult.LinkTypesBySourceRecordID)
	if err != nil {
		return MergeResult{}, err
	}
	timelineInvalidations := mergeTimelineInvalidations(mentionTimelineInvalidations, relationshipTimelineInvalidations)
	timelineRecordIDs := make([]uuid.UUID, 0, len(timelineInvalidations))
	for _, invalidation := range timelineInvalidations {
		timelineRecordIDs = append(timelineRecordIDs, invalidation.RecordID)
	}
	if err := s.ports.timeline.RefreshTimelineProjectionRowsTx(ctx, tx, timelineRecordIDs); err != nil {
		return MergeResult{}, err
	}

	result := MergeResult{
		StatusCode:       http.StatusOK,
		IncidentID:       incidentID,
		RecordType:       survivorMeta.RecordType,
		SurvivorRecordID: survivorRecordID,
		LoserRecordID:    request.LoserRecordID,
		ChangeSetID:      changeSetID,
		MergeSummary: MergeSummary{
			RecordType:                    survivorMeta.RecordType,
			RepointedMentionResolutionCnt: counts.RepointedMentions,
			RepointedLinkCount:            counts.RepointedLinks,
			DedupedLinkCount:              counts.DedupedLinks,
			RepointedTagCount:             counts.RepointedTags,
			DedupedTagCount:               counts.DedupedTags,
			RepointedAssessmentCount:      counts.RepointedAssessments,
			ExactMatchClasses:             carryPlan.ExactMatchClasses,
			SuggestionAliasesCopiedCount:  carryPlan.SuggestionAliasesCopiedCount,
			SuggestionAliasDuplicateNoop:  carryPlan.SuggestionAliasDuplicateNoop,
			ProvenanceOnlyRetainedCount:   countProvenanceOnlyIdentifiers(loserIdentifiers),
		},
		TimelineInvalidations: timelineInvalidations,
	}
	switch survivorMeta.RecordType {
	case "host":
		result.SurvivorRowVersion = survivorHost.RowVersion
		result.LoserRowVersion = loserHost.RowVersion
	case "identity":
		result.SurvivorRowVersion = survivorIdentity.RowVersion
		result.LoserRowVersion = loserIdentity.RowVersion
	}
	result.Payload = buildMergePayload(result)
	if err := s.appendMergeIntentsTx(ctx, tx, actor.ID, request.ClientTxnID, result, now.UTC()); err != nil {
		return MergeResult{}, err
	}

	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, result.Payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MergeResult{}, authn.ErrClientTxnConflict
		}
		return MergeResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MergeResult{}, fmt.Errorf("commit merge transaction: %w", err)
	}

	return result, nil
}
