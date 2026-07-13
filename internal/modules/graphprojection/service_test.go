package graphprojection

import (
	"context"
	"testing"
)

func TestProjectEphemeralReturnsFullNonRetainedResult(t *testing.T) {
	nonces := []string{"ephemeral-nonce-1", "ephemeral-nonce-2"}
	service := NewService(ServiceOptions{
		Now: fixedTime,
		NewNonce: func() (string, error) {
			nonce := nonces[0]
			nonces = nonces[1:]
			return nonce, nil
		},
	})
	input := mustJSON(t, minimalInput(t, "ephemeral-result"))
	first, err := service.ProjectEphemeral(context.Background(), EphemeralProjectionRequest{ProjectionInput: input})
	if err != nil {
		t.Fatalf("first ephemeral projection: %v", err)
	}
	second, err := service.ProjectEphemeral(context.Background(), EphemeralProjectionRequest{ProjectionInput: input})
	if err != nil {
		t.Fatalf("second ephemeral projection: %v", err)
	}
	firstID, _ := first.Data["ephemeral_projection_id"].(string)
	secondID, _ := second.Data["ephemeral_projection_id"].(string)
	if firstID == "" || secondID == "" || firstID == secondID || firstID[:4] != "gpe_" || secondID[:4] != "gpe_" {
		t.Fatalf("unexpected ephemeral identities: first=%q second=%q", firstID, secondID)
	}
	for _, field := range []string{"projection_schema_id", "graph_view_id", "graph_view_key", "source_snapshot_id", "projection_version", "generated_at", "properties", "metadata", "schema_registry", "vertices", "edges", "validation_summary", "consumer_capabilities"} {
		if _, ok := first.Data[field]; !ok {
			t.Fatalf("ephemeral result omitted %s: %#v", field, first.Data)
		}
	}
	for _, forbidden := range []string{"projection_run_id", "accepted_at", "started_at", "completed_at", "retention_expires_at", "idempotency_expires_at"} {
		if _, ok := first.Data[forbidden]; ok {
			t.Fatalf("ephemeral result leaked %s", forbidden)
		}
	}
}
