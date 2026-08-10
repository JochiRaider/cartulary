package artifacts

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

type IncidentStateCapability interface {
	EnsureOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type MemberReferenceCapability interface {
	ValidateIncidentMemberUserTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, string) error
}

type RecordEnvelopeCapability interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (records.Envelope, error)
}

type LinkCapability interface {
	ValidatePartyRefCollectionTx(context.Context, pgx.Tx, links.PartyRefCollectionValidation) error
	ValidateRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionValidation) error
	ValidateTagCollectionTx(context.Context, pgx.Tx, links.TagCollectionValidation) error
	ApplyPartyRefCollectionWithMutationValuesTx(context.Context, pgx.Tx, links.PartyRefCollectionCommand) (links.CollectionMutationResult, error)
	ApplyRecordRefCollectionWithMutationValuesTx(context.Context, pgx.Tx, links.RecordRefCollectionCommand) (links.CollectionMutationResult, error)
	ApplyTagCollectionWithMutationValuesTx(context.Context, pgx.Tx, links.TagCollectionCommand) (links.CollectionMutationResult, error)
	InsertLinkedNoteReferenceTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) (links.RecordLink, bool, error)
	LoadRecordLinkValueTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
}

type RevisionCapability interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendNonRowMutationTx(context.Context, pgx.Tx, revisions.AppendNonRowMutationParams) error
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendRecordRevisionAndIntentTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
	LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicts.RevisionWindowRow, error)
}

type artifactSourceMutationPort interface {
	insertRowTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, createParams, time.Time) error
	applyDirectChangeTx(context.Context, pgx.Tx, uuid.UUID, string, FieldValue, time.Time) (bool, error)
	applyHandoffRiskRefPayloadTx(context.Context, pgx.Tx, RecordEnvelopeCapability, uuid.UUID, uuid.UUID, uuid.UUID, riskRefActionPayload, time.Time) (bool, error)
	normalizeFindingLifecycleTx(context.Context, pgx.Tx, uuid.UUID, time.Time) (bool, error)
	touchRowTx(context.Context, pgx.Tx, uuid.UUID, time.Time) error
}

// MutationDependencies is assembled at the application composition root. The
// Artifacts facade never constructs peer-owner or platform stores.
type MutationDependencies struct {
	IncidentState        IncidentStateCapability
	MemberReferences     MemberReferenceCapability
	Idempotency          IdempotencyCapability
	RecordEnvelopes      RecordEnvelopeCapability
	Links                LinkCapability
	Projections          artifactprojection.Rows
	Revisions            RevisionCapability
	ConflictFields       conflicts.FieldResolver
	KeepSavedIdempotency conflicts.IdempotencyPort
}

func (d MutationDependencies) validate() error {
	required := []struct {
		name  string
		value any
	}{
		{name: "Incident admission", value: d.IncidentState},
		{name: "Member validation", value: d.MemberReferences},
		{name: "Route idempotency", value: d.Idempotency},
		{name: "Record envelopes", value: d.RecordEnvelopes},
		{name: "Links", value: d.Links},
		{name: "Projections", value: d.Projections},
		{name: "Revisions/history", value: d.Revisions},
		{name: "Conflict fields", value: d.ConflictFields},
		{name: "Keep-saved idempotency", value: d.KeepSavedIdempotency},
	}
	for _, dependency := range required {
		if dependency.value == nil {
			return fmt.Errorf("artifacts mutation dependencies: %s is required", dependency.name)
		}
	}
	return nil
}

type memberReferenceValidator struct{}

func NewMemberReferenceCapability() MemberReferenceCapability {
	return memberReferenceValidator{}
}

func (memberReferenceValidator) ValidateIncidentMemberUserTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	userID uuid.UUID,
	field string,
) error {
	return validateIncidentMemberUserTx(ctx, tx, incidentID, userID, field)
}

func validateIncidentMemberUserTx(ctx context.Context, tx pgx.Tx, incidentID, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM users u
    JOIN incident_memberships m ON m.user_id = u.id
   WHERE u.id = $1
     AND u.is_active = true
     AND m.incident_id = $2
)`, userID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate user: %w", err)
	}
	if !exists {
		return &ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

type artifactRecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func (f *MutationFacade) loadArtifactRecordMetaForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (artifactRecordMeta, error) {
	envelope, err := f.recordEnvelopes.LoadEnvelopeTx(ctx, tx, recordID, true)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return artifactRecordMeta{}, pgx.ErrNoRows
	}
	if err != nil {
		return artifactRecordMeta{}, err
	}
	if envelope.DeletedAt != nil {
		return artifactRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return artifactRecordMeta{
		IncidentID: envelope.IncidentID,
		RecordType: envelope.RecordType,
		RowVersion: envelope.RowVersion,
	}, nil
}

func validateArtifactViewRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, viewSchemaID string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM artifacts
     WHERE record_id = $1
       AND artifact_type = $2
)
`, recordID, artifactTypeForView(viewSchemaID)).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	return nil
}

func touchesArtifactField(changes []PatchChange, field string) bool {
	for _, change := range changes {
		if change.FieldKey == field {
			return true
		}
	}
	return false
}

func changedFieldKeys(before map[string]any, after map[string]any) []string {
	afterCells, _ := after["cells"].(map[string]any)
	beforeCells := map[string]any{}
	if before != nil {
		beforeCells, _ = before["cells"].(map[string]any)
	}
	keys := make([]string, 0)
	for fieldKey, afterValue := range afterCells {
		if beforeValue, ok := beforeCells[fieldKey]; !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			keys = append(keys, fieldKey)
		}
	}
	slices.Sort(keys)
	return keys
}

func workbookVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func rowVersionFromCanonicalRow(row map[string]any) int64 {
	switch value := row["row_version"].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return 0
	}
}

func contextualLinkFacts(sourceRecordID *uuid.UUID) *ContextualLink {
	if sourceRecordID == nil {
		return nil
	}
	return &ContextualLink{SourceRecordID: *sourceRecordID, LinkType: "references_artifact"}
}
