package sourceport

import (
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidActorCatalog = errors.New("incident bundle actor catalog is invalid")

type ActorDescriptor struct {
	SourceActorID       string
	DisplayName         string
	EmailHint           string
	ProviderSubjectHint string
}

type ActorCatalog struct {
	bySourceActorID map[string]ActorDescriptor
}

func NewActorCatalog(descriptors []ActorDescriptor) (ActorCatalog, error) {
	catalog := ActorCatalog{bySourceActorID: make(map[string]ActorDescriptor, len(descriptors))}
	for _, descriptor := range descriptors {
		descriptor.SourceActorID = strings.TrimSpace(descriptor.SourceActorID)
		if _, err := uuid.Parse(descriptor.SourceActorID); err != nil {
			return ActorCatalog{}, ErrInvalidActorCatalog
		}
		if _, duplicate := catalog.bySourceActorID[descriptor.SourceActorID]; duplicate {
			return ActorCatalog{}, ErrInvalidActorCatalog
		}
		catalog.bySourceActorID[descriptor.SourceActorID] = descriptor
	}
	return catalog, nil
}

func (c ActorCatalog) Lookup(sourceActorID string) (ActorDescriptor, bool) {
	descriptor, ok := c.bySourceActorID[sourceActorID]
	return descriptor, ok
}

func (c ActorCatalog) Len() int {
	return len(c.bySourceActorID)
}

// Descriptors returns the admitted descriptors in source actor order.
func (c ActorCatalog) Descriptors() []ActorDescriptor {
	result := make([]ActorDescriptor, 0, len(c.bySourceActorID))
	for _, descriptor := range c.bySourceActorID {
		result = append(result, descriptor)
	}
	slices.SortFunc(result, func(left, right ActorDescriptor) int { return strings.Compare(left.SourceActorID, right.SourceActorID) })
	return result
}
