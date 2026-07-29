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

const ExtensionImportFacadesOverrideKey = "imports.extension_import_facades"

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

type ExtensionImportMappingRequest struct {
	IncidentID           uuid.UUID
	ActorUserID          uuid.UUID
	TargetKind           string
	ExtensionProfileID   string
	ImportSessionID      uuid.UUID
	ImportUnitID         uuid.UUID
	SourceCapability     ImportSourceCapability
	OwnerMappingSchemaID string
	OwnerMapping         json.RawMessage
	ClientTxnID          string
}

type ExtensionImportMappingResult struct {
	OwnerMapping        json.RawMessage
	MappingFingerprint  string
	OwnerResultSchemaID string
	OwnerResult         map[string]any
}

type ExtensionImportApplyResult struct {
	ResourceRefs  []jobs.ResourceRef
	OwnerResponse map[string]any
}

type ExtensionImportFacade interface {
	PrepareImportUnitMapping(context.Context, ExtensionImportMappingRequest) (ExtensionImportMappingResult, error)
	ValidateImportUnitMappingResult(ExtensionImportMappingResult) error
	ApplyImportUnitTx(context.Context, pgx.Tx, ExtensionImportApplyRequest) (ExtensionImportApplyResult, error)
}

func extensionImportFacadesFromDependencies(deps httpapi.DependencySet) (map[string]ExtensionImportFacade, error) {
	facades := map[string]ExtensionImportFacade{}
	override, ok := deps.ModuleOverrides[ExtensionImportFacadesOverrideKey]
	if !ok || override == nil {
		return facades, nil
	}
	typed, ok := override.(map[string]ExtensionImportFacade)
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

func ExtensionImportFacadeKey(targetKind string, extensionProfileID string) string {
	return targetKind + ":" + extensionProfileID
}

func extensionImportFacadeKey(target importTarget) string {
	return ExtensionImportFacadeKey(target.TargetKind, target.ExtensionProfileID)
}

func (s *Service) applyExtensionOwnerUnitTx(
	ctx context.Context,
	tx pgx.Tx,
	actor authn.UserRecord,
	start ApplyStartResult,
	unit ApplyUnitData,
	target importTarget,
) (appliedUnitCommit, error) {
	facade := s.extensionImportFacades[extensionImportFacadeKey(target)]
	if facade == nil {
		return appliedUnitCommit{}, importApplyBlockedError("owner_apply_contract_unavailable")
	}
	result, err := facade.ApplyImportUnitTx(ctx, tx, ExtensionImportApplyRequest{
		IncidentID:         start.IncidentID,
		ActorUserID:        actor.ID,
		TargetKind:         target.TargetKind,
		ExtensionProfileID: target.ExtensionProfileID,
		ImportSessionID:    start.ImportSessionID,
		ImportUnitID:       unit.UnitID,
		SourceCapability: ImportSourceCapability{
			SourceStreamRef:     unit.SourceStreamRef,
			SourceContentSHA256: unit.SourceContentSHA256,
		},
		ExpectedSourceContentSHA256: unit.SourceContentSHA256,
		MappingFingerprint:          unit.MappingFingerprint,
		OwnerMappingSchemaID:        unit.ApprovedMapping.OwnerMappingSchemaID,
		OwnerMapping:                append(json.RawMessage(nil), unit.ApprovedMapping.OwnerMapping...),
		ClientTxnID:                 start.ClientTxnID,
	})
	if err != nil {
		return appliedUnitCommit{}, err
	}
	return appliedUnitCommit{
		OwnerResult:  result.OwnerResponse,
		ResourceRefs: append([]jobs.ResourceRef(nil), result.ResourceRefs...),
	}, nil
}
