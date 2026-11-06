package sourcestate

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type decodedFiles struct{ links, tagCatalog, recordTags []map[string]any }

var (
	testIncidentID = uuid.MustParse("41000000-0000-0000-0000-000000000001")
	testActorID    = uuid.MustParse("41000000-0000-0000-0000-000000000002")
	testSourceID   = uuid.MustParse("41000000-0000-0000-0000-000000000003")
	testTargetID   = uuid.MustParse("41000000-0000-0000-0000-000000000004")
)

func TestPrepareImportEnforcesStrictNDJSONAndExactMembers(t *testing.T) {
	port, err := NewSourcePort()
	if err != nil {
		t.Fatalf("construct source port: %v", err)
	}
	valid := validDecodedFiles()
	validBundle := bundleFromDecoded(t, valid)
	if _, err := port.PrepareImport(context.Background(), validBundle, testImportContext("valid")); err != nil {
		t.Fatalf("prepare valid import: %v", err)
	}

	tests := []struct {
		name   string
		bundle portabilitysupport.MapBundle
		check  func(error) bool
	}{
		{
			name: "blank line",
			bundle: func() portabilitysupport.MapBundle {
				bundle := bundleFromDecoded(t, validDecodedFiles())
				bundle["data/record_links.ndjson"] = append([]byte{'\n'}, bundle["data/record_links.ndjson"]...)
				return bundle
			}(),
			check: func(err error) bool {
				var malformed *incidentportability.MalformedPayloadError
				return errors.As(err, &malformed)
			},
		},
		{
			name: "multiple values",
			bundle: portabilitysupport.MapBundle{
				"data/record_links.ndjson": []byte("{} {}\n"),
				"data/tags.ndjson":         {},
				"data/record_tags.ndjson":  {},
			},
			check: func(err error) bool {
				var malformed *incidentportability.MalformedPayloadError
				return errors.As(err, &malformed)
			},
		},
		{
			name: "unknown member",
			bundle: func() portabilitysupport.MapBundle {
				files := validDecodedFiles()
				files.links[0]["future"] = true
				return bundleFromDecoded(t, files)
			}(),
			check: func(err error) bool {
				return failureInvariant(err) == "links_tags.source_identity_admitted"
			},
		},
		{
			name: "missing required file",
			bundle: func() portabilitysupport.MapBundle {
				bundle := bundleFromDecoded(t, validDecodedFiles())
				delete(bundle, "data/tags.ndjson")
				return bundle
			}(),
			check: func(err error) bool {
				var failure *incidentportability.VerificationFailure
				return errors.As(err, &failure)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := port.PrepareImport(context.Background(), test.bundle, testImportContext(test.name))
			if err == nil || !test.check(err) {
				t.Fatalf("PrepareImport error = %T %[1]v", err)
			}
		})
	}
}

func TestValidationPriorityIsFixedAcrossMultipleDefects(t *testing.T) {
	foreignIncident := uuid.MustParse("41000000-0000-0000-0000-000000000099")
	tests := []struct {
		name       string
		mutate     func(*decodedFiles)
		incidentID uuid.UUID
		want       string
	}{
		{
			name: "source identity precedes every later defect", want: "links_tags.source_identity_admitted", incidentID: foreignIncident,
			mutate: func(files *decodedFiles) {
				files.links[0]["unknown"] = true
				files.links[0]["link_type"] = "unknown"
				files.links[0]["deleted_at"] = testTimestamp()
				files.tagCatalog = nil
			},
		},
		{
			name: "link tuple precedes deletion and tags", want: "links_tags.link_tuple_legal", incidentID: testIncidentID,
			mutate: func(files *decodedFiles) {
				files.links[0]["link_type"] = "unknown"
				files.links[0]["deleted_at"] = testTimestamp()
				files.recordTags[0]["tag_name"] = " bad "
			},
		},
		{
			name: "deletion precedes tag normalization", want: "links_tags.deletion_tuple_legal", incidentID: testIncidentID,
			mutate: func(files *decodedFiles) {
				files.links[0]["deleted_at"] = testTimestamp()
				files.recordTags[0]["tag_name"] = " bad "
			},
		},
		{
			name: "tag normalization precedes catalog", want: "links_tags.tag_normalized", incidentID: testIncidentID,
			mutate: func(files *decodedFiles) {
				files.recordTags[0]["tag_name"] = " bad "
				files.tagCatalog = nil
			},
		},
	}
	stateManifest, err := loadManifest()
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files := validDecodedFiles()
			test.mutate(&files)
			importContext := testImportContext(test.name)
			importContext.IncidentID = test.incidentID
			_, err := prepareImport(context.Background(), stateManifest.descriptor(), stateManifest, bundleFromDecoded(t, files), importContext)
			got := failureInvariant(err)
			if got != test.want {
				t.Fatalf("selected invariant = %q, want %q", got, test.want)
			}
		})
	}
}

