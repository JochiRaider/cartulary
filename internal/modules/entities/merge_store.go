package entities

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/links"
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
	RecordID          uuid.UUID
	Scope             string
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *MergeRowVersionConflictError) Error() string {
	return fmt.Sprintf("entities: merge row version conflict for %s", e.RecordID)
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

type mergeMentionRecord struct {
	EntityMentionID  uuid.UUID
	SourceRecordID   uuid.UUID
	SourceFieldKey   string
	EntityType       string
	OriginKind       string
	OriginLocator    string
	RawText          string
	NormalizedText   string
	ResolutionStatus string
	RowVersion       int64
	ResolvedRecordID *uuid.UUID
	ResolvedByUserID *uuid.UUID
	ResolvedAt       *time.Time
	ResolutionMethod *string
}

type mergeLinkRecord struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
	Provenance   string
	Confidence   *int
	OwnerUserID  uuid.UUID
	DecidedAt    time.Time
	CreatedAt    time.Time
	DeletedAt    *time.Time
}

type mergeTagRecord struct {
	RecordTagID       uuid.UUID
	IncidentID        uuid.UUID
	RecordID          uuid.UUID
	TagName           string
	NormalizedTagName string
	CreatedByUserID   uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	DeletedByUserID   *uuid.UUID
}

type mergeAssessmentRecord struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	SubjectRecordID uuid.UUID
	SubjectType     string
	AssessmentState string
	ConfidenceScore *int
	Rationale       string
	AssessorUserID  uuid.UUID
	AssessedAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
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
	SurvivorHost        HostRecord
	SurvivorIdentity    IdentityRecord
	IdentifierInserts   []mergeIdentifierInsert
	AliasAdds           []CollectionAction
	ExactMatchClasses   []MergeExactMatchClassSummary
	IdentifierMutations []mergeIdentifierMutation
	AliasMutations      []mergeAliasMutation
}

