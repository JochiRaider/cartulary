package incidentportabilityassembly

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/tasksdecisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func NewCatalog() (*sourceport.Catalog, error) {
	indicatorContribution, err := indicators.NewIncidentBundleContribution()
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: Indicators contribution: %w", err)
	}
	recordSubtypeCatalog, err := subtypepresence.NewCatalog([]subtypepresence.Contribution{
		timeline.IncidentBundleSubtypeContribution(),
		entities.IncidentBundleSubtypeContribution(),
		parties.IncidentBundleSubtypeContribution(),
		indicatorContribution.SubtypePresence,
		artifacts.IncidentBundleSubtypeContribution(),
		tasksdecisions.IncidentBundleSubtypeContribution(),
		evidence.IncidentBundleSubtypeContribution(),
		assessments.IncidentBundleSubtypeContribution(),
	})
	if err != nil {
		return nil, err
	}
	targetSemantics, err := revisionassembly.CurrentTargetSemanticsCatalog()
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: Revisions target semantics catalog: %w", err)
	}
	providerContributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: Revisions provider contributions: %w", err)
	}
	revisionsValidation, err := revisions.NewIncidentBundleValidationCatalog(
		incidentBundleRecordEnvelopeReader{store: records.NewStore()},
		targetSemantics,
		providerContributions,
	)
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: revisions validation catalog: %w", err)
	}
	artifactsSourcePort, err := artifacts.NewIncidentBundleSourcePort()
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: Artifacts source port: %w", err)
	}
	tasksDecisionsSourcePort, err := tasksdecisions.NewIncidentBundleSourcePort(tasksdecisionassembly.NewLinkFactsCapability())
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: Tasks/Decisions source port: %w", err)
	}
	linksSourcePort, err := links.NewIncidentBundleSourcePort()
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: Links source port: %w", err)
	}
	incidentsSourcePort, err := constructIncidentsSourcePort(incidents.NewIncidentBundleSourcePort)
	if err != nil {
		return nil, err
	}
	v3 := []string{
		"data/incident.json", "data/actors.ndjson", "data/records.ndjson",
		"data/timeline_time_profiles.ndjson", "data/timeline_records.ndjson",
		"data/timeline_source_provenance.ndjson", "data/parties.ndjson",
		"data/entity_mentions.ndjson", "data/hosts.ndjson", "data/identities.ndjson",
		"data/entity_preserved_identifiers.ndjson", "data/entity_aliases.ndjson",
		"data/indicators.ndjson", "data/indicator_observations.ndjson",
		"data/indicator_state_intervals.ndjson", "data/artifacts.ndjson",
		"data/artifact_findings.ndjson", "data/artifact_investigative_queries.ndjson",
		"data/artifact_forensic_keywords.ndjson", "data/handoff_risk_refs.ndjson",
		"data/task_requests.ndjson", "data/decisions.ndjson",
		"data/evidence_records.ndjson", "data/evidence_custody_events.ndjson",
		"data/object_blobs.ndjson", "data/compromise_assessments.ndjson",
		"data/record_links.ndjson", "data/tags.ndjson", "data/record_tags.ndjson",
		"data/change_sets.ndjson", "data/change_set_mutations.ndjson",
		"data/record_revisions.ndjson", "data/saved_views.ndjson",
		"data/reference_pack_refs.json",
	}
	special := map[string]string{
		"data/actors.ndjson":            "actors",
		"data/reference_pack_refs.json": "reference_pack_refs",
	}
	return sourceport.NewCatalog(sourceport.CatalogOptions{
		Ports: []sourceport.Port{
			incidentsSourcePort,
			records.NewIncidentBundleSourcePort(recordSubtypeCatalog),
			timeline.NewIncidentBundleSourcePort(),
			parties.NewIncidentBundleContribution(),
			entities.NewIncidentBundleSourcePort(),
			indicatorContribution.SourcePort,
			artifactsSourcePort,
			tasksDecisionsSourcePort,
			evidence.NewIncidentBundleSourcePort(),
			assessments.NewIncidentBundleSourcePort(),
			linksSourcePort,
			revisions.NewIncidentBundleSourcePort(revisionsValidation),
			savedviews.NewIncidentBundleSourcePort(),
		},
		RequiredPathsByVersion: map[int][]string{3: v3},
		AllowedRelationIDs: map[string]struct{}{
			"incident-core": {}, "record-envelope": {}, "record-revisions": {},
			"timeline-source": {},
			"parties":         {}, "entity-source": {}, "indicator-source": {},
			"artifacts-and-optional-surfaces": {}, "tasks-and-decisions": {},
			"evidence-source-and-handles": {}, "assessment-source": {},
			"links-and-tags": {}, "savedviews": {},
		},
		SpecialConsumers: map[int]map[string]string{3: special},
	})
}

func constructIncidentsSourcePort(
	constructor func() (sourceport.Port, error),
) (sourceport.Port, error) {
	port, err := constructor()
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: Incidents source port: %w", err)
	}
	return port, nil
}

type incidentBundleRecordEnvelopeReader struct {
	store *records.Store
}

func (reader incidentBundleRecordEnvelopeReader) RecordTypeTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (string, error) {
	if reader.store == nil {
		return "", pgx.ErrNoRows
	}
	envelope, err := reader.store.LoadEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		if errors.Is(err, records.ErrEnvelopeNotFound) {
			return "", pgx.ErrNoRows
		}
		return "", err
	}
	if envelope.IncidentID != incidentID {
		return "", pgx.ErrNoRows
	}
	return envelope.RecordType, nil
}
