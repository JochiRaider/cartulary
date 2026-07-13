package networkflow

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

func (s *Service) commitTableRenameRoute(ctx context.Context, incidentID uuid.UUID, tableID string, actorUserID uuid.UUID, request tableRenameRequest, requestHash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	idempotencyKey := tableMutationIdempotencyKey(routeKeyTablesPatch, actorUserID, incidentID, tableID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayTableMutationIfPresent(ctx, idempotencyKey, requestHash); replayed || apiErr != nil {
		return payload, status, apiErr
	}
	var table TableRecord
	var payload map[string]any
	err := s.store.transactionRun.WithinTx(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.store.renameTableTx(ctx, tx, RenameTableParams{
			IncidentID:       incidentID,
			ActorUserID:      actorUserID,
			TableID:          tableID,
			BaseTableVersion: request.BaseTableVersion,
			DisplayName:      request.DisplayName,
			ClientTxnID:      request.ClientTxnID,
			RequestID:        requestID,
			SafeDigester:     s.safeDigester,
			Now:              s.now(),
		})
		if err != nil {
			return err
		}
		payload = tableMutationPayload(table)
		return authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload)
	})
	if err != nil {
		if payload, status, replayed, apiErr := s.replayTableMutationIfPresent(ctx, idempotencyKey, requestHash); replayed || apiErr != nil {
			return payload, status, apiErr
		}
		if authn.IsUniqueViolation(err) {
			return nil, 0, httpapi.ClientTxnConflictError(request.ClientTxnID)
		}
		return nil, 0, tableMutationAPIError(err, request.ClientTxnID)
	}
	if table.TableVersion != request.BaseTableVersion {
		s.publishTableResourceChange(incidentID, table.TableID, platformws.ExtensionResourceChangeKindInvalidate, platformws.ExtensionResourceReasonRenamed)
	}
	return payload, http.StatusOK, nil
}

func (s *Service) commitTableSoftDeleteRoute(ctx context.Context, incidentID uuid.UUID, tableID string, actorUserID uuid.UUID, request tableSoftDeleteRequest, requestHash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	idempotencyKey := tableMutationIdempotencyKey(routeKeyTablesDelete, actorUserID, incidentID, tableID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayTableMutationIfPresent(ctx, idempotencyKey, requestHash); replayed || apiErr != nil {
		return payload, status, apiErr
	}
	var table TableRecord
	var payload map[string]any
	err := s.store.transactionRun.WithinTx(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.store.softDeleteTableTx(ctx, tx, SoftDeleteTableParams{
			IncidentID:       incidentID,
			ActorUserID:      actorUserID,
			TableID:          tableID,
			BaseTableVersion: request.BaseTableVersion,
			ClientTxnID:      request.ClientTxnID,
			RequestID:        requestID,
			Now:              s.now(),
		})
		if err != nil {
			return err
		}
		payload = tableMutationPayload(table)
		return authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload)
	})
	if err != nil {
		if payload, status, replayed, apiErr := s.replayTableMutationIfPresent(ctx, idempotencyKey, requestHash); replayed || apiErr != nil {
			return payload, status, apiErr
		}
		if authn.IsUniqueViolation(err) {
			return nil, 0, httpapi.ClientTxnConflictError(request.ClientTxnID)
		}
		return nil, 0, tableMutationAPIError(err, request.ClientTxnID)
	}
	s.publishTableResourceChange(incidentID, table.TableID, platformws.ExtensionResourceChangeKindRemove, platformws.ExtensionResourceReasonSoftDeleted)
	return payload, http.StatusOK, nil
}

func (s *Service) publishTableResourceChange(incidentID uuid.UUID, tableID string, changeKind string, reasonCode string) {
	if s == nil || s.hub == nil || incidentID == uuid.Nil || strings.TrimSpace(tableID) == "" {
		return
	}
	_ = s.hub.PublishExtensionResourceChange(incidentID, platformws.ExtensionResourceChangePayload{
		ExtensionProfileID: ProfileID,
		ResourceKind:       "network_flow_table",
		ResourceID:         tableID,
		ChangeKind:         changeKind,
		ReasonCode:         reasonCode,
		WorkspaceRefs: []platformws.ExtensionWorkspaceRef{{
			Kind:               "extension_workspace",
			ExtensionProfileID: ProfileID,
			WorkspaceKey:       WorkspaceKeyNetworkAnalysis,
		}},
	})
}
