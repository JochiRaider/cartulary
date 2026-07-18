package entities_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/assertx"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/golden"
	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	timeline "github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func mustDefaultQueryMeta(t testing.TB, viewSchemaID string) viewschema.QueryMeta {
	t.Helper()
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("view schema %q not registered", viewSchemaID)
	}
	return schema.DefaultQueryMeta()
}

func requireTimelineMutationAfterRowMatchesQuery(t testing.TB, db postgres.DB, changeSetID uuid.UUID, queryRow map[string]any) {
	t.Helper()

	var rawAfterValue []byte
	if err := db.QueryRow(context.Background(), `
SELECT after_value
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND target_kind = 'timeline_record'
   AND operation_kind = 'patch'
 ORDER BY sequence_no DESC
 LIMIT 1
`, changeSetID).Scan(&rawAfterValue); err != nil {
		t.Fatalf("query timeline mutation after row: %v", err)
	}
	var mutationRow map[string]any
	if err := json.Unmarshal(rawAfterValue, &mutationRow); err != nil {
		t.Fatalf("decode timeline mutation after row: %v", err)
	}
	normalizedQueryRow := normalizeJSONMap(t, queryRow)
	if !reflect.DeepEqual(mutationRow, normalizedQueryRow) {
		t.Fatalf("timeline lifecycle mutation row drifted from query row:\nmutation=%#v\nquery=%#v", mutationRow, normalizedQueryRow)
	}
}

func normalizeJSONMap(t testing.TB, value map[string]any) map[string]any {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode row for normalization: %v", err)
	}
	var normalized map[string]any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		t.Fatalf("decode row for normalization: %v", err)
	}
	return normalized
}

