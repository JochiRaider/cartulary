package records

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRecordsIncidentBundleSourcePortRejectsDuplicateIdentity_Unit(t *testing.T) {
	recordID := uuid.New()
	incidentID := uuid.New()
	actorID := uuid.New()
	now := time.Date(2026, time.July, 29, 20, 0, 0, 1, time.UTC)
	row := validRecordsPortableMap(recordID, incidentID, actorID, now)
	payload := append(encodeRecordsPortableRow(t, row), encodeRecordsPortableRow(t, row)...)
	port := newRecordsTestPort(t, true)
	_, err := port.PrepareImport(
		context.Background(),
		sourceport.MapBundle{recordsBundlePath: payload},
		recordsTestImportContext(t, incidentID, actorID, "records-duplicate", nil),
	)
	requireRecordsFailure(t, err, "records.envelope_legal")
}

func TestRecordsSourcePortRejectsMissingSubtypeCatalog_Unit(t *testing.T) {
	validator, ok := NewIncidentBundleSourcePort(nil).(sourceport.ContractValidator)
	if !ok {
		t.Fatal("records source port does not expose contract validation")
	}
	if err := validator.ValidateSourcePortContract(); !errors.Is(err, sourceport.ErrInvalidCatalog) {
		t.Fatalf("missing subtype catalog error = %v; want ErrInvalidCatalog", err)
	}
}

func TestRecordsIncidentBundleSourcePortAppliesAndRollsBack_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-incident-bundle-source-port")
	actorID, incidentID := seedEnvelopeOwnerContext(t, db, "records-source-port")
	recordID := uuid.New()
	now := time.Date(2026, time.July, 29, 20, 1, 0, 1, time.UTC)
	payload := encodeRecordsPortableRow(
		t,
		validRecordsPortableMap(recordID, incidentID, actorID, now),
	)
	port := newRecordsTestPort(t, true)
	attributions := &recordsTestAttributions{}
	importContext := recordsTestImportContext(
		t, incidentID, actorID, "records-apply-rollback", attributions,
	)
	prepared, err := port.PrepareImport(
		ctx,
		sourceport.MapBundle{recordsBundlePath: payload},
		importContext,
	)
	if err != nil {
		t.Fatalf("prepare records import: %v", err)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin records apply transaction: %v", err)
	}
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("apply records import: %v", err)
	}
	if err := port.ValidateImportTx(ctx, tx, prepared, importContext); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("validate records import: %v", err)
	}
	var insideCount int
	if err := tx.QueryRow(
		ctx,
		`SELECT count(*) FROM records WHERE record_id = $1`,
		recordID,
	).Scan(&insideCount); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("query applied record: %v", err)
	}
	if insideCount != 1 {
		_ = tx.Rollback(ctx)
		t.Fatalf("applied record count = %d; want 1", insideCount)
	}
	if got := len(attributions.rows); got != 2 {
		_ = tx.Rollback(ctx)
		t.Fatalf("recorded attributions = %d; want 2", got)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("roll back records import: %v", err)
	}
	var outsideCount int
	if err := db.QueryRow(
		ctx,
		`SELECT count(*) FROM records WHERE record_id = $1`,
		recordID,
	).Scan(&outsideCount); err != nil {
		t.Fatalf("query rolled-back record: %v", err)
	}
	if outsideCount != 0 {
		t.Fatalf("rolled-back record count = %d; want 0", outsideCount)
	}
}

