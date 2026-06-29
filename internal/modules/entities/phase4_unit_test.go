package entities_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	. "github.com/JochiRaider/cartulary/internal/modules/entities"
	timeline "github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/assertx"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	phase4storetest "github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
)

func mustDefaultQueryMeta(t testing.TB, viewSchemaID string) viewschema.QueryMeta {
	t.Helper()
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("view schema %q not registered", viewSchemaID)
	}
	return schema.DefaultQueryMeta()
}

// U-4-03 / REQ-02-034, REQ-02-038, REQ-02-054..REQ-02-055 / AC-020, AC-021, AC-186.
func TestPhase4_CreateFromMention_U_4_03(t *testing.T) {
	t.Run("explicit host resolve links the selected mention to the canonical record", func(t *testing.T) {
		harness := phase4storetest.StartStore(t, "phase4-u-4-03-host-reuse")
		store := NewStore(harness.DB)
		actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u403-host@example.test", "U403 Host", "U403HostPhase4Pass1!", false, false, true)
		incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-03-host-incident", "IR-U403-H", "Phase 4 U-4-03 host")

		phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineRecordID)
		phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineSiblingRecordID)
		phase4storetest.SeedMention(t, harness.DB, actor.ID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)
		phase4storetest.SeedMention(t, harness.DB, actor.ID, golden.Phase4ResolvedHostMentionID, golden.Phase4TimelineSiblingRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)

		tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		targetID := golden.Phase4CanonicalHostRecordID
		result, err := store.ResolveOrCreateFromMentionTx(context.Background(), tx, actor, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, golden.Phase4HostMentionID, &targetID, golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("resolve mention: %v", err)
		}
		if err := tx.Commit(context.Background()); err != nil {
			t.Fatalf("commit tx: %v", err)
		}

		if result.RecordID != golden.Phase4CanonicalHostRecordID || result.EntityType != "host" {
			t.Fatalf("expected selected mention to reuse canonical host, got %#v", result)
		}
		selected := phase4storetest.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, selected, golden.Phase4MentionStatusResolved)
		if selected.ResolvedRecordID == nil || *selected.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected selected mention to resolve to canonical host, got %#v", selected)
		}
		if selected.RawText != "WS-023" {
			t.Fatalf("expected selected raw text preservation, got %#v", selected)
		}

		sibling := phase4storetest.LookupMention(t, harness.DB, golden.Phase4ResolvedHostMentionID)
		assertx.RequireMentionStatus(t, sibling, golden.Phase4MentionStatusUnresolved)
		if sibling.ResolvedRecordID != nil {
			t.Fatalf("expected sibling mention to remain unresolved, got %#v", sibling)
		}
		if got := phase4storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 1 {
			t.Fatalf("expected explicit resolve to avoid creating a second host, got %d rows", got)
		}
		var (
			projectedDisplayName string
			projectedHostname    string
			projectedState       string
		)
		if err := harness.DB.QueryRow(context.Background(), `
SELECT display_name, hostname, host_state
  FROM host_grid_projection
 WHERE record_id = $1
`, golden.Phase4CanonicalHostRecordID).Scan(&projectedDisplayName, &projectedHostname, &projectedState); err != nil {
			t.Fatalf("lookup host projection after explicit resolve: %v", err)
		}
		if projectedDisplayName != "WS-023" || projectedHostname != "WS-023" || projectedState != "canonical" {
			t.Fatalf("unexpected host projection after explicit resolve: display=%q hostname=%q state=%q", projectedDisplayName, projectedHostname, projectedState)
		}
	})

	t.Run("identity resolve without an explicit target is rejected without creating a stub", func(t *testing.T) {
		harness := phase4storetest.StartStore(t, "phase4-u-4-03-identity-create")
		store := NewStore(harness.DB)
		actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u403-identity@example.test", "U403 Identity", "U403IdentityPhase4Pass1!", false, false, true)
		incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-03-identity-incident", "IR-U403-I", "Phase 4 U-4-03 identity")

		phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineMixedRecordID)
		phase4storetest.SeedMention(t, harness.DB, actor.ID, golden.Phase4IdentityMentionID, golden.Phase4TimelineMixedRecordID, golden.Phase4FieldTimelineIdentityRefs, "identity", "alex.analyst@example.test", "unresolved", nil, nil)

		tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		_, err = store.ResolveOrCreateFromMentionTx(context.Background(), tx, actor, golden.Phase4TimelineMixedRecordID, golden.Phase4FieldTimelineIdentityRefs, golden.Phase4IdentityMentionID, nil, golden.Phase4BaseTime)
		if !errors.Is(err, ErrInvalidMentionResolution) {
			t.Fatalf("expected missing target rejection, got %v", err)
		}

		mention := phase4storetest.LookupMention(t, harness.DB, golden.Phase4IdentityMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusUnresolved)
		if mention.ResolvedRecordID != nil {
			t.Fatalf("expected mention to remain unresolved, got %#v", mention)
		}
		if got := phase4storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id = $1`, incident.ID); got != 0 {
			t.Fatalf("expected missing-target resolve to create no identity rows, got %d", got)
		}
	})
}

// U-4-04 / REQ-02-039..REQ-02-041 / AC-188..AC-190, AC-224, AC-225.
func TestPhase4_DismissRestoreMentionLifecycle_U_4_04(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase4-u-4-04")
	store := NewStore(harness.DB)
	timelineStore := timeline.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u404@example.test", "U404", "U404Phase4Pass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-04-incident", "IR-U404", "Phase 4 U-4-04")

	phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
	normalizedHostToken, ok := fieldnorm.NormalizeMentionToken("WS-023")
	if !ok {
		t.Fatal("normalize mention token")
	}
	summary := "dismiss and restore relationship row"
	created, err := timelineStore.CreateRow(context.Background(), actor, incident.ID, timeline.CreateRequest{
		ClientTxnID:          "txn-phase4-u-4-04-row",
		ActivitySynopsisText: &summary,
		HostRefs: &timeline.CollectionActionPayload{
			Actions: []timeline.CollectionAction{
				{
					Op:             "add_resolved_ref",
					RawText:        "WS-023",
					NormalizedText: normalizedHostToken,
					ResolvedRecord: &golden.Phase4CanonicalHostRecordID,
				},
			},
		},
	}, []byte("txn-phase4-u-4-04-row"), "req-phase4-u-4-04-row", golden.Phase4BaseTime)
	if err != nil {
		t.Fatalf("create resolved relationship row: %v", err)
	}
	initialRows, err := timelineStore.QueryRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query initial timeline rows: %v", err)
	}
	initialRow := phase4storetest.FindRow(t, initialRows, created.RecordID.String())
	initialItem := phase4storetest.RequireSingleCollectionItem(t, initialRow, golden.Phase4FieldTimelineHostRefs)
	mentionID := phase4storetest.MentionIDFromItemRef(t, initialItem["item_ref"].(string))

	tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
	if err != nil {
		t.Fatalf("begin dismiss tx: %v", err)
	}
	if err := store.ApplyMentionLifecycleTx(context.Background(), tx, actor, created.RecordID, golden.Phase4FieldTimelineHostRefs, mentionID, "dismiss_item", nil, golden.Phase4BaseTime); err != nil {
		t.Fatalf("dismiss mention: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit dismiss tx: %v", err)
	}

	dismissed := phase4storetest.LookupMention(t, harness.DB, mentionID)
	assertx.RequireMentionStatus(t, dismissed, golden.Phase4MentionStatusDismissed)
	if dismissed.ResolvedRecordID != nil || dismissed.ResolvedAt != nil || dismissed.ResolutionMethod != nil {
		t.Fatalf("expected dismissed mention to clear resolution metadata, got %#v", dismissed)
	}
	if got := phase4storetest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incident.ID, created.RecordID, golden.Phase4CanonicalHostRecordID); got != 0 {
		t.Fatalf("expected dismiss to remove active derived link, got %d rows", got)
	}
	dismissedRows, err := timelineStore.QueryRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query timeline rows after dismiss: %v", err)
	}
	dismissedRow := phase4storetest.FindRow(t, dismissedRows, created.RecordID.String())
	if got := phase4storetest.CollectionItems(t, dismissedRow, golden.Phase4FieldTimelineHostRefs); len(got) != 0 {
		t.Fatalf("dismissed mention must be excluded from current relationship-cell values, got %#v", got)
	}

	tx, err = harness.DB.BeginTx(context.Background(), pgxTxOptions())
	if err != nil {
		t.Fatalf("begin restore tx: %v", err)
	}
	if err := store.ApplyMentionLifecycleTx(context.Background(), tx, actor, created.RecordID, golden.Phase4FieldTimelineHostRefs, mentionID, "revert_to_unresolved", nil, golden.Phase4BaseTime.Add(time.Minute)); err != nil {
		t.Fatalf("restore mention: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit restore tx: %v", err)
	}

	restored := phase4storetest.LookupMention(t, harness.DB, mentionID)
	assertx.RequireMentionStatus(t, restored, golden.Phase4MentionStatusUnresolved)
	if restored.RawText != "WS-023" || restored.ResolvedRecordID != nil || restored.ResolutionMethod != nil {
		t.Fatalf("expected durable restore-to-unresolved semantics, got %#v", restored)
	}
	restoredRows, err := timelineStore.QueryRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query timeline rows after restore: %v", err)
	}
	restoredRow := phase4storetest.FindRow(t, restoredRows, created.RecordID.String())
	restoredItem := phase4storetest.RequireSingleCollectionItem(t, restoredRow, golden.Phase4FieldTimelineHostRefs)
	if restoredItem["item_kind"] != "unresolved_mention" || restoredItem["raw_text"] != "WS-023" {
		t.Fatalf("restore must surface the unresolved mention in current-state reads, got %#v", restoredItem)
	}
	if _, ok := restoredItem["resolved_record_id"]; ok {
		t.Fatalf("restore must not silently relink the historical target, got %#v", restoredItem)
	}
}

// U-4-05 / REQ-02-059..REQ-02-063 / AC-021, AC-022.
func TestPhase4_ExactMatchPrecedence_U_4_05(t *testing.T) {
	nullableString := func(value string) any {
		if value == "" {
			return nil
		}
		return value
	}
	startFixture := func(t *testing.T, suffix string) (*phase4storetest.StoreHarness, *Store, authn.UserRecord, uuid.UUID) {
		t.Helper()

		harness := phase4storetest.StartStore(t, "phase4-u-4-05-"+suffix)
		store := NewStore(harness.DB)
		actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u405-"+suffix+"@example.test", "U405 "+suffix, "U405Phase4Pass1!", false, false, true)
		incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-05-"+suffix, "IR-U405-"+suffix, "Phase 4 U-4-05 "+suffix)
		return harness, store, actor, incident.ID
	}
	seedHost := func(t *testing.T, harness *phase4storetest.StoreHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string, aadDeviceID string, fqdn string, hostname string) {
		t.Helper()

		phase4storetest.SeedHostRecord(t, harness.DB, incidentID, actorID, recordID, displayName, "seed-hostname", "seed.example.test", "")
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
	seedIdentity := func(t *testing.T, harness *phase4storetest.StoreHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string, aadObjectID string, sid string, upn string, email string, samAccountName string) {
		t.Helper()

		phase4storetest.SeedIdentityRecord(t, harness.DB, incidentID, actorID, recordID, displayName, "seed-upn@example.test", "seed-email@example.test", "SEEDSAM")
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

				reuse, err := store.CreateHostRow(context.Background(), actor, incidentID, CreateRequest{
					ClientTxnID: "txn-phase4-u-4-05-" + tc.suffix,
					Values:      tc.values,
				}, []byte("txn-phase4-u-4-05-"+tc.suffix), "req-"+tc.suffix, golden.Phase4BaseTime)
				if err != nil {
					t.Fatalf("host exact-match reuse: %v", err)
				}

				wantRecordID := hostHostnameRecordID
				switch tc.wantSelector {
				case "aad":
					wantRecordID = hostAADRecordID
				case "fqdn":
					wantRecordID = hostFQDNRecordID
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

				reuse, err := store.CreateIdentityRow(context.Background(), actor, incidentID, CreateRequest{
					ClientTxnID: "txn-phase4-u-4-05-" + tc.suffix,
					Values:      tc.values,
				}, []byte("txn-phase4-u-4-05-"+tc.suffix), "req-"+tc.suffix, golden.Phase4BaseTime.Add(2*time.Minute))
				if err != nil {
					t.Fatalf("identity exact-match reuse: %v", err)
				}

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
			phase4storetest.SeedEntityAlias(t, harness.DB, incidentID, actor.ID, hostAliasRecordID, "host", "Workstation 23")

			hostAliasOnly, err := store.CreateHostRow(context.Background(), actor, incidentID, CreateRequest{
				ClientTxnID: "txn-phase4-u-4-05-host-alias",
				Values: map[string]string{
					"host.display_name": "Workstation 23",
				},
			}, []byte("txn-phase4-u-4-05-host-alias"), "req-host-alias", golden.Phase4BaseTime.Add(time.Minute))
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
			phase4storetest.SeedEntityAlias(t, harness.DB, incidentID, actor.ID, identityAliasRecordID, "identity", "Case Owner")

			identityFuzzyNonMatch, err := store.CreateIdentityRow(context.Background(), actor, incidentID, CreateRequest{
				ClientTxnID: "txn-phase4-u-4-05-identity-fuzzy",
				Values: map[string]string{
					"identity.display_name": "Case Ownr",
				},
			}, []byte("txn-phase4-u-4-05-identity-fuzzy"), "req-identity-fuzzy", golden.Phase4BaseTime.Add(3*time.Minute))
			if err != nil {
				t.Fatalf("identity fuzzy non-match create: %v", err)
			}
			if identityFuzzyNonMatch.RecordID == identityAliasRecordID || identityFuzzyNonMatch.StatusCode != 201 {
				t.Fatalf("expected fuzzy non-match to avoid implicit reuse, got %#v", identityFuzzyNonMatch)
			}
		})
	})
}

// U-4-06 / REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
func TestPhase4_ExplicitEntityMerge_U_4_06(t *testing.T) {
	t.Run("host merge preserves raw mentions, loser lineage, and survivor reuse", func(t *testing.T) {
		harness := phase4storetest.StartStore(t, "phase4-u-4-06-host")
		store := NewStore(harness.DB)
		actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u406@example.test", "U406", "U406Phase4Pass1!", false, false, true)
		incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-06-incident", "IR-U406", "Phase 4 U-4-06")

		phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
		phase4storetest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateHostRecordID, "host", "Workstation 23")
		phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineRecordID)
		phase4storetest.SeedResolvedMention(t, harness.DB, actor.ID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateHostRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023")
		phase4storetest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateLinkID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateHostRecordID, "observed_on_host", "manual", nil)
		phase4storetest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, golden.Phase4TagIDSurvivor, golden.Phase4CanonicalHostRecordID, "critical-host")
		phase4storetest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, golden.Phase4TagIDLoser, golden.Phase4DuplicateHostRecordID, "critical-host")
		phase4storetest.SeedAssessment(t, harness.DB, incident.ID, actor.ID, golden.Phase4AssessmentHostID, golden.Phase4DuplicateHostRecordID, "host", "confirmed")
		beforeMention := phase4storetest.LookupMention(t, harness.DB, golden.Phase4HostMentionID)

		result, err := store.MergeEntity(context.Background(), actor, golden.Phase4CanonicalHostRecordID, MergeRequest{
			LoserRecordID:          golden.Phase4DuplicateHostRecordID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-phase4-u-4-06-merge",
		}, []byte("txn-phase4-u-4-06-merge"), "req-merge", golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("merge entity: %v", err)
		}
		if result.SurvivorRecordID != golden.Phase4CanonicalHostRecordID || result.LoserRecordID != golden.Phase4DuplicateHostRecordID {
			t.Fatalf("unexpected merge result: %#v", result)
		}

		mention := phase4storetest.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected merge to repoint mention resolution to survivor, got %#v", mention)
		}
		assertx.RequireRawTextPreserved(t, beforeMention.RawText, mention.RawText)
		link := phase4storetest.LookupActiveLink(t, harness.DB, incident.ID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(t, link, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host", "manual", nil)
		if got := phase4storetest.LookupAssessmentSubject(t, harness.DB, golden.Phase4AssessmentHostID); got != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected assessment repoint to survivor, got %s", got)
		}
		if got := phase4storetest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incident.ID, golden.Phase4CanonicalHostRecordID); got != 1 {
			t.Fatalf("expected one active deduped survivor tag, got %d", got)
		}
		if got := phase4storetest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, golden.Phase4DuplicateLinkID); got != 0 {
			t.Fatalf("expected loser-targeted active link to disappear, got %d rows", got)
		}
		state, mergedInto, rowVersion, _ := phase4storetest.LookupHostState(t, harness.DB, golden.Phase4DuplicateHostRecordID)
		if state != "merged" || mergedInto == nil || *mergedInto != golden.Phase4CanonicalHostRecordID || rowVersion != 2 {
			t.Fatalf("expected loser host lineage state after merge, got state=%s merged_into=%v row_version=%d", state, mergedInto, rowVersion)
		}

		reuse, err := store.CreateHostRow(context.Background(), actor, incident.ID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-06-reuse",
			Values: map[string]string{
				"host.fqdn": "ws-023.corp.example.test",
			},
		}, []byte("txn-phase4-u-4-06-reuse"), "req-reuse", golden.Phase4BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge exact-match reuse: %v", err)
		}
		if reuse.RecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected exact-match reuse to carry forward to survivor, got %#v", reuse)
		}
	})

	t.Run("identity merge preserves raw mentions, loser lineage, and survivor reuse", func(t *testing.T) {
		harness := phase4storetest.StartStore(t, "phase4-u-4-06-identity")
		store := NewStore(harness.DB)
		actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u406-identity@example.test", "U406 Identity", "U406IdentityPhase4Pass1!", false, false, true)
		incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-06-identity-incident", "IR-U406-I", "Phase 4 U-4-06 identity")

		phase4storetest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalIdentityID, "Alex Analyst", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		phase4storetest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateIdentityID, "Alex Duplicate", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		phase4storetest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateIdentityID, "identity", "Case Owner")
		phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineRecordID)
		phase4storetest.SeedResolvedMention(t, harness.DB, actor.ID, golden.Phase4IdentityMentionID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateIdentityID, golden.Phase4FieldTimelineIdentityRefs, "identity", "Case Owner")
		phase4storetest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateLinkID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateIdentityID, "observed_as_identity", "manual", nil)
		phase4storetest.SeedAssessment(t, harness.DB, incident.ID, actor.ID, golden.Phase4AssessmentIdentID, golden.Phase4DuplicateIdentityID, "identity", "confirmed")
		beforeMention := phase4storetest.LookupMention(t, harness.DB, golden.Phase4IdentityMentionID)

		result, err := store.MergeEntity(context.Background(), actor, golden.Phase4CanonicalIdentityID, MergeRequest{
			LoserRecordID:          golden.Phase4DuplicateIdentityID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-phase4-u-4-06-identity-merge",
		}, []byte("txn-phase4-u-4-06-identity-merge"), "req-identity-merge", golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("merge identity: %v", err)
		}
		if result.SurvivorRecordID != golden.Phase4CanonicalIdentityID || result.LoserRecordID != golden.Phase4DuplicateIdentityID {
			t.Fatalf("unexpected identity merge result: %#v", result)
		}

		mention := phase4storetest.LookupMention(t, harness.DB, golden.Phase4IdentityMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.Phase4CanonicalIdentityID {
			t.Fatalf("expected identity merge to repoint mention resolution to survivor, got %#v", mention)
		}
		assertx.RequireRawTextPreserved(t, beforeMention.RawText, mention.RawText)
		link := phase4storetest.LookupActiveLink(t, harness.DB, incident.ID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalIdentityID, "observed_as_identity")
		assertx.RequireActiveLink(t, link, golden.Phase4TimelineRecordID, golden.Phase4CanonicalIdentityID, "observed_as_identity", "manual", nil)
		if got := phase4storetest.LookupAssessmentSubject(t, harness.DB, golden.Phase4AssessmentIdentID); got != golden.Phase4CanonicalIdentityID {
			t.Fatalf("expected identity assessment repoint to survivor, got %s", got)
		}

		var (
			state         string
			mergedIntoRaw sql.NullString
			rowVersion    int64
		)
		if err := harness.DB.QueryRow(context.Background(), `
SELECT identity_state, merged_into_record_id::text, row_version
  FROM identities
 WHERE record_id = $1
`, golden.Phase4DuplicateIdentityID).Scan(&state, &mergedIntoRaw, &rowVersion); err != nil {
			t.Fatalf("lookup loser identity state: %v", err)
		}
		if state != "merged" || !mergedIntoRaw.Valid || mergedIntoRaw.String != golden.Phase4CanonicalIdentityID.String() || rowVersion != 2 {
			t.Fatalf("expected loser identity lineage after merge, got state=%s merged_into=%v row_version=%d", state, mergedIntoRaw, rowVersion)
		}

		reuse, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-06-identity-reuse",
			Values: map[string]string{
				"identity.email": "alex.analyst@example.test",
			},
		}, []byte("txn-phase4-u-4-06-identity-reuse"), "req-identity-reuse", golden.Phase4BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge identity exact-match reuse: %v", err)
		}
		if reuse.RecordID != golden.Phase4CanonicalIdentityID {
			t.Fatalf("expected identity exact-match reuse to carry forward to survivor, got %#v", reuse)
		}
	})

	t.Run("host merge exposes carried secondary reusable identifiers", func(t *testing.T) {
		harness := phase4storetest.StartStore(t, "phase4-u-4-06-host-reusable-row")
		store := NewStore(harness.DB)
		actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u406-host-reusable@example.test", "U406 Host Reusable", "U406HostReusablePhase4Pass1!", false, false, true)
		incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-06-host-reusable-incident", "IR-U406-HR", "Phase 4 U-4-06 host reusable rows")
		survivorID := uuid.New()
		loserID := uuid.New()

		phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, survivorID, "WS-023", "WS-023", "ws-023.current.example.test", "")
		phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, loserID, "Legacy WS-023", "LEGACY-WS-023", "legacy-ws-023.example.test", "")

		if _, err := store.MergeEntity(context.Background(), actor, survivorID, MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-phase4-u-4-06-host-reusable-merge",
		}, []byte("txn-phase4-u-4-06-host-reusable-merge"), "req-host-reusable-merge", golden.Phase4BaseTime); err != nil {
			t.Fatalf("merge host reusable identifiers: %v", err)
		}

		rows, err := store.QueryHostRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, HostsViewSchemaID))
		if err != nil {
			t.Fatalf("query host rows after reusable merge: %v", err)
		}
		survivorRow := requireEntityRow(t, rows, survivorID)
		requireReusableIdentifierItem(t, survivorRow, "host.reusable_identifiers", "fqdn", "legacy-ws-023.example.test")
		requireNoReusableIdentifierItem(t, survivorRow, "host.reusable_identifiers", "fqdn", "ws-023.current.example.test")

		reuse, err := store.CreateHostRow(context.Background(), actor, incident.ID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-06-host-reusable-create",
			Values: map[string]string{
				"host.fqdn": "legacy-ws-023.example.test",
			},
		}, []byte("txn-phase4-u-4-06-host-reusable-create"), "req-host-reusable-create", golden.Phase4BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge host reusable exact match: %v", err)
		}
		if reuse.RecordID != survivorID {
			t.Fatalf("expected reusable identifier to match survivor, got %#v", reuse)
		}
		reuseRow := reuse.Payload["row"].(map[string]any)
		requireReusableIdentifierItem(t, reuseRow, "host.reusable_identifiers", "fqdn", "legacy-ws-023.example.test")
	})

	t.Run("identity merge exposes carried secondary reusable identifiers", func(t *testing.T) {
		harness := phase4storetest.StartStore(t, "phase4-u-4-06-identity-reusable-row")
		store := NewStore(harness.DB)
		actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u406-identity-reusable@example.test", "U406 Identity Reusable", "U406IdentityReusablePhase4Pass1!", false, false, true)
		incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-06-identity-reusable-incident", "IR-U406-IR", "Phase 4 U-4-06 identity reusable rows")
		survivorID := uuid.New()
		loserID := uuid.New()

		phase4storetest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, survivorID, "Alex Survivor", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		phase4storetest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, loserID, "Alex Analyst Legacy", "alex.legacy@example.test", "alex.legacy@example.test", "ALEXLEGACY")

		if _, err := store.MergeEntity(context.Background(), actor, survivorID, MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-phase4-u-4-06-identity-reusable-merge",
		}, []byte("txn-phase4-u-4-06-identity-reusable-merge"), "req-identity-reusable-merge", golden.Phase4BaseTime); err != nil {
			t.Fatalf("merge identity reusable identifiers: %v", err)
		}

		rows, err := store.QueryIdentityRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, IdentitiesViewSchemaID))
		if err != nil {
			t.Fatalf("query identity rows after reusable merge: %v", err)
		}
		survivorRow := requireEntityRow(t, rows, survivorID)
		requireReusableIdentifierItem(t, survivorRow, "identity.reusable_identifiers", "email", "alex.legacy@example.test")
		requireNoReusableIdentifierItem(t, survivorRow, "identity.reusable_identifiers", "email", "alex.survivor@example.test")

		reuse, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-06-identity-reusable-create",
			Values: map[string]string{
				"identity.email": "alex.legacy@example.test",
			},
		}, []byte("txn-phase4-u-4-06-identity-reusable-create"), "req-identity-reusable-create", golden.Phase4BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge identity reusable exact match: %v", err)
		}
		if reuse.RecordID != survivorID {
			t.Fatalf("expected reusable identifier to match survivor, got %#v", reuse)
		}
		reuseRow := reuse.Payload["row"].(map[string]any)
		requireReusableIdentifierItem(t, reuseRow, "identity.reusable_identifiers", "email", "alex.legacy@example.test")
	})
}

// U-4-07 / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestPhase4_IndicatorObservationSeparation_U_4_07(t *testing.T) {
	startFixture := func(t *testing.T, suffix string) (*phase4storetest.StoreHarness, *Store, authn.UserRecord, uuid.UUID) {
		t.Helper()

		harness := phase4storetest.StartStore(t, "phase4-u-4-07-"+suffix)
		store := NewStore(harness.DB)
		actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u407-"+suffix+"@example.test", "U407 "+suffix, "U407Phase4Pass1!", false, false, true)
		incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-07-"+suffix, "IR-U407-"+suffix, "Phase 4 U-4-07 "+suffix)
		return harness, store, actor, incident.ID
	}

	t.Run("indicator create dedupes canonically within one incident and isolates incidents", func(t *testing.T) {
		harness, store, actor, incidentID := startFixture(t, "dedupe")
		createOne, err := store.CreateIndicatorRow(context.Background(), actor, incidentID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-07-indicator-1",
			Values: map[string]string{
				"indicator.indicator_type":   golden.Phase4IndicatorExamples[0].IndicatorType,
				"indicator.value_kind":       golden.Phase4IndicatorExamples[0].ValueKind,
				"indicator.display_value":    golden.Phase4IndicatorExamples[0].DisplayValue,
				"indicator.normalized_value": golden.Phase4IndicatorExamples[0].NormalizedValue,
				"indicator.defanged_value":   golden.Phase4IndicatorExamples[0].DefangedValue,
				"indicator.stix_pattern":     golden.Phase4IndicatorExamples[0].STIXPattern,
			},
		}, []byte("txn-phase4-u-4-07-indicator-1"), "req-indicator-1", golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("create indicator one: %v", err)
		}
		createReplay, err := store.CreateIndicatorRow(context.Background(), actor, incidentID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-07-indicator-2",
			Values: map[string]string{
				"indicator.indicator_type":   golden.Phase4IndicatorExamples[0].IndicatorType,
				"indicator.value_kind":       golden.Phase4IndicatorExamples[0].ValueKind,
				"indicator.display_value":    golden.Phase4IndicatorExamples[0].DisplayValue,
				"indicator.normalized_value": golden.Phase4IndicatorExamples[0].NormalizedValue,
				"indicator.defanged_value":   "203(.)0(.)113(.)24",
				"indicator.stix_pattern":     "[ipv4-addr:value = '203.0.113.24']",
			},
		}, []byte("txn-phase4-u-4-07-indicator-2"), "req-indicator-2", golden.Phase4BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("create indicator replay: %v", err)
		}
		if createReplay.RecordID != createOne.RecordID || createReplay.StatusCode != 200 {
			t.Fatalf("expected canonical dedupe within an incident, got %#v %#v", createOne, createReplay)
		}

		otherActor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u407-other@example.test", "U407 Other", "U407OtherPhase4Pass1!", false, false, true)
		otherIncident := phase4storetest.CreateIncidentInStore(t, harness.DB, otherActor, "txn-phase4-u-4-07-other-incident", "IR-U407-B", "Phase 4 U-4-07 other")
		otherCreate, err := store.CreateIndicatorRow(context.Background(), otherActor, otherIncident.ID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-07-other-indicator",
			Values: map[string]string{
				"indicator.indicator_type":   golden.Phase4IndicatorExamples[0].IndicatorType,
				"indicator.value_kind":       golden.Phase4IndicatorExamples[0].ValueKind,
				"indicator.display_value":    golden.Phase4IndicatorExamples[0].DisplayValue,
				"indicator.normalized_value": golden.Phase4IndicatorExamples[0].NormalizedValue,
			},
		}, []byte("txn-phase4-u-4-07-other-indicator"), "req-other-indicator", golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("create other-incident indicator: %v", err)
		}
		if otherCreate.RecordID == createOne.RecordID {
			t.Fatalf("expected incident-scoped canonical identity, got reused record %#v", otherCreate)
		}
	})

	t.Run("source-bound observations stay distinct and drive projection roll-up", func(t *testing.T) {
		harness, store, actor, incidentID := startFixture(t, "observations")
		created, err := store.CreateIndicatorRow(context.Background(), actor, incidentID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-07-observation-create",
			Values: map[string]string{
				"indicator.indicator_type":   golden.Phase4IndicatorExamples[0].IndicatorType,
				"indicator.value_kind":       golden.Phase4IndicatorExamples[0].ValueKind,
				"indicator.display_value":    golden.Phase4IndicatorExamples[0].DisplayValue,
				"indicator.normalized_value": golden.Phase4IndicatorExamples[0].NormalizedValue,
			},
		}, []byte("txn-phase4-u-4-07-observation-create"), "req-observation-create", golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("create indicator for observations: %v", err)
		}
		phase4storetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, golden.Phase4TimelineRecordID)
		phase4storetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, golden.Phase4TimelineSiblingRecordID)

		observationOne, _, err := store.CreateIndicatorObservation(context.Background(), actor, IndicatorObservationCreateParams{
			IncidentID:     incidentID,
			SourceRecordID: golden.Phase4TimelineRecordID,
			SourceFieldKey: golden.Phase4FieldTimelineSourceText,
			OriginKind:     "interactive_cell",
			OriginLocator:  "view:timeline/record:1/cell:timeline.raw_activity_text/span:12-24",
			ObservedText:   "203[.]0[.]113[.]24",
			CreatedAt:      golden.Phase4PastTime,
		})
		if err != nil {
			t.Fatalf("create observation one: %v", err)
		}
		observationTwo, _, err := store.CreateIndicatorObservation(context.Background(), actor, IndicatorObservationCreateParams{
			IncidentID:     incidentID,
			SourceRecordID: golden.Phase4TimelineSiblingRecordID,
			SourceFieldKey: golden.Phase4FieldTimelineSummary,
			OriginKind:     "interactive_cell",
			OriginLocator:  "view:timeline/record:2/cell:timeline.activity_synopsis_text/span:5-17",
			ObservedText:   "203[.]0[.]113[.]24",
			CreatedAt:      golden.Phase4BaseTime,
		})
		if err != nil {
			t.Fatalf("create observation two: %v", err)
		}
		if observationOne.ObservationID == observationTwo.ObservationID {
			t.Fatalf("expected repeated same-value observations to stay distinct, got %#v %#v", observationOne, observationTwo)
		}

		if _, _, err := store.ResolveIndicatorObservation(context.Background(), actor, IndicatorObservationResolveParams{
			ObservationID:             observationOne.ObservationID,
			ResolvedIndicatorRecordID: created.RecordID,
			ResolvedAt:                golden.Phase4BaseTime,
		}); err != nil {
			t.Fatalf("resolve observation one: %v", err)
		}
		if _, _, err := store.ResolveIndicatorObservation(context.Background(), actor, IndicatorObservationResolveParams{
			ObservationID:             observationTwo.ObservationID,
			ResolvedIndicatorRecordID: created.RecordID,
			ResolvedAt:                golden.Phase4BaseTime.Add(2 * time.Minute),
		}); err != nil {
			t.Fatalf("resolve observation two: %v", err)
		}

		projected := phase4storetest.LookupIndicatorProjection(t, harness.DB, created.RecordID)
		if projected.ObservationCount != 2 {
			t.Fatalf("expected observation_count=2, got %#v", projected)
		}
		if projected.FirstObservedAt == nil || !projected.FirstObservedAt.UTC().Equal(golden.Phase4PastTime) {
			t.Fatalf("expected first_observed_at=%s, got %#v", golden.Phase4PastTime, projected)
		}
		if projected.LastObservedAt == nil || !projected.LastObservedAt.UTC().Equal(golden.Phase4BaseTime) {
			t.Fatalf("expected last_observed_at=%s, got %#v", golden.Phase4BaseTime, projected)
		}
	})

	t.Run("lifecycle intervals stay separate from observation-derived timestamps", func(t *testing.T) {
		harness, store, actor, incidentID := startFixture(t, "lifecycle")
		created, err := store.CreateIndicatorRow(context.Background(), actor, incidentID, CreateRequest{
			ClientTxnID: "txn-phase4-u-4-07-lifecycle-create",
			Values: map[string]string{
				"indicator.indicator_type":   golden.Phase4IndicatorExamples[0].IndicatorType,
				"indicator.value_kind":       golden.Phase4IndicatorExamples[0].ValueKind,
				"indicator.display_value":    golden.Phase4IndicatorExamples[0].DisplayValue,
				"indicator.normalized_value": golden.Phase4IndicatorExamples[0].NormalizedValue,
			},
		}, []byte("txn-phase4-u-4-07-lifecycle-create"), "req-lifecycle-create", golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("create indicator for lifecycle: %v", err)
		}

		interval, _, err := store.AppendIndicatorLifecycleInterval(context.Background(), actor, IndicatorLifecycleAppendParams{
			IncidentID:        incidentID,
			IndicatorRecordID: created.RecordID,
			LifecycleState:    "active",
			ValidFrom:         golden.Phase4PastTime,
			CreatedAt:         golden.Phase4PastTime,
		})
		if err != nil {
			t.Fatalf("append lifecycle interval: %v", err)
		}
		phase4storetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, golden.Phase4TimelineMixedRecordID)
		if _, _, err := store.CreateIndicatorObservation(context.Background(), actor, IndicatorObservationCreateParams{
			IncidentID:                incidentID,
			SourceRecordID:            golden.Phase4TimelineMixedRecordID,
			SourceFieldKey:            golden.Phase4FieldTimelineSourceText,
			OriginKind:                "interactive_cell",
			OriginLocator:             "view:timeline/record:3/cell:timeline.raw_activity_text/span:1-9",
			ObservedText:              "203[.]0[.]113[.]24",
			ResolvedIndicatorRecordID: &created.RecordID,
			CreatedAt:                 golden.Phase4BaseTime,
		}); err != nil {
			t.Fatalf("create resolved observation: %v", err)
		}

		projected := phase4storetest.LookupIndicatorProjection(t, harness.DB, created.RecordID)
		if projected.LifecycleSummary == nil || *projected.LifecycleSummary != "active" {
			t.Fatalf("expected lifecycle_summary=active, got %#v", projected)
		}
		if projected.FirstObservedAt == nil || !projected.FirstObservedAt.UTC().Equal(golden.Phase4BaseTime) {
			t.Fatalf("expected observation timestamps to derive from observations, got %#v", projected)
		}

		intervalRow := phase4storetest.LookupIndicatorLifecycleInterval(t, harness.DB, interval.IntervalID)
		if !intervalRow.ValidFrom.UTC().Equal(golden.Phase4PastTime) {
			t.Fatalf("expected lifecycle valid_from to remain distinct from observation timestamps, got %#v", intervalRow)
		}
	})
}

func pgxTxOptions() pgx.TxOptions {
	return pgx.TxOptions{}
}

func requireEntityRow(t testing.TB, rows []map[string]any, recordID uuid.UUID) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("expected row record_id=%s in %#v", recordID, rows)
	return nil
}

func collectionItemsFromEntityRow(t testing.TB, row map[string]any, fieldKey string) []map[string]any {
	t.Helper()
	cells := row["cells"].(map[string]any)
	cell := cells[fieldKey].(map[string]any)
	value := cell["value"].(map[string]any)
	switch rawItems := value["items"].(type) {
	case []map[string]any:
		items := make([]map[string]any, 0, len(rawItems))
		items = append(items, rawItems...)
		return items
	case []any:
		items := make([]map[string]any, 0, len(rawItems))
		for _, rawItem := range rawItems {
			items = append(items, rawItem.(map[string]any))
		}
		return items
	default:
		t.Fatalf("unexpected collection item payload for %s: %#v", fieldKey, value["items"])
	}
	return nil
}

func requireReusableIdentifierItem(t testing.TB, row map[string]any, fieldKey string, identifierClass string, rawValue string) map[string]any {
	t.Helper()
	normalized, ok := fieldnorm.NormalizeIdentifier(identifierClass, rawValue)
	if !ok {
		t.Fatalf("test raw value %q does not normalize for %s", rawValue, identifierClass)
	}
	for _, item := range collectionItemsFromEntityRow(t, row, fieldKey) {
		if item["identifier_class"] == identifierClass && item["normalized_value"] == normalized {
			if item["item_kind"] != "reusable_identifier" {
				t.Fatalf("expected reusable identifier item kind, got %#v", item)
			}
			if item["raw_value"] != rawValue {
				t.Fatalf("expected raw_value=%q, got %#v", rawValue, item)
			}
			if item["item_ref"] == "" {
				t.Fatalf("expected reusable identifier item_ref, got %#v", item)
			}
			return item
		}
	}
	t.Fatalf("expected reusable identifier class=%s normalized=%s in %s, got %#v", identifierClass, normalized, fieldKey, collectionItemsFromEntityRow(t, row, fieldKey))
	return nil
}

func requireNoReusableIdentifierItem(t testing.TB, row map[string]any, fieldKey string, identifierClass string, rawValue string) {
	t.Helper()
	normalized, ok := fieldnorm.NormalizeIdentifier(identifierClass, rawValue)
	if !ok {
		t.Fatalf("test raw value %q does not normalize for %s", rawValue, identifierClass)
	}
	for _, item := range collectionItemsFromEntityRow(t, row, fieldKey) {
		if item["identifier_class"] == identifierClass && item["normalized_value"] == normalized {
			t.Fatalf("did not expect reusable identifier class=%s normalized=%s in %s, got %#v", identifierClass, normalized, fieldKey, item)
		}
	}
}