// entity-storage / REQ-02-034, REQ-02-038, REQ-02-054..REQ-02-055 / AC-020, AC-021, AC-186.
func TestCreateFromMention_Unit(t *testing.T) {
	t.Run("explicit host resolve links the selected mention to the canonical record", func(t *testing.T) {
		harness := recordstoretest.StartStore(t, "entity_linking-u-4-03-host-reuse")
		mentionStore := mentions.NewStore(harness.DB)
		actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u403-host@example.test", "U403 Host", "U403HostEntityLinkingPass1!", false, false, true)
		incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-03-host-incident", "IR-U403-H", "Record relationships entity-storage host")

		recordstoretest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		recordstoretest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordTimelineRecordID)
		recordstoretest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordTimelineSiblingRecordID)
		recordstoretest.SeedMention(t, harness.DB, actor.ID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordFieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)
		recordstoretest.SeedMention(t, harness.DB, actor.ID, golden.RecordResolvedHostMentionID, golden.RecordTimelineSiblingRecordID, golden.RecordFieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)

		tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		targetID := golden.RecordCanonicalHostRecordID
		result, err := mentionStore.ResolveExistingFromMentionTx(context.Background(), tx, actor, golden.RecordTimelineRecordID, golden.RecordFieldTimelineHostRefs, golden.RecordHostMentionID, &targetID, golden.RecordBaseTime)
		if err != nil {
			t.Fatalf("resolve mention: %v", err)
		}
		if err := tx.Commit(context.Background()); err != nil {
			t.Fatalf("commit tx: %v", err)
		}

		if result.RecordID != golden.RecordCanonicalHostRecordID || result.EntityType != "host" {
			t.Fatalf("expected selected mention to reuse canonical host, got %#v", result)
		}
		selected := recordstoretest.LookupMention(t, harness.DB, golden.RecordHostMentionID)
		assertx.RequireMentionStatus(t, selected, golden.RecordMentionStatusResolved)
		if selected.ResolvedRecordID == nil || *selected.ResolvedRecordID != golden.RecordCanonicalHostRecordID {
			t.Fatalf("expected selected mention to resolve to canonical host, got %#v", selected)
		}
		if selected.RawText != "WS-023" {
			t.Fatalf("expected selected raw text preservation, got %#v", selected)
		}

		sibling := recordstoretest.LookupMention(t, harness.DB, golden.RecordResolvedHostMentionID)
		assertx.RequireMentionStatus(t, sibling, golden.RecordMentionStatusUnresolved)
		if sibling.ResolvedRecordID != nil {
			t.Fatalf("expected sibling mention to remain unresolved, got %#v", sibling)
		}
		if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 1 {
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
`, golden.RecordCanonicalHostRecordID).Scan(&projectedDisplayName, &projectedHostname, &projectedState); err != nil {
			t.Fatalf("lookup host projection after explicit resolve: %v", err)
		}
		if projectedDisplayName != "WS-023" || projectedHostname != "WS-023" || projectedState != "canonical" {
			t.Fatalf("unexpected host projection after explicit resolve: display=%q hostname=%q state=%q", projectedDisplayName, projectedHostname, projectedState)
		}
	})

	t.Run("identity resolve without an explicit target is rejected without creating a stub", func(t *testing.T) {
		harness := recordstoretest.StartStore(t, "entity_linking-u-4-03-identity-create")
		mentionStore := mentions.NewStore(harness.DB)
		actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u403-identity@example.test", "U403 Identity", "U403IdentityEntityLinkingPass1!", false, false, true)
		incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-03-identity-incident", "IR-U403-I", "Record relationships entity-storage identity")

		recordstoretest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordTimelineMixedRecordID)
		recordstoretest.SeedMention(t, harness.DB, actor.ID, golden.RecordIdentityMentionID, golden.RecordTimelineMixedRecordID, golden.RecordFieldTimelineIdentityRefs, "identity", "alex.analyst@example.test", "unresolved", nil, nil)

		tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		_, err = mentionStore.ResolveExistingFromMentionTx(context.Background(), tx, actor, golden.RecordTimelineMixedRecordID, golden.RecordFieldTimelineIdentityRefs, golden.RecordIdentityMentionID, nil, golden.RecordBaseTime)
		if !errors.Is(err, mentions.ErrInvalidMentionResolution) {
			t.Fatalf("expected missing target rejection, got %v", err)
		}

		mention := recordstoretest.LookupMention(t, harness.DB, golden.RecordIdentityMentionID)
		assertx.RequireMentionStatus(t, mention, golden.RecordMentionStatusUnresolved)
		if mention.ResolvedRecordID != nil {
			t.Fatalf("expected mention to remain unresolved, got %#v", mention)
		}
		if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id = $1`, incident.ID); got != 0 {
			t.Fatalf("expected missing-target resolve to create no identity rows, got %d", got)
		}
	})
}

