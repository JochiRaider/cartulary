package reporting

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityreporting "github.com/JochiRaider/cartulary/internal/modules/entities/mentions/reportingprovider"
	incidentreporting "github.com/JochiRaider/cartulary/internal/modules/incidents/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	recordreporting "github.com/JochiRaider/cartulary/internal/modules/records/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/sourceboundary"
	timelinereporting "github.com/JochiRaider/cartulary/internal/modules/timeline/reportingprovider"
)

type reportingIncidentProvider interface {
	GetIncidentSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (exportprovider.IncidentSnapshot, error)
}

type reportingExportMaterializer struct {
	incidentProvider   reportingIncidentProvider
	sourceBoundary     SourceBoundaryResolver
	supportRefProvider exportprovider.SupportReferenceProvider
	fieldProviders     []exportprovider.FieldProvider
}

type reportingIncidentProviderFunc struct{}

func (reportingIncidentProviderFunc) GetIncidentSnapshotTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (exportprovider.IncidentSnapshot, error) {
	return incidentreporting.GetIncidentSnapshotTx(ctx, tx, incidentID)
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
	sourceBoundary SourceBoundaryResolver,
	supportRefProvider exportprovider.SupportReferenceProvider,
	contributions ...exportprovider.FieldProvider,
) (reportingExportMaterializer, error) {
	if sourceBoundary == nil {
		return reportingExportMaterializer{}, errors.New("reporting source-boundary resolver is required")
	}
	if supportRefProvider == nil || supportRefProvider.ProviderKey() == "" {
		return reportingExportMaterializer{}, errors.New("reporting support-reference provider is required")
	}
	fieldProviders := []exportprovider.FieldProvider{
		reportingExportFieldProviderFunc{key: "records", collect: recordreporting.CollectFactsTx},
		reportingExportFieldProviderFunc{key: "timeline", collect: timelinereporting.CollectFactsTx},
		parties.NewReportingContribution(),
		entityreporting.New(),
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
		incidentProvider:   reportingIncidentProviderFunc{},
		sourceBoundary:     sourceBoundary,
		supportRefProvider: supportRefProvider,
		fieldProviders:     fieldProviders,
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

type SnapshotSource struct {
	SourceFacts exportprovider.SourceFacts
	Fields      []ExportField
	Timeline    []exportprovider.TimelineSortFact
}

func (m reportingExportMaterializer) CollectSourceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (SnapshotSource, error) {
	supportRefs, err := m.supportRefProvider.CollectSupportRefsTx(ctx, tx, incidentID)
	if err != nil {
		return SnapshotSource{}, fmt.Errorf("collect reporting support provider %s: %w", m.supportRefProvider.ProviderKey(), err)
	}
	source := SnapshotSource{SourceFacts: exportprovider.EmptySourceFacts(), Fields: []ExportField{}, Timeline: []exportprovider.TimelineSortFact{}}
	seenEvents := map[string]bool{}
	seenPaths := map[string]string{}
	for _, provider := range m.fieldProviders {
		output, err := provider.CollectFactsTx(ctx, tx, incidentID, supportRefs)
		if err != nil {
			return SnapshotSource{}, fmt.Errorf("collect reporting export provider %s: %w", provider.ProviderKey(), err)
		}
		if output.ProviderKey != provider.ProviderKey() {
			return SnapshotSource{}, fmt.Errorf("collect reporting export provider %s: output provider key %q", provider.ProviderKey(), output.ProviderKey)
		}
		boundFields, err := output.BoundFields()
		if err != nil {
			return SnapshotSource{}, fmt.Errorf("collect reporting export provider %s: %w", provider.ProviderKey(), err)
		}
		source.SourceFacts.Parties = append(source.SourceFacts.Parties, output.SourceFacts.Parties...)
		source.SourceFacts.Assignments = append(source.SourceFacts.Assignments, output.SourceFacts.Assignments...)
		source.SourceFacts.DeletedRecords = append(source.SourceFacts.DeletedRecords, output.SourceFacts.DeletedRecords...)
		for _, event := range output.TimelineEventFacts {
			if seenEvents[event.TimelineEventID] {
				return SnapshotSource{}, fmt.Errorf("collect reporting export provider %s: overlapping timeline identity", provider.ProviderKey())
			}
			seenEvents[event.TimelineEventID] = true
			source.Timeline = append(source.Timeline, event.Sort.Clone())
		}
		for _, field := range boundFields {
			if field.Path == "" {
				return SnapshotSource{}, fmt.Errorf("collect reporting export provider %s: empty field path", provider.ProviderKey())
			}
			if existingProvider, exists := seenPaths[field.Path]; exists {
				return SnapshotSource{}, fmt.Errorf("collect reporting export provider %s: duplicate field path %s already emitted by %s", provider.ProviderKey(), field.Path, existingProvider)
			}
			seenPaths[field.Path] = provider.ProviderKey()
			projected := exportFieldFromProvider(field.FieldFact)
			projected.SourceRecordID = field.RecordID
			projected.SourceRecordType = field.RecordType
			source.Fields = append(source.Fields, projected)
		}
	}
	sort.Slice(source.Fields, func(i, j int) bool {
		return source.Fields[i].Path < source.Fields[j].Path
	})
	sort.Slice(source.Timeline, func(i, j int) bool { return source.Timeline[i].RecordID < source.Timeline[j].RecordID })
	source.SourceFacts.Sort()
	if err := source.SourceFacts.Validate(); err != nil {
		return SnapshotSource{}, err
	}
	return source, nil
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

func exportFieldFromProvider(field exportprovider.FieldFact) ExportField {
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
