package networkflow

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type importApplyParticipant struct {
	request imports.ExtensionImportApplyRequest
	limits  EffectiveLimits
}

// ImportTransaction binds the approved Core contexts before the coordinator
// builds input. Source bytes are read only through the eventual unit transaction.
func (f *importFacade) ImportTransaction(request imports.ExtensionImportApplyRequest) (crossownertransaction.Participant, error) {
	if f == nil || f.store == nil || f.sourceStore == nil || f.safeDigester == nil {
		return nil, importOwnerError("network_flow_target_unavailable", map[string]any{"reason_code": "owner_apply_contract_unavailable"})
	}
	if request.TargetKind != targetKindNetworkFlowTable || request.ExtensionProfileID != ProfileID || request.OwnerMappingSchemaID != importMappingSchemaID {
		return nil, importOwnerError("network_flow_mapping_invalid", map[string]any{"reason_code": "variant_member_conflict"})
	}
	if request.SourceCapability.SourceContentSHA256 != request.ExpectedSourceContentSHA256 {
		return nil, importOwnerError("network_flow_source_changed", nil)
	}
	request.OwnerMapping = append(json.RawMessage(nil), request.OwnerMapping...)
	return &importApplyParticipant{request: request, limits: f.limits}, nil
}

func (*importApplyParticipant) ID() string { return ImportApplyParticipantID }

func (p *importApplyParticipant) BuildInput(context.Context, crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	r := p.request
	// Preserve the owner's source-change classification before shared wire
	// admission. This check is bounded input construction and opens no source.
	mapping, err := decodeApprovedMapping(r.OwnerMapping, p.limits)
	if err != nil {
		return crossownertransaction.Input{}, facadeError(err)
	}
	fingerprint, err := mappingFingerprint(mapping, r.ExpectedSourceContentSHA256)
	if err != nil {
		return crossownertransaction.Input{}, facadeError(err)
	}
	if fingerprint != r.MappingFingerprint {
		return crossownertransaction.Input{}, importOwnerError("network_flow_source_changed", nil)
	}
	data, err := canonicaljson.MarshalExtension(map[string]any{
		"schema_id":   importApplyTransactionInputSchema,
		"incident_id": r.IncidentID.String(), "actor_user_id": r.ActorUserID.String(),
		"import_session_id": r.ImportSessionID.String(), "import_unit_id": r.ImportUnitID.String(),
		"source_stream_ref":       r.SourceCapability.SourceStreamRef,
		"source_content_sha256":   r.ExpectedSourceContentSHA256,
		"mapping_fingerprint":     r.MappingFingerprint,
		"owner_mapping_schema_id": r.OwnerMappingSchemaID, "owner_mapping": r.OwnerMapping,
		"client_txn_id": r.ClientTxnID,
	})
	return crossownertransaction.Input{SchemaID: importApplyTransactionInputSchema, CanonicalBytes: data}, err
}

func (p *importApplyParticipant) Prepare(context.Context, crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	return crossownertransaction.PrepareResult{SerializationKeys: []crossownertransaction.SerializationKey{
		{KeyKind: "network_flow_activity.import_unit", Key: p.request.ImportUnitID.String()},
		{KeyKind: "network_flow_activity.incident", Key: p.request.IncidentID.String()},
	}}, nil
}

type importApplyReadCapability interface {
	crossownertransaction.ReadCapability
	ValidateImportApply(context.Context, imports.ExtensionImportApplyRequest) error
}

type importApplyWriteCapability interface {
	crossownertransaction.WriteCapability
	WriteImportApply(context.Context, imports.ExtensionImportApplyRequest) (imports.ExtensionImportApplyResult, error)
}

type importApplyReadAccess struct {
	tx      pgx.Tx
	sources ImportSourcePort
}

type importApplyWriteAccess struct {
	tx          pgx.Tx
	importOwner *importFacade
}

func (*importApplyReadAccess) ParticipantScope() string  { return ImportApplyParticipantID }
func (*importApplyWriteAccess) ParticipantScope() string { return ImportApplyParticipantID }

func (c *importApplyReadAccess) ValidateImportApply(ctx context.Context, request imports.ExtensionImportApplyRequest) error {
	if c == nil || c.tx == nil || c.sources == nil {
		return crossownertransaction.ErrUnavailable
	}
	return c.sources.ValidateExtensionApplyPreconditionsTx(
		ctx, c.tx, request.IncidentID, request.ImportSessionID, request.ImportUnitID,
		request.SourceCapability.SourceStreamRef, request.ExpectedSourceContentSHA256,
	)
}

func (p *importApplyParticipant) Validate(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	access, ok := invocation.ReadAccess.(importApplyReadCapability)
	if !ok || access.ParticipantScope() != p.ID() {
		return crossownertransaction.ValidationResult{}, crossownertransaction.ErrUnavailable
	}
	if err := access.ValidateImportApply(ctx, p.request); err != nil {
		return crossownertransaction.ValidationResult{}, err
	}
	return crossownertransaction.Valid(), nil
}

func (p *importApplyParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	access, ok := invocation.WriteAccess.(importApplyWriteCapability)
	if !ok || access.ParticipantScope() != p.ID() {
		return crossownertransaction.WriteResult{}, crossownertransaction.ErrUnavailable
	}
	result, err := access.WriteImportApply(ctx, p.request)
	if err != nil {
		return crossownertransaction.WriteResult{}, err
	}
	return crossownertransaction.Written(result), nil
}

func (c *importApplyWriteAccess) WriteImportApply(ctx context.Context, request imports.ExtensionImportApplyRequest) (imports.ExtensionImportApplyResult, error) {
	if c == nil || c.tx == nil || c.importOwner == nil {
		return imports.ExtensionImportApplyResult{}, crossownertransaction.ErrUnavailable
	}
	return c.importOwner.applyImportUnitTx(ctx, c.tx, request)
}
