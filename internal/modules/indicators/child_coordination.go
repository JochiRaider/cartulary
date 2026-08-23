package indicators

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	observationCreateRouteKey  = "indicators.observations.create"
	observationResolveRouteKey = "indicators.observations.resolve"
	observationDismissRouteKey = "indicators.observations.dismiss"
	observationRestoreRouteKey = "indicators.observations.restore"
	lifecycleAppendRouteKey    = "indicators.lifecycle.append"
)

func (s *Store) childReplayKey(routeKey string, actorID uuid.UUID, scopeID uuid.UUID, clientTxnID string) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{RouteKey: routeKey, ActorUserID: actorID, ScopeKey: scopeID.String(), ClientTxnID: clientTxnID}
}

func loadObservationReplay(ctx context.Context, store *authn.Store, key authn.RouteIdempotencyKey, requestHash []byte) (IndicatorObservationMutationResult, bool, error) {
	existing, err := store.GetRouteIdempotency(ctx, key)
	if errors.Is(err, authn.ErrNotFound) {
		return IndicatorObservationMutationResult{}, false, nil
	}
	if err != nil {
		return IndicatorObservationMutationResult{}, false, fmt.Errorf("query Indicator observation idempotency: %w", err)
	}
	if !bytes.Equal(existing.RequestHash, requestHash) {
		return IndicatorObservationMutationResult{}, false, authn.ErrClientTxnConflict
	}
	var result IndicatorObservationMutationResult
	if err := json.Unmarshal(existing.ResponseJSON, &result); err != nil {
		return IndicatorObservationMutationResult{}, false, fmt.Errorf("decode Indicator observation replay: %w", err)
	}
	result.Replayed = true
	return result, true, nil
}

func loadLifecycleReplay(ctx context.Context, store *authn.Store, key authn.RouteIdempotencyKey, requestHash []byte) (IndicatorLifecycleMutationResult, bool, error) {
	existing, err := store.GetRouteIdempotency(ctx, key)
	if errors.Is(err, authn.ErrNotFound) {
		return IndicatorLifecycleMutationResult{}, false, nil
	}
	if err != nil {
		return IndicatorLifecycleMutationResult{}, false, fmt.Errorf("query Indicator lifecycle idempotency: %w", err)
	}
	if !bytes.Equal(existing.RequestHash, requestHash) {
		return IndicatorLifecycleMutationResult{}, false, authn.ErrClientTxnConflict
	}
	var result IndicatorLifecycleMutationResult
	if err := json.Unmarshal(existing.ResponseJSON, &result); err != nil {
		return IndicatorLifecycleMutationResult{}, false, fmt.Errorf("decode Indicator lifecycle replay: %w", err)
	}
	result.Replayed = true
	return result, true, nil
}

func mutationPayload(value any) (map[string]any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func validateChildMutationIdentity(clientTxnID string, requestID string, requestHash []byte, baseRowVersion int64) error {
	if strings.TrimSpace(clientTxnID) == "" || strings.TrimSpace(requestID) == "" || len(requestHash) == 0 || baseRowVersion < 1 {
		return ErrInvalidCreateRequest
	}
	return nil
}

func sortedRecordIDs(values ...uuid.UUID) []uuid.UUID {
	set := map[uuid.UUID]struct{}{}
	for _, value := range values {
		if value != uuid.Nil {
			set[value] = struct{}{}
		}
	}
	result := make([]uuid.UUID, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Slice(result, func(left int, right int) bool { return result[left].String() < result[right].String() })
	return result
}

func (s *Store) lockAffectedRecordsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]records.Envelope, error) {
	return s.recordStore.LoadEnvelopesTx(ctx, tx, sortedRecordIDs(recordIDs...), true)
}

func activeIncidentEnvelope(envelopes map[uuid.UUID]records.Envelope, incidentID uuid.UUID, recordID uuid.UUID) (records.Envelope, bool) {
	envelope, found := envelopes[recordID]
	if !found || envelope.IncidentID != incidentID || envelope.DeletedAt != nil {
		return records.Envelope{}, false
	}
	return envelope, true
}

func validateObservationSourceEnvelope(envelopes map[uuid.UUID]records.Envelope, incidentID uuid.UUID, recordID uuid.UUID, expectedVersion int64) error {
	envelope, available := activeIncidentEnvelope(envelopes, incidentID, recordID)
	if !available {
		return ErrIndicatorSourceNotFound
	}
	if envelope.RowVersion != expectedVersion {
		return ErrRowVersionConflict
	}
	return nil
}

