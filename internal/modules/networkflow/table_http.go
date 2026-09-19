package networkflow

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
)

func tableOutcomeHTTP(outcome tableMutationOutcome, err error, role, txn string) (map[string]any, int, *httpapi.APIError) {
	if err != nil {
		var denied *admission.Denied
		if errors.As(err, &denied) {
			return nil, 0, savedGraphAdmissionError(err, role)
		}
		return nil, 0, tableMutationAPIError(err, txn)
	}
	payload, status := tableOutcomeResponse(outcome)
	return payload, status, nil
}
func (s *routeService) commitTableRenameRoute(ctx context.Context, incident uuid.UUID, table string, actor uuid.UUID, request tableRenameRequest, hash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	outcome, err := s.tables.rename(ctx, incident, table, actor, request, hash, requestID)
	return tableOutcomeHTTP(outcome, err, "editor|admin", request.ClientTxnID)
}
func (s *routeService) commitTableSoftDeleteRoute(ctx context.Context, incident uuid.UUID, table string, actor uuid.UUID, request tableSoftDeleteRequest, hash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	outcome, err := s.tables.softDelete(ctx, incident, table, actor, request, hash, requestID)
	return tableOutcomeHTTP(outcome, err, "reviewer|admin", request.ClientTxnID)
}
