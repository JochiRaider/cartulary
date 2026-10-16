package reporting

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func assertFrozenReleaseBindings(t *testing.T, payload releaseCreateJobPayload, jobID uuid.UUID) {
	t.Helper()
	assertFrozenTokenValueBinding(t, payload.renderInput)
	assertFrozenMermaidSourceBinding(t, payload.renderInput)
	id := reportingResourceID(jobID, "release")
	rendered, _, err := fixtureRenderer().Render(context.Background(), payload.renderInput)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateReleaseCandidateBinding(payload, jobID, id, rendered); err != nil {
		t.Fatal("valid completion rejected", err)
	}
	if rendered.RenderBundle.Manifest.AdmittedExportModelSHA256 != payload.ExportModelSHA256 || rendered.RenderBundle.Manifest.ExportModelSHA256 != rendered.Redaction.ModelSHA256 || rendered.Redaction.ModelSHA256 == payload.ExportModelSHA256 {
		t.Fatal("bundle conflated frozen admission with assembled model")
	}
	for name, mutate := range map[string]func(map[string]any){
		"missing admission digest": func(m map[string]any) { delete(m, "admitted_export_model_sha256") },
		"null admission digest":    func(m map[string]any) { m["admitted_export_model_sha256"] = nil },
		"missing assembled digest": func(m map[string]any) { delete(m, "export_model_sha256") },
		"retired bundle": func(m map[string]any) {
			m["schema_id"], m["manifest_version"] = "cartulary.render_bundle_manifest.v2", "2"
		},
	} {
		t.Run(name, func(t *testing.T) {
			var object map[string]any
			if err := json.Unmarshal(rendered.RenderBundle.ManifestJSON, &object); err != nil {
				t.Fatal(err)
			}
			mutate(object)
			candidate := rendered.RenderBundle
			candidate.ManifestJSON, err = canonicalJSON(object)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(candidate.ManifestJSON, &candidate.Manifest); err != nil {
				t.Fatal(err)
			}
			candidate.ManifestSHA256 = reportingObjectDigest(candidate.Manifest.SchemaID, candidate.ManifestJSON)
			if validateShareableBundle(candidate) == nil {
				t.Fatal("incomplete or retired identity admitted")
			}
		})
	}
	assertFrozenSlidevSourceBinding(t, payload.renderInput, rendered)
	record := ReleaseRecord{
		ReleaseID: id, CreateJobID: jobID, IncidentID: uuid.MustParse(payload.IncidentID), SnapshotID: uuid.MustParse(payload.SnapshotID),
		CreatedByUserID: uuid.MustParse(payload.ActorUserID), ClientTxnID: payload.ClientTxnID, ReleaseScope: payload.ReleaseScope,
		SnapshotAt: payload.SnapshotAt, SourceChangeSetHighWatermark: payload.SourceChangeSetHighWatermark, DerivationVersion: payload.DerivationVersion,
		ExportModelSHA256: payload.ExportModelSHA256, TemplateID: payload.TemplateID, TemplateVersion: payload.TemplateVersion,
		RedactionProfileID: payload.RedactionProfileID, RedactionProfileVersion: payload.RedactionProfileVersion, RedactionProfileSHA256: rendered.ProfileSHA256,
		OutputKind: payload.OutputKind, OutputOptions: payload.OutputOptions, GraphProjectionRefs: payload.GraphProjectionRefs,
		RenderAdmittedAt: payload.RenderAdmittedAt, RecipientPartitionRefs: payload.RecipientPartitionRefs, ReleaseState: ReleaseStateApproved,
		OutputSHA256: &rendered.OutputSHA256, OutputMediaType: &rendered.OutputMediaType,
		RedactionManifestSHA256: &rendered.RedactionManifestSHA256, RedactionManifestJSON: rendered.RedactionManifestJSON,
	}
	if err := validateReleaseAdmissionBinding(payload, jobID, record); err != nil {
		t.Fatal(err)
	}
	if err := validateReleaseOutputBinding(record, rendered); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*ReleaseRecord){
		"identity":            func(r *ReleaseRecord) { r.ReleaseID = jobID },
		"job":                 func(r *ReleaseRecord) { r.CreateJobID = id },
		"snapshot":            func(r *ReleaseRecord) { r.SnapshotID = jobID },
		"incident":            func(r *ReleaseRecord) { r.IncidentID = jobID },
		"actor":               func(r *ReleaseRecord) { r.CreatedByUserID = jobID },
		"transaction":         func(r *ReleaseRecord) { r.ClientTxnID = "different" },
		"source":              func(r *ReleaseRecord) { r.ExportModelSHA256 = strings.Repeat("0", 64) },
		"boundary":            func(r *ReleaseRecord) { r.SourceChangeSetHighWatermark = "different" },
		"capture time":        func(r *ReleaseRecord) { r.SnapshotAt = r.SnapshotAt.Add(time.Second) },
		"derivation":          func(r *ReleaseRecord) { r.DerivationVersion = "retired" },
		"admission time":      func(r *ReleaseRecord) { r.RenderAdmittedAt = r.RenderAdmittedAt.Add(time.Second) },
		"template":            func(r *ReleaseRecord) { r.TemplateID = "other" },
		"template version":    func(r *ReleaseRecord) { r.TemplateVersion = "other" },
		"profile":             func(r *ReleaseRecord) { r.RedactionProfileID = "other" },
		"profile version":     func(r *ReleaseRecord) { r.RedactionProfileVersion = "other" },
		"kind":                func(r *ReleaseRecord) { r.OutputKind = OutputKindMermaid },
		"scope":               func(r *ReleaseRecord) { r.ReleaseScope = ReleaseScopeExternal },
		"options":             func(r *ReleaseRecord) { r.OutputOptions = json.RawMessage(`{}`) },
		"graphs":              func(r *ReleaseRecord) { r.GraphProjectionRefs = json.RawMessage(`null`) },
		"composition":         func(r *ReleaseRecord) { r.CompositionID = &jobID },
		"composition version": func(r *ReleaseRecord) { r.CompositionVersion = &payload.TemplateVersion },
		"composition digest":  func(r *ReleaseRecord) { r.CompositionSHA256 = &payload.ExportModelSHA256 },
		"recipients":          func(r *ReleaseRecord) { r.RecipientPartitionRefs = []string{"public"} },
	} {
		t.Run("completion row "+name, func(t *testing.T) {
			changed := record
			mutate(&changed)
			if validateReleaseAdmissionBinding(payload, jobID, changed) == nil {
				t.Fatal("substituted admission accepted")
			}
		})
	}
	for name, mutate := range map[string]func(*RenderBundleManifest){
		"renderer identity": func(m *RenderBundleManifest) { m.RendererID = "other" },
		"retired renderer":  func(m *RenderBundleManifest) { m.RendererVersion = "8" },
		"source":            func(m *RenderBundleManifest) { m.AdmittedExportModelSHA256 = strings.Repeat("a", 64) },
		"assembled model":   func(m *RenderBundleManifest) { m.ExportModelSHA256 = strings.Repeat("b", 64) },
		"template":          func(m *RenderBundleManifest) { m.TemplateID = "other" },
		"template digest":   func(m *RenderBundleManifest) { m.TemplateManifestSHA256 = strings.Repeat("b", 64) },
		"scope": func(m *RenderBundleManifest) {
			if m.ReleaseScope == ReleaseScopeInternalDraft {
				m.ReleaseScope = ReleaseScopeInternalReview
			} else {
				m.ReleaseScope = ReleaseScopeInternalDraft
			}
		},
		"kind":                 func(m *RenderBundleManifest) { m.OutputKind = OutputKindMermaid },
		"reference provenance": func(m *RenderBundleManifest) { m.ReferencePacks.SHA256 = strings.Repeat("c", 64) },
	} {
		t.Run("self consistent output "+name, func(t *testing.T) {
			changed := rendered
			mutate(&changed.RenderBundle.Manifest)
			changed.RenderBundle.ManifestJSON, err = canonicalJSON(changed.RenderBundle.Manifest)
			if err != nil {
				t.Fatal(err)
			}
			changed.RenderBundle.ManifestSHA256 = reportingObjectDigest(RenderBundleManifestSchemaID, changed.RenderBundle.ManifestJSON)
			changed.OutputSHA256 = changed.RenderBundle.ManifestSHA256
			shareableErr := validateShareableBundle(changed.RenderBundle)
			if name == "kind" {
				// The retained Slidev version now independently contradicts a
				// substituted Mermaid kind at the shared artifact boundary.
				if shareableErr == nil {
					t.Fatal("toolchain accepted a contradictory output kind")
				}
			} else if shareableErr != nil {
				t.Fatal("fixture must remain internally consistent", shareableErr)
			}
			if validateReleaseCandidateBinding(payload, jobID, id, changed) == nil {
				t.Fatal("rehashed output replaced admitted input")
			}
		})
	}
	if validateReleaseCandidateBinding(payload, jobID, uuid.New(), rendered) == nil {
		t.Fatal("completion replaced derived release ID")
	}
	for _, state := range []string{ReleaseStatePendingApproval, ReleaseStateApproved, ReleaseStatePublished, ReleaseStateInvalidated} {
		changed := record
		changed.ReleaseState = state
		if validateReleaseAdmissionBinding(payload, jobID, changed) != nil || validateReleaseOutputBinding(changed, rendered) != nil {
			t.Fatal("lifecycle transition replaced immutable binding")
		}
	}
	retained := jobs.RetainedExtensionJob{Resource: jobs.Resource{Status: jobs.StatusSucceeded, ResultSummary: &jobs.ResultSummary{
		Code: reportingJobReleaseCreatedCode, ResourceRefs: []jobs.ResourceRef{{Kind: "release", ID: id.String(), Route: "/api/v1/releases/" + id.String()}},
	}}}
	if err := validateReleaseJobOutcome(record, retained); err != nil {
		t.Fatal(err)
	}
	retained.Resource.ResultSummary.ResourceRefs[0].ID = jobID.String()
	if validateReleaseJobOutcome(record, retained) == nil {
		t.Fatal("wrong completed resource admitted")
	}
	retained.ExpiredAt = &payload.RenderAdmittedAt
	retained.Resource.ResultSummary = nil
	if validateReleaseJobOutcome(record, retained) != nil {
		t.Fatal("expired successful history rejected")
	}
	retained.Resource.Status = jobs.StatusFailed
	if validateReleaseJobOutcome(record, retained) == nil {
		t.Fatal("failed Job accepted for successful release")
	}
}

