package incidents

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestAdmittedMutationHashesMatchCompatibilityDigests_Unit(t *testing.T) {
	t.Parallel()
	create := func(body string) ([sha256.Size]byte, *AdmissionError) {
		request, err := AdmitIncidentCreateJSON(strings.NewReader(body))
		return request.requestHash, err
	}
	lifecycle := func(action LifecycleAction) func(string) ([sha256.Size]byte, *AdmissionError) {
		return func(body string) ([sha256.Size]byte, *AdmissionError) {
			request, err := AdmitIncidentLifecycleJSON(action, strings.NewReader(body))
			return request.requestHash, err
		}
	}
	membership := func(body string) ([sha256.Size]byte, *AdmissionError) {
		request, err := AdmitMembershipCreateJSON(strings.NewReader(body))
		return request.requestHash, err
	}
	const createBody = `{"client_txn_id":"txn-create-null","incident_key":"IR-NULL","title":"Null fields"}`
	const fullBody = `{"client_txn_id":"txn-create-full","current_phase":"containment","description":"description","incident_key":"IR-FULL","primary_external_case_ref":"CASE-7","severity":"high","title":"Full fields","tlp":"TLP:AMBER"}`
	const closeBody = `{"base_incident_version":7,"client_txn_id":"txn-close","reason":"resolved"}`
	const memberBody = `{"client_txn_id":"txn-member-email","email":"analyst@example.test","role":"viewer"}`

	// Fixed vectors protect the current persisted receipt format. Every vector
	// enters through owner admission; private, inadmissible payloads are excluded.
	tests := []struct {
		name          string
		admit         func(string) ([sha256.Size]byte, *AdmissionError)
		body          string
		wantDigest    string
		sameAs        string
		differentFrom string
	}{
		{name: "create omitted", admit: create, body: createBody, wantDigest: "ed4af5c74f0d155181649a402ad4904d4264abc1c75c7f9d914ed4a2f4b692f5"},
		{name: "create null", admit: create, body: strings.TrimSuffix(createBody, "}") + `,"description":null,"severity":null,"tlp":null,"current_phase":null,"primary_external_case_ref":null}`, sameAs: "create omitted"},
		{name: "create normalized", admit: create, body: strings.ReplaceAll(createBody, "IR-NULL", "  IR-NULL  "), sameAs: "create omitted"},
		{name: "create populated", admit: create, body: fullBody, wantDigest: "a5a2b1390eb0222298db3b26fc76b3aac0b1af544515edddacf82af0588513d5", differentFrom: "create omitted"},
		{name: "create changed title", admit: create, body: strings.ReplaceAll(createBody, "Null fields", "Changed fields"), differentFrom: "create omitted"},
		{name: "create changed description", admit: create, body: strings.ReplaceAll(fullBody, `:"description"`, `:"changed"`), differentFrom: "create populated"},
		{name: "create exact transaction identity", admit: create, body: strings.ReplaceAll(createBody, "txn-create-null", " txn-create-null "), differentFrom: "create omitted"},
		{name: "close", admit: lifecycle(LifecycleActionClose), body: closeBody, wantDigest: "a344ec7cee83b4201fb2e9e7e41eae647551b5f0d85fd2845f813ca6c293a926"},
		{name: "reopen", admit: lifecycle(LifecycleActionReopen), body: closeBody, wantDigest: "e4128c1e5d3cfcc711df4cd861ab4c56b256c4e6c2c8096f9bdd8d1abf8f7329", differentFrom: "close"},
		{name: "lifecycle normalized reason", admit: lifecycle(LifecycleActionClose), body: strings.ReplaceAll(closeBody, "resolved", "  resolved  "), sameAs: "close"},
		{name: "lifecycle changed reason", admit: lifecycle(LifecycleActionClose), body: strings.ReplaceAll(closeBody, "resolved", "investigating"), differentFrom: "close"},
		{name: "lifecycle changed base", admit: lifecycle(LifecycleActionClose), body: strings.ReplaceAll(closeBody, ":7", ":8"), differentFrom: "close"},
		{name: "lifecycle transaction key outside digest", admit: lifecycle(LifecycleActionClose), body: strings.ReplaceAll(closeBody, "txn-close", "other-transaction"), sameAs: "close"},
		{name: "member email", admit: membership, body: memberBody, wantDigest: "0f84db03598a88d155323931c437cf1fb0ea34992a52d3ab04d9e5e1bcf1c95d"},
		{name: "member normalized email", admit: membership, body: strings.ReplaceAll(memberBody, "analyst@example.test", "  Analyst@Example.Test  "), sameAs: "member email"},
		{name: "member changed email", admit: membership, body: strings.ReplaceAll(memberBody, "analyst@example.test", "other@example.test"), differentFrom: "member email"},
		{name: "member changed role", admit: membership, body: strings.ReplaceAll(memberBody, "viewer", "admin"), differentFrom: "member email"},
		{name: "member user", admit: membership, body: `{"client_txn_id":"txn-member-user","user_id":"00000000-0000-0000-0000-000000000707","role":"admin"}`, wantDigest: "d7101c84fc34237d1fc511dfc418fe0221d44b9e58324da798bdbcd2d05253f8"},
		{name: "member selector distinction", admit: membership, body: `{"client_txn_id":"txn-member-email","user_id":"00000000-0000-0000-0000-000000000707","role":"viewer"}`, differentFrom: "member email"},
	}
	digests := make(map[string][sha256.Size]byte, len(tests))
	for _, test := range tests {
		hash, admissionErr := test.admit(test.body)
		if admissionErr != nil {
			t.Fatalf("%s: admit mutation: %v", test.name, admissionErr)
		}
		digests[test.name] = hash
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hash := digests[test.name]
			if test.wantDigest != "" && hex.EncodeToString(hash[:]) != test.wantDigest {
				t.Fatalf("admitted digest = %x, want %s", hash, test.wantDigest)
			}
			if test.sameAs != "" && hash != digests[test.sameAs] {
				t.Fatalf("normalized request differs from %s", test.sameAs)
			}
			if test.differentFrom != "" && hash == digests[test.differentFrom] {
				t.Fatalf("different request equals %s", test.differentFrom)
			}
		})
	}
}
