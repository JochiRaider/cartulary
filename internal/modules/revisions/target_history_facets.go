package revisions

import (
	"fmt"

	"github.com/google/uuid"
)

type HistoryAddressability string

const (
	HistorySingleEntry                HistoryAddressability = "single_history_entry"
	HistoryNotIndividuallyAddressable HistoryAddressability = "not_individually_addressable"
)

type StoredMutation struct {
	TargetKind    string
	TargetID      string
	OperationKind string
	BeforeValue   map[string]any
	AfterValue    map[string]any
}

type HistoryTargetDescription struct {
	TargetKind            string
	TargetID              string
	LogicalItemIdentity   string
	HistoryRecordIDs      []uuid.UUID
	HistoryEntryRecordIDs []uuid.UUID
}

type historySemanticsContract struct {
	directTargetID bool
	recordIDFields []string
	addressability HistoryAddressability
}

type HistoryFacet struct {
	contract historySemanticsContract
}

func NewDirectRecordHistoryFacet(addressability HistoryAddressability) HistoryFacet {
	return HistoryFacet{contract: historySemanticsContract{directTargetID: true, addressability: addressability}}
}

func NewFieldAssociationHistoryFacet(recordIDFields []string, addressability HistoryAddressability) HistoryFacet {
	return HistoryFacet{contract: historySemanticsContract{
		recordIDFields: append([]string(nil), recordIDFields...),
		addressability: addressability,
	}}
}

func (facet HistoryFacet) historyContract() historySemanticsContract {
	return historySemanticsContract{
		directTargetID: facet.contract.directTargetID,
		recordIDFields: append([]string(nil), facet.contract.recordIDFields...),
		addressability: facet.contract.addressability,
	}
}

func (facet HistoryFacet) isZero() bool {
	return !facet.contract.directTargetID && len(facet.contract.recordIDFields) == 0 && facet.contract.addressability == ""
}

func (facet HistoryFacet) DescribeMutation(mutation StoredMutation) (HistoryTargetDescription, error) {
	contract := facet.historyContract()
	ids := make([]uuid.UUID, 0)
	if contract.directTargetID {
		recordID, err := uuid.Parse(mutation.TargetID)
		if err != nil || recordID == uuid.Nil {
			return HistoryTargetDescription{}, fmt.Errorf("%w: target %q id %q", ErrInvalidTargetSemantics, mutation.TargetKind, mutation.TargetID)
		}
		ids = append(ids, recordID)
	} else {
		for _, value := range []map[string]any{mutation.BeforeValue, mutation.AfterValue} {
			for _, field := range contract.recordIDFields {
				text, ok := value[field].(string)
				if !ok || text == "" {
					continue
				}
				recordID, err := uuid.Parse(text)
				if err != nil || recordID == uuid.Nil {
					return HistoryTargetDescription{}, fmt.Errorf("%w: target %q field %q", ErrInvalidTargetSemantics, mutation.TargetKind, field)
				}
				ids = append(ids, recordID)
			}
		}
	}
	ids = canonicalRecordIDs(ids)
	if len(ids) == 0 {
		return HistoryTargetDescription{}, fmt.Errorf("%w: target %q has no history association", ErrInvalidTargetSemantics, mutation.TargetKind)
	}
	entryIDs := append(make([]uuid.UUID, 0, len(ids)), ids...)
	if contract.addressability == HistoryNotIndividuallyAddressable {
		entryIDs = []uuid.UUID{}
	}
	return HistoryTargetDescription{
		TargetKind:            mutation.TargetKind,
		TargetID:              mutation.TargetID,
		LogicalItemIdentity:   mutation.TargetKind + ":" + mutation.TargetID,
		HistoryRecordIDs:      ids,
		HistoryEntryRecordIDs: entryIDs,
	}, nil
}
