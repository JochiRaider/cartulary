package incidents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

var (
	incidentSourceTestID       = uuid.MustParse("11111111-1111-4a11-8b11-11111111111c")
	incidentSourceCreatedActor = uuid.MustParse("22222222-2222-4a22-8b22-22222222222c")
	incidentSourceUpdatedActor = uuid.MustParse("33333333-3333-4a33-8b33-33333333333c")
	incidentSourceImportActor  = uuid.MustParse("44444444-4444-4444-8444-444444444444")
)

func TestIncidentSourcePrepareEnforcesExclusiveInvariantPrecedence_Unit(t *testing.T) {
	contract := mustIncidentSourceContract(t)
	importContext := newIncidentSourceImportContextForTest(t, "codec-precedence")

	tests := []struct {
		name      string
		payload   func() []byte
		invariant string
	}{
		{
			name: "missing identity outranks unknown member",
			payload: func() []byte {
				row := validIncidentSourceObject()
				delete(row, "id")
				row["unknown"] = true
				return marshalIncidentSourceObject(t, row)
			},
			invariant: "incident.source_identity_admitted",
		},
		{
			name: "duplicate identity outranks shape",
			payload: func() []byte {
				payload := string(validIncidentSourcePayload(t))
				return []byte(strings.Replace(payload, `"id":"`+incidentSourceTestID.String()+`"`, `"id":"`+incidentSourceTestID.String()+`","id":"`+incidentSourceTestID.String()+`"`, 1))
			},
			invariant: "incident.source_identity_admitted",
		},
		{
			name: "shape outranks lifecycle and attribution",
			payload: func() []byte {
				row := validIncidentSourceObject()
				row["unknown"] = true
				row["status"] = "invalid"
				row["created_by_user_id"] = uuid.NewString()
				return marshalIncidentSourceObject(t, row)
			},
			invariant: "incident.exact_shape",
		},
		{
			name: "lifecycle outranks attribution",
			payload: func() []byte {
				row := validIncidentSourceObject()
				row["status"] = "invalid"
				row["created_by_user_id"] = uuid.NewString()
				return marshalIncidentSourceObject(t, row)
			},
			invariant: "incident.identity_key_lifecycle",
		},
		{
			name: "attribution after admitted lifecycle",
			payload: func() []byte {
				row := validIncidentSourceObject()
				row["created_by_user_id"] = uuid.NewString()
				return marshalIncidentSourceObject(t, row)
			},
			invariant: "incident.attribution_version",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeIncidentSourceRow(contract, test.payload(), importContext)
			requireIncidentSourceInvariant(t, err, test.invariant)
		})
	}
}

