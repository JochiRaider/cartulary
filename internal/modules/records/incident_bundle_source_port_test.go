package records

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestRecordsIncidentBundleSourcePortRejectsDuplicateIdentity_Unit(t *testing.T) {
	recordID := uuid.New()
	incidentID := uuid.New()
	actorID := uuid.New()
	now := time.Date(2026, time.July, 29, 20, 0, 0, 1000, time.UTC)
	row := validRecordsPortableMap(recordID, incidentID, actorID, now)
	payload := append(encodeRecordsPortableRow(t, row), encodeRecordsPortableRow(t, row)...)
	port := newRecordsTestPort(t, true)
	_, err := port.PrepareImport(
		context.Background(),
		portabilitysupport.MapBundle{recordsBundlePath: payload},
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
	now := time.Date(2026, time.July, 29, 20, 1, 0, 1000, time.UTC)
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
		portabilitysupport.MapBundle{recordsBundlePath: payload},
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
	t.Run("deleted row clocks round trip exactly", func(t *testing.T) {
		recordID := uuid.New()
		row := validRecordsPortableMap(recordID, incidentID, actorID, time.Date(2026, time.July, 29, 20, 0, 0, 1000, time.UTC))
		row["updated_at"] = "2026-07-29T20:00:00.999999Z"
		row["deleted_at"] = "2026-07-29T20:00:00.123456Z"
		row["deleted_by_user_id"] = actorID.String()
		input := recordsTestImportContext(t, incidentID, actorID, "records-clock-preservation", &recordsTestAttributions{})
		prepared, err := port.PrepareImport(ctx, portabilitysupport.MapBundle{recordsBundlePath: encodeRecordsPortableRow(t, row)}, input)
		if err != nil {
			t.Fatal(err)
		}
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		var created, updated, deleted time.Time
		if err := tx.QueryRow(ctx, `SELECT created_at,updated_at,deleted_at FROM records WHERE record_id=$1`, recordID).Scan(&created, &updated, &deleted); err != nil {
			t.Fatal(err)
		}
		if created.UTC().Format(time.RFC3339Nano) != "2026-07-29T20:00:00.000001Z" ||
			updated.UTC().Format(time.RFC3339Nano) != "2026-07-29T20:00:00.999999Z" ||
			deleted.UTC().Format(time.RFC3339Nano) != "2026-07-29T20:00:00.123456Z" {
			t.Fatalf("stored row clocks changed: %v %v %v", created, updated, deleted)
		}
		files, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: tx, IncidentID: incidentID})
		if err != nil || len(files) != 1 || files[0].Path != recordsBundlePath {
			t.Fatalf("clock re-export: %v %v", files, err)
		}
		rows, err := incidentportability.DecodeNDJSON(files[0].Payload)
		if err != nil || len(rows) != 1 {
			t.Fatalf("clock re-export rows: %v %v", rows, err)
		}
		if rows[0]["created_at"] != "2026-07-29T20:00:00.000001Z" ||
			rows[0]["updated_at"] != "2026-07-29T20:00:00.999999Z" ||
			rows[0]["deleted_at"] != "2026-07-29T20:00:00.123456Z" {
			t.Fatalf("re-exported row clocks changed: %#v", rows[0])
		}
	})
	for _, bundleVersion := range []int{8} {
		t.Run(fmt.Sprintf("bundle_version_%d", bundleVersion), func(t *testing.T) {
			recordID := uuid.New()
			now := time.Date(2026, time.July, 29, 19, bundleVersion, 0, bundleVersion*1000, time.UTC)
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
				portabilitysupport.MapBundle{recordsBundlePath: payload},
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
	recordID := uuid.MustParse("abcdef01-2345-4678-9abc-def012345678")
	now := time.Date(2026, time.July, 29, 20, 2, 0, 123000, time.UTC)
	base := validRecordsPortableMap(recordID, incidentID, actorID, now)
	port := newRecordsTestPort(t, true)
	baseContext := recordsTestImportContext(
		t, incidentID, actorID, "records-invalid-category", nil,
	)
	t.Run("historical envelope validation", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			mutate func(map[string]any)
			valid  bool
		}{
			{"complete", func(map[string]any) {}, true},
			{"SQL clock spelling", func(row map[string]any) { row["created_at"] = "2026-07-29T20:02:00.000123+00:00" }, true},
			{"decoded number", func(row map[string]any) { row["row_version"] = float64(1) }, true},
			{"submicrosecond clock", func(row map[string]any) { row["created_at"] = "2026-07-29T20:02:00.0000001Z" }, false},
			{"record binding", func(row map[string]any) { row["record_id"] = uuid.NewString() }, false},
			{"incident binding", func(row map[string]any) { row["incident_id"] = uuid.NewString() }, false},
			{"type binding", func(row map[string]any) { row["record_type"] = "host" }, false},
			{"missing nullable", func(row map[string]any) { delete(row, "deleted_at") }, false},
			{"unknown member", func(row map[string]any) { row["derived"] = true }, false},
			{"nil actor", func(row map[string]any) { row["created_by_user_id"] = uuid.Nil.String() }, false},
			{"malformed actor", func(row map[string]any) { row["updated_by_user_id"] = "not-a-uuid" }, false},
			{"invalid version", func(row map[string]any) { row["row_version"] = "1" }, false},
			{"reversed clocks", func(row map[string]any) { row["updated_at"] = "2000-01-01T00:00:00Z" }, false},
			{"partial deletion", func(row map[string]any) { row["deleted_at"] = row["updated_at"] }, false},
			{"complete deletion", func(row map[string]any) {
				row["deleted_at"] = row["updated_at"]
				row["deleted_by_user_id"] = actorID.String()
			}, true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				row := cloneRecordsPortableMap(base)
				tc.mutate(row)
				before := encodeRecordsPortableRow(t, row)
				err := ValidateSnapshotEnvelope(incidentID, recordID, "timeline_event", row)
				if (err == nil) != tc.valid {
					t.Fatalf("snapshot validation = %v, valid %t", err, tc.valid)
				}
				if !bytes.Equal(before, encodeRecordsPortableRow(t, row)) {
					t.Fatal("snapshot validation changed history")
				}
			})
		}
	})

	t.Run("historical and portable exact clock forms", func(t *testing.T) {
		for _, field := range []string{"created_at", "updated_at", "deleted_at"} {
			for _, test := range []struct {
				value                string
				historical, portable bool
			}{
				{"0001-01-01T00:00:00Z", true, true},
				{"9999-12-31T23:59:59.999999Z", true, true},
				{"2026-07-29T20:02:00.000001Z", true, true},
				{"2026-07-29T20:02:00.000123+00:00", true, false},
				{"0000-01-01T00:00:00Z", false, false},
				{"10000-01-01T00:00:00Z", false, false},
				{"2026-07-29T21:02:00.000123+01:00", false, false},
				{"2026-07-29T19:02:00.000123-01:00", false, false},
				{"2026-07-29T20:02:00.000123-00:00", false, false},
				{"2026-07-29T20:02:00.100000Z", false, false},
				{"2026-07-29T20:02:00.0001230Z", false, false},
				{"2026-07-29T20:02:00.0000001Z", false, false},
				{"2026-07-29T20:02:00,000123Z", false, false},
				{"2026-07-29T20:02:00Z ", false, false},
			} {
				t.Run(field+"/"+test.value, func(t *testing.T) {
					row := cloneRecordsPortableMap(base)
					row["created_at"], row["updated_at"] = "0001-01-01T00:00:00Z", "9999-12-31T23:59:59.999999Z"
					row[field] = test.value
					if field == "deleted_at" {
						row["deleted_by_user_id"] = actorID.String()
					}
					before := encodeRecordsPortableRow(t, row)
					err := ValidateSnapshotEnvelope(incidentID, recordID, "timeline_event", row)
					if test.historical {
						if err != nil {
							t.Fatalf("canonical history rejected: %v", err)
						}
					} else if err == nil {
						t.Fatal("noncanonical historical timestamp accepted")
					}
					if !bytes.Equal(before, encodeRecordsPortableRow(t, row)) {
						t.Fatal("historical clock rewritten")
					}
					bundle := portabilitysupport.MapBundle{recordsBundlePath: bytes.Clone(before)}
					_, err = port.PrepareImport(t.Context(), bundle, baseContext)
					if test.portable {
						if err != nil {
							t.Fatalf("canonical portable clock rejected: %v", err)
						}
					} else {
						requireRecordsFailure(t, err, "records.envelope_legal")
					}
					if !bytes.Equal(before, bundle[recordsBundlePath]) {
						t.Fatal("portable clock rewritten")
					}
				})
			}
		}
	})
	t.Run("row clock precision", func(t *testing.T) {
		for _, field := range []string{"created_at", "updated_at", "deleted_at"} {
			for _, value := range []string{
				"2026-07-29T20:02:00Z", "2026-07-29T20:02:00.000001Z", "2026-07-29T20:02:00.999999Z",
				"2026-07-29T20:02:00.1234567Z", "2026-07-29T20:02:00.000000001Z",
			} {
				t.Run(field+"/"+value, func(t *testing.T) {
					row := cloneRecordsPortableMap(base)
					row["created_at"] = "2026-07-29T20:02:00Z"
					row["updated_at"] = "2026-07-29T20:02:01Z"
					row[field] = value
					if field == "deleted_at" {
						row["deleted_by_user_id"] = actorID.String()
					}
					_, err := port.PrepareImport(t.Context(), portabilitysupport.MapBundle{
						recordsBundlePath: encodeRecordsPortableRow(t, row),
					}, baseContext)
					if value == "2026-07-29T20:02:00.1234567Z" || value == "2026-07-29T20:02:00.000000001Z" {
						requireRecordsFailure(t, err, "records.envelope_legal")
					} else if err != nil {
						t.Fatalf("representable %s rejected: %v", field, err)
					}
				})
			}
		}
	})

	for member := range base {
		t.Run("missing_"+member, func(t *testing.T) {
			row := cloneRecordsPortableMap(base)
			delete(row, member)
			_, err := port.PrepareImport(
				context.Background(),
				portabilitysupport.MapBundle{recordsBundlePath: encodeRecordsPortableRow(t, row)},
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
				portabilitysupport.MapBundle{recordsBundlePath: testCase.payload},
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
	now := time.Date(2026, time.July, 29, 20, 3, 0, 1000, time.UTC)
	port := newRecordsTestPort(t, false)
	attributions := &recordsTestAttributions{}
	importContext := recordsTestImportContext(
		t, incidentID, actorID, "records-subtype-incomplete", attributions,
	)
	prepared, err := port.PrepareImport(
		ctx,
		portabilitysupport.MapBundle{
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
	sourceActorID := uuid.New()
	createdAt := time.Date(2026, time.July, 29, 20, 4, 0, 123000, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	resolver := recordsTestPortableResolver{"created_by_user_id": {}, "updated_by_user_id": {}}
	expected := make(map[string]map[string]any)
	const count = 257
	for i := 0; i < count; i++ {
		recordID := uuid.New()
		resolver["created_by_user_id"][recordID.String()] = sourceActorID.String()
		resolver["updated_by_user_id"][recordID.String()] = sourceActorID.String()
		expected[recordID.String()] = map[string]any{
			"record_id": recordID.String(), "incident_id": incidentID.String(),
			"record_type": "timeline_event", "row_version": json.Number("2"),
			"created_at": createdAt.Format(time.RFC3339Nano), "created_by_user_id": sourceActorID.String(),
			"updated_at": updatedAt.Format(time.RFC3339Nano), "updated_by_user_id": sourceActorID.String(),
			"deleted_at": nil, "deleted_by_user_id": nil,
		}

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
	}
	files, err := portabilitysupport.Export(newRecordsTestPort(t, true),
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
	if len(rows) != count {
		t.Fatalf("exported %d records, want %d across page boundaries", len(rows), count)
	}
	previous := ""
	for _, row := range rows {
		id := row["record_id"].(string)
		if id <= previous || expected[id] == nil {
			t.Fatalf("invalid or duplicate exported envelope: %#v", row)
		}
		previous = id
		if !reflect.DeepEqual(row, expected[id]) {
			t.Fatalf("exported envelope differs from seeded values: got=%#v want=%#v", row, expected[id])
		}
	}
}

type recordsTestSubtypeSource struct {
	types  []subtypepresence.RecordType
	mirror bool
}

func (s recordsTestSubtypeSource) SupportedRecordTypes() []subtypepresence.RecordType {
	return append([]subtypepresence.RecordType(nil), s.types...)
}

func (s recordsTestSubtypeSource) VisitSubtypeBindingsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	visit func(subtypepresence.Binding) error,
) error {
	if !s.mirror {
		return nil
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
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var binding subtypepresence.Binding
		if err := rows.Scan(
			&binding.RecordID,
			&binding.IncidentID,
			&binding.RecordType,
		); err != nil {
			return err
		}
		if err := visit(binding); err != nil {
			return err
		}
	}
	return rows.Err()
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
		BundleVersion: 8,
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
		failure.FamilyID() != "records" ||
		failure.InvariantID() != invariantID {
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
		validRecordsPortableMap(uuid.New(), incidentID, actorID, time.Date(2026, time.July, 29, 20, 5, 0, 123000, time.UTC)),
	)
	payload = bytes.TrimSuffix(payload, []byte{'\n'})
	payload = append(payload, []byte(" true\n")...)
	_, err := newRecordsTestPort(t, true).PrepareImport(
		context.Background(),
		portabilitysupport.MapBundle{recordsBundlePath: payload},
		recordsTestImportContext(t, incidentID, actorID, "records-trailing", nil),
	)
	requireRecordsFailure(t, err, "records.envelope_legal")
}

func (r *recordsTestAttributions) VisitImportedAttributions(table, rowID string, visit func(incidentportability.ImportedAttribution) error) error {
	for _, row := range r.ImportedAttributions() {
		if row.SourceTable == table && (rowID == "" || row.SourceRowID == rowID) {
			if err := visit(row); err != nil {
				return err
			}
		}
	}
	return nil
}
