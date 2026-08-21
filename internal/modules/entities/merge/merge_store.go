package merge

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
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

type mergeIdentifierInsert struct {
	Seed        identifierSeed
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
	SurvivorHost                 HostRecord
	SurvivorIdentity             IdentityRecord
	IdentifierInserts            []mergeIdentifierInsert
	AliasAdds                    []CollectionAction
	ExactMatchClasses            []MergeExactMatchClassSummary
	IdentifierMutations          []mergeIdentifierMutation
	AliasMutations               []mergeAliasMutation
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

	protectedRecordIDs, err := s.planMergeProtectedRecordIDsTx(ctx, tx, survivorRecordID, request.LoserRecordID)
	if err != nil {
		return MergeResult{}, err
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
		survivorHost     HostRecord
		loserHost        HostRecord
		survivorIdentity IdentityRecord
		loserIdentity    IdentityRecord
		incidentID       = survivorMeta.IncidentID
	)
	switch survivorMeta.RecordType {
	case "host":
		survivorHost, err = loadHostByRecordIDTx(ctx, tx, survivorRecordID)
		if err != nil {
			if errors.Is(err, ErrMergeTargetNotFound) {
				return MergeResult{}, &MergePreconditionError{ReasonCode: "survivor_not_found"}
			}
			return MergeResult{}, err
		}
		loserHost, err = loadHostByRecordIDTx(ctx, tx, request.LoserRecordID)
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
		survivorIdentity, err = loadIdentityByRecordIDTx(ctx, tx, survivorRecordID)
		if err != nil {
			if errors.Is(err, ErrMergeTargetNotFound) {
				return MergeResult{}, &MergePreconditionError{ReasonCode: "survivor_not_found"}
			}
			return MergeResult{}, err
		}
		loserIdentity, err = loadIdentityByRecordIDTx(ctx, tx, request.LoserRecordID)
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

	assessmentMutations, repointedAssessments, err := s.ports.assessments.RepointMergedAssessmentsTx(ctx, tx, incidentID, survivorMeta.RecordType, survivorRecordID, request.LoserRecordID, protectedRecordSet, now)
	if err != nil {
		return MergeResult{}, err
	}
	counts.RepointedAssessments = repointedAssessments
	mutations = append(mutations, assessmentMutations...)

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
		if err := hostidentity.UpdateHostTx(ctx, tx, nextSurvivor); err != nil {
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
		if err := hostidentity.UpdateHostTx(ctx, tx, nextLoser); err != nil {
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
		if err := hostidentity.UpdateIdentityTx(ctx, tx, nextSurvivor); err != nil {
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
		if err := hostidentity.UpdateIdentityTx(ctx, tx, nextLoser); err != nil {
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
	result.Payload = BuildMergePayload(result)
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

func mergeTimelineInvalidations(groups ...[]MergeTimelineInvalidation) []MergeTimelineInvalidation {
	byRecord := map[uuid.UUID]MergeTimelineInvalidation{}
	for _, group := range groups {
		for _, invalidation := range group {
			current := byRecord[invalidation.RecordID]
			if current.RecordID == uuid.Nil {
				current.RecordID = invalidation.RecordID
				current.RowVersion = invalidation.RowVersion
			}
			current.ChangedFieldKeys = append(current.ChangedFieldKeys, invalidation.ChangedFieldKeys...)
			byRecord[invalidation.RecordID] = current
		}
	}
	recordIDs := make([]uuid.UUID, 0, len(byRecord))
	for recordID := range byRecord {
		recordIDs = append(recordIDs, recordID)
	}
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	result := make([]MergeTimelineInvalidation, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		invalidation := byRecord[recordID]
		slices.Sort(invalidation.ChangedFieldKeys)
		invalidation.ChangedFieldKeys = slices.Compact(invalidation.ChangedFieldKeys)
		result = append(result, invalidation)
	}
	return result
}

func classifyMissingMergeTargetTx(ctx context.Context, tx pgx.Tx, survivorRecordID uuid.UUID, loserRecordID uuid.UUID) error {
	if _, err := loadMergeTargetMetaTx(ctx, tx, survivorRecordID); err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return ErrMergeTargetNotFound
		}
		return err
	}
	if _, err := loadMergeTargetMetaTx(ctx, tx, loserRecordID); err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return ErrMergeTargetNotFound
		}
		return err
	}
	return &MergePreconditionError{ReasonCode: "target_not_found"}
}

func (s *Store) planMergeProtectedRecordIDsTx(ctx context.Context, tx pgx.Tx, survivorRecordID uuid.UUID, loserRecordID uuid.UUID) ([]uuid.UUID, error) {
	recordIDs := []uuid.UUID{survivorRecordID, loserRecordID}
	survivorMeta, err := loadMergeTargetMetaTx(ctx, tx, survivorRecordID)
	if err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return nil, ErrMergeTargetNotFound
		}
		return nil, err
	}
	loserMeta, err := loadMergeTargetMetaTx(ctx, tx, loserRecordID)
	if err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return nil, ErrMergeTargetNotFound
		}
		return nil, err
	}
	if loserMeta.IncidentID != survivorMeta.IncidentID {
		return nil, ErrMergeTargetNotFound
	}
	if survivorRecordID == loserRecordID ||
		(survivorMeta.RecordType != "host" && survivorMeta.RecordType != "identity") ||
		loserMeta.RecordType != survivorMeta.RecordType {
		return recordIDs, nil
	}
	assessmentRecordIDs, err := s.ports.assessments.LoadMergeProtectedRecordIDsTx(ctx, tx, survivorMeta.IncidentID, survivorMeta.RecordType, loserRecordID)
	if err != nil {
		return nil, err
	}
	return append(recordIDs, assessmentRecordIDs...), nil
}