func TestIncidentSourcePrepareRejectsEveryDocumentAndScalarDefect_Unit(t *testing.T) {
	contract := mustIncidentSourceContract(t)
	importContext := newIncidentSourceImportContextForTest(t, "codec-defects")
	otherID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

	tests := []struct {
		name      string
		payload   func() []byte
		invariant string
	}{
		{name: "malformed", payload: func() []byte { return []byte(`{"id":`) }, invariant: "incident.exact_shape"},
		{name: "multiple values", payload: func() []byte { return append(validIncidentSourcePayload(t), []byte(`{}`)...) }, invariant: "incident.exact_shape"},
		{name: "non object", payload: func() []byte { return []byte(`[]`) }, invariant: "incident.exact_shape"},
		{name: "wrong identity", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["id"] = otherID.String() }), invariant: "incident.source_identity_admitted"},
		{name: "uppercase identity", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["id"] = strings.ToUpper(incidentSourceTestID.String()) }), invariant: "incident.source_identity_admitted"},
		{name: "nil identity", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["id"] = uuid.Nil.String() }), invariant: "incident.source_identity_admitted"},
		{name: "missing member", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { delete(row, "description") }), invariant: "incident.exact_shape"},
		{name: "unknown member", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["incident_id"] = incidentSourceTestID.String() }), invariant: "incident.exact_shape"},
		{name: "wrong scalar type", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["title"] = 7 }), invariant: "incident.exact_shape"},
		{name: "null required scalar", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["status"] = nil }), invariant: "incident.exact_shape"},
		{name: "wrong nullable type", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["severity"] = false }), invariant: "incident.exact_shape"},
		{name: "wrong version type", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["incident_version"] = "7" }), invariant: "incident.exact_shape"},
		{name: "key mismatch", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["incident_key_canonical"] = "IR-OTHER" }), invariant: "incident.identity_key_lifecycle"},
		{name: "key bytes", payload: mutateIncidentSourcePayload(t, func(row map[string]any) {
			value := strings.Repeat("x", 129)
			row["incident_key"], row["incident_key_canonical"] = value, value
		}), invariant: "incident.identity_key_lifecycle"},
		{name: "non NFC title", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["title"] = "e\u0301" }), invariant: "incident.identity_key_lifecycle"},
		{name: "title control", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["title"] = "line\nbreak" }), invariant: "incident.identity_key_lifecycle"},
		{name: "empty description", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["description"] = "" }), invariant: "incident.identity_key_lifecycle"},
		{name: "description carriage return", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["description"] = "line\rbreak" }), invariant: "incident.identity_key_lifecycle"},
		{name: "empty metadata", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["severity"] = "" }), invariant: "incident.identity_key_lifecycle"},
		{name: "unknown TLP", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["tlp"] = "TLP:BLUE" }), invariant: "incident.identity_key_lifecycle"},
		{name: "unknown status", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["status"] = "resolved" }), invariant: "incident.identity_key_lifecycle"},
		{name: "active with closed time", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["status"] = "active" }), invariant: "incident.identity_key_lifecycle"},
		{name: "closed without closed time", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["closed_at"] = nil }), invariant: "incident.identity_key_lifecycle"},
		{name: "unadmitted created actor", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["created_by_user_id"] = otherID.String() }), invariant: "incident.attribution_version"},
		{name: "uppercase updated actor", payload: mutateIncidentSourcePayload(t, func(row map[string]any) {
			row["updated_by_user_id"] = strings.ToUpper(incidentSourceUpdatedActor.String())
		}), invariant: "incident.attribution_version"},
		{name: "Z timestamp", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["created_at"] = "2026-08-28T12:00:00Z" }), invariant: "incident.attribution_version"},
		{name: "seven fractional digits", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["updated_at"] = "2026-08-28T12:00:02.1234567+00:00" }), invariant: "incident.attribution_version"},
		{name: "created after updated", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["created_at"] = "2026-08-28T12:00:03+00:00" }), invariant: "incident.attribution_version"},
		{name: "closed before created", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["closed_at"] = "2026-08-28T11:59:59+00:00" }), invariant: "incident.attribution_version"},
		{name: "zero version", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["incident_version"] = 0 }), invariant: "incident.attribution_version"},
		{name: "noncanonical version", payload: mutateIncidentSourcePayload(t, func(row map[string]any) { row["incident_version"] = json.RawMessage(`1.0`) }), invariant: "incident.attribution_version"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeIncidentSourceRow(contract, test.payload(), importContext)
			requireIncidentSourceInvariant(t, err, test.invariant)
		})
	}
}

func TestIncidentSourcePrepareBindsPortOperationIncidentVersionAndContract_Unit(t *testing.T) {
	contract := mustIncidentSourceContract(t)
	importContext := newIncidentSourceImportContextForTest(t, "codec-binding")
	prepared, err := prepareIncidentBundleIncident(
		contract,
		validIncidentSourcePayload(t),
		importContext,
	)
	if err != nil {
		t.Fatalf("prepare valid incident source: %v", err)
	}
	if !prepared.matches(contract, importContext) {
		t.Fatal("valid prepared source does not match its creating context")
	}
	for name, mutate := range map[string]func(*preparedIncidentSource){
		"port":      func(value *preparedIncidentSource) { value.portKey = "module.other:incident" },
		"path":      func(value *preparedIncidentSource) { value.logicalPath = "data/other.json" },
		"schema":    func(value *preparedIncidentSource) { value.schemaID = "cartulary.other.v1" },
		"operation": func(value *preparedIncidentSource) { value.operationID = "other" },
		"incident":  func(value *preparedIncidentSource) { value.incidentID = uuid.New() },
		"version":   func(value *preparedIncidentSource) { value.bundleVersion = 4 },
		"contract":  func(value *preparedIncidentSource) { value.contractMajor++ },
	} {
		t.Run(name, func(t *testing.T) {
			changed := prepared
			mutate(&changed)
			if changed.matches(contract, importContext) {
				t.Fatal("mutated prepared source retained binding")
			}
		})
	}

	wrongVersion := importContext
	wrongVersion.bundleVersion = 4
	if _, err := prepareIncidentBundleIncident(contract, validIncidentSourcePayload(t), wrongVersion); !errors.Is(err, errIncidentSourceCatalog) {
		t.Fatalf("wrong source generation error = %v", err)
	}
}

