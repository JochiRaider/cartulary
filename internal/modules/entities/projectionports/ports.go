package projectionports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// MutationRows is the complete typed Entities mutation boundary for derived
// workbook rows. Transactions remain owned by the caller.
type MutationRows interface {
	RefreshHostTx(context.Context, pgx.Tx, uuid.UUID) error
	RefreshIdentityTx(context.Context, pgx.Tx, uuid.UUID) error
	DeleteHostTx(context.Context, pgx.Tx, uuid.UUID) error
	DeleteIdentityTx(context.Context, pgx.Tx, uuid.UUID) error
}

type HostQueryProjection struct {
	RecordID          uuid.UUID
	HostState         string
	LinkedEventCount  int
	EvidenceCount     int
	Location          *string
	OSPlatform        *string
	BusinessOwner     *string
	Criticality       *string
	ContainmentStatus *string
}

type IdentityQueryProjection struct {
	RecordID         uuid.UUID
	IdentityState    string
	LinkedEventCount int
	EvidenceCount    int
	PrivilegeLevel   *string
	MFAState         *string
	ResetStatus      *string
}

type QueryReader interface {
	SelectHostQueryProjections(context.Context, uuid.UUID, viewschema.QueryMeta, querypage.Window) ([]HostQueryProjection, error)
	SelectIdentityQueryProjections(context.Context, uuid.UUID, viewschema.QueryMeta, querypage.Window) ([]IdentityQueryProjection, error)
}

type DerivedFact struct {
	RecordID     uuid.UUID
	RecordType   string
	ContentClass string
	Value        map[string]any
}

type ReportingReader interface {
	CollectHostDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]DerivedFact, error)
	CollectIdentityDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]DerivedFact, error)
}
