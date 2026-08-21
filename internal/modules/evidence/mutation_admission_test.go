package evidence_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
)

func TestEvidenceMutationAdmissionAndReplayHashing_Unit(t *testing.T) {
	t.Run("create normalization and default are canonical", func(t *testing.T) {
		omitted, apiErr := evidence.DecodeCreateRequest(strings.NewReader(`{
			"client_txn_id":"txn-evidence-create",
			"evidence.title":"  Disk image  "
		}`))
		if apiErr != nil {
			t.Fatalf("decode omitted-default create: %#v", apiErr)
		}
		explicit, apiErr := evidence.DecodeCreateRequest(strings.NewReader(`{
			"client_txn_id":"txn-evidence-create-other",
			"evidence.lifecycle_state":"requested",
			"evidence.title":"Disk image"
		}`))
		if apiErr != nil {
			t.Fatalf("decode explicit-default create: %#v", apiErr)
		}
		if got := *omitted.Values["evidence.title"].Text; got != "Disk image" {
			t.Fatalf("normalized title got %q", got)
		}
		omittedHash := hex.EncodeToString(evidence.CreateRequestHash(omitted))
		explicitHash := hex.EncodeToString(evidence.CreateRequestHash(explicit))
		const want = "29e3f59295df8149b5ada9c51cdde43dcfb30469a7a5c546af274a1e1e9a4df4"
		if omittedHash != want || explicitHash != want {
			t.Fatalf("create hashes got omitted=%s explicit=%s want=%s", omittedHash, explicitHash, want)
		}
	})

	t.Run("patch changes sort before hashing", func(t *testing.T) {
		request, apiErr := evidence.DecodePatchRequest(strings.NewReader(`{
			"view_schema_id":"cartulary.view.evidence.v1",
			"base_row_version":3,
			"client_txn_id":"txn-evidence-patch",
			"changes":[
				{"field_key":"evidence.title","value":"Disk image"},
				{"field_key":"evidence.lifecycle_state","value":"received"}
			]
		}`))
		if apiErr != nil {
			t.Fatalf("decode patch: %#v", apiErr)
		}
		if request.Changes[0].FieldKey != "evidence.lifecycle_state" || request.Changes[1].FieldKey != "evidence.title" {
			t.Fatalf("patch changes are not canonical: %#v", request.Changes)
		}
		const want = "2687e07b9b9fd93ee54cfb6051cd6ff1224f523230c0e7dec6dafa8fbec61153"
		if got := hex.EncodeToString(evidence.PatchRequestHash(request)); got != want {
			t.Fatalf("patch hash got %s want %s", got, want)
		}
	})

	t.Run("conflict claims and keep-saved hash stay owner-owned", func(t *testing.T) {
		claims := evidence.ConflictClaims{
			RecordID:          uuid.MustParse("00000000-0000-4000-8000-000000210101"),
			ViewSchemaID:      evidence.ViewSchemaID,
			FieldKey:          "evidence.title",
			CurrentRowVersion: 4,
		}
		request, apiErr := evidence.DecodeConflictResolveRequest(
			strings.NewReader(`{
				"conflict_token":"opaque-token",
				"resolution_kind":"keep_saved",
				"client_txn_id":"txn-evidence-conflict"
			}`),
			"opaque-token",
			claims,
		)
		if apiErr != nil {
			t.Fatalf("decode conflict resolution: %#v", apiErr)
		}
		const want = "5bf28191f2c841cb695adbf505fff2ea22c50a0b3f7e143c35276688c430ad5e"
		if got := hex.EncodeToString(evidence.ConflictResolveRequestHash(claims, request)); got != want {
			t.Fatalf("conflict hash got %s want %s", got, want)
		}
	})

	t.Run("initial blob input is closed and typed", func(t *testing.T) {
		for name, body := range map[string]string{
			"null":      `{"client_txn_id":"txn","evidence.title":"Disk image","evidence.initial_object_blob_id":null}`,
			"malformed": `{"client_txn_id":"txn","evidence.title":"Disk image","evidence.initial_object_blob_id":"not-a-uuid"}`,
			"unknown":   `{"client_txn_id":"txn","evidence.title":"Disk image","evidence.unowned":true}`,
		} {
			t.Run(name, func(t *testing.T) {
				if _, apiErr := evidence.DecodeCreateRequest(strings.NewReader(body)); apiErr == nil || apiErr.Code != "invalid_mutation_payload" {
					t.Fatalf("decode got %#v want invalid_mutation_payload", apiErr)
				}
			})
		}
	})
}