type mergeMutation struct {
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
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
			return MergeResult{}, &MergePreconditionError{ReasonCode: "loser_not_found"}
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
	if loserMeta.RecordType != survivorMeta.RecordType {
		return MergeResult{}, &MergePreconditionError{
			ReasonCode: "record_type_mismatch",
			Details: map[string]any{
				"survivor_record_type": survivorMeta.RecordType,
				"loser_record_type":    loserMeta.RecordType,
			},
		}
	}
	if loserMeta.IncidentID != survivorMeta.IncidentID {
		return MergeResult{}, &MergePreconditionError{ReasonCode: "cross_incident_pair"}
	}

	recordIDs := []uuid.UUID{survivorRecordID, request.LoserRecordID}
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	for _, recordID := range recordIDs {
		if err := lockMergeTargetTx(ctx, tx, survivorMeta.RecordType, recordID); err != nil {
			return MergeResult{}, err
		}
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
				return MergeResult{}, &MergePreconditionError{ReasonCode: "loser_not_found"}
			}
			return MergeResult{}, err
		}
		if err := validateHostMergePair(survivorHost, loserHost); err != nil {
			return MergeResult{}, err
		}
		if survivorHost.RowVersion != request.SurvivorBaseRowVersion {
			return MergeResult{}, &MergeRowVersionConflictError{
				RecordID:          survivorHost.RecordID,
				Scope:             "survivor",
				BaseRowVersion:    request.SurvivorBaseRowVersion,
				CurrentRowVersion: survivorHost.RowVersion,
			}
		}
		if loserHost.RowVersion != request.LoserBaseRowVersion {
			return MergeResult{}, &MergeRowVersionConflictError{
				RecordID:          loserHost.RecordID,
				Scope:             "loser",
				BaseRowVersion:    request.LoserBaseRowVersion,
				CurrentRowVersion: loserHost.RowVersion,
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
				return MergeResult{}, &MergePreconditionError{ReasonCode: "loser_not_found"}
			}
			return MergeResult{}, err
		}
		if err := validateIdentityMergePair(survivorIdentity, loserIdentity); err != nil {
			return MergeResult{}, err
		}
		if survivorIdentity.RowVersion != request.SurvivorBaseRowVersion {
			return MergeResult{}, &MergeRowVersionConflictError{
				RecordID:          survivorIdentity.RecordID,
				Scope:             "survivor",
				BaseRowVersion:    request.SurvivorBaseRowVersion,
				CurrentRowVersion: survivorIdentity.RowVersion,
			}
		}
		if loserIdentity.RowVersion != request.LoserBaseRowVersion {
			return MergeResult{}, &MergeRowVersionConflictError{
				RecordID:          loserIdentity.RecordID,
				Scope:             "loser",
				BaseRowVersion:    request.LoserBaseRowVersion,
				CurrentRowVersion: loserIdentity.RowVersion,
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
		survivorBefore any
		survivorAfter  any
		loserBefore    any
		loserAfter     any
		counts         mergeCounts
		mutations      []mergeMutation
	)

	switch survivorMeta.RecordType {
	case "host":
		survivorHost.SuggestionOnlyAliases = aliasTextsFromRecords(survivorAliases)
		loserHost.SuggestionOnlyAliases = aliasTextsFromRecords(loserAliases)
		survivorBefore = buildHostMutationValue(survivorHost)
		loserBefore = buildHostMutationValue(loserHost)
	case "identity":
		survivorIdentity.SuggestionOnlyAliases = aliasTextsFromRecords(survivorAliases)
		loserIdentity.SuggestionOnlyAliases = aliasTextsFromRecords(loserAliases)
		survivorBefore = buildIdentityMutationValue(survivorIdentity)
		loserBefore = buildIdentityMutationValue(loserIdentity)
	}

	mentionMutations, mentionCounts, invalidatedRecords, err := s.repointMergedMentionsTx(ctx, tx, incidentID, survivorMeta.RecordType, survivorRecordID, request.LoserRecordID, actor.ID, now)
	if err != nil {
		return MergeResult{}, err
	}
	counts.RepointedMentions = mentionCounts
	mutations = append(mutations, mentionMutations...)

	linkMutations, repointedLinks, dedupedLinks, linkedSourceRecordIDs, err := s.repointMergedLinksTx(ctx, tx, incidentID, survivorRecordID, request.LoserRecordID, actor.ID, now)
	if err != nil {
		return MergeResult{}, err
	}
	counts.RepointedLinks = repointedLinks
	counts.DedupedLinks = dedupedLinks
	mutations = append(mutations, linkMutations...)
	for sourceRecordID, fieldKeys := range linkedSourceRecordIDs {
		current := invalidatedRecords[sourceRecordID]
		current = append(current, fieldKeys...)
		invalidatedRecords[sourceRecordID] = current
	}

	tagMutations, repointedTags, dedupedTags, err := s.repointMergedTagsTx(ctx, tx, incidentID, survivorRecordID, request.LoserRecordID, actor.ID, now)
	if err != nil {
		return MergeResult{}, err
	}
	counts.RepointedTags = repointedTags
	counts.DedupedTags = dedupedTags
	mutations = append(mutations, tagMutations...)

	assessmentMutations, repointedAssessments, err := s.repointMergedAssessmentsTx(ctx, tx, incidentID, survivorMeta.RecordType, survivorRecordID, request.LoserRecordID, actor.ID, now)
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
		nextSurvivor.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, survivorHost.RecordID, actor.ID, now.UTC())
		if err != nil {
			return MergeResult{}, err
		}
		nextSurvivor.UpdatedAt = now.UTC()
		nextSurvivor.UpdatedByUser = actor.ID
		if err := updateHostTx(ctx, tx, nextSurvivor); err != nil {
			return MergeResult{}, err
		}
		if err := upsertHostProjectionTx(ctx, tx, nextSurvivor); err != nil {
			return MergeResult{}, err
		}

		nextLoser := loserHost
		nextLoser.HostState = "merged"
		nextLoser.MergedIntoRecordID = &survivorRecordID
		nextLoser.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, loserHost.RecordID, actor.ID, now.UTC())
		if err != nil {
			return MergeResult{}, err
		}
		nextLoser.UpdatedAt = now.UTC()
		nextLoser.UpdatedByUser = actor.ID
		if err := updateHostTx(ctx, tx, nextLoser); err != nil {
			return MergeResult{}, err
		}
		if err := deleteEntityProjectionTx(ctx, tx, "host", nextLoser.RecordID); err != nil {
			return MergeResult{}, err
		}

		nextSurvivor.SuggestionOnlyAliases = append([]string(nil), aliasTextsFromActions(carryPlan.AliasAdds)...)
		nextSurvivor.SuggestionOnlyAliases = append(survivorHost.SuggestionOnlyAliases, nextSurvivor.SuggestionOnlyAliases...)
		nextSurvivor.SuggestionOnlyAliases = loadUniqueSortedAliasTexts(nextSurvivor.SuggestionOnlyAliases)
		nextLoser.SuggestionOnlyAliases = loserHost.SuggestionOnlyAliases
		survivorAfter = buildHostMutationValue(nextSurvivor)
		loserAfter = buildHostMutationValue(nextLoser)
		survivorHost = nextSurvivor
		loserHost = nextLoser
	case "identity":
		nextSurvivor := carryPlan.SurvivorIdentity
		nextSurvivor.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, survivorIdentity.RecordID, actor.ID, now.UTC())
		if err != nil {
			return MergeResult{}, err
		}
		nextSurvivor.UpdatedAt = now.UTC()
		nextSurvivor.UpdatedByUser = actor.ID
		if err := updateIdentityTx(ctx, tx, nextSurvivor); err != nil {
			return MergeResult{}, err
		}
		if err := upsertIdentityProjectionTx(ctx, tx, nextSurvivor); err != nil {
			return MergeResult{}, err
		}

		nextLoser := loserIdentity
		nextLoser.IdentityState = "merged"
		nextLoser.MergedIntoRecordID = &survivorRecordID
		nextLoser.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, loserIdentity.RecordID, actor.ID, now.UTC())
		if err != nil {
			return MergeResult{}, err
		}
		nextLoser.UpdatedAt = now.UTC()
		nextLoser.UpdatedByUser = actor.ID
		if err := updateIdentityTx(ctx, tx, nextLoser); err != nil {
			return MergeResult{}, err
		}
		if err := deleteEntityProjectionTx(ctx, tx, "identity", nextLoser.RecordID); err != nil {
			return MergeResult{}, err
		}

		nextSurvivor.SuggestionOnlyAliases = append([]string(nil), aliasTextsFromActions(carryPlan.AliasAdds)...)
		nextSurvivor.SuggestionOnlyAliases = append(survivorIdentity.SuggestionOnlyAliases, nextSurvivor.SuggestionOnlyAliases...)
		nextSurvivor.SuggestionOnlyAliases = loadUniqueSortedAliasTexts(nextSurvivor.SuggestionOnlyAliases)
		nextLoser.SuggestionOnlyAliases = loserIdentity.SuggestionOnlyAliases
		survivorAfter = buildIdentityMutationValue(nextSurvivor)
		loserAfter = buildIdentityMutationValue(nextLoser)
		survivorIdentity = nextSurvivor
		loserIdentity = nextLoser
	}

	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
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
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      "host",
			TargetID:        survivorHost.RecordID.String(),
			OperationKind:   "patch",
			BeforeVersionID: &beforeVersionID,
			AfterVersionID:  &afterVersionID,
			BeforeValue:     survivorBefore,
			AfterValue:      survivorAfter,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
		loserBeforeVersionID := entityVersionID("host", loserHost.RecordID, loserHost.RowVersion-1)
		loserAfterVersionID := entityVersionID("host", loserHost.RecordID, loserHost.RowVersion)
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      "host",
			TargetID:        loserHost.RecordID.String(),
			OperationKind:   "patch",
			BeforeVersionID: &loserBeforeVersionID,
			AfterVersionID:  &loserAfterVersionID,
			BeforeValue:     loserBefore,
			AfterValue:      loserAfter,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
		if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    survivorHost.RecordID,
			RowVersion:  survivorHost.RowVersion,
			BeforeValue: survivorBefore,
			AfterValue:  survivorAfter,
		}); err != nil {
			return MergeResult{}, err
		}
		if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    loserHost.RecordID,
			RowVersion:  loserHost.RowVersion,
			BeforeValue: loserBefore,
			AfterValue:  loserAfter,
		}); err != nil {
			return MergeResult{}, err
		}
	case "identity":
		beforeVersionID := entityVersionID("identity", survivorIdentity.RecordID, survivorIdentity.RowVersion-1)
		afterVersionID := entityVersionID("identity", survivorIdentity.RecordID, survivorIdentity.RowVersion)
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      "identity",
			TargetID:        survivorIdentity.RecordID.String(),
			OperationKind:   "patch",
			BeforeVersionID: &beforeVersionID,
			AfterVersionID:  &afterVersionID,
			BeforeValue:     survivorBefore,
			AfterValue:      survivorAfter,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
		loserBeforeVersionID := entityVersionID("identity", loserIdentity.RecordID, loserIdentity.RowVersion-1)
		loserAfterVersionID := entityVersionID("identity", loserIdentity.RecordID, loserIdentity.RowVersion)
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      "identity",
			TargetID:        loserIdentity.RecordID.String(),
			OperationKind:   "patch",
			BeforeVersionID: &loserBeforeVersionID,
			AfterVersionID:  &loserAfterVersionID,
			BeforeValue:     loserBefore,
			AfterValue:      loserAfter,
		}); err != nil {
			return MergeResult{}, err
		}
		sequenceNo++
		if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    survivorIdentity.RecordID,
			RowVersion:  survivorIdentity.RowVersion,
			BeforeValue: survivorBefore,
			AfterValue:  survivorAfter,
		}); err != nil {
			return MergeResult{}, err
		}
		if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    loserIdentity.RecordID,
			RowVersion:  loserIdentity.RowVersion,
			BeforeValue: loserBefore,
			AfterValue:  loserAfter,
		}); err != nil {
			return MergeResult{}, err
		}
	}

	for _, mutation := range mutations {
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
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

	timelineInvalidations, err := loadMergeTimelineInvalidationsTx(ctx, tx, invalidatedRecords)
	if err != nil {
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

func lockMergeTargetTx(ctx context.Context, tx pgx.Tx, recordType string, recordID uuid.UUID) error {
	var err error
	switch recordType {
	case "host":
		err = tx.QueryRow(ctx, `SELECT record_id FROM hosts WHERE record_id = $1 FOR UPDATE NOWAIT`, recordID).Scan(&recordID)
	case "identity":
		err = tx.QueryRow(ctx, `SELECT record_id FROM identities WHERE record_id = $1 FOR UPDATE NOWAIT`, recordID).Scan(&recordID)
	default:
		return &MergePreconditionError{ReasonCode: "unsupported_record_type"}
	}
	if err == nil {
		return nil
	}
	if isLockUnavailable(err) {
		return &MergeRecordLockedError{RecordID: recordID}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return &MergePreconditionError{ReasonCode: "target_not_found"}
	}
	return fmt.Errorf("lock merge target %s: %w", recordID, err)
}

func loadHostByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (HostRecord, error) {
	record, err := scanHostRecord(tx.QueryRow(ctx, `
SELECT
    h.record_id,
    h.incident_id,
    h.display_name,
    h.aad_device_id,
    h.fqdn,
    h.hostname,
    h.host_state,
    h.merged_into_record_id,
    h.entity_origin,
    h.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM hosts h
  JOIN records r
    ON r.record_id = h.record_id
 WHERE h.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return HostRecord{}, ErrMergeTargetNotFound
	}
	if err != nil {
		return HostRecord{}, fmt.Errorf("load host merge target: %w", err)
	}
	return record, nil
}

func loadIdentityByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (IdentityRecord, error) {
	record, err := scanIdentityRecord(tx.QueryRow(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.display_name,
    i.aad_object_id,
    i.sid,
    i.upn,
    i.email::text,
    i.sam_account_name,
    i.identity_state,
    i.merged_into_record_id,
    i.entity_origin,
    i.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM identities i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
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

func (s *Store) repointMergedMentionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordType string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, actorUserID uuid.UUID, now time.Time) ([]mergeMutation, int, map[uuid.UUID][]string, error) {
	rows, err := tx.Query(ctx, `
SELECT
    entity_mention_id,
    source_record_id,
    source_field_key,
    entity_type,
    origin_kind,
    origin_locator,
    raw_text,
    normalized_text,
    resolution_status,
    row_version,
    resolved_record_id,
    resolved_by_user_id,
    resolved_at,
    resolution_method
  FROM entity_mentions
 WHERE entity_type = $1
   AND resolution_status = 'resolved'
   AND resolved_record_id = $2
 ORDER BY source_record_id ASC, source_field_key ASC, entity_mention_id ASC
 FOR UPDATE
`, recordType, loserRecordID)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("load merged mentions: %w", err)
	}
	defer rows.Close()

	mutations := make([]mergeMutation, 0)
	invalidations := make(map[uuid.UUID][]string)
	records := make([]mergeMentionRecord, 0)
	for rows.Next() {
		record, err := scanMergeMentionRecord(rows)
		if err != nil {
			return nil, 0, nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, nil, fmt.Errorf("iterate merged mentions: %w", err)
	}
	rows.Close()
	count := 0
	for _, record := range records {
		before := buildMergeMentionValue(record)
		beforeVersion := mentionVersionID(record.EntityMentionID, record.RowVersion)
		record.RowVersion++
		record.ResolvedRecordID = &survivorRecordID
		if _, err := tx.Exec(ctx, `
UPDATE entity_mentions
   SET resolved_record_id = $2,
       row_version = $3
 WHERE entity_mention_id = $1
`, record.EntityMentionID, survivorRecordID, record.RowVersion); err != nil {
			return nil, 0, nil, fmt.Errorf("repoint merged mention: %w", err)
		}
		afterVersion := mentionVersionID(record.EntityMentionID, record.RowVersion)
		mutations = append(mutations, mergeMutation{
			TargetKind:      "entity_mention",
			TargetID:        record.EntityMentionID.String(),
			OperationKind:   "patch",
			BeforeVersionID: &beforeVersion,
			AfterVersionID:  &afterVersion,
			BeforeValue:     before,
			AfterValue:      buildMergeMentionValue(record),
		})
		current := invalidations[record.SourceRecordID]
		current = append(current, record.SourceFieldKey)
		invalidations[record.SourceRecordID] = current
		count++
	}
	return mutations, count, invalidations, nil
}

func (s *Store) repointMergedLinksTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, actorUserID uuid.UUID, now time.Time) ([]mergeMutation, int, int, map[uuid.UUID][]string, error) {
	rows, err := tx.Query(ctx, `
SELECT
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    decided_at,
    created_at,
    deleted_at
  FROM record_links
 WHERE incident_id = $1
   AND deleted_at IS NULL
   AND dst_record_id = $2
 ORDER BY src_record_id ASC, link_type ASC, record_link_id ASC
 FOR UPDATE
`, incidentID, loserRecordID)
	if err != nil {
		return nil, 0, 0, nil, fmt.Errorf("load merged links: %w", err)
	}
	defer rows.Close()

	mutations := make([]mergeMutation, 0)
	invalidations := make(map[uuid.UUID][]string)
	records := make([]mergeLinkRecord, 0)
	for rows.Next() {
		record, err := scanMergeLinkRecord(rows)
		if err != nil {
			return nil, 0, 0, nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, 0, nil, fmt.Errorf("iterate merged links: %w", err)
	}
	rows.Close()
	repointed := 0
	deduped := 0
	for _, record := range records {
		if record.DstRecordID != loserRecordID {
			continue
		}
		before := buildMergeLinkValue(record)
		existing, err := s.linkStore.GetActiveLinkTx(ctx, tx, incidentID, record.SrcRecordID, survivorRecordID, record.LinkType)
		switch {
		case errors.Is(err, links.ErrRecordLinkNotFound):
			tombstoned, err := s.linkStore.TombstoneLinkTx(ctx, tx, record.RecordLinkID, actorUserID, now.UTC())
			if err != nil {
				return nil, 0, 0, nil, fmt.Errorf("tombstone merged link before repoint: %w", err)
			}
			mutations = append(mutations, mergeMutation{
				TargetKind:    "record_link",
				TargetID:      tombstoned.RecordLinkID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeLinkValue(mergeLinkRecordFromRecordLink(tombstoned)),
			})
			created, inserted, err := s.linkStore.UpsertLinkTx(ctx, tx, incidentID, record.SrcRecordID, survivorRecordID, record.LinkType, record.Provenance, record.Confidence, record.OwnerUserID, record.DecidedAt)
			if err != nil {
				return nil, 0, 0, nil, fmt.Errorf("create repointed merged link: %w", err)
			}
			if !inserted {
				return nil, 0, 0, nil, fmt.Errorf("create repointed merged link: expected insert for %s", record.RecordLinkID)
			}
			mutations = append(mutations, mergeMutation{
				TargetKind:    "record_link",
				TargetID:      created.RecordLinkID.String(),
				OperationKind: "create",
				AfterValue:    buildMergeLinkValue(mergeLinkRecordFromRecordLink(created)),
			})
			repointed++
		case err != nil:
			return nil, 0, 0, nil, err
		default:
			tombstoned, err := s.linkStore.TombstoneLinkTx(ctx, tx, record.RecordLinkID, actorUserID, now.UTC())
			if err != nil {
				return nil, 0, 0, nil, err
			}
			_ = existing
			mutations = append(mutations, mergeMutation{
				TargetKind:    "record_link",
				TargetID:      record.RecordLinkID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeLinkValue(mergeLinkRecordFromRecordLink(tombstoned)),
			})
			deduped++
		}
		fieldKey := mergeLinkTypeFieldKey(record.LinkType)
		if fieldKey != "" {
			current := invalidations[record.SrcRecordID]
			current = append(current, fieldKey)
			invalidations[record.SrcRecordID] = current
		}
	}
	return mutations, repointed, deduped, invalidations, nil
}

func (s *Store) repointMergedTagsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, actorUserID uuid.UUID, now time.Time) ([]mergeMutation, int, int, error) {
	rows, err := tx.Query(ctx, `
SELECT
    record_tag_id,
    incident_id,
    record_id,
    tag_name,
    normalized_tag_name,
    created_by_user_id,
    created_at,
    updated_at,
    deleted_at,
    deleted_by_user_id
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND deleted_at IS NULL
 ORDER BY normalized_tag_name ASC, record_tag_id ASC
 FOR UPDATE
`, incidentID, loserRecordID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("load merged tags: %w", err)
	}
	defer rows.Close()

	mutations := make([]mergeMutation, 0)
	records := make([]mergeTagRecord, 0)
	for rows.Next() {
		record, err := scanMergeTagRecord(rows)
		if err != nil {
			return nil, 0, 0, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("iterate merged tags: %w", err)
	}
	rows.Close()
	repointed := 0
	deduped := 0
	for _, record := range records {
		before := buildMergeTagValue(record)
		var existingID uuid.UUID
		err = tx.QueryRow(ctx, `
SELECT record_tag_id
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = $3
   AND deleted_at IS NULL
 LIMIT 1
`, incidentID, survivorRecordID, record.NormalizedTagName).Scan(&existingID)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			if _, err := tx.Exec(ctx, `
UPDATE record_tags
   SET record_id = $2,
       updated_at = $3
 WHERE record_tag_id = $1
`, record.RecordTagID, survivorRecordID, now.UTC()); err != nil {
				return nil, 0, 0, fmt.Errorf("repoint merged tag: %w", err)
			}
			record.RecordID = survivorRecordID
			record.UpdatedAt = now.UTC()
			mutations = append(mutations, mergeMutation{
				TargetKind:    "record_tag",
				TargetID:      record.RecordTagID.String(),
				OperationKind: "patch",
				BeforeValue:   before,
				AfterValue:    buildMergeTagValue(record),
			})
			repointed++
		case err != nil:
			return nil, 0, 0, fmt.Errorf("lookup survivor tag collision: %w", err)
		default:
			if _, err := tx.Exec(ctx, `
UPDATE record_tags
   SET deleted_at = COALESCE(deleted_at, $2),
       deleted_by_user_id = COALESCE(deleted_by_user_id, $3),
       updated_at = $2
 WHERE record_tag_id = $1
`, record.RecordTagID, now.UTC(), actorUserID); err != nil {
				return nil, 0, 0, fmt.Errorf("dedupe merged tag: %w", err)
			}
			record.DeletedAt = timePointer(now.UTC())
			record.DeletedByUserID = &actorUserID
			record.UpdatedAt = now.UTC()
			mutations = append(mutations, mergeMutation{
				TargetKind:    "record_tag",
				TargetID:      record.RecordTagID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeTagValue(record),
			})
			deduped++
			_ = existingID
		}
	}
	return mutations, repointed, deduped, nil
}

func (s *Store) repointMergedAssessmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordType string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, actorUserID uuid.UUID, now time.Time) ([]mergeMutation, int, error) {
	rows, err := tx.Query(ctx, `
SELECT
    record_id,
    incident_id,
    subject_record_id,
    subject_type,
    assessment_state,
    confidence_score,
    rationale,
    assessor_user_id,
    assessed_at,
    created_at,
    updated_at,
    deleted_at,
    deleted_by_user_id
  FROM assessments
 WHERE incident_id = $1
   AND subject_type = $2
   AND subject_record_id = $3
   AND deleted_at IS NULL
 ORDER BY assessed_at ASC, record_id ASC
 FOR UPDATE
`, incidentID, recordType, loserRecordID)
	if err != nil {
		return nil, 0, fmt.Errorf("load merged assessments: %w", err)
	}
	defer rows.Close()

	mutations := make([]mergeMutation, 0)
	records := make([]mergeAssessmentRecord, 0)
	for rows.Next() {
		record, err := scanMergeAssessmentRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate merged assessments: %w", err)
	}
	rows.Close()
	count := 0
	for _, record := range records {
		before := buildMergeAssessmentValue(record)
		if _, err := tx.Exec(ctx, `
UPDATE assessments
   SET subject_record_id = $2,
       updated_at = $3
 WHERE record_id = $1
`, record.RecordID, survivorRecordID, now.UTC()); err != nil {
			return nil, 0, fmt.Errorf("repoint merged assessment: %w", err)
		}
		record.SubjectRecordID = survivorRecordID
		record.UpdatedAt = now.UTC()
		if err := s.projectionStore.RefreshAssessmentTx(ctx, tx, record.RecordID); err != nil {
			return nil, 0, err
		}
		mutations = append(mutations, mergeMutation{
			TargetKind:    "assessment",
			TargetID:      record.RecordID.String(),
			OperationKind: "patch",
			BeforeValue:   before,
			AfterValue:    buildMergeAssessmentValue(record),
		})
		count++
	}
	_ = actorUserID
	return mutations, count, nil
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
		if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, survivor.RecordID, "host", []identifierSeed{candidate.Seed}, actorUserID, now); err != nil {
			return mergeCarryPlan{}, err
		}
		plan.IdentifierMutations = append(plan.IdentifierMutations, mergeIdentifierMutation{
			After: buildMergePreservedIdentifierValueFromSeed(incidentID, survivor.RecordID, "host", candidate.Seed),
		})
	}
	plan.SurvivorHost = next
	if len(loserAliases) > 0 {
		actions := filterAliasActions(aliasActionsFromRecords(loserAliases), survivorAliases)
		if _, err := syncEntityAliasesTx(ctx, tx, incidentID, survivor.RecordID, "host", actions, actorUserID, now); err != nil {
			return mergeCarryPlan{}, err
		}
		plan.AliasAdds = append(plan.AliasAdds, actions...)
		for _, action := range actions {
			plan.AliasMutations = append(plan.AliasMutations, mergeAliasMutation{
				After: buildMergeAliasValueFromAction(incidentID, survivor.RecordID, "host", action),
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
		if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, survivor.RecordID, "identity", []identifierSeed{candidate.Seed}, actorUserID, now); err != nil {
			return mergeCarryPlan{}, err
		}
		plan.IdentifierMutations = append(plan.IdentifierMutations, mergeIdentifierMutation{
			After: buildMergePreservedIdentifierValueFromSeed(incidentID, survivor.RecordID, "identity", candidate.Seed),
		})
	}
	plan.SurvivorIdentity = next
	if len(loserAliases) > 0 {
		actions := filterAliasActions(aliasActionsFromRecords(loserAliases), survivorAliases)
		if _, err := syncEntityAliasesTx(ctx, tx, incidentID, survivor.RecordID, "identity", actions, actorUserID, now); err != nil {
			return mergeCarryPlan{}, err
		}
		plan.AliasAdds = append(plan.AliasAdds, actions...)
		for _, action := range actions {
			plan.AliasMutations = append(plan.AliasMutations, mergeAliasMutation{
				After: buildMergeAliasValueFromAction(incidentID, survivor.RecordID, "identity", action),
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
					ReasonCode: "exact_match_conflict",
					Details: map[string]any{
						"record_type":           entityType,
						"identifier_class":      identifierClass,
						"normalized_value":      candidate.NormalizedValue,
						"conflicting_record_id": conflictingRecordID.String(),
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
			if hostCanonicalNormalized(record, identifierClass) == normalizedValue {
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
			if identityCanonicalNormalized(record, identifierClass) == normalizedValue {
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

func deleteEntityProjectionTx(ctx context.Context, tx pgx.Tx, entityType string, recordID uuid.UUID) error {
	var query string
	switch entityType {
	case "host":
		query = `DELETE FROM host_grid_projection WHERE record_id = $1`
	case "identity":
		query = `DELETE FROM identity_grid_projection WHERE record_id = $1`
	default:
		return fmt.Errorf("delete entity projection: unsupported entity type %q", entityType)
	}
	if _, err := tx.Exec(ctx, query, recordID); err != nil {
		return fmt.Errorf("delete entity projection: %w", err)
	}
	return nil
}

func loadMergeTimelineInvalidationsTx(ctx context.Context, tx pgx.Tx, fieldKeysByRecord map[uuid.UUID][]string) ([]MergeTimelineInvalidation, error) {
	if len(fieldKeysByRecord) == 0 {
		return nil, nil
	}
	recordIDs := make([]uuid.UUID, 0, len(fieldKeysByRecord))
	for recordID := range fieldKeysByRecord {
		recordIDs = append(recordIDs, recordID)
	}
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	result := make([]MergeTimelineInvalidation, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		var rowVersion int64
		if err := tx.QueryRow(ctx, `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&rowVersion); err != nil {
			return nil, fmt.Errorf("load record invalidation row_version: %w", err)
		}
		fieldKeys := append([]string(nil), fieldKeysByRecord[recordID]...)
		fieldKeys = loadUniqueSortedAliasTexts(fieldKeys)
		result = append(result, MergeTimelineInvalidation{
			RecordID:         recordID,
			RowVersion:       rowVersion,
			ChangedFieldKeys: fieldKeys,
		})
	}
	return result, nil
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

func aliasTextsFromRecords(records []mergeAliasRecord) []string {
	values := make([]string, 0, len(records))
	for _, record := range records {
		values = append(values, record.RawText)
	}
	return values
}

func aliasTextsFromActions(actions []CollectionAction) []string {
	values := make([]string, 0, len(actions))
	for _, action := range actions {
		values = append(values, action.RawText)
	}
	return values
}

func aliasActionsFromRecords(records []mergeAliasRecord) []CollectionAction {
	actions := make([]CollectionAction, 0, len(records))
	for _, record := range records {
		actions = append(actions, CollectionAction{
			Op:             "add_token",
			RawText:        record.RawText,
			NormalizedText: record.NormalizedText,
		})
	}
	return actions
}

func filterAliasActions(actions []CollectionAction, existing []mergeAliasRecord) []CollectionAction {
	if len(actions) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(existing))
	for _, record := range existing {
		seen[record.NormalizedText] = struct{}{}
	}
	filtered := make([]CollectionAction, 0, len(actions))
	for _, action := range actions {
		if _, ok := seen[action.NormalizedText]; ok {
			continue
		}
		seen[action.NormalizedText] = struct{}{}
		filtered = append(filtered, action)
	}
	return filtered
}

func loadUniqueSortedAliasTexts(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	slices.Sort(result)
	return result
}

func mergeScopeKey(survivorRecordID uuid.UUID, loserRecordID uuid.UUID) string {
	return survivorRecordID.String() + ":" + loserRecordID.String()
}

func isLockUnavailable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}

func mergeLinkTypeFieldKey(linkType string) string {
	switch linkType {
	case "observed_on_host":
		return "timeline.host_refs"
	case "observed_as_identity":
		return "timeline.identity_refs"
	default:
		return ""
	}
}

func mergeLinkRecordFromRecordLink(record links.RecordLink) mergeLinkRecord {
	return mergeLinkRecord{
		RecordLinkID: record.RecordLinkID,
		IncidentID:   record.IncidentID,
		SrcRecordID:  record.SrcRecordID,
		DstRecordID:  record.DstRecordID,
		LinkType:     record.LinkType,
		Provenance:   record.Provenance,
		Confidence:   record.Confidence,
		OwnerUserID:  record.OwnerUserID,
		DecidedAt:    record.DecidedAt,
		CreatedAt:    record.CreatedAt,
		DeletedAt:    record.DeletedAt,
	}
}

func timePointer(value time.Time) *time.Time {
	copy := value.UTC()
	return &copy
}

func buildHostMutationValue(record HostRecord) map[string]any {
	return map[string]any{
		"record_id":               record.RecordID.String(),
		"incident_id":             record.IncidentID.String(),
		"display_name":            record.DisplayName,
		"aad_device_id":           derefString(record.AADDeviceID),
		"fqdn":                    derefString(record.FQDN),
		"hostname":                derefString(record.Hostname),
		"host_state":              record.HostState,
		"merged_into_record_id":   formatUUIDPointer(record.MergedIntoRecordID),
		"entity_origin":           record.EntityOrigin,
		"seed_entity_mention_id":  formatUUIDPointer(record.SeedMentionID),
		"row_version":             record.RowVersion,
		"suggestion_only_aliases": append([]string(nil), record.SuggestionOnlyAliases...),
	}
}

func buildIdentityMutationValue(record IdentityRecord) map[string]any {
	return map[string]any{
		"record_id":               record.RecordID.String(),
		"incident_id":             record.IncidentID.String(),
		"display_name":            record.DisplayName,
		"aad_object_id":           derefString(record.AADObjectID),
		"sid":                     derefString(record.SID),
		"upn":                     derefString(record.UPN),
		"email":                   derefString(record.Email),
		"sam_account_name":        derefString(record.SamAccountName),
		"identity_state":          record.IdentityState,
		"merged_into_record_id":   formatUUIDPointer(record.MergedIntoRecordID),
		"entity_origin":           record.EntityOrigin,
		"seed_entity_mention_id":  formatUUIDPointer(record.SeedMentionID),
		"row_version":             record.RowVersion,
		"suggestion_only_aliases": append([]string(nil), record.SuggestionOnlyAliases...),
	}
}

func buildMergeMentionValue(record mergeMentionRecord) map[string]any {
	return map[string]any{
		"entity_mention_id":   record.EntityMentionID.String(),
		"source_record_id":    record.SourceRecordID.String(),
		"source_field_key":    record.SourceFieldKey,
		"entity_type":         record.EntityType,
		"origin_kind":         record.OriginKind,
		"origin_locator":      record.OriginLocator,
		"raw_text":            record.RawText,
		"normalized_text":     record.NormalizedText,
		"resolution_status":   record.ResolutionStatus,
		"row_version":         record.RowVersion,
		"resolved_record_id":  formatUUIDPointer(record.ResolvedRecordID),
		"resolved_by_user_id": formatUUIDPointer(record.ResolvedByUserID),
		"resolved_at":         formatTimestampPointer(record.ResolvedAt),
		"resolution_method":   derefString(record.ResolutionMethod),
	}
}

func buildMergeLinkValue(record mergeLinkRecord) map[string]any {
	return map[string]any{
		"record_link_id": record.RecordLinkID.String(),
		"incident_id":    record.IncidentID.String(),
		"src_record_id":  record.SrcRecordID.String(),
		"dst_record_id":  record.DstRecordID.String(),
		"link_type":      record.LinkType,
		"provenance":     record.Provenance,
		"confidence":     record.Confidence,
		"deleted_at":     formatTimestampPointer(record.DeletedAt),
	}
}

func buildMergeTagValue(record mergeTagRecord) map[string]any {
	return map[string]any{
		"record_tag_id":       record.RecordTagID.String(),
		"incident_id":         record.IncidentID.String(),
		"record_id":           record.RecordID.String(),
		"tag_name":            record.TagName,
		"normalized_tag_name": record.NormalizedTagName,
		"deleted_at":          formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id":  formatUUIDPointer(record.DeletedByUserID),
	}
}

func buildMergeAssessmentValue(record mergeAssessmentRecord) map[string]any {
	return map[string]any{
		"record_id":          record.RecordID.String(),
		"incident_id":        record.IncidentID.String(),
		"subject_record_id":  record.SubjectRecordID.String(),
		"subject_type":       record.SubjectType,
		"assessment_state":   record.AssessmentState,
		"confidence_score":   record.ConfidenceScore,
		"rationale":          record.Rationale,
		"assessor_user_id":   record.AssessorUserID.String(),
		"assessed_at":        formatTimestamp(record.AssessedAt),
		"deleted_at":         formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id": formatUUIDPointer(record.DeletedByUserID),
	}
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

func buildMergeAliasValueFromAction(incidentID uuid.UUID, recordID uuid.UUID, entityType string, action CollectionAction) map[string]any {
	return map[string]any{
		"incident_id":     incidentID.String(),
		"record_id":       recordID.String(),
		"entity_type":     entityType,
		"raw_text":        action.RawText,
		"normalized_text": action.NormalizedText,
		"classification":  "suggestion_only",
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
	if after != nil {
		if value, ok := after["normalized_value"].(string); ok {
			return value
		}
	}
	if before != nil {
		if value, ok := before["normalized_value"].(string); ok {
			return value
		}
	}
	return ""
}

func aliasMutationTargetID(before map[string]any, after map[string]any) string {
	if after != nil {
		if value, ok := after["normalized_text"].(string); ok {
			return value
		}
	}
	if before != nil {
		if value, ok := before["normalized_text"].(string); ok {
			return value
		}
	}
	return ""
}

func scanMergeMentionRecord(scanner interface{ Scan(dest ...any) error }) (mergeMentionRecord, error) {
	var (
		record           mergeMentionRecord
		resolvedRecordID pgtype.UUID
		resolvedByUserID pgtype.UUID
		resolvedAt       pgtype.Timestamptz
		resolutionMethod pgtype.Text
	)
	if err := scanner.Scan(
		&record.EntityMentionID,
		&record.SourceRecordID,
		&record.SourceFieldKey,
		&record.EntityType,
		&record.OriginKind,
		&record.OriginLocator,
		&record.RawText,
		&record.NormalizedText,
		&record.ResolutionStatus,
		&record.RowVersion,
		&resolvedRecordID,
		&resolvedByUserID,
		&resolvedAt,
		&resolutionMethod,
	); err != nil {
		return mergeMentionRecord{}, fmt.Errorf("scan merged mention: %w", err)
	}
	record.ResolvedRecordID = uuidPointerFromPG(resolvedRecordID)
	record.ResolvedByUserID = uuidPointerFromPG(resolvedByUserID)
	if resolvedAt.Valid {
		value := resolvedAt.Time.UTC()
		record.ResolvedAt = &value
	}
	if resolutionMethod.Valid {
		value := resolutionMethod.String
		record.ResolutionMethod = &value
	}
	return record, nil
}

func scanMergeLinkRecord(scanner interface{ Scan(dest ...any) error }) (mergeLinkRecord, error) {
	var (
		record     mergeLinkRecord
		confidence pgtype.Int4
		deletedAt  pgtype.Timestamptz
	)
	if err := scanner.Scan(
		&record.RecordLinkID,
		&record.IncidentID,
		&record.SrcRecordID,
		&record.DstRecordID,
		&record.LinkType,
		&record.Provenance,
		&confidence,
		&record.OwnerUserID,
		&record.DecidedAt,
		&record.CreatedAt,
		&deletedAt,
	); err != nil {
		return mergeLinkRecord{}, fmt.Errorf("scan merged link: %w", err)
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		record.Confidence = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.DecidedAt = record.DecidedAt.UTC()
	record.CreatedAt = record.CreatedAt.UTC()
	return record, nil
}

func scanMergeTagRecord(scanner interface{ Scan(dest ...any) error }) (mergeTagRecord, error) {
	var (
		record          mergeTagRecord
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := scanner.Scan(
		&record.RecordTagID,
		&record.IncidentID,
		&record.RecordID,
		&record.TagName,
		&record.NormalizedTagName,
		&record.CreatedByUserID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return mergeTagRecord{}, fmt.Errorf("scan merged tag: %w", err)
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.DeletedByUserID = uuidPointerFromPG(deletedByUserID)
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()
	return record, nil
}

func scanMergeAssessmentRecord(scanner interface{ Scan(dest ...any) error }) (mergeAssessmentRecord, error) {
	var (
		record          mergeAssessmentRecord
		confidence      pgtype.Int4
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.SubjectRecordID,
		&record.SubjectType,
		&record.AssessmentState,
		&confidence,
		&record.Rationale,
		&record.AssessorUserID,
		&record.AssessedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return mergeAssessmentRecord{}, fmt.Errorf("scan merged assessment: %w", err)
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		record.ConfidenceScore = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.DeletedByUserID = uuidPointerFromPG(deletedByUserID)
	record.AssessedAt = record.AssessedAt.UTC()
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()
	return record, nil
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
