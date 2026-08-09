package merge

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidents.Access
	ports          entityStorePorts
}

type TimelineEffectsPort interface {
	LoadTimelineInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error)
	LoadRelationshipInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error)
	RefreshTimelineProjectionRowsTx(context.Context, pgx.Tx, []uuid.UUID) error
}

type StoreOption func(*entityStorePorts)

func WithTimelineEffects(effects TimelineEffectsPort) StoreOption {
	return func(ports *entityStorePorts) {
		ports.timeline = effects
	}
}

func WithCollaborationIntents(appender collaboration.IntentAppender) StoreOption {
	return func(ports *entityStorePorts) {
		ports.collaboration = appender
	}
}

func WithAssessmentEffects(effects *assessments.MergeEffects) StoreOption {
	return func(ports *entityStorePorts) {
		ports.assessments = entityAssessmentAdapter{effects: effects}
	}
}

func WithWorkbookProjection(writer workbookprojection.Writer) StoreOption {
	return func(ports *entityStorePorts) {
		ports.projections = writer
	}
}

func NewStore(pool postgres.DB, appender *revisions.Appender, options ...StoreOption) *Store {
	ports := newEntityStorePorts(pool, appender, nil)
	for _, option := range options {
		if option != nil {
			option(&ports)
		}
	}
	if ports.assessments == nil {
		panic("compose entity merge store: assessment effects are required")
	}
	if ports.projections == nil {
		panic("compose entity merge store: workbook projection writer is required")
	}
	ports.mentions = entityMentionAdapter{store: mentions.NewStore(
		pool,
		appender,
		mentions.WithWorkbookProjection(ports.projections),
	)}
	return &Store{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		incidentAccess: incidents.NewAccess(pool),
		ports:          ports,
	}
}
