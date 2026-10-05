package reference_data

import (
	"context"
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrClientTxnConflict = errors.New("reference_data: client transaction conflict")

type receiptKey struct {
	RouteKey    string
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}
type receiptOutcome uint8

const (
	receiptUnknown receiptOutcome = iota
	receiptAccepted
	receiptCompleted
)

type operationReceipt struct {
	RequestHash  []byte
	ResponseJSON []byte
	Outcome      receiptOutcome
}
type actionReceipt struct {
	Version AdministrativeVersion `json:"pack_version"`
}

// The Core bridge retains the existing receipt representation. Application
// operations select a semantic outcome; only this adapter knows its persisted
// HTTP status. All writes use the caller's publication transaction.
func lockReceiptTx(ctx context.Context, tx pgx.Tx, key receiptKey) error {
	return authn.LockRouteIdempotencyTx(ctx, tx, authn.RouteIdempotencyKey(key))
}
func readReceipt(ctx context.Context, reader interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, key receiptKey) (operationReceipt, error) {
	r, err := authn.GetRouteIdempotencyTx(ctx, reader, authn.RouteIdempotencyKey(key))
	if errors.Is(err, authn.ErrNotFound) {
		return operationReceipt{}, ErrNotFound
	}
	if err != nil {
		return operationReceipt{}, err
	}
	outcome := receiptUnknown
	switch r.StatusCode {
	case http.StatusAccepted:
		outcome = receiptAccepted
	case http.StatusOK:
		outcome = receiptCompleted
	}
	return operationReceipt{RequestHash: r.RequestHash, ResponseJSON: r.ResponseJSON, Outcome: outcome}, nil
}
func writeReceiptTx(ctx context.Context, tx pgx.Tx, key receiptKey, hash []byte, outcome receiptOutcome, payload any) error {
	var status int
	switch outcome {
	case receiptAccepted:
		status = http.StatusAccepted
	case receiptCompleted:
		status = http.StatusOK
	default:
		return errors.New("reference pack: invalid receipt outcome")
	}
	return authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey(key), nil, hash, status, payload)
}
