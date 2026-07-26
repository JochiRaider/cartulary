package hostidentity

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

const (
	hostClipboardPasteRouteKey     = "entities.hosts.clipboard_paste"
	identityClipboardPasteRouteKey = "entities.identities.clipboard_paste"
)

type ClipboardPasteResult struct {
	Payload     map[string]any
	StatusCode  int
	Replayed    bool
	IncidentID  uuid.UUID
	ChangeSetID uuid.UUID
	ClientTxnID string
	Rows        []ClipboardPasteRowResult
}

type ClipboardPasteRowResult struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ChangedFieldKeys []string
	Row              map[string]any
}

func (s *Store) ApplyClipboardPastePlan(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, viewSchemaID string, plan tabularingest.TabularRowPlanV1, requestHash []byte, requestID string, now time.Time) (ClipboardPasteResult, error) {
	if err := plan.Validate(); err != nil {
		return ClipboardPasteResult{}, fmt.Errorf("validate entity clipboard paste plan: %w", err)
	}
	if plan.ViewSchemaID != viewSchemaID {
		return ClipboardPasteResult{}, fmt.Errorf("entity clipboard paste plan view mismatch: %s != %s", plan.ViewSchemaID, viewSchemaID)
	}
	routeKey, targetKind, err := entityClipboardRoute(viewSchemaID)
	if err != nil {
		return ClipboardPasteResult{}, err
	}
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    incidentID.String() + ":" + viewSchemaID,
		ClientTxnID: plan.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return ClipboardPasteResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return ClipboardPasteResult{}, fmt.Errorf("decode replayed entity clipboard paste payload: %w", err)
		}
		return ClipboardPasteResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: incidentID, ClientTxnID: plan.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return ClipboardPasteResult{}, fmt.Errorf("query entity clipboard paste idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ClipboardPasteResult{}, fmt.Errorf("begin entity clipboard paste transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return ClipboardPasteResult{}, err
	}
	changeSetID, err := s.ports.revisions.AppendChangeSetTx(ctx, tx, entityChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		ClientTxnID: &plan.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return ClipboardPasteResult{}, err
	}

	resultRows := make([]ClipboardPasteRowResult, 0, len(plan.Rows))
	payloadRows := make([]map[string]any, 0, len(plan.Rows))
	sequenceNo := 1
	for _, rowPlan := range plan.Rows {
		request, err := entityCreateRequestFromRowPlan(plan.ClientTxnID, rowPlan)
		if err != nil {
			return ClipboardPasteResult{}, err
		}
		var (
			recordID       uuid.UUID
			rowVersion     int64
			beforeRow      map[string]any
			afterRow       map[string]any
			operationKind  string
			aliasMutations []AliasMutationValue
		)
		switch viewSchemaID {
		case HostsViewSchemaID:
			record, before, operation, _, err := s.upsertHostTx(ctx, tx, actor, incidentID, request, now.UTC())
			if err != nil {
				return ClipboardPasteResult{}, err
			}
			if err := s.ports.projections.RefreshEntityRowTx(ctx, tx, record.RecordID, "host"); err != nil {
				return ClipboardPasteResult{}, err
			}
			recordID = record.RecordID
			rowVersion = record.RowVersion
			beforeRow = before
			afterRow = BuildHostRow(record)
			operationKind = operation
			aliasMutations = record.AliasMutations
		case IdentitiesViewSchemaID:
			record, before, operation, _, err := s.upsertIdentityTx(ctx, tx, actor, incidentID, request, now.UTC())
			if err != nil {
				return ClipboardPasteResult{}, err
			}
			if err := s.ports.projections.RefreshEntityRowTx(ctx, tx, record.RecordID, "identity"); err != nil {
				return ClipboardPasteResult{}, err
			}
			recordID = record.RecordID
			rowVersion = record.RowVersion
			beforeRow = before
			afterRow = BuildIdentityRow(record)
			operationKind = operation
			aliasMutations = record.AliasMutations
		}
		var beforeVersionID *string
		if beforeRow != nil {
			beforeVersion := rowVersion
			if !reflect.DeepEqual(beforeRow, afterRow) && rowVersion > 1 {
				beforeVersion = rowVersion - 1
			}
			value := entityVersionID(targetKind, recordID, beforeVersion)
			beforeVersionID = &value
		}
		afterVersionID := entityVersionID(targetKind, recordID, rowVersion)
		if err := s.ports.revisions.AppendMutationTx(ctx, tx, entityMutationParams{
			ChangeSetID:     changeSetID,
			SequenceNo:      sequenceNo,
			TargetKind:      targetKind,
			TargetID:        recordID.String(),
			OperationKind:   operationKind,
			BeforeVersionID: beforeVersionID,
			AfterVersionID:  &afterVersionID,
			BeforeValue:     beforeRow,
			AfterValue:      afterRow,
		}); err != nil {
			return ClipboardPasteResult{}, err
		}
		sequenceNo++
		if err := s.appendAliasCreateMutationsTx(ctx, tx, changeSetID, sequenceNo, aliasMutations); err != nil {
			return ClipboardPasteResult{}, err
		}
		sequenceNo += len(aliasMutations)
		if beforeRow == nil || !reflect.DeepEqual(beforeRow, afterRow) {
			if err := s.ports.revisions.AppendRecordRevisionTx(ctx, tx, entityRecordRevisionParams{
				ChangeSetID: changeSetID,
				RecordID:    recordID,
				RowVersion:  rowVersion,
				BeforeValue: beforeRow,
				AfterValue:  afterRow,
			}); err != nil {
				return ClipboardPasteResult{}, err
			}
		}
		changed := entityChangedFieldKeys(beforeRow, afterRow)
		resultRows = append(resultRows, ClipboardPasteRowResult{
			RecordID:         recordID,
			RowVersion:       rowVersion,
			ChangedFieldKeys: changed,
			Row:              afterRow,
		})
		payloadRows = append(payloadRows, afterRow)
	}

	payload := map[string]any{
		"view_schema_id": viewSchemaID,
		"change_set_id":  changeSetID.String(),
		"rows":           payloadRows,
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return ClipboardPasteResult{}, authn.ErrClientTxnConflict
		}
		return ClipboardPasteResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ClipboardPasteResult{}, fmt.Errorf("commit entity clipboard paste transaction: %w", err)
	}
	return ClipboardPasteResult{
		Payload:     payload,
		StatusCode:  http.StatusOK,
		IncidentID:  incidentID,
		ChangeSetID: changeSetID,
		ClientTxnID: plan.ClientTxnID,
		Rows:        resultRows,
	}, nil
}

