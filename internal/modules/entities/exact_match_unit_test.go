package entities_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestExactMatchPrecedence_Unit(t *testing.T) {
	nullableString := func(value string) any {
		if value == "" {
			return nil
		}
		return value
	}
	startFixture := func(t *testing.T, suffix string) (*appsupport.StoreHarness, *hostidentity.Store, authn.UserRecord, uuid.UUID) {
		t.Helper()

		harness := appsupport.StartStore(t, "entity_linking-u-4-05-"+suffix)
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u405-"+suffix+"@example.test", "U405 "+suffix, "U405EntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-05-"+suffix, "IR-U405-"+suffix, "Record relationships entity-storage "+suffix)
		return harness, store, actor, incident.ID
	}
	seedHost := func(t *testing.T, harness *appsupport.StoreHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string, aadDeviceID string, fqdn string, hostname string) {
		t.Helper()

		entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, recordID, displayName, "seed-hostname", "seed.example.test", "")
		if _, err := harness.DB.Exec(context.Background(), `
UPDATE hosts
   SET aad_device_id = $2,
       fqdn = $3,
       hostname = $4
 WHERE record_id = $1
`, recordID, nullableString(aadDeviceID), nullableString(fqdn), nullableString(hostname)); err != nil {
			t.Fatalf("normalize seeded host identifiers: %v", err)
		}
	}
	seedIdentity := func(t *testing.T, harness *appsupport.StoreHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string, aadObjectID string, sid string, upn string, email string, samAccountName string) {
		t.Helper()

		entitytest.SeedIdentityRecord(t, harness.DB, incidentID, actorID, recordID, displayName, "seed-upn@example.test", "seed-email@example.test", "SEEDSAM")
		if _, err := harness.DB.Exec(context.Background(), `
UPDATE identities
   SET aad_object_id = $2,
       sid = $3,
       upn = $4,
       email = $5,
       sam_account_name = $6
 WHERE record_id = $1
`, recordID, nullableString(aadObjectID), nullableString(sid), nullableString(upn), nullableString(email), nullableString(samAccountName)); err != nil {
			t.Fatalf("normalize seeded identity identifiers: %v", err)
		}
	}

	t.Run("host precedence ladder honors aad_device_id, then fqdn, then hostname", func(t *testing.T) {
		hostCases := []struct {
			name         string
			suffix       string
			values       map[string]string
			wantSelector string
		}{
			{
				name:   "aad_device_id outranks fqdn and hostname",
				suffix: "host-aad",
				values: map[string]string{
					"host.display_name":  "Host ladder aad",
					"host.aad_device_id": "AAD-DEVICE-01",
					"host.fqdn":          "ladder.example.test",
					"host.hostname":      "host-ladder",
				},
				wantSelector: "aad",
			},
			{
				name:   "fqdn outranks hostname when aad_device_id is absent",
				suffix: "host-fqdn",
				values: map[string]string{
					"host.display_name": "Host ladder fqdn",
					"host.fqdn":         "ladder.example.test",
					"host.hostname":     "host-ladder",
				},
				wantSelector: "fqdn",
			},
			{
				name:   "hostname matches when higher-precedence identifiers are absent",
				suffix: "host-hostname",
				values: map[string]string{
					"host.display_name": "Host ladder hostname",
					"host.hostname":     "host-ladder",
				},
				wantSelector: "hostname",
			},
		}
		for _, tc := range hostCases {
			t.Run(tc.name, func(t *testing.T) {
				harness, store, actor, incidentID := startFixture(t, tc.suffix)

				hostAADRecordID := uuid.New()
				hostFQDNRecordID := uuid.New()
				hostHostnameRecordID := uuid.New()
				seedHost(t, harness, incidentID, actor.ID, hostAADRecordID, "AAD Host", "AAD-DEVICE-01", "", "")
				seedHost(t, harness, incidentID, actor.ID, hostFQDNRecordID, "FQDN Host", "", "ladder.example.test", "")
				seedHost(t, harness, incidentID, actor.ID, hostHostnameRecordID, "Hostname Host", "", "", "host-ladder")

				reuse, err := store.CreateHostRow(context.Background(), actor, incidentID, hostidentity.CreateRequest{
					ClientTxnID: "txn-entity_linking-u-4-05-" + tc.suffix,
					Values:      tc.values,
				}, []byte("txn-entity_linking-u-4-05-"+tc.suffix), "req-"+tc.suffix, entitytest.BaseTime)

				wantRecordID := hostHostnameRecordID
				switch tc.wantSelector {
				case "aad":
					wantRecordID = hostAADRecordID
				case "fqdn":
					wantRecordID = hostFQDNRecordID
				}
				if tc.wantSelector != "hostname" {
					var conflict *hostidentity.ExactMatchConflictError
					if !errors.As(err, &conflict) || conflict.IdentifierClass != map[string]string{
						"aad":  "aad_device_id",
						"fqdn": "fqdn",
					}[tc.wantSelector] || len(conflict.CandidateRecords) < 2 {
						t.Fatalf("host cross-record match error = %#v, want ordered conflict", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("host exact-match reuse: %v", err)
				}
				if reuse.RecordID != wantRecordID || reuse.StatusCode != 200 {
					t.Fatalf("unexpected host precedence result: got %#v want_record=%s", reuse, wantRecordID)
				}
			})
		}
	})

	t.Run("identity precedence ladder honors aad_object_id, sid, upn, email, then sam_account_name", func(t *testing.T) {
		identityCases := []struct {
			name         string
			suffix       string
			values       map[string]string
			wantSelector string
		}{
			{
				name:   "aad_object_id outranks sid, upn, email, and sam_account_name",
				suffix: "identity-aad",
				values: map[string]string{
					"identity.display_name":     "Identity ladder aad",
					"identity.aad_object_id":    "AAD-OBJECT-01",
					"identity.sid":              "S-1-5-21-405-500-1001",
					"identity.upn":              "upn.identity@example.test",
					"identity.email":            "email.identity@example.test",
					"identity.sam_account_name": "SAMMATCH",
				},
				wantSelector: "aad",
			},
			{
				name:   "sid outranks upn, email, and sam_account_name when aad_object_id is absent",
				suffix: "identity-sid",
				values: map[string]string{
					"identity.display_name":     "Identity ladder sid",
					"identity.sid":              "S-1-5-21-405-500-1001",
					"identity.upn":              "upn.identity@example.test",
					"identity.email":            "email.identity@example.test",
					"identity.sam_account_name": "SAMMATCH",
				},
				wantSelector: "sid",
			},
			{
				name:   "upn outranks email and sam_account_name when higher-precedence identifiers are absent",
				suffix: "identity-upn",
				values: map[string]string{
					"identity.display_name":     "Identity ladder upn",
					"identity.upn":              "upn.identity@example.test",
					"identity.email":            "email.identity@example.test",
					"identity.sam_account_name": "SAMMATCH",
				},
				wantSelector: "upn",
			},
			{
				name:   "email outranks sam_account_name when higher-precedence identifiers are absent",
				suffix: "identity-email",
				values: map[string]string{
					"identity.display_name":     "Identity ladder email",
					"identity.email":            "email.identity@example.test",
					"identity.sam_account_name": "SAMMATCH",
				},
				wantSelector: "email",
			},
			{
				name:   "sam_account_name matches when it is the only exact-match identifier",
				suffix: "identity-sam",
				values: map[string]string{
					"identity.display_name":     "Identity ladder sam",
					"identity.sam_account_name": "SAMMATCH",
				},
				wantSelector: "sam",
			},
		}
		for _, tc := range identityCases {
			t.Run(tc.name, func(t *testing.T) {
				harness, store, actor, incidentID := startFixture(t, tc.suffix)

				identityAADRecordID := uuid.New()
				identitySIDRecordID := uuid.New()
				identityUPNRecordID := uuid.New()
				identityEmailRecordID := uuid.New()
				identitySAMRecordID := uuid.New()
				seedIdentity(t, harness, incidentID, actor.ID, identityAADRecordID, "AAD Identity", "AAD-OBJECT-01", "", "", "", "")
				seedIdentity(t, harness, incidentID, actor.ID, identitySIDRecordID, "SID Identity", "", "S-1-5-21-405-500-1001", "", "", "")
				seedIdentity(t, harness, incidentID, actor.ID, identityUPNRecordID, "UPN Identity", "", "", "upn.identity@example.test", "", "")
				seedIdentity(t, harness, incidentID, actor.ID, identityEmailRecordID, "Email Identity", "", "", "", "email.identity@example.test", "")
				seedIdentity(t, harness, incidentID, actor.ID, identitySAMRecordID, "SAM Identity", "", "", "", "", "SAMMATCH")

				reuse, err := store.CreateIdentityRow(context.Background(), actor, incidentID, hostidentity.CreateRequest{
					ClientTxnID: "txn-entity_linking-u-4-05-" + tc.suffix,
					Values:      tc.values,
				}, []byte("txn-entity_linking-u-4-05-"+tc.suffix), "req-"+tc.suffix, entitytest.BaseTime.Add(2*time.Minute))
				wantRecordID := identitySAMRecordID
				switch tc.wantSelector {
				case "aad":
					wantRecordID = identityAADRecordID
				case "sid":
					wantRecordID = identitySIDRecordID
				case "upn":
					wantRecordID = identityUPNRecordID
				case "email":
					wantRecordID = identityEmailRecordID
				}
				if tc.wantSelector != "sam" {
					var conflict *hostidentity.ExactMatchConflictError
					if !errors.As(err, &conflict) || conflict.IdentifierClass != map[string]string{
						"aad":   "aad_object_id",
						"sid":   "sid",
						"upn":   "upn",
						"email": "email",
					}[tc.wantSelector] || len(conflict.CandidateRecords) < 2 {
						t.Fatalf("identity cross-record match error = %#v, want ordered conflict", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("identity exact-match reuse: %v", err)
				}
				if reuse.RecordID != wantRecordID || reuse.StatusCode != 200 {
					t.Fatalf("unexpected identity precedence result: got %#v want_record=%s", reuse, wantRecordID)
				}
			})
		}
	})

	t.Run("suggestion-only aliases and fuzzy non-matches stay non-authoritative", func(t *testing.T) {
		t.Run("host suggestion-only alias does not trigger implicit reuse", func(t *testing.T) {
			harness, store, actor, incidentID := startFixture(t, "host-alias")

			hostAliasRecordID := uuid.New()
			seedHost(t, harness, incidentID, actor.ID, hostAliasRecordID, "Canonical Alias Host", "", "ws-023.corp.example.test", "WS-023")
			entitytest.SeedEntityAlias(t, harness.DB, incidentID, actor.ID, hostAliasRecordID, "host", "Workstation 23")

			hostAliasOnly, err := store.CreateHostRow(context.Background(), actor, incidentID, hostidentity.CreateRequest{
				ClientTxnID: "txn-entity_linking-u-4-05-host-alias",
				Values: map[string]string{
					"host.display_name": "Workstation 23",
				},
			}, []byte("txn-entity_linking-u-4-05-host-alias"), "req-host-alias", entitytest.BaseTime.Add(time.Minute))
			if err != nil {
				t.Fatalf("host alias-only create: %v", err)
			}
			if hostAliasOnly.RecordID == hostAliasRecordID || hostAliasOnly.StatusCode != 201 {
				t.Fatalf("expected suggestion-only alias to avoid implicit reuse, got %#v", hostAliasOnly)
			}
		})

		t.Run("identity fuzzy near-match does not trigger implicit reuse", func(t *testing.T) {
			harness, store, actor, incidentID := startFixture(t, "identity-fuzzy")

			identityAliasRecordID := uuid.New()
			seedIdentity(t, harness, incidentID, actor.ID, identityAliasRecordID, "Case Owner", "", "", "", "", "CASEOWNER")
			entitytest.SeedEntityAlias(t, harness.DB, incidentID, actor.ID, identityAliasRecordID, "identity", "Case Owner")

			identityFuzzyNonMatch, err := store.CreateIdentityRow(context.Background(), actor, incidentID, hostidentity.CreateRequest{
				ClientTxnID: "txn-entity_linking-u-4-05-identity-fuzzy",
				Values: map[string]string{
					"identity.display_name": "Case Ownr",
				},
			}, []byte("txn-entity_linking-u-4-05-identity-fuzzy"), "req-identity-fuzzy", entitytest.BaseTime.Add(3*time.Minute))
			if err != nil {
				t.Fatalf("identity fuzzy non-match create: %v", err)
			}
			if identityFuzzyNonMatch.RecordID == identityAliasRecordID || identityFuzzyNonMatch.StatusCode != 201 {
				t.Fatalf("expected fuzzy non-match to avoid implicit reuse, got %#v", identityFuzzyNonMatch)
			}
		})
	})
}

// entity-storage / REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
