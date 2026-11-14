package artifacts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

const artifactParentPath = "data/artifacts.ndjson"
const artifactFindingPath = "data/artifact_findings.ndjson"
const artifactQueryPath = "data/artifact_investigative_queries.ndjson"
const artifactKeywordPath = "data/artifact_forensic_keywords.ndjson"
const artifactRiskPath = "data/handoff_risk_refs.ndjson"

var artifactPortPaths = []string{artifactParentPath, artifactFindingPath, artifactQueryPath, artifactKeywordPath, artifactRiskPath}
var artifactSourceActor = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
var artifactSourceOtherActor = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")

type artifactPortableRows map[string][]map[string]any

func (rows artifactPortableRows) clone() artifactPortableRows {
	copy := artifactPortableRows{}
	for path, values := range rows {
		for _, row := range values {
			copy[path] = append(copy[path], maps.Clone(row))
		}
	}
	return copy
}
func artifactPortRecordID(index int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("10000000-0000-4000-8000-%012d", index))
}
func artifactPortableParent(incident, id uuid.UUID, kind string) map[string]any {
	// Known inactive parent columns intentionally remain populated source state.
	return map[string]any{
		"record_id": id.String(), "incident_id": incident.String(), "artifact_type": kind,
		"title": "Parent title\u200d", "body": "Body\n\tSecond line", "timestamp_utc": "2000-01-01T00:00:00.000001+00:00",
		"created_at": "2026-01-01T00:00:00.123456+00:00", "updated_at": "2026-01-03T00:00:00+00:00",
		"comm_id": "comm-source-label", "comm_type": "briefing", "audience": "Audience", "channel_or_meeting": "Bridge",
		"summary": "Summary", "next_report_at": "2030-01-01T00:00:00+00:00", "privilege_tag": "Privilege",
		"handoff_id": "handoff-source-label", "outgoing_owner_user_id": artifactSourceActor.String(), "incoming_owner_user_id": artifactSourceOtherActor.String(),
		"current_state_summary": "Current\nstate", "next_checks": "Next\nchecks", "acknowledged_at": "2030-02-01T00:00:00+00:00",
		"status_review_id": "status-source-label", "review_owner_user_id": artifactSourceOtherActor.String(), "active_risks_summary": "Active risks",
		"lesson_id": "lesson-source-label", "owner_user_id": artifactSourceActor.String(), "closure_state": "open", "created_by_user_id": artifactSourceOtherActor.String(),
	}
}
func artifactPortableFixture(incident uuid.UUID) artifactPortableRows {
	rows := artifactPortableRows{}
	for i, kind := range []string{"note", "comm_log", "handoff", "status_review", "lesson", "finding", "investigative_query", "forensic_keyword"} {
		rows[artifactParentPath] = append(rows[artifactParentPath], artifactPortableParent(incident, artifactPortRecordID(i+1), kind))
	}
	// Independent child clocks deliberately differ from the parent clocks.
	rows[artifactFindingPath] = []map[string]any{{"record_id": artifactPortRecordID(6).String(), "incident_id": incident.String(), "kind": "hypothesis", "statement": "Finding\nstatement", "state": "closed", "confidence_score": json.Number("75"), "owner_user_id": artifactSourceActor.String(), "closed_at": "2026-01-03T12:00:00+00:00", "created_at": "2026-01-02T00:00:00+00:00", "updated_at": "2026-01-04T00:00:00+00:00"}}
	rows[artifactQueryPath] = []map[string]any{{"record_id": artifactPortRecordID(7).String(), "incident_id": incident.String(), "query_id": "query-source-label", "platform": "Platform", "purpose": "Purpose", "query_text": "Query\n\ttext", "created_by_user_id": artifactSourceActor.String(), "created_at": "2026-01-02T00:00:00+00:00", "updated_at": "2026-01-05T00:00:00+00:00"}}
	rows[artifactKeywordPath] = []map[string]any{{"record_id": artifactPortRecordID(8).String(), "incident_id": incident.String(), "keyword_id": "keyword-source-label", "pattern": "Pattern\u200d", "reason": "Reason\n\ttext", "match_mode": "regex", "case_sensitive": true, "created_at": "2026-01-02T00:00:00+00:00", "updated_at": "2026-01-06T00:00:00+00:00"}}
	for i, text := range []string{"Canonical Risk", "canonical risk", "Canonical Risk"} {
		row := map[string]any{"risk_ref_id": fmt.Sprintf("30000000-0000-4000-8000-%012d", i+1), "incident_id": incident.String(), "handoff_record_id": artifactPortRecordID(3).String(), "risk_ref_text": text, "created_by_user_id": artifactSourceActor.String(), "created_at": "2026-01-02T00:00:00+00:00", "deleted_at": nil, "deleted_by_user_id": nil}
		if i == 2 {
			row["deleted_at"], row["deleted_by_user_id"] = "2026-01-03T00:00:00+00:00", artifactSourceOtherActor.String()
		}
		rows[artifactRiskPath] = append(rows[artifactRiskPath], row)
	}
	return rows
}
func artifactPortBundle(t testing.TB, rows artifactPortableRows) portabilitysupport.MapBundle {
	t.Helper()
	bundle := portabilitysupport.MapBundle{}
	for _, path := range artifactPortPaths {
		var bytes bytes.Buffer
		encoder := json.NewEncoder(&bytes)
		for _, row := range rows[path] {
			if err := encoder.Encode(row); err != nil {
				t.Fatal(err)
			}
		}
		bundle[path] = bytes.Bytes()
	}
	return bundle
}
func artifactPortInput(t testing.TB, incident, local uuid.UUID) sourceport.ImportContext {
	t.Helper()
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{
		{SourceActorID: artifactSourceActor.String(), DisplayName: "Historical Artifact actor"},
		{SourceActorID: artifactSourceOtherActor.String(), DisplayName: "Other historical Artifact actor"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return sourceport.ImportContext{IncidentID: incident, ActorUserID: local, OperationID: "artifact-portable", BundleVersion: 8, Actors: actors}
}
func requireArtifactPortFailure(t testing.TB, err error, invariant string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID() != "artifacts" || failure.InvariantID() != "artifacts."+invariant {
		t.Fatalf("failure=%v, want artifacts.%s", err, invariant)
	}
}

func TestArtifactIncidentBundleAdmission_Unit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	incident, local := uuid.New(), uuid.New()
	input := artifactPortInput(t, incident, local)
	port, err := artifacts.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatal(err)
	}
	base := artifactPortableFixture(incident)
	prepare := func(t testing.TB, rows artifactPortableRows) error {
		t.Helper()
		bundle := artifactPortBundle(t, rows)
		before := map[string][]byte{}
		for path, payload := range bundle {
			before[path] = bytes.Clone(payload)
		}
		_, err := port.PrepareImport(ctx, bundle, input)
		for path, payload := range bundle {
			if !bytes.Equal(payload, before[path]) {
				t.Fatal("Prepare changed source bytes")
			}
		}
		return err
	}
	if err := prepare(t, base); err != nil {
		t.Fatalf("complete valid fixture: %v", err)
	}
	if err := prepare(t, artifactPortableRows{}); err != nil {
		t.Fatalf("empty owner: %v", err)
	}
	for _, path := range artifactPortPaths {
		t.Run(path+" exact shape", func(t *testing.T) {
			for field := range base[path][0] {
				for _, missing := range []bool{false, true} {
					rows := base.clone()
					if missing {
						delete(rows[path][0], field)
					} else {
						rows[path][0][field] = []any{"wrong scalar"}
					}
					invariant := "row_shape_exact"
					if field == "record_id" || field == "risk_ref_id" {
						invariant = "source_identity_admitted"
					}
					requireArtifactPortFailure(t, prepare(t, rows), invariant)
				}
			}
			for field := range base[path][0] {
				rows := base.clone()
				rows[path][0][field] = nil
				invariant := "row_shape_exact"
				if field == "record_id" || field == "risk_ref_id" {
					invariant = "source_identity_admitted"
				}
				if path == artifactParentPath && !strings.Contains(" record_id incident_id artifact_type created_at updated_at ", " "+field+" ") {
					invariant = ""
				}
				if path == artifactFindingPath && field == "confidence_score" {
					invariant = ""
				}
				if path == artifactFindingPath && field == "closed_at" {
					invariant = "lifecycle_fields_legal"
				}
				if path == artifactRiskPath && (field == "deleted_at" || field == "deleted_by_user_id") {
					invariant = ""
				}
				if invariant == "" {
					if err := prepare(t, rows); err != nil {
						t.Fatalf("nullable %s/%s: %v", path, field, err)
					}
				} else {
					requireArtifactPortFailure(t, prepare(t, rows), invariant)
				}
			}
			rows := base.clone()
			rows[path][0]["unknown_user_id"] = artifactSourceActor.String()
			requireArtifactPortFailure(t, prepare(t, rows), "row_shape_exact")
			rows = base.clone()
			rows[path] = append([]map[string]any{maps.Clone(rows[path][0])}, rows[path]...)
			requireArtifactPortFailure(t, prepare(t, rows), "source_identity_admitted")
		})
	}
	// These lists are authored test expectations, independent of the decoder's field groups.
	for _, binding := range []struct{ path, fields string }{
		{artifactParentPath, "record_id incident_id outgoing_owner_user_id incoming_owner_user_id review_owner_user_id owner_user_id created_by_user_id"},
		{artifactFindingPath, "record_id incident_id owner_user_id"},
		{artifactQueryPath, "record_id incident_id created_by_user_id"},
		{artifactKeywordPath, "record_id incident_id"},
		{artifactRiskPath, "risk_ref_id incident_id handoff_record_id created_by_user_id deleted_by_user_id"},
	} {
		for _, field := range strings.Fields(binding.fields) {
			for _, bad := range []string{"00000000-0000-0000-0000-000000000000", "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", " aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "urn:uuid:aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"} {
				rows := base.clone()
				rows[binding.path][0][field] = bad
				invariant := "row_shape_exact"
				if field == "record_id" || field == "risk_ref_id" {
					invariant = "source_identity_admitted"
				}
				requireArtifactPortFailure(t, prepare(t, rows), invariant)
			}
		}
	}
	for _, binding := range []struct{ path, fields string }{
		{artifactParentPath, "timestamp_utc created_at updated_at next_report_at acknowledged_at"},
		{artifactFindingPath, "created_at updated_at closed_at"}, {artifactQueryPath, "created_at updated_at"},
		{artifactKeywordPath, "created_at updated_at"}, {artifactRiskPath, "created_at deleted_at"},
	} {
		for _, field := range strings.Fields(binding.fields) {
			for _, bad := range []string{"0000-01-01T00:00:00Z", "2026-01-01T00:00:00.1234567Z", "2026-01-01T00:00:00.100Z", "2026-01-01T01:00:00+01:00", "not an instant"} {
				rows := base.clone()
				rows[binding.path][0][field] = bad
				requireArtifactPortFailure(t, prepare(t, rows), "row_shape_exact")
			}
		}
	}
	for _, path := range artifactPortPaths {
		rows := base.clone()
		rows[path][0]["incident_id"] = uuid.NewString()
		requireArtifactPortFailure(t, prepare(t, rows), "references_same_incident")
	}
	for _, score := range []any{nil, json.Number("0"), json.Number("100")} {
		rows := base.clone()
		rows[artifactFindingPath][0]["confidence_score"] = score
		if err := prepare(t, rows); err != nil {
			t.Fatal(err)
		}
	}
	for _, score := range []json.Number{"-0", "0.0", "1e2", "9223372036854775808"} {
		rows := base.clone()
		rows[artifactFindingPath][0]["confidence_score"] = score
		requireArtifactPortFailure(t, prepare(t, rows), "row_shape_exact")
	}
	for _, score := range []json.Number{"-1", "101", "9007199254740993"} {
		rows := base.clone()
		rows[artifactFindingPath][0]["confidence_score"] = score
		requireArtifactPortFailure(t, prepare(t, rows), "subtype_exact")
	}
	for _, binding := range []struct{ path, field string }{
		{artifactParentPath, "artifact_type"}, {artifactParentPath, "comm_type"}, {artifactParentPath, "closure_state"},
		{artifactFindingPath, "kind"}, {artifactFindingPath, "state"}, {artifactKeywordPath, "match_mode"},
	} {
		rows := base.clone()
		rows[binding.path][0][binding.field] = "unknown"
		requireArtifactPortFailure(t, prepare(t, rows), "subtype_exact")
	}
	rows := base.clone()
	rows[artifactParentPath][0]["title"], rows[artifactParentPath][0]["body"] = nil, nil
	if err := prepare(t, rows); err != nil {
		t.Fatalf("cleared retained Note: %v", err)
	}
	for _, binding := range []struct {
		index  int
		fields string
	}{
		{1, "comm_id comm_type timestamp_utc audience channel_or_meeting summary"},
		{2, "handoff_id timestamp_utc outgoing_owner_user_id incoming_owner_user_id current_state_summary"},
		{3, "status_review_id timestamp_utc review_owner_user_id current_state_summary"},
		{4, "lesson_id timestamp_utc summary owner_user_id"},
	} {
		for _, field := range strings.Fields(binding.fields) {
			rows := base.clone()
			rows[artifactParentPath][binding.index][field] = nil
			requireArtifactPortFailure(t, prepare(t, rows), "subtype_exact")
		}
	}
	rows = base.clone()
	rows[artifactRiskPath][0]["normalized_risk_ref_text"] = "canonical risk"
	requireArtifactPortFailure(t, prepare(t, rows), "row_shape_exact")
	requireArtifactPortableTextCases(t, base, prepare)
	requireArtifactPortableLifecycleCases(t, base, prepare)
	requireArtifactPortableFailurePriority(t, base, prepare)
	for _, path := range artifactPortPaths {
		for _, raw := range []string{"null\n", "\n", "{} {}\n", "{\"record_id\":\"one\",\"record_id\":\"two\"}\n", "{\"value\":\"\\ud800\"}\n", string([]byte{'{', '"', 'v', '"', ':', '"', 0xff, '"', '}', '\n'})} {
			bundle := artifactPortBundle(t, base)
			bundle[path] = []byte(raw)
			_, err := port.PrepareImport(ctx, bundle, input)
			requireArtifactPortFailure(t, err, "row_shape_exact")
		}
	}
	for _, path := range []string{artifactParentPath, artifactRiskPath} {
		rows := base.clone()
		rows[path][0], rows[path][1] = rows[path][1], rows[path][0]
		requireArtifactPortFailure(t, prepare(t, rows), "source_identity_admitted")
	}
	for _, binding := range []struct {
		path, field string
		values      []string
	}{
		{artifactParentPath, "comm_type", []string{"meeting", "notification", "approval", "briefing", "handoff"}},
		{artifactParentPath, "closure_state", []string{"open", "closed"}},
		{artifactFindingPath, "kind", []string{"finding", "hypothesis"}},
		{artifactKeywordPath, "match_mode", []string{"literal", "regex"}},
	} {
		for _, value := range binding.values {
			rows := base.clone()
			rows[binding.path][0][binding.field] = value
			if err := prepare(t, rows); err != nil {
				t.Fatalf("known %s token %q: %v", binding.field, value, err)
			}
		}
	}
	for _, reverse := range []bool{false, true} {
		early, late := artifactParentPath, artifactRiskPath
		if reverse {
			early, late = late, early
		}
		rows := base.clone()
		key := "record_id"
		if early == artifactRiskPath {
			key = "risk_ref_id"
		}
		rows[early][0][key] = "invalid"
		rows[late][0]["unknown"] = true
		requireArtifactPortFailure(t, prepare(t, rows), "source_identity_admitted")
		bundle := artifactPortBundle(t, rows)
		bundle[late] = []byte("null\n")
		_, err := port.PrepareImport(ctx, bundle, input)
		requireArtifactPortFailure(t, err, "row_shape_exact")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := port.PrepareImport(cancelled, artifactPortBundle(t, base), input); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Prepare=%v", err)
	}
}

