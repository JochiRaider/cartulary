package incidentportabilityassembly

import (
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
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func NewCatalog() (*sourceport.Catalog, error) {
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
			records.NewIncidentBundleSourcePort(),
			timeline.NewIncidentBundleSourcePort(),
			parties.NewIncidentBundleSourcePort(),
			entities.NewIncidentBundleSourcePort(),
			indicators.NewIncidentBundleSourcePort(),
			artifacts.NewIncidentBundleSourcePort(),
			tasksdecisions.NewIncidentBundleSourcePort(),
			evidence.NewIncidentBundleSourcePort(),
			assessments.NewIncidentBundleSourcePort(),
			links.NewIncidentBundleSourcePort(),
			revisions.NewIncidentBundleSourcePort(),
			savedviews.NewIncidentBundleSourcePort(),
		},
		RequiredPathsByVersion: map[int][]string{1: v1, 2: v2},
		AllowedRelationIDs: map[string]struct{}{
			"incident-core": {}, "record-revisions": {}, "timeline-source": {},
			"parties": {}, "entity-source": {}, "indicator-source": {},
			"artifacts-and-optional-surfaces": {}, "tasks-and-decisions": {},
			"evidence-source-and-handles": {}, "assessment-source": {},
			"links-and-tags": {}, "savedviews": {},
		},
		SpecialConsumers: map[int]map[string]string{1: special, 2: special},
	})
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