// entity-storage / REQ-02-039..REQ-02-041 / AC-188..AC-190, AC-224, AC-225.
func TestDismissRestoreMentionLifecycle_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "entity_linking-u-4-04")
	mentionStore := mentions.NewStore(harness.DB)
	timelineFacade := timeline.NewFacade(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u404@example.test", "U404", "U404EntityLinkingPass1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-04-incident", "IR-U404", "Record relationships entity-storage")

	recordstoretest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
	normalizedHostToken, ok := fieldnorm.NormalizeMentionToken("WS-023")
	if !ok {
		t.Fatal("normalize mention token")
	}
	summary := "dismiss and restore relationship row"
	createRequest := timeline.CreateRequest{
		ClientTxnID:          "txn-entity_linking-u-4-04-row",
		ActivitySynopsisText: &summary,
		HostRefs: &timeline.CollectionActionPayload{
			Actions: []timeline.CollectionAction{
				{
					Op:             "add_resolved_ref",
					RawText:        "WS-023",
					NormalizedText: normalizedHostToken,
					ResolvedRecord: &golden.RecordCanonicalHostRecordID,
				},
			},
		},
	}
	created, err := timelineFacade.CreateRow(context.Background(), timeline.CreateRowCommand{
		Actor:       actor,
		IncidentID:  incident.ID,
		Request:     createRequest,
		RequestHash: []byte("txn-entity_linking-u-4-04-row"),
		RequestID:   "req-entity_linking-u-4-04-row",
		Now:         golden.RecordBaseTime,
	})
	if err != nil {
		t.Fatalf("create resolved relationship row: %v", err)
	}
	initialRows, err := timelineFacade.QueryTimelineRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query initial timeline rows: %v", err)
	}
	initialRow := recordstoretest.FindRow(t, initialRows, created.RecordID.String())
	initialItem := recordstoretest.RequireSingleCollectionItem(t, initialRow, golden.RecordFieldTimelineHostRefs)
	mentionID := recordstoretest.MentionIDFromItemRef(t, initialItem["item_ref"].(string))

	initialMention := recordstoretest.LookupMention(t, harness.DB, mentionID)
	dismissRequest := mentions.MentionActionRequest{
		BaseMentionRowVersion: initialMention.RowVersion,
		ClientTxnID:           "txn-entity_linking-u-4-04-dismiss",
		Action:                "dismiss_item",
	}
	dismissResult, err := mentionStore.ApplyMentionAction(context.Background(), actor, mentionID, dismissRequest, mentions.MentionActionRequestHash(dismissRequest), "req-entity_linking-u-4-04-dismiss", golden.RecordBaseTime)
	if err != nil {
		t.Fatalf("dismiss mention: %v", err)
	}

	dismissed := recordstoretest.LookupMention(t, harness.DB, mentionID)
	assertx.RequireMentionStatus(t, dismissed, golden.RecordMentionStatusDismissed)
	if dismissed.ResolvedRecordID != nil || dismissed.ResolvedAt != nil || dismissed.ResolutionMethod != nil {
		t.Fatalf("expected dismissed mention to clear resolution metadata, got %#v", dismissed)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incident.ID, created.RecordID, golden.RecordCanonicalHostRecordID); got != 0 {
		t.Fatalf("expected dismiss to remove active derived link, got %d rows", got)
	}
	dismissedRows, err := timelineFacade.QueryTimelineRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query timeline rows after dismiss: %v", err)
	}
	dismissedRow := recordstoretest.FindRow(t, dismissedRows, created.RecordID.String())
	if got := recordstoretest.CollectionItems(t, dismissedRow, golden.RecordFieldTimelineHostRefs); len(got) != 0 {
		t.Fatalf("dismissed mention must be excluded from current relationship-cell values, got %#v", got)
	}
	requireTimelineMutationAfterRowMatchesQuery(t, harness.DB, dismissResult.ChangeSetID, dismissedRow)

	restoreRequest := mentions.MentionActionRequest{
		BaseMentionRowVersion: dismissed.RowVersion,
		ClientTxnID:           "txn-entity_linking-u-4-04-restore",
		Action:                "revert_to_unresolved",
	}
	restoreResult, err := mentionStore.ApplyMentionAction(context.Background(), actor, mentionID, restoreRequest, mentions.MentionActionRequestHash(restoreRequest), "req-entity_linking-u-4-04-restore", golden.RecordBaseTime.Add(time.Minute))
	if err != nil {
		t.Fatalf("restore mention: %v", err)
	}

	restored := recordstoretest.LookupMention(t, harness.DB, mentionID)
	assertx.RequireMentionStatus(t, restored, golden.RecordMentionStatusUnresolved)
	if restored.RawText != "WS-023" || restored.ResolvedRecordID != nil || restored.ResolutionMethod != nil {
		t.Fatalf("expected durable restore-to-unresolved semantics, got %#v", restored)
	}
	restoredRows, err := timelineFacade.QueryTimelineRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query timeline rows after restore: %v", err)
	}
	restoredRow := recordstoretest.FindRow(t, restoredRows, created.RecordID.String())
	restoredItem := recordstoretest.RequireSingleCollectionItem(t, restoredRow, golden.RecordFieldTimelineHostRefs)
	if restoredItem["item_kind"] != "unresolved_mention" || restoredItem["raw_text"] != "WS-023" {
		t.Fatalf("restore must surface the unresolved mention in current-state reads, got %#v", restoredItem)
	}
	requireTimelineMutationAfterRowMatchesQuery(t, harness.DB, restoreResult.ChangeSetID, restoredRow)
	if _, ok := restoredItem["resolved_record_id"]; ok {
		t.Fatalf("restore must not silently relink the historical target, got %#v", restoredItem)
	}
}