func requireArtifactPortableTextCases(t *testing.T, base artifactPortableRows, prepare func(testing.TB, artifactPortableRows) error) {
	t.Helper()
	for _, binding := range []struct {
		path, field string
		limit       int
		multiline   bool
	}{
		{artifactParentPath, "title", 512, false}, {artifactParentPath, "body", 16384, true},
		{artifactParentPath, "audience", 256, false}, {artifactParentPath, "channel_or_meeting", 512, false},
		{artifactParentPath, "summary", 512, false}, {artifactParentPath, "privilege_tag", 512, false},
		{artifactParentPath, "current_state_summary", 16384, true}, {artifactParentPath, "next_checks", 16384, true}, {artifactParentPath, "active_risks_summary", 16384, true},
		{artifactFindingPath, "statement", 16384, true},
		{artifactQueryPath, "platform", 512, false}, {artifactQueryPath, "purpose", 512, false}, {artifactQueryPath, "query_text", 16384, true},
		{artifactKeywordPath, "pattern", 512, false}, {artifactKeywordPath, "reason", 4096, true}, {artifactRiskPath, "risk_ref_text", 512, false},
	} {
		t.Run(binding.path+"/"+binding.field, func(t *testing.T) {
			for _, good := range []string{"é  preserved\u200d text", strings.Repeat("😀", binding.limit)} {
				rows := base.clone()
				rows[binding.path][0][binding.field] = good
				if err := prepare(t, rows); err != nil {
					t.Fatalf("canonical text: %v", err)
				}
			}
			bad := []string{"", " ", " padded", "padded ", "e\u0301", strings.Repeat("😀", binding.limit+1), "before\r\nafter", "before\rafter"}
			for r := rune(0); r <= 159; r++ {
				if r > 31 && r < 127 {
					continue
				}
				if binding.multiline && (r == '\n' || r == '\t') {
					continue
				}
				bad = append(bad, "before"+string(r)+"after")
			}
			for _, value := range bad {
				rows := base.clone()
				rows[binding.path][0][binding.field] = value
				requireArtifactPortFailure(t, prepare(t, rows), "subtype_exact")
			}
			rows := base.clone()
			rows[binding.path][0][binding.field] = "before\n\tafter"
			err := prepare(t, rows)
			if binding.multiline {
				if err != nil {
					t.Fatalf("canonical multiline: %v", err)
				}
			} else {
				requireArtifactPortFailure(t, err, "subtype_exact")
			}
		})
	}
	for _, binding := range []struct{ path, fields string }{
		{artifactParentPath, "comm_id handoff_id status_review_id lesson_id"}, {artifactQueryPath, "query_id"}, {artifactKeywordPath, "keyword_id"},
	} {
		for _, field := range strings.Fields(binding.fields) {
			for _, bad := range []string{"", " \t ", "bad\x00label"} {
				rows := base.clone()
				rows[binding.path][0][field] = bad
				requireArtifactPortFailure(t, prepare(t, rows), "subtype_exact")
			}
		}
	}
	for _, binding := range []struct{ path, fields string }{
		{artifactParentPath, "outgoing_owner_user_id incoming_owner_user_id review_owner_user_id owner_user_id created_by_user_id"},
		{artifactFindingPath, "owner_user_id"}, {artifactQueryPath, "created_by_user_id"}, {artifactRiskPath, "created_by_user_id deleted_by_user_id"},
	} {
		for _, field := range strings.Fields(binding.fields) {
			rows := base.clone()
			index := 0
			if field == "deleted_by_user_id" {
				index = 2
			}
			rows[binding.path][index][field] = uuid.NewString()
			requireArtifactPortFailure(t, prepare(t, rows), "attribution_exact")
		}
	}
}
func requireArtifactPortableLifecycleCases(t *testing.T, base artifactPortableRows, prepare func(testing.TB, artifactPortableRows) error) {
	t.Helper()
	for _, path := range []string{artifactParentPath, artifactFindingPath, artifactQueryPath, artifactKeywordPath} {
		rows := base.clone()
		rows[path][0]["updated_at"] = "2000-01-01T00:00:00Z"
		requireArtifactPortFailure(t, prepare(t, rows), "lifecycle_fields_legal")
	}
	for _, change := range []map[string]any{
		{"state": "open"}, {"closed_at": nil}, {"closed_at": "2000-01-01T00:00:00Z"}, {"closed_at": "2030-01-01T00:00:00Z"},
	} {
		rows := base.clone()
		for k, v := range change {
			rows[artifactFindingPath][0][k] = v
		}
		requireArtifactPortFailure(t, prepare(t, rows), "lifecycle_fields_legal")
	}
	rows := base.clone()
	rows[artifactFindingPath][0]["state"], rows[artifactFindingPath][0]["closed_at"] = "open", nil
	if err := prepare(t, rows); err != nil {
		t.Fatalf("open Finding control: %v", err)
	}
	for _, change := range []map[string]any{
		{"deleted_at": nil}, {"deleted_by_user_id": nil}, {"deleted_at": "2000-01-01T00:00:00Z"},
	} {
		rows := base.clone()
		for k, v := range change {
			rows[artifactRiskPath][2][k] = v
		}
		requireArtifactPortFailure(t, prepare(t, rows), "lifecycle_fields_legal")
	}
}
func requireArtifactPortableFailurePriority(t *testing.T, base artifactPortableRows, prepare func(testing.TB, artifactPortableRows) error) {
	t.Helper()
	// Exchange competing defects between early and late rows without changing
	// stable-identity order. The expected phase is independent of traversal order.
	for _, swapped := range []bool{false, true} {
		first, last := 0, 7
		if swapped {
			first, last = last, first
		}
		for _, pair := range []struct {
			earlyField string
			earlyValue any
			lateField  string
			lateValue  any
			want       string
		}{
			{"title", []any{}, "created_by_user_id", uuid.NewString(), "row_shape_exact"},
			{"incident_id", uuid.NewString(), "title", " unnormalized", "references_same_incident"},
			{"title", " unnormalized", "updated_at", "2000-01-01T00:00:00Z", "subtype_exact"},
			{"updated_at", "2000-01-01T00:00:00Z", "created_by_user_id", uuid.NewString(), "lifecycle_fields_legal"},
		} {
			rows := base.clone()
			rows[artifactParentPath][first][pair.earlyField] = pair.earlyValue
			rows[artifactParentPath][last][pair.lateField] = pair.lateValue
			requireArtifactPortFailure(t, prepare(t, rows), pair.want)
		}
	}
}

