package administrativeaudit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/gen/administrativeauditregistry"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	ScopeDeployment = administrativeauditregistry.ScopeDeployment
	ScopeIncident   = administrativeauditregistry.ScopeIncident

	ActorOperator = administrativeauditregistry.ActorOperator
	ActorSystem   = administrativeauditregistry.ActorSystem
	ActorUser     = administrativeauditregistry.ActorUser

	SourceAPI      = administrativeauditregistry.SourceApi
	SourceOperator = administrativeauditregistry.SourceOperator
	SourceStartup  = administrativeauditregistry.SourceStartup
	SourceSystem   = administrativeauditregistry.SourceSystem
	SourceUI       = administrativeauditregistry.SourceUi

	ValueRedacted = administrativeauditregistry.ValueRedacted
	ValueVisible  = administrativeauditregistry.ValueVisible

	ActionAccountPreferencesUpdated    = administrativeauditregistry.ActionAccountPreferencesUpdated
	ActionAuthBindingCreated           = administrativeauditregistry.ActionAuthBindingCreated
	ActionAuthBindingRetired           = administrativeauditregistry.ActionAuthBindingRetired
	ActionAuthBindingRotated           = administrativeauditregistry.ActionAuthBindingRotated
	ActionBackupCreated                = administrativeauditregistry.ActionBackupCreated
	ActionBootstrapAdminCreated        = administrativeauditregistry.ActionBootstrapAdminCreated
	ActionDeploymentAdminGranted       = administrativeauditregistry.ActionDeploymentAdminGranted
	ActionDeploymentAdminRevoked       = administrativeauditregistry.ActionDeploymentAdminRevoked
	ActionMembershipCreated            = administrativeauditregistry.ActionMembershipCreated
	ActionMembershipDeleted            = administrativeauditregistry.ActionMembershipDeleted
	ActionMembershipRoleChanged        = administrativeauditregistry.ActionMembershipRoleChanged
	ActionPasswordChanged              = administrativeauditregistry.ActionPasswordChanged
	ActionPasswordReset                = administrativeauditregistry.ActionPasswordReset
	ActionRestoreCompleted             = administrativeauditregistry.ActionRestoreCompleted
	ActionRestoreFailed                = administrativeauditregistry.ActionRestoreFailed
	ActionRestoreStarted               = administrativeauditregistry.ActionRestoreStarted
	ActionRestoreVerificationCompleted = administrativeauditregistry.ActionRestoreVerificationCompleted
	ActionSessionsRevoked              = administrativeauditregistry.ActionSessionsRevoked
	ActionTOTPEnrollmentBegun          = administrativeauditregistry.ActionTotpEnrollmentBegun
	ActionTOTPEnrollmentCompleted      = administrativeauditregistry.ActionTotpEnrollmentCompleted
	ActionTOTPReset                    = administrativeauditregistry.ActionTotpReset
	ActionUserCreated                  = administrativeauditregistry.ActionUserCreated
	ActionUserProfileUpdated           = administrativeauditregistry.ActionUserProfileUpdated
	ActionUserStatusChanged            = administrativeauditregistry.ActionUserStatusChanged

	TargetAccountPreferences = administrativeauditregistry.TargetAccountPreferences
	TargetAuthBinding        = administrativeauditregistry.TargetAuthBinding
	TargetBackupSet          = administrativeauditregistry.TargetBackupSet
	TargetIncidentMembership = administrativeauditregistry.TargetIncidentMembership
	TargetRestoreOperation   = administrativeauditregistry.TargetRestoreOperation
	TargetUser               = administrativeauditregistry.TargetUser
)

var (
	ErrInvalidEvent  = errors.New("administrative audit event is invalid")
	ErrUnsafeChanges = errors.New("administrative audit changes are unsafe")
)

type Change struct {
	FieldPath  string `json:"field_path"`
	ValueState string `json:"value_state"`
	Before     any    `json:"before"`
	After      any    `json:"after"`
}

type Event struct {
	ScopeKind   string
	ScopeID     *uuid.UUID
	OccurredAt  time.Time
	ActorKind   string
	ActorUserID *uuid.UUID
	Source      string
	ActionCode  string
	TargetKind  string
	TargetID    *string
	Changes     []Change
	ReasonCode  *string
}

type RawEvent struct {
	ActorUserID  *uuid.UUID
	TargetUserID *uuid.UUID
	IncidentID   *uuid.UUID
	EventSource  string
	EventKind    string
	ReasonCode   *string
	ClientTxnID  *string
	RequestID    *string
	Before       any
	After        any
	OccurredAt   time.Time
}

type actionBinding struct {
	scopeKind  string
	targetKind string
}

var actionBindings = func() map[string][]actionBinding {
	bindings := make(map[string][]actionBinding, len(administrativeauditregistry.ActionBindings))
	for _, binding := range administrativeauditregistry.ActionBindings {
		bindings[binding.ActionCode] = append(bindings[binding.ActionCode], actionBinding{
			scopeKind:  binding.ScopeKind,
			targetKind: binding.TargetKind,
		})
	}
	return bindings
}()

