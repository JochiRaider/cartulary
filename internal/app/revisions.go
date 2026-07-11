package app

import (
	"fmt"

	artifactsdelete "github.com/JochiRaider/cartulary/internal/modules/artifacts/deleterestore"
	artifactrollback "github.com/JochiRaider/cartulary/internal/modules/artifacts/rollbackprovider"
	assessmentsdelete "github.com/JochiRaider/cartulary/internal/modules/assessments/deleterestore"
	assessmentrollback "github.com/JochiRaider/cartulary/internal/modules/assessments/rollbackprovider"
	entitiesdelete "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/deleterestore"
	entityrollback "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/rollbackprovider"
	mentionrollback "github.com/JochiRaider/cartulary/internal/modules/entities/mentions/rollbackprovider"
	evidencedelete "github.com/JochiRaider/cartulary/internal/modules/evidence/deleterestore"
	evidencerollback "github.com/JochiRaider/cartulary/internal/modules/evidence/rollbackprovider"
	indicatorsdelete "github.com/JochiRaider/cartulary/internal/modules/indicators/deleterestore"
	indicatorrollback "github.com/JochiRaider/cartulary/internal/modules/indicators/rollbackprovider"
	linkrevisionprovider "github.com/JochiRaider/cartulary/internal/modules/links/revisionprovider"
	partiesdelete "github.com/JochiRaider/cartulary/internal/modules/parties/deleterestore"
	partyrollback "github.com/JochiRaider/cartulary/internal/modules/parties/rollbackprovider"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	tasksdecisionsdelete "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/deleterestore"
	tasksdecisionsrollback "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/rollbackprovider"
	timelinedelete "github.com/JochiRaider/cartulary/internal/modules/timeline/deleterestore"
	timelinerollback "github.com/JochiRaider/cartulary/internal/modules/timeline/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func NewRevisionsCommandService(db postgres.DB, attributionResolver revisions.ImportedAttributionResolver) (*revisions.CommandService, error) {
	recordTypes := []string{
		"artifact",
		"assessment",
		"decision",
		"evidence",
		"host",
		"identity",
		"indicator",
		"party",
		"task_request",
		"timeline_event",
	}
	deleteRestoreProviders, err := revisions.NewDeleteRestoreProviderCatalog(recordTypes,
		revisions.DeleteRestoreProviderRegistration{RecordType: "timeline_event", Provider: timelinedelete.NewProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "host", Provider: entitiesdelete.HostProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "identity", Provider: entitiesdelete.IdentityProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "party", Provider: partiesdelete.NewProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "indicator", Provider: indicatorsdelete.NewProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "artifact", Provider: artifactsdelete.NewProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "task_request", Provider: tasksdecisionsdelete.TaskRequestProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "decision", Provider: tasksdecisionsdelete.DecisionProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "evidence", Provider: evidencedelete.NewProvider()},
		revisions.DeleteRestoreProviderRegistration{RecordType: "assessment", Provider: assessmentsdelete.NewProvider()},
	)
	if err != nil {
		return nil, fmt.Errorf("compose revisions delete/restore providers: %w", err)
	}
	rowRollbackProviders, err := revisions.NewRowProviderCatalog(recordTypes,
		revisions.RowProviderRegistration{RecordType: "timeline_event", Provider: timelinerollback.NewTimelineProvider()},
		revisions.RowProviderRegistration{RecordType: "host", Provider: entityrollback.NewHostProvider()},
		revisions.RowProviderRegistration{RecordType: "party", Provider: partyrollback.NewPartyProvider()},
		revisions.RowProviderRegistration{RecordType: "identity", Provider: entityrollback.NewIdentityProvider()},
		revisions.RowProviderRegistration{RecordType: "evidence", Provider: evidencerollback.NewProvider()},
		revisions.RowProviderRegistration{RecordType: "indicator", Provider: indicatorrollback.NewProvider()},
		revisions.RowProviderRegistration{RecordType: "assessment", Provider: assessmentrollback.NewProvider()},
		revisions.RowProviderRegistration{RecordType: "task_request", Provider: tasksdecisionsrollback.NewTaskRequestProvider()},
		revisions.RowProviderRegistration{RecordType: "decision", Provider: tasksdecisionsrollback.NewDecisionProvider()},
		revisions.RowProviderRegistration{RecordType: "artifact", Provider: artifactrollback.NewProvider()},
	)
	if err != nil {
		return nil, fmt.Errorf("compose revisions row rollback providers: %w", err)
	}
	linkProvider := linkrevisionprovider.NewProvider()
	entityCollectionProvider := entityrollback.NewCollectionProvider()
	indicatorChildProvider := indicatorrollback.NewChildProvider()
	nonRowRollbackProviders, err := revisions.NewNonRowProviderCatalog([]string{
		"entity_alias",
		"entity_mention",
		"entity_preserved_identifier",
		"indicator_observation",
		"indicator_state_interval",
		"record_link",
		"record_tag",
	},
		revisions.NonRowProviderRegistration{TargetKind: "record_link", Provider: linkProvider},
		revisions.NonRowProviderRegistration{TargetKind: "record_tag", Provider: linkProvider},
		revisions.NonRowProviderRegistration{TargetKind: "entity_mention", Provider: mentionrollback.NewMentionProvider()},
		revisions.NonRowProviderRegistration{TargetKind: "entity_alias", Provider: entityCollectionProvider},
		revisions.NonRowProviderRegistration{TargetKind: "entity_preserved_identifier", Provider: entityCollectionProvider},
		revisions.NonRowProviderRegistration{TargetKind: "indicator_observation", Provider: indicatorChildProvider},
		revisions.NonRowProviderRegistration{TargetKind: "indicator_state_interval", Provider: indicatorChildProvider},
	)
	if err != nil {
		return nil, fmt.Errorf("compose revisions non-row rollback providers: %w", err)
	}
	return revisions.NewCommandService(revisions.CommandServiceDependencies{
		Database:                    db,
		ImportedAttributionResolver: attributionResolver,
		ProjectionRebuilder:         projectionadapters.NewRowProjector(db),
		DeleteRestoreProviders:      deleteRestoreProviders,
		RowRollbackProviders:        rowRollbackProviders,
		NonRowRollbackProviders:     nonRowRollbackProviders,
	})
}
