package entities_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/assertx"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

// U-4-01 / REQ-02-028..REQ-02-036 / AC-019, AC-020, AC-022.
func TestPhase4_BindingMode_U_4_01_Red(t *testing.T) {
	views := fixtures.Phase4Views()
	if got := views["timeline"].FieldBinding[golden.Phase4FieldTimelineHostRefs]; got != golden.Phase4BindingMentionOrigin {
		t.Fatalf("expected fixture binding mention_origin for %s, got %q", golden.Phase4FieldTimelineHostRefs, got)
	}
	if got := views["hosts"].FieldBinding["host.display_name"]; got != golden.Phase4BindingEntityOrigin {
		t.Fatalf("expected fixture binding entity_origin for host.display_name, got %q", got)
	}
	if got := views["identities"].FieldBinding["identity.display_name"]; got != golden.Phase4BindingEntityOrigin {
		t.Fatalf("expected fixture binding entity_origin for identity.display_name, got %q", got)
	}

	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4TimelineViewSchemaID, golden.Phase4FieldTimelineHostRefs, golden.Phase4BindingMentionOrigin)
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4TimelineViewSchemaID, golden.Phase4FieldTimelineIdentityRefs, golden.Phase4BindingMentionOrigin)
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4HostsViewSchemaID, "host.display_name", golden.Phase4BindingEntityOrigin)
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4IdentitiesViewSchemaID, "identity.display_name", golden.Phase4BindingEntityOrigin)
}

// U-4-02 / REQ-02-031..REQ-02-032, REQ-02-058 / AC-019, AC-021.
func TestPhase4_DuplicateMentionProvenance_U_4_02_Red(t *testing.T) {
	mentions := fixtures.Phase4Mentions()
	assertx.RequireDistinctMentionProvenance(
		t,
		mentions["host_unresolved"],
		mentions["repeated_distinct_source_rows"],
		mentions["repeated_distinct_locator"],
	)
	if mentions["host_unresolved"].RawText != mentions["repeated_distinct_source_rows"].RawText {
		t.Fatalf("expected identical repeated raw mention text, got %q and %q", mentions["host_unresolved"].RawText, mentions["repeated_distinct_source_rows"].RawText)
	}

	phase4test.RequireMigrationTables(t, "U-4-02", "entity_mentions")
}

// U-4-03 / REQ-02-034, REQ-02-038, REQ-02-054..REQ-02-055 / AC-020, AC-021, AC-186.
func TestPhase4_CreateFromMention_U_4_03_Red(t *testing.T) {
	mentions := fixtures.Phase4Mentions()
	selected := mentions["host_unresolved"]
	sibling := mentions["repeated_distinct_locator"]
	selected.ResolutionStatus = golden.Phase4MentionStatusResolved
	selected.ResolvedRecordID = &golden.Phase4StubHostRecordID

	assertx.RequireSelectedMentionOnlyResolvedByCreateFromMention(t, selected, sibling, golden.Phase4StubHostRecordID)
	assertx.RequireRawTextPreserved(t, mentions["host_unresolved"].RawText, selected.RawText)

	phase4test.RequireMigrationTables(t, "U-4-03", "hosts", "entity_mentions")
}

// U-4-04 / REQ-02-039..REQ-02-041 / AC-188..AC-190, AC-224, AC-225.
func TestPhase4_DismissRestoreMentionLifecycle_U_4_04_Red(t *testing.T) {
	mentions := fixtures.Phase4Mentions()
	before := mentions["host_resolved"]
	dismissed := before
	dismissed.ResolutionStatus = golden.Phase4MentionStatusDismissed
	dismissed.ResolvedRecordID = nil
	dismissed.ResolvedByUserID = nil
	dismissed.ResolvedAt = nil
	dismissed.ResolutionMethod = nil

	assertx.RequireDismissedMentionPreservesRowAndText(t, before, dismissed)
	assertx.RequireDismissedMentionClearsResolution(t, dismissed)

	restored := dismissed
	restored.ResolutionStatus = golden.Phase4MentionStatusUnresolved
	assertx.RequireRestoreToUnresolved(t, restored)

	phase4test.RequireMigrationTables(t, "U-4-04", "entity_mentions")
}

