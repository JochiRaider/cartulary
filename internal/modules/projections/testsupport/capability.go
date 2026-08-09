// Package testsupport exposes named, typed projection capabilities to shared
// application tests without exposing production assembly aggregates or
// concrete Projections runtime types.
package testsupport

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	projectionruntime "github.com/JochiRaider/cartulary/internal/modules/projections/internal/runtime"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func ApplyAssessmentFixtureMutationTx(
	ctx context.Context,
	db postgres.DB,
	tx pgx.Tx,
	mutation assessmentprojection.ProjectionMutation,
) error {
	return projectionruntime.NewAssessmentRows(db, nil).ApplyAssessmentMutationTx(ctx, tx, mutation)
}

func NewIndicatorRows(
	db postgres.DB,
	source indicatorprojection.SourceReader,
) indicatorprojection.Rows {
	return projectionruntime.NewIndicatorRows(db, source)
}

type Dependencies struct {
	TimelineRebuilder  timelineprojection.Rebuilder
	EntityRebuilder    entityprojection.Rebuilder
	IndicatorRebuilder indicatorprojection.Rebuilder
	IndicatorRows      indicatorprojection.Rows
	EvidenceRows       evidenceprojection.Rows
}

type Capability struct {
	timelineRebuilder  timelineprojection.Rebuilder
	entityRebuilder    entityprojection.Rebuilder
	indicatorRebuilder indicatorprojection.Rebuilder
	indicatorRows      indicatorprojection.Rows
	evidenceRows       evidenceprojection.Rows
}

func New(dependencies Dependencies) *Capability {
	return &Capability{
		timelineRebuilder:  dependencies.TimelineRebuilder,
		entityRebuilder:    dependencies.EntityRebuilder,
		indicatorRebuilder: dependencies.IndicatorRebuilder,
		indicatorRows:      dependencies.IndicatorRows,
		evidenceRows:       dependencies.EvidenceRows,
	}
}

func (capability *Capability) RebuildTimeline(ctx context.Context, incidentID uuid.UUID) error {
	if capability == nil || capability.timelineRebuilder == nil {
		return errors.New("timeline projection rebuild capability is unavailable")
	}
	return capability.timelineRebuilder.RebuildTimeline(ctx, incidentID)
}

func (capability *Capability) RebuildHosts(ctx context.Context, incidentID uuid.UUID) error {
	if capability == nil || capability.entityRebuilder == nil {
		return errors.New("host projection rebuild capability is unavailable")
	}
	return capability.entityRebuilder.RebuildHosts(ctx, incidentID)
}

func (capability *Capability) RebuildIdentities(ctx context.Context, incidentID uuid.UUID) error {
	if capability == nil || capability.entityRebuilder == nil {
		return errors.New("identity projection rebuild capability is unavailable")
	}
	return capability.entityRebuilder.RebuildIdentities(ctx, incidentID)
}

func (capability *Capability) RebuildIndicators(ctx context.Context, incidentID uuid.UUID) error {
	if capability == nil || capability.indicatorRebuilder == nil {
		return errors.New("indicator projection rebuild capability is unavailable")
	}
	return capability.indicatorRebuilder.RebuildIndicators(ctx, incidentID)
}

func (capability *Capability) IndicatorProjectionPort() indicatorprojection.Rows {
	if capability == nil {
		return nil
	}
	return capability.indicatorRows
}

func (capability *Capability) EvidencePort() evidenceprojection.Rows {
	if capability == nil {
		return nil
	}
	return capability.evidenceRows
}
