package reporting

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/sourceboundary"
)

func TestReportingEvidenceProviderMalformedContributionPublishesNoPartialOutput(t *testing.T) {
	materializer, err := newReportingExportMaterializer(sourceboundary.NewResolver(), staticSupportRefProvider{})
	if err != nil {
		t.Fatalf("construct Reporting materializer: %v", err)
	}
	materializer.supportRefProvider = staticSupportRefProvider{}
	materializer.fieldProviders = []exportprovider.FieldProvider{
		validEvidenceAdjacentProvider{},
		malformedEvidenceAdjacentProvider{},
	}
	source, err := materializer.CollectSourceTx(context.Background(), nil, uuid.New())
	if err == nil || !strings.Contains(err.Error(), "invalid provider output schema") {
		t.Fatalf("malformed provider error = %v, want invalid provider output schema", err)
	}
	if len(source.Fields) != 0 || len(source.Timeline) != 0 {
		t.Fatalf("malformed provider published partial fields: %#v", source)
	}
}

func TestReportingEvidenceLogicalSupportIdentity(t *testing.T) {
	evidenceID := uuid.NewString()
	ref := "/evidence/" + evidenceID
	if got, err := resolveSupportSelector(ref, "incident"); err != nil || got.ref != "evidence:"+evidenceID || got.kind != "evidence_item" {
		t.Fatalf("Evidence support target = %#v, error=%v, want evidence:%s", got, err, evidenceID)
	}
	for source, target := range map[string]string{
		"/incident/status":           "record:incident_metadata:incident",
		"/record_envelopes/record-1": "record:record_envelope:record-1",
		"/hosts/host-1":              "record:host:host-1",
		"/timeline/event-1":          "timeline:event-1",
		"/relationships/link-1":      "relationship:link-1",
	} {
		got, err := resolveSupportSelector(source, "incident")
		if err != nil || got.ref != target || !validateSupportTarget(got.kind, got.ref) {
			t.Fatalf("logical support mapping %q = %#v, error=%v", source, got, err)
		}
	}
	for _, ref := range []string{"", "evidence/a", "/evidence/", "/evidence/a/", "//evidence/a", "/evidence/a/b", "/post/a", "/unowned/a", "/evidence/ a", "/evidence/a ", "/evidence/a#fragment", "/evidence/a\\b", "/evidence/a\x00", "/evidence/a\u0085", "/evidence/" + strings.Repeat("x", 129)} {
		t.Run("invalid selector "+strings.ReplaceAll(ref, "\x00", "NUL"), func(t *testing.T) {
			got, err := resolveSupportSelector(ref, "incident")
			var invalid *renderValidationError
			if got != (supportTarget{}) || !errors.As(err, &invalid) || invalid.FailureCode != "export_model_invalid" || invalid.ReasonCode != "invalid_support_ref" {
				t.Fatal("invalid support selector acquired an identity", got, err)
			}
			source := fixtureSnapshotSource(fixtureOwnerFields([]ExportField{{Path: "/hosts/source", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: "source", SupportRefs: []string{ref}}}))
			model, digest, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "boundary", source)
			if !errors.As(err, &invalid) || invalid.ReasonCode != "invalid_support_ref" || digest != "" || model.SnapshotID != "" {
				t.Fatal("malformed support published partial snapshot material", err)
			}
		})
	}
	sourceFields := []ExportField{
		{Path: "/hosts/source", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: "source", SupportRefs: []string{"/evidence/metadata"}},
		{Path: "/evidence/metadata", SourceFamily: "evidence", ContentClass: ContentClassSourceEvidence, Value: "metadata"},
	}
	model := fixtureStructuredRenderModel(t, sourceFields)
	t.Run("frozen inventory closure", func(t *testing.T) { testSnapshotSupportClosure(t, model, sourceFields) })
	for _, target := range []string{"record:unknown:0123456789abcdef", "evidence:", "record:post:p", "evidence:storage/blob", "timeline:wrong-kind"} {
		candidate := model
		candidate.SupportIndex = append([]ReportingSupportRef(nil), model.SupportIndex...)
		candidate.SupportIndex[0].SupportTargetRef = target
		var invalid *renderValidationError
		if err := validateRenderModelIdentity(candidate); !errors.As(err, &invalid) || invalid.ReasonCode != "invalid_support_ref" {
			t.Fatal("retained malformed support target was accepted", err)
		}
	}
	incidentModel := fixtureStructuredRenderModel(t, []ExportField{{Path: "/hosts/source", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: "source", SupportRefs: []string{"/incident/status"}}})
	if err := validateRenderModelIdentity(incidentModel); err != nil {
		t.Fatal("support for the model's incident was rejected", err)
	}
	for _, wrongID := range []string{"status", "another-incident"} {
		candidate := incidentModel
		candidate.SupportIndex = append([]ReportingSupportRef(nil), incidentModel.SupportIndex...)
		candidate.SupportIndex[0].SupportTargetRef = "record:incident_metadata:" + wrongID
		var invalid *renderValidationError
		if err := validateRenderModelIdentity(candidate); !errors.As(err, &invalid) || invalid.ReasonCode != "invalid_support_ref" {
			t.Fatal("retained support targeted a different incident", err)
		}
	}
}

type staticSupportRefProvider struct{}

func (staticSupportRefProvider) ProviderKey() string { return "links" }

func (staticSupportRefProvider) CollectSupportRefsTx(context.Context, pgx.Tx, uuid.UUID) (map[string][]string, error) {
	return nil, nil
}

type validEvidenceAdjacentProvider struct{}

func (validEvidenceAdjacentProvider) ProviderKey() string { return "evidence" }

func (validEvidenceAdjacentProvider) CollectFactsTx(
	context.Context,
	pgx.Tx,
	uuid.UUID,
	map[string][]string,
) (exportprovider.ProviderOutput, error) {
	return exportprovider.NewProviderOutput("evidence", []exportprovider.BoundField{{
		RecordID: "00000000-0000-0000-0000-000000700001", RecordType: "evidence",
		FieldFact: exportprovider.FieldFact{
			SchemaID:     exportprovider.FieldFactSchemaID,
			Path:         "/evidence/00000000-0000-0000-0000-000000700001",
			ContentClass: "source_evidence",
			SourceFamily: "evidence",
			Value:        map[string]any{"record_id": "00000000-0000-0000-0000-000000700001"},
		}}}, nil)
}

type malformedEvidenceAdjacentProvider struct{}

func (malformedEvidenceAdjacentProvider) ProviderKey() string { return "zz.evidence_malformed" }

func (malformedEvidenceAdjacentProvider) CollectFactsTx(
	context.Context,
	pgx.Tx,
	uuid.UUID,
	map[string][]string,
) (exportprovider.ProviderOutput, error) {
	return exportprovider.ProviderOutput{
		SchemaID:    "cartulary.invalid_reporting_owner_provider_output.v1",
		ProviderKey: "zz.evidence_malformed",
	}, nil
}
