package reportingprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityports "github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

type provider struct {
	derived entityports.ReportingReader
}

func New(derived entityports.ReportingReader) (exportprovider.FieldProvider, error) {
	if derived == nil {
		return nil, fmt.Errorf("compose Entities reporting provider: projection derived-fact reader is required")
	}
	return &provider{derived: derived}, nil
}

func (*provider) ProviderKey() string { return "entities.hostidentity" }

func (provider *provider) CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
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
	fields := make([]exportprovider.BoundField, 0, len(hostFacts)+len(identityFacts))
	fields = appendDerivedFacts(fields, "hosts", hostFacts, supportRefs)
	fields = appendDerivedFacts(fields, "identities", identityFacts, supportRefs)
	return exportprovider.NewProviderOutput(provider.ProviderKey(), fields, nil)
}

func appendDerivedFacts(
	fields []exportprovider.BoundField,
	prefix string,
	facts []entityports.DerivedFact,
	supportRefs map[string][]string,
) []exportprovider.BoundField {
	for _, fact := range facts {
		recordID := fact.RecordID.String()
		fields = append(fields, exportprovider.BoundField{
			RecordID: recordID, RecordType: fact.RecordType,
			FieldFact: exportprovider.FieldFact{
				SchemaID:     exportprovider.FieldFactSchemaID,
				Path:         "/" + prefix + "/" + recordID,
				ContentClass: fact.ContentClass,
				SourceFamily: fact.RecordType,
				Value:        fact.Value,
				SupportRefs:  exportprovider.CloneStrings(supportRefs[recordID]),
			}})
	}
	return fields
}
