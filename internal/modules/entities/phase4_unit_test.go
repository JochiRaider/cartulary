package entities_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	. "github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/assertx"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

// U-4-03 / REQ-02-034, REQ-02-038, REQ-02-054..REQ-02-055 / AC-020, AC-021, AC-186.
func TestPhase4_CreateFromMention_U_4_03(t *testing.T) {
	t.Run("unique host exact match reuses the canonical record and resolves only the selected mention", func(t *testing.T) {
		harness := phase4test.StartStore(t, "phase4-u-4-03-host-reuse")
		store := NewStore(harness.Pool)
		actor := phase4test.SeedLocalUserFlags(t, harness.DB, "u403-host@example.test", "U403 Host", "U403HostPhase4Pass1!", false, false, true)
		incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-u-4-03-host-incident", "IR-U403-H", "Phase 4 U-4-03 host")

		phase4test.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		phase4test.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineRecordID)
		phase4test.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineSiblingRecordID)
		phase4test.SeedMention(t, harness.DB, actor.ID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)
		phase4test.SeedMention(t, harness.DB, actor.ID, golden.Phase4ResolvedHostMentionID, golden.Phase4TimelineSiblingRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)

		tx, err := harness.Pool.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		result, err := store.ResolveOrCreateFromMentionTx(context.Background(), tx, actor, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, golden.Phase4HostMentionID, nil, golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("resolve or create from mention: %v", err)
		}
		if err := tx.Commit(context.Background()); err != nil {
			t.Fatalf("commit tx: %v", err)
		}

		if result.RecordID != golden.Phase4CanonicalHostRecordID || result.EntityType != "host" {
			t.Fatalf("expected selected mention to reuse canonical host, got %#v", result)
		}
		selected := phase4test.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, selected, golden.Phase4MentionStatusResolved)
		if selected.ResolvedRecordID == nil || *selected.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected selected mention to resolve to canonical host, got %#v", selected)
		}
		if selected.RawText != "WS-023" {
			t.Fatalf("expected selected raw text preservation, got %#v", selected)
		}

		sibling := phase4test.LookupMention(t, harness.DB, golden.Phase4ResolvedHostMentionID)
		assertx.RequireMentionStatus(t, sibling, golden.Phase4MentionStatusUnresolved)
		if sibling.ResolvedRecordID != nil {
			t.Fatalf("expected sibling mention to remain unresolved, got %#v", sibling)
		}
		if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 1 {
			t.Fatalf("expected create-from-mention reuse to avoid creating a second host, got %d rows", got)
		}
	})

	t.Run("identity create-from-mention creates a stub with seed provenance when no exact match exists", func(t *testing.T) {
		harness := phase4test.StartStore(t, "phase4-u-4-03-identity-create")
		store := NewStore(harness.Pool)
		actor := phase4test.SeedLocalUserFlags(t, harness.DB, "u403-identity@example.test", "U403 Identity", "U403IdentityPhase4Pass1!", false, false, true)
		incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-u-4-03-identity-incident", "IR-U403-I", "Phase 4 U-4-03 identity")

		phase4test.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineMixedRecordID)
		phase4test.SeedMention(t, harness.DB, actor.ID, golden.Phase4IdentityMentionID, golden.Phase4TimelineMixedRecordID, golden.Phase4FieldTimelineIdentityRefs, "identity", "alex.analyst@example.test", "unresolved", nil, nil)

		tx, err := harness.Pool.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		result, err := store.ResolveOrCreateFromMentionTx(context.Background(), tx, actor, golden.Phase4TimelineMixedRecordID, golden.Phase4FieldTimelineIdentityRefs, golden.Phase4IdentityMentionID, nil, golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("resolve or create from mention: %v", err)
		}
		if err := tx.Commit(context.Background()); err != nil {
			t.Fatalf("commit tx: %v", err)
		}

		if result.EntityType != "identity" {
			t.Fatalf("expected identity create-from-mention, got %#v", result)
		}
		mention := phase4test.LookupMention(t, harness.DB, golden.Phase4IdentityMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != result.RecordID {
			t.Fatalf("expected mention to resolve to created identity, got %#v", mention)
		}

		var (
			state          string
			entityOrigin   string
			seedMentionID  sql.NullString
			displayName    string
			email          sql.NullString
			samAccountName sql.NullString
		)
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT identity_state, entity_origin, seed_entity_mention_id::text, display_name, email::text, sam_account_name
  FROM identities
 WHERE record_id = $1
`, result.RecordID).Scan(&state, &entityOrigin, &seedMentionID, &displayName, &email, &samAccountName); err != nil {
			t.Fatalf("lookup created identity: %v", err)
		}
		if state != "stub" || entityOrigin != "created_from_mention" || !seedMentionID.Valid || seedMentionID.String != golden.Phase4IdentityMentionID.String() {
			t.Fatalf("expected created-from-mention stub provenance, got state=%q origin=%q seed=%v", state, entityOrigin, seedMentionID)
		}
		if displayName != "alex.analyst@example.test" || !email.Valid || email.String != "alex.analyst@example.test" || samAccountName.Valid {
			t.Fatalf("unexpected identity seed values: display=%q email=%v sam=%v", displayName, email, samAccountName)
		}
	})
}

// U-4-04 / REQ-02-039..REQ-02-041 / AC-188..AC-190, AC-224, AC-225.
func TestPhase4_DismissRestoreMentionLifecycle_U_4_04(t *testing.T) {
	harness := phase4test.StartStore(t, "phase4-u-4-04")
	store := NewStore(harness.Pool)
	actor := phase4test.SeedLocalUserFlags(t, harness.DB, "u404@example.test", "U404", "U404Phase4Pass1!", false, false, true)
	incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-u-4-04-incident", "IR-U404", "Phase 4 U-4-04")

	phase4test.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineRecordID)
	phase4test.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
	phase4test.SeedResolvedMention(t, harness.DB, actor.ID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023")
	phase4test.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, golden.Phase4ManualLinkID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host", "manual", nil)

	tx, err := harness.Pool.BeginTx(context.Background(), pgxTxOptions())
	if err != nil {
		t.Fatalf("begin dismiss tx: %v", err)
	}
	if err := store.ApplyMentionLifecycleTx(context.Background(), tx, actor, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, golden.Phase4HostMentionID, "dismiss_item", nil, golden.Phase4BaseTime); err != nil {
		t.Fatalf("dismiss mention: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit dismiss tx: %v", err)
	}

	dismissed := phase4test.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
	assertx.RequireMentionStatus(t, dismissed, golden.Phase4MentionStatusDismissed)
	if dismissed.ResolvedRecordID != nil || dismissed.ResolvedAt != nil || dismissed.ResolutionMethod != nil {
		t.Fatalf("expected dismissed mention to clear resolution metadata, got %#v", dismissed)
	}
	if got := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incident.ID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID); got != 0 {
		t.Fatalf("expected dismiss to remove active derived link, got %d rows", got)
	}

	tx, err = harness.Pool.BeginTx(context.Background(), pgxTxOptions())
	if err != nil {
		t.Fatalf("begin restore tx: %v", err)
	}
	if err := store.ApplyMentionLifecycleTx(context.Background(), tx, actor, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, golden.Phase4HostMentionID, "revert_to_unresolved", nil, golden.Phase4BaseTime.Add(time.Minute)); err != nil {
		t.Fatalf("restore mention: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit restore tx: %v", err)
	}

	restored := phase4test.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
	assertx.RequireMentionStatus(t, restored, golden.Phase4MentionStatusUnresolved)
	if restored.RawText != "WS-023" || restored.ResolvedRecordID != nil || restored.ResolutionMethod != nil {
		t.Fatalf("expected durable restore-to-unresolved semantics, got %#v", restored)
	}
}

// U-4-05 / REQ-02-059..REQ-02-063 / AC-021, AC-022.
func TestPhase4_ExactMatchPrecedence_U_4_05(t *testing.T) {
	harness := phase4test.StartStore(t, "phase4-u-4-05")
	store := NewStore(harness.Pool)
	actor := phase4test.SeedLocalUserFlags(t, harness.DB, "u405@example.test", "U405", "U405Phase4Pass1!", false, false, true)
	incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-u-4-05-incident", "IR-U405", "Phase 4 U-4-05")

	phase4test.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalHostRecordID, "Canonical Host", "WS-023", "ws-023.corp.example.test", "")
	phase4test.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalHostRecordID, "host", "Workstation 23")
	phase4test.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalIdentityID, "Alex Analyst", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
	phase4test.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalIdentityID, "identity", "Case Owner")

	hostReuse, err := store.CreateHostRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID: "txn-phase4-u-4-05-host-reuse",
		Values: map[string]string{
			"host.display_name": "Host reuse",
			"host.fqdn":         "ws-023.corp.example.test",
		},
	}, []byte("txn-phase4-u-4-05-host-reuse"), "req-host-reuse", golden.Phase4BaseTime)
	if err != nil {
		t.Fatalf("host exact-match reuse: %v", err)
	}
	if hostReuse.RecordID != golden.Phase4CanonicalHostRecordID || hostReuse.StatusCode != 200 {
		t.Fatalf("expected fqdn exact match to reuse canonical host, got %#v", hostReuse)
	}

	hostAliasOnly, err := store.CreateHostRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID: "txn-phase4-u-4-05-host-alias",
		Values: map[string]string{
			"host.display_name": "Workstation 23",
		},
	}, []byte("txn-phase4-u-4-05-host-alias"), "req-host-alias", golden.Phase4BaseTime.Add(time.Minute))
	if err != nil {
		t.Fatalf("host alias-only create: %v", err)
	}
	if hostAliasOnly.RecordID == golden.Phase4CanonicalHostRecordID || hostAliasOnly.StatusCode != 201 {
		t.Fatalf("expected suggestion-only alias to avoid implicit reuse, got %#v", hostAliasOnly)
	}

	identityReuse, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID: "txn-phase4-u-4-05-identity-reuse",
		Values: map[string]string{
			"identity.display_name": "Identity reuse",
			"identity.email":        "alex.analyst@example.test",
		},
	}, []byte("txn-phase4-u-4-05-identity-reuse"), "req-identity-reuse", golden.Phase4BaseTime.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("identity exact-match reuse: %v", err)
	}
	if identityReuse.RecordID != golden.Phase4CanonicalIdentityID || identityReuse.StatusCode != 200 {
		t.Fatalf("expected email exact match to reuse canonical identity, got %#v", identityReuse)
	}

	if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1 AND host_state IN ('stub', 'canonical')`, incident.ID); got != 2 {
		t.Fatalf("expected one reused host and one new stub host, got %d active rows", got)
	}
	if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id = $1 AND identity_state IN ('stub', 'canonical')`, incident.ID); got != 1 {
		t.Fatalf("expected identity exact-match reuse without creating a duplicate, got %d active rows", got)
	}
}

// U-4-06 / REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
func TestPhase4_ExplicitEntityMerge_U_4_06(t *testing.T) {
	harness := phase4test.StartStore(t, "phase4-u-4-06")
	store := NewStore(harness.Pool)
	actor := phase4test.SeedLocalUserFlags(t, harness.DB, "u406@example.test", "U406", "U406Phase4Pass1!", false, false, true)
	incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-u-4-06-incident", "IR-U406", "Phase 4 U-4-06")

	phase4test.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
	phase4test.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
	phase4test.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateHostRecordID, "host", "Workstation 23")
	phase4test.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineRecordID)
	phase4test.SeedResolvedMention(t, harness.DB, actor.ID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateHostRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023")
	phase4test.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, golden.Phase4DuplicateLinkID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateHostRecordID, "observed_on_host", "manual", nil)
	phase4test.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, golden.Phase4TagIDSurvivor, golden.Phase4CanonicalHostRecordID, "critical-host")
	phase4test.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, golden.Phase4TagIDLoser, golden.Phase4DuplicateHostRecordID, "critical-host")
	phase4test.SeedAssessment(t, harness.DB, incident.ID, actor.ID, golden.Phase4AssessmentHostID, golden.Phase4DuplicateHostRecordID, "host", "confirmed")

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

	mention := phase4test.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
	assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusResolved)
	if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
		t.Fatalf("expected merge to repoint mention resolution to survivor, got %#v", mention)
	}
	link := phase4test.LookupActiveLink(t, harness.DB, incident.ID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host")
	assertx.RequireActiveLink(t, link, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host", "manual", nil)
	if got := phase4test.LookupAssessmentSubject(t, harness.DB, golden.Phase4AssessmentHostID); got != golden.Phase4CanonicalHostRecordID {
		t.Fatalf("expected assessment repoint to survivor, got %s", got)
	}
	if got := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incident.ID, golden.Phase4CanonicalHostRecordID); got != 1 {
		t.Fatalf("expected one active deduped survivor tag, got %d", got)
	}
	if got := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, golden.Phase4DuplicateLinkID); got != 0 {
		t.Fatalf("expected loser-targeted active link to disappear, got %d rows", got)
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
}

// U-4-07 / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestPhase4_IndicatorObservationSeparation_U_4_07(t *testing.T) {
	startFixture := func(t *testing.T, suffix string) (*phase4test.StoreHarness, *Store, authn.UserRecord, uuid.UUID) {
		t.Helper()

		harness := phase4test.StartStore(t, "phase4-u-4-07-"+suffix)
		store := NewStore(harness.Pool)
		actor := phase4test.SeedLocalUserFlags(t, harness.DB, "u407-"+suffix+"@example.test", "U407 "+suffix, "U407Phase4Pass1!", false, false, true)
		incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-u-4-07-"+suffix, "IR-U407-"+suffix, "Phase 4 U-4-07 "+suffix)
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

		otherActor := phase4test.SeedLocalUserFlags(t, harness.DB, "u407-other@example.test", "U407 Other", "U407OtherPhase4Pass1!", false, false, true)
		otherIncident := phase4test.CreateIncidentInStore(t, harness.Pool, otherActor, "txn-phase4-u-4-07-other-incident", "IR-U407-B", "Phase 4 U-4-07 other")
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
		phase4test.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, golden.Phase4TimelineRecordID)
		phase4test.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, golden.Phase4TimelineSiblingRecordID)

		observationOne, _, err := store.CreateIndicatorObservation(context.Background(), actor, IndicatorObservationCreateParams{
			IncidentID:     incidentID,
			SourceRecordID: golden.Phase4TimelineRecordID,
			SourceFieldKey: golden.Phase4FieldTimelineSourceText,
			OriginKind:     "interactive_cell",
			OriginLocator:  "view:timeline/record:1/cell:timeline.source_text/span:12-24",
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
			OriginLocator:  "view:timeline/record:2/cell:timeline.summary/span:5-17",
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

		projected := lookupIndicatorProjection(t, harness.DB, created.RecordID)
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
		phase4test.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, golden.Phase4TimelineMixedRecordID)
		if _, _, err := store.CreateIndicatorObservation(context.Background(), actor, IndicatorObservationCreateParams{
			IncidentID:                incidentID,
			SourceRecordID:            golden.Phase4TimelineMixedRecordID,
			SourceFieldKey:            golden.Phase4FieldTimelineSourceText,
			OriginKind:                "interactive_cell",
			OriginLocator:             "view:timeline/record:3/cell:timeline.source_text/span:1-9",
			ObservedText:              "203[.]0[.]113[.]24",
			ResolvedIndicatorRecordID: &created.RecordID,
			CreatedAt:                 golden.Phase4BaseTime,
		}); err != nil {
			t.Fatalf("create resolved observation: %v", err)
		}

		projected := lookupIndicatorProjection(t, harness.DB, created.RecordID)
		if projected.LifecycleSummary == nil || *projected.LifecycleSummary != "active" {
			t.Fatalf("expected lifecycle_summary=active, got %#v", projected)
		}
		if projected.FirstObservedAt == nil || !projected.FirstObservedAt.UTC().Equal(golden.Phase4BaseTime) {
			t.Fatalf("expected observation timestamps to derive from observations, got %#v", projected)
		}

		intervalRow := lookupIndicatorLifecycleInterval(t, harness.DB, interval.IntervalID)
		if !intervalRow.ValidFrom.UTC().Equal(golden.Phase4PastTime) {
			t.Fatalf("expected lifecycle valid_from to remain distinct from observation timestamps, got %#v", intervalRow)
		}
	})
}

func pgxTxOptions() pgx.TxOptions {
	return pgx.TxOptions{}
}