func TestIncidentSourceApplyUsesExplicitColumnsAndPreservesBothAttributions_Unit(t *testing.T) {
	contract, prepared, importContext := mustPreparedIncidentSource(t, "codec-apply")
	recorder := &incidentSourceAttributionRecorder{}
	importContext.attributions = recorder
	db := &incidentSourceDatabaseFake{execTag: pgconn.NewCommandTag("INSERT 0 1")}
	if err := applyIncidentBundleIncidentTx(context.Background(), db, prepared, importContext, contract); err != nil {
		t.Fatalf("apply prepared source: %v", err)
	}
	if db.execCalls != 1 || len(db.execArgs) != 16 || strings.Contains(db.execSQL, "jsonb_populate_record") ||
		strings.Contains(db.execSQL, "SELECT *") {
		t.Fatalf("incident apply SQL is not explicit: calls=%d args=%d sql=%q", db.execCalls, len(db.execArgs), db.execSQL)
	}
	if db.execArgs[10] != incidentSourceImportActor || db.execArgs[13] != incidentSourceImportActor {
		t.Fatalf("stored actor remap = created %#v updated %#v", db.execArgs[10], db.execArgs[13])
	}
	wantAttributions := []incidentportability.ImportedAttribution{
		{SourceTable: "incidents", SourceRowID: incidentSourceTestID.String(), SourceColumn: "created_by_user_id", SourceActorID: incidentSourceCreatedActor.String()},
		{SourceTable: "incidents", SourceRowID: incidentSourceTestID.String(), SourceColumn: "updated_by_user_id", SourceActorID: incidentSourceUpdatedActor.String()},
	}
	if !slices.Equal(recorder.rows, wantAttributions) {
		t.Fatalf("source actor attributions = %#v, want %#v", recorder.rows, wantAttributions)
	}

	zeroRows := &incidentSourceDatabaseFake{execTag: pgconn.NewCommandTag("INSERT 0 0")}
	importContext.attributions = &incidentSourceAttributionRecorder{}
	err := applyIncidentBundleIncidentTx(context.Background(), zeroRows, prepared, importContext, contract)
	if path, ok := incidentportability.FixedImportFailurePath(err); !ok || path != contract.source.Path.LogicalPath {
		t.Fatalf("affected-row mismatch = %v, path %q, %v", err, path, ok)
	}
	if err := applyIncidentBundleIncidentTx(context.Background(), db, "wrong", importContext, contract); !errors.Is(err, errIncidentSourceCatalog) {
		t.Fatalf("wrong prepared type = %v", err)
	}
}