func TestRecordsPortableEnvelopeInvariantContract_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-portable-envelope-invariants")
	actorID, incidentID := seedEnvelopeOwnerContext(t, db, "portable-envelope")
	port := newRecordsTestPort(t, true)
	for _, bundleVersion := range []int{1, 2} {
		t.Run(fmt.Sprintf("bundle_version_%d", bundleVersion), func(t *testing.T) {
			recordID := uuid.New()
			now := time.Date(2026, time.July, 29, 19, bundleVersion, 0, bundleVersion, time.UTC)
			payload := encodeRecordsPortableRow(
				t,
				validRecordsPortableMap(recordID, incidentID, actorID, now),
			)
			attributions := &recordsTestAttributions{}
			importContext := recordsTestImportContext(
				t,
				incidentID,
				actorID,
				fmt.Sprintf("records-portable-envelope-v%d", bundleVersion),
				attributions,
			)
			importContext.BundleVersion = bundleVersion
			prepared, err := port.PrepareImport(
				ctx,
				sourceport.MapBundle{recordsBundlePath: payload},
				importContext,
			)
			if err != nil {
				t.Fatalf("prepare valid v%d portable envelope: %v", bundleVersion, err)
			}
			tx, err := db.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin valid v%d import transaction: %v", bundleVersion, err)
			}
			if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("apply valid v%d portable envelope: %v", bundleVersion, err)
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, importContext); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("validate valid v%d portable envelope: %v", bundleVersion, err)
			}
			var count int
			if err := tx.QueryRow(ctx, `
SELECT count(*)
  FROM records
 WHERE record_id = $1
   AND incident_id = $2
   AND record_type = 'timeline_event'
   AND row_version = 1
   AND deleted_at IS NULL
   AND deleted_by_user_id IS NULL
`, recordID, incidentID).Scan(&count); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("query valid v%d portable envelope: %v", bundleVersion, err)
			}
			if count != 1 {
				_ = tx.Rollback(ctx)
				t.Fatalf("valid v%d portable envelope count = %d; want 1", bundleVersion, count)
			}
			if err := tx.Rollback(ctx); err != nil {
				t.Fatalf("roll back valid v%d portable envelope: %v", bundleVersion, err)
			}
			if err := db.QueryRow(
				ctx,
				`SELECT count(*) FROM records WHERE record_id = $1`,
				recordID,
			).Scan(&count); err != nil {
				t.Fatalf("query v%d portable-envelope rollback: %v", bundleVersion, err)
			}
			if count != 0 {
				t.Fatalf("rolled-back v%d portable envelope count = %d; want 0", bundleVersion, count)
			}
		})
	}
}

func TestRecordsPortableEnvelopeRejectsEveryInvalidCategory_Unit(t *testing.T) {
	incidentID := uuid.New()
	actorID := uuid.New()
	recordID := uuid.New()
	now := time.Date(2026, time.July, 29, 20, 2, 0, 123, time.UTC)
	base := validRecordsPortableMap(recordID, incidentID, actorID, now)
	port := newRecordsTestPort(t, true)
	baseContext := recordsTestImportContext(
		t, incidentID, actorID, "records-invalid-category", nil,
	)

	for member := range base {
		t.Run("missing_"+member, func(t *testing.T) {
			row := cloneRecordsPortableMap(base)
			delete(row, member)
			_, err := port.PrepareImport(
				context.Background(),
				sourceport.MapBundle{recordsBundlePath: encodeRecordsPortableRow(t, row)},
				baseContext,
			)
			requireRecordsFailure(t, err, "records.envelope_legal")
		})
	}

	cases := map[string]struct {
		payload     []byte
		context     sourceport.ImportContext
		invariantID string
	}{
		"unknown_member": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["extra"] = true
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"wrong_member_type": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["record_id"] = true
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"noncanonical_uuid": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["record_id"] = strings.ToUpper(recordID.String())
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"incident_mismatch": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["incident_id"] = uuid.NewString()
			}),
			context: baseContext, invariantID: "records.incident_scope",
		},
		"closed_type": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["record_type"] = "other"
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"actor_not_cataloged": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["created_by_user_id"] = uuid.NewString()
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"integer_zero": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["row_version"] = 0
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"integer_fraction": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["row_version"] = json.Number("1.0")
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"timestamp_non_utc": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["created_at"] = "2026-07-29T16:02:00-04:00"
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"timestamp_order": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["updated_at"] = now.Add(-time.Second).Format(time.RFC3339Nano)
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"delete_tuple": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["deleted_by_user_id"] = actorID.String()
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"delete_range": {
			payload: mutateRecordsPayload(t, base, func(row map[string]any) {
				row["deleted_at"] = now.Add(time.Second).Format(time.RFC3339Nano)
				row["deleted_by_user_id"] = actorID.String()
			}),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"blank_line": {
			payload: append(encodeRecordsPortableRow(t, base), '\n'),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"multiple_values": {
			payload: []byte("{} {}\n"),
			context: baseContext, invariantID: "records.envelope_legal",
		},
		"duplicate_member": {
			payload: []byte(fmt.Sprintf(
				`{"record_id":%q,"record_id":%q}`+"\n",
				recordID.String(),
				recordID.String(),
			)),
			context: baseContext, invariantID: "records.envelope_legal",
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			testCase.context.OperationID = "records-invalid-" + name
			_, err := port.PrepareImport(
				context.Background(),
				sourceport.MapBundle{recordsBundlePath: testCase.payload},
				testCase.context,
			)
			requireRecordsFailure(t, err, testCase.invariantID)
		})
	}
}

