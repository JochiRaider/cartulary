package reportcomposition

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	strictjson "github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// PreviewReceiptReconciler validates the original owner response at a terminal
// Job transition. The preview acknowledgement is immutable; polling the Job is
// the separate interface for its current execution state.
type PreviewReceiptReconciler struct{}

func (PreviewReceiptReconciler) ReconcileFinalIdempotencyOutcomeTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey, requestHash []byte, job jobs.Resource) (bool, error) {
	invalid := errors.New("reportcomposition: incompatible preview admission receipt")
	jobID, err := uuid.Parse(job.JobID)
	if err != nil || key.RouteKey != "report_compositions.preview" || key.ActorUserID == uuid.Nil || key.ClientTxnID == "" {
		return false, invalid
	}
	source, err := PreviewSourceForRenderTx(ctx, tx, jobID)
	if err != nil {
		return false, err
	}
	if source.CreatedByUserID != key.ActorUserID || source.CompositionID.String() != key.ScopeKey ||
		job.Scope.Kind != jobs.ScopeKindIncident || job.Scope.IncidentID == nil || *job.Scope.IncidentID != source.IncidentID ||
		job.SubmittedByUserID == nil || *job.SubmittedByUserID != key.ActorUserID.String() {
		return false, invalid
	}
	receipt, err := authn.GetRouteIdempotencyTx(ctx, tx, authn.RouteIdempotencyKey{RouteKey: key.RouteKey, ActorUserID: key.ActorUserID, ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID})
	if errors.Is(err, authn.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if receipt.StatusCode != http.StatusAccepted || !bytes.Equal(receipt.RequestHash, requestHash) {
		return false, invalid
	}
	value, err := strictjson.DecodeStrict(receipt.ResponseJSON)
	if err != nil {
		return false, invalid
	}
	actual, err := canonicalJSON(value)
	if err != nil {
		return false, invalid
	}
	expected, err := canonicalJSON(previewView(source))
	if err != nil || !bytes.Equal(actual, expected) {
		return false, invalid
	}
	return true, nil
}