func ActionCodes(scopeKind string) []string {
	actionCodes := make([]string, 0, len(actionBindings))
	for actionCode, bindings := range actionBindings {
		for _, binding := range bindings {
			if binding.scopeKind == scopeKind {
				actionCodes = append(actionCodes, actionCode)
				break
			}
		}
	}
	sort.Strings(actionCodes)
	return actionCodes
}

func TargetKinds(scopeKind string) []string {
	targetSet := map[string]struct{}{}
	for _, bindings := range actionBindings {
		for _, binding := range bindings {
			if binding.scopeKind == scopeKind {
				targetSet[binding.targetKind] = struct{}{}
			}
		}
	}
	targetKinds := make([]string, 0, len(targetSet))
	for targetKind := range targetSet {
		targetKinds = append(targetKinds, targetKind)
	}
	sort.Strings(targetKinds)
	return targetKinds
}

var forbiddenVisibleFieldTokens = administrativeauditregistry.ForbiddenVisibleFieldTokens

func Visible(fieldPath string, before any, after any) Change {
	return Change{FieldPath: fieldPath, ValueState: ValueVisible, Before: before, After: after}
}

func Redacted(fieldPath string) Change {
	return Change{FieldPath: fieldPath, ValueState: ValueRedacted}
}

func AppendTx(ctx context.Context, tx pgx.Tx, raw RawEvent, event Event) (uuid.UUID, error) {
	changes, err := validateAndNormalizeEvent(event)
	if err != nil {
		return uuid.Nil, err
	}
	if err := validateRawProjectionPair(raw, event); err != nil {
		return uuid.Nil, err
	}
	changesJSON, err := json.Marshal(changes)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal administrative audit changes: %w", err)
	}
	eventID, err := appendRawTx(ctx, tx, raw)
	if err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO administrative_audit_projections (
    audit_event_id, scope_kind, scope_id, occurred_at, actor_kind, actor_user_id,
    source, action_code, target_kind, target_id, changes, reason_code
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12)
`,
		eventID,
		event.ScopeKind,
		event.ScopeID,
		event.OccurredAt.UTC(),
		event.ActorKind,
		event.ActorUserID,
		event.Source,
		event.ActionCode,
		event.TargetKind,
		event.TargetID,
		changesJSON,
		event.ReasonCode,
	); err != nil {
		return uuid.Nil, fmt.Errorf("insert administrative audit projection: %w", err)
	}
	return eventID, nil
}

func AppendRawTx(ctx context.Context, tx pgx.Tx, raw RawEvent) (uuid.UUID, error) {
	return appendRawTx(ctx, tx, raw)
}

func Append(ctx context.Context, db postgres.DB, raw RawEvent, event Event) (uuid.UUID, error) {
	if db == nil {
		return uuid.Nil, fmt.Errorf("%w: database is required", ErrInvalidEvent)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin administrative audit transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	eventID, err := AppendTx(ctx, tx, raw, event)
	if err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit administrative audit transaction: %w", err)
	}
	return eventID, nil
}

func AppendRaw(ctx context.Context, db postgres.DB, raw RawEvent) (uuid.UUID, error) {
	if db == nil {
		return uuid.Nil, fmt.Errorf("%w: database is required", ErrInvalidEvent)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin raw administrative audit transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	eventID, err := appendRawTx(ctx, tx, raw)
	if err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit raw administrative audit transaction: %w", err)
	}
	return eventID, nil
}

func appendRawTx(ctx context.Context, tx pgx.Tx, raw RawEvent) (uuid.UUID, error) {
	if tx == nil || strings.TrimSpace(raw.EventSource) == "" || strings.TrimSpace(raw.EventKind) == "" || raw.OccurredAt.IsZero() {
		return uuid.Nil, fmt.Errorf("%w: raw event fields are incomplete", ErrInvalidEvent)
	}
	beforeJSON, err := encodeRawValue(raw.Before)
	if err != nil {
		return uuid.Nil, err
	}
	afterJSON, err := encodeRawValue(raw.After)
	if err != nil {
		return uuid.Nil, err
	}
	eventID := uuid.New()
	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (
    id, actor_user_id, target_user_id, incident_id, event_source, event_kind,
    reason_code, client_txn_id, request_id, before_json, after_json, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11::jsonb, $12)
`,
		eventID,
		raw.ActorUserID,
		raw.TargetUserID,
		raw.IncidentID,
		raw.EventSource,
		raw.EventKind,
		raw.ReasonCode,
		raw.ClientTxnID,
		raw.RequestID,
		beforeJSON,
		afterJSON,
		raw.OccurredAt.UTC(),
	); err != nil {
		return uuid.Nil, fmt.Errorf("insert raw administrative audit event: %w", err)
	}
	return eventID, nil
}