func TestRecordsPortableEnvelopeRejectsIncompleteSubtypeBindings_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-subtype-incomplete")
	actorID, incidentID := seedEnvelopeOwnerContext(t, db, "records-subtype")
	recordID := uuid.New()
	now := time.Date(2026, time.July, 29, 20, 3, 0, 1, time.UTC)
	port := newRecordsTestPort(t, false)
	attributions := &recordsTestAttributions{}
	importContext := recordsTestImportContext(
		t, incidentID, actorID, "records-subtype-incomplete", attributions,
	)
	prepared, err := port.PrepareImport(
		ctx,
		sourceport.MapBundle{
			recordsBundlePath: encodeRecordsPortableRow(
				t,
				validRecordsPortableMap(recordID, incidentID, actorID, now),
			),
		},
		importContext,
	)
	if err != nil {
		t.Fatalf("prepare subtype-incomplete import: %v", err)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin subtype-incomplete transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("apply subtype-incomplete import: %v", err)
	}
	requireRecordsFailure(
		t,
		port.ValidateImportTx(ctx, tx, prepared, importContext),
		"records.subtype_complete",
	)
}

func TestRecordsPortableExportPreservesSourceActorsAndExactShape_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-portable-export")
	runtimeActorID, incidentID := seedEnvelopeOwnerContext(t, db, "records-export")
	recordID := uuid.New()
	sourceActorID := uuid.New()
	createdAt := time.Date(2026, time.July, 29, 20, 4, 0, 123000, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	if _, err := db.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type,
    created_by_user_id, created_at,
    updated_by_user_id, updated_at,
    row_version
)
VALUES ($1, $2, 'timeline_event', $3, $4, $3, $5, 2)
`, recordID, incidentID, runtimeActorID, createdAt, updatedAt); err != nil {
		t.Fatalf("seed exported envelope: %v", err)
	}
	resolver := recordsTestPortableResolver{
		"created_by_user_id": {recordID.String(): sourceActorID.String()},
		"updated_by_user_id": {recordID.String(): sourceActorID.String()},
	}
	files, err := newRecordsTestPort(t, true).Export(
		ctx,
		sourceport.ExportContext{
			Query:                db,
			IncidentID:           incidentID,
			PortableAttributions: resolver,
		},
	)
	if err != nil {
		t.Fatalf("export records: %v", err)
	}
	if len(files) != 1 || files[0].Path != recordsBundlePath {
		t.Fatalf("exported files = %#v", files)
	}
	rows, err := incidentportability.DecodeStrictNDJSONObjects(
		files[0].Payload,
		recordsBundlePath,
	)
	if err != nil {
		t.Fatalf("decode exported records: %v", err)
	}
	if len(rows) != 1 || len(rows[0]) != 10 {
		t.Fatalf("exported record rows = %#v", rows)
	}
	if rows[0]["created_by_user_id"] != sourceActorID.String() ||
		rows[0]["updated_by_user_id"] != sourceActorID.String() ||
		rows[0]["created_at"] != createdAt.Format(time.RFC3339Nano) ||
		rows[0]["updated_at"] != updatedAt.Format(time.RFC3339Nano) {
		t.Fatalf("portable attribution or timestamps changed: %#v", rows[0])
	}
}

type recordsTestSubtypeSource struct {
	types  []subtypepresence.RecordType
	mirror bool
}

func (s recordsTestSubtypeSource) SupportedRecordTypes() []subtypepresence.RecordType {
	return append([]subtypepresence.RecordType(nil), s.types...)
}

func (s recordsTestSubtypeSource) ListSubtypeBindingsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]subtypepresence.Binding, error) {
	if !s.mirror {
		return nil, nil
	}
	typeNames := make([]string, 0, len(s.types))
	for _, recordType := range s.types {
		typeNames = append(typeNames, string(recordType))
	}
	rows, err := tx.Query(ctx, `