type artifactPortAttributions struct {
	local uuid.UUID
	rows  []incidentportability.ImportedAttribution
}

func (a *artifactPortAttributions) RecordImportedAttribution(table, row, column, actor string) error {
	a.rows = append(a.rows, incidentportability.ImportedAttribution{SourceTable: table, SourceRowID: row, SourceColumn: column, SourceActorID: actor, LocalUserID: a.local})
	return nil
}
func (a *artifactPortAttributions) VisitImportedAttributions(table, row string, visit func(incidentportability.ImportedAttribution) error) error {
	for _, value := range a.rows {
		if value.SourceTable == table && (row == "" || value.SourceRowID == row) {
			if err := visit(value); err != nil {
				return err
			}
		}
	}
	return nil
}
func (a *artifactPortAttributions) ResolvePortableSourceActors(_ context.Context, _ incidentportability.Queryer, _ uuid.UUID, table, column string, ids []string) (map[string]string, error) {
	result := map[string]string{}
	for _, id := range ids {
		for _, value := range a.rows {
			if value.SourceTable == table && value.SourceColumn == column && value.SourceRowID == id {
				result[id] = value.SourceActorID
			}
		}
	}
	return result, nil
}

func TestArtifactIncidentBundleTransaction_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "artifact-source-port")
	actor, otherActor, incident, foreign := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{actor, otherActor} {
		if _, err := db.Exec(ctx, `INSERT INTO users (id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES ($1,$2,'Artifact source actor','test-only',false,true,true)`, id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []uuid.UUID{incident, foreign} {
		if _, err := db.Exec(ctx, `INSERT INTO incidents (id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES ($1,$2,$2,'Artifact source','active',$3,$3)`, id, "ART-"+id.String(), actor); err != nil {
			t.Fatal(err)
		}
	}
	port, err := artifacts.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatal(err)
	}
	begin := func(t testing.TB) pgx.Tx {
		t.Helper()
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		return tx
	}
	seed := func(t testing.TB, tx pgx.Tx, rows artifactPortableRows) {
		t.Helper()
		for _, row := range rows[artifactParentPath] {
			if _, err := tx.Exec(ctx, `INSERT INTO records (record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES ($1,$2,'artifact',$3,$3)`, row["record_id"], incident, actor); err != nil {
				t.Fatal(err)
			}
		}
	}
	prepare := func(t testing.TB, rows artifactPortableRows) (sourceport.Prepared, sourceport.ImportContext, *artifactPortAttributions) {
		t.Helper()
		input := artifactPortInput(t, incident, actor)
		prepared, err := port.PrepareImport(ctx, artifactPortBundle(t, rows), input)
		if err != nil {
			t.Fatal(err)
		}
		attributions := &artifactPortAttributions{local: actor}
		input.Attributions = attributions
		return prepared, input, attributions
	}
	for _, variant := range []string{"all fields", "nullable and deleted envelopes", "page boundaries"} {
		t.Run(variant, func(t *testing.T) {
			tx := begin(t)
			rows := artifactPortableFixture(incident)
			switch variant {
			case "nullable and deleted envelopes":
				for _, field := range []string{"title", "body", "timestamp_utc", "comm_id", "comm_type", "audience", "channel_or_meeting", "summary", "next_report_at", "privilege_tag", "handoff_id", "outgoing_owner_user_id", "incoming_owner_user_id", "current_state_summary", "next_checks", "acknowledged_at", "status_review_id", "review_owner_user_id", "active_risks_summary", "lesson_id", "owner_user_id", "closure_state", "created_by_user_id"} {
					rows[artifactParentPath][0][field] = nil
				}
				rows[artifactFindingPath][0]["confidence_score"] = nil
				rows[artifactFindingPath][0]["state"] = "open"
				rows[artifactFindingPath][0]["closed_at"] = nil
			case "page boundaries":
				for i := 9; i <= 131; i++ {
					rows[artifactParentPath] = append(rows[artifactParentPath], artifactPortableParent(incident, artifactPortRecordID(i), "note"))
				}
				for i := 4; i <= 130; i++ {
					risk := maps.Clone(rows[artifactRiskPath][0])
					risk["risk_ref_id"] = fmt.Sprintf("30000000-0000-4000-8000-%012d", i)
					risk["risk_ref_text"] = fmt.Sprintf("Risk %d", i)
					rows[artifactRiskPath] = append(rows[artifactRiskPath], risk)
				}
			}
			seed(t, tx, rows)
			if variant == "nullable and deleted envelopes" {
				if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at=updated_at,deleted_by_user_id=$2 WHERE incident_id=$1`, incident, actor); err != nil {
					t.Fatal(err)
				}
			}
			prepared, input, attributions := prepare(t, rows)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			requireArtifactPortableStored(t, tx, rows, actor, attributions)
			files, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: tx, IncidentID: incident, PortableAttributions: attributions})
			if err != nil || len(files) != 5 {
				t.Fatalf("export files=%d err=%v", len(files), err)
			}
			remaining := map[string]bool{}
			for _, path := range artifactPortPaths {
				remaining[path] = true
			}
			for _, file := range files {
				if !remaining[file.Path] {
					t.Fatalf("unexpected or duplicate export path %s", file.Path)
				}
				delete(remaining, file.Path)
				actual, err := incidentportability.DecodeNDJSON(file.Payload)
				if err != nil || !reflect.DeepEqual(actual, rows[file.Path]) {
					t.Fatalf("export %s differs from independently authored source: err=%v got=%#v want=%#v", file.Path, err, actual, rows[file.Path])
				}
			}
			if variant == "all fields" {

				for _, path := range artifactPortPaths {
					t.Run("extra stored row "+path, func(t *testing.T) {
						fewer := rows.clone()
						fewer[path] = fewer[path][:len(fewer[path])-1]
						// For parents remove the Note, preserving all subtype parents.
						if path == artifactParentPath {
							fewer[path] = rows[path][1:]
						}
						shortPrepared, shortInput, _ := prepare(t, fewer)
						shortInput.Attributions = attributions
						requireArtifactPortFailure(t, port.ValidateImportTx(ctx, tx, shortPrepared, shortInput), "source_state_exact")
					})
				}
				requireArtifactPortableStateMutations(t, tx, port, prepared, input, rows, otherActor)
				requireArtifactPortableSidecarMutations(t, tx, port, prepared, input, attributions)
				requireArtifactPortFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), "source_identity_admitted")
			}
			if err := tx.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			requireArtifactPortableEmpty(t, db, incident)
		})
	}
	// The last parent or a child may have a bad envelope. Every source relation
	// must still be empty when this preflight fails, without relying on rollback.
	for _, path := range []string{artifactParentPath, artifactFindingPath, artifactQueryPath, artifactKeywordPath} {
		for _, defect := range []string{"missing", "foreign", "wrong type"} {
			t.Run(path+" envelope "+defect, func(t *testing.T) {
				tx := begin(t)
				rows := artifactPortableFixture(incident)
				seed(t, tx, rows)
				id := rows[path][len(rows[path])-1]["record_id"]
				query := `DELETE FROM records WHERE record_id=$1`
				args := []any{id}
				if defect == "foreign" {
					query = `UPDATE records SET incident_id=$2 WHERE record_id=$1`
					args = append(args, foreign)
				}
				if defect == "wrong type" {
					query = `UPDATE records SET record_type='host' WHERE record_id=$1`
				}
				if _, err := tx.Exec(ctx, query, args...); err != nil {
					t.Fatal(err)
				}
				prepared, input, _ := prepare(t, rows)
				requireArtifactPortFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), "envelope_type_scope")
				requireArtifactPortableEmpty(t, tx, incident)
			})
		}
	}
	for _, defect := range []string{"missing finding", "missing query", "missing keyword", "wrong finding parent", "wrong query parent", "wrong keyword parent", "orphan finding", "orphan query", "orphan keyword", "missing risk parent", "wrong risk parent", "foreign risk parent", "duplicate active risk", "duplicate query ID", "duplicate keyword ID"} {
		t.Run(defect, func(t *testing.T) {
			tx := begin(t)
			rows := artifactPortableFixture(incident)
			seed(t, tx, rows)
			want := "subtype_exact"
			validateFailure := false
			switch defect {
			case "missing finding":
				rows[artifactFindingPath] = nil
				validateFailure = true
			case "missing query":
				rows[artifactQueryPath] = nil
				validateFailure = true
			case "missing keyword":
				rows[artifactKeywordPath] = nil
				validateFailure = true
			case "wrong finding parent":
				rows[artifactFindingPath][0]["record_id"] = artifactPortRecordID(1).String()
			case "wrong query parent":
				rows[artifactQueryPath][0]["record_id"] = artifactPortRecordID(1).String()
			case "wrong keyword parent":
				rows[artifactKeywordPath][0]["record_id"] = artifactPortRecordID(1).String()
			case "orphan finding":
				rows[artifactParentPath] = append(rows[artifactParentPath][:5], rows[artifactParentPath][6:]...)
			case "orphan query":
				rows[artifactParentPath] = append(rows[artifactParentPath][:6], rows[artifactParentPath][7:]...)
			case "orphan keyword":
				rows[artifactParentPath] = rows[artifactParentPath][:7]
			case "missing risk parent":
				rows[artifactRiskPath][0]["handoff_record_id"] = uuid.NewString()
				want = "handoff_risk_target"
			case "wrong risk parent":
				rows[artifactRiskPath][0]["handoff_record_id"] = artifactPortRecordID(1).String()
				want = "handoff_risk_target"
			case "foreign risk parent":
				id := uuid.New()
				if _, err := tx.Exec(ctx, `INSERT INTO records (record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES ($1,$2,'artifact',$3,$3)`, id, foreign, actor); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `INSERT INTO artifacts (record_id,incident_id,artifact_type) VALUES ($1,$2,'handoff')`, id, foreign); err != nil {
					t.Fatal(err)
				}
				rows[artifactRiskPath][0]["handoff_record_id"] = id.String()
				want = "handoff_risk_target"
			case "duplicate active risk":
				rows[artifactRiskPath][1]["risk_ref_text"] = "Canonical Risk"
				want = "handoff_risk_legal"
			case "duplicate query ID", "duplicate keyword ID":
				path, kind := artifactQueryPath, "investigative_query"
				if defect == "duplicate keyword ID" {
					path, kind = artifactKeywordPath, "forensic_keyword"
				}
				id := artifactPortRecordID(9)
				parent := artifactPortableParent(incident, id, kind)
				rows[artifactParentPath] = append(rows[artifactParentPath], parent)
				seed(t, tx, artifactPortableRows{artifactParentPath: {parent}})
				child := maps.Clone(rows[path][0])
				child["record_id"] = id.String()
				rows[path] = append(rows[path], child)
			}
			prepared, input, _ := prepare(t, rows)
			err := port.ApplyImportTx(ctx, tx, prepared, input)
			if validateFailure {
				if err != nil {
					t.Fatal(err)
				}
				err = port.ValidateImportTx(ctx, tx, prepared, input)
			}
			requireArtifactPortFailure(t, err, want)
			// These failures occur after earlier source writes. The caller owns rollback.
			var parents int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM artifacts WHERE incident_id=$1`, incident).Scan(&parents); err != nil || parents == 0 {
				t.Fatalf("expected earlier writes: %d, %v", parents, err)
			}
			if err := tx.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			requireArtifactPortableEmpty(t, db, incident)
		})
	}
	t.Run("binding and immutable members", func(t *testing.T) {
		tx := begin(t)
		rows := artifactPortableFixture(incident)
		seed(t, tx, rows)
		input := artifactPortInput(t, incident, actor)
		bundle := artifactPortBundle(t, rows)
		prepared, err := port.PrepareImport(ctx, bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		input.Attributions = &artifactPortAttributions{local: actor}
		for _, field := range []string{"incident", "actor", "operation", "version", "attributions"} {
			other := input
			switch field {
			case "incident":
				other.IncidentID = uuid.New()
			case "actor":
				other.ActorUserID = uuid.New()
			case "operation":
				other.OperationID = "another"
			case "version":
				other.BundleVersion = 7
			case "attributions":
				other.Attributions = nil
			}
			for _, err := range []error{port.ApplyImportTx(ctx, tx, prepared, other), port.ValidateImportTx(ctx, tx, prepared, other)} {
				if !errors.Is(err, sourceport.ErrPreparedBinding) {
					t.Fatalf("%s binding=%v", field, err)
				}
			}
		}
		other, err := artifacts.NewIncidentBundleSourcePort()
		if err != nil {
			t.Fatal(err)
		}
		for _, err := range []error{other.ApplyImportTx(ctx, tx, prepared, input), other.ValidateImportTx(ctx, tx, prepared, input), port.ApplyImportTx(ctx, nil, prepared, input), port.ValidateImportTx(ctx, nil, prepared, input)} {
			if !errors.Is(err, sourceport.ErrPreparedBinding) {
				t.Fatalf("foreign/nil binding=%v", err)
			}
		}
		bundle[artifactRiskPath] = append(bundle[artifactRiskPath], ' ')
		var verification *incidentportability.VerificationFailure
		if err := port.ApplyImportTx(ctx, tx, prepared, input); !errors.As(err, &verification) {
			t.Fatalf("changed bytes=%v", err)
		}
		requireArtifactPortableEmpty(t, tx, incident)
	})
}

// Fixed test expectations do not derive actor meaning from column-name suffixes.
var artifactExpectedActorColumns = map[string][]string{
	artifactParentPath:  {"created_by_user_id", "incoming_owner_user_id", "outgoing_owner_user_id", "owner_user_id", "review_owner_user_id"},
	artifactFindingPath: {"owner_user_id"}, artifactQueryPath: {"created_by_user_id"}, artifactKeywordPath: {}, artifactRiskPath: {"created_by_user_id", "deleted_by_user_id"},
}

func requireArtifactPortableStored(t testing.TB, tx pgx.Tx, source artifactPortableRows, actor uuid.UUID, attributions *artifactPortAttributions) {
	t.Helper()
	expectedActors := map[incidentportability.ImportedAttribution]bool{}
	for _, path := range artifactPortPaths {
		table := strings.TrimSuffix(strings.TrimPrefix(path, "data/"), ".ndjson")
		key := "record_id"
		if path == artifactRiskPath {
			key = "risk_ref_id"
		}
		for _, row := range source[path] {
			expected := maps.Clone(row)
			for _, field := range artifactExpectedActorColumns[path] {
				if raw := row[field]; raw != nil {
					expected[field] = actor.String()
					expectedActors[incidentportability.ImportedAttribution{SourceTable: table, SourceRowID: row[key].(string), SourceColumn: field, SourceActorID: raw.(string), LocalUserID: actor}] = true
				}
			}
			var raw []byte
			if err := tx.QueryRow(t.Context(), `SELECT to_jsonb(s) FROM `+table+` s WHERE `+key+`=$1`, row[key]).Scan(&raw); err != nil {
				t.Fatal(err)
			}
			actual, err := incidentportability.DecodeNDJSON(raw)
			if err != nil || !reflect.DeepEqual(actual, []map[string]any{expected}) {
				t.Fatalf("stored %s/%s: err=%v got=%#v want=%#v", table, row[key], err, actual, expected)
			}
		}
	}
	for _, actual := range attributions.rows {
		if !expectedActors[actual] {
			t.Fatalf("unexpected/duplicate attribution %#v", actual)
		}
		delete(expectedActors, actual)
	}
	if len(expectedActors) != 0 {
		t.Fatalf("missing attributions %#v", expectedActors)
	}
}
func requireArtifactPortableEmpty(t testing.TB, query incidentportability.Queryer, incident uuid.UUID) {
	t.Helper()
	for _, table := range []string{"artifacts", "artifact_findings", "artifact_investigative_queries", "artifact_forensic_keywords", "handoff_risk_refs"} {
		var count int
		if err := query.QueryRow(t.Context(), `SELECT count(*) FROM `+table+` WHERE incident_id=$1`, incident).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s remaining=%d err=%v", table, count, err)
		}
	}
}
func requireArtifactPortableStateMutations(t *testing.T, tx pgx.Tx, port sourceport.Port, prepared sourceport.Prepared, input sourceport.ImportContext, rows artifactPortableRows, otherActor uuid.UUID) {
	t.Helper()
	for _, path := range artifactPortPaths {
		table := strings.TrimSuffix(strings.TrimPrefix(path, "data/"), ".ndjson")
		key := "record_id"
		if path == artifactRiskPath {
			key = "risk_ref_id"
		}
		for field, value := range rows[path][0] {
			if field == key || field == "incident_id" || field == "handoff_record_id" || field == "deleted_at" || field == "deleted_by_user_id" {
				continue
			}
			t.Run(table+" changes "+field, func(t *testing.T) {
				nested, err := tx.Begin(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = nested.Rollback(t.Context()) }()
				expression := "$2"
				args := []any{rows[path][0][key], "Changed canonical source"}
				switch field {
				case "artifact_type":
					args[1] = "comm_log"
				case "comm_type":
					args[1] = "approval"
				case "closure_state":
					args[1] = "closed"
				case "kind":
					args[1] = "finding"
				case "state":
					args[1] = "open"
				case "match_mode":
					args[1] = "literal"
				case "confidence_score":
					args[1] = 42
				case "case_sensitive":
					args[1] = false
				case "timestamp_utc", "created_at", "updated_at", "next_report_at", "acknowledged_at", "closed_at":
					expression = field + `+interval '1 microsecond'`
					args = args[:1]
				default:
					if slices.Contains(artifactExpectedActorColumns[path], field) {
						args[1] = otherActor
					}
				}
				if value == nil {
					t.Fatalf("mutation field unexpectedly NULL: %s", field)
				}
				if _, err := nested.Exec(t.Context(), `UPDATE `+table+` SET `+field+`=`+expression+` WHERE `+key+`=$1`, args...); err != nil {
					t.Fatal(err)
				}
				requireArtifactPortFailure(t, port.ValidateImportTx(t.Context(), nested, prepared, input), "source_state_exact")
			})
		}
	}
	for _, change := range []struct{ name, sql, want string }{
		{"parent null", `UPDATE artifacts SET body=NULL WHERE artifact_type='note'`, "source_state_exact"},
		{"finding confidence null", `UPDATE artifact_findings SET confidence_score=NULL`, "source_state_exact"},
		{"risk deleted", `UPDATE handoff_risk_refs SET deleted_at=created_at,deleted_by_user_id=created_by_user_id WHERE risk_ref_text='canonical risk'`, "source_state_exact"},
		{"missing note", `DELETE FROM artifacts WHERE artifact_type='note'`, "source_state_exact"},
		{"missing finding", `DELETE FROM artifact_findings`, "subtype_exact"},
		{"missing query", `DELETE FROM artifact_investigative_queries`, "subtype_exact"},
		{"missing keyword", `DELETE FROM artifact_forensic_keywords`, "subtype_exact"},
		{"missing risk", `DELETE FROM handoff_risk_refs`, "source_state_exact"},
		{"risk wrong target", `UPDATE handoff_risk_refs SET handoff_record_id='10000000-0000-4000-8000-000000000001'`, "handoff_risk_target"},
		{"envelope before state", `UPDATE records SET record_type='host' WHERE record_id='10000000-0000-4000-8000-000000000001'`, "envelope_type_scope"},
	} {
		t.Run(change.name, func(t *testing.T) {
			nested, err := tx.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = nested.Rollback(t.Context()) }()
			query := change.sql + ` WHERE incident_id=$1`
			if strings.Contains(change.sql, " WHERE ") {
				query = change.sql + ` AND incident_id=$1`
			}
			if _, err := nested.Exec(t.Context(), query, input.IncidentID); err != nil {
				t.Fatal(err)
			}
			requireArtifactPortFailure(t, port.ValidateImportTx(t.Context(), nested, prepared, input), change.want)
		})
	}
}
func requireArtifactPortableSidecarMutations(t *testing.T, tx pgx.Tx, port sourceport.Port, prepared sourceport.Prepared, input sourceport.ImportContext, attributions *artifactPortAttributions) {
	t.Helper()
	for _, change := range []string{"missing", "extra row", "duplicate", "changed actor", "changed local", "changed column", "changed table", "unexpected keyword"} {
		t.Run("sidecar "+change, func(t *testing.T) {
			saved := slices.Clone(attributions.rows)
			defer func() { attributions.rows = saved }()
			switch change {
			case "missing":
				attributions.rows = attributions.rows[1:]
			case "extra row":
				extra := attributions.rows[0]
				extra.SourceRowID = uuid.NewString()
				attributions.rows = append(attributions.rows, extra)
			case "duplicate":
				attributions.rows = append(attributions.rows, attributions.rows[0])
			case "changed actor":
				attributions.rows[0].SourceActorID = uuid.NewString()
			case "changed local":
				attributions.rows[0].LocalUserID = uuid.New()
			case "changed column":
				attributions.rows[0].SourceColumn = "not_an_actor"
			case "changed table":
				attributions.rows[0].SourceTable = "artifact_forensic_keywords"
			case "unexpected keyword":
				attributions.rows = append(attributions.rows, incidentportability.ImportedAttribution{SourceTable: "artifact_forensic_keywords", SourceRowID: artifactPortRecordID(8).String(), SourceColumn: "creator_user_id", SourceActorID: artifactSourceActor.String(), LocalUserID: input.ActorUserID})
			}
			requireArtifactPortFailure(t, port.ValidateImportTx(t.Context(), tx, prepared, input), "attribution_exact")
		})
	}
}
