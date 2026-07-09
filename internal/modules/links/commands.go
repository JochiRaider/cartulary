package links

import (
	"time"

	"github.com/google/uuid"
)

type LinkType string

func (t LinkType) String() string {
	return string(t)
}

type LinkProvenance string

func (p LinkProvenance) String() string {
	return string(p)
}

type UpsertLinkCommand struct {
	IncidentID  uuid.UUID
	SrcRecordID uuid.UUID
	DstRecordID uuid.UUID
	LinkType    LinkType
	Provenance  LinkProvenance
	Confidence  *int
	OwnerUserID uuid.UUID
	Now         time.Time
}

type SyncFieldReferenceCommand struct {
	IncidentID  uuid.UUID
	SrcRecordID uuid.UUID
	TargetID    *uuid.UUID
	FieldKey    string
	LinkType    LinkType
	ActorUserID uuid.UUID
	Now         time.Time
}

type InsertSupersedesCommand struct {
	IncidentID          uuid.UUID
	ReplacementRecordID uuid.UUID
	SupersededRecordID  uuid.UUID
	OwnerUserID         uuid.UUID
	Now                 time.Time
}