func entityClipboardRoute(viewSchemaID string) (string, string, error) {
	switch viewSchemaID {
	case HostsViewSchemaID:
		return hostClipboardPasteRouteKey, "host", nil
	case IdentitiesViewSchemaID:
		return identityClipboardPasteRouteKey, "identity", nil
	default:
		return "", "", fmt.Errorf("unsupported entity clipboard paste view %q", viewSchemaID)
	}
}

func entityCreateRequestFromRowPlan(clientTxnID string, rowPlan tabularingest.RowPlanV1) (CreateRequest, error) {
	request := CreateRequest{
		ClientTxnID: clientTxnID,
		Values:      make(map[string]string),
		AliasAdds:   make(map[string][]CollectionAction),
	}
	for _, cell := range rowPlan.Cells {
		switch cell.FieldKey {
		case "host.aliases", "identity.aliases":
			normalized, ok := fieldnorm.NormalizeAliasText(cell.RawValue)
			if !ok {
				continue
			}
			request.AliasAdds[cell.FieldKey] = append(request.AliasAdds[cell.FieldKey], CollectionAction{
				Op:             "add_alias",
				RawText:        normalized,
				NormalizedText: normalized,
			})
		default:
			normalized, ok := fieldnorm.NormalizeLine(cell.RawValue)
			if !ok {
				continue
			}
			request.Values[cell.FieldKey] = normalized
		}
	}
	if len(request.Values) == 0 && len(request.AliasAdds) == 0 {
		return CreateRequest{}, ErrInvalidCreateRequest
	}
	return request, nil
}

func entityChangedFieldKeys(before map[string]any, after map[string]any) []string {
	cells, _ := after["cells"].(map[string]any)
	keys := make([]string, 0, len(cells))
	for fieldKey := range cells {
		if before != nil {
			beforeCells, _ := before["cells"].(map[string]any)
			if reflect.DeepEqual(beforeCells[fieldKey], cells[fieldKey]) {
				continue
			}
		}
		keys = append(keys, fieldKey)
	}
	return keys
}

func EntityClipboardPasteRequestHash(viewSchemaID string, clientTxnID string, clipboardText string, format string, startFieldKey string, columns []string) []byte {
	_ = clientTxnID
	payload := map[string]any{
		"view_schema_id":  viewSchemaID,
		"clipboard_text":  clipboardText,
		"format":          format,
		"start_field_key": startFieldKey,
		"columns":         append([]string(nil), columns...),
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}
