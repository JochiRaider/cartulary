package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	ImportApplyParticipantID   = "network_flow_activity.import_apply_v1"
	IndicatorLinkParticipantID = "network_flow_activity.indicator_link_v1"
)

type importApplyReadCapability interface {
	crossownertransaction.ReadCapability
	ValidateImportApply(context.Context, imports.ExtensionImportApplyRequest) error
}

type importApplyWriteCapability interface {
	crossownertransaction.WriteCapability
	CreateImportedTable(context.Context, CreateTableParams) (TableRecord, error)
}

type indicatorLinkReadCapability interface {
	crossownertransaction.ReadCapability
	ValidateIndicatorLinkTarget(context.Context, indicatorLinkMutation) error
}

type indicatorLinkWriteCapability interface {
	crossownertransaction.WriteCapability
	WriteIndicatorLink(context.Context, indicatorLinkMutation) (indicatorLinkCommitResult, error)
}

type transactionCapability struct {
	participantID string
	tx            pgx.Tx
	store         *Store
	imports       importTransactionPort
}

func (c *transactionCapability) ParticipantScope() string {
	if c == nil {
		return ""
	}
	return c.participantID
}

func (c *transactionCapability) ValidateImportApply(ctx context.Context, request imports.ExtensionImportApplyRequest) error {
	if c == nil || c.participantID != ImportApplyParticipantID || c.imports == nil {
		return crossownertransaction.ErrUnavailable
	}
	return c.imports.ValidateExtensionApplyPreconditionsTx(
		ctx, c.tx, request.IncidentID, request.ImportSessionID, request.ImportUnitID,
		request.SourceCapability.SourceStreamRef, request.ExpectedSourceContentSHA256,
	)
}

func (c *transactionCapability) CreateImportedTable(ctx context.Context, params CreateTableParams) (TableRecord, error) {
	if c == nil || c.participantID != ImportApplyParticipantID || c.store == nil {
		return TableRecord{}, crossownertransaction.ErrUnavailable
	}
	return c.store.CreateTableTx(ctx, c.tx, params)
}

type indicatorLinkMutation struct {
	IncidentID   uuid.UUID
	Actor        authn.UserRecord
	Request      indicatorLinkRequest
	Resolved     resolvedIndicatorLinkSelector
	TargetType   string
	RequestHash  []byte
	RequestID    string
	Now          time.Time
	SafeDigester SafeDigester
}

type indicatorLinkCommitResult struct {
	Binding   IndicatorBindingRecord
	Duplicate bool
	Payload   map[string]any
	Status    int
}

func (c *transactionCapability) ValidateIndicatorLinkTarget(ctx context.Context, mutation indicatorLinkMutation) error {
	if c == nil || c.participantID != IndicatorLinkParticipantID || c.store == nil || c.store.indicators == nil {
		return crossownertransaction.ErrUnavailable
	}
	if err := c.store.lockIncidentTx(ctx, c.tx, mutation.IncidentID); err != nil {
		return err
	}
	if mutation.Request.Target.Mode != "existing_indicator" {
		return nil
	}
	record, err := c.store.indicators.GetActiveIndicatorParticipantTx(ctx, c.tx, mutation.IncidentID, mutation.Request.Target.IndicatorID)
	if err != nil {
		return err
	}
	return validateIndicatorTargetLogical(record, mutation.Resolved.CandidateValue, mutation.TargetType)
}