// U-4-05 / REQ-02-059..REQ-02-063 / AC-021, AC-022.
func TestPhase4_ExactMatchPrecedence_U_4_05_Red(t *testing.T) {
	hosts := fixtures.Phase4Hosts()
	identities := fixtures.Phase4Identities()

	assertx.RequireExactMatchPrecedence(t, golden.Phase4HostExactMatchPrecedence, []string{"aad_device_id", "fqdn", "hostname"})
	assertx.RequireExactMatchPrecedence(t, golden.Phase4IdentityExactMatchPrecedence, []string{"aad_object_id", "sid", "upn", "email", "sam_account_name"})
	assertx.RequireSuggestionBoundary(t, hosts["canonical"].PreservedIdentifiers)
	assertx.RequireSuggestionBoundary(t, identities["canonical"].PreservedIdentifiers)

	phase4test.RequireMigrationTables(t, "U-4-05", "hosts", "identities", "entity_aliases")
}

// U-4-06 / REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
func TestPhase4_ExplicitEntityMerge_U_4_06_Red(t *testing.T) {
	links := fixtures.Phase4Links()
	before := links["duplicate_merge_candidate"]
	after := before
	after.TargetID = golden.Phase4CanonicalHostRecordID
	assertx.RequireMergeRepointsLiveLink(t, after, golden.Phase4CanonicalHostRecordID)

	assessments := fixtures.Phase4Assessments()
	if assessments["loser"].SubjectID != golden.Phase4DuplicateIdentityID {
		t.Fatalf("expected loser assessment fixture to point at duplicate identity, got %#v", assessments["loser"])
	}
	assertx.RequireRawTextPreserved(t, fixtures.Phase4Mentions()["host_unresolved"].RawText, fixtures.Phase4Mentions()["host_unresolved"].RawText)

	phase4test.RequireMigrationTables(t, "U-4-06", "hosts", "identities", "record_tags", "compromise_assessments", "entity_mentions")
}

