package reportingprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

type Provider struct {
	derived entityprojection.Reader
}

func New(derived entityprojection.Reader) (*Provider, error) {
	if derived == nil {
		return nil, fmt.Errorf("compose Entities reporting provider: projection derived-fact reader is required")
	}
	return &Provider{derived: derived}, nil
}

func (*Provider) ProviderKey() string { return "entities.hostidentity" }

func (provider *Provider) CollectFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) ([]exportprovider.Field, error) {
	output, err := provider.CollectFactsTx(ctx, tx, incidentID, supportRefs)
	if err != nil {
		return nil, err
	}
	return output.Fields(), nil
}

func (provider *Provider) CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
	if provider == nil || provider.derived == nil {
		return exportprovider.ProviderOutput{}, fmt.Errorf("collect Entities reporting facts: projection derived-fact reader is required")
	}
	hostFacts, err := provider.derived.CollectHostDerivedFactsTx(ctx, tx, incidentID)
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	identityFacts, err := provider.derived.CollectIdentityDerivedFactsTx(ctx, tx, incidentID)
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	fields := make([]exportprovider.FieldFact, 0, len(hostFacts)+len(identityFacts))
	fields = appendDerivedFacts(fields, "hosts", hostFacts, supportRefs)
	fields = appendDerivedFacts(fields, "identities", identityFacts, supportRefs)
	return exportprovider.NewProviderOutput(provider.ProviderKey(), fields)
}

func appendDerivedFacts(
	fields []exportprovider.FieldFact,
	prefix string,
	facts []entityprojection.DerivedFact,
	supportRefs map[string][]string,
) []exportprovider.FieldFact {
	for _, fact := range facts {
		recordID := fact.RecordID.String()
		fields = append(fields, exportprovider.FieldFact{
			SchemaID:     exportprovider.FieldFactSchemaID,
			Path:         "/" + prefix + "/" + recordID,
			ContentClass: fact.ContentClass,
			SourceFamily: fact.RecordType,
			Value:        fact.Value,
			SupportRefs:  exportprovider.CloneStrings(supportRefs[recordID]),
		})
	}
	return fields
}
