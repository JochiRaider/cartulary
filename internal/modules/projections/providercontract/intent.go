package providercontract

import (
	"fmt"
	"strings"
)

// SurfaceIntent describes source-owned query meaning without exposing any
// physical table, join, expression, discriminator SQL, or scan strategy.
type SurfaceIntent struct {
	ViewSchemaID          string
	FieldKeys             []string
	CanonicalSourceFilter *SourceFilterIntent
}

type SourceFilterIntent struct {
	Kind  string
	Field string
	Value string
}

func (intent SurfaceIntent) Clone() SurfaceIntent {
	intent.FieldKeys = cloneStrings(intent.FieldKeys)
	if intent.CanonicalSourceFilter != nil {
		filter := *intent.CanonicalSourceFilter
		intent.CanonicalSourceFilter = &filter
	}
	return intent
}

type Contribution struct {
	sourceOwnerModule string
	descriptors       []ProviderDescriptor
	intents           []SurfaceIntent
}

func NewContribution(
	sourceOwnerModule string,
	descriptors []ProviderDescriptor,
	intents []SurfaceIntent,
) (Contribution, error) {
	if strings.TrimSpace(sourceOwnerModule) == "" {
		return Contribution{}, fmt.Errorf("projection contribution source owner is required")
	}
	if len(descriptors) == 0 {
		return Contribution{}, fmt.Errorf("projection contribution %q has no descriptors", sourceOwnerModule)
	}
	contribution := Contribution{
		sourceOwnerModule: sourceOwnerModule,
		descriptors:       make([]ProviderDescriptor, 0, len(descriptors)),
		intents:           make([]SurfaceIntent, 0, len(intents)),
	}
	for _, descriptor := range descriptors {
		contribution.descriptors = append(contribution.descriptors, descriptor.Clone())
	}
	for _, intent := range intents {
		contribution.intents = append(contribution.intents, intent.Clone())
	}
	return contribution, nil
}

func (contribution Contribution) IsZero() bool {
	return contribution.sourceOwnerModule == "" || len(contribution.descriptors) == 0
}

func (contribution Contribution) SourceOwnerModule() string {
	return contribution.sourceOwnerModule
}

func (contribution Contribution) Descriptors() []ProviderDescriptor {
	result := make([]ProviderDescriptor, 0, len(contribution.descriptors))
	for _, descriptor := range contribution.descriptors {
		result = append(result, descriptor.Clone())
	}
	return result
}

func (contribution Contribution) SurfaceIntents() []SurfaceIntent {
	result := make([]SurfaceIntent, 0, len(contribution.intents))
	for _, intent := range contribution.intents {
		result = append(result, intent.Clone())
	}
	return result
}
