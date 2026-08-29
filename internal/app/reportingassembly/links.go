package reportingassembly

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

type LinksProvider struct {
	reader                 links.FactReader
	logicalTargetProviders []exportprovider.LogicalSupportTargetProvider
}

func NewLinksProvider(logicalTargetProviders ...exportprovider.LogicalSupportTargetProvider) (LinksProvider, error) {
	providers := append([]exportprovider.LogicalSupportTargetProvider(nil), logicalTargetProviders...)
	seen := make(map[string]struct{}, len(providers))
	for _, provider := range providers {
		if nilLogicalSupportTargetProvider(provider) {
			return LinksProvider{}, errors.New("reporting logical support target provider is required")
		}
		key := provider.ProviderKey()
		if key == "" {
			return LinksProvider{}, errors.New("reporting logical support target provider key is required")
		}
		if _, exists := seen[key]; exists {
			return LinksProvider{}, fmt.Errorf("duplicate Reporting logical support target provider %q", key)
		}
		seen[key] = struct{}{}
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].ProviderKey() < providers[j].ProviderKey()
	})
	return LinksProvider{
		reader:                 links.FactReader{},
		logicalTargetProviders: providers,
	}, nil
}

func nilLogicalSupportTargetProvider(provider exportprovider.LogicalSupportTargetProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (LinksProvider) ProviderKey() string {
	return "links"
}

func (provider LinksProvider) CollectSupportRefsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) (map[string][]string, error) {
	targets, err := provider.collectLogicalSupportTargetsTx(ctx, tx, incidentID)
	if err != nil {
		return nil, err
	}
	facts, err := provider.reader.LoadIncidentTx(ctx, tx, incidentID)
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

func (provider LinksProvider) collectLogicalSupportTargetsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) (map[string]string, error) {
	targets := map[string]string{}
	targetOwners := map[string]string{}
	for _, targetProvider := range provider.logicalTargetProviders {
		contribution, err := targetProvider.CollectLogicalSupportTargetsTx(ctx, tx, incidentID)
		if err != nil {
			return nil, fmt.Errorf("collect Reporting logical support targets from %s: %w", targetProvider.ProviderKey(), err)
		}
		keys := make([]string, 0, len(contribution))
		for recordID := range contribution {
			keys = append(keys, recordID)
		}
		sort.Strings(keys)
		for _, recordID := range keys {
			target := contribution[recordID]
			if existing, exists := targets[recordID]; exists && existing != target {
				return nil, fmt.Errorf(
					"conflicting Reporting logical support target %q from providers %s and %s",
					recordID,
					targetOwners[recordID],
					targetProvider.ProviderKey(),
				)
			}
			targets[recordID] = target
			targetOwners[recordID] = targetProvider.ProviderKey()
		}
	}
	return targets, nil
}

func (provider LinksProvider) CollectFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	supportRefs map[string][]string,
) (exportprovider.ProviderOutput, error) {
	facts, err := provider.reader.LoadIncidentTx(ctx, tx, incidentID)
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
