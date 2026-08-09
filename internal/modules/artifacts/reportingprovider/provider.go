package reportingprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

type Provider struct {
	derived artifactprojection.Reader
}

func New(derived artifactprojection.Reader) (*Provider, error) {
	if derived == nil {
		return nil, fmt.Errorf("compose Artifact reporting provider: projection derived-fact reader is required")
	}
	return &Provider{derived: derived}, nil
}

func (*Provider) ProviderKey() string { return "artifacts" }

func (provider *Provider) CollectFieldsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
) ([]exportprovider.Field, error) {
	output, err := provider.CollectFactsTx(ctx, tx, incidentID, supportRefs)
	if err != nil {
		return nil, err
	}
	return output.Fields(), nil
}

func (provider *Provider) CollectFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
) (exportprovider.ProviderOutput, error) {
	if provider == nil || provider.derived == nil {
		return exportprovider.ProviderOutput{}, fmt.Errorf("collect Artifact reporting facts: projection derived-fact reader is required")
	}
	facts, err := provider.derived.CollectDerivedFactsTx(ctx, tx, incidentID)
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	fields := make([]exportprovider.FieldFact, 0, len(facts))
	for _, fact := range facts {
		prefix, sourceFamily, contentClass, included := reportingSemantics(fact)
		if !included {
			continue
		}
		recordID := fact.RecordID.String()
		fields = append(fields, exportprovider.FieldFact{
			SchemaID:     exportprovider.FieldFactSchemaID,
			Path:         "/" + prefix + "/" + recordID,
			ContentClass: contentClass,
			SourceFamily: sourceFamily,
			Value:        fact.Value,
			SupportRefs:  exportprovider.CloneStrings(supportRefs[recordID]),
		})
	}
	return exportprovider.NewProviderOutput(provider.ProviderKey(), fields)
}

func reportingSemantics(fact artifactprojection.DerivedFact) (string, string, string, bool) {
	switch fact.ArtifactType {
	case "note":
		return "notes", "note", "working_material", true
	case "finding":
		contentClass := "working_material"
		if fact.FindingKind != nil && *fact.FindingKind == "finding" {
			contentClass = "curated_narrative"
		}
		return "findings", "finding_hypothesis", contentClass, true
	case "comm_log":
		return "comm_log", "comm_log", "working_material", true
	case "handoff":
		return "handoffs", "handoff", "working_material", true
	case "status_review":
		return "status_reviews", "status_review", "working_material", true
	case "lesson":
		return "lessons", "lesson", "working_material", true
	default:
		return "", "", "", false
	}
}
