package revisions

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type runtimeHistoryActorProbe struct {
	commandServiceTestAttributionResolver
	resolve func(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, []string) (map[string]string, error)
}

func (p runtimeHistoryActorProbe) ResolveImportedHistoryActorsTx(ctx context.Context, tx pgx.Tx, incident, changeSet uuid.UUID, ids []string) (map[string]string, error) {
	return p.resolve(ctx, tx, incident, changeSet, ids)
}

func exerciseRuntimeHistoryActors(t *testing.T) {
	t.Helper()
	const first = "00000000-0000-4000-8000-000000000001"
	const second = "00000000-0000-4000-8000-000000000002"
	const local = "00000000-0000-4000-8000-000000000003"
	incident, changeSet := uuid.New(), uuid.New()
	tx := &commandObservedTx{}
	ctx := context.Background()
	failure := errors.New("actor resolver failure")
	for _, test := range []struct {
		name      string
		mapping   map[string]string
		err       error
		wantError bool
	}{
		{name: "native"},
		{name: "imported", mapping: map[string]string{first: local, second: local}},
		{name: "missing", mapping: map[string]string{first: local}, wantError: true},
		{name: "zero", mapping: map[string]string{first: local, second: uuid.Nil.String()}, wantError: true},
		{name: "resolver failure", err: failure, wantError: true},
	} {
		t.Run("runtime actors "+test.name, func(t *testing.T) {
			value := map[string]any{"owner_user_id": second, "deleted_by_user_id": nil, "record_id": first, "number": json.Number("9007199254740993"), "nested": []any{map[string]any{"created_by_user_id": first, "label": second}}}
			original, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			store := &commandStore{importedAttributions: runtimeHistoryActorProbe{resolve: func(gotCtx context.Context, gotTx pgx.Tx, gotIncident, gotChangeSet uuid.UUID, ids []string) (map[string]string, error) {
				calls++
				if gotCtx != ctx || gotTx != tx || gotIncident != incident || gotChangeSet != changeSet || !reflect.DeepEqual(ids, []string{first, second}) {
					t.Fatalf("actor resolution escaped its scope: %v %v %v", gotIncident, gotChangeSet, ids)
				}
				return test.mapping, test.err
			}}}
			got, err := store.runtimeHistoryValuesTx(ctx, tx, incident, changeSet, nil, value)
			if calls != 1 || (err != nil) != test.wantError || (test.err != nil && !errors.Is(err, test.err)) {
				t.Fatalf("resolution calls=%d error=%v", calls, err)
			}
			if test.wantError {
				return
			}
			if len(got) != 2 || got[0] != nil {
				t.Fatalf("snapshot sides changed: %#v", got)
			}
			want := value
			if test.mapping != nil {
				want = map[string]any{"owner_user_id": local, "deleted_by_user_id": nil, "record_id": first, "number": json.Number("9007199254740993"), "nested": []any{map[string]any{"created_by_user_id": local, "label": second}}}
			}
			if !reflect.DeepEqual(got[1], want) {
				t.Fatalf("runtime image=%#v want=%#v", got[1], want)
			}
			if test.mapping != nil {
				got[1]["owner_user_id"] = "provider mutation"
				got[1]["nested"].([]any)[0].(map[string]any)["label"] = "provider nested mutation"
			}
			after, err := json.Marshal(value)
			if err != nil || string(after) != string(original) {
				t.Fatalf("retained history mutated: %s %v", after, err)
			}
		})
	}
}