func TestIncidentSourceValidateComparesEveryStoredColumnAfterActorRemap_Unit(t *testing.T) {
	contract, prepared, importContext := mustPreparedIncidentSource(t, "codec-validate")
	expected := prepared.row
	expected.CreatedByUserID = importContext.actorUserID
	expected.UpdatedByUserID = importContext.actorUserID

	if err := validateIncidentBundleIncidentTx(
		context.Background(),
		validationIncidentSourceDatabase(1, expected),
		prepared,
		importContext,
		contract,
	); err != nil {
		t.Fatalf("validate exact stored row: %v", err)
	}
	if err := validateIncidentBundleIncidentTx(
		context.Background(),
		validationIncidentSourceDatabase(0, expected),
		prepared,
		importContext,
		contract,
	); err == nil {
		t.Fatal("missing stored row passed validation")
	} else {
		requireIncidentSourceInvariant(t, err, "incident.source_identity_admitted")
	}

	tests := []struct {
		name      string
		invariant string
		mutate    func(*incidentSourceRow)
	}{
		{name: "id", invariant: "incident.source_identity_admitted", mutate: func(row *incidentSourceRow) { row.ID = uuid.New() }},
		{name: "incident key", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.IncidentKey = "changed" }},
		{name: "canonical key", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.IncidentKeyCanonical = "changed" }},
		{name: "title", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.Title = "changed" }},
		{name: "description", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.Description = stringPointer("changed") }},
		{name: "status", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.Status = "active" }},
		{name: "severity", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.Severity = stringPointer("changed") }},
		{name: "TLP", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.TLP = stringPointer("TLP:RED") }},
		{name: "phase", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.CurrentPhase = stringPointer("changed") }},
		{name: "case ref", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { row.PrimaryExternalCaseRef = stringPointer("changed") }},
		{name: "created actor", invariant: "incident.attribution_version", mutate: func(row *incidentSourceRow) { row.CreatedByUserID = uuid.New() }},
		{name: "created time", invariant: "incident.attribution_version", mutate: func(row *incidentSourceRow) { row.CreatedAt = row.CreatedAt.Add(time.Microsecond) }},
		{name: "updated time", invariant: "incident.attribution_version", mutate: func(row *incidentSourceRow) { row.UpdatedAt = row.UpdatedAt.Add(time.Microsecond) }},
		{name: "updated actor", invariant: "incident.attribution_version", mutate: func(row *incidentSourceRow) { row.UpdatedByUserID = uuid.New() }},
		{name: "version", invariant: "incident.attribution_version", mutate: func(row *incidentSourceRow) { row.IncidentVersion++ }},
		{name: "closed time", invariant: "incident.identity_key_lifecycle", mutate: func(row *incidentSourceRow) { value := row.ClosedAt.Add(time.Microsecond); row.ClosedAt = &value }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stored := cloneIncidentSourceRow(expected)
			test.mutate(&stored)
			err := validateIncidentBundleIncidentTx(
				context.Background(),
				validationIncidentSourceDatabase(1, stored),
				prepared,
				importContext,
				contract,
			)
			requireIncidentSourceInvariant(t, err, test.invariant)
		})
	}
}

func TestIncidentSourceExportNamesFixedProjectionAndPreservesV3Bytes_Unit(t *testing.T) {
	const portableRow = `{"closed_at":null,"created_at":"2026-08-28T12:34:56.123456+00:00","created_by_user_id":"22222222-2222-4222-8222-222222222222","current_phase":"triage","description":"Portable description","id":"11111111-1111-4111-8111-111111111111","incident_key":"IR-PORTABLE-1","incident_key_canonical":"IR-PORTABLE-1","incident_version":7,"primary_external_case_ref":null,"severity":"high","status":"active","title":"Portable incident","tlp":"TLP:AMBER","updated_at":"2026-08-28T12:34:56.123456+00:00","updated_by_user_id":"33333333-3333-4333-8333-333333333333"}`
	query := &incidentSourceExportQueryFake{payload: []byte(portableRow), incidentKey: "IR-PORTABLE-1"}
	payload, key, err := exportIncidentBundleIncident(context.Background(), query, incidentSourceTestID)
	if err != nil {
		t.Fatalf("export incident source: %v", err)
	}
	if string(payload) != portableRow+"\n" || key != "IR-PORTABLE-1" {
		t.Fatalf("exported source = %q, key %q", payload, key)
	}
	if strings.Contains(query.query, "to_jsonb(") || strings.Contains(query.query, "SELECT *") {
		t.Fatalf("export query retained table-shaped projection: %s", query.query)
	}
	contract := mustIncidentSourceContract(t)
	for _, column := range contract.columns {
		if !strings.Contains(query.query, "'"+column+"'") {
			t.Fatalf("export query omits fixed column %q: %s", column, query.query)
		}
	}
}

