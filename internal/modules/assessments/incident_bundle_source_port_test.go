package assessments_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

const assessmentBundlePath = "data/compromise_assessments.ndjson"

var assessmentHistoricalActor = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
var assessmentHistoricalDeleter = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")

func assessmentPortableRow(incident, record, subject uuid.UUID) map[string]any {
	return map[string]any{
		"record_id": record.String(), "incident_id": incident.String(), "subject_record_id": subject.String(),
		"subject_type": "host", "assessment_state": "suspected", "confidence_score": json.Number("25"),
		"assessor_user_id": assessmentHistoricalActor.String(), "assessed_at": "2000-01-01T00:00:00.000001Z",
		"created_at": "2026-01-01T00:00:00.123456Z", "updated_at": "2026-01-03T00:00:00+00:00",
		"deleted_at": nil, "deleted_by_user_id": nil, "rationale": "Evidence\n\tReason",
	}
}
func assessmentBundle(t testing.TB, rows ...map[string]any) portabilitysupport.MapBundle {
	t.Helper()
	var payload bytes.Buffer
	encoder := json.NewEncoder(&payload)
	for _, row := range rows {
		if err := encoder.Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	return portabilitysupport.MapBundle{assessmentBundlePath: payload.Bytes()}
}
func assessmentImportContext(t testing.TB, incident, local uuid.UUID) sourceport.ImportContext {
	t.Helper()
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{
		{SourceActorID: assessmentHistoricalActor.String(), DisplayName: "Historical assessor"},
		{SourceActorID: assessmentHistoricalDeleter.String(), DisplayName: "Historical deleter"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return sourceport.ImportContext{IncidentID: incident, ActorUserID: local, OperationID: "assessment-portable", BundleVersion: 8, Actors: actors}
}
func requireAssessmentPortFailure(t testing.TB, err error, invariant string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID() != "assessments" || failure.InvariantID() != "assessments."+invariant {
		t.Fatalf("failure = %v, want assessments.%s", err, invariant)
	}
}

func TestAssessmentIncidentBundleAdmission_Unit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	incident, local, record, subject := uuid.New(), uuid.New(), uuid.MustParse("10000000-0000-4000-8000-000000000001"), uuid.New()
	input := assessmentImportContext(t, incident, local)
	base := assessmentPortableRow(incident, record, subject)
	port := assessments.NewIncidentBundleSourcePort()
	prepare := func(t testing.TB, rows ...map[string]any) error {
		t.Helper()
		bundle := assessmentBundle(t, rows...)
		before := bytes.Clone(bundle[assessmentBundlePath])
		_, err := port.PrepareImport(ctx, bundle, input)
		if !bytes.Equal(before, bundle[assessmentBundlePath]) {
			t.Fatal("preparation changed source bytes")
		}
		return err
	}
	for _, subjectType := range []string{"host", "identity"} {
		for _, state := range []string{"unknown", "suspected", "confirmed", "disproven", "cleared"} {
			for _, score := range []any{nil, json.Number("0"), json.Number("100")} {
				row := maps.Clone(base)
				row["subject_type"], row["assessment_state"], row["confidence_score"] = subjectType, state, score
				if err := prepare(t, row); err != nil {
					t.Fatalf("valid %s/%s/%v: %v", subjectType, state, score, err)
				}
			}
		}
	}
	deleted := maps.Clone(base)
	deleted["deleted_at"], deleted["deleted_by_user_id"] = "2026-01-02T00:00:00Z", assessmentHistoricalDeleter.String()
	deleted["assessed_at"] = "2030-01-01T00:00:00Z"
	if err := prepare(t, deleted); err != nil {
		t.Fatal(err)
	}
	if err := prepare(t); err != nil {
		t.Fatalf("empty member: %v", err)
	}
	for field := range base {
		t.Run("missing_"+field, func(t *testing.T) {
			row := maps.Clone(base)
			delete(row, field)
			invariant := "row_shape_exact"
			if field == "record_id" {
				invariant = "source_identity_admitted"
			}
			requireAssessmentPortFailure(t, prepare(t, row), invariant)
		})
		t.Run("mistyped_"+field, func(t *testing.T) {
			row := maps.Clone(base)
			row[field] = []any{}
			invariant := "row_shape_exact"
			if field == "record_id" {
				invariant = "source_identity_admitted"
			}
			requireAssessmentPortFailure(t, prepare(t, row), invariant)
		})
	}
	for _, test := range []struct {
		name, field string
		value       any
		invariant   string
	}{
		{"unknown member", "extra", "secret-source-value", "row_shape_exact"},
		{"zero identity", "record_id", uuid.Nil.String(), "source_identity_admitted"},
		{"uppercase identity", "record_id", strings.ToUpper(assessmentHistoricalActor.String()), "source_identity_admitted"},
		{"wrong incident", "incident_id", uuid.NewString(), "subject_type_scope"},
		{"invalid subject", "subject_type", "artifact", "subject_type_scope"},
		{"self subject", "subject_record_id", record.String(), "subject_type_scope"},
		{"unknown state", "assessment_state", "contained", "state_confidence_rationale_legal"},
		{"negative score", "confidence_score", json.Number("-1"), "state_confidence_rationale_legal"},
		{"large score", "confidence_score", json.Number("101"), "state_confidence_rationale_legal"},
		{"exact large score", "confidence_score", json.Number("9007199254740993"), "state_confidence_rationale_legal"},
		{"overflow score", "confidence_score", json.Number("9223372036854775808"), "row_shape_exact"},
		{"exponent score", "confidence_score", json.Number("1e1"), "row_shape_exact"},
		{"fraction score", "confidence_score", json.Number("1.0"), "row_shape_exact"},
		{"negative zero", "confidence_score", json.Number("-0"), "row_shape_exact"},
		{"empty rationale", "rationale", "", "state_confidence_rationale_legal"},
		{"trim rationale", "rationale", " Reason ", "state_confidence_rationale_legal"},
		{"NFC rationale", "rationale", "Cafe\u0301", "state_confidence_rationale_legal"},
		{"CR rationale", "rationale", "Reason\r\nmore", "state_confidence_rationale_legal"},
		{"control rationale", "rationale", "Reason\x00", "state_confidence_rationale_legal"},
		{"format rationale", "rationale", "Reason\u200b", "state_confidence_rationale_legal"},
		{"non-UTC time", "assessed_at", "2026-01-01T01:00:00+01:00", "row_shape_exact"},
		{"nanosecond time", "assessed_at", "2026-01-01T00:00:00.1234567Z", "row_shape_exact"},
		{"fraction zero time", "assessed_at", "2026-01-01T00:00:00.100Z", "row_shape_exact"},
		{"year zero", "assessed_at", "0000-01-01T00:00:00Z", "row_shape_exact"},
		{"reversed time", "updated_at", "2000-01-01T00:00:00Z", "timestamps_lifecycle_legal"},
		{"deletion without actor", "deleted_at", "2026-01-02T00:00:00Z", "timestamps_lifecycle_legal"},
		{"actor without deletion", "deleted_by_user_id", assessmentHistoricalDeleter.String(), "timestamps_lifecycle_legal"},
		{"unknown actor", "assessor_user_id", uuid.NewString(), "attribution_exact"},
	} {
		t.Run(test.name, func(t *testing.T) {
			row := maps.Clone(base)
			row[test.field] = test.value
			err := prepare(t, row)
			requireAssessmentPortFailure(t, err, test.invariant)
			if strings.Contains(err.Error(), "secret-source-value") {
				t.Fatal("source value disclosed")
			}
		})
	}
	for _, field := range []string{"incident_id", "subject_record_id", "assessor_user_id", "deleted_by_user_id"} {
		for _, value := range []string{uuid.Nil.String(), strings.ToUpper(assessmentHistoricalActor.String()), "uuid"} {
			row := maps.Clone(base)
			row[field] = value
			requireAssessmentPortFailure(t, prepare(t, row), "row_shape_exact")
		}
	}
	for _, timestamp := range []string{"2025-01-01T00:00:00Z", "2030-01-01T00:00:00Z"} {
		row := maps.Clone(deleted)
		row["deleted_at"] = timestamp
		requireAssessmentPortFailure(t, prepare(t, row), "timestamps_lifecycle_legal")
	}
	unknownDeleter := maps.Clone(deleted)
	unknownDeleter["deleted_by_user_id"] = uuid.NewString()
	requireAssessmentPortFailure(t, prepare(t, unknownDeleter), "attribution_exact")
	requireAssessmentPortFailure(t, prepare(t, base, base), "source_identity_admitted")
	later := maps.Clone(base)
	later["record_id"] = "20000000-0000-4000-8000-000000000002"
	requireAssessmentPortFailure(t, prepare(t, later, base), "source_identity_admitted")
	for _, raw := range []string{"{} {}\n", "[]\n", "\n", "{\"record_id\":\"a\",\"record_id\":\"b\"}\n", "{\"bad\":\"\xff\"}\n"} {
		_, err := port.PrepareImport(ctx, portabilitysupport.MapBundle{assessmentBundlePath: []byte(raw)}, input)
		requireAssessmentPortFailure(t, err, "row_shape_exact")
	}
	// Swap the locations of competing defects while preserving admitted ID order.
	for _, reverse := range []bool{false, true} {
		first, second := maps.Clone(base), maps.Clone(later)
		if reverse {
			first["rationale"] = ""
			second["subject_type"] = "artifact"
		} else {
			first["subject_type"] = "artifact"
			second["rationale"] = ""
		}
		requireAssessmentPortFailure(t, prepare(t, first, second), "subject_type_scope")
		first, second = maps.Clone(base), maps.Clone(later)
		if reverse {
			first["updated_at"] = "2000-01-01T00:00:00Z"
			second["rationale"] = ""
		} else {
			first["rationale"] = ""
			second["updated_at"] = "2000-01-01T00:00:00Z"
		}
		requireAssessmentPortFailure(t, prepare(t, first, second), "state_confidence_rationale_legal")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := port.PrepareImport(cancelled, assessmentBundle(t, base), input); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v", err)
	}
}

type assessmentAttributions struct {
	local uuid.UUID
	rows  []incidentportability.ImportedAttribution
}

func (a *assessmentAttributions) RecordImportedAttribution(table, row, column, actor string) error {
	a.rows = append(a.rows, incidentportability.ImportedAttribution{SourceTable: table, SourceRowID: row, SourceColumn: column, SourceActorID: actor, LocalUserID: a.local})
	return nil
}
func (a *assessmentAttributions) VisitImportedAttributions(table, row string, visit func(incidentportability.ImportedAttribution) error) error {
	for _, value := range a.rows {
		if value.SourceTable == table && (row == "" || value.SourceRowID == row) {
			if err := visit(value); err != nil {
				return err
			}
		}
	}
	return nil
}
func (a *assessmentAttributions) ResolvePortableSourceActors(_ context.Context, _ incidentportability.Queryer, _ uuid.UUID, table, column string, ids []string) (map[string]string, error) {
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

func TestAssessmentIncidentBundleTransaction_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "assessment-source-port")
	actor, incident, foreign := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(ctx, `INSERT INTO users (id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES ($1,$2,'Assessment source actor','test-only',false,true,true)`, actor, actor.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []uuid.UUID{incident, foreign} {
		if _, err := db.Exec(ctx, `INSERT INTO incidents (id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES ($1,$2,$2,'Assessment source','active',$3,$3)`, id, "ASSESS-"+id.String(), actor); err != nil {
			t.Fatal(err)
		}
	}
	seed := func(id, scope uuid.UUID, kind string) {
		t.Helper()
		if _, err := db.Exec(ctx, `INSERT INTO records (record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES ($1,$2,$3,$4,$4)`, id, scope, kind, actor); err != nil {
			t.Fatal(err)
		}
	}
	port := assessments.NewIncidentBundleSourcePort()
	begin := func(t testing.TB) pgx.Tx {
		t.Helper()
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		return tx
	}
	for _, kind := range []string{"host", "identity"} {
		for _, deleted := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s_deleted_%t", kind, deleted), func(t *testing.T) {
				record, subject := uuid.New(), uuid.New()
				seed(record, incident, "assessment")
				seed(subject, incident, kind)
				row := assessmentPortableRow(incident, record, subject)
				row["subject_type"] = kind
				if kind == "identity" {
					row["confidence_score"] = nil
				}
				if deleted {
					row["deleted_at"], row["deleted_by_user_id"] = "2026-01-02T00:00:00Z", assessmentHistoricalDeleter.String()
					if _, err := db.Exec(ctx, `UPDATE records SET deleted_at=updated_at,deleted_by_user_id=$2 WHERE record_id IN ($1,$3)`, record, actor, subject); err != nil {
						t.Fatal(err)
					}
				}
				input := assessmentImportContext(t, incident, actor)
				bundle := assessmentBundle(t, row)
				prepared, err := port.PrepareImport(ctx, bundle, input)
				if err != nil {
					t.Fatal(err)
				}
				attributions := &assessmentAttributions{local: actor}
				input.Attributions = attributions
				tx := begin(t)
				if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
					t.Fatal(err)
				}
				if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
					t.Fatal(err)
				}
				var raw []byte
				if err := tx.QueryRow(ctx, `SELECT to_jsonb(a) FROM assessments a WHERE record_id=$1`, record).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				actual, err := incidentportability.DecodeNDJSON(raw)
				if err != nil {
					t.Fatal(err)
				}
				// Independent complete stored oracle, including explicit NULLs and runtime actor remapping.
				expected := map[string]any{"record_id": record.String(), "incident_id": incident.String(), "subject_record_id": subject.String(), "subject_type": kind, "assessment_state": "suspected", "confidence_score": json.Number("25"), "assessor_user_id": actor.String(), "assessed_at": "2000-01-01T00:00:00.000001+00:00", "created_at": "2026-01-01T00:00:00.123456+00:00", "updated_at": "2026-01-03T00:00:00+00:00", "deleted_at": nil, "deleted_by_user_id": nil, "rationale": "Evidence\n\tReason"}
				expectedActors := []incidentportability.ImportedAttribution{{SourceTable: "assessments", SourceRowID: record.String(), SourceColumn: "assessor_user_id", SourceActorID: assessmentHistoricalActor.String(), LocalUserID: actor}}
				if kind == "identity" {
					expected["confidence_score"] = nil
				}
				if deleted {
					expected["deleted_at"], expected["deleted_by_user_id"] = "2026-01-02T00:00:00+00:00", actor.String()
					expectedActors = append(expectedActors, incidentportability.ImportedAttribution{SourceTable: "assessments", SourceRowID: record.String(), SourceColumn: "deleted_by_user_id", SourceActorID: assessmentHistoricalDeleter.String(), LocalUserID: actor})
				}
				if !reflect.DeepEqual(actual, []map[string]any{expected}) || !reflect.DeepEqual(attributions.rows, expectedActors) {
					t.Fatalf("stored=%#v, actors=%#v; want %#v / %#v", actual, attributions.rows, expected, expectedActors)
				}
				files, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: tx, IncidentID: incident, PortableAttributions: attributions})
				if err != nil || len(files) != 1 || files[0].Path != assessmentBundlePath {
					t.Fatalf("export=%v err=%v", files, err)
				}
				expected["assessor_user_id"] = assessmentHistoricalActor.String()
				if deleted {
					expected["deleted_by_user_id"] = assessmentHistoricalDeleter.String()
				}
				exported, err := incidentportability.DecodeNDJSON(files[0].Payload)
				if err != nil || !reflect.DeepEqual(exported, []map[string]any{expected}) {
					t.Fatalf("export=%#v err=%v want=%#v", exported, err, expected)
				}
				for _, change := range []struct{ name, sql string }{
					{"judgment", `UPDATE assessments SET assessment_state='confirmed' WHERE record_id=$1`},
					{"rationale", `UPDATE assessments SET rationale='Different legal rationale' WHERE record_id=$1`},
					{"confidence", `UPDATE assessments SET confidence_score=99 WHERE record_id=$1`},
					{"assessed instant", `UPDATE assessments SET assessed_at=assessed_at+interval '1 microsecond' WHERE record_id=$1`},
					{"created instant", `UPDATE assessments SET created_at=created_at-interval '1 microsecond' WHERE record_id=$1`},
					{"updated instant", `UPDATE assessments SET updated_at=updated_at+interval '1 microsecond' WHERE record_id=$1`},
					{"missing source", `DELETE FROM assessments WHERE record_id=$1`},
				} {
					t.Run(change.name, func(t *testing.T) {
						nested, err := tx.Begin(ctx)
						if err != nil {
							t.Fatal(err)
						}
						defer func() { _ = nested.Rollback(ctx) }()
						if _, err := nested.Exec(ctx, change.sql, record); err != nil {
							t.Fatal(err)
						}
						requireAssessmentPortFailure(t, port.ValidateImportTx(ctx, nested, prepared, input), "source_state_exact")
					})
				}
				for _, change := range []string{"missing", "extra", "duplicate", "changed actor", "changed local"} {
					t.Run(change, func(t *testing.T) {
						saved := append([]incidentportability.ImportedAttribution(nil), attributions.rows...)
						defer func() { attributions.rows = saved }()
						switch change {
						case "missing":
							attributions.rows = attributions.rows[1:]
						case "extra":
							extra := attributions.rows[0]
							extra.SourceRowID = uuid.NewString()
							attributions.rows = append(attributions.rows, extra)
						case "duplicate":
							attributions.rows = append(attributions.rows, attributions.rows[0])
						case "changed actor":
							attributions.rows[0].SourceActorID = assessmentHistoricalDeleter.String()
						case "changed local":
							attributions.rows[0].LocalUserID = uuid.New()
						}
						requireAssessmentPortFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "attribution_exact")
					})
				}
				if err := tx.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
				var remaining int
				if err := db.QueryRow(ctx, `SELECT count(*) FROM assessments WHERE record_id=$1`, record).Scan(&remaining); err != nil || remaining != 0 {
					t.Fatalf("rollback count=%d err=%v", remaining, err)
				}
			})
		}
	}
	// A bad later reference must prevent even the first valid source insertion.
	for _, defect := range []string{"missing envelope", "wrong envelope type", "foreign envelope", "missing subject", "wrong subject type", "foreign subject"} {
		t.Run(defect, func(t *testing.T) {
			first, second := uuid.MustParse("10000000-0000-4000-8000-000000000001"), uuid.MustParse("20000000-0000-4000-8000-000000000002")
			tx := begin(t)
			// These fixtures are inside the nested transaction so identities can be reused.
			for _, id := range []uuid.UUID{first, second} {
				if _, err := tx.Exec(ctx, `INSERT INTO records (record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES ($1,$2,'assessment',$3,$3)`, id, incident, actor); err != nil {
					t.Fatal(err)
				}
			}
			subject := uuid.New()
			if _, err := tx.Exec(ctx, `INSERT INTO records (record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES ($1,$2,'host',$3,$3)`, subject, incident, actor); err != nil {
				t.Fatal(err)
			}
			a, b := assessmentPortableRow(incident, first, subject), assessmentPortableRow(incident, second, subject)
			switch defect {
			case "missing envelope":
				b["record_id"] = "30000000-0000-4000-8000-000000000003"
			case "wrong envelope type":
				if _, err := tx.Exec(ctx, `UPDATE records SET record_type='host' WHERE record_id=$1`, second); err != nil {
					t.Fatal(err)
				}
			case "foreign envelope":
				if _, err := tx.Exec(ctx, `UPDATE records SET incident_id=$2 WHERE record_id=$1`, second, foreign); err != nil {
					t.Fatal(err)
				}
			case "missing subject":
				b["subject_record_id"] = uuid.NewString()
			case "wrong subject type":
				b["subject_type"] = "identity"
			case "foreign subject":
				id := uuid.New()
				if _, err := tx.Exec(ctx, `INSERT INTO records (record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES ($1,$2,'host',$3,$3)`, id, foreign, actor); err != nil {
					t.Fatal(err)
				}
				b["subject_record_id"] = id.String()
			}
			input := assessmentImportContext(t, incident, actor)
			prepared, err := port.PrepareImport(ctx, assessmentBundle(t, a, b), input)
			if err != nil {
				t.Fatal(err)
			}
			input.Attributions = &assessmentAttributions{local: actor}
			requireAssessmentPortFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), "subject_type_scope")
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM assessments WHERE incident_id=$1`, incident).Scan(&count); err != nil || count != 0 {
				t.Fatalf("preflight wrote %d rows, err=%v", count, err)
			}
		})
	}
	t.Run("binding and immutable members", func(t *testing.T) {
		record, subject := uuid.New(), uuid.New()
		seed(record, incident, "assessment")
		seed(subject, incident, "host")
		input := assessmentImportContext(t, incident, actor)
		bundle := assessmentBundle(t, assessmentPortableRow(incident, record, subject))
		prepared, err := port.PrepareImport(ctx, bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		input.Attributions = &assessmentAttributions{local: actor}
		tx := begin(t)
		for _, field := range []string{"incident", "actor", "operation", "version"} {
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
			}
			if err := port.ApplyImportTx(ctx, tx, prepared, other); !errors.Is(err, sourceport.ErrPreparedBinding) {
				t.Fatalf("%s binding = %v", field, err)
			}
		}
		if err := assessments.NewIncidentBundleSourcePort().ApplyImportTx(ctx, tx, prepared, input); !errors.Is(err, sourceport.ErrPreparedBinding) {
			t.Fatalf("foreign port binding = %v", err)
		}
		bundle[assessmentBundlePath] = append(bundle[assessmentBundlePath], ' ')
		var verification *incidentportability.VerificationFailure
		if err := port.ApplyImportTx(ctx, tx, prepared, input); !errors.As(err, &verification) {
			t.Fatalf("changed member = %v", err)
		}
	})
}