func uuidSet(recordIDs []uuid.UUID) map[uuid.UUID]struct{} {
	result := make(map[uuid.UUID]struct{}, len(recordIDs))
	for _, recordID := range recordIDs {
		result[recordID] = struct{}{}
	}
	return result
}

func loadMergeTargetMetaTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (mergeTargetMeta, error) {
	row := tx.QueryRow(ctx, `
SELECT incident_id, record_type
  FROM records
 WHERE record_id = $1
`, recordID)
	var meta mergeTargetMeta
	meta.RecordID = recordID
	if err := row.Scan(&meta.IncidentID, &meta.RecordType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return mergeTargetMeta{}, ErrMergeTargetNotFound
		}
		return mergeTargetMeta{}, fmt.Errorf("load merge target meta: %w", err)
	}
	return meta, nil
}

func loadHostByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (HostRecord, error) {
	record, err := hostidentity.LoadHostByRecordIDTx(ctx, tx, recordID)
	if errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
		return HostRecord{}, ErrMergeTargetNotFound
	}
	if err != nil {
		return HostRecord{}, fmt.Errorf("load host merge target: %w", err)
	}
	return record, nil
}

func loadIdentityByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (IdentityRecord, error) {
	record, err := hostidentity.LoadIdentityByRecordIDTx(ctx, tx, recordID)
	if errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
		return IdentityRecord{}, ErrMergeTargetNotFound
	}
	if err != nil {
		return IdentityRecord{}, fmt.Errorf("load identity merge target: %w", err)
	}
	return record, nil
}

func validateHostMergePair(survivor HostRecord, loser HostRecord) error {
	if survivor.IncidentID != loser.IncidentID {
		return &MergePreconditionError{ReasonCode: "cross_incident_pair"}
	}
	if survivor.HostState != "stub" && survivor.HostState != "canonical" {
		return &MergePreconditionError{ReasonCode: "survivor_not_active"}
	}
	if loser.HostState != "stub" && loser.HostState != "canonical" {
		return &MergePreconditionError{ReasonCode: "loser_not_active"}
	}
	return nil
}

func validateIdentityMergePair(survivor IdentityRecord, loser IdentityRecord) error {
	if survivor.IncidentID != loser.IncidentID {
		return &MergePreconditionError{ReasonCode: "cross_incident_pair"}
	}
	if survivor.IdentityState != "stub" && survivor.IdentityState != "canonical" {
		return &MergePreconditionError{ReasonCode: "survivor_not_active"}
	}
	if loser.IdentityState != "stub" && loser.IdentityState != "canonical" {
		return &MergePreconditionError{ReasonCode: "loser_not_active"}
	}
	return nil
}

func (s *Store) planHostMergeCarryForwardTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, survivor HostRecord, loser HostRecord, survivorIdentifiers []mergePreservedIdentifierRecord, loserIdentifiers []mergePreservedIdentifierRecord, survivorAliases []mergeAliasRecord, loserAliases []mergeAliasRecord, actorUserID uuid.UUID, now time.Time) (mergeCarryPlan, error) {
	summary, candidates := buildMergeClassSummary(hostExactMatchPrecedence, hostCanonicalCandidates(loser), loserIdentifiers)
	next := survivor
	existingValues, canonicalFilled := hostExistingIdentifierState(survivor, survivorIdentifiers)
	plan, err := s.applyCarryPlanTx(ctx, tx, incidentID, "host", survivor.RecordID, loser.RecordID, hostExactMatchPrecedence, existingValues, canonicalFilled, candidates, summary, actorUserID, now)
	if err != nil {
		return mergeCarryPlan{}, err
	}
	for _, candidate := range plan.IdentifierInserts {
		switch candidate.Seed.IdentifierType {
		case "aad_device_id":
			if next.AADDeviceID == nil && candidate.MutationTag == "promoted" {
				next.AADDeviceID = stringPointer(candidate.Seed.RawValue)
			}
		case "fqdn":
			if next.FQDN == nil && candidate.MutationTag == "promoted" {
				next.FQDN = stringPointer(candidate.Seed.RawValue)
			}
		case "hostname":
			if next.Hostname == nil && candidate.MutationTag == "promoted" {
				next.Hostname = stringPointer(candidate.Seed.RawValue)
			}
		}
		if _, err := hostidentity.SyncPreservedIdentifiersTx(ctx, tx, incidentID, survivor.RecordID, "host", []identifierSeed{candidate.Seed}, actorUserID, now); err != nil {
			return mergeCarryPlan{}, err
		}
		plan.IdentifierMutations = append(plan.IdentifierMutations, mergeIdentifierMutation{
			After: buildMergePreservedIdentifierValueFromSeed(incidentID, survivor.RecordID, "host", candidate.Seed),
		})
	}
	plan.SurvivorHost = next
	if len(loserAliases) > 0 {
		actions := aliasActionsFromRecords(loserAliases)
		syncResult, err := hostidentity.SyncEntityAliasesTx(ctx, tx, incidentID, survivor.RecordID, "host", actions, actorUserID, now)
		if err != nil {
			return mergeCarryPlan{}, err
		}
		plan.SuggestionAliasesCopiedCount += len(syncResult.Added)
		plan.SuggestionAliasDuplicateNoop += syncResult.DuplicateNoopCount
		for _, alias := range syncResult.Added {
			plan.AliasMutations = append(plan.AliasMutations, mergeAliasMutation{
				After: alias.MutationValue(),
			})
		}
	}
	return plan, nil
}

func (s *Store) planIdentityMergeCarryForwardTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, survivor IdentityRecord, loser IdentityRecord, survivorIdentifiers []mergePreservedIdentifierRecord, loserIdentifiers []mergePreservedIdentifierRecord, survivorAliases []mergeAliasRecord, loserAliases []mergeAliasRecord, actorUserID uuid.UUID, now time.Time) (mergeCarryPlan, error) {
	summary, candidates := buildMergeClassSummary(identityExactMatchPrecedence, identityCanonicalCandidates(loser), loserIdentifiers)
	next := survivor
	existingValues, canonicalFilled := identityExistingIdentifierState(survivor, survivorIdentifiers)
	plan, err := s.applyCarryPlanTx(ctx, tx, incidentID, "identity", survivor.RecordID, loser.RecordID, identityExactMatchPrecedence, existingValues, canonicalFilled, candidates, summary, actorUserID, now)
	if err != nil {
		return mergeCarryPlan{}, err
	}
	for _, candidate := range plan.IdentifierInserts {
		switch candidate.Seed.IdentifierType {
		case "aad_object_id":
			if next.AADObjectID == nil && candidate.MutationTag == "promoted" {
				next.AADObjectID = stringPointer(candidate.Seed.RawValue)
			}
		case "sid":
			if next.SID == nil && candidate.MutationTag == "promoted" {
				next.SID = stringPointer(candidate.Seed.RawValue)
			}
		case "upn":
			if next.UPN == nil && candidate.MutationTag == "promoted" {
				next.UPN = stringPointer(candidate.Seed.RawValue)
			}
		case "email":
			if next.Email == nil && candidate.MutationTag == "promoted" {
				next.Email = stringPointer(candidate.Seed.RawValue)
			}
		case "sam_account_name":
			if next.SamAccountName == nil && candidate.MutationTag == "promoted" {
				next.SamAccountName = stringPointer(candidate.Seed.RawValue)
			}
		}
		if _, err := hostidentity.SyncPreservedIdentifiersTx(ctx, tx, incidentID, survivor.RecordID, "identity", []identifierSeed{candidate.Seed}, actorUserID, now); err != nil {
			return mergeCarryPlan{}, err
		}
		plan.IdentifierMutations = append(plan.IdentifierMutations, mergeIdentifierMutation{
			After: buildMergePreservedIdentifierValueFromSeed(incidentID, survivor.RecordID, "identity", candidate.Seed),
		})
	}
	plan.SurvivorIdentity = next
	if len(loserAliases) > 0 {
		actions := aliasActionsFromRecords(loserAliases)
		syncResult, err := hostidentity.SyncEntityAliasesTx(ctx, tx, incidentID, survivor.RecordID, "identity", actions, actorUserID, now)
		if err != nil {
			return mergeCarryPlan{}, err
		}
		plan.SuggestionAliasesCopiedCount += len(syncResult.Added)
		plan.SuggestionAliasDuplicateNoop += syncResult.DuplicateNoopCount
		for _, alias := range syncResult.Added {
			plan.AliasMutations = append(plan.AliasMutations, mergeAliasMutation{
				After: alias.MutationValue(),
			})
		}
	}
	return plan, nil
}

