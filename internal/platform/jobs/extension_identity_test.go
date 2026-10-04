package jobs

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func testAttributedAdmission(t *testing.T) {
	operator := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	human := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	for _, local := range []bool{false, true} {
		actor, operation, txn := human, uuid.Nil, "http-transaction"
		if local {
			actor, operation, txn = uuid.Nil, operator, operator.String()
		}
		admission, err := NewAttributedExtensionJobAdmission("test_profile", NewRouteIdempotencyKey("test.run", actor, "deployment", txn), Scope{Kind: ScopeKindDeployment}, operation, []byte(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		definition := Definition{Extension: &ExtensionPolicy{OwnerProfileID: "test_profile", IdentitySchemaID: AttributedRouteIdentitySchema}}
		params := EnqueueParams{Scope: Scope{Kind: ScopeKindDeployment}, AuthPolicy: AuthPolicyDeploymentAdmin, SubmittedByUserID: actor, OperatorOperationID: operation, Extension: admission}
		if err := validateExtensionAdmission(params, definition); err != nil {
			t.Fatal(err)
		}
		original := append([]byte(nil), admission.IdempotencyIdentity...)
		var fields map[string]any
		_ = json.Unmarshal(original, &fields)
		for name, value := range fields {
			for _, mode := range []string{"omitted", "null", "wrong_type"} {
				var mutated map[string]any
				_ = json.Unmarshal(original, &mutated)
				switch mode {
				case "omitted":
					delete(mutated, name)
				case "null":
					mutated[name] = nil
				case "wrong_type":
					mutated[name] = []any{}
				}
				admission.IdempotencyIdentity, _ = json.Marshal(mutated)
				err := validateExtensionAdmission(params, definition)
				wantValid := mode == "null" && value == nil
				if (err == nil) != wantValid {
					t.Fatalf("local=%t %s %s = %v", local, name, mode, err)
				}
			}
		}
		for _, raw := range [][]byte{
			bytes.Replace(original, []byte(`"schema_id":`), []byte(`"schema_id":"duplicate","schema_id":`), 1),
			bytes.Replace(original, []byte(`"route_identity":`), []byte(`"unknown":true,"route_identity":`), 1),
			bytes.Replace(original, []byte(AttributedRouteIdentitySchema), []byte(HumanRouteIdentitySchema), 1),
		} {
			admission.IdempotencyIdentity = raw
			if validateExtensionAdmission(params, definition) == nil {
				t.Fatal("malformed identity admitted")
			}
		}
		admission.IdempotencyIdentity = original
		definition.Extension.IdentitySchemaID = HumanRouteIdentitySchema
		if validateExtensionAdmission(params, definition) == nil {
			t.Fatal("v1 policy admitted v2")
		}
		if local {
			definition.Extension.IdentitySchemaID = AttributedRouteIdentitySchema
			params.SubmittedByUserID = human
			if validateExtensionAdmission(params, definition) == nil {
				t.Fatal("mixed actor admitted")
			}
			params.SubmittedByUserID = uuid.Nil
			params.OperatorOperationID = uuid.New()
			if validateExtensionAdmission(params, definition) == nil {
				t.Fatal("different local operation admitted")
			}
		}
	}
	if _, err := NewAttributedExtensionJobAdmission("test_profile", NewRouteIdempotencyKey("test.run", uuid.Nil, "deployment", "wrong"), Scope{Kind: ScopeKindDeployment}, operator, []byte(`{}`)); err == nil {
		t.Fatal("local transaction differs from operation")
	}
}
