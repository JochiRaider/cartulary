package reporting

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func TestReportingEvidenceProviderMalformedContributionPublishesNoPartialOutput(t *testing.T) {
	materializer, err := newReportingExportMaterializer()
	if err != nil {
		t.Fatalf("construct Reporting materializer: %v", err)
	}
	materializer.supportRefProvider = staticSupportRefProvider{}
	materializer.fieldProviders = []exportprovider.FieldProvider{
		validEvidenceAdjacentProvider{},
		malformedEvidenceAdjacentProvider{},
	}
	fields, err := materializer.CollectFieldsTx(context.Background(), nil, uuid.New())
	if err == nil || !strings.Contains(err.Error(), "invalid provider output schema") {
		t.Fatalf("malformed provider error = %v, want invalid provider output schema", err)
	}
	if fields != nil {
		t.Fatalf("malformed provider published partial fields: %#v", fields)
	}
}

func TestReportingEvidenceLogicalSupportIdentity(t *testing.T) {
	evidenceID := uuid.NewString()
	ref := "/evidence/" + evidenceID
	if got := supportKindFromRef(ref); got != "evidence_item" {
		t.Fatalf("Evidence support kind = %q, want evidence_item", got)
	}
	if got := supportTargetFromRef(ref); got != "evidence:"+evidenceID {
		t.Fatalf("Evidence support target = %q, want evidence:%s", got, evidenceID)
	}
}

type staticSupportRefProvider struct{}

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
	return exportprovider.NewProviderOutput("evidence", []exportprovider.FieldFact{{
		SchemaID:     exportprovider.FieldFactSchemaID,
		Path:         "/evidence/00000000-0000-0000-0000-000000700001",
		ContentClass: "source_evidence",
		SourceFamily: "evidence",
		Value:        map[string]any{"record_id": "00000000-0000-0000-0000-000000700001"},
	}})
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