// entity-storage / REQ-02-059..REQ-02-063 / AC-021, AC-022.
func TestExactMatchPrecedence_Unit(t *testing.T) {
	nullableString := func(value string) any {
		if value == "" {
			return nil
		}
		return value
	}
	startFixture := func(t *testing.T, suffix string) (*recordstoretest.StoreHarness, *hostidentity.Store, authn.UserRecord, uuid.UUID) {
		t.Helper()

		harness := recordstoretest.StartStore(t, "entity_linking-u-4-05-"+suffix)
		store := hostidentity.NewStore(harness.DB)
		actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u405-"+suffix+"@example.test", "U405 "+suffix, "U405EntityLinkingPass1!", false, false, true)
		incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-05-"+suffix, "IR-U405-"+suffix, "Record relationships entity-storage "+suffix)
		return harness, store, actor, incident.ID
	}
	seedHost := func(t *testing.T, harness *recordstoretest.StoreHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string, aadDeviceID string, fqdn string, hostname string) {
		t.Helper()

		recordstoretest.SeedHostRecord(t, harness.DB, incidentID, actorID, recordID, displayName, "seed-hostname", "seed.example.test", "")
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
	seedIdentity := func(t *testing.T, harness *recordstoretest.StoreHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string, aadObjectID string, sid string, upn string, email string, samAccountName string) {
		t.Helper()

		recordstoretest.SeedIdentityRecord(t, harness.DB, incidentID, actorID, recordID, displayName, "seed-upn@example.test", "seed-email@example.test", "SEEDSAM")
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
				}, []byte("txn-entity_linking-u-4-05-"+tc.suffix), "req-"+tc.suffix, golden.RecordBaseTime)
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

				reuse, err := store.CreateIdentityRow(context.Background(), actor, incidentID, hostidentity.CreateRequest{
					ClientTxnID: "txn-entity_linking-u-4-05-" + tc.suffix,
					Values:      tc.values,
				}, []byte("txn-entity_linking-u-4-05-"+tc.suffix), "req-"+tc.suffix, golden.RecordBaseTime.Add(2*time.Minute))
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
			recordstoretest.SeedEntityAlias(t, harness.DB, incidentID, actor.ID, hostAliasRecordID, "host", "Workstation 23")

			hostAliasOnly, err := store.CreateHostRow(context.Background(), actor, incidentID, hostidentity.CreateRequest{
				ClientTxnID: "txn-entity_linking-u-4-05-host-alias",
				Values: map[string]string{
					"host.display_name": "Workstation 23",
				},
			}, []byte("txn-entity_linking-u-4-05-host-alias"), "req-host-alias", golden.RecordBaseTime.Add(time.Minute))
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
			recordstoretest.SeedEntityAlias(t, harness.DB, incidentID, actor.ID, identityAliasRecordID, "identity", "Case Owner")

			identityFuzzyNonMatch, err := store.CreateIdentityRow(context.Background(), actor, incidentID, hostidentity.CreateRequest{
				ClientTxnID: "txn-entity_linking-u-4-05-identity-fuzzy",
				Values: map[string]string{
					"identity.display_name": "Case Ownr",
				},
			}, []byte("txn-entity_linking-u-4-05-identity-fuzzy"), "req-identity-fuzzy", golden.RecordBaseTime.Add(3*time.Minute))
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
func TestExplicitEntityMerge_Unit(t *testing.T) {
	t.Run("host merge preserves raw mentions, loser lineage, and survivor reuse", func(t *testing.T) {
		harness := recordstoretest.StartStore(t, "entity_linking-u-4-06-host")
		store := hostidentity.NewStore(harness.DB)
		actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u406@example.test", "U406", "U406EntityLinkingPass1!", false, false, true)
		incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-incident", "IR-U406", "Record relationships entity-storage")

		recordstoretest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		recordstoretest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordDuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
		recordstoretest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, golden.RecordDuplicateHostRecordID, "host", "Workstation 23")
		recordstoretest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordTimelineRecordID)
		outgoingTargetID := uuid.New()
		outgoingLinkID := uuid.New()
		recordstoretest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, outgoingTargetID)
		recordstoretest.SeedResolvedMention(t, harness.DB, actor.ID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordDuplicateHostRecordID, golden.RecordFieldTimelineHostRefs, "host", "WS-023")
		recordstoretest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, golden.RecordDuplicateLinkID, golden.RecordTimelineRecordID, golden.RecordDuplicateHostRecordID, "observed_on_host", "manual", nil)
		recordstoretest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, outgoingLinkID, golden.RecordDuplicateHostRecordID, outgoingTargetID, "references_record", "manual", nil)
		recordstoretest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, golden.RecordTagIDSurvivor, golden.RecordCanonicalHostRecordID, "critical-host")
		recordstoretest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, golden.RecordTagIDLoser, golden.RecordDuplicateHostRecordID, "critical-host")
		recordstoretest.SeedAssessment(t, harness.DB, incident.ID, actor.ID, golden.RecordAssessmentHostID, golden.RecordDuplicateHostRecordID, "host", "confirmed")
		beforeMention := recordstoretest.LookupMention(t, harness.DB, golden.RecordHostMentionID)

		result, err := merge.NewStore(harness.DB).MergeEntity(context.Background(), actor, golden.RecordCanonicalHostRecordID, merge.MergeRequest{
			LoserRecordID:          golden.RecordDuplicateHostRecordID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-merge",
		}, []byte("txn-entity_linking-u-4-06-merge"), "req-merge", golden.RecordBaseTime)
		if err != nil {
			t.Fatalf("merge entity: %v", err)
		}
		if result.SurvivorRecordID != golden.RecordCanonicalHostRecordID || result.LoserRecordID != golden.RecordDuplicateHostRecordID {
			t.Fatalf("unexpected merge result: %#v", result)
		}
		if result.MergeSummary.RepointedLinkCount != 2 {
			t.Fatalf("expected merge to repoint incoming and outgoing links, got summary %#v", result.MergeSummary)
		}

		mention := recordstoretest.LookupMention(t, harness.DB, golden.RecordHostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.RecordMentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.RecordCanonicalHostRecordID {
			t.Fatalf("expected merge to repoint mention resolution to survivor, got %#v", mention)
		}
		assertx.RequireRawTextPreserved(t, beforeMention.RawText, mention.RawText)
		link := recordstoretest.LookupActiveLink(t, harness.DB, incident.ID, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(t, link, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, "observed_on_host", "manual", nil)
		outgoing := recordstoretest.LookupActiveLink(t, harness.DB, incident.ID, golden.RecordCanonicalHostRecordID, outgoingTargetID, "references_record")
		assertx.RequireActiveLink(t, outgoing, golden.RecordCanonicalHostRecordID, outgoingTargetID, "references_record", "manual", nil)
		if got := recordstoretest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, outgoingLinkID); got != 0 {
			t.Fatalf("expected loser-sourced active link to disappear, got %d rows", got)
		}
		if got := recordstoretest.LookupAssessmentSubject(t, harness.DB, golden.RecordAssessmentHostID); got != golden.RecordCanonicalHostRecordID {
			t.Fatalf("expected assessment repoint to survivor, got %s", got)
		}
		if got := recordstoretest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incident.ID, golden.RecordCanonicalHostRecordID); got != 1 {
			t.Fatalf("expected one active deduped survivor tag, got %d", got)
		}
		if got := recordstoretest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, golden.RecordDuplicateLinkID); got != 0 {
			t.Fatalf("expected loser-targeted active link to disappear, got %d rows", got)
		}
		state, mergedInto, rowVersion, _ := recordstoretest.LookupHostState(t, harness.DB, golden.RecordDuplicateHostRecordID)
		if state != "merged" || mergedInto == nil || *mergedInto != golden.RecordCanonicalHostRecordID || rowVersion != 2 {
			t.Fatalf("expected loser host lineage state after merge, got state=%s merged_into=%v row_version=%d", state, mergedInto, rowVersion)
		}

		reuse, err := store.CreateHostRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-reuse",
			Values: map[string]string{
				"host.fqdn": "ws-023.corp.example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-reuse"), "req-reuse", golden.RecordBaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge exact-match reuse: %v", err)
		}
		if reuse.RecordID != golden.RecordCanonicalHostRecordID {
			t.Fatalf("expected exact-match reuse to carry forward to survivor, got %#v", reuse)
		}
	})

	t.Run("identity merge preserves raw mentions, loser lineage, and survivor reuse", func(t *testing.T) {
		harness := recordstoretest.StartStore(t, "entity_linking-u-4-06-identity")
		store := hostidentity.NewStore(harness.DB)
		actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u406-identity@example.test", "U406 Identity", "U406IdentityEntityLinkingPass1!", false, false, true)
		incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-identity-incident", "IR-U406-I", "Record relationships entity-storage identity")

		recordstoretest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordCanonicalIdentityID, "Alex Analyst", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		recordstoretest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordDuplicateIdentityID, "Alex Duplicate", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		recordstoretest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, golden.RecordDuplicateIdentityID, "identity", "Case Owner")
		recordstoretest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordTimelineRecordID)
		recordstoretest.SeedResolvedMention(t, harness.DB, actor.ID, golden.RecordIdentityMentionID, golden.RecordTimelineRecordID, golden.RecordDuplicateIdentityID, golden.RecordFieldTimelineIdentityRefs, "identity", "Case Owner")
		recordstoretest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, golden.RecordDuplicateLinkID, golden.RecordTimelineRecordID, golden.RecordDuplicateIdentityID, "observed_as_identity", "manual", nil)
		recordstoretest.SeedAssessment(t, harness.DB, incident.ID, actor.ID, golden.RecordAssessmentIdentID, golden.RecordDuplicateIdentityID, "identity", "confirmed")
		beforeMention := recordstoretest.LookupMention(t, harness.DB, golden.RecordIdentityMentionID)

		result, err := merge.NewStore(harness.DB).MergeEntity(context.Background(), actor, golden.RecordCanonicalIdentityID, merge.MergeRequest{
			LoserRecordID:          golden.RecordDuplicateIdentityID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-identity-merge",
		}, []byte("txn-entity_linking-u-4-06-identity-merge"), "req-identity-merge", golden.RecordBaseTime)
		if err != nil {
			t.Fatalf("merge identity: %v", err)
		}
		if result.SurvivorRecordID != golden.RecordCanonicalIdentityID || result.LoserRecordID != golden.RecordDuplicateIdentityID {
			t.Fatalf("unexpected identity merge result: %#v", result)
		}

		mention := recordstoretest.LookupMention(t, harness.DB, golden.RecordIdentityMentionID)
		assertx.RequireMentionStatus(t, mention, golden.RecordMentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.RecordCanonicalIdentityID {
			t.Fatalf("expected identity merge to repoint mention resolution to survivor, got %#v", mention)
		}
		assertx.RequireRawTextPreserved(t, beforeMention.RawText, mention.RawText)
		link := recordstoretest.LookupActiveLink(t, harness.DB, incident.ID, golden.RecordTimelineRecordID, golden.RecordCanonicalIdentityID, "observed_as_identity")
		assertx.RequireActiveLink(t, link, golden.RecordTimelineRecordID, golden.RecordCanonicalIdentityID, "observed_as_identity", "manual", nil)
		if got := recordstoretest.LookupAssessmentSubject(t, harness.DB, golden.RecordAssessmentIdentID); got != golden.RecordCanonicalIdentityID {
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
`, golden.RecordDuplicateIdentityID).Scan(&state, &mergedIntoRaw, &rowVersion); err != nil {
			t.Fatalf("lookup loser identity state: %v", err)
		}
		if state != "merged" || !mergedIntoRaw.Valid || mergedIntoRaw.String != golden.RecordCanonicalIdentityID.String() || rowVersion != 2 {
			t.Fatalf("expected loser identity lineage after merge, got state=%s merged_into=%v row_version=%d", state, mergedIntoRaw, rowVersion)
		}

		reuse, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-identity-reuse",
			Values: map[string]string{
				"identity.email": "alex.analyst@example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-identity-reuse"), "req-identity-reuse", golden.RecordBaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge identity exact-match reuse: %v", err)
		}
		if reuse.RecordID != golden.RecordCanonicalIdentityID {
			t.Fatalf("expected identity exact-match reuse to carry forward to survivor, got %#v", reuse)
		}
	})

	t.Run("host merge exposes carried secondary reusable identifiers", func(t *testing.T) {
		harness := recordstoretest.StartStore(t, "entity_linking-u-4-06-host-reusable-row")
		store := hostidentity.NewStore(harness.DB)
		actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u406-host-reusable@example.test", "U406 Host Reusable", "U406HostReusableEntityLinkingPass1!", false, false, true)
		incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-host-reusable-incident", "IR-U406-HR", "Record relationships entity-storage host reusable rows")
		survivorID := uuid.New()
		loserID := uuid.New()

		recordstoretest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, survivorID, "WS-023", "WS-023", "ws-023.current.example.test", "")
		recordstoretest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, loserID, "Legacy WS-023", "LEGACY-WS-023", "legacy-ws-023.example.test", "")

		if _, err := merge.NewStore(harness.DB).MergeEntity(context.Background(), actor, survivorID, merge.MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-host-reusable-merge",
		}, []byte("txn-entity_linking-u-4-06-host-reusable-merge"), "req-host-reusable-merge", golden.RecordBaseTime); err != nil {
			t.Fatalf("merge host reusable identifiers: %v", err)
		}

		rows, err := store.QueryHostRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, hostidentity.HostsViewSchemaID))
		if err != nil {
			t.Fatalf("query host rows after reusable merge: %v", err)
		}
		survivorRow := requireEntityRow(t, rows, survivorID)
		requireReusableIdentifierItem(t, survivorRow, "host.reusable_identifiers", "fqdn", "legacy-ws-023.example.test")
		requireNoReusableIdentifierItem(t, survivorRow, "host.reusable_identifiers", "fqdn", "ws-023.current.example.test")

		reuse, err := store.CreateHostRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-host-reusable-create",
			Values: map[string]string{
				"host.fqdn": "legacy-ws-023.example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-host-reusable-create"), "req-host-reusable-create", golden.RecordBaseTime.Add(time.Minute))
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
		harness := recordstoretest.StartStore(t, "entity_linking-u-4-06-identity-reusable-row")
		store := hostidentity.NewStore(harness.DB)
		actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u406-identity-reusable@example.test", "U406 Identity Reusable", "U406IdentityReusableEntityLinkingPass1!", false, false, true)
		incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-identity-reusable-incident", "IR-U406-IR", "Record relationships entity-storage identity reusable rows")
		survivorID := uuid.New()
		loserID := uuid.New()

		recordstoretest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, survivorID, "Alex Survivor", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		recordstoretest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, loserID, "Alex Analyst Legacy", "alex.legacy@example.test", "alex.legacy@example.test", "ALEXLEGACY")

		if _, err := merge.NewStore(harness.DB).MergeEntity(context.Background(), actor, survivorID, merge.MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-identity-reusable-merge",
		}, []byte("txn-entity_linking-u-4-06-identity-reusable-merge"), "req-identity-reusable-merge", golden.RecordBaseTime); err != nil {
			t.Fatalf("merge identity reusable identifiers: %v", err)
		}

		rows, err := store.QueryIdentityRows(context.Background(), incident.ID, mustDefaultQueryMeta(t, hostidentity.IdentitiesViewSchemaID))
		if err != nil {
			t.Fatalf("query identity rows after reusable merge: %v", err)
		}
		survivorRow := requireEntityRow(t, rows, survivorID)
		requireReusableIdentifierItem(t, survivorRow, "identity.reusable_identifiers", "email", "alex.legacy@example.test")
		requireNoReusableIdentifierItem(t, survivorRow, "identity.reusable_identifiers", "email", "alex.survivor@example.test")

		reuse, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-identity-reusable-create",
			Values: map[string]string{
				"identity.email": "alex.legacy@example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-identity-reusable-create"), "req-identity-reusable-create", golden.RecordBaseTime.Add(time.Minute))
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
