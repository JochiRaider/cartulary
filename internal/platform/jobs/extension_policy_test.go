package jobs

import (
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
	"github.com/google/uuid"
	"strings"
	"testing"
)

func testJobPolicyMatrix(t *testing.T) {
	actor := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	for _, proof := range []bool{true, false} {
		for _, identity := range []string{"required", "none"} {
			for _, cancellation := range []string{"precommit_observable", "not_cancelable"} {
				t.Run(fmt.Sprintf("proof=%t_identity=%s_cancel=%s", proof, identity, cancellation), func(t *testing.T) {
					policy := &ExtensionPolicy{OwnerProfileID: "test_profile", OperationKind: "test_profile.run", ContractSHA256: strings.Repeat("a", 64), ProofRequired: proof, MaxProofBytes: 4096, IdempotencyPolicy: identity, CancellationPolicy: cancellation}
					admission := &ExtensionJobAdmission{OwnerProfileID: "test_profile", IdempotencyIdentity: json.RawMessage(`null`), NormalizedRequestSHA256: strings.Repeat("b", 64)}
					if identity == "required" {
						policy.IdentitySchemaID = HumanRouteIdentitySchema
						var err error
						admission, err = NewExtensionJobAdmission("test_profile", NewRouteIdempotencyKey("test.run", actor, "deployment", "test"), Scope{Kind: ScopeKindDeployment}, []byte(`{}`))
						if err != nil {
							t.Fatal(err)
						}
					}
					definition := Definition{JobKind: "test_profile.run_v1", ProgressUnitID: "test_profile.run.attempt.v1", HandlerName: "test_profile.worker_v1", Terminal: jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"done"}}, Extension: policy}
					if _, err := NewCatalog([]Definition{definition}); err != nil {
						t.Fatalf("policy rejected: %v", err)
					}
					params := EnqueueParams{JobKind: definition.JobKind, Scope: Scope{Kind: ScopeKindDeployment}, SubmittedByUserID: actor, Extension: admission, Cancelable: cancellation == "precommit_observable"}
					if err := validateExtensionAdmission(params, definition); err != nil {
						t.Fatalf("policy admission: %v", err)
					}
					if cancellation == "not_cancelable" {
						params.Cancelable = true
						if validateExtensionAdmission(params, definition) == nil {
							t.Fatal("caller overrode cancellation policy")
						}
						params.Cancelable = false
					}
					if identity == "none" {
						params.SubmittedByUserID = uuid.Nil
						params.OperatorOperationID = actor
						params.AuthPolicy = AuthPolicyDeploymentAdmin
						if err := validateExtensionAdmission(params, definition); err != nil {
							t.Fatal("local attribution without replay identity", err)
						}
						for _, raw := range []json.RawMessage{nil, json.RawMessage(`{}`)} {
							admission.IdempotencyIdentity = raw
							if validateExtensionAdmission(params, definition) == nil {
								t.Fatal("none policy accepted absent/object identity")
							}
						}
					}
				})
			}
		}
	}
}