func TestPreparedValueIsOpaqueCopiedAndWrongTypesDoNotPanic(t *testing.T) {
	stateManifest, err := loadManifest()
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	descriptor := stateManifest.descriptor()
	prepared, err := prepareImport(
		context.Background(), descriptor, stateManifest, bundleFromDecoded(t, validDecodedFiles()), testImportContext("opaque"),
	)
	if err != nil {
		t.Fatalf("prepare import: %v", err)
	}
	var retained map[string]any
	if err := prepared.walk(context.Background(), pathRecordLinks, func(row map[string]any) error { retained = row; return nil }); err != nil {
		t.Fatal(err)
	}
	retained["link_type"] = "corrupted"
	if err := prepared.walk(context.Background(), pathRecordLinks, func(row map[string]any) error {
		if row["link_type"] != "references_record" {
			t.Fatal("prepared member exposed retained mutable rows")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("wrong prepared type panicked: %v", recovered)
		}
	}()
	err = applyPreparedTx(
		context.Background(), nil, struct{}{}, testImportContext("wrong"), stateManifest, descriptor,
	)
	var stateError *sourceStateError
	if !errors.As(err, &stateError) || err.Error() != "links: invalid prepared source state" {
		t.Fatalf("wrong prepared type error = %T %[1]v, want private safe error", err)
	}
}

func validDecodedFiles() decodedFiles {
	now := testTimestamp()
	linkID := uuid.MustParse("41000000-0000-0000-0000-000000000010")
	tagID := uuid.MustParse("41000000-0000-0000-0000-000000000011")
	return decodedFiles{
		links: []map[string]any{{
			"record_link_id": linkID.String(), "incident_id": testIncidentID.String(),
			"src_record_id": testSourceID.String(), "dst_record_id": testTargetID.String(),
			"link_type": "references_record", "field_key": nil, "provenance": "manual", "confidence": nil,
			"owner_user_id": testActorID.String(), "created_by_user_id": testActorID.String(),
			"decided_at": now, "created_at": now, "deleted_at": nil, "deleted_by_user_id": nil,
		}},
		tagCatalog: []map[string]any{{"tag_name": "Urgent", "normalized_tag_name": "urgent"}},
		recordTags: []map[string]any{{
			"record_tag_id": tagID.String(), "incident_id": testIncidentID.String(), "record_id": testSourceID.String(),
			"tag_name": "Urgent", "normalized_tag_name": "urgent", "created_by_user_id": testActorID.String(),
			"created_at": now, "updated_at": now, "deleted_at": nil, "deleted_by_user_id": nil,
		}},
	}
}

func testTimestamp() string {
	return time.Date(2026, time.August, 27, 14, 30, 0, 123000000, time.UTC).Format("2006-01-02T15:04:05.999999") + "+00:00"
}

func bundleFromDecoded(t *testing.T, files decodedFiles) portabilitysupport.MapBundle {
	t.Helper()
	return portabilitysupport.MapBundle{
		"data/record_links.ndjson": encodeTestRows(t, files.links),
		"data/tags.ndjson":         encodeTestRows(t, files.tagCatalog),
		"data/record_tags.ndjson":  encodeTestRows(t, files.recordTags),
	}
}

func encodeTestRows(t *testing.T, rows []map[string]any) []byte {
	t.Helper()
	var result []byte
	for _, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("encode test row: %v", err)
		}
		result = append(result, encoded...)
		result = append(result, '\n')
	}
	return result
}

func testImportContext(operationID string) sourceport.ImportContext {
	return sourceport.ImportContext{
		IncidentID: testIncidentID, ActorUserID: testActorID, BundleVersion: 7,
		OperationID: operationID,
	}
}

func failureInvariant(err error) string {
	var failure *sourceport.Failure
	if !errors.As(err, &failure) {
		return ""
	}
	return failure.InvariantID()
}
