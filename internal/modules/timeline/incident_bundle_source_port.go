package timeline

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	binding := &timelinePreparationBinding{contractMajor: sourceport.ContractMajor}
	descriptor := sourceport.Descriptor{
		FamilyID: "timeline", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.timeline", OwnerRelationIDs: []string{"timeline-source"},
		Dependencies: []string{"records"},
		Paths: []sourceport.Path{
			{LogicalPath: timelineBundleProfilesPath, ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"incident_id"}, StableIdentityInvariantID: "timeline.source_identity_admitted"},
			{LogicalPath: timelineBundleRecordsPath, ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "timeline.source_identity_admitted"},
			{LogicalPath: timelineBundleProvenancePath, ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"record_id", "source_row_ordinal", "source_column_ordinal", "source_identity_sha256"}, StableIdentityInvariantID: "timeline.source_identity_admitted"},
		},
		InvariantIDs: []string{
			"timeline.version_shape_exact", "timeline.envelope_type_scope",
			"timeline.lifecycle_coherent", "timeline.generated_time_coherent",
			"timeline.paired_time_coherent", "timeline.provenance_unique",
			"timeline.provenance_non_orphaned",
			"timeline.source_identity_admitted",
			"timeline.profile_coherent", "timeline.profile_attribution_exact", "timeline.profile_state_exact",
			"timeline.event_attribution_exact", "timeline.event_state_exact",
			"timeline.provenance_coherent", "timeline.provenance_state_exact",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: ExportIncidentBundleFiles,
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareTimelineImport(ctx, bundle, importContext, descriptor, binding)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedTimelineImport)
			if !ok || !prepared.matches(importContext, binding) || tx == nil {
				return sourceport.ErrPreparedBinding
			}
			// Apply reads only the staged preparation streams (Core 01
			// REQ-01-641); attribution is flushed by the core in step 9.
			store := importContext.Preparation
			if err := prepared.preflightEvents(ctx, tx, store, descriptor); err != nil {
				return err
			}
			if err := prepared.applyProfile(ctx, tx, descriptor); err != nil {
				return err
			}
			if err := prepared.applyEvents(ctx, tx, store, descriptor); err != nil {
				return err
			}
			return prepared.applyProvenance(ctx, tx, store, descriptor)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedTimelineImport)
			if !ok || !prepared.matches(importContext, binding) || tx == nil {
				return sourceport.ErrPreparedBinding
			}
			store := importContext.Preparation
			if err := prepared.validateProfile(ctx, tx, descriptor); err != nil {
				return err
			}
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM timeline_events event
      LEFT JOIN records record ON record.record_id = event.record_id
     WHERE event.incident_id = $1
       AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> 'timeline_event')
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return descriptor.DeclaredFailure("timeline.envelope_type_scope")
			}
			if err := prepared.validateEvents(ctx, tx, store, descriptor); err != nil {
				return err
			}
			return prepared.validateProvenance(ctx, tx, store, descriptor)
		},
		// Attribution was staged while preparing; the core flushes it in step 9.
		Attributions: func(value any) []incidentportability.AttributionStream {
			prepared, ok := value.(preparedTimelineImport)
			if !ok {
				return nil
			}
			return []incidentportability.AttributionStream{prepared.attribution}
		},
	})
}
