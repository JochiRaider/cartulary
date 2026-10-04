package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

const HumanRouteIdentitySchema = "cartulary.route_scoped_idempotency_identity.v1"
const AttributedRouteIdentitySchema = "cartulary.route_scoped_idempotency_identity.v2"

type attributedRouteIdentity struct {
	SchemaID            string  `json:"schema_id"`
	ActorKind           string  `json:"actor_kind"`
	ActorUserID         *string `json:"actor_user_id"`
	OperatorOperationID *string `json:"operator_operation_id"`
	RouteIdentity       string  `json:"route_identity"`
	ScopeKind           string  `json:"scope_kind"`
	ScopeID             *string `json:"scope_id"`
	ClientTxnID         string  `json:"client_txn_id"`
}

// decodeRouteIdentity admits the original bytes before ordinary decoding can
// collapse duplicate keys or malformed Unicode. Schema selection is required
// owner policy, never autodetection or a legacy fallback for a changed job.
func decodeRouteIdentity(raw []byte, schema string) (attributedRouteIdentity, error) {
	var identity attributedRouteIdentity
	invalid := fmt.Errorf("%w: invalid extension idempotency identity", ErrInvalidJobDefinition)
	value, err := canonicaljson.DecodeStrict(raw)
	if err != nil {
		return identity, invalid
	}
	fields, ok := value.(map[string]any)
	names := []string{"schema_id", "actor_user_id", "route_identity", "scope_kind", "scope_id", "client_txn_id"}
	switch schema {
	case HumanRouteIdentitySchema:
	case AttributedRouteIdentitySchema:
		names = append(names, "actor_kind", "operator_operation_id")
	default:
		return identity, invalid
	}
	if !ok || len(fields) != len(names) {
		return identity, invalid
	}
	for _, name := range names {
		if _, exists := fields[name]; !exists {
			return identity, invalid
		}
	}
	if json.Unmarshal(raw, &identity) != nil || identity.SchemaID != schema || identity.RouteIdentity == "" || identity.ClientTxnID == "" {
		return identity, invalid
	}
	if schema == HumanRouteIdentitySchema {
		identity.ActorKind = "user"
	}
	switch identity.ScopeKind {
	case ScopeKindDeployment:
		if identity.ScopeID != nil {
			return identity, invalid
		}
	case ScopeKindIncident:
		if !canonicalIdentityUUID(identity.ScopeID) {
			return identity, invalid
		}
	default:
		return identity, invalid
	}
	switch identity.ActorKind {
	case "user":
		if !canonicalIdentityUUID(identity.ActorUserID) || identity.OperatorOperationID != nil {
			return identity, invalid
		}
	case "local_operator":
		if schema != AttributedRouteIdentitySchema || identity.ActorUserID != nil || !canonicalIdentityUUID(identity.OperatorOperationID) || identity.ScopeKind != ScopeKindDeployment || identity.ClientTxnID != *identity.OperatorOperationID {
			return identity, invalid
		}
	default:
		return identity, invalid
	}
	return identity, nil
}

func canonicalIdentityUUID(value *string) bool {
	if value == nil {
		return false
	}
	id, err := uuid.Parse(*value)
	return err == nil && id != uuid.Nil && id.String() == *value
}

func validIdentitySchema(schema string) bool {
	return schema == HumanRouteIdentitySchema || schema == AttributedRouteIdentitySchema
}
