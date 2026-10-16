package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func testCompactRenderAdmission(t *testing.T, payload releaseCreateJobPayload, jobID, incidentID, actorID uuid.UUID, at time.Time) (renderAdmission, renderInput) {
	t.Helper()
	profile, err := selectedRedactionProfile(payload.RedactionProfileID, payload.RedactionProfileVersion, payload.RecipientPartitionRefs)
	if err != nil {
		t.Fatal(err)
	}
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	templateSHA, err := templateManifestDigest(payload.TemplateContract)
	if err != nil {
		t.Fatal(err)
	}
	graphSHA, err := graphSourcesDigest(payload.GraphSources)
	if err != nil {
		t.Fatal(err)
	}
	admission := renderAdmission{
		SchemaID: renderAdmissionSchema, Algorithm: renderPreparationAlgorithm, Identity: payload.ExportModel.RenderIdentity,
		ActorUserID: payload.ActorUserID, IncidentID: payload.IncidentID, SnapshotID: payload.SnapshotID, SnapshotAt: payload.SnapshotAt,
		SourceChangeSetHighWatermark: payload.SourceChangeSetHighWatermark, DerivationVersion: payload.DerivationVersion,
		SourceCaptureSHA256: strings.Repeat("a", 64), ReferencePacks: payload.ExportModel.ReferencePacks,
		TemplateContract: payload.TemplateContract, TemplateFiles: []TemplatePackFile{}, TemplateSHA256: templateSHA,
		Profile: profile, ProfileSHA256: profileSHA, ReleaseScope: payload.ReleaseScope, RecipientPartitionRefs: []string{},
		OutputKind: payload.OutputKind, OutputOptions: payload.OutputOptions, GraphProjectionRefs: payload.GraphProjectionRefs,
		GraphSources: payload.GraphSources, GraphSourcesSHA256: graphSHA, ClientTxnID: payload.ClientTxnID, NormalizedRequest: payload.NormalizedRequest,
	}
	captureRaw, captureDigest, err := freezeSnapshotCapture(IncidentMetadataSnapshot{ID: payload.IncidentID, Title: "Frozen", Status: "active", Version: 1}, payload.SnapshotID, payload.SnapshotAt, payload.SourceChangeSetHighWatermark, payload.ExportModel.ReferencePacks, fixtureSnapshotSource(nil))
	if err != nil {
		t.Fatal(err)
	}
	admission.SourceCaptureSHA256 = captureDigest
	raw, err := marshalRenderAdmission(admission)
	if err != nil {
		t.Fatal(err)
	}
	digest := reportingObjectDigest(renderAdmissionSchema, raw)
	if _, err := decodeRenderAdmission(raw, digest, jobID, incidentID, actorID, at); err != nil {
		t.Fatal(err)
	}
	var original map[string]json.RawMessage
	if json.Unmarshal(raw, &original) != nil {
		t.Fatal("fixture")
	}
	for _, forbidden := range []string{"export_model", "sections", "source_facts"} {
		if _, exists := original[forbidden]; exists {
			t.Fatal("admission eagerly retained expanded source", forbidden)
		}
	}
	record := SnapshotRecord{SnapshotID: uuid.MustParse(admission.SnapshotID), IncidentID: incidentID, SnapshotAt: admission.SnapshotAt,
		SourceChangeSetHighWatermark: admission.SourceChangeSetHighWatermark, DerivationVersion: admission.DerivationVersion,
		SourceCaptureJSON: captureRaw, SourceCaptureSHA256: captureDigest}
	input, prepared, err := materializeRenderAdmission(context.Background(), admission, digest, record)
	if err != nil {
		t.Fatal("prepare frozen recipe", err)
	}
	if input.ExportModelSHA256 != payload.ExportModelSHA256 || !sameReportingValue(input.ExportModel, payload.ExportModel) {
		t.Fatal("recipe changed projection bytes")
	}
	after, err := marshalRenderAdmission(admission)
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatal("preparation mutated admission")
	}
	preparedRaw, err := canonicalJSON(prepared)
	if err != nil {
		t.Fatal(err)
	}
	preparedDigest := reportingObjectDigest(renderPreparationSchema, preparedRaw)
	if _, err := decodeRenderPreparation(preparedRaw, preparedDigest, admission, digest); err != nil {
		t.Fatal(err)
	}
	if _, err := decodeRenderPreparation(preparedRaw, preparedDigest, admission, strings.Repeat("0", 64)); err == nil {
		t.Fatal("preparation changed admission")
	}
	changed := record
	changed.SourceCaptureSHA256 = strings.Repeat("0", 64)
	if _, _, err := materializeRenderAdmission(context.Background(), admission, digest, changed); err == nil {
		t.Fatal("preparation accepted substituted capture")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := materializeRenderAdmission(canceled, admission, digest, record); err != context.Canceled {
		t.Fatal("preparation ignored cancellation", err)
	}
	for name, edit := range map[string]func(map[string]json.RawMessage){
		"unknown": func(v map[string]json.RawMessage) { v["future"] = json.RawMessage(`true`) },
		"retired": func(v map[string]json.RawMessage) { v["preparation_algorithm"] = json.RawMessage(`"retired"`) },
		"identity": func(v map[string]json.RawMessage) {
			v["snapshot_id"] = json.RawMessage(`"00000000-0000-0000-0000-000000000412"`)
		},
		"scope":                   func(v map[string]json.RawMessage) { v["release_scope"] = json.RawMessage(`"external_release"`) },
		"missing capture binding": func(v map[string]json.RawMessage) { delete(v, "source_capture_sha256") },
		"missing policy":          func(v map[string]json.RawMessage) { delete(v, "authored_redaction_profile") },
		"different request":       func(v map[string]json.RawMessage) { v["client_txn_id"] = json.RawMessage(`"different"`) },
		"null template files":     func(v map[string]json.RawMessage) { v["template_files"] = json.RawMessage(`null`) },
		"null recipients":         func(v map[string]json.RawMessage) { v["recipient_partition_refs"] = json.RawMessage(`null`) },
		"legacy model":            func(v map[string]json.RawMessage) { v["export_model"] = json.RawMessage(`{}`) },
		"invalid timestamp": func(v map[string]json.RawMessage) {
			v["snapshot_at"] = json.RawMessage(`"2026-10-14T12:00:00.0000001Z"`)
		},
		"after admission":         func(v map[string]json.RawMessage) { v["snapshot_at"] = json.RawMessage(`"2026-10-15T12:00:00Z"`) },
		"altered authored policy": func(v map[string]json.RawMessage) { v["authored_redaction_profile"] = json.RawMessage(`{}`) },
		"altered template":        func(v map[string]json.RawMessage) { v["template_contract"] = json.RawMessage(`{}`) },
		"altered graphs":          func(v map[string]json.RawMessage) { v["graph_sources"] = json.RawMessage(`{}`) },
		"altered request bytes":   func(v map[string]json.RawMessage) { v["normalized_request"] = json.RawMessage(`"e30="`) },
	} {
		t.Run(name, func(t *testing.T) {
			value := map[string]json.RawMessage{}
			for k, v := range original {
				value[k] = v
			}
			edit(value)
			changed, err := canonicalJSON(value)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decodeRenderAdmission(changed, reportingObjectDigest(renderAdmissionSchema, changed), jobID, incidentID, actorID, at); err == nil {
				t.Fatal("malformed recipe accepted after rehash")
			}
		})
	}
	altered := bytes.Replace(raw, []byte(admission.SourceCaptureSHA256), []byte(strings.Repeat("b", 64)), 1)
	if _, err := decodeRenderAdmission(altered, digest, jobID, incidentID, actorID, at); err == nil {
		t.Fatal("changed recipe bypassed independent digest")
	}
	return admission, input
}