func validateAndNormalizeEvent(event Event) ([]Change, error) {
	if event.OccurredAt.IsZero() {
		return nil, fmt.Errorf("%w: occurred_at is required", ErrInvalidEvent)
	}
	switch event.ScopeKind {
	case ScopeDeployment:
		if event.ScopeID != nil {
			return nil, fmt.Errorf("%w: deployment scope_id must be nil", ErrInvalidEvent)
		}
	case ScopeIncident:
		if event.ScopeID == nil {
			return nil, fmt.Errorf("%w: incident scope_id is required", ErrInvalidEvent)
		}
	default:
		return nil, fmt.Errorf("%w: unknown scope kind %q", ErrInvalidEvent, event.ScopeKind)
	}
	switch event.ActorKind {
	case ActorUser:
		if event.ActorUserID == nil {
			return nil, fmt.Errorf("%w: user actor requires actor_user_id", ErrInvalidEvent)
		}
	case ActorOperator, ActorSystem:
		if event.ActorUserID != nil {
			return nil, fmt.Errorf("%w: non-user actor forbids actor_user_id", ErrInvalidEvent)
		}
	default:
		return nil, fmt.Errorf("%w: unknown actor kind %q", ErrInvalidEvent, event.ActorKind)
	}
	switch event.Source {
	case SourceAPI, SourceOperator, SourceStartup, SourceSystem, SourceUI:
	default:
		return nil, fmt.Errorf("%w: unknown source %q", ErrInvalidEvent, event.Source)
	}
	if !validActionBinding(event.ActionCode, event.ScopeKind, event.TargetKind) {
		return nil, fmt.Errorf("%w: action %q cannot target %s/%s", ErrInvalidEvent, event.ActionCode, event.ScopeKind, event.TargetKind)
	}
	if event.TargetID == nil || strings.TrimSpace(*event.TargetID) == "" {
		return nil, fmt.Errorf("%w: target_id is required", ErrInvalidEvent)
	}

	changes := append([]Change(nil), event.Changes...)
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].FieldPath < changes[j].FieldPath
	})
	if len(changes) == 0 {
		return nil, fmt.Errorf("%w: current action requires changes", ErrInvalidEvent)
	}
	for index, change := range changes {
		if strings.TrimSpace(change.FieldPath) == "" {
			return nil, fmt.Errorf("%w: field_path is required", ErrUnsafeChanges)
		}
		if index > 0 && changes[index-1].FieldPath == change.FieldPath {
			return nil, fmt.Errorf("%w: duplicate field_path %q", ErrUnsafeChanges, change.FieldPath)
		}
		switch change.ValueState {
		case ValueRedacted:
			if change.Before != nil || change.After != nil {
				return nil, fmt.Errorf("%w: redacted field %q must contain null values", ErrUnsafeChanges, change.FieldPath)
			}
		case ValueVisible:
			if containsForbiddenToken(change.FieldPath) || containsForbiddenJSONKey(change.Before) || containsForbiddenJSONKey(change.After) {
				return nil, fmt.Errorf("%w: visible field %q is secret-bearing", ErrUnsafeChanges, change.FieldPath)
			}
		default:
			return nil, fmt.Errorf("%w: unknown value_state %q", ErrUnsafeChanges, change.ValueState)
		}
	}
	return changes, nil
}

func validateRawProjectionPair(raw RawEvent, event Event) error {
	if raw.OccurredAt.IsZero() || !raw.OccurredAt.Equal(event.OccurredAt) {
		return fmt.Errorf("%w: raw and projected occurrence times must match", ErrInvalidEvent)
	}
	if !equalUUIDPointers(raw.ActorUserID, event.ActorUserID) {
		return fmt.Errorf("%w: raw and projected actor_user_id must match", ErrInvalidEvent)
	}
	switch event.ScopeKind {
	case ScopeDeployment:
		if raw.IncidentID != nil {
			return fmt.Errorf("%w: deployment projection cannot pair with incident raw event", ErrInvalidEvent)
		}
	case ScopeIncident:
		if !equalUUIDPointers(raw.IncidentID, event.ScopeID) {
			return fmt.Errorf("%w: raw incident_id must match projected scope_id", ErrInvalidEvent)
		}
	}
	return nil
}

func equalUUIDPointers(left *uuid.UUID, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func validActionBinding(actionCode string, scopeKind string, targetKind string) bool {
	for _, binding := range actionBindings[actionCode] {
		if binding.scopeKind == scopeKind && binding.targetKind == targetKind {
			return true
		}
	}
	return false
}

func containsForbiddenToken(value string) bool {
	normalized := strings.ToLower(value)
	for _, token := range forbiddenVisibleFieldTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func containsForbiddenObjectKey(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if containsForbiddenToken(key) || containsForbiddenObjectKey(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsForbiddenObjectKey(child) {
				return true
			}
		}
	}
	return false
}

func containsForbiddenJSONKey(value any) bool {
	if value == nil {
		return false
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return true
	}
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return true
	}
	return containsForbiddenObjectKey(decoded)
}

func encodeRawValue(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal raw administrative audit value: %w", err)
	}
	return encoded, nil
}
