package incidents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/internal/sourcestate"
)

func NewIncidentBundleSourcePort() (sourceport.Port, error) {
	contract, err := newIncidentSourceContract()
	if err != nil {
		return nil, fmt.Errorf("incidents source port: %w", err)
	}
	if contract.source.ContractMajor != sourceport.ContractMajor {
		return nil, fmt.Errorf("incidents source port: %w: contract generation mismatch", sourceport.ErrInvalidCatalog)
	}
	descriptor := sourceport.Descriptor{
		FamilyID:         contract.source.FamilyID,
		ContractMajor:    contract.source.ContractMajor,
		OwnerID:          contract.source.OwnerID,
		OwnerRelationIDs: contract.source.OwnerRelationIDs,
		Paths: []sourceport.Path{{
			LogicalPath:               contract.source.Path.LogicalPath,
			ContentRole:               contract.source.Path.ContentRole,
			SchemaID:                  contract.source.Path.SchemaID,
			Versions:                  contract.source.Path.Versions,
			StableIdentity:            contract.source.Path.StableIdentity,
			StableIdentityInvariantID: contract.source.Path.StableIdentityInvariantID,
		}},
		InvariantIDs: contract.source.InvariantIDs,
	}
	logicalPath := contract.source.Path.LogicalPath
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Export: func(ctx context.Context, exportContext sourceport.ExportContext) error {
			payload, _, err := exportIncidentBundleIncident(ctx, exportContext.Query, exportContext.IncidentID)
			if err != nil {
				return err
			}
			if exportContext.PortableAttributions == nil {
				return sourceport.ErrInvalidCatalog
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(payload, &fields); err != nil {
				return err
			}
			for _, column := range []string{"created_by_user_id", "updated_by_user_id"} {
				var local uuid.UUID
				if err := json.Unmarshal(fields[column], &local); err != nil {
					return err
				}
				historical, err := exportContext.PortableAttributions.ResolvePortableSourceActors(ctx, exportContext.Query, exportContext.IncidentID, "incidents", column, []string{exportContext.IncidentID.String()})
				if err != nil {
					return err
				}
				actor, err := incidentportability.ExportActorID(exportContext.Actors, historical[exportContext.IncidentID.String()], local)
				if err != nil {
					return err
				}
				fields[column], err = json.Marshal(actor)
				if err != nil {
					return err
				}
			}
			payload, err = json.Marshal(fields)
			if err != nil {
				return err
			}
			payload, err = incidentportability.CanonicalRawJSON(payload)
			if err != nil {
				return err
			}
			_, err = exportContext.Members.WriteMember(ctx, logicalPath, bytes.NewReader(payload))
			return err
		},
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			files, err := sourceport.PrepareMembers(ctx, descriptor, bundle, importContext.BundleVersion)
			if err != nil {
				return nil, err
			}
			payload, err := incidentportability.ReadControl(ctx, files, logicalPath, incidentportability.MaxControlMemberBytes)
			if err != nil {
				return nil, err
			}
			prepared, err := prepareIncidentBundleIncident(contract, payload, incidentSourcePortImportContext(importContext))
			return prepared, incidentSourcePortError(descriptor, err)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			err := applyIncidentBundleIncidentTx(ctx, tx, value, incidentSourcePortImportContext(importContext), contract)
			return incidentSourcePortError(descriptor, err)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			err := validateIncidentBundleIncidentTx(ctx, tx, value, incidentSourcePortImportContext(importContext), contract)
			return incidentSourcePortError(descriptor, err)
		},
		ValidateContract: sourcestate.Validate,
	}), nil
}

func incidentSourcePortImportContext(importContext sourceport.ImportContext) incidentSourceImportContext {
	return incidentSourceImportContext{
		incidentID:    importContext.IncidentID,
		actorUserID:   importContext.ActorUserID,
		bundleVersion: importContext.BundleVersion,
		operationID:   importContext.OperationID,
		attributions:  importContext.Attributions,
		actorAdmitted: func(sourceActorID string) bool {
			_, admitted := importContext.Actors.Lookup(sourceActorID)
			return admitted
		},
	}
}

func incidentSourcePortError(descriptor sourceport.Descriptor, err error) error {
	if err == nil {
		return nil
	}
	var invariantFailure *incidentSourceInvariantFailure
	if errors.As(err, &invariantFailure) {
		return descriptor.DeclaredFailure(invariantFailure.invariantID)
	}
	if errors.Is(err, errIncidentSourcePreparedBinding) {
		return fmt.Errorf("%w: %v", sourceport.ErrPreparedBinding, err)
	}
	if errors.Is(err, errIncidentSourceCatalog) {
		return fmt.Errorf("%w: %v", sourceport.ErrInvalidCatalog, err)
	}
	return err
}