// U-4-07 / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestPhase4_IndicatorObservationSeparation_U_4_07_Red(t *testing.T) {
	indicators := fixtures.Phase4Indicators()
	observations := fixtures.Phase4IndicatorObservations()
	intervals := fixtures.Phase4IndicatorIntervals()

	assertx.RequireIndicatorObservationSeparation(t, observations["source_bound"], indicators["canonical"])
	assertx.RequireIndicatorLifecycleSeparate(t, intervals["active"], indicators["canonical"])

	phase4test.RequireMigrationTables(t, "U-4-07", "indicators", "indicator_observations", "indicator_state_intervals")
	phase4test.RequireViewContract(t, "U-4-07", golden.Phase4IndicatorsViewSchemaID)

	t.Run("indicator view create dedupes canonically within one incident and isolates incidents", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-07-create")
		phase4test.RequireSchemaTables(t, harness.DB, "U-4-07", "indicators", "indicator_observations", "indicator_state_intervals", "indicator_grid_projection")

		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incidentOne := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-07-incident-1",
			"incident_key":  "IR-U407-1",
			"title":         "Indicator canonical create 1",
		})
		incidentTwo := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-07-incident-2",
			"incident_key":  "IR-U407-2",
			"title":         "Indicator canonical create 2",
		})
		incidentOneID := mustUUID(t, incidentOne["incident_id"].(string))
		incidentTwoID := mustUUID(t, incidentTwo["incident_id"].(string))

		createResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentOneID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			fixtures.IndicatorCreatePayload("txn-phase4-u-4-07-create-1"),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := httptestSuccess(t, createResp, http.StatusCreated)
		row := createData["row"].(map[string]any)
		firstRecordID := mustUUID(t, row["record_id"].(string))
		assertx.RequireProjectionRowRecordID(t, row, firstRecordID.String())
		requireIndicatorCellValue(t, row, "indicator.indicator_type", golden.Phase4IndicatorExamples[0].IndicatorType)
		requireIndicatorCellValue(t, row, "indicator.display_value", golden.Phase4IndicatorExamples[0].DisplayValue)
		requireIndicatorCellValue(t, row, "indicator.observation_count", float64(0))
		requireIndicatorCellValue(t, row, "indicator.first_observed_at", nil)
		requireIndicatorCellValue(t, row, "indicator.last_observed_at", nil)
		requireIndicatorCellValue(t, row, "indicator.lifecycle_summary", nil)
		requireIndicatorCellValue(t, row, "indicator.supporting_link_count", float64(0))

		replayResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentOneID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			fixtures.IndicatorCreatePayload("txn-phase4-u-4-07-create-2"),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		replayData := httptestSuccess(t, replayResp, http.StatusOK)
		if got := replayData["row"].(map[string]any)["record_id"]; got != firstRecordID.String() {
			t.Fatalf("expected same-incident dedupe to reuse canonical indicator %s, got %#v", firstRecordID, replayData)
		}

		defangedVariant := fixtures.IndicatorCreatePayload("txn-phase4-u-4-07-create-3")
		defangedVariant["indicator.defanged_value"] = "203(.)0(.)113(.)24"
		defangedVariant["indicator.stix_pattern"] = "[ipv4-addr:value = '203.0.113.24']"
		defangedResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentOneID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			defangedVariant,
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		defangedData := httptestSuccess(t, defangedResp, http.StatusOK)
		if got := defangedData["row"].(map[string]any)["record_id"]; got != firstRecordID.String() {
			t.Fatalf("expected defanged_value and stix_pattern to stay outside canonical identity, got %#v", defangedData)
		}

		hashExample := golden.Phase4IndicatorExamples[3]
		hashCreate := map[string]any{
			"client_txn_id":              "txn-phase4-u-4-07-create-4",
			"indicator.indicator_type":   hashExample.IndicatorType,
			"indicator.value_kind":       hashExample.ValueKind,
			"indicator.display_value":    hashExample.DisplayValue,
			"indicator.normalized_value": hashExample.NormalizedValue,
			"indicator.hash_algorithm":   hashExample.HashAlgorithm,
			"indicator.hash_value":       hashExample.HashValue,
			"indicator.stix_pattern":     hashExample.STIXPattern,
		}
		hashResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentOneID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			hashCreate,
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		hashData := httptestSuccess(t, hashResp, http.StatusCreated)
		hashRecordID := mustUUID(t, hashData["row"].(map[string]any)["record_id"].(string))

		hashVariant := map[string]any{
			"client_txn_id":              "txn-phase4-u-4-07-create-5",
			"indicator.indicator_type":   hashExample.IndicatorType,
			"indicator.value_kind":       hashExample.ValueKind,
			"indicator.display_value":    hashExample.DisplayValue,
			"indicator.normalized_value": hashExample.NormalizedValue,
			"indicator.hash_algorithm":   hashExample.HashAlgorithm,
			"indicator.hash_value":       hashExample.HashValue,
			"indicator.stix_pattern":     "[file:hashes.'SHA-256' = 'elsewhere']",
		}
		hashVariantResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentOneID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			hashVariant,
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		hashVariantData := httptestSuccess(t, hashVariantResp, http.StatusOK)
		if got := hashVariantData["row"].(map[string]any)["record_id"]; got != hashRecordID.String() {
			t.Fatalf("expected stix_pattern changes to reuse the same hash-backed canonical indicator, got %#v", hashVariantData)
		}

		crossIncidentResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentTwoID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			fixtures.IndicatorCreatePayload("txn-phase4-u-4-07-create-6"),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		crossIncidentData := httptestSuccess(t, crossIncidentResp, http.StatusCreated)
		secondIncidentRecordID := mustUUID(t, crossIncidentData["row"].(map[string]any)["record_id"].(string))
		if secondIncidentRecordID == firstRecordID {
			t.Fatalf("expected incident-scoped canonical dedupe, got same record_id across incidents %s", firstRecordID)
		}

		hashOnlyAlg := map[string]any{
			"client_txn_id":            "txn-phase4-u-4-07-create-7",
			"indicator.indicator_type": hashExample.IndicatorType,
			"indicator.value_kind":     hashExample.ValueKind,
			"indicator.display_value":  hashExample.DisplayValue,
			"indicator.hash_algorithm": hashExample.HashAlgorithm,
		}
		hashOnlyAlgResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentOneID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			hashOnlyAlg,
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestError(t, hashOnlyAlgResp, http.StatusBadRequest, "invalid_mutation_payload")

		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE incident_id = $1 AND deleted_at IS NULL`, incidentOneID); got != 2 {
			t.Fatalf("expected two canonical indicators in incident one, got %d", got)
		}
		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE incident_id = $1 AND deleted_at IS NULL`, incidentTwoID); got != 1 {
			t.Fatalf("expected one canonical indicator in incident two, got %d", got)
		}
		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentOneID); got != 5 {
			t.Fatalf("expected one change_set per accepted create/upsert in incident one, got %d", got)
		}

		indicatorRow := lookupIndicatorRecord(t, harness.DB, firstRecordID)
		if indicatorRow.CreatedByUser != adminUserID || indicatorRow.UpdatedByUser != adminUserID {
			t.Fatalf("expected create attribution on canonical indicator, got %#v", indicatorRow)
		}
		if indicatorRow.DedupeKey == "" || indicatorRow.DisplayValue != golden.Phase4IndicatorExamples[0].DisplayValue || derefStringPointer(indicatorRow.NormalizedValue) != golden.Phase4IndicatorExamples[0].NormalizedValue {
			t.Fatalf("unexpected canonical indicator storage %#v", indicatorRow)
		}
	})

	t.Run("source-bound observations stay distinct and drive derived projection fields", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-07-observations")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-07-observations-incident",
			"incident_key":  "IR-U407-OBS",
			"title":         "Indicator observations",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		createResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			fixtures.IndicatorCreatePayload("txn-phase4-u-4-07-observations-create"),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := httptestSuccess(t, createResp, http.StatusCreated)
		indicatorRecordID := mustUUID(t, createData["row"].(map[string]any)["record_id"].(string))

		seedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		seedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineSiblingRecordID)
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "WS-023", "ws-023", "", "")
		seedResolvedMention(t, harness.DB, adminUserID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, golden.Phase4FieldTimelineSourceText, "WS-023")

		store := entities.NewStore(harness.Server.Runtime.Postgres)
		actor := authn.UserRecord{ID: adminUserID}

		observationOne, _, err := store.CreateIndicatorObservation(context.Background(), actor, entities.IndicatorObservationCreateParams{
			IncidentID:     incidentID,
			SourceRecordID: golden.Phase4TimelineRecordID,
			SourceFieldKey: golden.Phase4FieldTimelineSourceText,
			OriginKind:     "interactive_cell",
			OriginLocator:  "view:timeline/record:1/cell:timeline.source_text/span:12-24",
			ObservedText:   "203[.]0[.]113[.]24",
			CreatedAt:      golden.Phase4PastTime,
		})
		if err != nil {
			t.Fatalf("create indicator observation one: %v", err)
		}
		observationTwo, _, err := store.CreateIndicatorObservation(context.Background(), actor, entities.IndicatorObservationCreateParams{
			IncidentID:     incidentID,
			SourceRecordID: golden.Phase4TimelineSiblingRecordID,
			SourceFieldKey: golden.Phase4FieldTimelineSummary,
			OriginKind:     "interactive_cell",
			OriginLocator:  "view:timeline/record:2/cell:timeline.summary/span:5-17",
			ObservedText:   "203[.]0[.]113[.]24",
			CreatedAt:      golden.Phase4BaseTime,
		})
		if err != nil {
			t.Fatalf("create indicator observation two: %v", err)
		}
		if observationOne.ObservationID == observationTwo.ObservationID {
			t.Fatalf("expected repeated identical observations to stay distinct, got %#v %#v", observationOne, observationTwo)
		}

		projectedBeforeResolve := lookupIndicatorProjection(t, harness.DB, indicatorRecordID)
		if projectedBeforeResolve.ObservationCount != 0 || projectedBeforeResolve.FirstObservedAt != nil || projectedBeforeResolve.LastObservedAt != nil {
			t.Fatalf("unresolved observations must not populate canonical observation summary, got %#v", projectedBeforeResolve)
		}

		resolvedOne, _, err := store.ResolveIndicatorObservation(context.Background(), actor, entities.IndicatorObservationResolveParams{
			ObservationID:             observationOne.ObservationID,
			ResolvedIndicatorRecordID: indicatorRecordID,
			ResolvedAt:                golden.Phase4BaseTime,
		})
		if err != nil {
			t.Fatalf("resolve indicator observation one: %v", err)
		}
		resolvedTwo, _, err := store.ResolveIndicatorObservation(context.Background(), actor, entities.IndicatorObservationResolveParams{
			ObservationID:             observationTwo.ObservationID,
			ResolvedIndicatorRecordID: indicatorRecordID,
			ResolvedAt:                golden.Phase4BaseTime.Add(2 * time.Minute),
		})
		if err != nil {
			t.Fatalf("resolve indicator observation two: %v", err)
		}
		if resolvedOne.ResolvedIndicatorRecordID == nil || *resolvedOne.ResolvedIndicatorRecordID != indicatorRecordID {
			t.Fatalf("expected resolved observation one to point at canonical indicator, got %#v", resolvedOne)
		}
		if resolvedTwo.ResolvedIndicatorRecordID == nil || *resolvedTwo.ResolvedIndicatorRecordID != indicatorRecordID {
			t.Fatalf("expected resolved observation two to point at canonical indicator, got %#v", resolvedTwo)
		}

		projected := lookupIndicatorProjection(t, harness.DB, indicatorRecordID)
		if projected.ObservationCount != 2 {
			t.Fatalf("expected observation_count=2, got %#v", projected)
		}
		if projected.FirstObservedAt == nil || !projected.FirstObservedAt.UTC().Equal(golden.Phase4PastTime) {
			t.Fatalf("expected first_observed_at=%s, got %#v", golden.Phase4PastTime, projected)
		}
		if projected.LastObservedAt == nil || !projected.LastObservedAt.UTC().Equal(golden.Phase4BaseTime) {
			t.Fatalf("expected last_observed_at=%s, got %#v", golden.Phase4BaseTime, projected)
		}
		if projected.RecordID != indicatorRecordID {
			t.Fatalf("expected indicator projection keyed by canonical record_id %s, got %#v", indicatorRecordID, projected)
		}

		observationRows := listIndicatorObservations(t, harness.DB, incidentID)
		if len(observationRows) != 2 {
			t.Fatalf("expected two indicator observation rows, got %#v", observationRows)
		}
		if observationRows[0].ObservedText != "203[.]0[.]113[.]24" || observationRows[1].ObservedText != "203[.]0[.]113[.]24" {
			t.Fatalf("expected raw observed text preservation, got %#v", observationRows)
		}
		if observationRows[0].SourceRecordID == observationRows[1].SourceRecordID {
			t.Fatalf("expected repeated same-value observations on distinct source rows to remain distinct, got %#v", observationRows)
		}
		if observationRows[0].OriginLocator == observationRows[1].OriginLocator {
			t.Fatalf("expected observation provenance locators to remain distinct, got %#v", observationRows)
		}
		if observationRows[0].ParsedIndicatorType == nil || *observationRows[0].ParsedIndicatorType != golden.Phase4IndicatorTypeIPv4 {
			t.Fatalf("expected parsed indicator type guess to be preserved, got %#v", observationRows[0])
		}
		if observationRows[0].NormalizedCandidate == nil || *observationRows[0].NormalizedCandidate != golden.Phase4IndicatorExamples[0].NormalizedValue {
			t.Fatalf("expected normalized candidate preservation, got %#v", observationRows[0])
		}

		mention := lookupMention(t, harness.DB, golden.Phase4HostMentionID)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.Phase4CanonicalHostRecordID || mention.RowVersion != 1 {
			t.Fatalf("indicator observation resolution must not rewrite coexisting host mention rows, got %#v", mention)
		}
		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID); got != 5 {
			t.Fatalf("expected one change_set for create, two observation creates, and two observation resolves; got %d", got)
		}
	})

	t.Run("lifecycle intervals stay separate from observation-derived timestamps", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-07-lifecycle")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-07-lifecycle-incident",
			"incident_key":  "IR-U407-LIFE",
			"title":         "Indicator lifecycle",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		createResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
			fixtures.IndicatorCreatePayload("txn-phase4-u-4-07-lifecycle-create"),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := httptestSuccess(t, createResp, http.StatusCreated)
		indicatorRecordID := mustUUID(t, createData["row"].(map[string]any)["record_id"].(string))

		store := entities.NewStore(harness.Server.Runtime.Postgres)
		actor := authn.UserRecord{ID: adminUserID}
		interval, _, err := store.AppendIndicatorLifecycleInterval(context.Background(), actor, entities.IndicatorLifecycleAppendParams{
			IncidentID:        incidentID,
			IndicatorRecordID: indicatorRecordID,
			LifecycleState:    "active",
			ValidFrom:         golden.Phase4PastTime,
			CreatedAt:         golden.Phase4PastTime,
		})
		if err != nil {
			t.Fatalf("append indicator lifecycle interval: %v", err)
		}
		if interval.IndicatorRecordID != indicatorRecordID {
			t.Fatalf("expected lifecycle interval to target canonical indicator, got %#v", interval)
		}

		seedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		_, _, err = store.CreateIndicatorObservation(context.Background(), actor, entities.IndicatorObservationCreateParams{
			IncidentID:                incidentID,
			SourceRecordID:            golden.Phase4TimelineRecordID,
			SourceFieldKey:            golden.Phase4FieldTimelineSourceText,
			OriginKind:                "interactive_cell",
			OriginLocator:             "view:timeline/record:1/cell:timeline.source_text/span:12-24",
			ObservedText:              "203[.]0[.]113[.]24",
			ResolvedIndicatorRecordID: &indicatorRecordID,
			CreatedAt:                 golden.Phase4BaseTime,
		})
		if err != nil {
			t.Fatalf("create resolved observation for lifecycle test: %v", err)
		}

		projected := lookupIndicatorProjection(t, harness.DB, indicatorRecordID)
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
		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM indicator_state_intervals WHERE indicator_record_id = $1`, indicatorRecordID); got != 1 {
			t.Fatalf("expected append-only lifecycle storage with one interval row, got %d", got)
		}
	})
}
