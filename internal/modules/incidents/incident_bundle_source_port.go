package incidents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "incident", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.incidents", OwnerRelationIDs: []string{"incident-core"},
		Paths: []sourceport.Path{{
			LogicalPath: "data/incident.json", ContentRole: "singleton_json",
			Versions: []int{3}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "incident.source_identity_admitted",
		}},
		InvariantIDs: []string{
			"incident.exact_shape", "incident.identity_key_lifecycle",
			"incident.attribution_version", "incident.source_identity_admitted",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export: func(ctx context.Context, exportContext sourceport.ExportContext) ([]incidentportability.File, error) {
			payload, _, err := exportIncidentBundleIncident(ctx, exportContext.Query, exportContext.IncidentID)
			if err != nil {
				return nil, err
			}
			return []incidentportability.File{{Path: "data/incident.json", Payload: payload}}, nil
		},
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			files, err := sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
			if err != nil {
				return nil, err
			}
			var row map[string]any
			if err := json.Unmarshal(files["data/incident.json"], &row); err != nil {
				return nil, descriptor.DeclaredFailure("incident.exact_shape")
			}
			if incidentportability.StringFromAny(row["id"]) != importContext.IncidentID.String() {
				return nil, descriptor.DeclaredFailure("incident.identity_key_lifecycle")
			}
			return files, nil
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			files, ok := value.(sourceport.PreparedFiles)
			if !ok {
				return fmt.Errorf("%w: Incidents prepared import has type %T", sourceport.ErrInvalidCatalog, value)
			}
			return importIncidentBundleIncidentTx(ctx, tx, files["data/incident.json"], importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE id = $1`, importContext.IncidentID).Scan(&count); err != nil {
				return err
			}
			if count != 1 {
				return descriptor.DeclaredFailure("incident.identity_key_lifecycle")
			}
			return nil
		},
	})
}
