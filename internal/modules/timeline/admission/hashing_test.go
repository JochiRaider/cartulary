package admission

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func TestTimelineAdmissionCreateAndHashContracts(t *testing.T) {
	t.Run("zero field create remains admitted", func(t *testing.T) {
		request, apiErr := DecodeTimelineCreateRequest(bytes.NewBufferString(`{
			"client_txn_id": "txn-support-timeline_mutation-zero"
		}`))
		if apiErr != nil {
			t.Fatalf("decode zero-field Timeline create: %#v", apiErr)
		}
		if timeline.CreateRequestHasUserValue(request) {
			t.Fatalf("zero-field request unexpectedly has user values: %#v", request)
		}
		if got, want := fmt.Sprintf("%x", CreateRequestHash(request)), "707fa241fc61891b68145853cf5fd56ed74baf61d63431a09c0af44df93efda8"; got != want {
			t.Fatalf("canonical zero-field create request hash changed: got %s want %s", got, want)
		}
	})

	t.Run("patch hash ignores admitted change ordering", func(t *testing.T) {
		left, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
			"view_schema_id": "cartulary.view.timeline.v2",
			"base_row_version": 3,
			"client_txn_id": "txn-support-timeline_mutation-hash",
			"changes": [
				{ "field_key": "timeline.activity_synopsis_text", "value": "summary" },
				{ "field_key": "timeline.raw_activity_text", "value": "details" }
			]
		}`))
		if apiErr != nil {
			t.Fatalf("decode left patch: %#v", apiErr)
		}
		right, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
			"view_schema_id": "cartulary.view.timeline.v2",
			"base_row_version": 3,
			"client_txn_id": "txn-support-timeline_mutation-hash",
			"changes": [
				{ "field_key": "timeline.raw_activity_text", "value": "details" },
				{ "field_key": "timeline.activity_synopsis_text", "value": "summary" }
			]
		}`))
		if apiErr != nil {
			t.Fatalf("decode right patch: %#v", apiErr)
		}
		if !bytes.Equal(PatchRequestHash(left), PatchRequestHash(right)) {
			t.Fatal("canonical patch request hash changed with input ordering")
		}
		if got, want := fmt.Sprintf("%x", PatchRequestHash(left)), "44f43180a1cd46648a6ff12fb685432f9922d23f3cc6e2fa6ed249f09e3a461b"; got != want {
			t.Fatalf("canonical patch request hash changed: got %s want %s", got, want)
		}
	})

	t.Run("action hash remains wire stable", func(t *testing.T) {
		reason := "superseding duplicate"
		replacementRecordID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
		got := fmt.Sprintf("%x", ActionRequestHash(
			4,
			"client-transaction-id-is-not-hashed",
			&reason,
			&replacementRecordID,
		))
		const want = "8029aa9e47b0d78cfbcf269b60a67a0818f479aa889d6a818f684850a23e5ca5"
		if got != want {
			t.Fatalf("normalized action request hash changed: got %s want %s", got, want)
		}
	})
}
