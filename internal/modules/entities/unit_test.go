package entities_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/jackc/pgx/v5"

	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timeline "github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func newEntityTestTimelineBundle(t testing.TB, pool postgres.DB) *timelineassembly.Bundle {
	t.Helper()
	revisionComposition := revisionsupport.MustComposition(t)
	return timelineassembly.NewBundle(
		pool,
		conflicttest.NewCodec("timeline"),
		revisionComposition.Runtime.Appender(),
		revisionComposition.Intents,
		evidence.NewTimelineAttachmentContribution(pool),
	)
}

func newEntityTestStore(t testing.TB, pool postgres.DB) *hostidentity.Store {
	t.Helper()
	return hostidentity.NewStore(pool, revisionsupport.MustAppender(t), nil)
}

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
		harness := appsupport.StartStore(t, "entity_linking-u-4-03-host-reuse")
		mentionStore := newEntityTestTimelineBundle(t, harness.DB).EntityMentionStore
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u403-host@example.test", "U403 Host", "U403HostEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-03-host-incident", "IR-U403-H", "Record relationships entity-storage host")

		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.SiblingRecordID)
		entitytest.SeedMention(t, harness.DB, actor.ID, entitytest.HostMentionID, timelinetest.RecordID, timelinetest.FieldHostRefs, "host", "WS-023", "unresolved", nil, nil)
		entitytest.SeedMention(t, harness.DB, actor.ID, entitytest.ResolvedHostMentionID, timelinetest.SiblingRecordID, timelinetest.FieldHostRefs, "host", "WS-023", "unresolved", nil, nil)

		tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		targetID := entitytest.CanonicalHostRecordID
		result, err := mentionStore.ResolveExistingFromMentionTx(context.Background(), tx, actor, timelinetest.RecordID, timelinetest.FieldHostRefs, entitytest.HostMentionID, &targetID, entitytest.BaseTime)
		if err != nil {
			t.Fatalf("resolve mention: %v", err)
		}
		if err := tx.Commit(context.Background()); err != nil {
			t.Fatalf("commit tx: %v", err)
		}

		if result.RecordID != entitytest.CanonicalHostRecordID || result.EntityType != "host" {
			t.Fatalf("expected selected mention to reuse canonical host, got %#v", result)
		}
		selected := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, selected, entitytest.MentionStatusResolved)
		if selected.ResolvedRecordID == nil || *selected.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected selected mention to resolve to canonical host, got %#v", selected)
		}
		if selected.RawText != "WS-023" {
			t.Fatalf("expected selected raw text preservation, got %#v", selected)
		}

		sibling := entitytest.LookupMention(t, harness.DB, entitytest.ResolvedHostMentionID)
		entitytest.RequireMentionStatus(t, sibling, entitytest.MentionStatusUnresolved)
		if sibling.ResolvedRecordID != nil {
			t.Fatalf("expected sibling mention to remain unresolved, got %#v", sibling)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 1 {
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
`, entitytest.CanonicalHostRecordID).Scan(&projectedDisplayName, &projectedHostname, &projectedState); err != nil {
			t.Fatalf("lookup host projection after explicit resolve: %v", err)
		}
		if projectedDisplayName != "WS-023" || projectedHostname != "WS-023" || projectedState != "canonical" {
			t.Fatalf("unexpected host projection after explicit resolve: display=%q hostname=%q state=%q", projectedDisplayName, projectedHostname, projectedState)
		}
	})

	t.Run("identity resolve without an explicit target is rejected without creating a stub", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-03-identity-create")
		mentionStore := newEntityTestTimelineBundle(t, harness.DB).EntityMentionStore
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u403-identity@example.test", "U403 Identity", "U403IdentityEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-03-identity-incident", "IR-U403-I", "Record relationships entity-storage identity")

		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.MixedRecordID)
		entitytest.SeedMention(t, harness.DB, actor.ID, entitytest.IdentityMentionID, timelinetest.MixedRecordID, timelinetest.FieldIdentityRefs, "identity", "alex.analyst@example.test", "unresolved", nil, nil)

		tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		_, err = mentionStore.ResolveExistingFromMentionTx(context.Background(), tx, actor, timelinetest.MixedRecordID, timelinetest.FieldIdentityRefs, entitytest.IdentityMentionID, nil, entitytest.BaseTime)
		if !errors.Is(err, mentions.ErrInvalidMentionResolution) {
			t.Fatalf("expected missing target rejection, got %v", err)
		}

		mention := entitytest.LookupMention(t, harness.DB, entitytest.IdentityMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusUnresolved)
		if mention.ResolvedRecordID != nil {
			t.Fatalf("expected mention to remain unresolved, got %#v", mention)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id = $1`, incident.ID); got != 0 {
			t.Fatalf("expected missing-target resolve to create no identity rows, got %d", got)
		}
	})
}

