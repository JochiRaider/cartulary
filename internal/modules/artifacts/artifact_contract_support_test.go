package artifacts_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func importTextFields(entries ...string) []ownerfacade.ImportFieldValue {
	fields := make([]ownerfacade.ImportFieldValue, 0, len(entries)/2)
	for index := 0; index < len(entries); index += 2 {
		value := entries[index+1]
		fields = append(fields, ownerfacade.ImportFieldValue{
			FieldKey:        entries[index],
			NormalizedValue: ownerfacade.NewTextImportScalar(value),
		})
	}
	return fields
}

func seedArtifactContractRecord(
	t testing.TB,
	harness *appsupport.StoreHarness,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	recordType string,
	now time.Time,
) uuid.UUID {
	t.Helper()
	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin record seed: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	recordID, err := records.NewStore().InsertTx(context.Background(), tx, records.InsertParams{
		IncidentID: incidentID, RecordType: recordType,
		CreatedByUserID: actorID, CreatedAt: now,
		UpdatedByUserID: actorID, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("insert %s record envelope: %v", recordType, err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit %s record envelope: %v", recordType, err)
	}
	return recordID
}

func artifactContractCount(t testing.TB, harness *appsupport.StoreHarness, query string, args ...any) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("count artifact contract state: %v", err)
	}
	return count
}

func requireArtifactRecordChangedIntent(
	t testing.TB,
	harness *appsupport.StoreHarness,
	result artifacts.MutationResult,
	actorID uuid.UUID,
) {
	t.Helper()
	selector := collaborationsupport.IntentSelector{
		EventFamily: "record_changed", SourceChangeSetID: result.ChangeSetID.String(),
		SourceRecordID: result.RecordID.String(), SourceRowVersion: &result.RowVersion,
	}
	collaborationsupport.RequireIntentCount(t, harness.DB, selector, 1)
	raw := collaborationsupport.LoadLatestIntent(t, harness.DB, selector).CanonicalPayload
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode artifact record_changed intent: %v", err)
	}
	if payload["record_id"] != result.RecordID.String() ||
		payload["change_set_id"] != result.ChangeSetID.String() ||
		payload["client_txn_id"] != result.ClientTxnID ||
		payload["actor_user_id"] != actorID.String() ||
		payload["row_version"] != float64(result.RowVersion) {
		t.Fatalf("artifact record_changed identity = %#v, want result %#v actor %s", payload, result, actorID)
	}
	gotKeys := jsonStringSlice(t, payload["changed_field_keys"])
	wantKeys := append([]string(nil), result.ChangedFieldKeys...)
	slices.Sort(gotKeys)
	slices.Sort(wantKeys)
	if !slices.Equal(gotKeys, wantKeys) {
		t.Fatalf("artifact record_changed changed keys = %#v, want %#v", gotKeys, wantKeys)
	}
	affected, ok := payload["affected_views"].([]any)
	if !ok || len(affected) != 1 {
		t.Fatalf("artifact record_changed affected views = %#v, want exactly one", payload["affected_views"])
	}
	view, ok := affected[0].(map[string]any)
	if !ok || view["view_schema_id"] != result.ViewSchemaID || view["change_kind"] != "patch" {
		t.Fatalf("artifact record_changed affected view = %#v, want patch for %s", affected[0], result.ViewSchemaID)
	}
}

func jsonStringSlice(t testing.TB, raw any) []string {
	t.Helper()
	values, ok := raw.([]any)
	if !ok {
		t.Fatalf("JSON string list has type %T", raw)
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("JSON string list contains %T", value)
		}
		result = append(result, text)
	}
	return result
}

func requireArtifactText(
	t testing.TB,
	harness *appsupport.StoreHarness,
	query string,
	recordID uuid.UUID,
	want string,
) {
	t.Helper()
	var got string
	if err := harness.DB.QueryRow(context.Background(), query, recordID).Scan(&got); err != nil {
		t.Fatalf("query restored artifact text: %v", err)
	}
	if got != want {
		t.Fatalf("restored artifact text = %q, want %q", got, want)
	}
}

