package merge

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess *admission.Checker
	hostIdentity   *hostidentity.MergeCapability
	ports          entityStorePorts
}

type TimelineEffectsPort interface {
	LoadTimelineInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error)
	LoadRelationshipInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error)
	RefreshTimelineProjectionRowsTx(context.Context, pgx.Tx, []uuid.UUID) error
}

type MentionEffectsPort interface {
	RepointMergedMentionsTx(context.Context, pgx.Tx, mentions.RepointMergedMentionsCommand) (mentions.RepointMergedMentionsResult, error)
}

type StoreDependencies struct {
	Postgres      postgres.DB
	Revisions     *revisions.Appender
	HostIdentity  *hostidentity.MergeCapability
	Assessments   AssessmentEffectsPort
	Mentions      MentionEffectsPort
	Links         LinkEffectsPort
	Timeline      TimelineEffectsPort
	Projections   workbookprojection.Writer
	Collaboration collaboration.IntentAppender
}

func NewStore(dependencies StoreDependencies) (*Store, error) {
	for _, dependency := range []struct {
		name  string
		value any
	}{
		{name: "Postgres", value: dependencies.Postgres},
		{name: "Revisions", value: dependencies.Revisions},
		{name: "HostIdentity", value: dependencies.HostIdentity},
		{name: "Assessments", value: dependencies.Assessments},
		{name: "Mentions", value: dependencies.Mentions},
		{name: "Links", value: dependencies.Links},
		{name: "Timeline", value: dependencies.Timeline},
		{name: "Projections", value: dependencies.Projections},
		{name: "Collaboration", value: dependencies.Collaboration},
	} {
		if isNilStoreDependency(dependency.value) {
			return nil, fmt.Errorf("compose Merge store: %s is required", dependency.name)
		}
	}
	ports := newEntityStorePorts(dependencies.Postgres, dependencies.Revisions, dependencies.Projections)
	ports.assessments = dependencies.Assessments
	ports.mentions = entityMentionAdapter{store: dependencies.Mentions}
	ports.links = dependencies.Links
	ports.timeline = dependencies.Timeline
	ports.collaboration = dependencies.Collaboration
	return &Store{
		pool:           dependencies.Postgres,
		authStore:      authn.NewStore(dependencies.Postgres),
		incidentAccess: admission.NewChecker(dependencies.Postgres),
		hostIdentity:   dependencies.HostIdentity,
		ports:          ports,
	}, nil
}

func isNilStoreDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
