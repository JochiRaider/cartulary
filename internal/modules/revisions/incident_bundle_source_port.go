package revisions

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func NewIncidentBundleSourcePort(
	envelopes IncidentBundleRecordEnvelopeReader,
	snapshots *RecordSnapshotCaptureCatalog,
	targets *TargetSemanticsCatalog,
) (sourceport.Port, error) {
	validation, err := newIncidentBundleValidationCatalog(envelopes, snapshots, targets)
	if err != nil {
		return nil, err
	}
	descriptor := revisionsIncidentBundleDescriptor()
	port := sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		ValidateContract: func() error {
			return validation.validateContract()
		},
		Export: func(ctx context.Context, exportContext sourceport.ExportContext) ([]incidentportability.File, error) {
			return exportIncidentBundleFiles(ctx, exportContext)
		},
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareRevisionsImport(bundle, importContext)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedRevisionsImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return applyPreparedRevisionsImportTx(ctx, tx, prepared, importContext, validation)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, _ sourceport.ImportContext) error {
			prepared, ok := value.(preparedRevisionsImport)
			if !ok {
				return sourceport.ErrPreparedBinding
			}
			return validatePreparedRevisionsImportTx(ctx, tx, prepared, validation)
		},
	})
	if err := port.ValidateSourcePortContract(); err != nil {
		return nil, err
	}
	return port, nil
}

func revisionsIncidentBundleDescriptor() sourceport.Descriptor {
	return sourceport.Descriptor{
		FamilyID: "revisions", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.revisions", OwnerRelationIDs: []string{"record-revisions"},
		Dependencies: []string{"links_tags"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/change_sets.ndjson", ContentRole: "source_rows", SchemaID: "cartulary.incident_bundle.change_sets.row.v1", Versions: []int{3}, StableIdentity: []string{"change_set_id"}, StableIdentityInvariantID: "revisions.source_identity_admitted"},
			{LogicalPath: "data/change_set_mutations.ndjson", ContentRole: "source_rows", SchemaID: "cartulary.incident_bundle.change_set_mutations.row.v1", Versions: []int{3}, StableIdentity: []string{"change_set_id", "sequence_no"}, StableIdentityInvariantID: "revisions.source_identity_admitted"},
			{LogicalPath: "data/record_revisions.ndjson", ContentRole: "source_rows", SchemaID: "cartulary.incident_bundle.record_revisions.row.v1", Versions: []int{3}, StableIdentity: []string{"revision_id"}, StableIdentityInvariantID: "revisions.source_identity_admitted"},
		},
		InvariantIDs: []string{
			"revisions.references_complete", "revisions.actor_references_complete",
			"revisions.mutation_sequence_contiguous", "revisions.record_version_unique",
			"revisions.history_reconstruction", "revisions.sequence_repair_after_validation",
			"revisions.source_identity_admitted",
		},
	}
}