func (s *Store) applyCarryPlanTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, precedence []string, survivorExisting map[string]map[string]struct{}, canonicalFilled map[string]bool, candidates map[string][]mergeExactMatchCandidate, summary map[string]MergeExactMatchClassSummary, actorUserID uuid.UUID, now time.Time) (mergeCarryPlan, error) {
	plan := mergeCarryPlan{
		IdentifierInserts: make([]mergeIdentifierInsert, 0),
		ExactMatchClasses: make([]MergeExactMatchClassSummary, 0, len(precedence)),
	}
	for _, identifierClass := range precedence {
		classSummary := summary[identifierClass]
		currentSet := survivorExisting[identifierClass]
		if currentSet == nil {
			currentSet = make(map[string]struct{})
			survivorExisting[identifierClass] = currentSet
		}
		promoted := false
		for _, candidate := range candidates[identifierClass] {
			if _, ok := currentSet[candidate.NormalizedValue]; ok {
				classSummary.DuplicateNoopCount++
				continue
			}
			conflictingRecordID, found, err := findThirdPartyExactMatchConflictTx(ctx, tx, incidentID, entityType, identifierClass, candidate.NormalizedValue, survivorRecordID, loserRecordID)
			if err != nil {
				return mergeCarryPlan{}, err
			}
			if found {
				classSummary.BlockedConflict++
				return mergeCarryPlan{}, &MergePreconditionError{
					ReasonCode: "carry_forward_identifier_collision",
					Details: map[string]any{
						"record_type":        entityType,
						"identifier_class":   identifierClass,
						"normalized_value":   candidate.NormalizedValue,
						"blocking_record_id": conflictingRecordID.String(),
					},
				}
			}

			insert := mergeIdentifierInsert{
				Seed: identifierSeed{
					IdentifierType: identifierClass,
					RawValue:       candidate.RawValue,
					Classification: "exact_match_reuse",
				},
				MutationTag: "carried",
			}
			if !promoted && !canonicalFilled[identifierClass] {
				insert.MutationTag = "promoted"
				classSummary.PromotedCount++
				promoted = true
				canonicalFilled[identifierClass] = true
			} else {
				classSummary.CarriedCount++
			}
			plan.IdentifierInserts = append(plan.IdentifierInserts, insert)
			currentSet[candidate.NormalizedValue] = struct{}{}
		}
		plan.ExactMatchClasses = append(plan.ExactMatchClasses, classSummary)
	}
	_ = actorUserID
	_ = now
	return plan, nil
}

func findThirdPartyExactMatchConflictTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, identifierClass string, normalizedValue string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID) (uuid.UUID, bool, error) {
	switch entityType {
	case "host":
		rows, err := tx.Query(ctx, `
SELECT record_id, aad_device_id, fqdn, hostname
  FROM hosts
 WHERE incident_id = $1
   AND host_state IN ('stub', 'canonical')
   AND record_id <> $2
   AND record_id <> $3
`, incidentID, survivorRecordID, loserRecordID)
		if err != nil {
			return uuid.UUID{}, false, fmt.Errorf("load host conflict candidates: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var (
				recordID    uuid.UUID
				aadDeviceID pgtype.Text
				fqdn        pgtype.Text
				hostname    pgtype.Text
			)
			if err := rows.Scan(&recordID, &aadDeviceID, &fqdn, &hostname); err != nil {
				return uuid.UUID{}, false, fmt.Errorf("scan host conflict candidate: %w", err)
			}
			record := HostRecord{
				RecordID:    recordID,
				AADDeviceID: textPointer(aadDeviceID),
				FQDN:        textPointer(fqdn),
				Hostname:    textPointer(hostname),
			}
			if hostidentity.HostCanonicalNormalized(record, identifierClass) == normalizedValue {
				return recordID, true, nil
			}
		}
		if err := rows.Err(); err != nil {
			return uuid.UUID{}, false, fmt.Errorf("iterate host conflict candidates: %w", err)
		}
	case "identity":
		rows, err := tx.Query(ctx, `
SELECT record_id, aad_object_id, sid, upn, email::text, sam_account_name
  FROM identities
 WHERE incident_id = $1
   AND identity_state IN ('stub', 'canonical')
   AND record_id <> $2
   AND record_id <> $3
`, incidentID, survivorRecordID, loserRecordID)
		if err != nil {
			return uuid.UUID{}, false, fmt.Errorf("load identity conflict candidates: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var (
				recordID       uuid.UUID
				aadObjectID    pgtype.Text
				sid            pgtype.Text
				upn            pgtype.Text
				email          pgtype.Text
				samAccountName pgtype.Text
			)
			if err := rows.Scan(&recordID, &aadObjectID, &sid, &upn, &email, &samAccountName); err != nil {
				return uuid.UUID{}, false, fmt.Errorf("scan identity conflict candidate: %w", err)
			}
			record := IdentityRecord{
				RecordID:       recordID,
				AADObjectID:    textPointer(aadObjectID),
				SID:            textPointer(sid),
				UPN:            textPointer(upn),
				Email:          textPointer(email),
				SamAccountName: textPointer(samAccountName),
			}
			if hostidentity.IdentityCanonicalNormalized(record, identifierClass) == normalizedValue {
				return recordID, true, nil
			}
		}
		if err := rows.Err(); err != nil {
			return uuid.UUID{}, false, fmt.Errorf("iterate identity conflict candidates: %w", err)
		}
	}

	rows, err := tx.Query(ctx, `
SELECT record_id
  FROM entity_preserved_identifiers
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
   AND classification = 'exact_match_reuse'
   AND deleted_at IS NULL
   AND record_id <> $5
   AND record_id <> $6
`, incidentID, entityType, identifierClass, normalizedValue, survivorRecordID, loserRecordID)
	if err != nil {
		return uuid.UUID{}, false, fmt.Errorf("load preserved identifier conflicts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var recordID uuid.UUID
		if err := rows.Scan(&recordID); err != nil {
			return uuid.UUID{}, false, fmt.Errorf("scan preserved identifier conflict: %w", err)
		}
		active, err := isEntityRecordActiveTx(ctx, tx, entityType, recordID)
		if err != nil {
			return uuid.UUID{}, false, err
		}
		if active {
			return recordID, true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return uuid.UUID{}, false, fmt.Errorf("iterate preserved identifier conflicts: %w", err)
	}
	return uuid.UUID{}, false, nil
}

func isEntityRecordActiveTx(ctx context.Context, tx pgx.Tx, entityType string, recordID uuid.UUID) (bool, error) {
	switch entityType {
	case "host":
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM hosts
     WHERE record_id = $1
       AND host_state IN ('stub', 'canonical')
)`, recordID).Scan(&exists); err != nil {
			return false, fmt.Errorf("query active host record: %w", err)
		}
		return exists, nil
	case "identity":
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM identities
     WHERE record_id = $1
       AND identity_state IN ('stub', 'canonical')
)`, recordID).Scan(&exists); err != nil {
			return false, fmt.Errorf("query active identity record: %w", err)
		}
		return exists, nil
	default:
		return false, nil
	}
}

func loadMergeAliasesTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) ([]mergeAliasRecord, error) {
	rows, err := tx.Query(ctx, `
SELECT
    entity_alias_id,
    incident_id,
    record_id,
    entity_type,
    raw_text,
    normalized_text,
    classification,
    created_at,
    deleted_at
  FROM entity_aliases
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY normalized_text ASC, created_at ASC, entity_alias_id ASC
 FOR UPDATE
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load merge aliases: %w", err)
	}
	defer rows.Close()

	records := make([]mergeAliasRecord, 0)
	for rows.Next() {
		record, err := scanMergeAliasRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate merge aliases: %w", err)
	}
	return records, nil
}

func loadMergePreservedIdentifiersTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) ([]mergePreservedIdentifierRecord, error) {
	rows, err := tx.Query(ctx, `
SELECT
    entity_preserved_identifier_id,
    incident_id,
    record_id,
    entity_type,
    identifier_type,
    raw_value,
    normalized_value,
    classification,
    created_at,
    deleted_at
  FROM entity_preserved_identifiers
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY identifier_type ASC, normalized_value ASC, created_at ASC, entity_preserved_identifier_id ASC
 FOR UPDATE
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load merge preserved identifiers: %w", err)
	}
	defer rows.Close()

	records := make([]mergePreservedIdentifierRecord, 0)
	for rows.Next() {
		record, err := scanMergePreservedIdentifierRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate merge preserved identifiers: %w", err)
	}
	return records, nil
}