// entity-storage / REQ-02-039..REQ-02-041 / AC-188..AC-190, AC-224, AC-225.
func TestDismissRestoreMentionLifecycle_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "entity_linking-u-4-04")
	timelineBundle := newEntityTestTimelineBundle(t, harness.DB)
	mentionStore := timelineBundle.EntityMentionStore
	timelineFacade := timelineBundle.Facade
	timelineProjectionStore := timelineBundle.ProjectionCatalog.Query
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u404@example.test", "U404", "U404EntityLinkingPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-04-incident", "IR-U404", "Record relationships entity-storage")

	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
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
					ResolvedRecord: &entitytest.CanonicalHostRecordID,
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
		Now:         entitytest.BaseTime,
	})
	if err != nil {
		t.Fatalf("create resolved relationship row: %v", err)
	}
	initialRows, err := timelineProjectionStore.QueryRows(context.Background(), incident.ID, timeline.TimelineViewSchemaID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query initial timeline rows: %v", err)
	}
	initialRow := workbookscenariotest.FindRow(t, initialRows, created.RecordID.String())
	initialItem := workbookscenariotest.RequireSingleCollectionItem(t, initialRow, timelinetest.FieldHostRefs)
	mentionID := entitytest.MentionIDFromItemRef(t, initialItem["item_ref"].(string))

	initialMention := entitytest.LookupMention(t, harness.DB, mentionID)
	dismissRequest := mentions.MentionActionRequest{
		BaseMentionRowVersion: initialMention.RowVersion,
		ClientTxnID:           "txn-entity_linking-u-4-04-dismiss",
		Action:                "dismiss_item",
	}
	dismissResult, err := mentionStore.ApplyMentionAction(context.Background(), actor, mentionID, dismissRequest, mentions.MentionActionRequestHash(dismissRequest), "req-entity_linking-u-4-04-dismiss", entitytest.BaseTime)
	if err != nil {
		t.Fatalf("dismiss mention: %v", err)
	}

	dismissed := entitytest.LookupMention(t, harness.DB, mentionID)
	entitytest.RequireMentionStatus(t, dismissed, entitytest.MentionStatusDismissed)
	if dismissed.ResolvedRecordID != nil || dismissed.ResolvedAt != nil || dismissed.ResolutionMethod != nil {
		t.Fatalf("expected dismissed mention to clear resolution metadata, got %#v", dismissed)
	}
	if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incident.ID, created.RecordID, entitytest.CanonicalHostRecordID); got != 0 {
		t.Fatalf("expected dismiss to remove active derived link, got %d rows", got)
	}
	dismissedRows, err := timelineProjectionStore.QueryRows(context.Background(), incident.ID, timeline.TimelineViewSchemaID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query timeline rows after dismiss: %v", err)
	}
	dismissedRow := workbookscenariotest.FindRow(t, dismissedRows, created.RecordID.String())
	if got := workbookscenariotest.CollectionItems(t, dismissedRow, timelinetest.FieldHostRefs); len(got) != 0 {
		t.Fatalf("dismissed mention must be excluded from current relationship-cell values, got %#v", got)
	}
	requireTimelineMutationAfterRowMatchesQuery(t, harness.DB, dismissResult.ChangeSetID, dismissedRow)

	restoreRequest := mentions.MentionActionRequest{
		BaseMentionRowVersion: dismissed.RowVersion,
		ClientTxnID:           "txn-entity_linking-u-4-04-restore",
		Action:                "revert_to_unresolved",
	}
	restoreResult, err := mentionStore.ApplyMentionAction(context.Background(), actor, mentionID, restoreRequest, mentions.MentionActionRequestHash(restoreRequest), "req-entity_linking-u-4-04-restore", entitytest.BaseTime.Add(time.Minute))
	if err != nil {
		t.Fatalf("restore mention: %v", err)
	}

	restored := entitytest.LookupMention(t, harness.DB, mentionID)
	entitytest.RequireMentionStatus(t, restored, entitytest.MentionStatusUnresolved)
	if restored.RawText != "WS-023" || restored.ResolvedRecordID != nil || restored.ResolutionMethod != nil {
		t.Fatalf("expected durable restore-to-unresolved semantics, got %#v", restored)
	}
	restoredRows, err := timelineProjectionStore.QueryRows(context.Background(), incident.ID, timeline.TimelineViewSchemaID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query timeline rows after restore: %v", err)
	}
	restoredRow := workbookscenariotest.FindRow(t, restoredRows, created.RecordID.String())
	restoredItem := workbookscenariotest.RequireSingleCollectionItem(t, restoredRow, timelinetest.FieldHostRefs)
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
				}, []byte("txn-entity_linking-u-4-05-"+tc.suffix), "req-"+tc.suffix, entitytest.BaseTime.Add(2*time.Minute))
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
func TestExplicitEntityMerge_Unit(t *testing.T) {
	t.Run("host merge preserves raw mentions, loser lineage, and survivor reuse", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-06-host")
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u406@example.test", "U406", "U406EntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-incident", "IR-U406", "Record relationships entity-storage")

		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, entitytest.DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
		entitytest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, entitytest.DuplicateHostRecordID, "host", "Workstation 23")
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
		outgoingTargetID := uuid.New()
		outgoingLinkID := uuid.New()
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, outgoingTargetID)
		entitytest.SeedResolvedMention(t, harness.DB, actor.ID, entitytest.HostMentionID, timelinetest.RecordID, entitytest.DuplicateHostRecordID, timelinetest.FieldHostRefs, "host", "WS-023")
		linktest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, linktest.DuplicateLinkID, timelinetest.RecordID, entitytest.DuplicateHostRecordID, "observed_on_host", "manual", nil)
		linktest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, outgoingLinkID, entitytest.DuplicateHostRecordID, outgoingTargetID, "references_record", "manual", nil)
		linktest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, linktest.TagIDSurvivor, entitytest.CanonicalHostRecordID, "critical-host")
		linktest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, linktest.TagIDLoser, entitytest.DuplicateHostRecordID, "critical-host")
		assessmenttest.SeedAssessment(t, harness.DB, incident.ID, actor.ID, assessmenttest.HostAssessmentID, entitytest.DuplicateHostRecordID, "host", "confirmed")
		beforeMention := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)

		result, err := newEntityTestTimelineBundle(t, harness.DB).EntityMergeStore.MergeEntity(context.Background(), actor, entitytest.CanonicalHostRecordID, merge.MergeRequest{
			LoserRecordID:          entitytest.DuplicateHostRecordID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-merge",
		}, []byte("txn-entity_linking-u-4-06-merge"), "req-merge", entitytest.BaseTime)
		if err != nil {
			t.Fatalf("merge entity: %v", err)
		}
		if result.SurvivorRecordID != entitytest.CanonicalHostRecordID || result.LoserRecordID != entitytest.DuplicateHostRecordID {
			t.Fatalf("unexpected merge result: %#v", result)
		}
		if result.MergeSummary.RepointedLinkCount != 2 {
			t.Fatalf("expected merge to repoint incoming and outgoing links, got summary %#v", result.MergeSummary)
		}

		mention := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected merge to repoint mention resolution to survivor, got %#v", mention)
		}
		entitytest.RequireRawTextPreserved(t, beforeMention.RawText, mention.RawText)
		link := linktest.LookupActiveLink(t, harness.DB, incident.ID, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host")
		linktest.RequireActiveLink(t, link, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host", "manual", nil)
		outgoing := linktest.LookupActiveLink(t, harness.DB, incident.ID, entitytest.CanonicalHostRecordID, outgoingTargetID, "references_record")
		linktest.RequireActiveLink(t, outgoing, entitytest.CanonicalHostRecordID, outgoingTargetID, "references_record", "manual", nil)
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, outgoingLinkID); got != 0 {
			t.Fatalf("expected loser-sourced active link to disappear, got %d rows", got)
		}
		if got := assessmenttest.LookupAssessmentSubject(t, harness.DB, assessmenttest.HostAssessmentID); got != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected assessment repoint to survivor, got %s", got)
		}
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incident.ID, entitytest.CanonicalHostRecordID); got != 1 {
			t.Fatalf("expected one active deduped survivor tag, got %d", got)
		}
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, linktest.DuplicateLinkID); got != 0 {
			t.Fatalf("expected loser-targeted active link to disappear, got %d rows", got)
		}
		state, mergedInto, rowVersion, _ := entitytest.LookupHostState(t, harness.DB, entitytest.DuplicateHostRecordID)
		if state != "merged" || mergedInto == nil || *mergedInto != entitytest.CanonicalHostRecordID || rowVersion != 2 {
			t.Fatalf("expected loser host lineage state after merge, got state=%s merged_into=%v row_version=%d", state, mergedInto, rowVersion)
		}

		reuse, err := store.CreateHostRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-reuse",
			Values: map[string]string{
				"host.fqdn": "ws-023.corp.example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-reuse"), "req-reuse", entitytest.BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge exact-match reuse: %v", err)
		}
		if reuse.RecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected exact-match reuse to carry forward to survivor, got %#v", reuse)
		}
	})

	t.Run("identity merge preserves raw mentions, loser lineage, and survivor reuse", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-06-identity")
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u406-identity@example.test", "U406 Identity", "U406IdentityEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-identity-incident", "IR-U406-I", "Record relationships entity-storage identity")

		entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, entitytest.CanonicalIdentityRecordID, "Alex Analyst", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, entitytest.DuplicateIdentityRecordID, "Alex Duplicate", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		entitytest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, entitytest.DuplicateIdentityRecordID, "identity", "Case Owner")
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
		entitytest.SeedResolvedMention(t, harness.DB, actor.ID, entitytest.IdentityMentionID, timelinetest.RecordID, entitytest.DuplicateIdentityRecordID, timelinetest.FieldIdentityRefs, "identity", "Case Owner")
		linktest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, linktest.DuplicateLinkID, timelinetest.RecordID, entitytest.DuplicateIdentityRecordID, "observed_as_identity", "manual", nil)
		assessmenttest.SeedAssessment(t, harness.DB, incident.ID, actor.ID, assessmenttest.IdentityAssessmentID, entitytest.DuplicateIdentityRecordID, "identity", "confirmed")
		beforeMention := entitytest.LookupMention(t, harness.DB, entitytest.IdentityMentionID)

		result, err := newEntityTestTimelineBundle(t, harness.DB).EntityMergeStore.MergeEntity(context.Background(), actor, entitytest.CanonicalIdentityRecordID, merge.MergeRequest{
			LoserRecordID:          entitytest.DuplicateIdentityRecordID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-identity-merge",
		}, []byte("txn-entity_linking-u-4-06-identity-merge"), "req-identity-merge", entitytest.BaseTime)
		if err != nil {
			t.Fatalf("merge identity: %v", err)
		}
		if result.SurvivorRecordID != entitytest.CanonicalIdentityRecordID || result.LoserRecordID != entitytest.DuplicateIdentityRecordID {
			t.Fatalf("unexpected identity merge result: %#v", result)
		}

		mention := entitytest.LookupMention(t, harness.DB, entitytest.IdentityMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != entitytest.CanonicalIdentityRecordID {
			t.Fatalf("expected identity merge to repoint mention resolution to survivor, got %#v", mention)
		}
		entitytest.RequireRawTextPreserved(t, beforeMention.RawText, mention.RawText)
		link := linktest.LookupActiveLink(t, harness.DB, incident.ID, timelinetest.RecordID, entitytest.CanonicalIdentityRecordID, "observed_as_identity")
		linktest.RequireActiveLink(t, link, timelinetest.RecordID, entitytest.CanonicalIdentityRecordID, "observed_as_identity", "manual", nil)
		if got := assessmenttest.LookupAssessmentSubject(t, harness.DB, assessmenttest.IdentityAssessmentID); got != entitytest.CanonicalIdentityRecordID {
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
`, entitytest.DuplicateIdentityRecordID).Scan(&state, &mergedIntoRaw, &rowVersion); err != nil {
			t.Fatalf("lookup loser identity state: %v", err)
		}
		if state != "merged" || !mergedIntoRaw.Valid || mergedIntoRaw.String != entitytest.CanonicalIdentityRecordID.String() || rowVersion != 2 {
			t.Fatalf("expected loser identity lineage after merge, got state=%s merged_into=%v row_version=%d", state, mergedIntoRaw, rowVersion)
		}

		reuse, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-identity-reuse",
			Values: map[string]string{
				"identity.email": "alex.analyst@example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-identity-reuse"), "req-identity-reuse", entitytest.BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge identity exact-match reuse: %v", err)
		}
		if reuse.RecordID != entitytest.CanonicalIdentityRecordID {
			t.Fatalf("expected identity exact-match reuse to carry forward to survivor, got %#v", reuse)
		}
	})

	t.Run("host merge exposes carried secondary reusable identifiers", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-06-host-reusable-row")
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u406-host-reusable@example.test", "U406 Host Reusable", "U406HostReusableEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-host-reusable-incident", "IR-U406-HR", "Record relationships entity-storage host reusable rows")
		survivorID := uuid.New()
		loserID := uuid.New()

		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, survivorID, "WS-023", "WS-023", "ws-023.current.example.test", "")
		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, loserID, "Legacy WS-023", "LEGACY-WS-023", "legacy-ws-023.example.test", "")

		if _, err := newEntityTestTimelineBundle(t, harness.DB).EntityMergeStore.MergeEntity(context.Background(), actor, survivorID, merge.MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-host-reusable-merge",
		}, []byte("txn-entity_linking-u-4-06-host-reusable-merge"), "req-host-reusable-merge", entitytest.BaseTime); err != nil {
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
		}, []byte("txn-entity_linking-u-4-06-host-reusable-create"), "req-host-reusable-create", entitytest.BaseTime.Add(time.Minute))
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
		harness := appsupport.StartStore(t, "entity_linking-u-4-06-identity-reusable-row")
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u406-identity-reusable@example.test", "U406 Identity Reusable", "U406IdentityReusableEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-identity-reusable-incident", "IR-U406-IR", "Record relationships entity-storage identity reusable rows")
		survivorID := uuid.New()
		loserID := uuid.New()

		entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, survivorID, "Alex Survivor", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, loserID, "Alex Analyst Legacy", "alex.legacy@example.test", "alex.legacy@example.test", "ALEXLEGACY")

		if _, err := newEntityTestTimelineBundle(t, harness.DB).EntityMergeStore.MergeEntity(context.Background(), actor, survivorID, merge.MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-identity-reusable-merge",
		}, []byte("txn-entity_linking-u-4-06-identity-reusable-merge"), "req-identity-reusable-merge", entitytest.BaseTime); err != nil {
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
		}, []byte("txn-entity_linking-u-4-06-identity-reusable-create"), "req-identity-reusable-create", entitytest.BaseTime.Add(time.Minute))
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
