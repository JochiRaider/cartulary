package incidents

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestIncidentIdempotencyHashCompatibility_Unit(t *testing.T) {
	t.Parallel()

	description := "description"
	severity := "high"
	tlp := "amber"
	phase := "containment"
	externalCase := "CASE-7"
	memberID := uuid.MustParse("00000000-0000-0000-0000-000000000707")
	email := "analyst@example.test"

	tests := []struct {
		name       string
		preimage   string
		wantDigest string
		got        [sha256.Size]byte
	}{
		{
			name:       "create nullable members remain explicit null",
			preimage:   `{"client_txn_id":"txn-create-null","current_phase":null,"description":null,"incident_key":"IR-NULL","primary_external_case_ref":null,"severity":null,"title":"Null fields","tlp":null}`,
			wantDigest: "ed4af5c74f0d155181649a402ad4904d4264abc1c75c7f9d914ed4a2f4b692f5",
			got: hashRequestPayload(incidentCreateIdempotencyPayload{
				ClientTxnID: "txn-create-null", IncidentKey: "IR-NULL", Title: "Null fields",
			}),
		},
		{
			name:       "create populated members retain canonical order",
			preimage:   `{"client_txn_id":"txn-create-full","current_phase":"containment","description":"description","incident_key":"IR-FULL","primary_external_case_ref":"CASE-7","severity":"high","title":"Full fields","tlp":"amber"}`,
			wantDigest: "7121cc4093ff20e0404c557cde7976cae82cebd1b5bebc208add35d690914012",
			got: hashRequestPayload(incidentCreateIdempotencyPayload{
				ClientTxnID:            "txn-create-full",
				IncidentKey:            "IR-FULL",
				Title:                  "Full fields",
				Description:            &description,
				Severity:               &severity,
				TLP:                    &tlp,
				CurrentPhase:           &phase,
				PrimaryExternalCaseRef: &externalCase,
			}),
		},
		{
			name:       "lifecycle close includes action route",
			preimage:   `{"action_route":"close","base_incident_version":7,"reason":"resolved"}`,
			wantDigest: "a344ec7cee83b4201fb2e9e7e41eae647551b5f0d85fd2845f813ca6c293a926",
			got: hashRequestPayload(incidentLifecycleIdempotencyPayload{
				ActionRoute: "close", BaseIncidentVersion: 7, Reason: "resolved",
			}),
		},
		{
			name:       "lifecycle reopen is distinct",
			preimage:   `{"action_route":"reopen","base_incident_version":7,"reason":"resolved"}`,
			wantDigest: "e4128c1e5d3cfcc711df4cd861ab4c56b256c4e6c2c8096f9bdd8d1abf8f7329",
			got: hashRequestPayload(incidentLifecycleIdempotencyPayload{
				ActionRoute: "reopen", BaseIncidentVersion: 7, Reason: "resolved",
			}),
		},
		{
			name:       "membership selected by user id",
			preimage:   `{"client_txn_id":"txn-member-user","email":null,"role":"admin","user_id":"00000000-0000-0000-0000-000000000707"}`,
			wantDigest: "d7101c84fc34237d1fc511dfc418fe0221d44b9e58324da798bdbcd2d05253f8",
			got: hashRequestPayload(membershipCreateIdempotencyPayload{
				ClientTxnID: "txn-member-user",
				UserID:      &memberID,
				Role:        "admin",
			}),
		},
		{
			name:       "membership selected by email",
			preimage:   `{"client_txn_id":"txn-member-email","email":"analyst@example.test","role":"viewer","user_id":null}`,
			wantDigest: "0f84db03598a88d155323931c437cf1fb0ea34992a52d3ab04d9e5e1bcf1c95d",
			got: hashRequestPayload(membershipCreateIdempotencyPayload{
				ClientTxnID: "txn-member-email",
				Email:       &email,
				Role:        "viewer",
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			preimageDigest := sha256.Sum256([]byte(test.preimage))
			if got := hex.EncodeToString(preimageDigest[:]); got != test.wantDigest {
				t.Fatalf("golden preimage digest = %s, want %s; preimage=%s", got, test.wantDigest, test.preimage)
			}
			if got := hex.EncodeToString(test.got[:]); got != test.wantDigest {
				t.Fatalf("application request digest = %s, want %s; preimage=%s", got, test.wantDigest, test.preimage)
			}
		})
	}
}

func TestAdmittedMutationHashesMatchCompatibilityDigests_Unit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		wantDigest string
		admit      func() ([sha256.Size]byte, *AdmissionError)
	}{
		{
			name:       "incident create",
			wantDigest: "ed4af5c74f0d155181649a402ad4904d4264abc1c75c7f9d914ed4a2f4b692f5",
			admit: func() ([sha256.Size]byte, *AdmissionError) {
				request, err := AdmitIncidentCreateJSON(strings.NewReader(
					`{"client_txn_id":"txn-create-null","incident_key":"IR-NULL","title":"Null fields"}`,
				))
				return request.requestHash, err
			},
		},
		{
			name:       "lifecycle close",
			wantDigest: "a344ec7cee83b4201fb2e9e7e41eae647551b5f0d85fd2845f813ca6c293a926",
			admit: func() ([sha256.Size]byte, *AdmissionError) {
				request, err := AdmitIncidentLifecycleJSON(LifecycleActionClose, strings.NewReader(
					`{"base_incident_version":7,"client_txn_id":"txn-close-not-in-preimage","reason":"resolved"}`,
				))
				return request.requestHash, err
			},
		},
		{
			name:       "lifecycle reopen",
			wantDigest: "e4128c1e5d3cfcc711df4cd861ab4c56b256c4e6c2c8096f9bdd8d1abf8f7329",
			admit: func() ([sha256.Size]byte, *AdmissionError) {
				request, err := AdmitIncidentLifecycleJSON(LifecycleActionReopen, strings.NewReader(
					`{"base_incident_version":7,"client_txn_id":"txn-reopen-not-in-preimage","reason":"resolved"}`,
				))
				return request.requestHash, err
			},
		},
		{
			name:       "membership by user ID",
			wantDigest: "d7101c84fc34237d1fc511dfc418fe0221d44b9e58324da798bdbcd2d05253f8",
			admit: func() ([sha256.Size]byte, *AdmissionError) {
				request, err := AdmitMembershipCreateJSON(strings.NewReader(
					`{"client_txn_id":"txn-member-user","user_id":"00000000-0000-0000-0000-000000000707","role":"admin"}`,
				))
				return request.requestHash, err
			},
		},
		{
			name:       "membership by email",
			wantDigest: "0f84db03598a88d155323931c437cf1fb0ea34992a52d3ab04d9e5e1bcf1c95d",
			admit: func() ([sha256.Size]byte, *AdmissionError) {
				request, err := AdmitMembershipCreateJSON(strings.NewReader(
					`{"client_txn_id":"txn-member-email","email":"analyst@example.test","role":"viewer"}`,
				))
				return request.requestHash, err
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			hash, admissionErr := test.admit()
			if admissionErr != nil {
				t.Fatalf("admit mutation: %v", admissionErr)
			}
			if got := hex.EncodeToString(hash[:]); got != test.wantDigest {
				t.Fatalf("admitted request digest = %s, want %s", got, test.wantDigest)
			}
		})
	}
}