func requireArtifactTimestamp(
	t testing.TB,
	harness *appsupport.StoreHarness,
	recordID uuid.UUID,
	want time.Time,
) {
	t.Helper()
	var got time.Time
	if err := harness.DB.QueryRow(context.Background(), `SELECT updated_at FROM artifacts WHERE record_id = $1`, recordID).Scan(&got); err != nil {
		t.Fatalf("query restored artifact parent timestamp: %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("restored artifact parent updated_at = %s, want %s", got, want)
	}
}

func requireCount(
	t testing.TB,
	harness *appsupport.StoreHarness,
	query string,
	argsAndWant ...any,
) {
	t.Helper()
	if len(argsAndWant) == 0 {
		t.Fatal("count assertion requires an expected value")
	}
	want := argsAndWant[len(argsAndWant)-1].(int)
	args := argsAndWant[:len(argsAndWant)-1]
	if got := artifactContractCount(t, harness, query, args...); got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}

func artifactTextValues(entries ...string) map[string]artifacts.FieldValue {
	values := make(map[string]artifacts.FieldValue, len(entries)/2)
	for index := 0; index < len(entries); index += 2 {
		values[entries[index]] = artifactTextValue(entries[index+1])
	}
	return values
}

func artifactTextValue(value string) artifacts.FieldValue {
	return artifacts.FieldValue{Text: &value}
}

func artifactFieldValuePtr(value artifacts.FieldValue) *artifacts.FieldValue {
	return &value
}

type artifactImportActiveUserLookup struct{}

var errInjectedArtifactImportFinalization = errors.New("injected artifact import finalization failure")

type failingArtifactImportAppender struct {
	*revisions.Appender
}

func (failingArtifactImportAppender) AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error {
	return errInjectedArtifactImportFinalization
}

func (artifactImportActiveUserLookup) IsActiveUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error) {
	var active bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND is_active = true)`, userID).Scan(&active)
	return active, err
}

func artifactImportDependencies(database postgres.DB, appender *revisions.Appender) artifacts.ImportDependencies {
	return artifacts.ImportDependencies{
		RecordEnvelopes: records.NewStore(),
		ActiveUsers:     artifactImportActiveUserLookup{},
		Projections:     appsupport.ArtifactProjectionRows(database),
		Revisions:       appender,
		Collaboration:   collaborationsupport.NewPublicationAppender(),
	}
}

func artifactTypeForTestView(viewSchemaID string) string {
	switch viewSchemaID {
	case artifacts.NotesViewSchemaID:
		return "note"
	case artifacts.CommLogViewSchemaID:
		return "comm_log"
	case artifacts.HandoffViewSchemaID:
		return "handoff"
	case artifacts.StatusReviewViewSchemaID:
		return "status_review"
	case artifacts.LessonViewSchemaID:
		return "lesson"
	case artifacts.FindingsViewSchemaID:
		return "finding"
	case artifacts.InvestigativeQueriesViewSchemaID:
		return "investigative_query"
	case artifacts.ForensicKeywordsViewSchemaID:
		return "forensic_keyword"
	default:
		return ""
	}
}

func mustArtifactMutationFacade(
	t testing.TB,
	pool postgres.DB,
	codec conflicts.ConflictTokenCodec,
) *artifacts.MutationFacade {
	t.Helper()
	facade, err := workbookassembly.NewArtifactMutationContribution(
		pool,
		codec,
		revisionsupport.MustAppender(t),
		mustConflictFieldResolver(t),
		appsupport.ArtifactProjectionRows(pool),
		collaborationsupport.NewPublicationAppender(),
	)
	if err != nil {
		t.Fatalf("compose Artifacts mutation facade: %v", err)
	}
	return facade
}

func mustConflictFieldResolver(t testing.TB) conflicts.FieldResolver {
	t.Helper()
	resolver, err := revisionassembly.CurrentConflictFieldResolver()
	if err != nil {
		t.Fatalf("compose conflict field resolver: %v", err)
	}
	return resolver
}
