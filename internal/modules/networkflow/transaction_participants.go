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
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	ImportApplyParticipantID   = "network_flow_activity.import_apply_v1"
	IndicatorLinkParticipantID = "network_flow_activity.indicator_link_v1"
)

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
	store         *store
	imports       ImportSourcePort
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

func (c *transactionCapability) CreateImportedTable(ctx context.Context, params createTableParams) (tableRecord, error) {
	if c == nil || c.participantID != ImportApplyParticipantID || c.store == nil {
		return tableRecord{}, crossownertransaction.ErrUnavailable
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
	SafeDigester safeDigester
}

type indicatorLinkCommitResult struct {
	Binding   indicatorBindingRecord
	Duplicate bool
}

func (c *transactionCapability) ValidateIndicatorLinkTarget(ctx context.Context, mutation indicatorLinkMutation) error {
	if c == nil || c.participantID != IndicatorLinkParticipantID || c.store == nil || c.store.indicators == nil {
		return crossownertransaction.ErrUnavailable
	}
	if err := c.store.lockIncidentTx(ctx, c.tx, mutation.IncidentID); err != nil {
		return err
	}
	if _, err := admission.NewChecker(c.store.pool).CheckTx(ctx, c.tx, mutation.IncidentID, mutation.Actor.ID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return err
	}
	if err := validateIndicatorLinkSourcesTx(ctx, c.tx, mutation); err != nil {
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

// Accepted flow rows and mappings are immutable. Under the incident lock,
// checking every selected table and captured row closes the deletion race,
// including graph tables whose contributors fall beyond the retained prefix.
func validateIndicatorLinkSourcesTx(ctx context.Context, tx pgx.Tx, mutation indicatorLinkMutation) error {
	tableIDs := map[string]bool{}
	for _, id := range mutation.Request.Selector.GraphQuery.SelectedTableIDs {
		tableIDs[id] = true
	}
	for _, ref := range mutation.Resolved.SourceRowRefs {
		tableIDs[ref.NetworkFlowTableID] = true
	}
	ordered := make([]string, 0, len(tableIDs))
	for id := range tableIDs {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	for _, id := range ordered {
		var status string
		err := tx.QueryRow(ctx, `SELECT table_status FROM network_flow_tables WHERE incident_id = $1 AND network_flow_table_id = $2`, mutation.IncidentID, id).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return linkSourceTableError(tableReadFailure(errTableNotFound), id)
		}
		if err == nil && status != "active" {
			return linkSourceTableError(tableReadFailure(errTableNotActive), id)
		}
		if err != nil {
			return err
		}
	}
	for _, ref := range mutation.Resolved.SourceRowRefs {
		var accepted bool
		err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM network_flow_rows WHERE incident_id = $1 AND network_flow_table_id = $2 AND network_flow_row_id = $3 AND source_row_number = $4 AND mapping_fingerprint = $5)`, mutation.IncidentID, ref.NetworkFlowTableID, ref.NetworkFlowRowID, ref.SourceRowNumber, ref.MappingFingerprint).Scan(&accepted)
		if err != nil {
			return err
		}
		if !accepted {
			return invalidIndicatorSelector("row_refs", "row_not_accepted")
		}
	}
	return nil
}

func (c *transactionCapability) WriteIndicatorLink(ctx context.Context, mutation indicatorLinkMutation) (indicatorLinkCommitResult, error) {
	if c == nil || c.participantID != IndicatorLinkParticipantID || c.store == nil || c.store.indicators == nil {
		return indicatorLinkCommitResult{}, crossownertransaction.ErrUnavailable
	}
	var target indicators.IndicatorReference
	var err error
	switch mutation.Request.Target.Mode {
	case "existing_indicator":
		target, err = c.store.indicators.GetActiveIndicatorParticipantTx(ctx, c.tx, mutation.IncidentID, mutation.Request.Target.IndicatorID)
	case "create_indicator":
		var result indicators.IndicatorFindOrCreateParticipantResult
		result, err = c.store.indicators.FindOrCreateIndicatorParticipantTx(ctx, c.tx, indicators.IndicatorFindOrCreateParticipantCommand{
			IncidentID: mutation.IncidentID, ActorUserID: mutation.Actor.ID,
			IndicatorType: mutation.Request.Target.IndicatorType, ValueKind: "atomic",
			DisplayValue: mutation.Resolved.CandidateValue, NormalizedValue: &mutation.Resolved.CandidateValue,
			OperationContext:  "network_flow_indicator_link",
			OperationOccurred: mutation.Now,
		})
		target = result.Indicator
	default:
		err = errInvalidStorageArgument
	}
	if err != nil {
		return indicatorLinkCommitResult{}, err
	}
	if err := validateIndicatorTargetLogical(target, mutation.Resolved.CandidateValue, mutation.TargetType); err != nil {
		return indicatorLinkCommitResult{}, err
	}
	binding, duplicate, err := c.store.CreateOrReuseIndicatorBindingTx(ctx, c.tx, createIndicatorBindingParams{
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
	if err := (indicatorLinkReceiptAdapter{}).saveTx(ctx, c.tx, mutation, binding, duplicate); err != nil {
		return indicatorLinkCommitResult{}, err
	}
	return indicatorLinkCommitResult{Binding: binding, Duplicate: duplicate}, nil
}

func validateIndicatorTargetLogical(record indicators.IndicatorReference, candidateValue string, targetType string) error {
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
