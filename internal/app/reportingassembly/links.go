package reportingassembly

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidencereporting "github.com/JochiRaider/cartulary/internal/modules/evidence/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

type LinksProvider struct {
	reader links.ActiveFactReader
}

func NewLinksProvider() LinksProvider {
	return LinksProvider{reader: links.ActiveFactReader{}}
}

func (LinksProvider) ProviderKey() string {
	return "links"
}

func (provider LinksProvider) CollectSupportRefsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) (map[string][]string, error) {
	targets, err := evidencereporting.CollectLogicalSupportTargetsTx(ctx, tx, incidentID)
	if err != nil {
		return nil, err
	}
	facts, err := provider.reader.LoadTx(ctx, tx, incidentID)
	if err != nil {
		return nil, err
	}
	result := map[string][]string{}
	for _, fact := range facts.RecordLinks {
		switch fact.LinkType {
		case links.LinkTypeSupportedBy, links.LinkTypeReferencesRecord, links.LinkTypeAttachedEvidence:
		default:
			continue
		}
		target := targets[fact.DstRecordID.String()]
		if target == "" {
			target = "/record_envelopes/" + fact.DstRecordID.String()
		}
		sourceID := fact.SrcRecordID.String()
		result[sourceID] = append(result[sourceID], target)
	}
	return result, nil
}

func (provider LinksProvider) CollectFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
) (exportprovider.ProviderOutput, error) {
	facts, err := provider.reader.LoadTx(ctx, tx, incidentID)
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	fields := make([]exportprovider.FieldFact, 0, len(facts.RecordLinks)+len(facts.RecordTags))
	for _, fact := range facts.RecordLinks {
		fields = append(fields, exportprovider.FieldFact{
			SchemaID:     exportprovider.FieldFactSchemaID,
			Path:         "/relationships/" + fact.RecordLinkID.String(),
			ContentClass: "derived_analytic",
			SourceFamily: "record_link",
			Value: map[string]any{
				"record_link_id":     fact.RecordLinkID.String(),
				"src_record_id":      fact.SrcRecordID.String(),
				"dst_record_id":      fact.DstRecordID.String(),
				"link_type":          fact.LinkType,
				"field_key":          nullableString(fact.FieldKey),
				"provenance":         fact.Provenance,
				"confidence":         nullableJSONNumber(fact.Confidence),
				"owner_user_id":      fact.OwnerUserID.String(),
				"created_by_user_id": fact.CreatedByUserID.String(),
				"decided_at":         postgresJSONTimestamp(fact.DecidedAt),
				"created_at":         postgresJSONTimestamp(fact.CreatedAt),
				"deleted_at":         nil,
				"deleted_by_user_id": nil,
			},
			SupportRefs: exportprovider.CloneStrings(supportRefs[fact.RecordLinkID.String()]),
		})
	}
	for _, fact := range facts.RecordTags {
		fields = append(fields, exportprovider.FieldFact{
			SchemaID:     exportprovider.FieldFactSchemaID,
			Path:         "/tags/" + fact.RecordTagID.String(),
			ContentClass: "derived_analytic",
			SourceFamily: "record_tag",
			Value: map[string]any{
				"record_tag_id":       fact.RecordTagID.String(),
				"record_id":           fact.RecordID.String(),
				"tag_name":            fact.TagName,
				"normalized_tag_name": fact.NormalizedTagName,
				"created_by_user_id":  fact.CreatedByUserID.String(),
				"created_at":          postgresJSONTimestamp(fact.CreatedAt),
				"updated_at":          postgresJSONTimestamp(fact.UpdatedAt),
				"deleted_at":          nil,
				"deleted_by_user_id":  nil,
			},
			SupportRefs: exportprovider.CloneStrings(supportRefs[fact.RecordTagID.String()]),
		})
	}
	return exportprovider.NewProviderOutput(provider.ProviderKey(), fields)
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableJSONNumber(value *int) any {
	if value == nil {
		return nil
	}
	return float64(*value)
}

func postgresJSONTimestamp(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.999999999+00:00")
}
