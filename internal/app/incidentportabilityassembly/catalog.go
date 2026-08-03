package incidentportabilityassembly

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
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
	taskDecisionContribution := tasksdecisions.NewIncidentBundleContribution()
	indicatorContribution := indicators.NewIncidentBundleContribution()
	recordSubtypeCatalog, err := subtypepresence.NewCatalog([]subtypepresence.Contribution{
		timeline.IncidentBundleSubtypeContribution(),
		entities.IncidentBundleSubtypeContribution(),
		parties.IncidentBundleSubtypeContribution(),
		indicatorContribution.SubtypePresence,
		artifacts.IncidentBundleSubtypeContribution(),
		taskDecisionContribution.SubtypePresence,
		evidence.IncidentBundleSubtypeContribution(),
		assessments.IncidentBundleSubtypeContribution(),
	})
	if err != nil {
		return nil, err
	}
	revisionsValidation, err := revisions.NewIncidentBundleValidationCatalog(
		incidentBundleRecordEnvelopeReader{store: records.NewStore()},
		revisionassembly.CurrentProviderContributions(),
	)
	if err != nil {
		return nil, fmt.Errorf("incident portability assembly: revisions validation catalog: %w", err)
	}
	v2 := []string{
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
	v1 := replaceTimelinePaths(v2)
	special := map[string]string{
		"data/actors.ndjson":            "actors",
		"data/reference_pack_refs.json": "reference_pack_refs",
	}
	return sourceport.NewCatalog(sourceport.CatalogOptions{
		Ports: []sourceport.Port{
			incidents.NewIncidentBundleSourcePort(),
			records.NewIncidentBundleSourcePort(recordSubtypeCatalog),
			timeline.NewIncidentBundleSourcePort(),
			parties.NewIncidentBundleSourcePort(),
			entities.NewIncidentBundleSourcePort(),
			indicatorContribution.SourcePort,
			artifacts.NewIncidentBundleSourcePort(),
			taskDecisionContribution.SourcePort,
			evidence.NewIncidentBundleSourcePort(),
			assessments.NewIncidentBundleSourcePort(),
			links.NewIncidentBundleSourcePort(),
			revisions.NewIncidentBundleSourcePort(revisionsValidation),
			savedviews.NewIncidentBundleSourcePort(),
		},
		RequiredPathsByVersion: map[int][]string{1: v1, 2: v2},
		AllowedRelationIDs: map[string]struct{}{
			"incident-core": {}, "record-envelope": {}, "record-revisions": {},
			"timeline-source": {},
			"parties":         {}, "entity-source": {}, "indicator-source": {},
			"artifacts-and-optional-surfaces": {}, "tasks-and-decisions": {},
			"evidence-source-and-handles": {}, "assessment-source": {},
			"links-and-tags": {}, "savedviews": {},
		},
		SpecialConsumers: map[int]map[string]string{1: special, 2: special},
	})
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

func replaceTimelinePaths(v2 []string) []string {
	v1 := make([]string, 0, len(v2)-1)
	for _, logicalPath := range v2 {
		switch logicalPath {
		case "data/timeline_time_profiles.ndjson",
			"data/timeline_records.ndjson",
			"data/timeline_source_provenance.ndjson":
			continue
		default:
			v1 = append(v1, logicalPath)
		}
	}
	return append(v1,
		"data/timeline_time_conversion_profiles.ndjson",
		"data/timeline_events.ndjson",
	)
}
