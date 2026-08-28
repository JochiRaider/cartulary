package links

import (
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/mutationvalue"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/vocabulary"
)

type LinkType = vocabulary.LinkType

type LinkProvenance = vocabulary.LinkProvenance

type Mutation = mutationvalue.Value

func ParseLinkType(value string) (LinkType, error) {
	return vocabulary.ParseLinkType(value)
}

func ParseLinkProvenance(value string) (LinkProvenance, error) {
	return vocabulary.ParseLinkProvenance(value)
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

type TombstoneActiveLinkCommand struct {
	IncidentID  uuid.UUID
	SrcRecordID uuid.UUID
	DstRecordID uuid.UUID
	LinkType    LinkType
	ActorUserID uuid.UUID
	Now         time.Time
}

type RecordLinkCommandResult struct {
	RecordLinkID uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     LinkType
	mutation     *Mutation
}

func (result RecordLinkCommandResult) Mutation() (Mutation, bool) {
	if result.mutation == nil {
		return Mutation{}, false
	}
	copy := mutationvalue.Copy([]Mutation{*result.mutation})
	return copy[0], true
}