func (c *transactionCapability) WriteIndicatorLink(ctx context.Context, mutation indicatorLinkMutation) (indicatorLinkCommitResult, error) {
	if c == nil || c.participantID != IndicatorLinkParticipantID || c.store == nil || c.store.indicators == nil {
		return indicatorLinkCommitResult{}, crossownertransaction.ErrUnavailable
	}
	var target indicators.IndicatorRecord
	var err error
	switch mutation.Request.Target.Mode {
	case "existing_indicator":
		target, err = c.store.indicators.GetActiveIndicatorParticipantTx(ctx, c.tx, mutation.IncidentID, mutation.Request.Target.IndicatorID)
	case "create_indicator":
		var result indicators.IndicatorFindOrCreateParticipantResult
		result, err = c.store.indicators.FindOrCreateIndicatorParticipantTx(ctx, c.tx, indicators.IndicatorFindOrCreateParticipantCommand{
			IncidentID: mutation.IncidentID, Actor: mutation.Actor,
			IndicatorType: mutation.Request.Target.IndicatorType, ValueKind: "atomic",
			DisplayValue: mutation.Resolved.CandidateValue, NormalizedValue: &mutation.Resolved.CandidateValue,
			OperationContext:  "network_flow_indicator_link",
			OperationOccurred: mutation.Now,
		})
		target = result.Indicator
	default:
		err = ErrInvalidStorageArgument
	}
	if err != nil {
		return indicatorLinkCommitResult{}, err
	}
	if err := validateIndicatorTargetLogical(target, mutation.Resolved.CandidateValue, mutation.TargetType); err != nil {
		return indicatorLinkCommitResult{}, err
	}
	binding, duplicate, err := c.store.CreateOrReuseIndicatorBindingTx(ctx, c.tx, CreateIndicatorBindingParams{
		IncidentID: mutation.IncidentID, ActorUserID: mutation.Actor.ID, TargetIndicator: target,
		SelectorKind: mutation.Resolved.SelectorKind, CandidateValue: mutation.Resolved.CandidateValue,
		SourceRowRefs: mutation.Resolved.SourceRowRefs, SourceRowRefsTruncated: mutation.Resolved.SourceRowRefsTruncated,
		SourceRowRefsTotalCount: mutation.Resolved.SourceRowRefsTotalCount,
		ClientTxnID:             mutation.Request.ClientTxnID, RequestID: mutation.RequestID,
		SafeDigester: mutation.SafeDigester, Now: mutation.Now,
	})
	if err != nil {
		return indicatorLinkCommitResult{}, err
	}
	status := 201
	if duplicate {
		status = 200
	}
	payload := indicatorLinkPayload(binding, duplicate)
	if err := authn.InsertRouteIdempotencyPayload(ctx, c.tx, indicatorLinkIdempotencyKey(mutation.Actor.ID, mutation.IncidentID, mutation.Request.ClientTxnID), nil, mutation.RequestHash, status, payload); err != nil {
		return indicatorLinkCommitResult{}, err
	}
	return indicatorLinkCommitResult{Binding: binding, Duplicate: duplicate, Payload: payload, Status: status}, nil
}

func validateIndicatorTargetLogical(record indicators.IndicatorRecord, candidateValue string, targetType string) error {
	if record.IndicatorType != targetType {
		return &indicatorTargetParticipantError{ReasonCode: "target_type_mismatch"}
	}
	if record.ValueKind != "atomic" || record.NormalizedValue == nil || *record.NormalizedValue != candidateValue {
		return &indicatorTargetParticipantError{ReasonCode: "target_value_mismatch"}
	}
	return nil
}

type indicatorTargetParticipantError struct {
	ReasonCode string
}

func (e *indicatorTargetParticipantError) Error() string {
	return "network flow indicator target invalid: " + e.ReasonCode
}

type importApplyParticipant struct {
	facade   *importFacade
	request  imports.ExtensionImportApplyRequest
	prepared preparedImportApply
}

func (p *importApplyParticipant) ID() string { return ImportApplyParticipantID }

func (p *importApplyParticipant) BuildInput(_ context.Context, _ crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	canonical, err := json.Marshal(map[string]any{
		"schema_id":                      "cartulary.network_flow_activity.import_apply_transaction_input.v1",
		"incident_id":                    p.request.IncidentID.String(),
		"actor_user_id":                  p.request.ActorUserID.String(),
		"target_kind":                    p.request.TargetKind,
		"extension_profile_id":           p.request.ExtensionProfileID,
		"import_session_id":              p.request.ImportSessionID.String(),
		"import_unit_id":                 p.request.ImportUnitID.String(),
		"source_stream_ref":              p.request.SourceCapability.SourceStreamRef,
		"expected_source_content_sha256": p.request.ExpectedSourceContentSHA256,
		"mapping_fingerprint":            p.request.MappingFingerprint,
		"owner_mapping_schema_id":        p.request.OwnerMappingSchemaID,
		"owner_mapping":                  json.RawMessage(p.request.OwnerMapping),
		"client_txn_id":                  p.request.ClientTxnID,
	})
	return crossownertransaction.Input{
		SchemaID:       "cartulary.network_flow_activity.import_apply_transaction_input.v1",
		CanonicalBytes: canonical,
	}, err
}

func (p *importApplyParticipant) Prepare(ctx context.Context, _ crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	prepared, err := p.facade.prepareImportApply(ctx, p.request)
	if err != nil {
		return crossownertransaction.PrepareResult{}, err
	}
	p.prepared = prepared
	return crossownertransaction.PrepareResult{SerializationKeys: []crossownertransaction.SerializationKey{
		{KeyKind: "network_flow_activity.import_unit", Key: p.request.ImportSessionID.String() + ":" + p.request.ImportUnitID.String()},
		{KeyKind: "network_flow_activity.incident", Key: p.request.IncidentID.String()},
	}}, nil
}

