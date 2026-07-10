package imports

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const extensionApplyFacadesOverrideKey = "imports.extension_apply_facades"

type ExtensionImportApplyRequest struct {
	IncidentID                  uuid.UUID
	ActorUserID                 uuid.UUID
	TargetKind                  string
	ExtensionProfileID          string
	ImportSessionID             uuid.UUID
	ImportUnitID                uuid.UUID
	SourceCapability            ImportSourceCapability
	ExpectedSourceContentSHA256 string
	MappingFingerprint          string
	OwnerMappingSchemaID        string
	OwnerMapping                json.RawMessage
	ClientTxnID                 string
}

type ExtensionImportApplyResult struct {
	ResourceRefs  []jobs.ResourceRef
	OwnerResponse map[string]any
}

type ExtensionImportApplyFacade interface {
	ApplyImportUnitTx(context.Context, pgx.Tx, ExtensionImportApplyRequest) (ExtensionImportApplyResult, error)
}

func extensionApplyFacadesFromDependencies(deps httpapi.DependencySet) (map[string]ExtensionImportApplyFacade, error) {
	facades := map[string]ExtensionImportApplyFacade{}
	override, ok := deps.ModuleOverrides[extensionApplyFacadesOverrideKey]
	if !ok || override == nil {
		return facades, nil
	}
	typed, ok := override.(map[string]ExtensionImportApplyFacade)
	if !ok {
		return nil, fmt.Errorf("imports extension apply facades override has type %T", override)
	}
	for key, facade := range typed {
		if facade != nil {
			facades[key] = facade
		}
	}
	return facades, nil
}

func extensionImportFacadeKey(target importTarget) string {
	return target.TargetKind + ":" + target.ExtensionProfileID
}

func (s *Service) applyExtensionOwnerUnit(ctx context.Context, actor authn.UserRecord, start ApplyStartResult, unit ApplyUnitData, target importTarget) ([]jobs.ResourceRef, error) {
	facade := s.extensionApplyFacades[extensionImportFacadeKey(target)]
	if facade == nil {
		return nil, importApplyBlockedError("owner_apply_contract_unavailable")
	}
	tx, err := s.store.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, start.IncidentID); err != nil {
		return nil, err
	}
	sourceCapability, err := s.store.sourceCapabilityForUnitTx(ctx, tx, start.ImportSessionID, unit.UnitID)
	if err != nil {
		return nil, err
	}
	if sourceCapability.SourceContentSHA256 != unit.SourceContentSHA256 {
		return nil, importApplyBlockedError("source_changed")
	}
	result, err := facade.ApplyImportUnitTx(ctx, tx, ExtensionImportApplyRequest{
		IncidentID:                  start.IncidentID,
		ActorUserID:                 actor.ID,
		TargetKind:                  target.TargetKind,
		ExtensionProfileID:          target.ExtensionProfileID,
		ImportSessionID:             start.ImportSessionID,
		ImportUnitID:                unit.UnitID,
		SourceCapability:            sourceCapability,
		ExpectedSourceContentSHA256: unit.SourceContentSHA256,
		MappingFingerprint:          unit.MappingFingerprint,
		OwnerMappingSchemaID:        unit.ApprovedMapping.OwnerMappingSchemaID,
		OwnerMapping:                append(json.RawMessage(nil), unit.ApprovedMapping.OwnerMapping...),
		ClientTxnID:                 start.ClientTxnID,
	})
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return append([]jobs.ResourceRef(nil), result.ResourceRefs...), nil
}
