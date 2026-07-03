package incidentportability

import (
	"fmt"
	"strings"
)

type ImportTargetDescriptor struct {
	SourceFamily             string
	LogicalBundlePath        string
	OwnerModule              string
	OwnerPortSymbol          string
	TargetRelation           string
	StableRowIdentity        []string
	RequiredColumnSet        []string
	NullableColumnPolicy     string
	PostImportInvariantCheck string
	VisibilityPrecondition   string

	registryKey string
}

var (
	TargetRecords = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "records", LogicalBundlePath: "data/records.ndjson", OwnerModule: "records", OwnerPortSymbol: "records.ImportIncidentBundleFilesTx", TargetRelation: "records", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id", "record_type", "row_version"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "records owner projection/revision validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetTimelineTimeConversionProfiles = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "timeline", LogicalBundlePath: "data/timeline_time_conversion_profiles.ndjson", OwnerModule: "timeline", OwnerPortSymbol: "timeline.ImportIncidentBundleFilesTx", TargetRelation: "timeline_time_conversion_profiles", StableRowIdentity: []string{"incident_id"}, RequiredColumnSet: []string{"incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "timeline owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetTimelineEvents = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "timeline", LogicalBundlePath: "data/timeline_events.ndjson", OwnerModule: "timeline", OwnerPortSymbol: "timeline.ImportIncidentBundleFilesTx", TargetRelation: "timeline_events", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "timeline owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetParties = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "parties", LogicalBundlePath: "data/parties.ndjson", OwnerModule: "parties", OwnerPortSymbol: "parties.ImportIncidentBundleFilesTx", TargetRelation: "parties", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "parties owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetEntityMentions = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "entities", LogicalBundlePath: "data/entity_mentions.ndjson", OwnerModule: "entities", OwnerPortSymbol: "entities.ImportIncidentBundleFilesTx", TargetRelation: "entity_mentions", StableRowIdentity: []string{"entity_mention_id"}, RequiredColumnSet: []string{"entity_mention_id", "source_record_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "entities owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetHosts = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "entities", LogicalBundlePath: "data/hosts.ndjson", OwnerModule: "entities", OwnerPortSymbol: "entities.ImportIncidentBundleFilesTx", TargetRelation: "hosts", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "entities owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetIdentities = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "entities", LogicalBundlePath: "data/identities.ndjson", OwnerModule: "entities", OwnerPortSymbol: "entities.ImportIncidentBundleFilesTx", TargetRelation: "identities", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "entities owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetEntityPreservedIdentifiers = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "entities", LogicalBundlePath: "data/entity_preserved_identifiers.ndjson", OwnerModule: "entities", OwnerPortSymbol: "entities.ImportIncidentBundleFilesTx", TargetRelation: "entity_preserved_identifiers", StableRowIdentity: []string{"entity_preserved_identifier_id"}, RequiredColumnSet: []string{"entity_preserved_identifier_id", "record_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "entities owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetEntityAliases = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "entities", LogicalBundlePath: "data/entity_aliases.ndjson", OwnerModule: "entities", OwnerPortSymbol: "entities.ImportIncidentBundleFilesTx", TargetRelation: "entity_aliases", StableRowIdentity: []string{"entity_alias_id"}, RequiredColumnSet: []string{"entity_alias_id", "record_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "entities owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetIndicators = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "indicators", LogicalBundlePath: "data/indicators.ndjson", OwnerModule: "indicators", OwnerPortSymbol: "indicators.ImportIncidentBundleFilesTx", TargetRelation: "indicators", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "indicators owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetIndicatorObservations = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "indicators", LogicalBundlePath: "data/indicator_observations.ndjson", OwnerModule: "indicators", OwnerPortSymbol: "indicators.ImportIncidentBundleFilesTx", TargetRelation: "indicator_observations", StableRowIdentity: []string{"indicator_observation_id"}, RequiredColumnSet: []string{"indicator_observation_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "indicators owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetIndicatorStateIntervals = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "indicators", LogicalBundlePath: "data/indicator_state_intervals.ndjson", OwnerModule: "indicators", OwnerPortSymbol: "indicators.ImportIncidentBundleFilesTx", TargetRelation: "indicator_state_intervals", StableRowIdentity: []string{"indicator_state_interval_id"}, RequiredColumnSet: []string{"indicator_state_interval_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "indicators owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetArtifacts = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "artifacts", LogicalBundlePath: "data/artifacts.ndjson", OwnerModule: "artifacts", OwnerPortSymbol: "artifacts.ImportIncidentBundleFilesTx", TargetRelation: "artifacts", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "artifacts owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetArtifactFindings = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "artifacts", LogicalBundlePath: "data/artifact_findings.ndjson", OwnerModule: "artifacts", OwnerPortSymbol: "artifacts.ImportIncidentBundleFilesTx", TargetRelation: "artifact_findings", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "artifacts owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetArtifactInvestigativeQueries = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "artifacts", LogicalBundlePath: "data/artifact_investigative_queries.ndjson", OwnerModule: "artifacts", OwnerPortSymbol: "artifacts.ImportIncidentBundleFilesTx", TargetRelation: "artifact_investigative_queries", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "artifacts owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetArtifactForensicKeywords = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "artifacts", LogicalBundlePath: "data/artifact_forensic_keywords.ndjson", OwnerModule: "artifacts", OwnerPortSymbol: "artifacts.ImportIncidentBundleFilesTx", TargetRelation: "artifact_forensic_keywords", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "artifacts owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetHandoffRiskRefs = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "artifacts", LogicalBundlePath: "data/handoff_risk_refs.ndjson", OwnerModule: "artifacts", OwnerPortSymbol: "artifacts.ImportIncidentBundleFilesTx", TargetRelation: "handoff_risk_refs", StableRowIdentity: []string{"risk_ref_id"}, RequiredColumnSet: []string{"risk_ref_id", "handoff_record_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "artifacts owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetTaskRequests = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "tasksdecisions", LogicalBundlePath: "data/task_requests.ndjson", OwnerModule: "tasksdecisions", OwnerPortSymbol: "tasksdecisions.ImportIncidentBundleFilesTx", TargetRelation: "task_requests", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "tasks/decisions owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetDecisions = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "tasksdecisions", LogicalBundlePath: "data/decisions.ndjson", OwnerModule: "tasksdecisions", OwnerPortSymbol: "tasksdecisions.ImportIncidentBundleFilesTx", TargetRelation: "decisions", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "tasks/decisions owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetEvidenceRecords = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "evidence", LogicalBundlePath: "data/evidence_records.ndjson", OwnerModule: "evidence", OwnerPortSymbol: "evidence.ImportIncidentBundleFilesTx", TargetRelation: "evidence", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "evidence owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetEvidenceCustodyEvents = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "evidence", LogicalBundlePath: "data/evidence_custody_events.ndjson", OwnerModule: "evidence", OwnerPortSymbol: "evidence.ImportIncidentBundleFilesTx", TargetRelation: "evidence_custody_events", StableRowIdentity: []string{"custody_event_id"}, RequiredColumnSet: []string{"custody_event_id", "evidence_record_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "evidence owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetObjectBlobs = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "evidence", LogicalBundlePath: "data/object_blobs.ndjson", OwnerModule: "evidence", OwnerPortSymbol: "evidence.ImportIncidentBundleFilesTx", TargetRelation: "object_blobs", StableRowIdentity: []string{"object_blob_id"}, RequiredColumnSet: []string{"object_blob_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "evidence blob owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetAssessments = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "assessments", LogicalBundlePath: "data/compromise_assessments.ndjson", OwnerModule: "assessments", OwnerPortSymbol: "assessments.ImportIncidentBundleFilesTx", TargetRelation: "assessments", StableRowIdentity: []string{"record_id"}, RequiredColumnSet: []string{"record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "assessments owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetRecordLinks = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "links", LogicalBundlePath: "data/record_links.ndjson", OwnerModule: "links", OwnerPortSymbol: "links.ImportIncidentBundleFilesTx", TargetRelation: "record_links", StableRowIdentity: []string{"record_link_id"}, RequiredColumnSet: []string{"record_link_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "links owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetRecordTags = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "links", LogicalBundlePath: "data/record_tags.ndjson", OwnerModule: "links", OwnerPortSymbol: "links.ImportIncidentBundleFilesTx", TargetRelation: "record_tags", StableRowIdentity: []string{"record_tag_id"}, RequiredColumnSet: []string{"record_tag_id", "record_id", "incident_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "links owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetChangeSets = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "revisions", LogicalBundlePath: "data/change_sets.ndjson", OwnerModule: "revisions", OwnerPortSymbol: "revisions.ImportIncidentBundleFilesTx", TargetRelation: "change_sets", StableRowIdentity: []string{"change_set_id"}, RequiredColumnSet: []string{"change_set_id", "incident_id", "actor_user_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "revisions owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetChangeSetMutations = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "revisions", LogicalBundlePath: "data/change_set_mutations.ndjson", OwnerModule: "revisions", OwnerPortSymbol: "revisions.ImportIncidentBundleFilesTx", TargetRelation: "change_set_mutations", StableRowIdentity: []string{"change_set_id", "sequence_no"}, RequiredColumnSet: []string{"change_set_id", "sequence_no"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "revisions owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetRecordRevisions = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "revisions", LogicalBundlePath: "data/record_revisions.ndjson", OwnerModule: "revisions", OwnerPortSymbol: "revisions.ImportIncidentBundleFilesTx", TargetRelation: "record_revisions", StableRowIdentity: []string{"revision_id"}, RequiredColumnSet: []string{"revision_id", "change_set_id", "record_id"}, NullableColumnPolicy: "owner-defined", PostImportInvariantCheck: "revisions owner validation", VisibilityPrecondition: "incident unpublished",
	})
	TargetSavedViews = newImportTargetDescriptor(ImportTargetDescriptor{
		SourceFamily: "savedviews", LogicalBundlePath: "data/saved_views.ndjson", OwnerModule: "savedviews", OwnerPortSymbol: "savedviews.ImportIncidentBundleFilesTx", TargetRelation: "saved_views", StableRowIdentity: []string{"saved_view_id"}, RequiredColumnSet: []string{"saved_view_id", "incident_id", "view_schema_id", "scope", "display_name", "query_json", "layout_json", "owner_user_id", "created_at", "updated_at", "saved_view_version"}, NullableColumnPolicy: "system owner_user_id may be null; private/shared owner_user_id required before remap", PostImportInvariantCheck: "savedviews owner validation", VisibilityPrecondition: "incident unpublished",
	})
)

var importTargetsByRegistryKey map[string]ImportTargetDescriptor

func newImportTargetDescriptor(target ImportTargetDescriptor) ImportTargetDescriptor {
	key := target.TargetRelation
	if strings.TrimSpace(key) == "" {
		panic("incident portability import target missing relation")
	}
	if strings.TrimSpace(target.LogicalBundlePath) == "" || strings.TrimSpace(target.OwnerModule) == "" || len(target.StableRowIdentity) == 0 {
		panic("incident portability import target missing required descriptor fields")
	}
	target.registryKey = key
	if importTargetsByRegistryKey == nil {
		importTargetsByRegistryKey = map[string]ImportTargetDescriptor{}
	}
	if _, exists := importTargetsByRegistryKey[key]; exists {
		panic("duplicate incident portability import target: " + key)
	}
	importTargetsByRegistryKey[key] = target
	return target
}

func RegisteredImportTargets() []ImportTargetDescriptor {
	targets := make([]ImportTargetDescriptor, 0, len(importTargetsByRegistryKey))
	for _, target := range importTargetsByRegistryKey {
		targets = append(targets, target)
	}
	return targets
}

func validateRegisteredImportTarget(target ImportTargetDescriptor) error {
	if strings.TrimSpace(target.registryKey) == "" {
		return &VerificationFailure{ReasonCode: "malformed_manifest"}
	}
	registered, ok := importTargetsByRegistryKey[target.registryKey]
	if !ok || registered.TargetRelation != target.TargetRelation || registered.LogicalBundlePath != target.LogicalBundlePath {
		return &VerificationFailure{ReasonCode: "malformed_manifest"}
	}
	return nil
}

func SourceRowID(target ImportTargetDescriptor, row map[string]any) string {
	if err := validateRegisteredImportTarget(target); err != nil {
		return ""
	}
	parts := make([]string, 0, len(target.StableRowIdentity))
	for _, column := range target.StableRowIdentity {
		value := StringFromAny(row[column])
		if strings.TrimSpace(value) == "" {
			return ""
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, ":")
}

func SourceRowIDForRelation(relation string, row map[string]any) string {
	target, ok := importTargetsByRegistryKey[relation]
	if !ok {
		return ""
	}
	return SourceRowID(target, row)
}

func ValidateRequiredColumns(target ImportTargetDescriptor, row map[string]any) error {
	if err := validateRegisteredImportTarget(target); err != nil {
		return err
	}
	for _, column := range target.RequiredColumnSet {
		if _, ok := row[column]; !ok {
			return &VerificationFailure{ReasonCode: "malformed_manifest"}
		}
	}
	if SourceRowID(target, row) == "" {
		return &VerificationFailure{ReasonCode: "malformed_manifest"}
	}
	return nil
}

func (target ImportTargetDescriptor) String() string {
	return fmt.Sprintf("%s:%s", target.SourceFamily, target.TargetRelation)
}