func (p *importApplyParticipant) Validate(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	access, ok := invocation.ReadAccess.(importApplyReadCapability)
	if !ok {
		return crossownertransaction.ValidationResult{}, crossownertransaction.ErrUnavailable
	}
	if err := access.ValidateImportApply(ctx, p.request); err != nil {
		return crossownertransaction.ValidationResult{}, err
	}
	return crossownertransaction.Valid(), nil
}

func (p *importApplyParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	access, ok := invocation.WriteAccess.(importApplyWriteCapability)
	if !ok {
		return crossownertransaction.WriteResult{}, crossownertransaction.ErrUnavailable
	}
	table, err := access.CreateImportedTable(ctx, p.prepared.params)
	if err != nil {
		return crossownertransaction.WriteResult{}, storeApplyError(err)
	}
	result := p.prepared.result(table)
	return crossownertransaction.Written(result), nil
}

type indicatorLinkParticipant struct {
	mutation indicatorLinkMutation
}

func (p *indicatorLinkParticipant) ID() string { return IndicatorLinkParticipantID }

func (p *indicatorLinkParticipant) BuildInput(_ context.Context, _ crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	canonical, err := json.Marshal(map[string]any{
		"schema_id":   "cartulary.network_flow_activity.indicator_link_transaction_input.v1",
		"incident_id": p.mutation.IncidentID.String(), "actor_user_id": p.mutation.Actor.ID.String(),
		"request":         indicatorSelectorHashResource(p.mutation.Request.Selector),
		"target":          indicatorTargetHashResource(p.mutation.Request.Target),
		"candidate_value": p.mutation.Resolved.CandidateValue,
		"client_txn_id":   p.mutation.Request.ClientTxnID,
	})
	return crossownertransaction.Input{
		SchemaID:       "cartulary.network_flow_activity.indicator_link_transaction_input.v1",
		CanonicalBytes: canonical,
	}, err
}

func (p *indicatorLinkParticipant) Prepare(_ context.Context, _ crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	keys := []crossownertransaction.SerializationKey{
		{KeyKind: "network_flow_activity.incident", Key: p.mutation.IncidentID.String()},
	}
	tableIDs := map[string]struct{}{}
	for _, ref := range p.mutation.Resolved.SourceRowRefs {
		tableIDs[ref.NetworkFlowTableID] = struct{}{}
	}
	sortedTableIDs := make([]string, 0, len(tableIDs))
	for tableID := range tableIDs {
		sortedTableIDs = append(sortedTableIDs, tableID)
	}
	sort.Strings(sortedTableIDs)
	for _, tableID := range sortedTableIDs {
		keys = append(keys, crossownertransaction.SerializationKey{KeyKind: "network_flow_activity.table", Key: tableID})
	}
	indicatorKey := p.mutation.Resolved.CandidateValue
	if p.mutation.Request.Target.Mode == "existing_indicator" {
		indicatorKey = p.mutation.Request.Target.IndicatorID.String()
	}
	keys = append(keys, crossownertransaction.SerializationKey{KeyKind: "network_flow_activity.indicator", Key: indicatorKey})
	return crossownertransaction.PrepareResult{SerializationKeys: keys}, nil
}

func (p *indicatorLinkParticipant) Validate(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	access, ok := invocation.ReadAccess.(indicatorLinkReadCapability)
	if !ok {
		return crossownertransaction.ValidationResult{}, crossownertransaction.ErrUnavailable
	}
	if err := access.ValidateIndicatorLinkTarget(ctx, p.mutation); err != nil {
		return crossownertransaction.ValidationResult{}, err
	}
	return crossownertransaction.Valid(), nil
}

func (p *indicatorLinkParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	access, ok := invocation.WriteAccess.(indicatorLinkWriteCapability)
	if !ok {
		return crossownertransaction.WriteResult{}, crossownertransaction.ErrUnavailable
	}
	result, err := access.WriteIndicatorLink(ctx, p.mutation)
	if err != nil {
		return crossownertransaction.WriteResult{}, err
	}
	return crossownertransaction.Written(result), nil
}

func participantResult[T any](result crossownertransaction.Result, participantID string) (T, error) {
	var zero T
	value, ok := result.ParticipantValues[participantID]
	if !ok {
		return zero, fmt.Errorf("network flow participant result missing")
	}
	typed, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("network flow participant result has type %T", value)
	}
	return typed, nil
}

func indicatorParticipantReason(err error) string {
	var target *indicatorTargetParticipantError
	if errors.As(err, &target) {
		return target.ReasonCode
	}
	return ""
}