SELECT record_id, incident_id, record_type
  FROM records
 WHERE incident_id = $1
   AND record_type = ANY($2::text[])
 ORDER BY record_id
`, incidentID, typeNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bindings []subtypepresence.Binding
	for rows.Next() {
		var binding subtypepresence.Binding
		if err := rows.Scan(
			&binding.RecordID,
			&binding.IncidentID,
			&binding.RecordType,
		); err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, rows.Err()
}

func newRecordsTestPort(t testing.TB, mirror bool) sourceport.Port {
	t.Helper()
	catalog, err := subtypepresence.NewCatalog([]subtypepresence.Contribution{
		{FamilyID: "timeline", Source: recordsTestSubtypeSource{types: []subtypepresence.RecordType{subtypepresence.RecordTypeTimelineEvent}, mirror: mirror}},
		{FamilyID: "entities", Source: recordsTestSubtypeSource{types: []subtypepresence.RecordType{subtypepresence.RecordTypeHost, subtypepresence.RecordTypeIdentity}, mirror: mirror}},
		{FamilyID: "parties", Source: recordsTestSubtypeSource{types: []subtypepresence.RecordType{subtypepresence.RecordTypeParty}, mirror: mirror}},
		{FamilyID: "indicators", Source: recordsTestSubtypeSource{types: []subtypepresence.RecordType{subtypepresence.RecordTypeIndicator}, mirror: mirror}},
		{FamilyID: "artifacts", Source: recordsTestSubtypeSource{types: []subtypepresence.RecordType{subtypepresence.RecordTypeArtifact}, mirror: mirror}},
		{FamilyID: "tasks_decisions", Source: recordsTestSubtypeSource{types: []subtypepresence.RecordType{subtypepresence.RecordTypeTaskRequest, subtypepresence.RecordTypeDecision}, mirror: mirror}},
		{FamilyID: "evidence", Source: recordsTestSubtypeSource{types: []subtypepresence.RecordType{subtypepresence.RecordTypeEvidence}, mirror: mirror}},
		{FamilyID: "assessments", Source: recordsTestSubtypeSource{types: []subtypepresence.RecordType{subtypepresence.RecordTypeAssessment}, mirror: mirror}},
	})
	if err != nil {
		t.Fatalf("new records test subtype catalog: %v", err)
	}
	return NewIncidentBundleSourcePort(catalog)
}

type recordsTestAttributions struct {
	rows []incidentportability.ImportedAttribution
}

func (r *recordsTestAttributions) RecordImportedAttribution(
	table string,
	sourceRowID string,
	column string,
	sourceActorID string,
) error {
	r.rows = append(r.rows, incidentportability.ImportedAttribution{
		SourceTable:   table,
		SourceRowID:   sourceRowID,
		SourceColumn:  column,
		SourceActorID: sourceActorID,
	})
	return nil
}

func (r *recordsTestAttributions) ImportedAttributions() []incidentportability.ImportedAttribution {
	return append([]incidentportability.ImportedAttribution(nil), r.rows...)
}

type recordsTestPortableResolver map[string]map[string]string

func (r recordsTestPortableResolver) ResolvePortableSourceActors(
	_ context.Context,
	_ incidentportability.Queryer,
	_ uuid.UUID,
	_ string,
	sourceColumn string,
	_ []string,
) (map[string]string, error) {
	resolved := map[string]string{}
	for rowID, actorID := range r[sourceColumn] {
		resolved[rowID] = actorID
	}
	return resolved, nil
}

func recordsTestImportContext(
	t testing.TB,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	operationID string,
	attributions incidentportability.AttributionRecorder,
) sourceport.ImportContext {
	t.Helper()
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{
		SourceActorID: actorID.String(),
		DisplayName:   "Records source actor",
	}})
	if err != nil {
		t.Fatalf("new records actor catalog: %v", err)
	}
	return sourceport.ImportContext{
		IncidentID:    incidentID,
		ActorUserID:   actorID,
		BundleVersion: 2,
		OperationID:   operationID,
		Attributions:  attributions,
		Actors:        actors,
	}
}

func validRecordsPortableMap(
	recordID uuid.UUID,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	now time.Time,
) map[string]any {
	return map[string]any{
		"record_id":          recordID.String(),
		"incident_id":        incidentID.String(),
		"record_type":        "timeline_event",
		"created_at":         now.Format(time.RFC3339Nano),
		"created_by_user_id": actorID.String(),
		"updated_at":         now.Format(time.RFC3339Nano),
		"updated_by_user_id": actorID.String(),
		"row_version":        1,
		"deleted_at":         nil,
		"deleted_by_user_id": nil,
	}
}

func mutateRecordsPayload(
	t testing.TB,
	base map[string]any,
	mutate func(map[string]any),
) []byte {
	t.Helper()
	row := cloneRecordsPortableMap(base)
	mutate(row)
	return encodeRecordsPortableRow(t, row)
}

func cloneRecordsPortableMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func encodeRecordsPortableRow(t testing.TB, row map[string]any) []byte {
	t.Helper()
	payload, err := incidentportability.CanonicalJSONString(row)
	if err != nil {
		t.Fatalf("encode records portable row: %v", err)
	}
	return payload
}

func requireRecordsFailure(t testing.TB, err error, invariantID string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) ||
		failure.FamilyID != "records" ||
		failure.InvariantID != invariantID {
		t.Fatalf("records failure = %#v, %v; want %s", failure, err, invariantID)
	}
	if strings.Contains(err.Error(), "record_id") ||
		strings.Contains(err.Error(), "SELECT") ||
		strings.Contains(err.Error(), "records.ndjson") {
		t.Fatalf("records failure exposed internal or row detail: %v", err)
	}
}

func TestRecordsPortableEnvelopeRejectsTrailingContent_Unit(t *testing.T) {
	incidentID := uuid.New()
	actorID := uuid.New()
	payload := encodeRecordsPortableRow(
		t,
		validRecordsPortableMap(uuid.New(), incidentID, actorID, time.Now().UTC()),
	)
	payload = bytes.TrimSuffix(payload, []byte{'\n'})
	payload = append(payload, []byte(" true\n")...)
	_, err := newRecordsTestPort(t, true).PrepareImport(
		context.Background(),
		sourceport.MapBundle{recordsBundlePath: payload},
		recordsTestImportContext(t, incidentID, actorID, "records-trailing", nil),
	)
	requireRecordsFailure(t, err, "records.envelope_legal")
}
