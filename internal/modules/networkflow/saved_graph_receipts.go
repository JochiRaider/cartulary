package networkflow

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/jackc/pgx/v5"
)

type savedGraphReceiptReader interface {
	GetRouteIdempotency(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error)
}

type savedGraphReceiptAdapter struct{ reader savedGraphReceiptReader }

// Receipt access/decoding failures are internal even when their causes happen to
// be transaction-conflict errors. They must never become a caller conflict.
type savedGraphReceiptFailure struct{ cause error }

func (e *savedGraphReceiptFailure) Error() string { return "saved graph receipt unavailable" }
func (e *savedGraphReceiptFailure) Unwrap() error { return e.cause }

func (a savedGraphReceiptAdapter) replay(ctx context.Context, key authn.RouteIdempotencyKey, hash []byte) (savedGraphOutcome, bool, error) {
	existing, err := a.reader.GetRouteIdempotency(ctx, key)
	if errors.Is(err, authn.ErrNotFound) {
		return savedGraphOutcome{}, false, nil
	}
	if err != nil {
		return savedGraphOutcome{}, false, &savedGraphReceiptFailure{err}
	}
	if !bytes.Equal(existing.RequestHash, hash) {
		return savedGraphOutcome{}, true, authn.ErrClientTxnConflict
	}
	payload, err := decodeStoredNetworkFlowResponse(existing.ResponseJSON)
	if err != nil {
		return savedGraphOutcome{}, true, &savedGraphReceiptFailure{err}
	}
	if err := validateGraphViewReceipt(key, existing.StatusCode, payload, nil); err != nil {
		return savedGraphOutcome{}, true, &savedGraphReceiptFailure{err}
	}
	var kind savedGraphOutcomeKind
	switch existing.StatusCode {
	case http.StatusAccepted:
		kind = savedGraphAccepted
	case http.StatusOK:
		kind = savedGraphChanged
	case http.StatusNoContent:
		kind = savedGraphRetired
	default:
		return savedGraphOutcome{}, true, &savedGraphReceiptFailure{fmt.Errorf("invalid receipt outcome")}
	}
	return savedGraphOutcome{kind: kind, receipt: payload}, true, nil
}

func (a savedGraphReceiptAdapter) saveTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, hash []byte, outcome savedGraphOutcome) error {
	return authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, hash, savedGraphOutcomeStatus(outcome.kind), savedGraphOutcomePayload(outcome))
}

func savedGraphOutcomeStatus(kind savedGraphOutcomeKind) int {
	switch kind {
	case savedGraphAccepted:
		return http.StatusAccepted
	case savedGraphChanged:
		return http.StatusOK
	case savedGraphRetired:
		return http.StatusNoContent
	default:
		return http.StatusInternalServerError
	}
}

func savedGraphOutcomePayload(outcome savedGraphOutcome) map[string]any {
	if outcome.receipt != nil {
		return outcome.receipt
	}
	switch outcome.kind {
	case savedGraphAccepted:
		return graphViewAcceptedPayload(outcome.declaration, outcome.jobID)
	case savedGraphChanged:
		return graphViewMutationPayload(outcome.declaration)
	case savedGraphRetired:
		return map[string]any{}
	default:
		return nil
	}
}