func validateResolvedIndicatorEnvelope(envelopes map[uuid.UUID]records.Envelope, incidentID uuid.UUID, recordID uuid.UUID) error {
	envelope, available := activeIncidentEnvelope(envelopes, incidentID, recordID)
	if !available || envelope.RecordType != "indicator" {
		return ErrResolvedIndicatorNotFound
	}
	return nil
}

func validateAddressedIndicatorEnvelope(envelopes map[uuid.UUID]records.Envelope, incidentID uuid.UUID, recordID uuid.UUID) error {
	envelope, available := activeIncidentEnvelope(envelopes, incidentID, recordID)
	if !available || envelope.RecordType != "indicator" {
		return ErrIndicatorNotFound
	}
	return nil
}

func validatePriorObservationDependencies(envelopes map[uuid.UUID]records.Envelope, current IndicatorObservationRecord) error {
	if _, available := activeIncidentEnvelope(envelopes, current.IncidentID, current.SourceRecordID); !available {
		return ErrIndicatorObservationNotFound
	}
	if current.ResolvedIndicatorRecordID == nil {
		return nil
	}
	envelope, available := activeIncidentEnvelope(envelopes, current.IncidentID, *current.ResolvedIndicatorRecordID)
	if !available || envelope.RecordType != "indicator" {
		return ErrIndicatorObservationNotFound
	}
	return nil
}

func validateLifecycleSupportEnvelopes(envelopes map[uuid.UUID]records.Envelope, incidentID uuid.UUID, supportRefs []uuid.UUID) error {
	for _, recordID := range supportRefs {
		if _, available := activeIncidentEnvelope(envelopes, incidentID, recordID); !available {
			return ErrInvalidCreateRequest
		}
	}
	return nil
}

func sourceSpan(text string, start int, end int) (string, error) {
	if !utf8.ValidString(text) || start < 0 || start >= end || end > len(text) {
		return "", ErrInvalidCreateRequest
	}
	if (start > 0 && !utf8.RuneStart(text[start])) || (end < len(text) && !utf8.RuneStart(text[end])) {
		return "", ErrInvalidCreateRequest
	}
	selected := text[start:end]
	if selected == "" || strings.IndexByte(selected, 0) >= 0 {
		return "", ErrInvalidCreateRequest
	}
	return selected, nil
}

func sourceOriginLocator(recordID uuid.UUID, fieldKey string, start int, end int) string {
	return fmt.Sprintf("record:%s:field:%s:bytes:%d-%d", recordID, fieldKey, start, end)
}

func (s *Store) advanceAffectedRecordsTx(ctx context.Context, tx pgx.Tx, actorID uuid.UUID, now time.Time, recordIDs []uuid.UUID) ([]AffectedRecordVersion, error) {
	result := make([]AffectedRecordVersion, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		rowVersion, err := s.recordStore.AdvanceVersionTx(ctx, tx, recordID, actorID, now)
		if err != nil {
			return nil, err
		}
		result = append(result, AffectedRecordVersion{RecordID: recordID, RowVersion: rowVersion})
	}
	return result, nil
}

func captureAffectedRecordSnapshotsTx(ctx context.Context, tx pgx.Tx, appender revisionAppendPort, recordIDs []uuid.UUID) (map[uuid.UUID]revisions.RecordSnapshot, error) {
	snapshots := make(map[uuid.UUID]revisions.RecordSnapshot, len(recordIDs))
	for _, recordID := range recordIDs {
		snapshot, err := appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
		if err != nil {
			return nil, err
		}
		snapshots[recordID] = snapshot
	}
	return snapshots, nil
}

func appendAffectedRecordRevisionsTx(
	ctx context.Context,
	tx pgx.Tx,
	appender revisionAppendPort,
	changeSetID uuid.UUID,
	versions []AffectedRecordVersion,
	beforeSnapshots map[uuid.UUID]revisions.RecordSnapshot,
	afterSnapshots map[uuid.UUID]revisions.RecordSnapshot,
	beforeRows map[uuid.UUID]map[string]any,
	afterRows map[uuid.UUID]map[string]any,
) error {
	for _, version := range versions {
		beforeSnapshot := beforeSnapshots[version.RecordID]
		afterSnapshot := afterSnapshots[version.RecordID]
		if err := appender.AppendRecordRevisionAndIntentTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID:    changeSetID,
			RecordID:       version.RecordID,
			RowVersion:     version.RowVersion,
			BeforeSnapshot: &beforeSnapshot,
			AfterSnapshot:  &afterSnapshot,
			LiveChange: revisions.LiveRecordChange{
				BeforeValue: beforeRows[version.RecordID],
				AfterValue:  afterRows[version.RecordID],
			},
		}); err != nil {
			return err
		}
	}
	return nil
}
