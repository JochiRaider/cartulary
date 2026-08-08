package jobs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// ExtensionJobAdmission is durable internal ownership and replay evidence. It
// is intentionally absent from Resource and therefore never enters the public
// common-job envelope.
type ExtensionJobAdmission struct {
	OwnerProfileID          string
	JobKind                 string
	IdempotencyIdentity     json.RawMessage
	IdempotencyRouteKey     string
	IdempotencyScopeKey     string
	NormalizedRequestSHA256 string
}

type routeScopedIdempotencyIdentity struct {
	SchemaID      string  `json:"schema_id"`
	ActorUserID   string  `json:"actor_user_id"`
	RouteIdentity string  `json:"route_identity"`
	ScopeKind     string  `json:"scope_kind"`
	ScopeID       *string `json:"scope_id"`
	ClientTxnID   string  `json:"client_txn_id"`
}

func NewExtensionJobAdmission(ownerProfileID string, jobKind string, key RouteIdempotencyKey, scope Scope, normalizedRequest []byte) (*ExtensionJobAdmission, error) {
	if ownerProfileID == "" || jobKind == "" || key.ActorUserID == uuid.Nil ||
		key.RouteKey == "" || key.ScopeKey == "" || key.ClientTxnID == "" || len(normalizedRequest) == 0 {
		return nil, fmt.Errorf("%w: incomplete extension job admission", ErrInvalidJobDefinition)
	}
	if err := validateScope(scope); err != nil {
		return nil, err
	}
	var scopeID *string
	if scope.Kind == ScopeKindIncident {
		value := scope.IncidentID.String()
		scopeID = &value
	}
	identity, err := json.Marshal(routeScopedIdempotencyIdentity{
		SchemaID:      "cartulary.route_scoped_idempotency_identity.v1",
		ActorUserID:   key.ActorUserID.String(),
		RouteIdentity: key.RouteKey + ":" + key.ScopeKey,
		ScopeKind:     scope.Kind,
		ScopeID:       scopeID,
		ClientTxnID:   key.ClientTxnID,
	})
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(normalizedRequest)
	return &ExtensionJobAdmission{
		OwnerProfileID:          ownerProfileID,
		JobKind:                 jobKind,
		IdempotencyIdentity:     identity,
		IdempotencyRouteKey:     key.RouteKey,
		IdempotencyScopeKey:     key.ScopeKey,
		NormalizedRequestSHA256: fmt.Sprintf("%x", digest[:]),
	}, nil
}
