package networkflow

import (
	"bytes"
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/jackc/pgx/v5"
)

type storedMutationReceipt struct {
	payload map[string]any
	status  int
}
type tableReceiptAdapter struct{}

func (tableReceiptAdapter) replayTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, hash []byte) (tableMutationOutcome, bool, error) {
	receipt, err := authn.GetRouteIdempotencyTx(ctx, tx, key)
	if errors.Is(err, authn.ErrNotFound) {
		return tableMutationOutcome{}, false, nil
	}
	if err != nil {
		return tableMutationOutcome{}, false, err
	}
	if !bytes.Equal(receipt.RequestHash, hash) {
		return tableMutationOutcome{}, true, authn.ErrClientTxnConflict
	}
	payload, err := decodeStoredNetworkFlowResponse(receipt.ResponseJSON)
	if err != nil {
		return tableMutationOutcome{}, true, err
	}
	return tableMutationOutcome{receipt: &storedMutationReceipt{payload: payload, status: receipt.StatusCode}}, true, nil
}
func (tableReceiptAdapter) saveTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, hash []byte, outcome tableMutationOutcome) error {
	payload, status := tableOutcomeResponse(outcome)
	return authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, hash, status, payload)
}
func tableOutcomeResponse(outcome tableMutationOutcome) (map[string]any, int) {
	if outcome.receipt != nil {
		return outcome.receipt.payload, outcome.receipt.status
	}
	return tableMutationPayload(outcome.table), 200
}
