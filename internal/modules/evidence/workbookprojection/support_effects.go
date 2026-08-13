package workbookprojection

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SupportProjectionEffectsTx interface {
	RefreshEvidenceAssociationEffects(
		context.Context,
		pgx.Tx,
		EvidenceAssociationEffectsInput,
	) (EvidenceAssociationEffectsResult, error)
}

type EvidenceAssociationEffectsInput struct {
	IncidentID uuid.UUID
	Subjects   []EvidenceAssociationSubject
}

type EvidenceAssociationSubject struct {
	RecordID   uuid.UUID
	RecordType string
}

type SupportChangeKind string

const (
	SupportChangePatch      SupportChangeKind = "patch"
	SupportChangeInvalidate SupportChangeKind = "invalidate"
)

type EvidenceAssociationEffectsResult struct {
	Changes []EvidenceSupportRowChange
}

type EvidenceSupportRowChange struct {
	RecordID      uuid.UUID
	RowVersion    int64
	AffectedViews []EvidenceAffectedViewChange
}

type EvidenceAffectedViewChange struct {
	ViewSchemaID     string
	ChangeKind       SupportChangeKind
	ChangedFieldKeys []string
	Patch            *EvidenceViewRowPatch
}

type EvidenceViewRowPatch struct {
	RecordID    uuid.UUID
	RowVersion  int64
	Cells       map[string]any
	GroupValues map[string]any
}
