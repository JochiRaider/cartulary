package networkflow

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *Service) commitTableRenameRoute(ctx context.Context, incidentID uuid.UUID, tableID string, actorUserID uuid.UUID, request tableRenameRequest, requestHash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	key := tableMutationIdempotencyKey(routeKeyTablesPatch, actorUserID, incidentID, tableID, request.ClientTxnID)
	return s.commitTableMutation(ctx, key, incidentID, actorUserID, admission.RolesEditorAdmin, "editor|admin", requestHash, func(tx pgx.Tx) (TableRecord, error) {
		return s.store.renameTableTx(ctx, tx, RenameTableParams{
			IncidentID: incidentID, ActorUserID: actorUserID, TableID: tableID,
			BaseTableVersion: request.BaseTableVersion, DisplayName: request.DisplayName,
			ClientTxnID: request.ClientTxnID, RequestID: requestID, SafeDigester: s.safeDigester, Now: s.now(),
		})
	})
}

func (s *Service) commitTableSoftDeleteRoute(ctx context.Context, incidentID uuid.UUID, tableID string, actorUserID uuid.UUID, request tableSoftDeleteRequest, requestHash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	key := tableMutationIdempotencyKey(routeKeyTablesDelete, actorUserID, incidentID, tableID, request.ClientTxnID)
	return s.commitTableMutation(ctx, key, incidentID, actorUserID, admission.RolesReviewerAdmin, "reviewer|admin", requestHash, func(tx pgx.Tx) (TableRecord, error) {
		table, err := s.store.softDeleteTableTx(ctx, tx, SoftDeleteTableParams{
			IncidentID: incidentID, ActorUserID: actorUserID, TableID: tableID,
			BaseTableVersion: request.BaseTableVersion, ClientTxnID: request.ClientTxnID, RequestID: requestID, Now: s.now(),
		})
		if err != nil {
			return TableRecord{}, err
		}
		if err := s.store.InvalidateGraphViewsForTableTx(ctx, tx, incidentID, tableID, table.UpdatedAt); err != nil {
			return TableRecord{}, err
		}
		return table, nil
	})
}

// Incident serialization orders current admission, immutable receipt lookup and
// fresh table mutation. Replay never consults the current table lifecycle/version.
func (s *Service) commitTableMutation(ctx context.Context, key authn.RouteIdempotencyKey, incidentID, actorID uuid.UUID, roles admission.RoleSet, requiredRole string, requestHash []byte, mutate func(pgx.Tx) (TableRecord, error)) (map[string]any, int, *httpapi.APIError) {
	var payload map[string]any
	status := http.StatusOK
	err := withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		if _, err := s.incidentAccess.CheckTx(ctx, tx, incidentID, actorID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleOpen}); err != nil {
			return err
		}
		receipt, err := authn.GetRouteIdempotencyTx(ctx, tx, key)
		if err == nil {
			if !bytes.Equal(receipt.RequestHash, requestHash) {
				return authn.ErrClientTxnConflict
			}
			payload, err = decodeStoredNetworkFlowResponse(receipt.ResponseJSON)
			status = receipt.StatusCode
			return err
		}
		if !errors.Is(err, authn.ErrNotFound) {
			return err
		}
		table, err := mutate(tx)
		if err != nil {
			return err
		}
		payload = tableMutationPayload(table)
		return authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload)
	})
	if err != nil {
		var denied *admission.Denied
		if errors.As(err, &denied) {
			return nil, 0, savedGraphAdmissionError(err, requiredRole)
		}
		return nil, 0, tableMutationAPIError(err, key.ClientTxnID)
	}
	return payload, status, nil
}
