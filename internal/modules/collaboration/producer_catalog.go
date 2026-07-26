package collaboration

import (
	"fmt"
	"sort"
	"strings"
)

// RecordProducerDescriptor is the code-backed admission catalog for record
// owners that publish record_changed intents. The catalog is deliberately
// independent of the validation-only projection-provider manifest.
type RecordProducerDescriptor struct {
	RecordType   string
	FieldPrefix  string
	ViewSchemaID string
}

var recordProducerCatalog = []RecordProducerDescriptor{
	{RecordType: "artifact", FieldPrefix: "comm_log", ViewSchemaID: "cartulary.view.comm_log.v1"},
	{RecordType: "artifact", FieldPrefix: "finding", ViewSchemaID: "cartulary.view.findings.v1"},
	{RecordType: "artifact", FieldPrefix: "handoff", ViewSchemaID: "cartulary.view.handoff.v1"},
	{RecordType: "artifact", FieldPrefix: "forensic_keyword", ViewSchemaID: "cartulary.view.forensic_keywords.v1"},
	{RecordType: "artifact", FieldPrefix: "lesson", ViewSchemaID: "cartulary.view.lesson.v1"},
	{RecordType: "artifact", FieldPrefix: "note", ViewSchemaID: "cartulary.view.notes.v1"},
	{RecordType: "artifact", FieldPrefix: "investigative_query", ViewSchemaID: "cartulary.view.investigative_queries.v1"},
	{RecordType: "artifact", FieldPrefix: "status_review", ViewSchemaID: "cartulary.view.status_review.v1"},
	{RecordType: "assessment", ViewSchemaID: "cartulary.view.assessments.v1"},
	{RecordType: "decision", ViewSchemaID: "cartulary.view.decisions.v1"},
	{RecordType: "evidence", ViewSchemaID: "cartulary.view.evidence.v1"},
	{RecordType: "host", ViewSchemaID: "cartulary.view.hosts.v1"},
	{RecordType: "identity", ViewSchemaID: "cartulary.view.identities.v1"},
	{RecordType: "indicator", ViewSchemaID: "cartulary.view.indicators.v1"},
	{RecordType: "party", ViewSchemaID: "cartulary.view.parties.v1"},
	{RecordType: "task_request", ViewSchemaID: "cartulary.view.task_requests.v1"},
	{RecordType: "timeline_event", ViewSchemaID: "cartulary.view.timeline.v2"},
}

func RecordProducerCatalog() []RecordProducerDescriptor {
	return append([]RecordProducerDescriptor(nil), recordProducerCatalog...)
}

func RecordProducerViewSchema(recordType string, row map[string]any) (string, error) {
	fieldPrefix := firstCellPrefix(row)
	if fieldPrefix == "" && recordType == "artifact" {
		fieldPrefix = artifactTypeFromSourceSnapshot(row)
	}
	for _, descriptor := range recordProducerCatalog {
		if descriptor.RecordType != recordType {
			continue
		}
		if descriptor.FieldPrefix == "" || descriptor.FieldPrefix == fieldPrefix {
			return descriptor.ViewSchemaID, nil
		}
	}
	return "", fmt.Errorf("record_changed producer is not mapped for record type %q and field prefix %q", recordType, fieldPrefix)
}

func artifactTypeFromSourceSnapshot(row map[string]any) string {
	source, ok := row["source"].(map[string]any)
	if !ok {
		return ""
	}
	artifactType, _ := source["artifact_type"].(string)
	return artifactType
}

func firstCellPrefix(row map[string]any) string {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return ""
	}
	keys := make([]string, 0, len(cells))
	for key := range cells {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return strings.SplitN(keys[0], ".", 2)[0]
}
