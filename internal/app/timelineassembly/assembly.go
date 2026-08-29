package timelineassembly

import (
	"fmt"
	"reflect"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/app/entitymergeassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelinefactassembly"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	entityports "github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	entityfacts "github.com/JochiRaider/cartulary/internal/modules/entities/timelinefacts"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/collectionfacts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Dependencies struct {
	Postgres            postgres.DB
	ConflictTokens      conflicttokens.ConflictTokenCodec
	ConflictFields      conflicttokens.FieldResolver
	Revisions           *revisions.Appender
	Collaboration       collaboration.RecordChangedAppender
	EvidenceAttachments evidence.TimelineAttachmentContribution
	TimelineProjection  workbookprojection.Writer
	EntityProjection    entityports.MutationRows
	AssessmentRows      assessmentprojection.Rows
}

type Bundle struct {
	Facade             *timeline.Facade
	PerformanceFixture *timeline.PerformanceFixtureContribution
	EntityMentionStore *mentions.Store
	EntityMergeStore   *merge.Store
}

type composition struct {
	entityMentionStore *mentions.Store
	entityMergeStore   *merge.Store
	collaborators      timeline.Collaborators
}

func NewBundle(dependencies Dependencies) (*Bundle, error) {
	if err := validateDependencies(dependencies); err != nil {
		return nil, err
	}
	components, err := compose(dependencies)
	if err != nil {
		return nil, err
	}
	facade := timeline.NewFacade(dependencies.Postgres, components.collaborators, dependencies.ConflictTokens)
	return &Bundle{
		Facade:             facade,
		PerformanceFixture: timeline.NewPerformanceFixtureContribution(facade),
		EntityMentionStore: components.entityMentionStore,
		EntityMergeStore:   components.entityMergeStore,
	}, nil
}

func validateDependencies(dependencies Dependencies) error {
	if isNilDependency(dependencies.Postgres) {
		return fmt.Errorf("compose Timeline bundle: Postgres is required")
	}
	if isNilDependency(dependencies.Revisions) {
		return fmt.Errorf("compose Timeline bundle: Revisions appender is required")
	}
	if isNilDependency(dependencies.ConflictFields) {
		return fmt.Errorf("compose Timeline bundle: Revisions conflict field resolver is required")
	}
	if isNilDependency(dependencies.Collaboration) {
		return fmt.Errorf("compose Timeline bundle: Collaboration intent appender is required")
	}
	if isNilDependency(dependencies.EvidenceAttachments) {
		return fmt.Errorf("compose Timeline bundle: Evidence attachment contribution is required")
	}
	if isNilDependency(dependencies.TimelineProjection) {
		return fmt.Errorf("compose Timeline bundle: Timeline projection writer is required")
	}
	if isNilDependency(dependencies.EntityProjection) {
		return fmt.Errorf("compose Timeline bundle: Entities projection writer is required")
	}
	if isNilDependency(dependencies.AssessmentRows) {
		return fmt.Errorf("compose Timeline bundle: Assessment projection rows are required")
	}
	return nil
}

func isNilDependency(dependency any) bool {
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

// NewCollaborators composes Timeline's typed application boundary for focused
// facade tests that replace one collaborator without starting a server.
func NewCollaborators(dependencies Dependencies) (timeline.Collaborators, error) {
	if err := validateDependencies(dependencies); err != nil {
		return timeline.Collaborators{}, err
	}
	components, err := compose(dependencies)
	if err != nil {
		return timeline.Collaborators{}, err
	}
	return components.collaborators, nil
}

func compose(dependencies Dependencies) (composition, error) {
	conflictFields, err := dependencies.ConflictFields.ResolveViewSchema(timeline.TimelineViewSchemaID)
	if err != nil {
		return composition{}, fmt.Errorf("compose Timeline bundle: resolve Timeline conflict fields: %w", err)
	}
	recordsPort := recordAdapter{
		store:   records.NewStore(),
		targets: records.NewRouteTargetResolver(dependencies.Postgres),
	}
	collectionFacts := collectionfacts.New(
		entityfacts.Reader{},
		timelinefactassembly.NewLinkReader(),
		evidence.TimelineFactReader{},
	)
	timelineWriter := dependencies.TimelineProjection
	entityProjectionWriter := dependencies.EntityProjection
	mentionEffects := mentioneffects.NewProvider(recordsPort, collectionFacts, timelineWriter)
	linkStore := links.NewStore()
	entityMentionStore, err := mentions.NewStore(mentions.StoreDependencies{
		Postgres:      dependencies.Postgres,
		Revisions:     dependencies.Revisions,
		Links:         mentionLinkAdapter{store: linkStore},
		Projections:   entityProjectionWriter,
		Timeline:      mentionEffects,
		Collaboration: dependencies.Collaboration,
	})
	if err != nil {
		return composition{}, fmt.Errorf("compose Timeline bundle: %w", err)
	}
	collaborators := timeline.Collaborators{
		Core: timeline.CoreCollaborators{
			Idempotency:    idempotencyAdapter{store: authn.NewStore(dependencies.Postgres)},
			Incidents:      incidentAdapter{access: admission.NewChecker(dependencies.Postgres)},
			Records:        recordsPort,
			Revisions:      revisionAdapter{appender: dependencies.Revisions, reader: conflicttokens.NewRevisionWindowReader()},
			ConflictFields: conflictFields,
		},
		Collections: timeline.CollectionCollaborators{
			Links:    linkAdapter{store: linkStore},
			Mentions: mentionAdapter{store: entityMentionStore},
			Entities: entityAdapter{store: hostidentity.NewSourceFacts()},
			Evidence: evidenceAdapter{attachments: dependencies.EvidenceAttachments},
			Facts:    collectionFacts,
		},
		Commit: timeline.CommitCollaborators{
			Projection:       timelineWriter,
			EntityProjection: entityProjectionWriter,
			Collaboration:    collaborationAdapter{appender: dependencies.Collaboration},
		},
	}
	assessmentEffects, err := assessmentassembly.NewMergeEffects(
		dependencies.AssessmentRows,
		dependencies.Revisions,
	)
	if err != nil {
		return composition{}, fmt.Errorf("compose Timeline bundle: assessment merge effects: %w", err)
	}
	entityAssessmentEffects, err := entitymergeassembly.NewAssessmentEffects(assessmentEffects)
	if err != nil {
		return composition{}, fmt.Errorf("compose Timeline bundle: %w", err)
	}
	entityMergeStore, err := merge.NewStore(merge.StoreDependencies{
		Postgres:      dependencies.Postgres,
		Revisions:     dependencies.Revisions,
		HostIdentity:  hostidentity.NewMergeCapability(),
		Assessments:   entityAssessmentEffects,
		Mentions:      entityMentionStore,
		Links:         entitymergeassembly.NewLinkEffects(),
		Timeline:      mentionEffects,
		Projections:   entityProjectionWriter,
		Collaboration: dependencies.Collaboration,
	})
	if err != nil {
		return composition{}, fmt.Errorf("compose Timeline bundle: %w", err)
	}
	return composition{
		entityMentionStore: entityMentionStore,
		entityMergeStore:   entityMergeStore,
		collaborators:      collaborators,
	}, nil
}
