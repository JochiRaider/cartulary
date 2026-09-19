package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type restoredPageTx struct {
	pgx.Tx // Unexpected lifecycle methods panic: the capability borrows this handle.
	rows   *restoredPageRows
	err    error
	calls  int
}

func (tx *restoredPageTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	tx.calls++
	return tx.rows, tx.err
}

type restoredPageRows struct {
	pgx.Rows
	index, count int
	closed       bool
	err, scanErr error
	buffer       []byte
}

func (r *restoredPageRows) Close()     { r.closed = true }
func (r *restoredPageRows) Next() bool { r.index++; return r.index <= r.count }
func (r *restoredPageRows) Err() error { return r.err }
func (r *restoredPageRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	*dest[0].(*uuid.UUID) = uuid.New()
	id := uuid.New()
	*dest[1].(**uuid.UUID) = &id
	r.buffer[0] = byte('0' + r.index)
	*dest[2].(*json.RawMessage) = r.buffer
	return nil
}

func TestRestoredNonterminalPageContract_Unit(t *testing.T) {
	ctx := context.Background()
	scope := RestoredNonterminalScope{JobKind: "test.kind", ExtensionOwnerProfileID: "test_owner"}
	for _, limit := range []int{-1, 0, 257} {
		tx := &restoredPageTx{}
		if page, err := ListRestoredNonterminalPageTx(ctx, tx, scope, nil, limit); page != nil || !errors.Is(err, ErrInvalidJobDefinition) || tx.calls != 0 {
			t.Fatalf("limit %d: %v %v", limit, page, err)
		}
	}
	zero := uuid.Nil
	for _, input := range []struct {
		ctx    context.Context
		tx     *restoredPageTx
		scope  RestoredNonterminalScope
		cursor *uuid.UUID
	}{
		{nil, &restoredPageTx{}, scope, nil}, {ctx, &restoredPageTx{}, RestoredNonterminalScope{}, nil}, {ctx, &restoredPageTx{}, scope, &zero},
	} {
		if _, err := ListRestoredNonterminalPageTx(input.ctx, input.tx, input.scope, input.cursor, 1); !errors.Is(err, ErrInvalidJobDefinition) || input.tx.calls != 0 {
			t.Fatalf("invalid input reached storage: %v", err)
		}
	}
	if _, err := ListRestoredNonterminalPageTx(ctx, nil, scope, nil, 1); !errors.Is(err, ErrInvalidJobDefinition) {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	tx := &restoredPageTx{}
	if _, err := ListRestoredNonterminalPageTx(canceled, tx, scope, nil, 1); !errors.Is(err, context.Canceled) || tx.calls != 0 {
		t.Fatalf("cancellation: %v", err)
	}
	rows := &restoredPageRows{count: 2, buffer: []byte("0")}
	tx = &restoredPageTx{rows: rows}
	page, err := ListRestoredNonterminalPageTx(ctx, tx, scope, nil, 256)
	if err != nil || !rows.closed || len(page) != 2 {
		t.Fatalf("complete closed page: %#v %v", page, err)
	}
	rows.buffer[0] = 'x'
	if string(page[0].HandlerPayloadJSON) != "1" || string(page[1].HandlerPayloadJSON) != "2" {
		t.Fatalf("borrowed payload bytes: %#v", page)
	}
	failure := errors.New("enumeration failure")
	for _, failing := range []*restoredPageTx{
		{err: failure}, {rows: &restoredPageRows{count: 2, scanErr: failure}}, {rows: &restoredPageRows{count: 2, buffer: []byte("0"), err: failure}},
	} {
		page, err := ListRestoredNonterminalPageTx(ctx, failing, scope, nil, 256)
		if page != nil || !errors.Is(err, failure) || (failing.rows != nil && !failing.rows.closed) {
			t.Fatalf("partial page or open rows: %v %v", page, err)
		}
	}
}
