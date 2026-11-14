package evidence_test

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

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const evidencePortablePath = "data/evidence_records.ndjson"

var evidencePortableActor = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

func evidencePortableRow(incident uuid.UUID, index int) map[string]any {
	return map[string]any{
		"record_id": fmt.Sprintf("10000000-0000-4000-8000-%012d", index), "incident_id": incident.String(),
		"title": "Evidence source", "lifecycle_state": "requested", "requested_at": "2026-01-01T00:00:00.123456+00:00", "received_at": nil,
		"storage_ref": "ticket://source", "blob_hash": nil, "collector_party_text": "Collector", "collector_party_id": nil,
		"source_party_text": "Source", "source_party_id": nil, "upload_state": "pending", "object_blob_id": nil, "owner_user_id": evidencePortableActor.String(),
		"created_at": "2026-01-01T00:00:00+00:00", "updated_at": "2026-01-02T00:00:00+00:00",
	}
}
func evidencePortableBundle(t testing.TB, rows ...map[string]any) portabilitysupport.MapBundle {
	t.Helper()
	var out bytes.Buffer
	for _, row := range rows {
		if err := json.NewEncoder(&out).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	return portabilitysupport.MapBundle{evidencePortablePath: out.Bytes(), "data/object_blobs.ndjson": nil, "data/evidence_custody_events.ndjson": nil}
}
func evidencePortableInput(t testing.TB, incident, local uuid.UUID) sourceport.ImportContext {
	t.Helper()
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: evidencePortableActor.String(), DisplayName: "Historical Evidence owner"}})
	if err != nil {
		t.Fatal(err)
	}
	return sourceport.ImportContext{IncidentID: incident, ActorUserID: local, OperationID: "evidence-portable", BundleVersion: 8, Actors: actors}
}
func requireEvidencePortableFailure(t testing.TB, err error, invariant string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID() != "evidence" || failure.InvariantID() != "evidence."+invariant {
		t.Fatalf("failure=%v want evidence.%s", err, invariant)
	}
}

type evidencePortableAttributions struct {
	local uuid.UUID
	rows  []incidentportability.ImportedAttribution
}

func (a *evidencePortableAttributions) RecordImportedAttribution(table, row, column, actor string) error {
	a.rows = append(a.rows, incidentportability.ImportedAttribution{SourceTable: table, SourceRowID: row, SourceColumn: column, SourceActorID: actor, LocalUserID: a.local})
	return nil
}
func (a *evidencePortableAttributions) VisitImportedAttributions(table, row string, visit func(incidentportability.ImportedAttribution) error) error {
	for _, value := range a.rows {
		if value.SourceTable == table && (row == "" || value.SourceRowID == row) {
			if err := visit(value); err != nil {
				return err
			}
		}
	}
	return nil
}
func (a *evidencePortableAttributions) ResolvePortableSourceActors(_ context.Context, _ incidentportability.Queryer, _ uuid.UUID, table, column string, ids []string) (map[string]string, error) {
	result := map[string]string{}
	for _, value := range a.rows {
		if value.SourceTable == table && value.SourceColumn == column && slices.Contains(ids, value.SourceRowID) {
			result[value.SourceRowID] = value.SourceActorID
		}
	}
	return result, nil
}

type forbiddenEvidencePortableTx struct{ pgx.Tx }

