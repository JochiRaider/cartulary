package imports

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

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

// ExtensionImportFacadeBinding is the complete self-identifying analytical
// facade contract projected from Core Table 17.2-F.
type ExtensionImportFacadeBinding struct {
	SchemaID               string
	TargetKind             string
	ExtensionProfileID     string
	OwnerContractRef       string
	FacadeID               string
	ContractMajor          int
	MappingSchemaID        string
	PreviewRequestSchemaID string
	PreviewResultSchemaID  string
	ApplyRequestSchemaID   string
	ApplyResultSchemaID    string
	ErrorSchemaID          string
	ErrorTranslationID     string
	CommitProtocolID       string
}

type ExtensionImportFacade interface {
	Binding() ExtensionImportFacadeBinding
	PrepareImportUnitMapping(context.Context, ExtensionImportMappingRequest) (ExtensionImportMappingResult, error)
	ValidateImportUnitMappingResult(ExtensionImportMappingResult) error
	ApplyImportUnitTx(context.Context, pgx.Tx, ExtensionImportApplyRequest) (ExtensionImportApplyResult, error)
	TranslateImportUnitError(error) (ExtensionImportErrorTranslation, bool)
	ValidateImportUnitError(ExtensionImportOwnerError) error
}

func validateExtensionImportFacades(
	contributions []ExtensionImportFacade,
	extensionProfileAdmitted func(string) bool,
) (map[analyticalImportTargetKey]ExtensionImportFacade, error) {
	validated := make(map[analyticalImportTargetKey]ExtensionImportFacade, len(contributions))
	facadeIDs := make(map[string]analyticalImportTargetKey, len(contributions))
	for index, facade := range contributions {
		if nilInterface(facade) {
			return nil, fmt.Errorf("imports analytical facade contribution %d is nil", index)
		}
		binding := facade.Binding()
		selector := analyticalImportTargetKey{
			TargetKind:         binding.TargetKind,
			ExtensionProfileID: binding.ExtensionProfileID,
		}
		target, known := analyticalImportTargets[selector]
		if !known {
			return nil, fmt.Errorf(
				"imports analytical facade contribution %d has unknown selector %q/%q",
				index,
				binding.TargetKind,
				binding.ExtensionProfileID,
			)
		}
		if _, duplicate := validated[selector]; duplicate {
			return nil, fmt.Errorf(
				"imports analytical facade selector %q/%q is duplicated",
				selector.TargetKind,
				selector.ExtensionProfileID,
			)
		}
		if prior, duplicate := facadeIDs[binding.FacadeID]; duplicate {
			return nil, fmt.Errorf(
				"imports analytical facade id %q is duplicated for %q/%q and %q/%q",
				binding.FacadeID,
				prior.TargetKind,
				prior.ExtensionProfileID,
				selector.TargetKind,
				selector.ExtensionProfileID,
			)
		}
		if err := validateExtensionImportFacadeBinding(binding, target); err != nil {
			return nil, fmt.Errorf("imports analytical facade %q: %w", binding.FacadeID, err)
		}
		validated[selector] = facade
		facadeIDs[binding.FacadeID] = selector
	}
	for selector, target := range analyticalImportTargets {
		required := target.ApplyStatus == applyStatusSupported ||
			target.ApplyStatus == applyStatusSupportedWhenClaimed && extensionProfileAdmitted(selector.ExtensionProfileID)
		if required && validated[selector] == nil {
			return nil, fmt.Errorf(
				"imports analytical target %q/%q requires its facade",
				selector.TargetKind,
				selector.ExtensionProfileID,
			)
		}
	}
	return validated, nil
}

func validateExtensionImportFacadeBinding(binding ExtensionImportFacadeBinding, target importTarget) error {
	want := ExtensionImportFacadeBinding{
		SchemaID:               target.BindingSchemaID,
		TargetKind:             target.TargetKind,
		ExtensionProfileID:     target.ExtensionProfileID,
		OwnerContractRef:       target.OwnerContractRef,
		FacadeID:               target.ApplyFacade,
		ContractMajor:          target.ContractMajor,
		MappingSchemaID:        target.MappingSchemaID,
		PreviewRequestSchemaID: target.PreviewRequestID,
		PreviewResultSchemaID:  target.PreviewResultID,
		ApplyRequestSchemaID:   target.ApplyRequestID,
		ApplyResultSchemaID:    target.ApplyResultID,
		ErrorSchemaID:          target.ErrorSchemaID,
		ErrorTranslationID:     target.ErrorTranslationID,
		CommitProtocolID:       target.CommitProtocolID,
	}
	checks := []struct {
		field string
		got   string
		want  string
	}{
		{field: "schema_id", got: binding.SchemaID, want: want.SchemaID},
		{field: "target_kind", got: binding.TargetKind, want: want.TargetKind},
		{field: "extension_profile_id", got: binding.ExtensionProfileID, want: want.ExtensionProfileID},
		{field: "owner_contract_ref", got: binding.OwnerContractRef, want: want.OwnerContractRef},
		{field: "facade_id", got: binding.FacadeID, want: want.FacadeID},
		{field: "mapping_schema_id", got: binding.MappingSchemaID, want: want.MappingSchemaID},
		{field: "preview_request_schema_id", got: binding.PreviewRequestSchemaID, want: want.PreviewRequestSchemaID},
		{field: "preview_result_schema_id", got: binding.PreviewResultSchemaID, want: want.PreviewResultSchemaID},
		{field: "apply_request_schema_id", got: binding.ApplyRequestSchemaID, want: want.ApplyRequestSchemaID},
		{field: "apply_result_schema_id", got: binding.ApplyResultSchemaID, want: want.ApplyResultSchemaID},
		{field: "error_schema_id", got: binding.ErrorSchemaID, want: want.ErrorSchemaID},
		{field: "error_translation_id", got: binding.ErrorTranslationID, want: want.ErrorTranslationID},
		{field: "commit_protocol_id", got: binding.CommitProtocolID, want: want.CommitProtocolID},
	}
	for _, check := range checks {
		if check.got != check.want {
			return fmt.Errorf("%s got %q want %q", check.field, check.got, check.want)
		}
	}
	if binding.ContractMajor != want.ContractMajor {
		return fmt.Errorf("contract_major got %d want %d", binding.ContractMajor, want.ContractMajor)
	}
	return nil
}

func (s *service) applyExtensionOwnerUnitTx(
	ctx context.Context,
	tx pgx.Tx,
	actor authn.UserRecord,
	start applyStartResult,
	unit applyUnitData,
	target importTarget,
) (appliedUnitCommit, error) {
	facade := s.extensionImportFacades[analyticalImportTargetKey{
		TargetKind:         target.TargetKind,
		ExtensionProfileID: target.ExtensionProfileID,
	}]
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
		OwnerMappingSchemaID:        unit.approvedMapping.OwnerMappingSchemaID,
		OwnerMapping:                append(json.RawMessage(nil), unit.approvedMapping.OwnerMapping...),
		ClientTxnID:                 start.ClientTxnID,
	})
	if err != nil {
		return appliedUnitCommit{}, &translatedImportUnitError{
			failure: translateExtensionOwnerFailure(target, facade, err),
			cause:   err,
		}
	}
	return appliedUnitCommit{
		OwnerResult:  result.OwnerResponse,
		ResourceRefs: append([]jobs.ResourceRef(nil), result.ResourceRefs...),
	}, nil
}