func assertFrozenTokenValueBinding(t *testing.T, payload renderInput) {
	t.Helper()
	fields := fixtureSourceFields(t, payload.ExportModel.SnapshotContent)
	fields = append(fields, ExportField{Path: "/parties/token-party", SourceFamily: "party", SourceRecordID: "token-party", SourceRecordType: "party", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Admitted private value"}})
	snapshot, _, err := buildStructuredSnapshotModel(payload.IncidentID, payload.SnapshotID, payload.SnapshotAt, payload.SourceChangeSetHighWatermark, fixtureSnapshotSource(fields))
	if err != nil {
		t.Fatal(err)
	}
	snapshot.ReferencePacks = payload.ExportModel.ReferencePacks
	payload.SourceFacts = snapshot.SourceFacts.Clone()
	payload.SourceFactsSHA256, err = sourceFactsDigest(payload.SourceFacts)
	if err != nil {
		t.Fatal(err)
	}
	payload.ExportModel, payload.ExportModelSHA256, err = bindRenderModel(snapshot, payload.ExportModel.RenderIdentity, payload.ReleaseScope, payload.RecipientPartitionRefs)
	if err != nil {
		t.Fatal(err)
	}
	payload.RedactionProfileID, payload.RedactionProfileVersion = TokenizedRedactionProfileID, DefaultRedactionProfileVersion
	payload.RedactionProfile, payload.RedactionProfileSHA256, err = ResolveRedactionProfile(payload.RedactionProfileID, payload.RedactionProfileVersion, payload.RecipientPartitionRefs, payload.ExportModel.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	rendered, _, err := fixtureRenderer().Render(context.Background(), payload)
	if err != nil || validateRenderCandidateBinding(payload, rendered) != nil {
		t.Fatal("valid tokenized candidate rejected", err)
	}
	private, _, err := decodePortableObject[RedactionRevealMap](rendered.Redaction.RevealMapJSON, rendered.Redaction.RevealMapSHA256)
	if err != nil || len(private.Entries) == 0 {
		t.Fatal("missing tokenized fixture", err)
	}
	private.Entries[0].CanonicalDisplayValue = "Invented private value"
	private.Entries[0].CanonicalDisplayValueSHA256, err = subjectDisplayValueDigest(private.Entries[0].StableSubjectRef, &private.Entries[0].CanonicalDisplayValue)
	if err != nil {
		t.Fatal(err)
	}
	tokens, _, err := decodePortableObject[RedactionTokenManifest](rendered.Redaction.TokenManifestJSON, rendered.Redaction.TokenManifestSHA256)
	if err != nil {
		t.Fatal(err)
	}
	tokens.Entries[0].CanonicalDisplayValueSHA256 = private.Entries[0].CanonicalDisplayValueSHA256
	rendered.Redaction.TokenManifestJSON, err = canonicalJSON(tokens)
	if err != nil {
		t.Fatal(err)
	}
	rendered.Redaction.TokenManifestSHA256 = reportingObjectDigest(tokens.SchemaID, rendered.Redaction.TokenManifestJSON)
	rendered.RenderBundle.Manifest.TokenManifestSHA256 = &rendered.Redaction.TokenManifestSHA256
	private.TokenManifestSHA256 = rendered.Redaction.TokenManifestSHA256
	rendered.Redaction.RevealMapJSON, err = canonicalJSON(private)
	if err != nil {
		t.Fatal(err)
	}
	rendered.Redaction.RevealMapSHA256 = reportingObjectDigest(private.SchemaID, rendered.Redaction.RevealMapJSON)
	if validateRevealMap(rendered) != nil {
		t.Fatal("fixture must retain valid inter-artifact relationships")
	}
	// Rebind the public member and root too. The rejection below must come
	// from the frozen source, not an accidentally stale outer file digest.
	for i := range rendered.RenderBundle.Files {
		file := &rendered.RenderBundle.Files[i]
		if file.Role == renderBundleRoleTokenManifest {
			file.Bytes = rendered.Redaction.TokenManifestJSON
			file.SizeBytes, file.SHA256 = int64(len(file.Bytes)), hashHex(file.Bytes)
			rendered.RenderBundle.Manifest.Files[i] = file.RenderBundleManifestFile
		}
	}
	rendered.RenderBundle.ManifestJSON, err = canonicalJSON(rendered.RenderBundle.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	rendered.RenderBundle.ManifestSHA256 = reportingObjectDigest(RenderBundleManifestSchemaID, rendered.RenderBundle.ManifestJSON)
	rendered.OutputSHA256 = rendered.RenderBundle.ManifestSHA256
	if err := validateShareableBundle(rendered.RenderBundle); err != nil {
		t.Fatal("tampered fixture must preserve the public artifact graph", err)
	}
	if validateRenderCandidateBinding(payload, rendered) == nil {
		t.Fatal("completion admitted invented private values with recomputed digests")
	}
}
