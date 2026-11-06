package networkflow

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/google/uuid"
	"io"
)

type committedPortabilitySources interface {
	ReadSource(context.Context, uuid.UUID) (imports.CommittedSource, error)
	OpenSource(context.Context, uuid.UUID) (imports.CommittedSource, io.ReadCloser, error)
}
type portableIndicatorReferences interface {
	FindIndicator(context.Context, uuid.UUID) (indicators.IndicatorReference, error)
}

type portabilityDependencies struct {
	incidentID uuid.UUID
	sources    committedPortabilitySources
	indicators portableIndicatorReferences
}

func (r *portabilityDependencies) ProfileScope() string     { return ProfileID }
func (r *portabilityDependencies) IncidentScope() uuid.UUID { return r.incidentID }

// Assembly supplies owner capabilities after shared admission. No owner receives
// another owner's physical storage, archive paths or transaction lifecycle.
func NewPortabilityDependencies(incidentID uuid.UUID, sources committedPortabilitySources, indicatorReferences portableIndicatorReferences) (incidentbundles.ImportDependencyAccess, error) {
	if incidentID == uuid.Nil || sources == nil || indicatorReferences == nil {
		return nil, incidentbundles.ErrPortabilityUnavailable
	}
	return &portabilityDependencies{incidentID: incidentID, sources: sources, indicators: indicatorReferences}, nil
}
func (r *portabilityDependencies) CommittedSources() committedPortabilitySources { return r.sources }
func (r *portabilityDependencies) IndicatorReferences() portableIndicatorReferences {
	return r.indicators
}