func TestEvidencePortableRecordAdmission_Unit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	incident, local := uuid.New(), uuid.New()
	port := evidence.NewIncidentBundleSourcePort()
	input := evidencePortableInput(t, incident, local)
	base := evidencePortableRow(incident, 1)
	prepare := func(rows ...map[string]any) error {
		bundle := evidencePortableBundle(t, rows...)
		before := maps.Clone(bundle)
		for path, value := range before {
			before[path] = bytes.Clone(value)
		}
		_, err := port.PrepareImport(ctx, bundle, input)
		if !reflect.DeepEqual(bundle, before) {
			t.Fatal("Prepare changed admitted bytes")
		}
		return err
	}
	if err := prepare(base); err != nil {
		t.Fatal(err)
	}
	if err := prepare(); err != nil {
		t.Fatal("empty owner", err)
	}
	required := map[string]bool{"record_id": true, "incident_id": true, "lifecycle_state": true, "upload_state": true, "created_at": true, "updated_at": true}
	for field := range base {
		for _, variant := range []string{"missing", "type", "null"} {
			t.Run(field+"/"+variant, func(t *testing.T) {
				row := maps.Clone(base)
				switch variant {
				case "missing":
					delete(row, field)
				case "type":
					row[field] = []any{}
				case "null":
					row[field] = nil
				}
				err := prepare(row)
				if variant == "null" && !required[field] {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				invariant := "row_shape_exact"
				if field == "record_id" {
					invariant = "source_identity_admitted"
				}
				requireEvidencePortableFailure(t, err, invariant)
			})
		}
	}
	for _, field := range []string{"record_id", "incident_id", "collector_party_id", "source_party_id", "object_blob_id", "owner_user_id"} {
		for _, value := range []any{uuid.Nil.String(), "not-a-uuid", "10000000000040008000000000000001", "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", 42, true} {
			row := maps.Clone(base)
			row[field] = value
			invariant := "row_shape_exact"
			if field == "record_id" {
				invariant = "source_identity_admitted"
			}
			requireEvidencePortableFailure(t, prepare(row), invariant)
		}
	}
	for _, field := range []string{"created_at", "updated_at", "requested_at", "received_at"} {
		for _, value := range []string{"0000-01-01T00:00:00Z", "2026-01-01T00:00:00.1234567Z", "2026-01-01T00:00:00.100000Z", "2026-01-01T01:00:00+01:00", "2026-01-01", "2026-01-01T00:00:00z"} {
			row := maps.Clone(base)
			row[field] = value
			requireEvidencePortableFailure(t, prepare(row), "row_shape_exact")
		}
		for _, value := range []string{"0001-01-01T00:00:00Z", "9999-12-31T23:59:59.999999+00:00", "2026-01-01T00:00:00.1Z"} {
			row := maps.Clone(base)
			row[field] = value
			if err := prepare(row); err != nil {
				t.Fatal(field, value, err)
			}
		}
	}
	for _, field := range []string{"title", "storage_ref", "collector_party_text", "source_party_text"} {
		for _, value := range []string{"", " ", " padded", "padded ", "e\u0301", "before\nafter", "before\x00after"} {
			row := maps.Clone(base)
			row[field] = value
			requireEvidencePortableFailure(t, prepare(row), "row_shape_exact")
		}
		row := maps.Clone(base)
		row[field] = "é preserved\u200d text"
		if err := prepare(row); err != nil {
			t.Fatal(field, err)
		}
	}
	for _, tc := range []struct {
		field     string
		value     any
		invariant string
	}{
		{"unknown", true, "row_shape_exact"}, {"incident_id", uuid.NewString(), "envelope_type_scope"},
		{"lifecycle_state", "unknown", "lifecycle_legal"}, {"upload_state", "unknown", "lifecycle_legal"},
		{"lifecycle_state", "available", "lifecycle_legal"}, {"lifecycle_state", "released", "lifecycle_legal"},
		{"storage_ref", "object://10000000-0000-4000-8000-000000000001", "storage_reference_legal"},
		{"owner_user_id", uuid.NewString(), "attribution_exact"},
	} {
		row := maps.Clone(base)
		row[tc.field] = tc.value
		requireEvidencePortableFailure(t, prepare(row), tc.invariant)
	}
	for _, value := range []string{"\x00digest", "dig\x00est", "digest\x00"} {
		row := maps.Clone(base)
		row["blob_hash"] = value
		requireEvidencePortableFailure(t, prepare(row), "row_shape_exact")
	}
	validHash := maps.Clone(base)
	validHash["blob_hash"] = strings.Repeat("a", 64)
	if err := prepare(validHash); err != nil {
		t.Fatal(err)
	}
	second := evidencePortableRow(incident, 2)
	if err := prepare(base, second); err != nil {
		t.Fatal(err)
	}
	requireEvidencePortableFailure(t, prepare(base, base), "source_identity_admitted")
	requireEvidencePortableFailure(t, prepare(second, base), "source_identity_admitted")
	malformed := []string{"null\n", "[]\n", "{} {}\n", "\n", "{\"record_id\":\"a\",\"record_id\":\"b\"}\n", "{\"title\":\"\\ud800\"}\n", "{\"title\":\"" + string([]byte{0xff}) + "\"}\n"}
	for _, raw := range malformed {
		bundle := evidencePortableBundle(t)
		bundle[evidencePortablePath] = []byte(raw)
		_, err := port.PrepareImport(ctx, bundle, input)
		requireEvidencePortableFailure(t, err, "row_shape_exact")
	}
	// A later syntax or shape failure outranks an earlier semantic failure.
	badIdentity := maps.Clone(base)
	badIdentity["record_id"] = "invalid"
	bundle := evidencePortableBundle(t, badIdentity)
	bundle[evidencePortablePath] = append(bundle[evidencePortablePath], []byte("null\n")...)
	_, err := port.PrepareImport(ctx, bundle, input)
	requireEvidencePortableFailure(t, err, "row_shape_exact")
	early := maps.Clone(base)
	early["lifecycle_state"] = "unknown"
	late := maps.Clone(second)
	delete(late, "title")
	requireEvidencePortableFailure(t, prepare(early, late), "row_shape_exact")
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = port.PrepareImport(cancelled, evidencePortableBundle(t, base), input)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation=%v", err)
	}
	prepared, err := port.PrepareImport(ctx, evidencePortableBundle(t, base), input)
	if err != nil {
		t.Fatal(err)
	}
	input.Attributions = &evidencePortableAttributions{local: local}
	input.RewrittenObjectBlobs = evidencePortableBundle(t)
	for _, variant := range []string{"instance", "operation", "incident", "actor", "version", "attributions"} {
		other := input
		target := port
		switch variant {
		case "instance":
			target = evidence.NewIncidentBundleSourcePort()
		case "operation":
			other.OperationID += "-other"
		case "incident":
			other.IncidentID = uuid.New()
		case "actor":
			other.ActorUserID = uuid.New()
		case "version":
			other.BundleVersion = 7
		case "attributions":
			other.Attributions = nil
		}
		for _, call := range []func(context.Context, pgx.Tx, sourceport.Prepared, sourceport.ImportContext) error{target.ApplyImportTx, target.ValidateImportTx} {
			if err := call(ctx, &forbiddenEvidencePortableTx{}, prepared, other); !errors.Is(err, sourceport.ErrPreparedBinding) {
				t.Fatalf("%s binding=%v", variant, err)
			}
		}
	}
	// Unknown source names must never be accepted through a user-reference suffix.
	row := maps.Clone(base)
	row["extra_user_id"] = evidencePortableActor.String()
	requireEvidencePortableFailure(t, prepare(row), "row_shape_exact")
	for _, binding := range []struct {
		field string
		limit int
	}{{"title", 512}, {"storage_ref", 1024}, {"collector_party_text", 256}, {"source_party_text", 256}} {
		row := maps.Clone(base)
		row[binding.field] = strings.Repeat("😀", binding.limit)
		if err := prepare(row); err != nil {
			t.Fatal(binding.field, err)
		}
		row[binding.field] = strings.Repeat("😀", binding.limit+1)
		requireEvidencePortableFailure(t, prepare(row), "row_shape_exact")
	}
}
