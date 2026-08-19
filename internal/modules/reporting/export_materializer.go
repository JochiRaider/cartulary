package reporting

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityreporting "github.com/JochiRaider/cartulary/internal/modules/entities/mentions/reportingprovider"
	evidencereporting "github.com/JochiRaider/cartulary/internal/modules/evidence/reportingprovider"
	incidentreporting "github.com/JochiRaider/cartulary/internal/modules/incidents/reportingprovider"
	linkreporting "github.com/JochiRaider/cartulary/internal/modules/links/reportingprovider"
	partyreporting "github.com/JochiRaider/cartulary/internal/modules/parties/reportingprovider"
	recordreporting "github.com/JochiRaider/cartulary/internal/modules/records/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/sourceboundary"
	timelinereporting "github.com/JochiRaider/cartulary/internal/modules/timeline/reportingprovider"
)

type reportingIncidentProvider interface {
	GetIncidentSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (exportprovider.IncidentSnapshot, error)
}

type reportingSupportRefProvider interface {
	CollectSupportRefsTx(context.Context, pgx.Tx, uuid.UUID) (map[string][]string, error)
}

type reportingExportMaterializer struct {
	incidentProvider   reportingIncidentProvider
	sourceBoundary     sourceboundary.Resolver
	supportRefProvider reportingSupportRefProvider
	fieldProviders     []exportprovider.FieldProvider
}

type reportingIncidentProviderFunc struct{}

func (reportingIncidentProviderFunc) GetIncidentSnapshotTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (exportprovider.IncidentSnapshot, error) {
	return incidentreporting.GetIncidentSnapshotTx(ctx, tx, incidentID)
}

type reportingSupportRefProviderFunc struct {
	collect func(context.Context, pgx.Tx, uuid.UUID) (map[string][]string, error)
}

func (p reportingSupportRefProviderFunc) CollectSupportRefsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (map[string][]string, error) {
	return p.collect(ctx, tx, incidentID)
}

type reportingExportFieldProviderFunc struct {
	key     string
	collect func(context.Context, pgx.Tx, uuid.UUID, map[string][]string) (exportprovider.ProviderOutput, error)
}

func (p reportingExportFieldProviderFunc) ProviderKey() string {
	return p.key
}

func (p reportingExportFieldProviderFunc) CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
	return p.collect(ctx, tx, incidentID, supportRefs)
}