func buildMergeClassSummary(precedence []string, canonicalCandidates map[string][]mergeExactMatchCandidate, preservedIdentifiers []mergePreservedIdentifierRecord) (map[string]MergeExactMatchClassSummary, map[string][]mergeExactMatchCandidate) {
	summary := make(map[string]MergeExactMatchClassSummary, len(precedence))
	candidates := make(map[string][]mergeExactMatchCandidate, len(precedence))
	for _, identifierClass := range precedence {
		summary[identifierClass] = MergeExactMatchClassSummary{IdentifierClass: identifierClass}
		seen := make(map[string]struct{})
		for _, candidate := range canonicalCandidates[identifierClass] {
			if _, ok := seen[candidate.NormalizedValue]; ok {
				continue
			}
			candidates[identifierClass] = append(candidates[identifierClass], candidate)
			seen[candidate.NormalizedValue] = struct{}{}
		}
		for _, identifier := range preservedIdentifiers {
			if identifier.IdentifierType != identifierClass {
				continue
			}
			current := summary[identifierClass]
			switch identifier.Classification {
			case "exact_match_reuse":
				if _, ok := seen[identifier.NormalizedValue]; ok {
					summary[identifierClass] = current
					continue
				}
				candidates[identifierClass] = append(candidates[identifierClass], mergeExactMatchCandidate{
					IdentifierClass: identifierClass,
					RawValue:        identifier.RawValue,
					NormalizedValue: identifier.NormalizedValue,
				})
				seen[identifier.NormalizedValue] = struct{}{}
			case "provenance_only":
				current.ProvenanceOnly++
			case "suggestion_only":
				current.SuggestionOnly++
			}
			summary[identifierClass] = current
		}
	}
	return summary, candidates
}