func mustIncidentSourceContract(t testing.TB) incidentSourceContract {
	t.Helper()
	contract, err := newIncidentSourceContract()
	if err != nil {
		t.Fatalf("construct incident source contract: %v", err)
	}
	return contract
}

func newIncidentSourceImportContextForTest(t testing.TB, operationID string) incidentSourceImportContext {
	t.Helper()
	actors := map[string]struct{}{
		incidentSourceCreatedActor.String(): {},
		incidentSourceUpdatedActor.String(): {},
	}
	return incidentSourceImportContext{
		incidentID: incidentSourceTestID, actorUserID: incidentSourceImportActor,
		bundleVersion: 3, operationID: operationID,
		actorAdmitted: func(actorID string) bool {
			_, admitted := actors[actorID]
			return admitted
		},
	}
}

func validIncidentSourceObject() map[string]any {
	return map[string]any{
		"id":                        incidentSourceTestID.String(),
		"incident_key":              "IR-PORTABLE-1",
		"incident_key_canonical":    "IR-PORTABLE-1",
		"title":                     "Portable incident",
		"description":               "Portable description\nwith\ttabs",
		"status":                    "closed",
		"severity":                  "high",
		"tlp":                       "TLP:AMBER",
		"current_phase":             "triage",
		"primary_external_case_ref": "CASE-1",
		"created_by_user_id":        incidentSourceCreatedActor.String(),
		"created_at":                "2026-08-28T12:00:00.123456+00:00",
		"updated_at":                "2026-08-28T12:00:02.123456+00:00",
		"updated_by_user_id":        incidentSourceUpdatedActor.String(),
		"incident_version":          7,
		"closed_at":                 "2026-08-28T12:00:01.123456+00:00",
	}
}

func validIncidentSourcePayload(t testing.TB) []byte {
	t.Helper()
	return marshalIncidentSourceObject(t, validIncidentSourceObject())
}

func mutateIncidentSourcePayload(t testing.TB, mutate func(map[string]any)) func() []byte {
	t.Helper()
	return func() []byte {
		row := validIncidentSourceObject()
		mutate(row)
		return marshalIncidentSourceObject(t, row)
	}
}

func marshalIncidentSourceObject(t testing.TB, row map[string]any) []byte {
	t.Helper()
	payload, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal incident source row: %v", err)
	}
	return append(payload, '\n')
}

func requireIncidentSourceInvariant(t testing.TB, err error, invariantID string) {
	t.Helper()
	var failure *incidentSourceInvariantFailure
	if !errors.As(err, &failure) || failure.invariantID != invariantID {
		t.Fatalf("source failure = %v, want invariant %s", err, invariantID)
	}
}

func mustPreparedIncidentSource(
	t testing.TB,
	operationID string,
) (incidentSourceContract, preparedIncidentSource, incidentSourceImportContext) {
	t.Helper()
	contract := mustIncidentSourceContract(t)
	importContext := newIncidentSourceImportContextForTest(t, operationID)
	prepared, err := prepareIncidentBundleIncident(
		contract,
		validIncidentSourcePayload(t),
		importContext,
	)
	if err != nil {
		t.Fatalf("prepare incident source: %v", err)
	}
	return contract, prepared, importContext
}

type incidentSourceAttributionRecorder struct {
	rows []incidentportability.ImportedAttribution
}

func (recorder *incidentSourceAttributionRecorder) RecordImportedAttribution(
	table string,
	rowID string,
	column string,
	actorID string,
) error {
	recorder.rows = append(recorder.rows, incidentportability.ImportedAttribution{
		SourceTable: table, SourceRowID: rowID, SourceColumn: column, SourceActorID: actorID,
	})
	return nil
}

func (recorder *incidentSourceAttributionRecorder) ImportedAttributions() []incidentportability.ImportedAttribution {
	return slices.Clone(recorder.rows)
}

type incidentSourceDatabaseFake struct {
	execTag   pgconn.CommandTag
	execErr   error
	execCalls int
	execSQL   string
	execArgs  []any
	rows      []pgx.Row
	rowIndex  int
}

