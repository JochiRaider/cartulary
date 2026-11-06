package subtypepresence

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidCatalog     = errors.New("records subtype-presence catalog is invalid")
	ErrIncompleteBindings = errors.New("records subtype bindings are incomplete")
)

type RecordType string

const (
	RecordTypeTimelineEvent RecordType = "timeline_event"
	RecordTypeHost          RecordType = "host"
	RecordTypeIdentity      RecordType = "identity"
	RecordTypeParty         RecordType = "party"
	RecordTypeIndicator     RecordType = "indicator"
	RecordTypeArtifact      RecordType = "artifact"
	RecordTypeTaskRequest   RecordType = "task_request"
	RecordTypeDecision      RecordType = "decision"
	RecordTypeEvidence      RecordType = "evidence"
	RecordTypeAssessment    RecordType = "assessment"
)

type Binding struct {
	RecordID   uuid.UUID
	IncidentID uuid.UUID
	RecordType RecordType
}

type Source interface {
	SupportedRecordTypes() []RecordType
	VisitSubtypeBindingsTx(context.Context, pgx.Tx, uuid.UUID, func(Binding) error) error
}

type Contribution struct {
	FamilyID string
	Source   Source
}

type catalogSource struct {
	familyID string
	source   Source
}

type Catalog struct {
	sources []catalogSource
}

var expectedFamilyByType = map[RecordType]string{
	RecordTypeTimelineEvent: "timeline",
	RecordTypeHost:          "entities",
	RecordTypeIdentity:      "entities",
	RecordTypeParty:         "parties",
	RecordTypeIndicator:     "indicators",
	RecordTypeArtifact:      "artifacts",
	RecordTypeTaskRequest:   "tasks_decisions",
	RecordTypeDecision:      "tasks_decisions",
	RecordTypeEvidence:      "evidence",
	RecordTypeAssessment:    "assessments",
}

func NewCatalog(contributions []Contribution) (*Catalog, error) {
	if len(contributions) == 0 {
		return nil, ErrInvalidCatalog
	}
	seenFamilies := make(map[string]struct{}, len(contributions))
	seenTypes := make(map[RecordType]struct{}, len(expectedFamilyByType))
	sources := make([]catalogSource, 0, len(contributions))
	for _, contribution := range contributions {
		if _, admitted := expectedTypesForFamily(contribution.FamilyID); !admitted ||
			isNilSource(contribution.Source) {
			return nil, ErrInvalidCatalog
		}
		if _, duplicate := seenFamilies[contribution.FamilyID]; duplicate {
			return nil, ErrInvalidCatalog
		}
		seenFamilies[contribution.FamilyID] = struct{}{}
		supported := contribution.Source.SupportedRecordTypes()
		if len(supported) == 0 {
			return nil, ErrInvalidCatalog
		}
		for _, recordType := range supported {
			expectedFamily, admitted := expectedFamilyByType[recordType]
			if !admitted || expectedFamily != contribution.FamilyID {
				return nil, ErrInvalidCatalog
			}
			if _, duplicate := seenTypes[recordType]; duplicate {
				return nil, ErrInvalidCatalog
			}
			seenTypes[recordType] = struct{}{}
		}
		sources = append(sources, catalogSource{
			familyID: contribution.FamilyID,
			source:   contribution.Source,
		})
	}
	if len(seenTypes) != len(expectedFamilyByType) {
		return nil, ErrInvalidCatalog
	}
	for recordType := range expectedFamilyByType {
		if _, present := seenTypes[recordType]; !present {
			return nil, ErrInvalidCatalog
		}
	}
	sort.Slice(sources, func(left, right int) bool {
		return sources[left].familyID < sources[right].familyID
	})
	return &Catalog{sources: sources}, nil
}

func (c *Catalog) ValidateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, expected Summary) error {
	if c == nil || tx == nil || incidentID == uuid.Nil || expected.incidentID != incidentID {
		return ErrInvalidCatalog
	}
	actual := NewInventory(incidentID)
	for _, entry := range c.sources {
		err := entry.source.VisitSubtypeBindingsTx(ctx, tx, incidentID, func(binding Binding) error {
			if expectedFamilyByType[binding.RecordType] != entry.familyID {
				return ErrIncompleteBindings
			}
			return actual.Add(binding)
		})
		if err != nil {
			return err
		}
	}
	if !expected.Equal(actual.Summary()) {
		return ErrIncompleteBindings
	}
	return nil
}

func expectedTypesForFamily(familyID string) ([]RecordType, bool) {
	types := make([]RecordType, 0, 2)
	for recordType, expectedFamily := range expectedFamilyByType {
		if expectedFamily == familyID {
			types = append(types, recordType)
		}
	}
	sort.Slice(types, func(left, right int) bool {
		return types[left] < types[right]
	})
	return types, len(types) > 0
}

func isNilSource(source Source) bool {
	if source == nil {
		return true
	}
	value := reflect.ValueOf(source)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (c *Catalog) ValidateContract() error {
	if c == nil || len(c.sources) == 0 {
		return ErrInvalidCatalog
	}
	seen := make(map[RecordType]struct{}, len(expectedFamilyByType))
	for _, entry := range c.sources {
		for _, recordType := range entry.source.SupportedRecordTypes() {
			if expectedFamilyByType[recordType] != entry.familyID {
				return fmt.Errorf("%w: invalid source contribution", ErrInvalidCatalog)
			}
			seen[recordType] = struct{}{}
		}
	}
	if len(seen) != len(expectedFamilyByType) {
		return ErrInvalidCatalog
	}
	return nil
}
