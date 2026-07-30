package reportingprovider

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestReportingEnvelopeFactsContract_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-reporting-envelope-facts")
	actorID := uuid.New()
	incidentID := uuid.New()
	activeID := uuid.MustParse("20000000-0000-4000-8000-000000000001")
	deletedID := uuid.MustParse("20000000-0000-4000-8000-000000000002")
	now := time.Date(2026, time.July, 29, 18, 0, 0, 987654000, time.UTC)
	if _, err := db.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Records reporting actor', 'test-only', false, true, true)
`, actorID, "records-reporting-"+actorID.String()+"@example.test"); err != nil {
		t.Fatalf("seed reporting actor: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, $2, 'Records reporting incident', 'active', $3, $3)
`, incidentID, "RECORDS-REPORTING-"+incidentID.String(), actorID); err != nil {
		t.Fatalf("seed reporting incident: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type,
    created_by_user_id, created_at, updated_by_user_id, updated_at, row_version,
    deleted_at, deleted_by_user_id
) VALUES
    ($1, $3, 'artifact', $4, $5, $4, $5, 3, NULL, NULL),
    ($2, $3, 'evidence', $4, $5, $4, $5, 2, $5, $4)
`, activeID, deletedID, incidentID, actorID, now); err != nil {
		t.Fatalf("seed reporting envelopes: %v", err)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin reporting transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	supportRefs := map[string][]string{
		activeID.String(): {"support:beta", "support:alpha"},
	}
	output, err := CollectFactsTx(ctx, tx, incidentID, supportRefs)
	if err != nil {
		t.Fatalf("collect reporting envelope facts: %v", err)
	}
	if output.SchemaID != exportprovider.ProviderOutputSchemaID || output.ProviderKey != "records" {
		t.Fatalf("provider identity = (%q, %q)", output.SchemaID, output.ProviderKey)
	}
	if len(output.FieldFacts) != 1 {
		t.Fatalf("field fact count = %d; want 1", len(output.FieldFacts))
	}
	field := output.FieldFacts[0]
	if field.Path != "/record_envelopes/"+activeID.String() ||
		field.SourceFamily != "record_envelope" ||
		field.ContentClass != "derived_analytic" {
		t.Fatalf("field fact identity = %#v", field)
	}
	if !reflect.DeepEqual(field.SupportRefs, supportRefs[activeID.String()]) {
		t.Fatalf("field support refs = %#v; want %#v", field.SupportRefs, supportRefs[activeID.String()])
	}
	value, ok := field.Value.(map[string]any)
	if !ok {
		t.Fatalf("field value type = %T; want map[string]any", field.Value)
	}
	if _, exists := value["incident_id"]; exists {
		t.Fatalf("field value exposes incident_id: %#v", value)
	}
	if value["record_id"] != activeID.String() ||
		value["record_type"] != "artifact" ||
		value["row_version"] != float64(3) ||
		value["deleted_at"] != nil ||
		value["deleted_by_user_id"] != nil {
		t.Fatalf("field value = %#v", value)
	}
	if len(output.RecordFacts) != 1 {
		t.Fatalf("record fact count = %d; want 1", len(output.RecordFacts))
	}
	recordFact := output.RecordFacts[0]
	if recordFact.RecordID != activeID.String() ||
		recordFact.RecordType != "record_envelope" ||
		recordFact.SourceFamily != "record_envelope" ||
		!reflect.DeepEqual(recordFact.FieldPaths, []string{field.Path}) ||
		!reflect.DeepEqual(recordFact.SupportRefs, []string{"support:alpha", "support:beta"}) {
		t.Fatalf("record fact = %#v", recordFact)
	}
}