func hostCanonicalCandidates(record HostRecord) map[string][]mergeExactMatchCandidate {
	candidates := make(map[string][]mergeExactMatchCandidate, len(hostExactMatchPrecedence))
	if normalized := normalizeOptionalIdentifier("aad_device_id", record.AADDeviceID); normalized != "" {
		candidates["aad_device_id"] = append(candidates["aad_device_id"], mergeExactMatchCandidate{IdentifierClass: "aad_device_id", RawValue: *record.AADDeviceID, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := normalizeOptionalIdentifier("fqdn", record.FQDN); normalized != "" {
		candidates["fqdn"] = append(candidates["fqdn"], mergeExactMatchCandidate{IdentifierClass: "fqdn", RawValue: *record.FQDN, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := normalizeOptionalIdentifier("hostname", record.Hostname); normalized != "" {
		candidates["hostname"] = append(candidates["hostname"], mergeExactMatchCandidate{IdentifierClass: "hostname", RawValue: *record.Hostname, NormalizedValue: normalized, FromCanonical: true})
	}
	return candidates
}

func identityCanonicalCandidates(record IdentityRecord) map[string][]mergeExactMatchCandidate {
	candidates := make(map[string][]mergeExactMatchCandidate, len(identityExactMatchPrecedence))
	if normalized := normalizeOptionalIdentifier("aad_object_id", record.AADObjectID); normalized != "" {
		candidates["aad_object_id"] = append(candidates["aad_object_id"], mergeExactMatchCandidate{IdentifierClass: "aad_object_id", RawValue: *record.AADObjectID, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := normalizeOptionalIdentifier("sid", record.SID); normalized != "" {
		candidates["sid"] = append(candidates["sid"], mergeExactMatchCandidate{IdentifierClass: "sid", RawValue: *record.SID, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := normalizeOptionalIdentifier("upn", record.UPN); normalized != "" {
		candidates["upn"] = append(candidates["upn"], mergeExactMatchCandidate{IdentifierClass: "upn", RawValue: *record.UPN, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := normalizeOptionalIdentifier("email", record.Email); normalized != "" {
		candidates["email"] = append(candidates["email"], mergeExactMatchCandidate{IdentifierClass: "email", RawValue: *record.Email, NormalizedValue: normalized, FromCanonical: true})
	}
	if normalized := normalizeOptionalIdentifier("sam_account_name", record.SamAccountName); normalized != "" {
		candidates["sam_account_name"] = append(candidates["sam_account_name"], mergeExactMatchCandidate{IdentifierClass: "sam_account_name", RawValue: *record.SamAccountName, NormalizedValue: normalized, FromCanonical: true})
	}
	return candidates
}

func hostExistingIdentifierState(record HostRecord, preserved []mergePreservedIdentifierRecord) (map[string]map[string]struct{}, map[string]bool) {
	set := make(map[string]map[string]struct{}, len(hostExactMatchPrecedence))
	filled := make(map[string]bool, len(hostExactMatchPrecedence))
	for _, identifierClass := range hostExactMatchPrecedence {
		set[identifierClass] = make(map[string]struct{})
	}
	if normalized := normalizeOptionalIdentifier("aad_device_id", record.AADDeviceID); normalized != "" {
		set["aad_device_id"][normalized] = struct{}{}
		filled["aad_device_id"] = true
	}
	if normalized := normalizeOptionalIdentifier("fqdn", record.FQDN); normalized != "" {
		set["fqdn"][normalized] = struct{}{}
		filled["fqdn"] = true
	}
	if normalized := normalizeOptionalIdentifier("hostname", record.Hostname); normalized != "" {
		set["hostname"][normalized] = struct{}{}
		filled["hostname"] = true
	}
	for _, identifier := range preserved {
		if identifier.Classification != "exact_match_reuse" {
			continue
		}
		current := set[identifier.IdentifierType]
		if current == nil {
			current = make(map[string]struct{})
			set[identifier.IdentifierType] = current
		}
		current[identifier.NormalizedValue] = struct{}{}
	}
	return set, filled
}

func identityExistingIdentifierState(record IdentityRecord, preserved []mergePreservedIdentifierRecord) (map[string]map[string]struct{}, map[string]bool) {
	set := make(map[string]map[string]struct{}, len(identityExactMatchPrecedence))
	filled := make(map[string]bool, len(identityExactMatchPrecedence))
	for _, identifierClass := range identityExactMatchPrecedence {
		set[identifierClass] = make(map[string]struct{})
	}
	if normalized := normalizeOptionalIdentifier("aad_object_id", record.AADObjectID); normalized != "" {
		set["aad_object_id"][normalized] = struct{}{}
		filled["aad_object_id"] = true
	}
	if normalized := normalizeOptionalIdentifier("sid", record.SID); normalized != "" {
		set["sid"][normalized] = struct{}{}
		filled["sid"] = true
	}
	if normalized := normalizeOptionalIdentifier("upn", record.UPN); normalized != "" {
		set["upn"][normalized] = struct{}{}
		filled["upn"] = true
	}
	if normalized := normalizeOptionalIdentifier("email", record.Email); normalized != "" {
		set["email"][normalized] = struct{}{}
		filled["email"] = true
	}
	if normalized := normalizeOptionalIdentifier("sam_account_name", record.SamAccountName); normalized != "" {
		set["sam_account_name"][normalized] = struct{}{}
		filled["sam_account_name"] = true
	}
	for _, identifier := range preserved {
		if identifier.Classification != "exact_match_reuse" {
			continue
		}
		current := set[identifier.IdentifierType]
		if current == nil {
			current = make(map[string]struct{})
			set[identifier.IdentifierType] = current
		}
		current[identifier.NormalizedValue] = struct{}{}
	}
	return set, filled
}

func aliasValuesFromRecords(records []mergeAliasRecord) []hostidentity.AliasValue {
	values := make([]hostidentity.AliasValue, 0, len(records))
	for _, record := range records {
		values = append(values, hostidentity.AliasValue{EntityAliasID: record.EntityAliasID, AliasText: record.NormalizedText})
	}
	return values
}

func aliasActionsFromRecords(records []mergeAliasRecord) []CollectionAction {
	actions := make([]CollectionAction, 0, len(records))
	for _, record := range records {
		actions = append(actions, CollectionAction{
			Op:             "add_alias",
			RawText:        record.RawText,
			NormalizedText: record.NormalizedText,
		})
	}
	return actions
}

func countProvenanceOnlyIdentifiers(values []mergePreservedIdentifierRecord) int {
	count := 0
	for _, value := range values {
		if value.Classification == "provenance_only" && value.DeletedAt == nil {
			count++
		}
	}
	return count
}

func mergeScopeKey(survivorRecordID uuid.UUID, loserRecordID uuid.UUID) string {
	return survivorRecordID.String() + ":" + loserRecordID.String()
}

func buildMergePreservedIdentifierValueFromSeed(incidentID uuid.UUID, recordID uuid.UUID, entityType string, seed identifierSeed) map[string]any {
	normalized, _ := fieldnorm.NormalizeIdentifier(seed.IdentifierType, seed.RawValue)
	return map[string]any{
		"incident_id":      incidentID.String(),
		"record_id":        recordID.String(),
		"entity_type":      entityType,
		"identifier_type":  seed.IdentifierType,
		"raw_value":        seed.RawValue,
		"normalized_value": normalized,
		"classification":   seed.Classification,
	}
}

func identifierMutationsToMergeMutations(records []mergeIdentifierMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(records))
	for _, record := range records {
		result = append(result, mergeMutation{
			TargetKind:    "entity_preserved_identifier",
			TargetID:      identifierMutationTargetID(record.Before, record.After),
			OperationKind: "create",
			BeforeValue:   record.Before,
			AfterValue:    record.After,
		})
	}
	return result
}

func aliasMutationsToMergeMutations(records []mergeAliasMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(records))
	for _, record := range records {
		result = append(result, mergeMutation{
			TargetKind:    "entity_alias",
			TargetID:      aliasMutationTargetID(record.Before, record.After),
			OperationKind: "create",
			BeforeValue:   record.Before,
			AfterValue:    record.After,
		})
	}
	return result
}

func identifierMutationTargetID(before map[string]any, after map[string]any) string {
	value := after
	if value == nil {
		value = before
	}
	if value == nil {
		return ""
	}
	return strings.Join([]string{
		"entity_preserved_identifier",
		targetIDComponent(stringMapValue(value, "record_id")),
		targetIDComponent(stringMapValue(value, "entity_type")),
		targetIDComponent(stringMapValue(value, "identifier_type")),
		targetIDComponent(stringMapValue(value, "normalized_value")),
		targetIDComponent(stringMapValue(value, "classification")),
	}, ":")
}

func aliasMutationTargetID(before map[string]any, after map[string]any) string {
	value := after
	if value == nil {
		value = before
	}
	if value == nil {
		return ""
	}
	if aliasID := stringMapValue(value, "entity_alias_id"); aliasID != "" {
		return "entity_alias:" + aliasID
	}
	return strings.Join([]string{
		"entity_alias",
		targetIDComponent(stringMapValue(value, "record_id")),
		targetIDComponent(stringMapValue(value, "entity_type")),
		targetIDComponent(stringMapValue(value, "normalized_text")),
	}, ":")
}

func stringMapValue(values map[string]any, key string) string {
	if value, ok := values[key].(string); ok {
		return value
	}
	return ""
}

func targetIDComponent(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func scanMergeAliasRecord(scanner interface{ Scan(dest ...any) error }) (mergeAliasRecord, error) {
	var (
		record    mergeAliasRecord
		deletedAt pgtype.Timestamptz
	)
	if err := scanner.Scan(
		&record.EntityAliasID,
		&record.IncidentID,
		&record.RecordID,
		&record.EntityType,
		&record.RawText,
		&record.NormalizedText,
		&record.Classification,
		&record.CreatedAt,
		&deletedAt,
	); err != nil {
		return mergeAliasRecord{}, fmt.Errorf("scan merge alias: %w", err)
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.CreatedAt = record.CreatedAt.UTC()
	return record, nil
}

func scanMergePreservedIdentifierRecord(scanner interface{ Scan(dest ...any) error }) (mergePreservedIdentifierRecord, error) {
	var (
		record    mergePreservedIdentifierRecord
		deletedAt pgtype.Timestamptz
	)
	if err := scanner.Scan(
		&record.EntityPreservedIdentifierID,
		&record.IncidentID,
		&record.RecordID,
		&record.EntityType,
		&record.IdentifierType,
		&record.RawValue,
		&record.NormalizedValue,
		&record.Classification,
		&record.CreatedAt,
		&deletedAt,
	); err != nil {
		return mergePreservedIdentifierRecord{}, fmt.Errorf("scan merge preserved identifier: %w", err)
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.CreatedAt = record.CreatedAt.UTC()
	return record, nil
}