func (db *incidentSourceDatabaseFake) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	db.execCalls++
	db.execSQL = query
	db.execArgs = slices.Clone(args)
	return db.execTag, db.execErr
}

func (db *incidentSourceDatabaseFake) QueryRow(context.Context, string, ...any) pgx.Row {
	if db.rowIndex >= len(db.rows) {
		return incidentSourceScanRow(func([]any) error { return errors.New("unexpected incident source query") })
	}
	row := db.rows[db.rowIndex]
	db.rowIndex++
	return row
}

type incidentSourceScanRow func([]any) error

func (row incidentSourceScanRow) Scan(destinations ...any) error {
	return row(destinations)
}

func validationIncidentSourceDatabase(count int, stored incidentSourceRow) *incidentSourceDatabaseFake {
	return &incidentSourceDatabaseFake{rows: []pgx.Row{
		incidentSourceScanRow(func(destinations []any) error {
			if len(destinations) != 1 {
				return fmt.Errorf("count destinations = %d", len(destinations))
			}
			*destinations[0].(*int) = count
			return nil
		}),
		incidentSourceStoredRow(stored),
	}}
}

func incidentSourceStoredRow(stored incidentSourceRow) pgx.Row {
	return incidentSourceScanRow(func(destinations []any) error {
		if len(destinations) != 16 {
			return fmt.Errorf("stored destinations = %d", len(destinations))
		}
		*destinations[0].(*pgtype.UUID) = pgtype.UUID{Bytes: [16]byte(stored.ID), Valid: true}
		*destinations[1].(*string) = stored.IncidentKey
		*destinations[2].(*string) = stored.IncidentKeyCanonical
		*destinations[3].(*string) = stored.Title
		*destinations[4].(*pgtype.Text) = pgText(stored.Description)
		*destinations[5].(*string) = stored.Status
		*destinations[6].(*pgtype.Text) = pgText(stored.Severity)
		*destinations[7].(*pgtype.Text) = pgText(stored.TLP)
		*destinations[8].(*pgtype.Text) = pgText(stored.CurrentPhase)
		*destinations[9].(*pgtype.Text) = pgText(stored.PrimaryExternalCaseRef)
		*destinations[10].(*pgtype.UUID) = pgtype.UUID{Bytes: [16]byte(stored.CreatedByUserID), Valid: true}
		*destinations[11].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: stored.CreatedAt, Valid: true}
		*destinations[12].(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: stored.UpdatedAt, Valid: true}
		*destinations[13].(*pgtype.UUID) = pgtype.UUID{Bytes: [16]byte(stored.UpdatedByUserID), Valid: true}
		*destinations[14].(*int64) = stored.IncidentVersion
		*destinations[15].(*pgtype.Timestamptz) = pgTime(stored.ClosedAt)
		return nil
	})
}

func pgText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func pgTime(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}

func cloneIncidentSourceRow(source incidentSourceRow) incidentSourceRow {
	result := source
	result.Description = cloneStringPointer(source.Description)
	result.Severity = cloneStringPointer(source.Severity)
	result.TLP = cloneStringPointer(source.TLP)
	result.CurrentPhase = cloneStringPointer(source.CurrentPhase)
	result.PrimaryExternalCaseRef = cloneStringPointer(source.PrimaryExternalCaseRef)
	if source.ClosedAt != nil {
		value := *source.ClosedAt
		result.ClosedAt = &value
	}
	return result
}

func stringPointer(value string) *string {
	return &value
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	return stringPointer(*value)
}

type incidentSourceExportQueryFake struct {
	query       string
	payload     []byte
	incidentKey string
}

func (*incidentSourceExportQueryFake) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected incident source Query call")
}

func (query *incidentSourceExportQueryFake) QueryRow(_ context.Context, statement string, _ ...any) pgx.Row {
	query.query = statement
	return incidentSourceScanRow(func(destinations []any) error {
		if len(destinations) != 2 {
			return fmt.Errorf("export destinations = %d", len(destinations))
		}
		*destinations[0].(*[]byte) = slices.Clone(query.payload)
		*destinations[1].(*string) = query.incidentKey
		return nil
	})
}
