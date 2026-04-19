package httptestx

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type AuthorizationOutcome struct {
	Status int
	Code   string
}

type MutationAttribution struct {
	ActorUserID string
	Source      string
	ClientTxnID string
	RequestID   string
	CreatedAt   time.Time
}

type ReplayCounts struct {
	ChangeSets   int
	MutationRows int
	Revisions    int
}

type ReplayExpectation struct {
	FirstStatus     int
	ReplayStatus    int
	DivergentStatus int
	DivergentCode   string
	StableBefore    ReplayCounts
	StableAfter     ReplayCounts
}

func RequireAuthorizationReDerived(t testing.TB, before AuthorizationOutcome, after AuthorizationOutcome) {
	t.Helper()
	if before.Status == after.Status && before.Code == after.Code {
		t.Fatalf("expected authorization outcome to change after re-derivation: before=%+v after=%+v", before, after)
	}
}

func RequireMutationAttribution(t testing.TB, got MutationAttribution, wantActorUserID string, wantSource string, wantClientTxnID string) {
	t.Helper()
	if got.ActorUserID == "" || got.Source == "" || got.RequestID == "" || got.CreatedAt.IsZero() {
		t.Fatalf("expected non-empty mutation attribution, got %+v", got)
	}
	if wantActorUserID != "" && got.ActorUserID != wantActorUserID {
		t.Fatalf("unexpected actor_user_id: got %q want %q", got.ActorUserID, wantActorUserID)
	}
	if wantSource != "" && got.Source != wantSource {
		t.Fatalf("unexpected mutation source: got %q want %q", got.Source, wantSource)
	}
	if wantClientTxnID != "" {
		if got.ClientTxnID == "" {
			t.Fatalf("expected non-empty client_txn_id, got %+v", got)
		}
		if got.ClientTxnID != wantClientTxnID {
			t.Fatalf("unexpected client_txn_id: got %q want %q", got.ClientTxnID, wantClientTxnID)
		}
	}
}

func RequireReplayScaffold(t testing.TB, got ReplayExpectation) {
	t.Helper()
	if got.FirstStatus == 0 || got.ReplayStatus == 0 || got.DivergentStatus == 0 {
		t.Fatalf("expected replay scaffold statuses to be populated, got %+v", got)
	}
	if got.DivergentCode == "" {
		t.Fatal("expected divergent replay code")
	}
	if got.StableBefore != got.StableAfter {
		t.Fatalf("expected replay counts to remain stable, before=%+v after=%+v", got.StableBefore, got.StableAfter)
	}
}

func RequireDivergentReplayRejected(t testing.TB, status int, code string, wantCode string) {
	t.Helper()
	if status == 0 {
		t.Fatal("expected divergent replay status")
	}
	if code != wantCode {
		t.Fatalf("unexpected divergent replay code: got %q want %q", code, wantCode)
	}
}

func RequireClosedVocabularyRejected(t testing.TB, code string, details map[string]any, wantField string, wantReasonCode string) {
	t.Helper()
	if code == "" {
		t.Fatal("expected closed-vocabulary rejection code")
	}
	if code != "invalid_mutation_payload" && code != "invalid_view_query" {
		t.Fatalf("unexpected closed-vocabulary rejection code: %q", code)
	}
	if details == nil {
		t.Fatal("expected closed-vocabulary rejection details")
	}
	if wantField != "" && details["field"] != wantField {
		t.Fatalf("unexpected closed-vocabulary field: got %v want %q", details["field"], wantField)
	}
	if wantReasonCode != "" && details["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected closed-vocabulary reason_code: got %v want %q", details["reason_code"], wantReasonCode)
	}
}

func RequireSecretSafePayload(t testing.TB, payload map[string]any, forbiddenKeys []string) {
	t.Helper()
	requireSecretSafeValue(t, payload, forbiddenKeys, "")
}

func RequireWritableStringNormalization(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("unexpected normalized string: got %q want %q", got, want)
	}
}

func RequireFieldKeyConformance(t testing.TB, fieldKeys []string, allowed []string) {
	t.Helper()
	if !slices.IsSorted(fieldKeys) {
		t.Fatalf("expected sorted field keys, got %v", fieldKeys)
	}
	for _, fieldKey := range fieldKeys {
		if !slices.Contains(allowed, fieldKey) {
			t.Fatalf("unexpected field key %q not in allowed set %v", fieldKey, allowed)
		}
	}
}

func RequireProjectionDeterminism(t testing.TB, first any, second any) {
	t.Helper()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("projection rebuild was not deterministic:\nfirst: %#v\nsecond: %#v", first, second)
	}
}

func RequireDefaultQueryMeta(t testing.TB, body map[string]any, viewSchemaID string) {
	t.Helper()

	metaValue, ok := body["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected success envelope meta object, got %T", body["meta"])
	}
	queryValue, ok := metaValue["query"].(map[string]any)
	if !ok {
		t.Fatalf("expected meta.query object, got %T", metaValue["query"])
	}

	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("view schema %q not registered", viewSchemaID)
	}
	expected := schema.DefaultQueryMeta()

	filters, ok := queryValue["filters"].([]any)
	if !ok {
		t.Fatalf("expected meta.query.filters array, got %T", queryValue["filters"])
	}
	if !reflect.DeepEqual(filters, expected.Filters) {
		t.Fatalf("unexpected meta.query.filters: got %#v want %#v", filters, expected.Filters)
	}

	sortValue, ok := queryValue["sort"].([]any)
	if !ok {
		t.Fatalf("expected meta.query.sort array, got %T", queryValue["sort"])
	}
	gotSort := make([]viewschema.SortEntry, 0, len(sortValue))
	for _, item := range sortValue {
		entry, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("expected sort entry object, got %T", item)
		}
		fieldKey, _ := entry["field_key"].(string)
		direction, _ := entry["direction"].(string)
		gotSort = append(gotSort, viewschema.SortEntry{
			FieldKey:  fieldKey,
			Direction: direction,
		})
	}
	if !reflect.DeepEqual(gotSort, expected.Sort) {
		t.Fatalf("unexpected meta.query.sort: got %#v want %#v", gotSort, expected.Sort)
	}
	if _, exists := queryValue["group_by"]; exists {
		t.Fatalf("expected default query meta to omit group_by, got %#v", queryValue["group_by"])
	}
}

func requireSecretSafeValue(t testing.TB, value any, forbiddenKeys []string, path string) {
	t.Helper()

	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if slices.Contains(forbiddenKeys, key) {
				t.Fatalf("payload exposed forbidden key %q at %s", key, joinPath(path, key))
			}
			requireSecretSafeValue(t, item, forbiddenKeys, joinPath(path, key))
		}
	case []any:
		for index, item := range typed {
			requireSecretSafeValue(t, item, forbiddenKeys, joinPath(path, index))
		}
	}
}

func joinPath(path string, part any) string {
	if path == "" {
		return toPathPart(part)
	}
	return path + "." + toPathPart(part)
}

func toPathPart(part any) string {
	switch typed := part.(type) {
	case string:
		return typed
	default:
		return "[" + fmt.Sprint(part) + "]"
	}
}
