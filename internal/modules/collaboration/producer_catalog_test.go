package collaboration

import "testing"

func TestArtifactProducerCatalogUsesPublicRowPrefixes(t *testing.T) {
	tests := []struct {
		prefix       string
		viewSchemaID string
	}{
		{prefix: "comm_log", viewSchemaID: "cartulary.view.comm_log.v1"},
		{prefix: "finding", viewSchemaID: "cartulary.view.findings.v1"},
		{prefix: "forensic_keyword", viewSchemaID: "cartulary.view.forensic_keywords.v1"},
		{prefix: "handoff", viewSchemaID: "cartulary.view.handoff.v1"},
		{prefix: "investigative_query", viewSchemaID: "cartulary.view.investigative_queries.v1"},
		{prefix: "lesson", viewSchemaID: "cartulary.view.lesson.v1"},
		{prefix: "note", viewSchemaID: "cartulary.view.notes.v1"},
		{prefix: "status_review", viewSchemaID: "cartulary.view.status_review.v1"},
	}
	for _, test := range tests {
		t.Run(test.prefix, func(t *testing.T) {
			got, err := RecordProducerViewSchema("artifact", map[string]any{
				"cells": map[string]any{test.prefix + ".value": map[string]any{"value": "test"}},
			})
			if err != nil {
				t.Fatalf("resolve producer: %v", err)
			}
			if got != test.viewSchemaID {
				t.Fatalf("view schema = %q, want %q", got, test.viewSchemaID)
			}
		})
	}
}

func TestArtifactProducerCatalogUsesSourceTypeForDeletedRows(t *testing.T) {
	got, err := RecordProducerViewSchema("artifact", map[string]any{
		"source": map[string]any{"artifact_type": "note"},
	})
	if err != nil {
		t.Fatalf("resolve deleted artifact producer: %v", err)
	}
	if got != "cartulary.view.notes.v1" {
		t.Fatalf("view schema = %q, want cartulary.view.notes.v1", got)
	}
}
