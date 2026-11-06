package imports

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestCommittedSourcePortableAdmission(t *testing.T) {
	ctx := context.Background()
	incidentID, actorID, unitID := uuid.New(), uuid.New(), uuid.New()
	original := []byte("Title,Unmapped\r\n\"Literal <script>界</script>\",\"original\r\nbytes\"\r\n")
	sum := sha256.Sum256(original)
	digest := hex.EncodeToString(sum[:])
	request, apiErr := decodeMappingRequest(strings.NewReader(`{"client_txn_id":"map","target_view_schema_id":"cartulary.view.timeline.v2","header_row_ref":1,"data_start_row_ref":2,"unknown_column_policy":"preserve_raw_capture","source_columns":[{"source_column_ordinal":1,"source_header_text":"Title","field_key":"timeline.activity_synopsis_text","entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"},{"source_column_ordinal":2,"source_header_text":"Unmapped","field_key":null,"entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"}]}`), nil)
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	mapping, err := json.Marshal(request.approvedMapping)
	if err != nil {
		t.Fatal(err)
	}
	source := CommittedSource{SchemaID: "cartulary.imports.committed_source.v1", CommittedSourceID: CommittedSourceID(unitID), IncidentID: incidentID, ContentSHA256: digest, ByteSize: int64(len(original)), SourceFileKind: "csv", SourceMediaType: "text/csv", OriginalFilename: "original.csv", ParserProfileID: parserProfileWorkbookImport, ParserVersion: parserVersionWorkbookImport, LocatorKind: "csv_file", Locator: "file", HeaderRowRef: 1, DataStartRowRef: 2, ApprovedMapping: mapping, MappingFingerprint: request.Fingerprint, TargetKind: ImportTargetKindViewSchema, TargetViewSchemaID: "cartulary.view.timeline.v2", OriginImportUnitID: unitID, OriginImportSessionID: uuid.New(), CreatedByUserID: actorID, CommittedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: actorID.String(), DisplayName: "Historical source author"}})
	if err != nil {
		t.Fatal(err)
	}
	makeBundle := func(value CommittedSource) portabilitysupport.MapBundle {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return portabilitysupport.MapBundle{committedSourcesPath: append(raw, '\n'), committedSourceBlobPrefix + digest: bytes.Clone(original)}
	}
	bundle := makeBundle(source)
	prepared, err := PrepareCommittedSources(ctx, bundle, incidentID, actors)
	if err != nil {
		t.Fatal(err)
	}
	retained, reader, err := prepared.OpenSource(ctx, source.CommittedSourceID)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, actual) || retained.CreatedByUserID != actorID || retained.OriginImportUnitID != unitID {
		t.Fatal("source bytes or historical origin changed")
	}
	if _, err := prepared.ReadSource(ctx, uuid.New()); err == nil {
		t.Fatal("unadmitted source read succeeded")
	}
	// A reusable preparation must still bind the exact admitted bytes on reopen.
	bundle[committedSourceBlobPrefix+digest] = []byte("changed")
	if _, reader, err := prepared.OpenSource(ctx, source.CommittedSourceID); err == nil {
		defer reader.Close()
		if _, err := io.ReadAll(reader); err == nil {
			t.Fatal("mutable backing member escaped preparation binding")
		}
	}
	for name, mutate := range map[string]func(*CommittedSource){
		"cross incident":   func(s *CommittedSource) { s.IncidentID = uuid.New() },
		"missing actor":    func(s *CommittedSource) { s.CreatedByUserID = uuid.New() },
		"changed mapping":  func(s *CommittedSource) { s.MappingFingerprint = strings.Repeat("0", 64) },
		"changed header":   func(s *CommittedSource) { s.HeaderRowRef = 2; s.DataStartRowRef = 3 },
		"changed target":   func(s *CommittedSource) { s.TargetViewSchemaID = "cartulary.view.hosts.v1" },
		"changed identity": func(s *CommittedSource) { s.OriginImportUnitID = uuid.New() },
		"wrong parser":     func(s *CommittedSource) { s.ParserVersion = "unsupported" },
		"wrong blob size":  func(s *CommittedSource) { s.ByteSize++ },
	} {
		t.Run(name, func(t *testing.T) {
			value := source
			mutate(&value)
			if _, err := PrepareCommittedSources(ctx, makeBundle(value), incidentID, actors); err == nil {
				t.Fatal("invalid committed source admitted")
			}
		})
	}
	for name, mutate := range map[string]func(portabilitysupport.MapBundle){
		"missing blob":  func(b portabilitysupport.MapBundle) { delete(b, committedSourceBlobPrefix+digest) },
		"tampered blob": func(b portabilitysupport.MapBundle) { b[committedSourceBlobPrefix+digest][0] = 'x' },
		"undeclared blob": func(b portabilitysupport.MapBundle) {
			sum := sha256.Sum256([]byte("unused"))
			b[committedSourceBlobPrefix+hex.EncodeToString(sum[:])] = []byte("unused")
		},
		"duplicate source": func(b portabilitysupport.MapBundle) {
			b[committedSourcesPath] = append(b[committedSourcesPath], b[committedSourcesPath]...)
		},
		"missing empty field": func(b portabilitysupport.MapBundle) {
			b[committedSourcesPath] = bytes.Replace(b[committedSourcesPath], []byte(`"source_rect_a1":"",`), nil, 1)
		},
		"unknown field": func(b portabilitysupport.MapBundle) {
			b[committedSourcesPath] = bytes.Replace(b[committedSourcesPath], []byte(`"schema_id":`), []byte(`"unknown":true,"schema_id":`), 1)
		},
	} {
		t.Run(name, func(t *testing.T) {
			b := makeBundle(source)
			mutate(b)
			if _, err := PrepareCommittedSources(ctx, b, incidentID, actors); err == nil {
				t.Fatal("invalid source members admitted")
			}
		})
	}
}
