package rollbackprovider

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestHostSourceForRollbackValue(t *testing.T) {
	t.Parallel()
	value := map[string]any{"source": map[string]any{
		"display_name": "Host A",
		"hostname":     nil,
		"host_state":   "canonical",
	}}
	got, ok := hostSourceForRollbackValue(value)
	if !ok {
		t.Fatal("hostSourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"display_name": "Host A", "hostname": nil, "host_state": "canonical"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("host source = %#v, want %#v", got, want)
	}
	if _, ok := hostSourceForRollbackValue(map[string]any{"record_id": "legacy", "display_name": "Legacy"}); ok {
		t.Fatal("schema-less direct row was accepted")
	}
	t.Run("lineage admission", func(t *testing.T) {
		assertEntityRollbackLineageAdmission(t, NewHostProvider(), "host_state")
	})
	t.Run("partial source preserves coherent lineage", assertHostRollbackLineageTuple)
}

func assertEntityRollbackLineageAdmission(t *testing.T, provider rollbackcontract.RowSourceProvider, stateKey string) {
	t.Helper()
	for _, invalid := range []any{"", " ", "invalid", nil, 7, true} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": map[string]any{stateKey: invalid}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Errorf("state %#v: error=%v, want nonreversible", invalid, err)
		}
	}
	for _, invalid := range []any{"", " ", "not-a-uuid", uuid.Nil.String(), 7, false, []any{}, map[string]any{}} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": map[string]any{"merged_into_record_id": invalid}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Errorf("merge reference %#v: error=%v, want nonreversible", invalid, err)
		}
	}
	for _, state := range []string{"stub", "canonical", "merged"} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": map[string]any{stateKey: state}}); err != nil {
			t.Errorf("valid partial state %q rejected: %v", state, err)
		}
	}
	for _, source := range []map[string]any{
		{"display_name": "Preserved"},
		{"merged_into_record_id": nil},
		{"merged_into_record_id": "00000000-0000-4000-8000-000000000123"},
	} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": source}); err != nil {
			t.Errorf("valid partial source %#v rejected: %v", source, err)
		}
	}
}

type hostLineageTx struct {
	pgx.Tx
	state  string
	target *uuid.UUID
	err    error
	writes int
}

func (tx *hostLineageTx) QueryRow(context.Context, string, ...any) pgx.Row { return tx }
func (tx *hostLineageTx) Scan(dest ...any) error {
	if tx.err != nil {
		return tx.err
	}
	*dest[0].(*string), *dest[1].(**uuid.UUID) = tx.state, tx.target
	return nil
}
func (tx *hostLineageTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	tx.writes++
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func assertHostRollbackLineageTuple(t *testing.T) {
	target := uuid.MustParse("00000000-0000-4000-8000-000000000123")
	readFailure := errors.New("lineage read failed")
	for _, tc := range []struct {
		name   string
		state  string
		target *uuid.UUID
		source map[string]any
		read   error
		want   error
	}{
		{"missing merge target", "canonical", nil, map[string]any{"host_state": "merged"}, nil, rollbackcontract.ErrTargetNotReversible},
		{"cleared merge target", "merged", &target, map[string]any{"merged_into_record_id": nil}, nil, rollbackcontract.ErrTargetNotReversible},
		{"restored canonical", "merged", &target, map[string]any{"host_state": "canonical", "merged_into_record_id": nil}, nil, nil},
		{"restored merged", "canonical", nil, map[string]any{"host_state": "merged", "merged_into_record_id": target.String()}, nil, nil},
		{"omitted lineage", "merged", &target, map[string]any{"display_name": "Preserved"}, nil, nil},
		{"missing source", "", nil, map[string]any{"display_name": "Preserved"}, pgx.ErrNoRows, rollbackcontract.ErrTargetNotFound},
		{"read failure", "", nil, map[string]any{"display_name": "Preserved"}, readFailure, readFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := &hostLineageTx{state: tc.state, target: tc.target, err: tc.read}
			err := NewHostProvider().RestoreTx(context.Background(), tx, rollbackcontract.RestoreRequest{RecordID: uuid.New(), RetainedValue: map[string]any{"source": tc.source}})
			if !errors.Is(err, tc.want) {
				t.Fatalf("restore error=%v, want %v", err, tc.want)
			}
			wantWrites := 0
			if tc.want == nil {
				wantWrites = 1
			}
			if tx.writes != wantWrites {
				t.Fatalf("writes=%d, want %d", tx.writes, wantWrites)
			}
		})
	}
}