func newReportingExportMaterializer(
	sourceBoundary sourceboundary.Resolver,
	contributions ...exportprovider.FieldProvider,
) (reportingExportMaterializer, error) {
	if sourceBoundary == nil {
		return reportingExportMaterializer{}, errors.New("reporting source-boundary resolver is required")
	}
	fieldProviders := []exportprovider.FieldProvider{
		reportingExportFieldProviderFunc{key: "records", collect: recordreporting.CollectFactsTx},
		reportingExportFieldProviderFunc{key: "timeline", collect: timelinereporting.CollectFactsTx},
		reportingExportFieldProviderFunc{key: "parties", collect: partyreporting.CollectFactsTx},
		reportingExportFieldProviderFunc{key: "evidence", collect: evidencereporting.CollectFactsTx},
		reportingExportFieldProviderFunc{key: "links", collect: linkreporting.CollectFactsTx},
		reportingExportFieldProviderFunc{key: "entities.mentions", collect: entityreporting.CollectFactsTx},
	}
	fieldProviders = append(fieldProviders, contributions...)
	seen := make(map[string]struct{}, len(fieldProviders))
	for _, provider := range fieldProviders {
		if provider == nil {
			return reportingExportMaterializer{}, errors.New("reporting export field provider is required")
		}
		key := provider.ProviderKey()
		if key == "" {
			return reportingExportMaterializer{}, errors.New("reporting export field provider key is required")
		}
		if _, exists := seen[key]; exists {
			return reportingExportMaterializer{}, fmt.Errorf("duplicate reporting export field provider %q", key)
		}
		seen[key] = struct{}{}
	}
	sort.Slice(fieldProviders, func(i, j int) bool {
		return fieldProviders[i].ProviderKey() < fieldProviders[j].ProviderKey()
	})
	return reportingExportMaterializer{
		incidentProvider: reportingIncidentProviderFunc{},
		sourceBoundary:   sourceBoundary,
		supportRefProvider: reportingSupportRefProviderFunc{
			collect: func(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (map[string][]string, error) {
				targets, err := evidencereporting.CollectLogicalSupportTargetsTx(ctx, tx, incidentID)
				if err != nil {
					return nil, err
				}
				return linkreporting.CollectSupportRefsTx(ctx, tx, incidentID, targets)
			},
		},
		fieldProviders: fieldProviders,
	}, nil
}

func getIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (IncidentMetadataSnapshot, error) {
	record, err := (reportingIncidentProviderFunc{}).GetIncidentSnapshotTx(ctx, tx, incidentID)
	if err != nil {
		if errors.Is(err, exportprovider.ErrNotFound) {
			return IncidentMetadataSnapshot{}, ErrNotFound
		}
		return IncidentMetadataSnapshot{}, err
	}
	return incidentMetadataFromProvider(record), nil
}

func (m reportingExportMaterializer) ResolveSourceBoundaryTx(
	ctx context.Context,
	tx pgx.Tx,
	incident IncidentMetadataSnapshot,
) (ResolvedSourceBoundary, error) {
	incidentID, err := uuid.Parse(incident.ID)
	if err != nil {
		return ResolvedSourceBoundary{}, err
	}
	boundary, err := m.sourceBoundary.ResolveCurrentTx(ctx, tx, sourceboundary.ResolveInput{
		IncidentID:      incidentID,
		IncidentVersion: incident.Version,
	})
	if err != nil {
		return ResolvedSourceBoundary{}, err
	}
	return ResolvedSourceBoundary{
		Token:         boundary.Token,
		CanonicalJSON: append([]byte(nil), boundary.CanonicalJSON...),
	}, nil
}

func (m reportingExportMaterializer) CollectFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]ExportField, error) {
	supportRefs, err := m.supportRefProvider.CollectSupportRefsTx(ctx, tx, incidentID)
	if err != nil {
		return nil, err
	}
	fields := []ExportField{}
	seenPaths := map[string]string{}
	for _, provider := range m.fieldProviders {
		output, err := provider.CollectFactsTx(ctx, tx, incidentID, supportRefs)
		if err != nil {
			return nil, fmt.Errorf("collect reporting export provider %s: %w", provider.ProviderKey(), err)
		}
		if output.ProviderKey != provider.ProviderKey() {
			return nil, fmt.Errorf("collect reporting export provider %s: output provider key %q", provider.ProviderKey(), output.ProviderKey)
		}
		if err := output.Validate(); err != nil {
			return nil, fmt.Errorf("collect reporting export provider %s: %w", provider.ProviderKey(), err)
		}
		for _, field := range output.Fields() {
			if field.Path == "" {
				return nil, fmt.Errorf("collect reporting export provider %s: empty field path", provider.ProviderKey())
			}
			if existingProvider, exists := seenPaths[field.Path]; exists {
				return nil, fmt.Errorf("collect reporting export provider %s: duplicate field path %s already emitted by %s", provider.ProviderKey(), field.Path, existingProvider)
			}
			seenPaths[field.Path] = provider.ProviderKey()
			fields = append(fields, exportFieldFromProvider(field))
		}
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	return fields, nil
}

func incidentMetadataFromProvider(record exportprovider.IncidentSnapshot) IncidentMetadataSnapshot {
	return IncidentMetadataSnapshot{
		ID:           record.ID,
		Title:        record.Title,
		Description:  record.Description,
		Status:       record.Status,
		Severity:     record.Severity,
		TLP:          record.TLP,
		CurrentPhase: record.CurrentPhase,
		Version:      record.Version,
	}
}

func exportFieldFromProvider(field exportprovider.Field) ExportField {
	return ExportField{
		Path:                    field.Path,
		ContentClass:            field.ContentClass,
		SourceFamily:            field.SourceFamily,
		Value:                   field.Value,
		DisclosurePartitionRefs: exportprovider.CloneStrings(field.DisclosurePartitionRefs),
		SupportRefs:             exportprovider.CloneStrings(field.SupportRefs),
		RawBlobSource:           field.RawBlobSource,
		OpaqueBinary:            field.OpaqueBinary,
		GeneratedPresentation:   field.GeneratedPresentation,
	}
}
